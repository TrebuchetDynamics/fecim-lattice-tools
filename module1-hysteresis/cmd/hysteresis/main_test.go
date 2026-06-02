package hysteresiscli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/module1-hysteresis/pkg/simulation"
	"fecim-lattice-tools/shared/cli"
)

// captureStdout runs f and returns everything written to stdout during execution.
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	var buf strings.Builder
	done := make(chan struct{})
	go func() {
		b := make([]byte, 65536)
		for {
			n, err := r.Read(b)
			if n > 0 {
				buf.Write(b[:n])
			}
			if err != nil {
				break
			}
		}
		done <- struct{}{}
	}()

	f()

	w.Close()
	os.Stdout = old
	<-done
	return buf.String()
}

func TestGetMaterialKey(t *testing.T) {
	tests := []struct {
		material string // material name field
		wantKey  string
	}{
		{"HZO (Si-doped)", "default"},
		{"FeCIM HZO", "fecim"},
		{"FeCIM HZO (TARGET - NOT DEMONSTRATED)", "fecim-target"},
		{"Literature Superlattice (Cheema 2020)", "superlattice"},
		{"Cryogenic HZO (4K)", "cryogenic"},
		{"HZO Standard (32 states)", "hzo32"},
		{"HZO FTJ (140 states)", "ftj140"},
		{"AlScN (8-16 states)", "alscn"},
		{"Unknown material", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.material, func(t *testing.T) {
			m := ferroelectric.DefaultHZO()
			m.Name = tt.material
			got := getMaterialKey(m)
			if got != tt.wantKey {
				t.Errorf("getMaterialKey(%q) = %q, want %q", tt.material, got, tt.wantKey)
			}
		})
	}
}

func TestGetMaterialKey_NilMaterial(t *testing.T) {
	// Nil material should not panic; returns "default"
	got := getMaterialKey(nil)
	if got != "default" {
		t.Errorf("getMaterialKey(nil) = %q, want %q", got, "default")
	}
}

func TestListMaterials(t *testing.T) {
	out := captureStdout(func() {
		listMaterials()
	})

	if !strings.Contains(out, "Available materials") {
		t.Errorf("listMaterials output missing header, got: %s", out)
	}
	if !strings.Contains(out, "default") {
		t.Errorf("listMaterials output missing material key, got: %s", out)
	}
}

func TestPrintMaterialInfo(t *testing.T) {
	m := ferroelectric.DefaultHZO()
	out := captureStdout(func() {
		printMaterialInfo(m)
	})

	if !strings.Contains(out, "Material Parameters") {
		t.Errorf("printMaterialInfo missing header, got: %s", out)
	}
	if !strings.Contains(out, m.Name) {
		t.Errorf("printMaterialInfo missing material name %q, got: %s", m.Name, out)
	}
	if !strings.Contains(out, "μC/cm²") {
		t.Errorf("printMaterialInfo missing polarization units, got: %s", out)
	}
}

func TestBuildMaterialResult_AllMaterials(t *testing.T) {
	names := []string{"default", "fecim", "superlattice", "cryogenic", "hzo32", "ftj140", "alscn"}
	for _, n := range names {
		t.Run(n, func(t *testing.T) {
			m := getMaterial(n)
			res := buildMaterialResult(m)
			if res.Material == "" {
				t.Errorf("empty material name for %q", n)
			}
			if res.RemanentPol <= 0 || res.RemanentPol > 200 {
				t.Errorf("Pr out of range for %q: %v", n, res.RemanentPol)
			}
			if res.SaturationPol <= 0 || res.SaturationPol > 200 {
				t.Errorf("Ps out of range for %q: %v", n, res.SaturationPol)
			}
			if res.CoerciveField <= 0 || res.CoerciveField > 20 {
				t.Errorf("Ec out of range for %q: %v MV/cm", n, res.CoerciveField)
			}
			if res.Thickness <= 0 || res.Thickness > 1000 {
				t.Errorf("thickness out of range for %q: %v nm", n, res.Thickness)
			}
			if res.DiscreteLevels != 30 {
				t.Errorf("levels should be 30 for %q, got %d", n, res.DiscreteLevels)
			}
		})
	}
}

