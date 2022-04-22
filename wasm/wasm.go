//go:build js

package main

import (
	"fmt"
	"syscall/js"
)

func main() {
	fmt.Println("Hello, world!")
	js.Global().Call("addEventListener", "message", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		fmt.Println("got a message", args)
		return nil
	}))
	js.Global().Call("postMessage", map[string]interface{}{"ready": true})
	fmt.Println("waiting")
	c := make(chan struct{})
	<-c
	fmt.Println("exiting")
}
