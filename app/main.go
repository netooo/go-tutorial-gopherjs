package app

import (
	"github.com/gopherjs/gopherjs/js"
)

func main() {
	document := js.Global.Get("document")
	meta := document.Call("createElement", "meta")
	meta.Set("name", "viewport")
	meta.Set("content", "width=device-width, initial-scale=1, shrink-to-fit=no")
	document.Get("head").Call("appendChild", meta)
}
