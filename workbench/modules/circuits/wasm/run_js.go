//go:build js && wasm

package wasm

import (
	"encoding/json"
	"fecim-lattice-tools/workbench/modules/circuits"
	"syscall/js"
)

func Run() {
	bridge := circuits.NewBridge()
	identity, _ := json.Marshal(circuits.Identity())
	js.Global().Set("fecimModuleIdentity", string(identity))
	dispatch := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) != 1 || args[0].Type() != js.TypeString || js.Global().Get("Object").Invoke(args[0]).Get("length").Int() > 2<<20 {
			return string(bridge.Dispatch(nil))
		}
		return string(bridge.Dispatch([]byte(args[0].String())))
	})
	js.Global().Set("fecimModuleDispatch", dispatch)
	select {}
}
