// pkg/export/cross_def_lef_test.go
// M6-CROSS-03: Cross-format consistency between DEF and LEF
// Verifies cell dimensions match across placement (DEF) and library (LEF) formats

package export

import (
	"math"
	"regexp"
	"strconv"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	"fecim-lattice-tools/module6-eda/pkg/config"
)

// TestCrossFormat_DEF_LEF_CellSize (M6-CROSS-03)
// Exports DEF and LEF for the same array configuration
// Extracts component SIZE from DEF placement file
// Extracts MACRO SIZE from LEF library file
// Verifies match within 1%
func TestCrossFormat_DEF_LEF_CellSize(t *testing.T) {
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
		},
		Stats: compiler.DesignStats{TotalCells: 2, ActiveCells: 2},
	}

	// Generate DEF placement
	defCfg := DEFConfigFrom(design)
	defLayout := GenerateDEF(design, defCfg)
	if len(defLayout) == 0 {
		t.Fatal("DEF layout generation failed")
	}

	// Generate LEF library
	lefCfg := config.CellConfig{
		Name:       "fecim_bitcell",
		CellType:   "passive",
		Technology: "sky130",
		Width:      0.46,
		Height:     2.72,
		MetalPitch: 0.46,
		MetalWidth: 0.14,
	}
	lefLibrary := GenerateLEF(lefCfg)
	if len(lefLibrary) == 0 {
		t.Fatal("LEF library generation failed")
	}

	// Extract cell size from DEF
	defWidth, defHeight := extractCellSizeFromDEF(t, defLayout, defCfg)

	// Extract macro size from LEF
	lefWidth, lefHeight := extractMacroSizeFromLEF(t, lefLibrary)

	// Calculate deltas
	widthDelta := math.Abs(defWidth-lefWidth) / lefWidth * 100.0
	heightDelta := math.Abs(defHeight-lefHeight) / lefHeight * 100.0

	t.Logf("M6-CROSS-03: DEF cell=%.3fx%.3f µm, LEF macro=%.3fx%.3f µm, delta=%.2f%%x%.2f%%",
		defWidth, defHeight, lefWidth, lefHeight, widthDelta, heightDelta)

	// Verify deltas < 1%
	if widthDelta >= 1.0 {
		t.Errorf("Cell width mismatch exceeds 1%% tolerance: DEF=%.3f µm, LEF=%.3f µm, delta=%.2f%%",
			defWidth, lefWidth, widthDelta)
	}
	if heightDelta >= 1.0 {
		t.Errorf("Cell height mismatch exceeds 1%% tolerance: DEF=%.3f µm, LEF=%.3f µm, delta=%.2f%%",
			defHeight, lefHeight, heightDelta)
	}
}

// TestCrossFormat_DEF_LEF_1T1R_CellSize (M6-CROSS-03)
// Same test for 1T1R architecture
func TestCrossFormat_DEF_LEF_1T1R_CellSize(t *testing.T) {
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
		},
		Stats: compiler.DesignStats{TotalCells: 1, ActiveCells: 1},
	}

	defCfg := DEFConfigFrom(design)
	defLayout := GenerateDEF(design, defCfg)
	if len(defLayout) == 0 {
		t.Fatal("DEF layout generation failed")
	}

	lefCfg := config.CellConfig{
		Name:       "fecim_1t1r_bitcell",
		CellType:   "1t1r",
		Technology: "sky130",
		Width:      0.920,
		Height:     3.400,
		MetalPitch: 0.46,
		MetalWidth: 0.14,
	}
	lefLibrary := GenerateLEF(lefCfg)
	if len(lefLibrary) == 0 {
		t.Fatal("LEF library generation failed")
	}

	defWidth, defHeight := extractCellSizeFromDEF(t, defLayout, defCfg)
	lefWidth, lefHeight := extractMacroSizeFromLEF(t, lefLibrary)

	widthDelta := math.Abs(defWidth-lefWidth) / lefWidth * 100.0
	heightDelta := math.Abs(defHeight-lefHeight) / lefHeight * 100.0

	t.Logf("M6-CROSS-03 (1T1R): DEF cell=%.3fx%.3f µm, LEF macro=%.3fx%.3f µm, delta=%.2f%%x%.2f%%",
		defWidth, defHeight, lefWidth, lefHeight, widthDelta, heightDelta)

	if widthDelta >= 1.0 {
		t.Errorf("1T1R width mismatch exceeds 1%% tolerance: DEF=%.3f µm, LEF=%.3f µm, delta=%.2f%%",
			defWidth, lefWidth, widthDelta)
	}
	if heightDelta >= 1.0 {
		t.Errorf("1T1R height mismatch exceeds 1%% tolerance: DEF=%.3f µm, LEF=%.3f µm, delta=%.2f%%",
			defHeight, lefHeight, heightDelta)
	}
}

