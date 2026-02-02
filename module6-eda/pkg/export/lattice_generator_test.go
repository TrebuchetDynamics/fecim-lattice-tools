// pkg/export/lattice_generator_test.go
// Tests for lattice generator functions with cell_{row}_{col} naming convention
package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ============================================================================
// GenerateLatticeVerilog Tests
// ============================================================================

func TestGenerateLatticeVerilog_Basic(t *testing.T) {
	verilog := GenerateLatticeVerilog(4, 4)

	// Check module declaration
	if !strings.Contains(verilog, "module lattice_4x4") {
		t.Error("Missing module declaration for lattice_4x4")
	}

	// Check WL/BL ports
	if !strings.Contains(verilog, "input  wire [3:0] WL") {
		t.Error("Missing WL port declaration")
	}
	if !strings.Contains(verilog, "inout  wire [3:0] BL") {
		t.Error("Missing BL port declaration")
	}

	// Check VPWR/VGND
	if !strings.Contains(verilog, "inout  wire       VPWR") {
		t.Error("Missing VPWR port")
	}
	if !strings.Contains(verilog, "inout  wire       VGND") {
		t.Error("Missing VGND port")
	}

	// Check endmodule
	if !strings.Contains(verilog, "endmodule") {
		t.Error("Missing endmodule")
	}
}

func TestGenerateLatticeVerilog_CellNaming(t *testing.T) {
	verilog := GenerateLatticeVerilog(2, 3)

	// Verify cell_{row}_{col} naming convention (different from R_{row}_{col})
	expectedCells := []string{
		"cell_0_0", "cell_0_1", "cell_0_2",
		"cell_1_0", "cell_1_1", "cell_1_2",
	}

	for _, cell := range expectedCells {
		if !strings.Contains(verilog, cell) {
			t.Errorf("Missing cell instance: %s", cell)
		}
	}

	// Total should be 6 cells
	cellCount := strings.Count(verilog, "fecim_bit cell_")
	if cellCount != 6 {
		t.Errorf("Expected 6 cell instances, got %d", cellCount)
	}
}

func TestGenerateLatticeVerilog_Connections(t *testing.T) {
	verilog := GenerateLatticeVerilog(2, 2)

	// Check that cells connect to correct WL
	if !strings.Contains(verilog, ".WL  (WL[0])") {
		t.Error("Missing WL[0] connection")
	}
	if !strings.Contains(verilog, ".WL  (WL[1])") {
		t.Error("Missing WL[1] connection")
	}

	// Check BL connections
	if !strings.Contains(verilog, ".BL  (BL[0])") {
		t.Error("Missing BL[0] connection")
	}
	if !strings.Contains(verilog, ".BL  (BL[1])") {
		t.Error("Missing BL[1] connection")
	}
}

func TestGenerateLatticeVerilog_LargeArray(t *testing.T) {
	verilog := GenerateLatticeVerilog(16, 32)

	// Check port widths
	if !strings.Contains(verilog, "input  wire [15:0] WL") {
		t.Error("WL should be [15:0] for 16 rows")
	}
	if !strings.Contains(verilog, "inout  wire [31:0] BL") {
		t.Error("BL should be [31:0] for 32 cols")
	}

	// Check corner cells
	if !strings.Contains(verilog, "cell_0_0") {
		t.Error("Missing corner cell cell_0_0")
	}
	if !strings.Contains(verilog, "cell_15_31") {
		t.Error("Missing corner cell cell_15_31")
	}
}

// ============================================================================
// GenerateLatticeDEF Tests
// ============================================================================

func TestGenerateLatticeDEF_Basic(t *testing.T) {
	def := GenerateLatticeDEF(4, 4)

	// Check VERSION
	if !strings.Contains(def, "VERSION 5.8") {
		t.Error("Missing VERSION 5.8")
	}

	// Check DESIGN name
	if !strings.Contains(def, "DESIGN lattice_4x4") {
		t.Error("Missing DESIGN lattice_4x4")
	}

	// Check UNITS
	if !strings.Contains(def, "UNITS DISTANCE MICRONS 1000") {
		t.Error("Missing UNITS declaration")
	}

	// Check END DESIGN
	if !strings.Contains(def, "END DESIGN") {
		t.Error("Missing END DESIGN")
	}
}

func TestGenerateLatticeDEF_Components(t *testing.T) {
	def := GenerateLatticeDEF(3, 3)

	// Should have 9 components
	if !strings.Contains(def, "COMPONENTS 9") {
		t.Error("Should declare COMPONENTS 9")
	}

	// Check cell naming matches Verilog (cell_{row}_{col})
	expectedCells := []string{
		"cell_0_0", "cell_0_1", "cell_0_2",
		"cell_1_0", "cell_1_1", "cell_1_2",
		"cell_2_0", "cell_2_1", "cell_2_2",
	}

	for _, cell := range expectedCells {
		if !strings.Contains(def, cell) {
			t.Errorf("Missing component: %s", cell)
		}
	}

	// Check FIXED placement
	if !strings.Contains(def, "+ FIXED (") {
		t.Error("Cells should have FIXED placement")
	}
}

