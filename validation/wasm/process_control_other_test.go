//go:build !linux && !js

package wasm_test

import (
	"fmt"
	"os/exec"
)

func protectProcessTree(cmd *exec.Cmd) error {
	return fmt.Errorf("the explicit Firefox parity gate currently requires Linux process-tree ownership")
}