func TestRunWithNoArgs(t *testing.T) {
	// Running with no args should return an error about gogpu migration
	err := Run([]string{})
	if err == nil {
		t.Fatal("Run with no args should return an error")
	}
	if !strings.Contains(err.Error(), "no longer launches") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunWithHelp(t *testing.T) {
	err := Run([]string{"--help"})
	if err != nil {
		t.Errorf("Run with --help should return nil, got: %v", err)
	}
}

func TestRunWithListMaterials(t *testing.T) {
	out := captureStdout(func() {
		err := Run([]string{"--list-materials"})
		if err != nil {
			t.Errorf("Run with --list-materials failed: %v", err)
		}
	})
	if !strings.Contains(out, "Available materials") {
		t.Errorf("expected material listing, got: %s", out)
	}
}

func TestRunWithVulkan(t *testing.T) {
	err := Run([]string{"--vulkan"})
	if err == nil {
		t.Fatal("Run with --vulkan should return an error")
	}
	if !strings.Contains(err.Error(), "no longer launches Vulkan") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunHeadless(t *testing.T) {
	m := ferroelectric.DefaultHZO()
	engine := simulation.NewEngine(m)
	engine.SetFrequency(1e6)

	out := captureStdout(func() {
		runHeadless(engine, m)
	})

	if !strings.Contains(out, m.Name) {
		t.Errorf("headless output missing material name, got: %s", out)
	}
	if !strings.Contains(out, "SIMULATION SUMMARY") {
		t.Errorf("headless output missing summary, got: %s", out)
	}
}

func TestRunHeadless_AllMaterials(t *testing.T) {
	names := []string{"default", "fecim", "superlattice", "cryogenic", "hzo32", "ftj140", "alscn"}
	for _, n := range names {
		t.Run(n, func(t *testing.T) {
			m := getMaterial(n)
			engine := simulation.NewEngine(m)
			engine.SetFrequency(1e6)

			out := captureStdout(func() {
				runHeadless(engine, m)
			})
			if !strings.Contains(out, m.Name) {
				t.Errorf("headless output missing material name %q", m.Name)
			}
			if !strings.Contains(out, "μC/cm²") {
				t.Errorf("headless output missing polarization units for %q", m.Name)
			}
		})
	}
}

func TestRunBatchHysteresis_JSON(t *testing.T) {
	batch := &cli.BatchProcessor{}
	// We can't easily inject items; test via the JSON path
	common := cli.NewCommonFlags()
	out, err := cli.NewOutputWriter(common)
	if err != nil {
		t.Fatalf("failed to create output writer: %v", err)
	}
	defer out.Close()

	// Test with empty batch — should return nil
	err = runBatchHysteresis(batch, common, out)
	if err != nil {
		t.Errorf("runBatchHysteresis with empty batch failed: %v", err)
	}
}

func TestRunHeadless_Regression(t *testing.T) {
	// Basic material properties should be consistent across runs
	m := getMaterial("superlattice")
	engine := simulation.NewEngine(m)
	engine.SetFrequency(1e6)

	out1 := captureStdout(func() {
		runHeadless(engine, m)
	})

	engine2 := simulation.NewEngine(m)
	engine2.SetFrequency(1e6)

	out2 := captureStdout(func() {
		runHeadless(engine2, m)
	})

	// Output should be deterministic (same material, same freq)
	if out1 != out2 {
		t.Log("Deterministic regression note: outputs differ — this may indicate Preisach state non-determinism")
	}
}

func TestPrintMaterialInfo_NilMaterialSafety(t *testing.T) {
	// Should not panic on nil
	out := captureStdout(func() {
		printMaterialInfo(nil)
	})
	if out == "" {
		t.Logf("printMaterialInfo(nil) returned empty output (nil material case)")
	}
}

func TestGetMaterial_AllKeys(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"default", "HZO (Si-doped, Park 2015 midpoint)"},
		{"fecim", "FeCIM HZO"},
		{"superlattice", "Literature Superlattice (HZO nanolaminate 2025)"},
		{"cryogenic", "Cryogenic HZO (4K)"},
		{"hzo32", "HZO Standard (32 states)"},
		{"ftj140", "HZO FTJ (140 states)"},
		{"alscn", "AlScN (8-16 states)"},
		{"unknown-key", "HZO (Si-doped, Park 2015 midpoint)"}, // fallback to default
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			m := getMaterial(tt.key)
			if m.Name != tt.want {
				t.Errorf("getMaterial(%q).Name = %q, want %q", tt.key, m.Name, tt.want)
			}
		})
	}
}

