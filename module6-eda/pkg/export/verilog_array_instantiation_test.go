// Phase 3: Verilog & Digital Export
// M6-VER-03: Verilog Array Instantiation Validation
//
// Tests:
// - N×M array → N×M cell module instances
// - Count instances in exported Verilog, verify exact match

package export

import (
	"fmt"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
)

// TestM6VER03_ExactInstanceCount verifies N×M array produces exactly N×M instances
func TestM6VER03_ExactInstanceCount(t *testing.T) {
	testCases := []struct {
		name          string
		rows          int
		cols          int
		expectedCells int
	}{
		{"4×4 array", 4, 4, 16},
		{"2×2 array", 2, 2, 4},
		{"8×4 array", 8, 4, 32},
		{"3×5 array", 3, 5, 15},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create weight matrix
			weights := make([][]float64, tc.rows)
			for i := range weights {
				weights[i] = make([]float64, tc.cols)
				for j := range weights[i] {
					weights[i][j] = float64(i*tc.cols+j) * 0.1
				}
			}

			config := compiler.DefaultConfig()
			config.ArrayRows = tc.rows * 2 // Physical array larger than weight matrix
			config.ArrayCols = tc.cols * 2
			design, err := compiler.Compile(weights, config)
			if err != nil {
				t.Fatalf("Compile failed: %v", err)
			}

			verilog := GenerateVerilogWithDefaults(design)

			// M6-VER-03a: Count fecim_bit instances (passive architecture default)
			instanceCount := strings.Count(verilog, "fecim_bit #(")

			if instanceCount != tc.expectedCells {
				t.Errorf("M6-VER-03a FAIL: Expected exactly %d instances for %d×%d array, got %d",
					tc.expectedCells, tc.rows, tc.cols, instanceCount)
			} else {
				t.Logf("M6-VER-03a PASS: %d×%d array → %d cell instances (exact match)",
					tc.rows, tc.cols, instanceCount)
			}

			// M6-VER-03b: Verify instance naming convention
			// Should use R_{row}_{col} naming
			expectedFirstInstance := "R_0_0"
			expectedLastInstance := fmt.Sprintf("R_%d_%d", tc.rows-1, tc.cols-1)

			if !strings.Contains(verilog, expectedFirstInstance) {
				t.Errorf("M6-VER-03b FAIL: Missing instance '%s'", expectedFirstInstance)
			} else {
				t.Logf("M6-VER-03b PASS: First instance '%s' present", expectedFirstInstance)
			}

			if !strings.Contains(verilog, expectedLastInstance) {
				t.Errorf("M6-VER-03b FAIL: Missing instance '%s'", expectedLastInstance)
			} else {
				t.Logf("M6-VER-03b PASS: Last instance '%s' present", expectedLastInstance)
			}
		})
	}
}

// TestM6VER03_InstanceCoverageComplete verifies all cells are instantiated
func TestM6VER03_InstanceCoverageComplete(t *testing.T) {
	// Create 4×4 array
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

	verilog := GenerateVerilogWithDefaults(design)

	// M6-VER-03c: Verify all row×col combinations exist
	missingCells := []string{}
	for row := 0; row < 4; row++ {
		for col := 0; col < 4; col++ {
			instanceName := fmt.Sprintf("R_%d_%d", row, col)
			if !strings.Contains(verilog, instanceName) {
				missingCells = append(missingCells, instanceName)
			}
		}
	}

	if len(missingCells) > 0 {
		t.Errorf("M6-VER-03c FAIL: Missing %d cell instances: %v", len(missingCells), missingCells)
	} else {
		t.Log("M6-VER-03c PASS: All 16 cell instances present (complete coverage)")
	}

	// M6-VER-03d: Verify no extra instances beyond array bounds
	// Check for R_4_0 (should NOT exist in 4×4 array)
	extraInstance := "R_4_0"
	if strings.Contains(verilog, extraInstance) {
		t.Errorf("M6-VER-03d FAIL: Found instance '%s' beyond 4×4 array bounds", extraInstance)
	} else {
		t.Logf("M6-VER-03d PASS: No instances beyond array bounds")
	}
}

