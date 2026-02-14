package export

import (
	"regexp"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

// TestM6LIB01_LibertySyntaxValidation — M6-LIB-01
// Export Liberty .lib file and parse back: check library, cell, pin, timing, capacitance
// Verify structure correctness
func TestM6LIB01_LibertySyntaxValidation(t *testing.T) {
	cfg := config.DefaultCellConfig()
	lib := GenerateLiberty(cfg)

	// Verify top-level library block
	if !strings.Contains(lib, "library(") {
		t.Fatal("missing library declaration")
	}

	// Extract library name
	reLibrary := regexp.MustCompile(`library\(([^)]+)\)`)
	if !reLibrary.MatchString(lib) {
		t.Fatal("invalid library declaration syntax")
	}

	// Verify essential units
	requiredUnits := []string{
		"time_unit",
		"voltage_unit",
		"current_unit",
		"capacitive_load_unit",
		"leakage_power_unit",
	}
	for _, unit := range requiredUnits {
		if !strings.Contains(lib, unit) {
			t.Fatalf("missing required unit: %s", unit)
		}
	}

	// Verify delay_model
	if !strings.Contains(lib, "delay_model : table_lookup") {
		t.Fatal("missing or invalid delay_model")
	}

	// Verify cell block
	if !strings.Contains(lib, "cell(") {
		t.Fatal("missing cell declaration")
	}

	// Verify cell attributes
	requiredCellAttrs := []string{
		"area :",
		"cell_leakage_power :",
	}
	for _, attr := range requiredCellAttrs {
		if !strings.Contains(lib, attr) {
			t.Fatalf("missing required cell attribute: %s", attr)
		}
	}

	// Verify pin blocks (BL output, WL input, VPWR, VGND)
	requiredPins := []string{"pin(BL)", "pin(WL)", "pin(VPWR)", "pin(VGND)"}
	for _, pin := range requiredPins {
		if !strings.Contains(lib, pin) {
			t.Fatalf("missing required pin: %s", pin)
		}
	}

	// Verify pin directions
	if !regexp.MustCompile(`pin\(BL\)\s*\{[^}]*direction\s*:\s*output`).MatchString(lib) {
		t.Fatal("BL pin missing or incorrect direction (expected output)")
	}
	if !regexp.MustCompile(`pin\(WL\)\s*\{[^}]*direction\s*:\s*input`).MatchString(lib) {
		t.Fatal("WL pin missing or incorrect direction (expected input)")
	}

	// Verify timing blocks
	if !strings.Contains(lib, "timing()") {
		t.Fatal("missing timing block")
	}
	if !strings.Contains(lib, "related_pin :") {
		t.Fatal("missing related_pin in timing block")
	}

	// Verify capacitance declaration on input pins
	if !regexp.MustCompile(`pin\(WL\)\s*\{[^}]*capacitance\s*:\s*[0-9.]+`).MatchString(lib) {
		t.Fatal("WL pin missing capacitance attribute")
	}

	// Verify NLDM table references
	nldmTables := []string{"cell_rise", "cell_fall", "rise_transition", "fall_transition"}
	for _, table := range nldmTables {
		pattern := table + "(fecim_nldm_7x7)"
		if !strings.Contains(lib, pattern) {
			t.Fatalf("missing NLDM table: %s", pattern)
		}
	}

	// Verify power/ground pin types
	if !regexp.MustCompile(`pin\(VPWR\)\s*\{[^}]*pg_type\s*:\s*primary_power`).MatchString(lib) {
		t.Fatal("VPWR pin missing pg_type : primary_power")
	}
	if !regexp.MustCompile(`pin\(VGND\)\s*\{[^}]*pg_type\s*:\s*primary_ground`).MatchString(lib) {
		t.Fatal("VGND pin missing pg_type : primary_ground")
	}

	t.Logf("M6-LIB-01 PASS: Liberty syntax validation complete")
	t.Logf("  - Library declaration: valid")
	t.Logf("  - Units: %d/%d present", len(requiredUnits), len(requiredUnits))
	t.Logf("  - Cell attributes: %d/%d present", len(requiredCellAttrs), len(requiredCellAttrs))
	t.Logf("  - Pins: %d/%d present", len(requiredPins), len(requiredPins))
	t.Logf("  - NLDM tables: %d/%d present", len(nldmTables), len(nldmTables))
}

// TestM6LIB01_LibertySyntaxValidation_1T1R validates 1T1R cell structure
func TestM6LIB01_LibertySyntaxValidation_1T1R(t *testing.T) {
	cfg := config.DefaultCellConfig()
	cfg.CellType = "1t1r"
	lib := GenerateLiberty(cfg)

	// 1T1R should have WL and SL input pins
	if !strings.Contains(lib, "pin(WL)") {
		t.Fatal("1T1R missing WL pin")
	}
	if !strings.Contains(lib, "pin(SL)") {
		t.Fatal("1T1R missing SL pin")
	}

	// Verify both have input direction
	if !regexp.MustCompile(`pin\(WL\)\s*\{[^}]*direction\s*:\s*input`).MatchString(lib) {
		t.Fatal("WL pin missing or incorrect direction")
	}
	if !regexp.MustCompile(`pin\(SL\)\s*\{[^}]*direction\s*:\s*input`).MatchString(lib) {
		t.Fatal("SL pin missing or incorrect direction")
	}

	// Verify function for 1T1R: (WL & SL)
	if !strings.Contains(lib, "function : \"(WL & SL)\"") {
		t.Fatal("1T1R missing correct boolean function (WL & SL)")
	}

	t.Logf("M6-LIB-01 PASS (1T1R): Liberty syntax validation for 1T1R cell")
}

// TestM6LIB01_LibertySyntaxValidation_2T1R validates 2T1R cell structure
func TestM6LIB01_LibertySyntaxValidation_2T1R(t *testing.T) {
	cfg := config.DefaultCellConfig()
	cfg.CellType = "2t1r"
	lib := GenerateLiberty(cfg)

	// 2T1R should have WL, CSL, and SL input pins
	requiredPins := []string{"pin(WL)", "pin(CSL)", "pin(SL)"}
	for _, pin := range requiredPins {
		if !strings.Contains(lib, pin) {
			t.Fatalf("2T1R missing %s", pin)
		}
	}

	// Verify function for 2T1R: (WL & CSL & SL)
	if !strings.Contains(lib, "function : \"(WL & CSL & SL)\"") {
		t.Fatal("2T1R missing correct boolean function (WL & CSL & SL)")
	}

	t.Logf("M6-LIB-01 PASS (2T1R): Liberty syntax validation for 2T1R cell")
}

// TestM6LIB01_OperatingConditions validates operating_conditions block
func TestM6LIB01_OperatingConditions(t *testing.T) {
	cfg := config.DefaultCellConfig()
	cfg.Voltage = 1.8
	cfg.Temperature = 25.0
	cfg.Process = 1.0
	lib := GenerateLiberty(cfg)

	// Check operating_conditions block exists
	if !strings.Contains(lib, "operating_conditions(") {
		t.Fatal("missing operating_conditions block")
	}

	// Validate process, temperature, voltage attributes
	requiredOC := []string{"process :", "temperature :", "voltage :"}
	for _, attr := range requiredOC {
		if !strings.Contains(lib, attr) {
			t.Fatalf("operating_conditions missing: %s", attr)
		}
	}

	// Check default_operating_conditions reference
	if !strings.Contains(lib, "default_operating_conditions :") {
		t.Fatal("missing default_operating_conditions declaration")
	}

	t.Logf("M6-LIB-01 PASS: Operating conditions block validated")
}
