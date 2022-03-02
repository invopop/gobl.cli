package internal

import (
	"context"
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/invopop/gobl/dsig"
	"github.com/stretchr/testify/assert"
	"gitlab.com/flimzy/testy"
)

var signingKey = new(dsig.PrivateKey)

const signingKeyText = `{"use":"sig","kty":"EC","kid":"b7cee60f-204e-438b-a88f-021d28af6991","crv":"P-256","alg":"ES256","x":"wLez6TfqNReD3FUUyVP4Q7HAGdokmAfE6LwfcM28DlQ","y":"CIxURqWtiFIu9TaatRa85NkNsw1LZHw_ZQ9A45GW_MU","d":"xNx9MxONcuLk8Ai6s2isqXMZaDi3HNGLkFX-qiNyyeo"}`

func init() {
	if err := json.Unmarshal([]byte(signingKeyText), signingKey); err != nil {
		panic(err)
	}
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
				opts.PrivateKey = signingKey
			}
			got, err := parseSets(opts)
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
		f, err := os.Open("testdata/nototals.json")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = f.Close() })

		return tt{
			opts: BuildOptions{
				Data: f,
			},
		}
	})
	tests.Add("merge YAML", func(t *testing.T) interface{} {
		f, err := os.Open("testdata/nototals.json")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = f.Close() })

		return tt{
			opts: BuildOptions{
				Data: f,
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
		err: `code=400, message=marshal: unregistered schema: https://example.com/duck`,
	})
	tests.Add("with template", func(t *testing.T) interface{} {
		f, err := os.Open("testdata/noname.json")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = f.Close() })

		return tt{
			opts: BuildOptions{
				Template: strings.NewReader(`{"doc":{"supplier":{"name": "Other Company"}}}`),
				Data:     f,
			},
		}
	})
	tests.Add("template with empty input", func(t *testing.T) interface{} {
		f, err := os.Open("testdata/noname.json")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = f.Close() })

		return tt{
			opts: BuildOptions{
				Template: f,
				Data:     strings.NewReader("{}"),
			},
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		t.Parallel()
		opts := tt.opts
		if opts.PrivateKey == nil {
			opts.PrivateKey = signingKey
		}
		got, err := Build(context.Background(), opts)
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
