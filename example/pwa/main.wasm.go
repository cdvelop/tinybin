//go:build wasm
// +build wasm

package main

import (
	"syscall/js"

	"github.com/cdvelop/tinybin"
	. "github.com/cdvelop/tinystring"
)

type User struct {
	ID    uint64
	Name  string
	Email string
	Age   uint8
	Admin bool
}

func main() {
	// Your WebAssembly code here ok

	// Crear el elemento div
	dom := js.Global().Get("document").Call("createElement", "div")
	dom.Set("innerHTML", "Hello, WebAssembly! 0")

	// Obtener el body del documento y agregar el elemento
	body := js.Global().Get("document").Get("body")
	body.Call("appendChild", dom)

	logger := func(msg ...any) {

		js.Global().Get("console").Call("log", Translate(msg...).String())
	}

	tb := tinybin.New(logger)

	user := User{
		ID:    1,
		Name:  "John Doe",
		Email: "john.doe@example.com",
		Age:   30,
		Admin: false,
	}

	out, err := tb.Encode(user)
	if err != nil {
		logger("Error encoding user:", err)
		return
	}

	newUser := &User{}

	err = tb.Decode(out, newUser)
	if err != nil {
		logger("Error decoding user:", err)
		return
	}

	logger("Encoded data name:", newUser.Name)
	logger("Encoded data email:", newUser.Email)
	logger("Encoded data age:", newUser.Age)
	logger("Encoded data admin:", newUser.Admin)

	select {}
}
