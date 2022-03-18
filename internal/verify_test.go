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
	tests.Add("missing sig", func(t *testing.T) interface{} {
		f, err := os.Open("testdata/nosig.json")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = f.Close() })

		return tt{
			in:  f,
			key: verifyKey,
			err: "code=422, message=sigs: cannot be blank.",
		}
	})
	tests.Add("missing key", func(t *testing.T) interface{} {
		f, err := os.Open("testdata/success.json")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = f.Close() })

		return tt{
			in:  f,
			err: "code=400, message=public key required",
		}
	})
	tests.Add("wrong public key", func(t *testing.T) interface{} {
		f, err := os.Open("testdata/success.json")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = f.Close() })

		var wrongKey = new(dsig.PublicKey)
		if err := json.Unmarshal([]byte(`{"use":"sig","kty":"EC","kid":"0f8726f9-5e1c-43e2-a0c8-55e696550800","crv":"P-256","alg":"ES256","x":"rmBvuoO4LJCnuo0htUaAhSgF7OdCEO6GAuI_o1dUP8s","y":"7qXu5-LHoQM52C_HJJ4eYoGlQnesFLcT76hIlDs-ZfU"}`), wrongKey); err != nil {
			panic(err)
		}

		return tt{
			in:  f,
			key: wrongKey,
			err: "code=422, message=key mismatch",
		}
	})
	tests.Add("draft", func(t *testing.T) interface{} {
		f, err := os.Open("testdata/draft.json")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = f.Close() })

		return tt{
			in:  f,
			key: verifyKey,
			err: "code=422, message=document is a draft",
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
