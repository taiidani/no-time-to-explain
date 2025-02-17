//go:build js && wasm

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"syscall/js"
)

var (
	document            js.Value
	addMessageBtn       js.Value
	addMessageDialog    js.Value
	addMessageForm      js.Value
	addMessageCancelBtn js.Value
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer cancel()

	document = js.Global().Get("document")
	addMessageBtn = document.Call("getElementById", "addMessage")
	addMessageDialog = document.Call("getElementById", "addMessageDialog")
	addMessageForm = document.Call("getElementById", "addMessageForm")
	addMessageCancelBtn = document.Call("getElementById", "addMessageCancel")
	events()

	fmt.Println("wasm loaded")
	<-ctx.Done()
}

func events() {
	addMessageBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		addMessageDialog.Call("setAttribute", "open", "true")
		return nil
	}))

	addMessageCancelBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		evt := args[0]
		evt.Call("preventDefault")
		evt.Call("stopPropagation")

		addMessageForm.Call("reset")
		addMessageDialog.Call("removeAttribute", "open")
		return nil
	}))
}
