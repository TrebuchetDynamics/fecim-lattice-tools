package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/config/physics"
	"fecim-lattice-tools/module6-eda/pkg/openlane"
	"fecim-lattice-tools/validation/configvalidator"
	"gopkg.in/yaml.v3"
)

func TestConfigValidation_EndToEnd_LoadAllConfigYAMLAndValidateSchema(t *testing.T) {
	configDir := filepath.Clean(filepath.Join("..", "config"))

	seen := map[string]bool{}
	err := filepath.WalkDir(configDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".yaml") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		base := filepath.Base(path)
		seen[base] = true
		if err := validateYAMLBySchema(base, data); err != nil {
			return fmt.Errorf("schema validation failed for %s: %w", path, err)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("failed validating config directory: %v", err)
	}

	// Ensure we exercised every split config file expected by loader.
	required := []string{
		"benchmarks.yaml",
		"calibration.yaml",
		"constants.yaml",
		"crossbar.yaml",
		"energy.yaml",
		"materials.yaml",
		"mnist.yaml",
		"preisach.yaml",
		"simulation.yaml",
		"timing.yaml",
		"training.yaml",
	}
	for _, name := range required {
		if !seen[name] {
			t.Fatalf("expected config file %s to be present and validated", name)
		}
	}
}

func TestConfigValidation_EndToEnd_MalformedConfigsProduceDescriptiveErrors(t *testing.T) {
	tests := []struct {
		name       string
		json       string
		errSubstrs []string
	}{
		{
			name: "missing required field",
			json: `{
				"material_name": "HZO",
				"num_levels": 4,
				"calibrations": {}
			}`,
			errSubstrs: []string{"version", "required"},
		},
		{
			name: "out of range value",
			json: `{
				"version": 1,
				"material": "HZO",
				"temperature_k": 300,
				"grid_size": 2,
				"distribution_type": "gaussian",
				"hysteron_states": [1, 1, 1],
				"alpha_mean": 1.0,
				"alpha_sigma": 1.0,
				"beta_mean": -1.0,
				"beta_sigma": 1.0,
				"correlation": 1.5
			}`,
			errSubstrs: []string{"correlation", "between"},
		},
		{
			name: "wrong type",
			json: `{
				"name": "weights",
				"rows": "2",
				"cols": 2,
				"weights": [[0.1, 0.2], [0.3, 0.4]]
			}`,
			errSubstrs: []string{"rows", "invalid type"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := configvalidator.ValidateJSON([]byte(tt.json))
			if result.Valid {
				t.Fatalf("expected invalid result for malformed config")
			}
			if len(result.Errors) == 0 {
				t.Fatalf("expected at least one validation error")
			}

			full := result.String()
			for _, sub := range tt.errSubstrs {
				if !strings.Contains(full, sub) {
					t.Fatalf("expected validation output to contain %q; got: %s", sub, full)
				}
			}

			for _, e := range result.Errors {
				if e.Field == "" {
					t.Fatalf("expected error field path to be populated: %+v", e)
				}
				msg := e.Error()
				if !strings.Contains(msg, e.Field) {
					t.Fatalf("expected formatted error to include field path %q, got %q", e.Field, msg)
				}
				if len(strings.TrimSpace(e.Message)) < 5 {
					t.Fatalf("expected descriptive error message, got %q", e.Message)
				}
			}
		})
	}
}

func TestConfigValidation_EndToEnd_DefaultGenerationProducesValidConfigs(t *testing.T) {
	// 1) Physics default loading path should always produce a usable config.
	cfg := physics.LoadWithDefaults()
	if cfg == nil {
		t.Fatal("LoadWithDefaults returned nil")
	}
	if cfg.Constants.RoomTemperature <= 0 {
		t.Fatalf("invalid defaults: room_temperature=%v", cfg.Constants.RoomTemperature)
	}
	if cfg.Constants.BitsPerCell <= 0 {
		t.Fatalf("invalid defaults: bits_per_cell=%v", cfg.Constants.BitsPerCell)
	}

	// 2) OpenLane default configuration should generate a valid schema-conforming config.
	ol := openlane.DefaultConfig()
	openlaneDoc := map[string]any{
		"DESIGN_NAME":      "fecim_default",
		"VERILOG_FILES":    "dir::src/fecim_default.v",
		"CLOCK_PERIOD":     10.0,
		"CLOCK_PORT":       "clk",
		"PDK":              ol.PDKVariant,
		"STD_CELL_LIBRARY": ol.SCLibrary,
	}
	openlaneJSON, err := json.Marshal(openlaneDoc)
	if err != nil {
		t.Fatalf("marshal default OpenLane doc: %v", err)
	}
	result := configvalidator.ValidateJSON(openlaneJSON)
	if !result.Valid {
		t.Fatalf("generated default OpenLane config should validate, got:\n%s", result.String())
	}
	if result.ConfigType != string(configvalidator.ConfigTypeOpenLane) {
		t.Fatalf("expected OpenLane config type, got %s", result.ConfigType)
	}
}

