package internal

import (
	"encoding/json"
	"io"
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
	// starting at 0.
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
func Bulk(in io.Reader) <-chan *BulkResponse {
	dec := json.NewDecoder(in)
	var seq int64
	resCh := make(chan *BulkResponse, 1)
	go func() {
		defer close(resCh)
		for {
			var req BulkRequest
			err := dec.Decode(&req)
			res := &BulkResponse{
				ReqID:   req.ReqID,
				SeqID:   seq,
				IsFinal: err != nil,
			}
			if err != nil && err != io.EOF {
				res.Error = err.Error()
			}
			resCh <- res
			if err != nil {
				return
			}
			seq++
		}
	}()
	return resCh
}
