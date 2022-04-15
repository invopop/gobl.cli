package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
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
				SeqID:   1,
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
		req, err := json.Marshal(map[string]interface{}{
			"action": "verify",
			"req_id": "asdf",
			"payload": map[string]interface{}{
				"data":      json.RawMessage(payload),
				"publickey": verifyKey,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		return tt{
			in: bytes.NewReader(req),
			want: []*BulkResponse{
				{
					ReqID:   "asdf",
					SeqID:   1,
					Payload: []byte(`{"ok":true}`),
					IsFinal: false,
				},
				{
					SeqID:   2,
					IsFinal: true,
				},
			},
		}
	})
	tests.Add("two verifications", func(t *testing.T) interface{} {
		payload, err := ioutil.ReadFile("testdata/success.json")
		if err != nil {
			t.Fatal(err)
		}
		req, err := json.Marshal(map[string]interface{}{
			"action": "verify",
			"req_id": "asdf",
			"payload": map[string]interface{}{
				"data":      json.RawMessage(payload),
				"publickey": verifyKey,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		return tt{
			in: io.MultiReader(bytes.NewReader(req), bytes.NewReader(req)),
			want: []*BulkResponse{
				{
					ReqID:   "asdf",
					SeqID:   1,
					Payload: []byte(`{"ok":true}`),
					IsFinal: false,
				},
				{
					ReqID:   "asdf",
					SeqID:   2,
					Payload: []byte(`{"ok":true}`),
					IsFinal: false,
				},
				{
					SeqID:   3,
					IsFinal: true,
				},
			},
		}
	})
	tests.Run(t, func(t *testing.T, tt tt) {
		ch := Bulk(context.Background(), tt.in)
		results := []*BulkResponse{}
		for res := range ch {
			results = append(results, res)
		}
		if d := cmp.Diff(results, tt.want); d != "" {
			t.Error(d)
		}
	})
}
