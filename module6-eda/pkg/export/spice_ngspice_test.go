package export

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
)

// TestM6_SPICE_04_NgspiceCompatibility validates SPICE netlist compatibility with ngspice.
// If ngspice is available: exports 2×2 array, runs .op analysis, checks exit code.
// If ngspice is not available: skips with informative message.
func TestM6_SPICE_04_NgspiceCompatibility_OpAnalysis(t *testing.T) {
	// Check if ngspice is available
	ngspicePath, err := exec.LookPath("ngspice")
	if err != nil {
		t.Skip("M6-SPICE-04: ngspice not found in PATH — skipping external tool validation (install ngspice to enable)")
	}

	t.Logf("M6-SPICE-04: Found ngspice at %s", ngspicePath)

	// Create 2×2 1T1R array
	cells := []compiler.CellAssignment{
		{Row: 0, Col: 0, Conductance: 50.0, Resistance: 20000.0, Level: 0},
		{Row: 0, Col: 1, Conductance: 60.0, Resistance: 16666.7, Level: 5},
		{Row: 1, Col: 0, Conductance: 70.0, Resistance: 14285.7, Level: 10},
		{Row: 1, Col: 1, Conductance: 80.0, Resistance: 12500.0, Level: 15},
	}

	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeStorage,
			Architecture: compiler.Arch1T1R,
		},
		Cells: cells,
		Stats: compiler.DesignStats{TotalCells: 4, ActiveCells: 4},
	}

	// Generate SPICE netlist
	netlist := GenerateSPICE(design, 1.8)

	// Write to temporary file
	tmpDir := t.TempDir()
	netlistPath := filepath.Join(tmpDir, "test_array.sp")
	err = os.WriteFile(netlistPath, []byte(netlist), 0644)
	if err != nil {
		t.Fatalf("M6-SPICE-04: Failed to write netlist to %s: %v", netlistPath, err)
	}

	t.Logf("M6-SPICE-04: Wrote netlist to %s (%d bytes)", netlistPath, len(netlist))

	// Run ngspice in batch mode
	cmd := exec.Command("ngspice", "-b", netlistPath, "-o", filepath.Join(tmpDir, "ngspice.log"))
	output, err := cmd.CombinedOutput()

	// Log output for debugging
	t.Logf("M6-SPICE-04: ngspice output:\n%s", string(output))

	// Check exit code
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Errorf("M6-SPICE-04: ngspice exited with non-zero code: %d\nOutput:\n%s",
				exitErr.ExitCode(), string(output))
		} else {
			t.Fatalf("M6-SPICE-04: Failed to run ngspice: %v", err)
		}
	}

	// Verify output contains expected markers
	outputStr := string(output)
	if !strings.Contains(outputStr, "Circuit:") && !strings.Contains(outputStr, "ngspice") {
		t.Errorf("M6-SPICE-04: ngspice output does not contain expected simulation markers")
	}

	// Check for common SPICE errors
	errorKeywords := []string{"error:", "ERROR:", "fatal:", "FATAL:", "syntax error", "unknown"}
	for _, keyword := range errorKeywords {
		if strings.Contains(strings.ToLower(outputStr), strings.ToLower(keyword)) {
			t.Errorf("M6-SPICE-04: ngspice output contains error keyword '%s'", keyword)
		}
	}

	t.Log("M6-SPICE-04 PASS: ngspice .op analysis completed successfully — netlist is tool-compatible")
}

