// pkg/validation/yosys.go
package validation

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ValidateVerilog validates Verilog syntax using Yosys
// Returns nil if validation passes, error with details if it fails
func ValidateVerilog(verilogPath string) error {
	// Check if file exists
	if _, err := os.Stat(verilogPath); os.IsNotExist(err) {
		return fmt.Errorf("verilog file not found: %s", verilogPath)
	}
	
	// Check if yosys is installed
	if _, err := exec.LookPath("yosys"); err != nil {
		return fmt.Errorf("yosys not found in PATH - cannot validate (install with: sudo apt install yosys)")
	}
	
	// Run yosys syntax check
	// Use 'read_verilog' to parse and check syntax
	cmd := exec.Command("yosys", "-p", fmt.Sprintf("read_verilog %s; hierarchy -check", verilogPath))
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("yosys validation failed:\n%s", string(output))
	}
	
	// Check for warnings or errors in output
	outputStr := string(output)
	if strings.Contains(outputStr, "ERROR") {
		return fmt.Errorf("yosys found errors:\n%s", outputStr)
	}
	
	return nil
}

// ValidateVerilogWithCell validates array Verilog with cell library blackbox
func ValidateVerilogWithCell(arrayPath, cellPath string) error {
	// Check if files exist
	if _, err := os.Stat(arrayPath); os.IsNotExist(err) {
		return fmt.Errorf("array verilog not found: %s", arrayPath)
	}
	if _, err := os.Stat(cellPath); os.IsNotExist(err) {
		return fmt.Errorf("cell verilog not found: %s", cellPath)
	}
	
	// Check if yosys is installed
	if _, err := exec.LookPath("yosys"); err != nil {
		return fmt.Errorf("yosys not found in PATH")
	}
	
	// Run yosys with cell as library (blackbox)
	cmd := exec.Command("yosys", "-p", 
		fmt.Sprintf("read_verilog -lib %s; read_verilog %s; hierarchy -check", cellPath, arrayPath))
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("yosys validation failed:\n%s", string(output))
	}
	
	outputStr := string(output)
	if strings.Contains(outputStr, "ERROR") {
		return fmt.Errorf("yosys found errors:\n%s", outputStr)
	}
	
	return nil
}
