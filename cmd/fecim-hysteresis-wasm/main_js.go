//go:build js && wasm

package main

import (
	"encoding/json"
	"fecim-lattice-tools/workbench/hysteresis"
	"syscall/js"
)

func main() {
	bridge := hysteresis.NewBridge()
	identity, _ := json.Marshal(hysteresis.BridgeIdentity())
	js.Global().Set("fecimHysteresisIdentity", string(identity))
	dispatch := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 1 || args[0].Type() != js.TypeString {
			return string(bridge.Dispatch(nil))
		}
		// JS UTF-16 length is a cheap conservative guard before copying to Go.
		if js.Global().Get("Object").Invoke(args[0]).Get("length").Int() > hysteresis.MaxEnvelopeBytes {
			return string(bridge.Dispatch(make([]byte, hysteresis.MaxEnvelopeBytes+1)))
		}
		return string(bridge.Dispatch([]byte(args[0].String())))
	})
	js.Global().Set("fecimHysteresisDispatch", dispatch)
	select {}
}