// TestM6_SPICE_04_NgspiceCompatibility_Passive validates passive array with ngspice
func TestM6_SPICE_04_NgspiceCompatibility_Passive(t *testing.T) {
	// Check if ngspice is available
	if _, err := exec.LookPath("ngspice"); err != nil {
		t.Skip("M6-SPICE-04: ngspice not found — skipping")
	}

	// Create passive (0T1R) array
	cells := []compiler.CellAssignment{
		{Row: 0, Col: 0, Conductance: 50.0, Resistance: 20000.0, Level: 0},
		{Row: 0, Col: 1, Conductance: 60.0, Resistance: 16666.7, Level: 5},
	}

	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeStorage,
			Architecture: compiler.ArchPassive,
		},
		Cells: cells,
		Stats: compiler.DesignStats{TotalCells: 2, ActiveCells: 2},
	}

	netlist := GenerateSPICE(design, 1.8)

	tmpDir := t.TempDir()
	netlistPath := filepath.Join(tmpDir, "passive_array.sp")
	err := os.WriteFile(netlistPath, []byte(netlist), 0644)
	if err != nil {
		t.Fatalf("M6-SPICE-04: Failed to write passive netlist: %v", err)
	}

	cmd := exec.Command("ngspice", "-b", netlistPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Logf("M6-SPICE-04: ngspice output:\n%s", string(output))
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Errorf("M6-SPICE-04: Passive array ngspice failed with exit code %d", exitErr.ExitCode())
		}
	}

	t.Log("M6-SPICE-04 PASS: Passive array ngspice compatibility verified")
}

// TestM6_SPICE_04_NgspiceCompatibility_2T1R validates 2T1R array with ngspice
func TestM6_SPICE_04_NgspiceCompatibility_2T1R(t *testing.T) {
	if _, err := exec.LookPath("ngspice"); err != nil {
		t.Skip("M6-SPICE-04: ngspice not found — skipping")
	}

	cells := []compiler.CellAssignment{
		{Row: 0, Col: 0, Conductance: 50.0, Resistance: 20000.0, Level: 0},
		{Row: 0, Col: 1, Conductance: 60.0, Resistance: 16666.7, Level: 5},
	}

	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeCompute,
			Architecture: compiler.Arch2T1R,
		},
		Cells: cells,
		Stats: compiler.DesignStats{TotalCells: 2, ActiveCells: 2},
	}

	netlist := GenerateSPICE(design, 1.8)

	tmpDir := t.TempDir()
	netlistPath := filepath.Join(tmpDir, "2t1r_array.sp")
	err := os.WriteFile(netlistPath, []byte(netlist), 0644)
	if err != nil {
		t.Fatalf("M6-SPICE-04: Failed to write 2T1R netlist: %v", err)
	}

	cmd := exec.Command("ngspice", "-b", netlistPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Logf("M6-SPICE-04: ngspice output:\n%s", string(output))
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Errorf("M6-SPICE-04: 2T1R array ngspice failed with exit code %d", exitErr.ExitCode())
		}
	}

	t.Log("M6-SPICE-04 PASS: 2T1R array ngspice compatibility verified")
}

// TestM6_SPICE_04_NetlistFileExport validates ExportSPICE writes valid file
func TestM6_SPICE_04_NetlistFileExport(t *testing.T) {
	cells := []compiler.CellAssignment{
		{Row: 0, Col: 0, Conductance: 50.0, Resistance: 20000.0, Level: 0},
	}

	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeStorage,
			Architecture: compiler.Arch1T1R,
		},
		Cells: cells,
		Stats: compiler.DesignStats{TotalCells: 1, ActiveCells: 1},
	}

	tmpDir := t.TempDir()
	netlistPath := filepath.Join(tmpDir, "export_test.sp")

	err := ExportSPICE(design, netlistPath, 1.8)
	if err != nil {
		t.Fatalf("M6-SPICE-04: ExportSPICE failed: %v", err)
	}

	// Verify file exists and is readable
	content, err := os.ReadFile(netlistPath)
	if err != nil {
		t.Fatalf("M6-SPICE-04: Failed to read exported netlist: %v", err)
	}

	// Verify content is non-empty and contains key markers
	if len(content) == 0 {
		t.Fatal("M6-SPICE-04: Exported netlist is empty")
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, ".subckt") {
		t.Error("M6-SPICE-04: Exported netlist missing .subckt directive")
	}
	if !strings.Contains(contentStr, ".end") {
		t.Error("M6-SPICE-04: Exported netlist missing .end directive")
	}

	t.Logf("M6-SPICE-04 PASS: ExportSPICE wrote valid netlist to %s (%d bytes)", netlistPath, len(content))
}
