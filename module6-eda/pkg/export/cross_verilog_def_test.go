// pkg/export/cross_verilog_def_test.go
// M6-CROSS-02: Cross-format consistency between Verilog and DEF
// Verifies instance counts match across RTL (Verilog) and physical (DEF) formats

package export

import (
	"regexp"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
)

// TestCrossFormat_Verilog_DEF_InstanceCount (M6-CROSS-02)
// Exports Verilog and DEF for the same array configuration
// Counts module instances in Verilog netlist
// Counts COMPONENTS in DEF placement file
// Verifies counts match exactly
func TestCrossFormat_Verilog_DEF_InstanceCount(t *testing.T) {
	// Create test array design
	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeStorage,
			Architecture: compiler.ArchPassive,
			Technology:   "sky130",
			Levels:       32,
			CellPitch:    0.46,
			RowHeight:    2.72,
		},
		Cells: []compiler.CellAssignment{
			{Row: 0, Col: 0, Conductance: 50.0, Resistance: 20000.0, Level: 5},
			{Row: 0, Col: 1, Conductance: 60.0, Resistance: 16666.7, Level: 10},
			{Row: 1, Col: 0, Conductance: 70.0, Resistance: 14285.7, Level: 15},
			{Row: 1, Col: 1, Conductance: 80.0, Resistance: 12500.0, Level: 20},
		},
		Stats: compiler.DesignStats{TotalCells: 4, ActiveCells: 4},
	}

	// Generate Verilog RTL
	verilogCfg := DefaultVerilogConfig()
	verilogNetlist := GenerateVerilog(design, verilogCfg)
	if len(verilogNetlist) == 0 {
		t.Fatal("Verilog netlist generation failed")
	}

	// Generate DEF placement
	defCfg := DEFConfigFrom(design)
	defLayout := GenerateDEF(design, defCfg)
	if len(defLayout) == 0 {
		t.Fatal("DEF layout generation failed")
	}

	// Count instances in Verilog
	verilogCount := countVerilogInstances(t, verilogNetlist)

	// Count components in DEF
	defCount := countDEFComponents(t, defLayout)

	t.Logf("M6-CROSS-02: Verilog instances=%d, DEF components=%d", verilogCount, defCount)

	// Verify counts match exactly
	if verilogCount != defCount {
		t.Errorf("Instance count mismatch: Verilog=%d, DEF=%d", verilogCount, defCount)
	}

	// Also verify against expected count from design
	expectedCount := len(design.Cells)
	if verilogCount != expectedCount {
		t.Errorf("Verilog count mismatch: got %d, expected %d", verilogCount, expectedCount)
	}
	if defCount != expectedCount {
		t.Errorf("DEF count mismatch: got %d, expected %d", defCount, expectedCount)
	}
}

// TestCrossFormat_Verilog_DEF_1T1R_InstanceCount (M6-CROSS-02)
// Same test for 1T1R architecture
func TestCrossFormat_Verilog_DEF_1T1R_InstanceCount(t *testing.T) {
	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeMemory,
			Architecture: compiler.Arch1T1R,
			Technology:   "sky130",
			Levels:       64,
			CellPitch:    0.92,
			RowHeight:    3.40,
		},
		Cells: []compiler.CellAssignment{
			{Row: 0, Col: 0, Conductance: 50.0, Resistance: 20000.0, Level: 10},
			{Row: 0, Col: 1, Conductance: 60.0, Resistance: 16666.7, Level: 20},
			{Row: 0, Col: 2, Conductance: 70.0, Resistance: 14285.7, Level: 30},
		},
		Stats: compiler.DesignStats{TotalCells: 3, ActiveCells: 3},
	}

	verilogCfg := DefaultVerilogConfig()
	verilogNetlist := GenerateVerilog(design, verilogCfg)
	if len(verilogNetlist) == 0 {
		t.Fatal("Verilog netlist generation failed")
	}

	defCfg := DEFConfigFrom(design)
	defLayout := GenerateDEF(design, defCfg)
	if len(defLayout) == 0 {
		t.Fatal("DEF layout generation failed")
	}

	verilogCount := countVerilogInstances(t, verilogNetlist)
	defCount := countDEFComponents(t, defLayout)

	t.Logf("M6-CROSS-02 (1T1R): Verilog instances=%d, DEF components=%d", verilogCount, defCount)

	if verilogCount != defCount {
		t.Errorf("1T1R instance count mismatch: Verilog=%d, DEF=%d", verilogCount, defCount)
	}

	expectedCount := len(design.Cells)
	if verilogCount != expectedCount || defCount != expectedCount {
		t.Errorf("1T1R count mismatch: Verilog=%d, DEF=%d, expected=%d",
			verilogCount, defCount, expectedCount)
	}
}