func validateYAMLBySchema(fileName string, data []byte) error {
	var root map[string]any
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("yaml parse: %w", err)
	}

	extract := func(key string) (map[string]any, error) {
		raw, ok := root[key]
		if !ok {
			return nil, fmt.Errorf("missing top-level key %q", key)
		}
		m, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("top-level key %q must be mapping", key)
		}
		return m, nil
	}

	decode := func(section map[string]any, out any) error {
		b, err := yaml.Marshal(section)
		if err != nil {
			return err
		}
		return yaml.Unmarshal(b, out)
	}

	switch fileName {
	case "constants.yaml":
		section, err := extract("constants")
		if err != nil {
			return err
		}
		var v physics.Constants
		if err := decode(section, &v); err != nil {
			return err
		}
		if v.RoomTemperature <= 0 || v.BitsPerCell <= 0 {
			return fmt.Errorf("constants contain invalid non-positive values")
		}
	case "materials.yaml":
		section, err := extract("materials")
		if err != nil {
			return err
		}
		var v map[string]*physics.Material
		if err := decode(section, &v); err != nil {
			return err
		}
		if len(v) == 0 {
			return fmt.Errorf("materials must not be empty")
		}
	case "crossbar.yaml":
		section, err := extract("crossbar")
		if err != nil {
			return err
		}
		var v physics.Crossbar
		if err := decode(section, &v); err != nil {
			return err
		}
		if v.DefaultRows <= 0 || v.DefaultCols <= 0 {
			return fmt.Errorf("crossbar default dimensions must be positive")
		}
	case "training.yaml":
		section, err := extract("training")
		if err != nil {
			return err
		}
		var v physics.Training
		if err := decode(section, &v); err != nil {
			return err
		}
		if v.DefaultBatchSize <= 0 {
			return fmt.Errorf("training default_batch_size must be positive")
		}
	case "energy.yaml":
		section, err := extract("energy")
		if err != nil {
			return err
		}
		var v physics.Energy
		if err := decode(section, &v); err != nil {
			return err
		}
		if v.ReadEnergyJ <= 0 || v.WriteEnergyJ <= 0 {
			return fmt.Errorf("energy read/write energies must be positive")
		}
	case "timing.yaml":
		section, err := extract("timing")
		if err != nil {
			return err
		}
		var v physics.Timing
		if err := decode(section, &v); err != nil {
			return err
		}
		if v.ReadLatencyS <= 0 || v.WriteLatencyS <= 0 {
			return fmt.Errorf("timing latencies must be positive")
		}
	case "preisach.yaml":
		section, err := extract("preisach")
		if err != nil {
			return err
		}
		var v physics.Preisach
		if err := decode(section, &v); err != nil {
			return err
		}
		if v.GridSize <= 0 {
			return fmt.Errorf("preisach grid_size must be positive")
		}
	case "calibration.yaml":
		section, err := extract("calibration")
		if err != nil {
			return err
		}
		var v physics.Calibration
		if err := decode(section, &v); err != nil {
			return err
		}
		if v.Iterations <= 0 {
			return fmt.Errorf("calibration iterations must be positive")
		}
	case "simulation.yaml":
		section, err := extract("simulation")
		if err != nil {
			return err
		}
		var v physics.Simulation
		if err := decode(section, &v); err != nil {
			return err
		}
		if v.FrameRateHz <= 0 || v.DtS <= 0 {
			return fmt.Errorf("simulation frame_rate_hz and dt_s must be positive")
		}
	case "mnist.yaml":
		section, err := extract("mnist")
		if err != nil {
			return err
		}
		var v physics.MNIST
		if err := decode(section, &v); err != nil {
			return err
		}
		if v.InputSize <= 0 || v.OutputSize <= 0 {
			return fmt.Errorf("mnist input/output sizes must be positive")
		}
	case "benchmarks.yaml":
		section, err := extract("benchmarks")
		if err != nil {
			return err
		}
		if len(section) == 0 {
			return fmt.Errorf("benchmarks section must not be empty")
		}
	default:
		return fmt.Errorf("no schema validator defined for %s", fileName)
	}

	return nil
}