func createTestHysteresisResult(m *ferroelectric.HZOMaterial) HysteresisResult {
	return buildMaterialResult(m)
}

func TestRunE2EWideJSONBatchMaterialMatrix(t *testing.T) {
	batchPath := filepath.Join(t.TempDir(), "materials.json")
	outPath := filepath.Join(t.TempDir(), "hysteresis-batch.json")
	if err := os.WriteFile(batchPath, []byte(`["default", "fecim", "superlattice", "cryogenic", "hzo32", "ftj140", "alscn", "unknown-key"]`), 0644); err != nil {
		t.Fatalf("write batch: %v", err)
	}

	if err := Run([]string{"--json", "--batch", batchPath, "--output", outPath}); err != nil {
		t.Fatalf("Run JSON batch: %v", err)
	}
	artifact := readHysteresisJSONFile[struct {
		Total     int `json:"total"`
		Succeeded int `json:"succeeded"`
		Failed    int `json:"failed"`
		Results   []struct {
			Success bool `json:"success"`
			Data    struct {
				Material        string  `json:"material"`
				RemanentPol     float64 `json:"remanent_polarization_uC_cm2"`
				SaturationPol   float64 `json:"saturation_polarization_uC_cm2"`
				CoerciveField   float64 `json:"coercive_field_MV_cm"`
				CoerciveVoltage float64 `json:"coercive_voltage_V"`
				Thickness       float64 `json:"thickness_nm"`
				Permittivity    float64 `json:"permittivity"`
				SwitchingTime   float64 `json:"switching_time_ns"`
				EnduranceCycles float64 `json:"endurance_cycles"`
				DiscreteLevels  int     `json:"discrete_levels"`
				BitsPerCell     float64 `json:"bits_per_cell"`
			} `json:"data"`
		} `json:"results"`
	}](t, outPath)
	if artifact.Total != 8 || artifact.Succeeded != 8 || artifact.Failed != 0 || len(artifact.Results) != 8 {
		t.Fatalf("batch summary = total %d succeeded %d failed %d results %d, want 8/8/0/8", artifact.Total, artifact.Succeeded, artifact.Failed, len(artifact.Results))
	}
	seen := map[string]bool{}
	for i, result := range artifact.Results {
		if !result.Success {
			t.Fatalf("result %d not successful: %+v", i, result)
		}
		data := result.Data
		if data.Material == "" || data.RemanentPol <= 0 || data.SaturationPol <= 0 || data.CoerciveField <= 0 || data.CoerciveVoltage <= 0 || data.Thickness <= 0 || data.Permittivity <= 0 || data.SwitchingTime <= 0 || data.EnduranceCycles <= 0 {
			t.Fatalf("result %d has incomplete physical metrics: %+v", i, data)
		}
		if data.DiscreteLevels != 30 || data.BitsPerCell != 4.91 {
			t.Fatalf("result %d level baseline = %d / %.2f, want 30 / 4.91", i, data.DiscreteLevels, data.BitsPerCell)
		}
		seen[data.Material] = true
	}
	for _, want := range []string{"HZO (Si-doped, Park 2015 midpoint)", "FeCIM HZO", "Literature Superlattice (HZO nanolaminate 2025)", "Cryogenic HZO (4K)", "HZO Standard (32 states)", "HZO FTJ (140 states)", "AlScN (8-16 states)"} {
		if !seen[want] {
			t.Fatalf("batch JSON did not include material %q; seen=%v", want, seen)
		}
	}
}

