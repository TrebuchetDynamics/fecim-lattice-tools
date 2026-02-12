package physics

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveToFile_RoundTrip(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "physics_out.yaml")

	if err := cfg.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(data) < 100 {
		t.Fatalf("saved file too small: %d bytes", len(data))
	}
}

func TestLoadWithDefaults_ReturnsValidConfig(t *testing.T) {
	cfg := LoadWithDefaults()
	if cfg == nil {
		t.Fatal("LoadWithDefaults returned nil")
	}
	// Config may have 0 FeCIMLevels from embedded defaults; just check non-nil
}

func TestReload_ReturnsConfig(t *testing.T) {
	cfg, err := Reload()
	if err != nil {
		t.Fatalf("Reload: %v", err)
	}
	if cfg == nil {
		t.Fatal("Reload returned nil config")
	}
}

func TestGetNumLevels_MaterialOverride(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Material with AnalogStates set should use that value
	m := cfg.GetMaterial("hzo_ftj_140")
	if m == nil {
		t.Skip("hzo_ftj_140 not in config")
	}
	levels := m.GetNumLevels(cfg)
	if m.AnalogStates > 0 && levels != m.AnalogStates {
		t.Errorf("expected %d levels from material, got %d", m.AnalogStates, levels)
	}

	// Material without AnalogStates should fall back to global or 30
	fake := &Material{}
	levels = fake.GetNumLevels(cfg)
	if levels < 1 {
		t.Errorf("expected positive fallback levels, got %d", levels)
	}

	// Nil config fallback should return 30
	levels = fake.GetNumLevels(nil)
	if levels != 30 {
		t.Errorf("expected ultimate fallback 30, got %d", levels)
	}
}

func TestGetMaterial_UnknownReturnsNil(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if m := cfg.GetMaterial("nonexistent_material_xyz"); m != nil {
		t.Error("expected nil for unknown material")
	}
}

func TestMaterial_PsMicroCcm2(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	m := cfg.DefaultMaterial()
	if m == nil {
		t.Fatal("no default material")
	}
	ps := m.PsMicroCcm2()
	if ps <= 0 {
		t.Errorf("expected positive Ps, got %f µC/cm²", ps)
	}
}
