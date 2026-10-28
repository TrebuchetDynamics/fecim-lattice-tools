//go:build !js

// Native protocol driver for exact qualification of the production transport.
package main

import (
	"encoding/json"
	"fecim-lattice-tools/workbench/hysteresis"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: fecim-hysteresis-wasm identity|REQUEST_ARRAY.json")
		os.Exit(2)
	}
	if os.Args[1] == "identity" {
		json.NewEncoder(os.Stdout).Encode(hysteresis.BridgeIdentity())
		return
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var requests []json.RawMessage
	if err := json.NewDecoder(f).Decode(&requests); err != nil {
		panic(err)
	}
	bridge := hysteresis.NewBridge()
	replies := make([]json.RawMessage, 0, len(requests))
	for _, q := range requests {
		replies = append(replies, bridge.Dispatch(q))
	}
	json.NewEncoder(os.Stdout).Encode(replies)
}