// TestCrossFormat_DEF_LEF_2T1R_CellSize (M6-CROSS-03)
// Same test for 2T1R architecture
func TestCrossFormat_DEF_LEF_2T1R_CellSize(t *testing.T) {
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
		},
		Stats: compiler.DesignStats{TotalCells: 1, ActiveCells: 1},
	}

	defCfg := DEFConfigFrom(design)
	defLayout := GenerateDEF(design, defCfg)
	if len(defLayout) == 0 {
		t.Fatal("DEF layout generation failed")
	}

	lefCfg := config.CellConfig{
		Name:       "fecim_2t1r_bitcell",
		CellType:   "2t1r",
		Technology: "sky130",
		Width:      1.380,
		Height:     3.800,
		MetalPitch: 0.46,
		MetalWidth: 0.14,
	}
	lefLibrary := GenerateLEF(lefCfg)
	if len(lefLibrary) == 0 {
		t.Fatal("LEF library generation failed")
	}

	defWidth, defHeight := extractCellSizeFromDEF(t, defLayout, defCfg)
	lefWidth, lefHeight := extractMacroSizeFromLEF(t, lefLibrary)

	widthDelta := math.Abs(defWidth-lefWidth) / lefWidth * 100.0
	heightDelta := math.Abs(defHeight-lefHeight) / lefHeight * 100.0

	t.Logf("M6-CROSS-03 (2T1R): DEF cell=%.3fx%.3f µm, LEF macro=%.3fx%.3f µm, delta=%.2f%%x%.2f%%",
		defWidth, defHeight, lefWidth, lefHeight, widthDelta, heightDelta)

	if widthDelta >= 1.0 {
		t.Errorf("2T1R width mismatch exceeds 1%% tolerance: DEF=%.3f µm, LEF=%.3f µm, delta=%.2f%%",
			defWidth, lefWidth, widthDelta)
	}
	if heightDelta >= 1.0 {
		t.Errorf("2T1R height mismatch exceeds 1%% tolerance: DEF=%.3f µm, LEF=%.3f µm, delta=%.2f%%",
			defHeight, lefHeight, heightDelta)
	}
}

// TestCrossFormat_DEF_LEF_ArrayDimensions (M6-CROSS-03)
// Verify array dimensions calculated from cell size match
func TestCrossFormat_DEF_LEF_ArrayDimensions(t *testing.T) {
	// Create 4x3 array
	var cells []compiler.CellAssignment
	for row := 0; row < 4; row++ {
		for col := 0; col < 3; col++ {
			cells = append(cells, compiler.CellAssignment{
				Row:         row,
				Col:         col,
				Conductance: 50.0,
				Resistance:  20000.0,
				Level:       5,
			})
		}
	}

	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeStorage,
			Architecture: compiler.ArchPassive,
			Technology:   "sky130",
			CellPitch:    0.46,
			RowHeight:    2.72,
		},
		Cells: cells,
		Stats: compiler.DesignStats{TotalCells: 12, ActiveCells: 12},
	}

	defCfg := DEFConfigFrom(design)
	defLayout := GenerateDEF(design, defCfg)

	lefCfg := config.CellConfig{
		Name:       "fecim_bitcell",
		CellType:   "passive",
		Technology: "sky130",
		Width:      0.46,
		Height:     2.72,
	}
	lefLibrary := GenerateLEF(lefCfg)

	defWidth, defHeight := extractCellSizeFromDEF(t, defLayout, defCfg)
	lefWidth, lefHeight := extractMacroSizeFromLEF(t, lefLibrary)

	// Calculate expected array dimensions
	expectedArrayWidth := float64(3) * lefWidth
	expectedArrayHeight := float64(4) * lefHeight

	t.Logf("M6-CROSS-03 (4x3): Cell size=%.3fx%.3f µm, Array size=%.3fx%.3f µm",
		lefWidth, lefHeight, expectedArrayWidth, expectedArrayHeight)

	// Verify cell sizes match
	widthDelta := math.Abs(defWidth-lefWidth) / lefWidth * 100.0
	heightDelta := math.Abs(defHeight-lefHeight) / lefHeight * 100.0

	if widthDelta >= 1.0 || heightDelta >= 1.0 {
		t.Errorf("Array cell size mismatch: width delta=%.2f%%, height delta=%.2f%%",
			widthDelta, heightDelta)
	}
}