func TestGenerateLatticeDEF_DIEAREA(t *testing.T) {
	def := GenerateLatticeDEF(4, 4)

	// Check DIEAREA exists
	if !strings.Contains(def, "DIEAREA") {
		t.Error("Missing DIEAREA")
	}

	// DIEAREA should start at origin
	if !strings.Contains(def, "DIEAREA ( 0 0 )") {
		t.Error("DIEAREA should start at ( 0 0 )")
	}
}

func TestGenerateLatticeDEF_Pins(t *testing.T) {
	def := GenerateLatticeDEF(2, 2)

	// Pins = rows + cols + 2 (VPWR + VGND) = 2 + 2 + 2 = 6
	if !strings.Contains(def, "PINS 6") {
		t.Error("Should declare PINS 6")
	}

	// Check WL pins
	if !strings.Contains(def, "WL[0]") || !strings.Contains(def, "WL[1]") {
		t.Error("Missing WL pins")
	}

	// Check BL pins
	if !strings.Contains(def, "BL[0]") || !strings.Contains(def, "BL[1]") {
		t.Error("Missing BL pins")
	}

	// Check power pins
	if !strings.Contains(def, "- VPWR") || !strings.Contains(def, "- VGND") {
		t.Error("Missing power pins")
	}

	// Check direction
	if !strings.Contains(def, "DIRECTION INPUT") {
		t.Error("WL should be DIRECTION INPUT")
	}
	if !strings.Contains(def, "DIRECTION INOUT") {
		t.Error("BL should be DIRECTION INOUT")
	}
	if !strings.Contains(def, "USE POWER") || !strings.Contains(def, "USE GROUND") {
		t.Error("Power pins should have USE POWER/GROUND")
	}
}

func TestGenerateLatticeDEF_Nets(t *testing.T) {
	def := GenerateLatticeDEF(2, 2)

	// Nets = rows + cols + 2 = 6
	if !strings.Contains(def, "NETS 6") {
		t.Error("Should declare NETS 6")
	}

	// Check WL nets connect to cells
	if !strings.Contains(def, "- WL[0]") {
		t.Error("Missing WL[0] net")
	}
	if !strings.Contains(def, "( cell_0_0 WL )") {
		t.Error("WL[0] net should connect to cell_0_0")
	}

	// Check BL nets
	if !strings.Contains(def, "- BL[0]") {
		t.Error("Missing BL[0] net")
	}

	// Check power nets
	if !strings.Contains(def, "- VPWR") {
		t.Error("Missing VPWR net")
	}
}

func TestGenerateLatticeDEF_CellPlacement(t *testing.T) {
	def := GenerateLatticeDEF(2, 2)

	// Cell dimensions: 0.46um width, 2.72um height
	// Origin: 10um, 10um
	// cell_0_0 should be at (10000, 10000) in DBU

	if !strings.Contains(def, "- cell_0_0 fecim_bit + FIXED ( 10000 10000 )") {
		t.Error("cell_0_0 should be at (10000, 10000)")
	}

	// cell_0_1 should be at (10460, 10000) - 0.46um = 460 DBU offset
	if !strings.Contains(def, "- cell_0_1 fecim_bit + FIXED ( 10460 10000 )") {
		t.Error("cell_0_1 should be at (10460, 10000)")
	}

	// cell_1_0 should be at (10000, 12720) - 2.72um = 2720 DBU offset
	if !strings.Contains(def, "- cell_1_0 fecim_bit + FIXED ( 10000 12720 )") {
		t.Error("cell_1_0 should be at (10000, 12720)")
	}
}

// ============================================================================
// Cross-Validation Tests
// ============================================================================

func TestLatticeVerilogDEF_InstanceNamesMatch(t *testing.T) {
	rows, cols := 4, 5

	verilog := GenerateLatticeVerilog(rows, cols)
	def := GenerateLatticeDEF(rows, cols)

	// Extract instance names from both
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			instanceName := strings.Join([]string{"cell_", itoa(row), "_", itoa(col)}, "")

			if !strings.Contains(verilog, instanceName) {
				t.Errorf("Verilog missing instance: %s", instanceName)
			}
			if !strings.Contains(def, instanceName) {
				t.Errorf("DEF missing instance: %s", instanceName)
			}
		}
	}
}

func TestLatticeVerilogDEF_ModuleNameMatch(t *testing.T) {
	sizes := []struct{ rows, cols int }{
		{4, 4},
		{8, 16},
		{32, 32},
	}

	for _, size := range sizes {
		verilog := GenerateLatticeVerilog(size.rows, size.cols)
		def := GenerateLatticeDEF(size.rows, size.cols)

		expectedModuleName := "lattice_" + itoa(size.rows) + "x" + itoa(size.cols)

		if !strings.Contains(verilog, "module "+expectedModuleName) {
			t.Errorf("Verilog module name should be %s", expectedModuleName)
		}
		if !strings.Contains(def, "DESIGN "+expectedModuleName) {
			t.Errorf("DEF design name should be %s", expectedModuleName)
		}
	}
}

