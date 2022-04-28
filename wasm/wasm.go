//go:build js

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"syscall/js"

	"github.com/invopop/gobl.cli/internal"
)

func main() {
	fmt.Println("Hello, world!")

	r, w := io.Pipe()
	js.Global().Call("addEventListener", "message", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		jsonEvent := js.Global().Get("JSON").Call("stringify", args[0].Get("data")).String()
		fmt.Println("got a message", jsonEvent)
		if _, err := fmt.Fprintln(w, jsonEvent); err != nil {
			fmt.Fprintln(os.Stderr, "failed to send event: %s", err)
		}
		return nil
	}))

	js.Global().Call("postMessage", map[string]interface{}{"ready": true})

	fmt.Println("waiting")
	for result := range internal.Bulk(context.TODO(), r) {
		response := js.Global().Get("Object").New()
		if result.ReqID != "" {
			response.Set("req_id", result.ReqID)
		}
		response.Set("seq_id", result.SeqID)
		if len(result.Payload) > 0 {
			response.Set("payload", string(result.Payload))
		}
		if result.Error != "" {
			response.Set("error", result.Error)
		}
		if result.IsFinal {
			response.Set("is_final", true)
		}
		js.Global().Call("postMessage", response)
	}
	fmt.Println("exiting")
}
