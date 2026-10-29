//go:build !js

package main

import (
	"context"
	"encoding/json"
	"fecim-lattice-tools/workbench/modules/circuits"
	"fmt"
	"os"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		panic("provide circuit command scenarios JSON")
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var scenarios []circuits.Scenario
	if err = json.Unmarshal(data, &scenarios); err != nil {
		panic(err)
	}
	outputs := [][]string{}
	for _, scenario := range scenarios {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		result, err := circuits.Replay(ctx, scenario)
		cancel()
		if err != nil {
			panic(err)
		}
		outputs = append(outputs, result)
	}
	data, err = json.Marshal(outputs)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(data))
}
