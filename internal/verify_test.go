package internal

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/invopop/gobl/dsig"
	"github.com/stretchr/testify/assert"
	"gitlab.com/flimzy/testy"
)

var verifyKey = new(dsig.PublicKey)

const verifyKeyText = `{"use":"sig","kty":"EC","kid":"11da5c50-fc2f-442e-a97f-44f7aea73548","crv":"P-256","alg":"ES256","x":"TWfhO3rcAtagXo84QvtApjsSEinAw9yHueiNYArZbBU","y":"RCjVid5EkxVBV-e-r2bqaH1uhsmr6rKmysHuI8dS84g"}`

func init() {
	if err := json.Unmarshal([]byte(verifyKeyText), verifyKey); err != nil {
		panic(err)
	}
}

func TestVerify(t *testing.T) {
	type tt struct {
		in  io.Reader
		key *dsig.PublicKey
		err string
	}

	tests := testy.NewTable()
	tests.Add("validation pass", func(t *testing.T) interface{} {
		f, err := os.Open("testdata/success.json")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = f.Close() })

		return tt{
			in:  f,
			key: verifyKey,
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		t.Parallel()
		err := Verify(context.Background(), tt.in, tt.key)
		if tt.err == "" {
			assert.Nil(t, err)
		} else {
			assert.EqualError(t, err, tt.err)
		}

	})
}