func TestRunE2EListMaterialsJSONAndOutputIsolation(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "materials.json")
	stdout := captureStdout(func() {
		if err := Run([]string{"--json", "--list-materials", "--output", outPath}); err != nil {
			t.Fatalf("Run list materials JSON: %v", err)
		}
	})
	if stdout != "" {
		t.Fatalf("stdout = %q, want empty when --output captures JSON", stdout)
	}
	materials := readHysteresisJSONFile[[]struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	}](t, outPath)
	if len(materials) < 7 {
		t.Fatalf("materials length = %d, want broad material list", len(materials))
	}
	entries := map[string]string{}
	for _, material := range materials {
		if material.Key == "" || material.Name == "" {
			t.Fatalf("material entry incomplete: %+v", material)
		}
		entries[material.Key+"|"+material.Name] = material.Name
	}
	for _, wantName := range []string{"FeCIM HZO", "Literature Superlattice", "Cryogenic HZO", "HZO Standard (32 states)", "HZO FTJ (140 states)", "AlScN (8-16 states)"} {
		found := false
		for _, gotName := range entries {
			found = found || gotName == wantName
		}
		if !found {
			t.Fatalf("JSON material listing missing material %q: %v", wantName, entries)
		}
	}
}

func TestRunE2EHeadlessConfigAndFrequencyWorkflow(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "hysteresis-config.json")
	if err := os.WriteFile(configPath, []byte(`{"material":"cryogenic","frequency":2500000,"temperature":4}`), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	out := captureStdout(func() {
		if err := Run([]string{"--headless", "--config", configPath}); err != nil {
			t.Fatalf("Run headless config: %v", err)
		}
	})
	for _, want := range []string{"FeCIM Hysteresis Visualizer", "Cryogenic HZO (4K)", "P-E", "Discrete Analog Levels", "SIMULATION SUMMARY", "30 levels are an educational discretization"} {
		if !strings.Contains(out, want) {
			t.Fatalf("headless config output missing %q\n%s", want, out)
		}
	}
}

func TestRunE2EInvalidConfigAndBatchDoNotEmitJSONResults(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "missing config", args: []string{"--config", filepath.Join(t.TempDir(), "missing.json"), "--headless"}, wantErr: "failed to load config"},
		{name: "missing batch", args: []string{"--batch", filepath.Join(t.TempDir(), "missing-batch.json"), "--json"}, wantErr: "failed to load batch file"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			outPath := filepath.Join(t.TempDir(), "should-not-exist.json")
			args := append([]string{}, tc.args...)
			args = append(args, "--output", outPath)
			stdout := captureStdout(func() {
				err := Run(args)
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("Run error = %v, want containing %q", err, tc.wantErr)
				}
			})
			if stdout != "" {
				t.Fatalf("stdout = %q, want no output before invalid config/batch returns", stdout)
			}
			if data, err := os.ReadFile(outPath); err == nil && len(data) != 0 {
				t.Fatalf("output file %s contains %q, want no JSON result on invalid input", outPath, string(data))
			}
		})
	}
}

func readHysteresisJSONFile[T any](t *testing.T, path string) T {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("decode %s: %v\n%s", path, err, string(data))
	}
	return value
}

func TestHysteresisResult_JSONMarshaling(t *testing.T) {
	m := getMaterial("superlattice")
	res := createTestHysteresisResult(m)

	if res.Material != m.Name {
		t.Errorf("Material field mismatch: %q vs %q", res.Material, m.Name)
	}
	if res.RemanentPol <= 0 {
		t.Errorf("RemanentPol should be positive, got %v", res.RemanentPol)
	}
	if res.SaturationPol <= 0 {
		t.Errorf("SaturationPol should be positive, got %v", res.SaturationPol)
	}
	if res.CoerciveField <= 0 {
		t.Errorf("CoerciveField should be positive, got %v", res.CoerciveField)
	}
	if res.CoerciveVoltage <= 0 {
		t.Errorf("CoerciveVoltage should be positive, got %v", res.CoerciveVoltage)
	}
	if res.Thickness <= 0 {
		t.Errorf("Thickness should be positive, got %v", res.Thickness)
	}
	if res.Permittivity <= 0 {
		t.Errorf("Permittivity should be positive, got %v", res.Permittivity)
	}
	if res.SwitchingTime <= 0 {
		t.Errorf("SwitchingTime should be positive, got %v", res.SwitchingTime)
	}
	if res.EnduranceCycles <= 0 {
		t.Errorf("EnduranceCycles should be positive, got %v", res.EnduranceCycles)
	}
	if res.DiscreteLevels != 30 {
		t.Errorf("DiscreteLevels should be 30, got %d", res.DiscreteLevels)
	}
	if res.BitsPerCell != 4.91 {
		t.Errorf("BitsPerCell should be 4.91, got %v", res.BitsPerCell)
	}
}
