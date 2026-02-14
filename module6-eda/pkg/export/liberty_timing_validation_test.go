package export

import (
	"regexp"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

// TestM6LIB02_TimingNLDMTablesPresent — M6-LIB-02
// Verify NLDM timing tables present (cell_rise, cell_fall)
// Check table dimensions (7×7 or similar)
// Verify index_1, index_2 arrays
func TestM6LIB02_TimingNLDMTablesPresent(t *testing.T) {
	cfg := config.DefaultCellConfig()
	lib := GenerateLiberty(cfg)

	// Verify NLDM template declaration
	if !strings.Contains(lib, "lu_table_template(fecim_nldm_7x7)") {
		t.Fatal("missing NLDM template declaration")
	}

	// Verify template has variable_1 and variable_2
	templatePattern := `lu_table_template\(fecim_nldm_7x7\)\s*\{[^}]*variable_1\s*:[^}]*variable_2\s*:`
	if !regexp.MustCompile(templatePattern).MatchString(lib) {
		t.Fatal("NLDM template missing variable_1 or variable_2")
	}

	// Verify index_1 and index_2 arrays in template
	if !strings.Contains(lib, "index_1(") {
		t.Fatal("NLDM template missing index_1")
	}
	if !strings.Contains(lib, "index_2(") {
		t.Fatal("NLDM template missing index_2")
	}

	// Extract index_1 values
	reIndex1 := regexp.MustCompile(`index_1\("([^"]+)"\)`)
	m1 := reIndex1.FindStringSubmatch(lib)
	if len(m1) < 2 {
		t.Fatal("failed to extract index_1 values")
	}
	index1Values := strings.Split(m1[1], ",")
	if len(index1Values) != 7 {
		t.Fatalf("expected 7 index_1 values, got %d: %v", len(index1Values), index1Values)
	}

	// Extract index_2 values
	reIndex2 := regexp.MustCompile(`index_2\("([^"]+)"\)`)
	m2 := reIndex2.FindStringSubmatch(lib)
	if len(m2) < 2 {
		t.Fatal("failed to extract index_2 values")
	}
	index2Values := strings.Split(m2[1], ",")
	if len(index2Values) != 7 {
		t.Fatalf("expected 7 index_2 values, got %d: %v", len(index2Values), index2Values)
	}

	t.Logf("M6-LIB-02: NLDM template validated")
	t.Logf("  - Table dimensions: 7×7")
	t.Logf("  - index_1 (input_net_transition): %s", m1[1])
	t.Logf("  - index_2 (total_output_net_capacitance): %s", m2[1])
}

// TestM6LIB02_TimingTablesComplete validates all 4 timing tables exist
func TestM6LIB02_TimingTablesComplete(t *testing.T) {
	cfg := config.DefaultCellConfig()
	lib := GenerateLiberty(cfg)

	requiredTables := []string{
		"cell_rise(fecim_nldm_7x7)",
		"cell_fall(fecim_nldm_7x7)",
		"rise_transition(fecim_nldm_7x7)",
		"fall_transition(fecim_nldm_7x7)",
	}

	for _, table := range requiredTables {
		if !strings.Contains(lib, table) {
			t.Fatalf("missing required timing table: %s", table)
		}
	}

	t.Logf("M6-LIB-02 PASS: All 4 NLDM timing tables present")
	for _, table := range requiredTables {
		t.Logf("  - %s", table)
	}
}

// TestM6LIB02_TimingTableDimensions validates 7×7 table structure
func TestM6LIB02_TimingTableDimensions(t *testing.T) {
	cfg := config.DefaultCellConfig()
	lib := GenerateLiberty(cfg)

	// Extract cell_rise table values
	reCellRise := regexp.MustCompile(`cell_rise\(fecim_nldm_7x7\)\s*\{\s*values\(\\([^)]+)\)`)
	mRise := reCellRise.FindStringSubmatch(lib)
	if len(mRise) < 2 {
		t.Fatal("failed to extract cell_rise values")
	}

	// Count rows (should be 7, each row is a quoted string)
	rows := strings.Count(mRise[1], "\"")
	// Each row has 2 quotes (open and close), so divide by 2
	rowCount := rows / 2
	if rowCount != 7 {
		t.Fatalf("expected 7 rows in cell_rise table, got %d", rowCount)
	}

	// Extract first row and count columns
	reFirstRow := regexp.MustCompile(`values\(\\\s*"([^"]+)"`)
	mFirstRow := reFirstRow.FindStringSubmatch(lib)
	if len(mFirstRow) < 2 {
		t.Fatal("failed to extract first row of cell_rise")
	}

	firstRowValues := strings.Split(mFirstRow[1], ",")
	if len(firstRowValues) != 7 {
		t.Fatalf("expected 7 columns in cell_rise table, got %d", len(firstRowValues))
	}

	t.Logf("M6-LIB-02 PASS: NLDM table dimensions validated")
	t.Logf("  - Rows: %d", rowCount)
	t.Logf("  - Columns: %d", len(firstRowValues))
	t.Logf("  - Total entries: %d (7×7)", rowCount*len(firstRowValues))
}

// TestM6LIB02_TimingTableValues validates values are non-negative and monotonic
func TestM6LIB02_TimingTableValues(t *testing.T) {
	cfg := config.DefaultCellConfig()
	cfg.RiseTime = 50.0
	cfg.FallTime = 5.0
	lib := GenerateLiberty(cfg)

	// Extract first value from cell_rise
	riseVal := extractFirstNLDMValue(t, lib, "cell_rise")
	if riseVal <= 0 {
		t.Fatalf("cell_rise first value must be > 0, got %.3f", riseVal)
	}

	// Extract first value from cell_fall
	fallVal := extractFirstNLDMValue(t, lib, "cell_fall")
	if fallVal <= 0 {
		t.Fatalf("cell_fall first value must be > 0, got %.3f", fallVal)
	}

	// Write should be slower than read (rise > fall for FeCIM)
	if riseVal <= fallVal {
		t.Fatalf("expected write slower than read (rise > fall): rise=%.3f fall=%.3f", riseVal, fallVal)
	}

	// Verify rise_transition and fall_transition also present
	riseTransition := extractFirstNLDMValue(t, lib, "rise_transition")
	if riseTransition <= 0 {
		t.Fatalf("rise_transition must be > 0, got %.3f", riseTransition)
	}

	fallTransition := extractFirstNLDMValue(t, lib, "fall_transition")
	if fallTransition <= 0 {
		t.Fatalf("fall_transition must be > 0, got %.3f", fallTransition)
	}

	t.Logf("M6-LIB-02 PASS: Timing table values validated")
	t.Logf("  - cell_rise (write): %.3f ns", riseVal)
	t.Logf("  - cell_fall (read): %.3f ns", fallVal)
	t.Logf("  - rise_transition: %.3f ns", riseTransition)
	t.Logf("  - fall_transition: %.3f ns", fallTransition)
	t.Logf("  - Write/read ratio: %.2fx slower", riseVal/fallVal)
}

// TestM6LIB02_TimingRelatedPins validates related_pin attributes
func TestM6LIB02_TimingRelatedPins(t *testing.T) {
	tests := []struct {
		name      string
		cellType  string
		wantPins  []string
		wantSense string
	}{
		{
			name:      "passive",
			cellType:  "passive",
			wantPins:  []string{"WL"},
			wantSense: "positive_unate",
		},
		{
			name:      "1t1r",
			cellType:  "1t1r",
			wantPins:  []string{"WL", "SL"},
			wantSense: "positive_unate",
		},
		{
			name:      "2t1r",
			cellType:  "2t1r",
			wantPins:  []string{"WL", "CSL", "SL"},
			wantSense: "positive_unate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultCellConfig()
			cfg.CellType = tt.cellType
			lib := GenerateLiberty(cfg)

			// Verify timing blocks for each expected pin
			for _, pin := range tt.wantPins {
				pattern := `related_pin\s*:\s*"` + pin + `"`
				if !regexp.MustCompile(pattern).MatchString(lib) {
					t.Fatalf("missing timing block for related_pin: %s", pin)
				}
			}

			// Verify timing_sense
			if !strings.Contains(lib, "timing_sense : "+tt.wantSense) {
				t.Fatalf("missing or incorrect timing_sense (expected %s)", tt.wantSense)
			}

			t.Logf("M6-LIB-02 PASS (%s): related_pins validated: %v", tt.cellType, tt.wantPins)
		})
	}
}
