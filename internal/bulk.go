package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/invopop/gobl/dsig"
)

// BulkRequest represents a single request in the stream of bulk requests.
type BulkRequest struct {
	// Action is the action to perform on the payload.
	Action string `json:"action"`
	// ReqID is an opaque request ID, which is returned with the associated
	// response.
	ReqID string `json:"req_id"`
	// Payload is the payload upon which to perform the action.
	Payload json.RawMessage `json:"payload"`
}

// BulkResponse represents a singel response in the stream of bulk responses.
type BulkResponse struct {
	// ReqID is an exact copy of the value provided in the request, if any.
	ReqID string `json:"req_id,omitempty"`
	// SeqID is the sequence ID of the request this response correspond to,
	// starting at 1.
	SeqID int64 `json:"seq_id"`
	// Payload is the response payload.
	Payload json.RawMessage `json:"payload"`
	// Error represents an error processing a request item.
	Error string `json:"error"`
	// IsFinal will be true once the end of the request input stream has been
	// reached, or an unrecoverable error has occurred.
	IsFinal bool `json:"is_final"`
}

// Bulk processes a stream of bulk requests.
func Bulk(ctx context.Context, in io.Reader) <-chan *BulkResponse {
	dec := json.NewDecoder(in)
	var seq int64
	resCh := make(chan *BulkResponse, 1)
	go func() {
		defer close(resCh)
		for {
			var req BulkRequest
			err := dec.Decode(&req)
			resCh <- processRequest(ctx, req, atomic.AddInt64(&seq, 1), err)
			if err != nil {
				return
			}
		}
	}()
	return resCh
}

func processRequest(ctx context.Context, req BulkRequest, seq int64, err error) *BulkResponse {
	res := &BulkResponse{
		ReqID: req.ReqID,
		SeqID: seq,
	}
	if err != nil {
		res.IsFinal = true
		if err != io.EOF {
			res.Error = err.Error()
		}
		return res
	}
	switch req.Action {
	case "verify":
		vrfy := &VerifyRequest{}
		if err := json.Unmarshal(req.Payload, vrfy); err != nil {
			res.Error = err.Error()
			return res
		}
		err := Verify(ctx, bytes.NewReader(vrfy.Data), vrfy.PublicKey)
		if err != nil {
			res.Error = err.Error()
			return res
		}
		res.Payload, _ = json.Marshal(VerifyResponse{OK: true})
	case "build":
		bld := &BuildRequest{}
		if err := json.Unmarshal(req.Payload, bld); err != nil {
			res.Error = err.Error()
			return res
		}
		env, err := Build(ctx, &BuildOptions{
			Data:       bytes.NewReader(bld.Data),
			PrivateKey: bld.PrivateKey,
		})
		if err != nil {
			res.Error = err.Error()
			return res
		}
		res.Payload, _ = json.Marshal(env)
	default:
		res.Error = fmt.Sprintf("Unrecognized action '%s'", req.Action)
	}
	return res
}

// VerifyRequest is the payload for a verification request.
type VerifyRequest struct {
	Data      json.RawMessage `json:"data"`
	PublicKey *dsig.PublicKey `json:"publickey"`
}

// VerifyResponse is the response to a verification request.
type VerifyResponse struct {
	OK bool `json:"ok"`
}

// BuildRequest is the payload for a build reqeuest.
type BuildRequest struct {
	Template   json.RawMessage  `json:"template"`
	Data       json.RawMessage  `json:"data"`
	PrivateKey *dsig.PrivateKey `json:"privatekey"`
	DocType    string           `json:"type"`
}
