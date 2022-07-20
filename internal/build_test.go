package internal

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/flimzy/testy"

	"github.com/invopop/gobl/dsig"
	"github.com/invopop/gobl/note"
)

var (
	privateKey    = new(dsig.PrivateKey)
	publicKey     = new(dsig.PublicKey)
	verifyKeyText string
)

const signingKeyFile = "testdata/private.jwk"

func init() {
	data, err := ioutil.ReadFile(signingKeyFile)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(data, privateKey); err != nil {
		panic(err)
	}
	publicKey = privateKey.Public()
	pub, err := json.Marshal(publicKey)
	if err != nil {
		panic(err)
	}
	verifyKeyText = string(pub)
}

func Test_parseSets(t *testing.T) {
	tests := []struct {
		name string
		opts BuildOptions
		err  string
	}{
		{
			name: "invalid yaml",
			opts: BuildOptions{
				SetYAML: map[string]string{
					"foo": "[bar",
				},
			},
			err: `yaml: line 1: did not find expected ',' or ']'`,
		},
		{
			name: "valid yaml",
			opts: BuildOptions{
				SetYAML: map[string]string{
					"sring":  "bar",
					"number": "1234",
					"bool":   "true",
					"array":  "[1,2,3]",
					"object": `{"foo":"bar"}`,
				},
			},
		},
		{
			name: "root key",
			opts: BuildOptions{
				SetYAML: map[string]string{
					".": `{"foo":"bar"}`,
				},
			},
		},
		{
			name: "literal period",
			opts: BuildOptions{
				SetYAML: map[string]string{
					"\\.": `foo`,
				},
			},
		},
		{
			name: "period",
			opts: BuildOptions{
				SetYAML: map[string]string{
					"foo.bar": "baz",
				},
			},
		},
		{
			name: "anchored at root",
			opts: BuildOptions{
				SetYAML: map[string]string{
					".foo": "bar",
				},
			},
		},
		{
			name: "unmergable",
			opts: BuildOptions{
				SetYAML: map[string]string{
					".": "foo",
				},
			},
			err: "src and dst must be of same type",
		},
		{
			name: "explicit string",
			opts: BuildOptions{
				SetString: map[string]string{
					"foo": "1234",
				},
			},
		},
		{
			name: "root string",
			opts: BuildOptions{
				SetString: map[string]string{
					".": "1234",
				},
			},
			err: "src and dst must be of same type",
		},
		{
			name: "missing file",
			opts: BuildOptions{
				SetFile: map[string]string{
					"foo": "notfound.yaml",
				},
			},
			err: `open notfound.yaml: no such file or directory`,
		},
		{
			name: "invalid file",
			opts: BuildOptions{
				SetFile: map[string]string{
					"foo": "testdata/invalid.yaml",
				},
			},
			err: `yaml: line 2: found unexpected end of stream`,
		},
		{
			name: "unmergable",
			opts: BuildOptions{
				SetFile: map[string]string{
					".": "testdata/unmergable.yaml",
				},
			},
			err: `src and dst must be of same type`,
		},
		{
			name: "valid file",
			opts: BuildOptions{
				SetFile: map[string]string{
					"foo": "testdata/valid.yaml",
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			opts := tt.opts
			if opts.PrivateKey == nil {
				opts.PrivateKey = privateKey
			}
			got, err := parseSets(&opts)
			if tt.err == "" {
				assert.Nil(t, err)
			} else {
				assert.EqualError(t, err, tt.err)
			}
			if err != nil {
				return
			}
			if d := testy.DiffInterface(testy.Snapshot(t), got); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestBuild(t *testing.T) {
	type tt struct {
		opts BuildOptions
		err  string
	}

	tests := testy.NewTable()
	tests.Add("success", func(t *testing.T) interface{} {
		return tt{
			opts: BuildOptions{
				Data: testFileReader(t, "testdata/nototals.json"),
			},
		}
	})
	tests.Add("merge YAML", func(t *testing.T) interface{} {
		return tt{
			opts: BuildOptions{
				Data: testFileReader(t, "testdata/nototals.json"),
				SetYAML: map[string]string{
					"doc.supplier.name": "Other Company",
				},
			},
		}
	})
	tests.Add("invalid type", tt{
		opts: BuildOptions{
			Data: strings.NewReader(`{
				"$schema": "https://gobl.org/draft-0/envelope",
				"head": {
					"uuid": "9d8eafd5-77be-11ec-b485-5405db9a3e49",
					"dig": {
						"alg": "sha256",
						"val": "dce3bc3c8bf28f3d209f783917b3082ddc0339a66e9ba3aa63849e4357db1422"
					}
				},
				doc: {
					"$schema": "https://example.com/duck",
					"walk": "like a duck",
					"talk": "like a duck",
					"look": "like a duck"
				}
			}`),
		},
		err: `code=400, message=marshal: unregistered or invalid schema`,
	})
	tests.Add("with template", func(t *testing.T) interface{} {
		return tt{
			opts: BuildOptions{
				Template: strings.NewReader(`{"doc":{"supplier":{"name": "Other Company"}}}`),
				Data:     testFileReader(t, "testdata/noname.json"),
			},
		}
	})
	tests.Add("template with empty input", func(t *testing.T) interface{} {
		return tt{
			opts: BuildOptions{
				Template: testFileReader(t, "testdata/nosig.json"),
				Data:     strings.NewReader("{}"),
			},
		}
	})
	tests.Add("with signature", func(t *testing.T) interface{} {
		return tt{
			opts: BuildOptions{
				Template: testFileReader(t, "testdata/signed.json"),
				Data:     strings.NewReader("{}"),
			},
			err: `code=409, message=document has already been signed`,
		}
	})
	tests.Add("explicit type", func(t *testing.T) interface{} {
		return tt{
			opts: BuildOptions{
				Data:    testFileReader(t, "testdata/notype.json"),
				DocType: "bill.Invoice",
			},
		}
	})
	tests.Add("draft", func(t *testing.T) interface{} {
		return tt{
			opts: BuildOptions{
				Data: testFileReader(t, "testdata/draft.json"),
			},
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		t.Parallel()
		opts := tt.opts
		if opts.PrivateKey == nil {
			opts.PrivateKey = privateKey
		}
		got, err := Build(context.Background(), &opts)
		if tt.err == "" {
			assert.Nil(t, err)
		} else {
			assert.EqualError(t, err, tt.err)
		}
		if err != nil {
			return
		}
		re := testy.Replacement{
			Regexp:      regexp.MustCompile(`(?s)"sigs": \[.*\]`),
			Replacement: `"sigs": ["signature data"]`,
		}
		if d := testy.DiffAsJSON(testy.Snapshot(t), got, re); d != nil {
			t.Error(d)
		}
	})
}

func TestBuildWithPartialEnvelope(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		opts := &BuildOptions{
			Data:       testFileReader(t, "testdata/message.env.yaml"),
			PrivateKey: privateKey,
		}
		got, err := Build(context.Background(), opts)
		require.NoError(t, err)
		assert.NotEmpty(t, got.Head.UUID.String())
		assert.Empty(t, got.Signatures)

		msg, ok := got.Extract().(*note.Message)
		if assert.True(t, ok) {
			assert.Equal(t, "https://gobl.org/draft-0/note/message", got.Document.Schema().String())
			assert.Equal(t, "Test Message", msg.Title)
			assert.Equal(t, "We hope you like this test message!", msg.Content)
		}
	})
}
