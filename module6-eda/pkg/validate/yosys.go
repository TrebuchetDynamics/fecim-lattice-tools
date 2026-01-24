// pkg/validate/yosys.go
package validate

import (
	"bytes"
	"fmt"
	"os/exec"
)

// RunYosysCheck executes yosys to validate the verilog file
// Returns the output log and any error encountered
func RunYosysCheck(verilogPath string) (string, error) {
	// Command: yosys -p "read_verilog <file>; hierarchy -check; check"
	cmdStr := fmt.Sprintf("read_verilog %s; hierarchy -check; check", verilogPath)
	
	cmd := exec.Command("yosys", "-p", cmdStr)
	
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	
	output := outBuf.String()
	if err != nil {
		// Append stderr if there was an error
		output += "\nERROR:\n" + errBuf.String()
		return output, fmt.Errorf("yosys validation failed: %w", err)
	}

	return output, nil
}
