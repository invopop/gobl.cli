package internal

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gitlab.com/flimzy/testy"
)

func TestBulk(t *testing.T) {
	type tt struct {
		in   io.Reader
		want []*BulkResponse
	}

	tests := testy.NewTable()
	tests.Add("invalid input", tt{
		in: strings.NewReader("this ain't json"),
		want: []*BulkResponse{
			{
				SeqID:   0,
				Error:   "invalid character 'h' in literal true (expecting 'r')",
				IsFinal: true,
			},
		},
	})
	tests.Add("one verification", func(t *testing.T) interface{} {
		payload, err := ioutil.ReadFile("testdata/success.json")
		if err != nil {
			t.Fatal(err)
		}
		req, err := json.Marshal(BulkRequest{
			Action:  "verify",
			ReqID:   "asdf",
			Payload: payload,
		})
		if err != nil {
			t.Fatal(err)
		}
		return tt{
			in: bytes.NewReader(req),
			want: []*BulkResponse{
				{
					ReqID:   "asdf",
					SeqID:   0,
					IsFinal: false,
				},
				{
					SeqID:   1,
					IsFinal: true,
				},
			},
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		ch := Bulk(tt.in)
		results := []*BulkResponse{}
		for res := range ch {
			results = append(results, res)
		}
		if d := cmp.Diff(results, tt.want, cmpopts.IgnoreFields(BulkResponse{}, "Payload")); d != "" {
			t.Error(d)
		}
	})
}
