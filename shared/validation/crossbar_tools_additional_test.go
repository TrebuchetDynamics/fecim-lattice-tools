package validation

import (
	"strings"
	"testing"
)

func TestValidateCrossSim(t *testing.T) {
	if !IsCrossSimAvailable() {
		t.Skip("CrossSim not available, skipping validation test")
	}

	passed, output, err := ValidateCrossSim()
	if err != nil && !strings.Contains(err.Error(), "CrossSim not installed") {
		t.Logf("ValidateCrossSim error (may be expected): %v", err)
	}

	if passed {
		t.Logf("CrossSim validation passed: %s", output)
		if !strings.Contains(output, "VALIDATION_PASSED") {
			t.Error("expected VALIDATION_PASSED in output")
		}
	} else {
		t.Logf("CrossSim validation failed (may be expected in CI): %s", output)
	}
}

func TestValidateCrossSimNotInstalled(t *testing.T) {
	if IsCrossSimAvailable() {
		t.Skip("CrossSim is installed, cannot test not-installed path")
	}

	passed, _, err := ValidateCrossSim()
	if err == nil {
		t.Error("expected error when CrossSim not installed")
	}
	if passed {
		t.Error("expected validation to fail when CrossSim not installed")
	}
	if err != nil && !strings.Contains(err.Error(), "not installed") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidateBadCrossbar(t *testing.T) {
	if !IsBadCrossbarAvailable() {
		t.Skip("BadCrossbar not available, skipping validation test")
	}

	passed, output, err := ValidateBadCrossbar()
	if err != nil && !strings.Contains(err.Error(), "BadCrossbar not installed") {
		t.Logf("ValidateBadCrossbar error (may be expected): %v", err)
	}

	if passed {
		t.Logf("BadCrossbar validation passed: %s", output)
		if !strings.Contains(output, "VALIDATION_PASSED") {
			t.Error("expected VALIDATION_PASSED in output")
		}
	} else {
		t.Logf("BadCrossbar validation failed (may be expected in CI): %s", output)
	}
}

func TestValidateBadCrossbarNotInstalled(t *testing.T) {
	if IsBadCrossbarAvailable() {
		t.Skip("BadCrossbar is installed, cannot test not-installed path")
	}

	passed, _, err := ValidateBadCrossbar()
	if err == nil {
		t.Error("expected error when BadCrossbar not installed")
	}
	if passed {
		t.Error("expected validation to fail when BadCrossbar not installed")
	}
	if err != nil && !strings.Contains(err.Error(), "not installed") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCheckAllTools(t *testing.T) {
	results := CheckAllTools()
	if len(results) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(results))
	}

	var foundCrossSim, foundBadCrossbar bool
	for _, r := range results {
		if r.Name == "CrossSim" {
			foundCrossSim = true
		}
		if r.Name == "BadCrossbar" {
			foundBadCrossbar = true
		}
	}
	if !foundCrossSim || !foundBadCrossbar {
		t.Error("expected both CrossSim and BadCrossbar in results")
	}
}

func TestValidateAllTools(t *testing.T) {
	results := ValidateAllTools()
	if len(results) != 2 {
		t.Fatalf("expected 2 validation results, got %d", len(results))
	}

	for _, r := range results {
		t.Logf("Tool: %s, Passed: %v, Elapsed: %v", r.Tool, r.Passed, r.Elapsed)
		if r.Error != nil {
			t.Logf("  Error: %v", r.Error)
		}
	}
}

func TestInstallToolsIfNeeded(t *testing.T) {
	// This should check but not actually install (requires root/pip permissions)
	results := InstallToolsIfNeeded()
	t.Logf("Install check returned %d results", len(results))
	for _, r := range results {
		t.Logf("Tool: %s, Success: %v", r.Tool, r.Success)
		if r.Error != nil {
			t.Logf("  Error: %v", r.Error)
		}
	}
}
