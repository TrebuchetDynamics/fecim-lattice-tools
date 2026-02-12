package openlane

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig_HasSaneDefaults(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.PDKVariant != "sky130A" {
		t.Errorf("expected sky130A, got %s", cfg.PDKVariant)
	}
	if cfg.SCLibrary != "sky130_fd_sc_hd" {
		t.Errorf("expected sky130_fd_sc_hd, got %s", cfg.SCLibrary)
	}
	if cfg.PreferredMode != ModeDocker {
		t.Errorf("expected ModeDocker, got %d", cfg.PreferredMode)
	}
	if cfg.TimeoutPlacement != 5*time.Minute {
		t.Errorf("unexpected placement timeout: %v", cfg.TimeoutPlacement)
	}
}

func TestSaveConfig_LoadConfig_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-config.json")

	orig := DefaultConfig()
	orig.PDKRoot = "/test/pdk"
	orig.PreferredMode = ModeNative
	orig.TimeoutSynthesis = 20 * time.Minute

	if err := SaveConfig(orig, path); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if loaded.PDKRoot != "/test/pdk" {
		t.Errorf("PDKRoot mismatch: %s", loaded.PDKRoot)
	}
	if loaded.PreferredMode != ModeNative {
		t.Errorf("PreferredMode mismatch: %d", loaded.PreferredMode)
	}
	if loaded.TimeoutSynthesis != 20*time.Minute {
		t.Errorf("TimeoutSynthesis mismatch: %v", loaded.TimeoutSynthesis)
	}
}

func TestLoadConfig_MissingFile_ReturnsDefault(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/path.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
	// Should still return a valid default config
	if cfg == nil {
		t.Fatal("expected non-nil default config on error")
	}
	if cfg.PDKVariant != "sky130A" {
		t.Errorf("expected default PDKVariant, got %s", cfg.PDKVariant)
	}
}

func TestLoadConfig_InvalidJSON_ReturnsDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	os.WriteFile(path, []byte("{invalid json!!!"), 0644)

	cfg, err := LoadConfig(path)
	if err == nil {
		t.Error("expected parse error")
	}
	if cfg == nil || cfg.PDKVariant != "sky130A" {
		t.Error("expected default config on parse error")
	}
}

func TestConfig_PathHelpers(t *testing.T) {
	cfg := DefaultConfig()
	cfg.PDKRoot = "/pdk"

	tlef := cfg.GetTechLEFPath()
	if tlef == "" {
		t.Error("empty tech LEF path")
	}
	clef := cfg.GetCellLEFPath()
	if clef == "" {
		t.Error("empty cell LEF path")
	}
	lib := cfg.GetLibertyPath()
	if lib == "" {
		t.Error("empty liberty path")
	}
}

func TestGetConfigPath_NonEmpty(t *testing.T) {
	p := GetConfigPath()
	if p == "" {
		t.Error("empty config path")
	}
}

func TestGetVolareSetupInstructions_NonEmpty(t *testing.T) {
	s := GetVolareSetupInstructions()
	if len(s) < 50 {
		t.Error("instructions too short")
	}
}
