// pkg/validation/circuit_image.go
// Circuit visualization using Yosys show command and OpenROAD save_image
// Generates schematic SVG and layout PNG from EDA files

package validation

import (
	"fmt"
	"os"
	"path/filepath"

	"fecim-lattice-tools/module6-eda/pkg/openlane"
)

// CircuitImageResult contains the result of circuit image generation
type CircuitImageResult struct {
	Success   bool
	ImagePath string
	RawOutput string
	Error     string
}

// GenerateYosysSchematic creates a circuit schematic SVG using Yosys show command
// Requires: Verilog file
// Output: SVG schematic diagram
func GenerateYosysSchematic(verilogPath string, outputPrefix string, topModule string, manager *openlane.Manager, config *openlane.Config) (*CircuitImageResult, error) {
	result := &CircuitImageResult{
		Success:   false,
		ImagePath: outputPrefix + ".svg",
	}

	// Check if Verilog file exists
	if _, err := os.Stat(verilogPath); os.IsNotExist(err) {
		result.Error = fmt.Sprintf("Verilog file not found: %s", verilogPath)
		return result, nil
	}

	// Check mode
	mode := manager.DetectMode()
	if mode == openlane.ModeNone {
		result.Error = "Yosys not available (install Docker with OpenLane image or native Yosys)"
		return result, nil
	}

	workDir := filepath.Dir(verilogPath)
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get absolute path: %v", err)
		return result, nil
	}

	// Yosys command to generate schematic
	// show -format svg -prefix <name> -viewer none
	verilogName := filepath.Base(verilogPath)
	outputName := filepath.Base(outputPrefix)

	yosysCmd := fmt.Sprintf(
		"read_verilog %s; hierarchy -check -top %s; show -format svg -prefix %s -viewer none",
		verilogName, topModule, outputName,
	)

	runner := openlane.NewRunner(manager, config)
	runResult, err := runner.RunYosys(yosysCmd, absWorkDir)

	if runResult != nil {
		result.RawOutput = runResult.Stdout + "\n" + runResult.Stderr
	}

	if err != nil {
		result.Error = fmt.Sprintf("Yosys execution failed: %v", err)
		return result, nil
	}

	// Check if output file was created (Yosys adds .svg extension)
	expectedOutput := filepath.Join(absWorkDir, outputName+".svg")
	if _, err := os.Stat(expectedOutput); os.IsNotExist(err) {
		result.Error = "Yosys did not produce SVG schematic"
		return result, nil
	}

	result.Success = true
	result.ImagePath = expectedOutput
	return result, nil
}

// openroadImageScript is the TCL script for OpenROAD image export
const openroadImageScript = `# save_layout_image.tcl - OpenROAD layout image export
# Environment: CELL_LEF, DEF_FILE, OUTPUT_PNG

puts "=== OpenROAD Layout Image Export ==="

# Read LEF and DEF
read_lef $env(CELL_LEF)
read_def $env(DEF_FILE)

puts "Design loaded, saving image..."

# Save layout image
save_image $env(OUTPUT_PNG)

puts "=== Image Export Complete ==="
exit
`

// GenerateOpenROADImage creates a layout PNG using OpenROAD save_image command
// Requires: LEF and DEF files
// Output: PNG layout image
func GenerateOpenROADImage(defPath string, lefPath string, outputPath string, manager *openlane.Manager, config *openlane.Config) (*CircuitImageResult, error) {
	result := &CircuitImageResult{
		Success:   false,
		ImagePath: outputPath,
	}

	// Check if files exist
	if _, err := os.Stat(defPath); os.IsNotExist(err) {
		result.Error = fmt.Sprintf("DEF file not found: %s", defPath)
		return result, nil
	}
	if _, err := os.Stat(lefPath); os.IsNotExist(err) {
		result.Error = fmt.Sprintf("LEF file not found: %s", lefPath)
		return result, nil
	}

	// Check mode
	mode := manager.DetectMode()
	if mode == openlane.ModeNone {
		result.Error = "OpenROAD not available (install Docker with OpenLane image or native OpenROAD)"
		return result, nil
	}

	workDir := filepath.Dir(defPath)
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get absolute path: %v", err)
		return result, nil
	}

	// Write TCL script
	scriptPath := filepath.Join(absWorkDir, "save_layout_image.tcl")
	if err := os.WriteFile(scriptPath, []byte(openroadImageScript), 0644); err != nil {
		result.Error = fmt.Sprintf("failed to write TCL script: %v", err)
		return result, nil
	}
	defer os.Remove(scriptPath)

	// Copy LEF to work directory if needed
	lefName := filepath.Base(lefPath)
	lefDst := filepath.Join(absWorkDir, lefName)
	if lefPath != lefDst {
		if lefData, err := os.ReadFile(lefPath); err == nil {
			os.WriteFile(lefDst, lefData, 0644)
			defer os.Remove(lefDst)
		}
	}

	// Set up environment variables
	var envVars map[string]string
	outputName := filepath.Base(outputPath)
	if mode == openlane.ModeDocker {
		envVars = map[string]string{
			"DEF_FILE":   "/design/" + filepath.Base(defPath),
			"CELL_LEF":   "/design/" + lefName,
			"OUTPUT_PNG": "/design/" + outputName,
		}
	} else {
		envVars = map[string]string{
			"DEF_FILE":   filepath.Join(absWorkDir, filepath.Base(defPath)),
			"CELL_LEF":   lefDst,
			"OUTPUT_PNG": filepath.Join(absWorkDir, outputName),
		}
	}

	// Run OpenROAD
	runner := openlane.NewRunner(manager, config)
	runResult, err := runner.RunOpenROAD("save_layout_image.tcl", absWorkDir, envVars)

	if runResult != nil {
		result.RawOutput = runResult.Stdout + "\n" + runResult.Stderr
	}

	if err != nil {
		result.Error = fmt.Sprintf("OpenROAD execution failed: %v", err)
		return result, nil
	}

	// Check if output file was created
	expectedOutput := filepath.Join(absWorkDir, outputName)
	if _, err := os.Stat(expectedOutput); os.IsNotExist(err) {
		result.Error = "OpenROAD did not produce layout image"
		return result, nil
	}

	// Move to final output path if different
	if expectedOutput != outputPath {
		os.Rename(expectedOutput, outputPath)
	}

	result.Success = true
	result.ImagePath = outputPath
	return result, nil
}