// ============================================================================
// File I/O Tests
// ============================================================================

func TestWriteLatticeVerilog(t *testing.T) {
	tmpDir := t.TempDir()

	filename, err := WriteLatticeVerilog(4, 4, tmpDir)
	if err != nil {
		t.Fatalf("WriteLatticeVerilog failed: %v", err)
	}

	expectedFilename := filepath.Join(tmpDir, "lattice_4x4.v")
	if filename != expectedFilename {
		t.Errorf("Expected filename %s, got %s", expectedFilename, filename)
	}

	// Verify file exists and has content
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Output file is empty")
	}
	if !strings.Contains(string(content), "module lattice_4x4") {
		t.Error("Output file missing module declaration")
	}
}

func TestWriteLatticeDEF(t *testing.T) {
	tmpDir := t.TempDir()

	filename, err := WriteLatticeDEF(4, 4, tmpDir)
	if err != nil {
		t.Fatalf("WriteLatticeDEF failed: %v", err)
	}

	expectedFilename := filepath.Join(tmpDir, "lattice_4x4.def")
	if filename != expectedFilename {
		t.Errorf("Expected filename %s, got %s", expectedFilename, filename)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if !strings.Contains(string(content), "DESIGN lattice_4x4") {
		t.Error("Output file missing DESIGN declaration")
	}
}

func TestGenerateLattice(t *testing.T) {
	tmpDir := t.TempDir()

	err := GenerateLattice(8, 8, tmpDir)
	if err != nil {
		t.Fatalf("GenerateLattice failed: %v", err)
	}

	// Check both files exist
	verilogPath := filepath.Join(tmpDir, "lattice_8x8.v")
	defPath := filepath.Join(tmpDir, "lattice_8x8.def")

	if _, err := os.Stat(verilogPath); os.IsNotExist(err) {
		t.Error("Verilog file not created")
	}
	if _, err := os.Stat(defPath); os.IsNotExist(err) {
		t.Error("DEF file not created")
	}
}

func TestGenerateLattice_InvalidDirectory(t *testing.T) {
	err := GenerateLattice(4, 4, "/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("Expected error for invalid directory")
	}
}

// ============================================================================
// Naming Convention Differentiation Test
// ============================================================================

func TestNamingConvention_DifferentFromMainGenerator(t *testing.T) {
	// Lattice generator uses cell_{row}_{col}
	latticeVerilog := GenerateLatticeVerilog(2, 2)

	// Should NOT use R_{row}_{col} naming
	if strings.Contains(latticeVerilog, "R_0_0") {
		t.Error("Lattice generator should use cell_{row}_{col}, not R_{row}_{col}")
	}

	// Should use cell_{row}_{col} naming
	if !strings.Contains(latticeVerilog, "cell_0_0") {
		t.Error("Lattice generator should use cell_{row}_{col} naming")
	}
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestGenerateLatticeVerilog_SingleCell(t *testing.T) {
	verilog := GenerateLatticeVerilog(1, 1)

	if !strings.Contains(verilog, "module lattice_1x1") {
		t.Error("Missing module declaration for 1x1 array")
	}
	if !strings.Contains(verilog, "[0:0] WL") {
		t.Error("WL should be [0:0] for single row")
	}
	if !strings.Contains(verilog, "cell_0_0") {
		t.Error("Missing single cell instance")
	}
}

func TestGenerateLatticeDEF_SingleCell(t *testing.T) {
	def := GenerateLatticeDEF(1, 1)

	if !strings.Contains(def, "COMPONENTS 1") {
		t.Error("Should have COMPONENTS 1")
	}
	if !strings.Contains(def, "PINS 4") {
		t.Error("Should have PINS 4 (WL[0], BL[0], VPWR, VGND)")
	}
}

func TestGenerateLatticeVerilog_RectangularArray(t *testing.T) {
	// Test non-square array
	verilog := GenerateLatticeVerilog(8, 4)

	if !strings.Contains(verilog, "[7:0] WL") {
		t.Error("WL should be [7:0] for 8 rows")
	}
	if !strings.Contains(verilog, "[3:0] BL") {
		t.Error("BL should be [3:0] for 4 cols")
	}

	// Check corner cells exist
	if !strings.Contains(verilog, "cell_7_3") {
		t.Error("Missing corner cell cell_7_3")
	}
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkGenerateLatticeVerilog_Small(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateLatticeVerilog(8, 8)
	}
}

func BenchmarkGenerateLatticeVerilog_Large(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateLatticeVerilog(64, 64)
	}
}

func BenchmarkGenerateLatticeDEF_Small(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateLatticeDEF(8, 8)
	}
}

func BenchmarkGenerateLatticeDEF_Large(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateLatticeDEF(64, 64)
	}
}
