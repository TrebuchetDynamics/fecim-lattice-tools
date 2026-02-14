// Phase 3: Verilog & Digital Export
// M6-VER-01: Verilog Syntax Validation
//
// Tests:
// - Export Verilog for 4×4 array
// - Check basic syntax: module, endmodule, input, output, wire
// - If Yosys available: run `read_verilog`, check exit code
// - If not: skip with message

package export

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
)

// TestM6VER01_BasicVerilogSyntax verifies basic Verilog syntax elements
func TestM6VER01_BasicVerilogSyntax(t *testing.T) {
	// Create 4×4 array design
	weights := [][]float64{
		{0.1, 0.2, 0.3, 0.4},
		{0.5, 0.6, 0.7, 0.8},
		{0.9, 1.0, -0.1, -0.2},
		{-0.3, -0.4, -0.5, -0.6},
	}

	config := compiler.DefaultConfig()
	config.ArrayRows = 8
	config.ArrayCols = 8
	design, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Generate Verilog
	verilog := GenerateVerilogWithDefaults(design)

	// M6-VER-01a: Check module declaration
	if !strings.Contains(verilog, "module fecim_crossbar") {
		t.Error("M6-VER-01a FAIL: Missing 'module' declaration")
	} else {
		t.Log("M6-VER-01a PASS: 'module' declaration present")
	}

	// M6-VER-01b: Check endmodule
	if !strings.Contains(verilog, "endmodule") {
		t.Error("M6-VER-01b FAIL: Missing 'endmodule' statement")
	} else {
		t.Log("M6-VER-01b PASS: 'endmodule' statement present")
	}

	// M6-VER-01c: Check input declarations
	if !strings.Contains(verilog, "input") {
		t.Error("M6-VER-01c FAIL: Missing 'input' declarations")
	} else {
		t.Log("M6-VER-01c PASS: 'input' declarations present")
	}

	// M6-VER-01d: Check wire declarations (either as port type or internal)
	if !strings.Contains(verilog, "wire") {
		t.Error("M6-VER-01d FAIL: Missing 'wire' declarations")
	} else {
		t.Log("M6-VER-01d PASS: 'wire' declarations present")
	}

	// M6-VER-01e: Check inout declarations (BL lines are bidirectional)
	if !strings.Contains(verilog, "inout") {
		t.Error("M6-VER-01e FAIL: Missing 'inout' declarations for bidirectional ports")
	} else {
		t.Log("M6-VER-01e PASS: 'inout' declarations present")
	}

	// M6-VER-01f: Check port list syntax (should have proper formatting)
	if !strings.Contains(verilog, "(") || !strings.Contains(verilog, ")") {
		t.Error("M6-VER-01f FAIL: Missing port list parentheses")
	} else {
		t.Log("M6-VER-01f PASS: Port list syntax present")
	}

	t.Logf("M6-VER-01 Summary: Generated Verilog size: %d bytes", len(verilog))
}

// TestM6VER01_YosysSyntaxCheck validates Verilog using Yosys (if available)
func TestM6VER01_YosysSyntaxCheck(t *testing.T) {
	// Check if Yosys is available
	yosysPath, err := exec.LookPath("yosys")
	if err != nil {
		t.Skip("M6-VER-01-Yosys SKIP: Yosys not found in PATH (install with: apt install yosys)")
		return
	}

	t.Logf("M6-VER-01-Yosys: Found Yosys at %s", yosysPath)

	// Create 4×4 array design
	weights := [][]float64{
		{0.1, 0.2, 0.3, 0.4},
		{0.5, 0.6, 0.7, 0.8},
		{0.9, 1.0, -0.1, -0.2},
		{-0.3, -0.4, -0.5, -0.6},
	}

	config := compiler.DefaultConfig()
	config.ArrayRows = 8
	config.ArrayCols = 8
	design, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Export Verilog to temporary file
	tmpDir := t.TempDir()
	verilogPath := filepath.Join(tmpDir, "test_4x4.v")
	if err := ExportVerilog(design, verilogPath); err != nil {
		t.Fatalf("ExportVerilog failed: %v", err)
	}

	t.Logf("M6-VER-01-Yosys: Exported Verilog to %s", verilogPath)

	// Create Yosys script to read Verilog
	scriptPath := filepath.Join(tmpDir, "yosys_check.ys")
	yosysScript := `# Yosys syntax check script
read_verilog -noassert ` + verilogPath + `
hierarchy -check
`
	if err := os.WriteFile(scriptPath, []byte(yosysScript), 0644); err != nil {
		t.Fatalf("Failed to write Yosys script: %v", err)
	}

	// Run Yosys
	cmd := exec.Command("yosys", "-s", scriptPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("M6-VER-01-Yosys FAIL: Yosys returned error: %v", err)
		t.Logf("Yosys output:\n%s", string(output))
	} else {
		t.Log("M6-VER-01-Yosys PASS: Yosys read_verilog succeeded (exit code 0)")
		// Check for warnings in output
		if strings.Contains(string(output), "Warning") {
			t.Logf("M6-VER-01-Yosys INFO: Yosys warnings detected:\n%s", string(output))
		}
	}
}

// TestM6VER01_4x4ArrayDimensions verifies correct dimensions in 4×4 export
func TestM6VER01_4x4ArrayDimensions(t *testing.T) {
	// Create 4×4 array design
	weights := [][]float64{
		{0.1, 0.2, 0.3, 0.4},
		{0.5, 0.6, 0.7, 0.8},
		{0.9, 1.0, -0.1, -0.2},
		{-0.3, -0.4, -0.5, -0.6},
	}

	config := compiler.DefaultConfig()
	config.ArrayRows = 8
	config.ArrayCols = 8
	design, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Generate Verilog
	verilog := GenerateVerilogWithDefaults(design)

	// M6-VER-01-DIM-01: Check WL port width (4 rows = [3:0])
	if !strings.Contains(verilog, "input  wire [3:0] WL") {
		t.Error("M6-VER-01-DIM-01 FAIL: Expected 'input  wire [3:0] WL' for 4 rows")
	} else {
		t.Log("M6-VER-01-DIM-01 PASS: WL port is [3:0] for 4 rows")
	}

	// M6-VER-01-DIM-02: Check BL port width (4 cols = [3:0])
	if !strings.Contains(verilog, "inout  wire [3:0] BL") {
		t.Error("M6-VER-01-DIM-02 FAIL: Expected 'inout  wire [3:0] BL' for 4 cols")
	} else {
		t.Log("M6-VER-01-DIM-02 PASS: BL port is [3:0] for 4 cols")
	}

	// M6-VER-01-DIM-03: Check ROWS parameter
	if !strings.Contains(verilog, "parameter ROWS = 4") {
		t.Error("M6-VER-01-DIM-03 FAIL: Expected 'parameter ROWS = 4'")
	} else {
		t.Log("M6-VER-01-DIM-03 PASS: ROWS parameter = 4")
	}

	// M6-VER-01-DIM-04: Check COLS parameter
	if !strings.Contains(verilog, "parameter COLS = 4") {
		t.Error("M6-VER-01-DIM-04 FAIL: Expected 'parameter COLS = 4'")
	} else {
		t.Log("M6-VER-01-DIM-04 PASS: COLS parameter = 4")
	}

	t.Logf("M6-VER-01-DIM Summary: 4×4 array dimensions validated")
}
