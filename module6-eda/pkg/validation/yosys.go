// pkg/validation/yosys.go
package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fecim-lattice-tools/module6-eda/pkg/openlane"
	"fecim-lattice-tools/shared/logging"
)

// ValidateVerilog validates Verilog syntax using Yosys (via Docker or native)
// Returns nil if validation passes, error with details if it fails
func ValidateVerilog(verilogPath string) error {
	logging.GlobalInfo("Running Yosys validation on: %s", verilogPath)

	// Check if file exists
	if _, err := os.Stat(verilogPath); os.IsNotExist(err) {
		return fmt.Errorf("verilog file not found: %s", verilogPath)
	}

	// Get absolute path and working directory
	absPath, err := filepath.Abs(verilogPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}
	workDir := filepath.Dir(absPath)
	fileName := filepath.Base(absPath)

	// Create runner with OpenLane manager
	manager := openlane.NewManager()
	config := openlane.DefaultConfig()
	runner := openlane.NewRunner(manager, config)

	// Build yosys command - paths relative to /design in Docker
	yosysCmd := fmt.Sprintf("read_verilog %s; hierarchy -check", fileName)

	result, err := runner.RunYosys(yosysCmd, workDir)
	if err != nil {
		logging.GlobalError("Yosys validation failed: %v", err)
		output := ""
		if result != nil {
			output = result.Stdout + result.Stderr
		}
		return fmt.Errorf("yosys validation failed:\n%s\n%v", output, err)
	}

	logging.GlobalDebug("Yosys output:\n%s", result.Stdout)

	// Check for errors in output
	outputStr := result.Stdout + result.Stderr
	if strings.Contains(outputStr, "ERROR") {
		logging.GlobalError("Yosys reported internal errors")
		return fmt.Errorf("yosys found errors:\n%s", outputStr)
	}

	logging.GlobalInfo("Yosys validation passed for %s", verilogPath)
	return nil
}

// ValidateVerilogWithCell validates array Verilog with cell library blackbox
// Both files should be in the same directory for Docker volume mounting
func ValidateVerilogWithCell(arrayPath, cellPath string) error {
	logging.GlobalInfo("Running Yosys validation on array: %s with cell: %s", arrayPath, cellPath)

	// Check if files exist
	if _, err := os.Stat(arrayPath); os.IsNotExist(err) {
		return fmt.Errorf("array verilog not found: %s", arrayPath)
	}
	if _, err := os.Stat(cellPath); os.IsNotExist(err) {
		return fmt.Errorf("cell verilog not found: %s", cellPath)
	}

	// Get absolute paths
	absArrayPath, err := filepath.Abs(arrayPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for array: %v", err)
	}
	absCellPath, err := filepath.Abs(cellPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for cell: %v", err)
	}

	// For Docker, we need both files accessible from a common mount point
	// Use the project root as the work directory
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v", err)
	}

	// Make paths relative to workDir for Docker
	relArrayPath, err := filepath.Rel(workDir, absArrayPath)
	if err != nil {
		relArrayPath = absArrayPath
	}
	relCellPath, err := filepath.Rel(workDir, absCellPath)
	if err != nil {
		relCellPath = absCellPath
	}

	// Create runner with OpenLane manager
	manager := openlane.NewManager()
	config := openlane.DefaultConfig()
	runner := openlane.NewRunner(manager, config)

	// Build yosys command with cell as library (blackbox)
	yosysCmd := fmt.Sprintf("read_verilog -lib %s; read_verilog %s; hierarchy -check", relCellPath, relArrayPath)

	result, err := runner.RunYosys(yosysCmd, workDir)
	if err != nil {
		logging.GlobalError("Yosys validation failed: %v", err)
		output := ""
		if result != nil {
			output = result.Stdout + result.Stderr
		}
		return fmt.Errorf("yosys validation failed:\n%s\n%v", output, err)
	}

	logging.GlobalDebug("Yosys output:\n%s", result.Stdout)

	outputStr := result.Stdout + result.Stderr
	if strings.Contains(outputStr, "ERROR") {
		logging.GlobalError("Yosys reported internal errors")
		return fmt.Errorf("yosys found errors:\n%s", outputStr)
	}

	logging.GlobalInfo("Yosys validation passed for %s + %s", arrayPath, cellPath)
	return nil
}