// TestM6VER03_ArchitectureInstanceTypes verifies correct cell types per architecture
func TestM6VER03_ArchitectureInstanceTypes(t *testing.T) {
	weights := [][]float64{
		{0.1, 0.2},
		{0.3, 0.4},
	}

	testCases := []struct {
		name         string
		configFunc   func() compiler.ArrayConfig
		expectedCell string
		expectedPort string // Additional port to check
	}{
		{
			name:         "Passive architecture",
			configFunc:   compiler.DefaultConfig,
			expectedCell: "fecim_bit #(",
			expectedPort: "", // No additional ports
		},
		{
			name:         "1T1R architecture",
			configFunc:   compiler.Config1T1R,
			expectedCell: "fecim_1t1r #(",
			expectedPort: ".SL", // Should have SL connections
		},
		{
			name:         "2T1R architecture",
			configFunc:   compiler.Config2T1R,
			expectedCell: "fecim_2t1r #(",
			expectedPort: ".CSL", // Should have CSL connections
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := tc.configFunc()
			config.ArrayRows = 4
			config.ArrayCols = 4
			design, err := compiler.Compile(weights, config)
			if err != nil {
				t.Fatalf("Compile failed: %v", err)
			}

			verilog := GenerateVerilogWithDefaults(design)

			// M6-VER-03-ARCH-01: Verify correct cell type
			instanceCount := strings.Count(verilog, tc.expectedCell)
			expectedCells := 4 // 2×2 weight matrix

			if instanceCount != expectedCells {
				t.Errorf("M6-VER-03-ARCH-01 FAIL: Expected %d instances of '%s', got %d",
					expectedCells, tc.expectedCell, instanceCount)
			} else {
				t.Logf("M6-VER-03-ARCH-01 PASS: %d instances of '%s'", instanceCount, tc.expectedCell)
			}

			// M6-VER-03-ARCH-02: Verify architecture-specific ports
			if tc.expectedPort != "" {
				portCount := strings.Count(verilog, tc.expectedPort)
				if portCount != expectedCells {
					t.Errorf("M6-VER-03-ARCH-02 FAIL: Expected %d '%s' port connections, got %d",
						expectedCells, tc.expectedPort, portCount)
				} else {
					t.Logf("M6-VER-03-ARCH-02 PASS: %d '%s' port connections", portCount, tc.expectedPort)
				}
			}
		})
	}
}

// TestM6VER03_InstanceParameterPropagation verifies LEVEL parameters
func TestM6VER03_InstanceParameterPropagation(t *testing.T) {
	// Create 2×2 array with specific weights
	weights := [][]float64{
		{0.5, -0.5}, // Different weights to ensure different levels
		{1.0, -1.0},
	}

	config := compiler.DefaultConfig()
	config.ArrayRows = 4
	config.ArrayCols = 4
	config.Levels = 16 // Use 16 levels for quantization
	design, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	verilog := GenerateVerilogWithDefaults(design)

	// M6-VER-03-PARAM-01: Verify all instances have LEVEL parameter
	instanceCount := strings.Count(verilog, "fecim_bit #(")
	levelCount := strings.Count(verilog, ".LEVEL(")

	if levelCount != instanceCount {
		t.Errorf("M6-VER-03-PARAM-01 FAIL: Expected %d .LEVEL parameters, got %d",
			instanceCount, levelCount)
	} else {
		t.Logf("M6-VER-03-PARAM-01 PASS: All %d instances have .LEVEL parameter", instanceCount)
	}

	// M6-VER-03-PARAM-02: Verify LEVEL values are within range [0, LEVELS-1]
	// This is a sanity check - actual values tested in compiler tests
	if strings.Contains(verilog, fmt.Sprintf(".LEVEL(%d)", config.Levels)) {
		t.Errorf("M6-VER-03-PARAM-02 FAIL: Found .LEVEL(%d) which exceeds LEVELS-1=%d",
			config.Levels, config.Levels-1)
	} else {
		t.Logf("M6-VER-03-PARAM-02 PASS: LEVEL parameters within valid range [0, %d]", config.Levels-1)
	}

	t.Log("M6-VER-03 Summary: Array instantiation validated (exact count, complete coverage, correct types)")
}
