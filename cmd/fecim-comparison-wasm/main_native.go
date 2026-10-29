//go:build !js

package main

import (
	"encoding/json"
	comparison "fecim-lattice-tools/workbench/modules/comparison"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "identity" {
		json.NewEncoder(os.Stdout).Encode(comparison.Identity())
		return
	}
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: fecim-comparison-wasm identity|requests.json")
		os.Exit(2)
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var requests []json.RawMessage
	if err = json.Unmarshal(data, &requests); err != nil {
		panic(err)
	}
	bridge := comparison.NewBridge()
	defer bridge.Close()
	out := make([]json.RawMessage, 0, len(requests))
	for _, q := range requests {
		out = append(out, bridge.Dispatch(q))
	}
	json.NewEncoder(os.Stdout).Encode(out)
}