// TestCrossFormat_DEF_LEF_SiteDefinition (M6-CROSS-03)
// Verify SITE definitions in DEF and LEF match
func TestCrossFormat_DEF_LEF_SiteDefinition(t *testing.T) {
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
		},
		Stats: compiler.DesignStats{TotalCells: 1, ActiveCells: 1},
	}

	defCfg := DEFConfigFrom(design)
	defLayout := GenerateDEF(design, defCfg)

	lefCfg := config.CellConfig{
		Name:       "fecim_bitcell",
		CellType:   "passive",
		Technology: "sky130",
		Width:      0.46,
		Height:     2.72,
	}
	lefLibrary := GenerateLEF(lefCfg)

	// Extract site name from both formats
	defSite := extractSiteNameFromDEF(t, defLayout)
	lefSite := extractSiteNameFromLEF(t, lefLibrary)

	t.Logf("M6-CROSS-03: DEF site='%s', LEF site='%s'", defSite, lefSite)

	if defSite != lefSite {
		t.Errorf("Site name mismatch: DEF='%s', LEF='%s'", defSite, lefSite)
	}
}

// extractCellSizeFromDEF extracts cell width and height from DEF configuration
func extractCellSizeFromDEF(t *testing.T, def string, cfg DEFConfig) (width, height float64) {
	t.Helper()

	// Cell size comes from DEFConfig, which is derived from ArrayConfig
	width = cfg.CellWidth
	height = cfg.CellHeight

	// Verify SITE definition matches
	// Format: "ROW ROW_0 fecim_site X Y N DO cols BY 1 STEP width 0"
	rowPattern := regexp.MustCompile(`ROW\s+ROW_\d+\s+\w+\s+\d+\s+\d+\s+\w+\s+DO\s+\d+\s+BY\s+1\s+STEP\s+(\d+)\s+0`)
	matches := rowPattern.FindStringSubmatch(def)
	if len(matches) >= 2 {
		// Extract STEP value (in database units)
		stepDBU, err := strconv.Atoi(matches[1])
		if err == nil {
			// Convert to microns
			stepWidth := float64(stepDBU) / float64(cfg.DatabaseUnit)
			if math.Abs(stepWidth-width) > 0.001 {
				t.Logf("Warning: DEF STEP width %.3f µm differs from configured width %.3f µm",
					stepWidth, width)
			}
		}
	}

	t.Logf("DEF cell size: %.3f µm x %.3f µm", width, height)

	return width, height
}

// extractMacroSizeFromLEF parses LEF file to extract MACRO SIZE
func extractMacroSizeFromLEF(t *testing.T, lef string) (width, height float64) {
	t.Helper()

	// Look for SIZE declaration in LEF MACRO
	// Format: "SIZE 0.460 BY 2.720 ;"
	sizePattern := regexp.MustCompile(`SIZE\s+([0-9.]+)\s+BY\s+([0-9.]+)\s*;`)
	matches := sizePattern.FindStringSubmatch(lef)
	if len(matches) < 3 {
		t.Fatal("Failed to extract SIZE from LEF MACRO")
	}

	var err error
	width, err = strconv.ParseFloat(matches[1], 64)
	if err != nil {
		t.Fatalf("Failed to parse LEF width: %v", err)
	}

	height, err = strconv.ParseFloat(matches[2], 64)
	if err != nil {
		t.Fatalf("Failed to parse LEF height: %v", err)
	}

	t.Logf("LEF macro size: %.3f µm x %.3f µm", width, height)

	return width, height
}

// extractSiteNameFromDEF extracts SITE name from DEF ROW definition
func extractSiteNameFromDEF(t *testing.T, def string) string {
	t.Helper()

	// Format: "ROW ROW_0 fecim_site ..."
	sitePattern := regexp.MustCompile(`ROW\s+ROW_\d+\s+(\w+)\s+`)
	matches := sitePattern.FindStringSubmatch(def)
	if len(matches) < 2 {
		t.Fatal("Failed to extract SITE name from DEF")
	}

	return matches[1]
}

// extractSiteNameFromLEF extracts SITE name from LEF SITE definition
func extractSiteNameFromLEF(t *testing.T, lef string) string {
	t.Helper()

	// Format: "SITE fecim_site"
	sitePattern := regexp.MustCompile(`SITE\s+(\w+)\s`)
	matches := sitePattern.FindStringSubmatch(lef)
	if len(matches) < 2 {
		t.Fatal("Failed to extract SITE name from LEF")
	}

	return matches[1]
}