// TestCrossFormat_Verilog_DEF_2T1R_InstanceCount (M6-CROSS-02)
// Same test for 2T1R architecture
func TestCrossFormat_Verilog_DEF_2T1R_InstanceCount(t *testing.T) {
	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeCompute,
			Architecture: compiler.Arch2T1R,
			Technology:   "sky130",
			Levels:       128,
			CellPitch:    1.38,
			RowHeight:    3.80,
		},
		Cells: []compiler.CellAssignment{
			{Row: 0, Col: 0, Conductance: 75.0, Resistance: 13333.3, Level: 32},
			{Row: 0, Col: 1, Conductance: 85.0, Resistance: 11764.7, Level: 64},
		},
		Stats: compiler.DesignStats{TotalCells: 2, ActiveCells: 2},
	}

	verilogCfg := DefaultVerilogConfig()
	verilogNetlist := GenerateVerilog(design, verilogCfg)
	if len(verilogNetlist) == 0 {
		t.Fatal("Verilog netlist generation failed")
	}

	defCfg := DEFConfigFrom(design)
	defLayout := GenerateDEF(design, defCfg)
	if len(defLayout) == 0 {
		t.Fatal("DEF layout generation failed")
	}

	verilogCount := countVerilogInstances(t, verilogNetlist)
	defCount := countDEFComponents(t, defLayout)

	t.Logf("M6-CROSS-02 (2T1R): Verilog instances=%d, DEF components=%d", verilogCount, defCount)

	if verilogCount != defCount {
		t.Errorf("2T1R instance count mismatch: Verilog=%d, DEF=%d", verilogCount, defCount)
	}

	expectedCount := len(design.Cells)
	if verilogCount != expectedCount || defCount != expectedCount {
		t.Errorf("2T1R count mismatch: Verilog=%d, DEF=%d, expected=%d",
			verilogCount, defCount, expectedCount)
	}
}

// TestCrossFormat_Verilog_DEF_LargeArray (M6-CROSS-02)
// Test with larger array to ensure consistency at scale
func TestCrossFormat_Verilog_DEF_LargeArray(t *testing.T) {
	// Create 8x8 array (64 cells)
	var cells []compiler.CellAssignment
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			cells = append(cells, compiler.CellAssignment{
				Row:         row,
				Col:         col,
				Conductance: 50.0 + float64(row*col),
				Resistance:  20000.0 - float64(row*col)*100,
				Level:       (row*8 + col) % 32,
			})
		}
	}

	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeCompute,
			Architecture: compiler.ArchPassive,
			Technology:   "sky130",
			Levels:       32,
			CellPitch:    0.46,
			RowHeight:    2.72,
		},
		Cells: cells,
		Stats: compiler.DesignStats{TotalCells: 64, ActiveCells: 64},
	}

	verilogCfg := DefaultVerilogConfig()
	verilogNetlist := GenerateVerilog(design, verilogCfg)

	defCfg := DEFConfigFrom(design)
	defLayout := GenerateDEF(design, defCfg)

	verilogCount := countVerilogInstances(t, verilogNetlist)
	defCount := countDEFComponents(t, defLayout)

	t.Logf("M6-CROSS-02 (8x8 array): Verilog instances=%d, DEF components=%d", verilogCount, defCount)

	if verilogCount != defCount {
		t.Errorf("Large array instance count mismatch: Verilog=%d, DEF=%d", verilogCount, defCount)
	}

	if verilogCount != 64 || defCount != 64 {
		t.Errorf("Large array count error: Verilog=%d, DEF=%d, expected=64",
			verilogCount, defCount)
	}
}

