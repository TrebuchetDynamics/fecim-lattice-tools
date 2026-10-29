// Command generate emits the authoritative, private local API schema. It opens
// no grants, constructs no module and starts no listener.
package main

import (
	"fecim-lattice-tools/workbench/moduleapi/httpapi"
	"fmt"
	"os"
)

func main() {
	data, err := httpapi.New(nil).JSON()
	if err != nil {
		fmt.Fprintln(os.Stderr, "schema generation failed")
		os.Exit(1)
	}
	data = append(data, '\n')
	if len(os.Args) == 2 {
		if err := os.WriteFile(os.Args[1], data, 0644); err != nil {
			fmt.Fprintln(os.Stderr, "schema output unavailable")
			os.Exit(1)
		}
	} else {
		_, _ = os.Stdout.Write(data)
	}
}
