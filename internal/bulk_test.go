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
	tests.Add("success the failure", func(t *testing.T) interface{} {
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
			in: io.MultiReader(bytes.NewReader(req), strings.NewReader("not json")),
			want: []*BulkResponse{
				{
					ReqID:   "asdf",
					SeqID:   1,
					Payload: []byte(`{"ok":true}`),
					IsFinal: false,
				},
				{
					SeqID:   2,
					Error:   "invalid character 'o' in literal null (expecting 'u')",
					IsFinal: true,
				},
			},
		}
	})
	tests.Add("non-fatal payload error", func(t *testing.T) interface{} {
		req, err := json.Marshal(map[string]interface{}{
			"action":  "verify",
			"req_id":  "asdf",
			"payload": "not an object",
		})
		if err != nil {
			t.Fatal(err)
		}
		return tt{
			in: io.MultiReader(bytes.NewReader(req)),
			want: []*BulkResponse{
				{
					ReqID:   "asdf",
					SeqID:   1,
					IsFinal: false,
					Error:   `json: cannot unmarshal string into Go value of type internal.VerifyRequest`,
				},
				{
					SeqID:   2,
					IsFinal: true,
				},
			},
		}
	})
	tests.Add("non-fatal data error", func(t *testing.T) interface{} {
		req, err := json.Marshal(map[string]interface{}{
			"action": "verify",
			"req_id": "asdf",
			"payload": map[string]interface{}{
				"data":      json.RawMessage(`"oink"`),
				"publickey": verifyKey,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		return tt{
			in: io.MultiReader(bytes.NewReader(req)),
			want: []*BulkResponse{
				{
					ReqID:   "asdf",
					SeqID:   1,
					IsFinal: false,
					Error:   `code=400, message=error unmarshaling JSON: json: cannot unmarshal string into Go value of type gobl.Envelope`,
				},
				{
					SeqID:   2,
					IsFinal: true,
				},
			},
		}
	})
	tests.Add("one build, already signed", func(t *testing.T) interface{} {
		payload, err := ioutil.ReadFile("testdata/success.json")
		if err != nil {
			t.Fatal(err)
		}
		req, err := json.Marshal(map[string]interface{}{
			"action": "build",
			"req_id": "asdf",
			"payload": map[string]interface{}{
				"data":       json.RawMessage(payload),
				"privatekey": signingKey,
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
					Error:   `code=409, message=document has already been signed`,
					IsFinal: false,
				},
				{
					SeqID:   2,
					IsFinal: true,
				},
			},
		}
	})
	tests.Add("one build, success", func(t *testing.T) interface{} {
		payload, err := ioutil.ReadFile("testdata/nosig.json")
		if err != nil {
			t.Fatal(err)
		}
		req, err := json.Marshal(map[string]interface{}{
			"action": "build",
			"req_id": "asdf",
			"payload": map[string]interface{}{
				"data":       json.RawMessage(payload),
				"privatekey": signingKey,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		return tt{
			in: bytes.NewReader(req),
			want: []*BulkResponse{
				{
					ReqID: "asdf",
					SeqID: 1,
					Payload: json.RawMessage(`{
	"$schema": "https://gobl.org/draft-0/envelope",
	"head": {},
	"doc": {}
					}`),
					IsFinal: false,
				},
				{
					SeqID:   2,
					IsFinal: true,
				},
			},
		}
	})
	tests.Add("non-fatal payload error, build", func(t *testing.T) interface{} {
		req, err := json.Marshal(map[string]interface{}{
			"action":  "build",
			"req_id":  "asdf",
			"payload": "not an object",
		})
		if err != nil {
			t.Fatal(err)
		}
		return tt{
			in: io.MultiReader(bytes.NewReader(req)),
			want: []*BulkResponse{
				{
					ReqID:   "asdf",
					SeqID:   1,
					IsFinal: false,
					Error:   `json: cannot unmarshal string into Go value of type internal.BuildRequest`,
				},
				{
					SeqID:   2,
					IsFinal: true,
				},
			},
		}
	})
	tests.Add("non-fatal data error, build", func(t *testing.T) interface{} {
		req, err := json.Marshal(map[string]interface{}{
			"action": "build",
			"req_id": "asdf",
			"payload": map[string]interface{}{
				"data":      json.RawMessage(`"oink"`),
				"publickey": verifyKey,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		return tt{
			in: io.MultiReader(bytes.NewReader(req)),
			want: []*BulkResponse{
				{
					ReqID:   "asdf",
					SeqID:   1,
					IsFinal: false,
					Error:   "code=400, message=yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `oink` into map[string]interface {}",
				},
				{
					SeqID:   2,
					IsFinal: true,
				},
			},
		}
	})
	tests.Add("unknown action", func(t *testing.T) interface{} {
		req, err := json.Marshal(map[string]interface{}{
			"action": "frobnicate",
			"req_id": "asdf",
		})
		if err != nil {
			t.Fatal(err)
		}
		return tt{
			in: io.MultiReader(bytes.NewReader(req)),
			want: []*BulkResponse{
				{
					ReqID:   "asdf",
					SeqID:   1,
					Error:   "Unrecognized action 'frobnicate'",
					IsFinal: false,
				},
				{
					SeqID:   2,
					IsFinal: true,
				},
			},
		}
	})
	tests.Add("keygen", func(t *testing.T) interface{} {
		req, err := json.Marshal(map[string]interface{}{
			"action": "keygen",
			"req_id": "asdf",
		})
		if err != nil {
			t.Fatal(err)
		}
		return tt{
			in: io.MultiReader(bytes.NewReader(req)),
			want: []*BulkResponse{
				{
					ReqID:   "asdf",
					SeqID:   1,
					IsFinal: false,
				},
				{
					SeqID:   2,
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
		if d := cmp.Diff(tt.want, results, cmpopts.IgnoreFields(BulkResponse{}, "Payload")); d != "" {
			t.Error(d)
		}
	})
}
