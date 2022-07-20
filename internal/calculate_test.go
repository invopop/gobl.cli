package internal

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/flimzy/testy"

	"github.com/invopop/gobl/note"
)

func TestCalculate(t *testing.T) {
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

	// TODO: Add test that asserts errors returned by `env.Calculate` are
	// returned.

	tests.Run(t, func(t *testing.T, tt tt) {
		t.Parallel()
		opts := tt.opts
		if opts.PrivateKey == nil {
			opts.PrivateKey = privateKey
		}
		got, err := Calculate(context.Background(), &opts)
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

func TestCalculateWithPartialEnvelope(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		opts := &BuildOptions{
			Data:       testFileReader(t, "testdata/message.env.yaml"),
			PrivateKey: privateKey,
		}
		got, err := Calculate(context.Background(), opts)
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
