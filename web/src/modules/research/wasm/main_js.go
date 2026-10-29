//go:build js && wasm

package main

import (
	"encoding/json"
	"fecim-lattice-tools/workbench/moduleapi/research"
	"syscall/js"
)

func main() {
	bridge := research.New()
	identity, _ := json.Marshal(research.Identity())
	js.Global().Set("fecimModuleIdentity", string(identity))
	dispatch := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 1 || args[0].Type() != js.TypeString || js.Global().Get("Object").Invoke(args[0]).Get("length").Int() > 12<<20 {
			return string(bridge.Dispatch(nil))
		}
		return string(bridge.Dispatch([]byte(args[0].String())))
	})
	js.Global().Set("fecimModuleDispatch", dispatch)
	select {}
}