// TestCrossFormat_Verilog_DEF_InstanceNames (M6-CROSS-02)
// Verify instance naming convention matches between Verilog and DEF
func TestCrossFormat_Verilog_DEF_InstanceNames(t *testing.T) {
	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeStorage,
			Architecture: compiler.ArchPassive,
			Technology:   "sky130",
			CellPitch:    0.46,
			RowHeight:    2.72,
		},
		Cells: []compiler.CellAssignment{
			{Row: 0, Col: 0, Conductance: 50.0, Level: 5},
			{Row: 2, Col: 3, Conductance: 60.0, Level: 10},
		},
		Stats: compiler.DesignStats{TotalCells: 2, ActiveCells: 2},
	}

	verilogCfg := DefaultVerilogConfig()
	verilogNetlist := GenerateVerilog(design, verilogCfg)

	defCfg := DEFConfigFrom(design)
	defLayout := GenerateDEF(design, defCfg)

	// Check for R_0_0 naming convention in both formats
	if !strings.Contains(verilogNetlist, "R_0_0") {
		t.Error("Verilog missing expected instance name R_0_0")
	}
	if !strings.Contains(defLayout, "R_0_0") {
		t.Error("DEF missing expected instance name R_0_0")
	}

	// Check for R_2_3 in both formats
	if !strings.Contains(verilogNetlist, "R_2_3") {
		t.Error("Verilog missing expected instance name R_2_3")
	}
	if !strings.Contains(defLayout, "R_2_3") {
		t.Error("DEF missing expected instance name R_2_3")
	}

	t.Log("M6-CROSS-02: Instance naming convention verified across Verilog and DEF")
}

// countVerilogInstances counts cell instances in Verilog netlist
// Looks for pattern: cellname #(.LEVEL(...)) instancename (
func countVerilogInstances(t *testing.T, verilog string) int {
	t.Helper()

	// Match instance pattern with R_{row}_{col} naming convention
	// Example: "fecim_bit #(.LEVEL(5)) R_0_0 ("
	// Use simple pattern: #(.LEVEL followed by R_\d+_\d+ on same line
	instancePattern := regexp.MustCompile(`#\(\.LEVEL\([^)]+\)\)\s+R_\d+_\d+`)
	matches := instancePattern.FindAllString(verilog, -1)

	count := len(matches)
	t.Logf("Verilog: found %d instances matching #(.LEVEL...) R_{row}_{col} pattern", count)

	return count
}

// countDEFComponents counts COMPONENTS in DEF placement file
// Parses the COMPONENTS section and counts component declarations
func countDEFComponents(t *testing.T, def string) int {
	t.Helper()

	// Find COMPONENTS count from header
	// Format: "COMPONENTS <count> ;"
	headerPattern := regexp.MustCompile(`COMPONENTS\s+(\d+)\s*;`)
	matches := headerPattern.FindStringSubmatch(def)
	if len(matches) < 2 {
		t.Fatal("Failed to find COMPONENTS header in DEF file")
	}

	// Extract count from header
	var headerCount int
	re := regexp.MustCompile(`(\d+)`)
	if numMatches := re.FindStringSubmatch(matches[1]); len(numMatches) > 1 {
		// Parse the count
		n := numMatches[1]
		for _, ch := range n {
			headerCount = headerCount*10 + int(ch-'0')
		}
	}

	// Also count actual component lines for verification
	// Format: "    - R_0_0 fecim_bit + FIXED ( 10000 10000 ) N ;"
	componentPattern := regexp.MustCompile(`\s+-\s+R_\d+_\d+\s+\w+\s+\+\s+FIXED`)
	actualMatches := componentPattern.FindAllString(def, -1)
	actualCount := len(actualMatches)

	t.Logf("DEF: COMPONENTS header=%d, actual component lines=%d", headerCount, actualCount)

	// Verify header count matches actual count
	if headerCount != actualCount {
		t.Errorf("DEF component count mismatch: header=%d, actual=%d", headerCount, actualCount)
	}

	return actualCount
}
