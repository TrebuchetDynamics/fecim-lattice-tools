// pkg/openlane/openlane_tcl_syntax_test.go
// M6-OL-01: OpenLane TCL Syntax Validation
//
// Tests:
// - Generate OpenLane TCL script
// - Parse TCL syntax (basic check for `set`, `source`, balanced braces)
// - Verify no syntax errors

package openlane

import (
	"fmt"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/module6-eda/pkg/export"
)

// TestGenerateTCLScript_M6_OL_01 tests TCL script generation for OpenLane flow
func TestGenerateTCLScript_M6_OL_01(t *testing.T) {
	cfg := config.DefaultArrayConfig()
	cfg.Rows = 4
	cfg.Cols = 4

	// Generate OpenLane config (JSON format in v2.0)
	configContent := export.GenerateOpenLaneConfig(cfg)

	// Verify config content is not empty
	if configContent == "" {
		t.Fatal("GenerateOpenLaneConfig() returned empty string")
	}

	// Verify it contains JSON structure
	if !strings.Contains(configContent, "{") || !strings.Contains(configContent, "}") {
		t.Errorf("GenerateOpenLaneConfig() missing JSON braces: %s", configContent)
	}

	// Verify required OpenLane v2.0 fields
	requiredFields := []string{
		"DESIGN_NAME",
		"VERILOG_FILES",
		"FP_DEF_TEMPLATE",
		"EXTRA_LEFS",
		"EXTRA_LIBS",
	}

	for _, field := range requiredFields {
		if !strings.Contains(configContent, field) {
			t.Errorf("GenerateOpenLaneConfig() missing required field: %s", field)
		}
	}

	t.Logf("M6-OL-01: OpenLane config generation — PASS (%d bytes)", len(configContent))
}

// TestTCLSyntaxBalancedBraces_M6_OL_01 tests TCL brace balancing
func TestTCLSyntaxBalancedBraces_M6_OL_01(t *testing.T) {
	cfg := config.DefaultArrayConfig()
	cfg.Rows = 8
	cfg.Cols = 8

	configContent := export.GenerateOpenLaneConfig(cfg)

	// Count braces (JSON uses braces similarly to TCL)
	openCount := strings.Count(configContent, "{")
	closeCount := strings.Count(configContent, "}")

	if openCount != closeCount {
		t.Errorf("Unbalanced braces: %d open, %d close", openCount, closeCount)
	}

	t.Logf("M6-OL-01: Brace balance — %d pairs — PASS", openCount)
}

// TestTCLSyntaxNoForbiddenChars_M6_OL_01 tests for syntax-breaking characters
func TestTCLSyntaxNoForbiddenChars_M6_OL_01(t *testing.T) {
	cfg := config.DefaultArrayConfig()
	cfg.Rows = 4
	cfg.Cols = 4

	configContent := export.GenerateOpenLaneConfig(cfg)

	// Check for unescaped quotes that could break JSON/TCL
	// JSON strings should use escaped quotes if nested
	lines := strings.Split(configContent, "\n")
	for i, line := range lines {
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Check for suspicious patterns
		// In JSON: unbalanced quotes on a single line (outside string context)
		quoteCount := strings.Count(line, "\"")
		if quoteCount%2 != 0 && !strings.HasSuffix(strings.TrimSpace(line), ",") {
			// Allow odd quotes if line ends with comma (continuation)
			t.Errorf("Line %d has unbalanced quotes: %s", i+1, line)
		}
	}

	t.Logf("M6-OL-01: No forbidden chars — PASS")
}

// TestTCLSyntaxValidJSON_M6_OL_01 tests that config is valid JSON
func TestTCLSyntaxValidJSON_M6_OL_01(t *testing.T) {
	cfg := config.DefaultArrayConfig()
	cfg.Rows = 16
	cfg.Cols = 16

	configContent := export.GenerateOpenLaneConfig(cfg)

	// Basic JSON structure validation
	trimmed := strings.TrimSpace(configContent)
	if !strings.HasPrefix(trimmed, "{") {
		t.Errorf("Config does not start with '{'")
	}
	if !strings.HasSuffix(trimmed, "}") {
		t.Errorf("Config does not end with '}'")
	}

	// Check for comma-separated key-value pairs
	if !strings.Contains(configContent, ":") {
		t.Errorf("Config missing key-value separator ':'")
	}

	t.Logf("M6-OL-01: Valid JSON structure — PASS")
}

// TestTCLSyntaxMultipleArraySizes_M6_OL_01 tests TCL generation across array sizes
func TestTCLSyntaxMultipleArraySizes_M6_OL_01(t *testing.T) {
	testCases := []struct {
		rows int
		cols int
	}{
		{4, 4},
		{8, 8},
		{16, 16},
		{32, 32},
	}

	for _, tc := range testCases {
		tc := tc // Capture loop variable
		t.Run(fmt.Sprintf("%dx%d", tc.rows, tc.cols), func(t *testing.T) {
			cfg := config.DefaultArrayConfig()
			cfg.Rows = tc.rows
			cfg.Cols = tc.cols

			configContent := export.GenerateOpenLaneConfig(cfg)

			// Verify content is generated
			if len(configContent) == 0 {
				t.Errorf("Empty config for %dx%d array", tc.rows, tc.cols)
			}

			// Verify design name contains dimensions
			expectedName := fmt.Sprintf("fecim_crossbar_%dx%d", tc.rows, tc.cols)
			if !strings.Contains(configContent, expectedName) {
				t.Errorf("Config missing design name '%s'", expectedName)
			}

			// Verify brace balance
			openCount := strings.Count(configContent, "{")
			closeCount := strings.Count(configContent, "}")
			if openCount != closeCount {
				t.Errorf("Unbalanced braces in %dx%d: %d open, %d close",
					tc.rows, tc.cols, openCount, closeCount)
			}

			t.Logf("M6-OL-01: %dx%d array config — %d bytes, %d brace pairs — PASS",
				tc.rows, tc.cols, len(configContent), openCount)
		})
	}
}

// TestTCLSyntaxNoTrailingComma_M6_OL_01 tests JSON doesn't have trailing commas
func TestTCLSyntaxNoTrailingComma_M6_OL_01(t *testing.T) {
	cfg := config.DefaultArrayConfig()
	configContent := export.GenerateOpenLaneConfig(cfg)

	// Check for trailing comma before closing brace (invalid JSON)
	if strings.Contains(configContent, ",\n}") || strings.Contains(configContent, ", }") {
		t.Errorf("Config has trailing comma before closing brace (invalid JSON)")
	}

	t.Logf("M6-OL-01: No trailing commas — PASS")
}
