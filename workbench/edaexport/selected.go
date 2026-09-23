package edaexport

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	edaconfig "fecim-lattice-tools/module6-eda/pkg/config"
	daexport "fecim-lattice-tools/module6-eda/pkg/export"
	"fecim-lattice-tools/workbench/experiment"
)

const maxExportCells = 16_384

type exportManifest struct {
	SchemaVersion int               `json:"schema_version"`
	RunID         string            `json:"run_id"`
	Design        any               `json:"resolved_design"`
	Technology    string            `json:"eda_technology"`
	Warning       string            `json:"warning"`
	FilesSHA256   map[string]string `json:"files_sha256"`
}

// ExportSelectedRun exports one persisted, integrity-verified successful run without rerunning it.
func ExportSelectedRun(root, runID string) (string, error) {
	run, err := experiment.LoadRun(root, runID)
	if err != nil {
		return "", err
	}
	if run.Design.Array.Rows <= 0 || run.Design.Array.Cols <= 0 {
		return "", fmt.Errorf("selected run has invalid array dimensions %dx%d", run.Design.Array.Rows, run.Design.Array.Cols)
	}
	if run.Design.Array.Rows > maxExportCells/run.Design.Array.Cols {
		return "", fmt.Errorf("array exceeds EDA export limit of %d cells", maxExportCells)
	}
	config := compiler.NewArrayConfig(compiler.ModeCompute, run.Design.Array.Rows, run.Design.Array.Cols)
	config.Name = "fecim_run_" + runID[:12]
	config.Levels = run.Design.Device.ConductanceLevels
	config.GMin = run.Design.Device.GMinS * 1e6
	config.GMax = run.Design.Device.GMaxS * 1e6
	config.Peripherals.ADCBits = run.Design.Circuit.ADCBits
	config.Peripherals.DACBits = run.Design.Circuit.DACBits
	config.Peripherals.TIAGain = run.Design.Circuit.TIAGainOhm
	array, err := compiler.GenerateDesign(config)
	if err != nil {
		return "", fmt.Errorf("compile selected run: %w", err)
	}

	outputDir := filepath.Join(root, "exports", runID)
	if err := os.MkdirAll(filepath.Dir(outputDir), 0o755); err != nil {
		return "", err
	}
	if err := os.Mkdir(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("create export directory: %w", err)
	}
	complete := false
	defer func() {
		if !complete {
			_ = os.RemoveAll(outputDir)
		}
	}()

	files := map[string]func(string) error{
		"design.spice": func(path string) error { return daexport.ExportSPICE(array, path, config.Peripherals.VDD) },
		"design.v":     func(path string) error { return daexport.ExportVerilog(array, path) },
		"design.def":   func(path string) error { return daexport.ExportDEF(array, path) },
	}
	cell := edaconfig.DefaultCellConfig()
	cell.Name = config.Name + "_cell"
	cell.Width = config.CellPitch
	cell.Height = config.RowHeight
	files["cell.lef"] = func(path string) error { return os.WriteFile(path, []byte(daexport.GenerateLEF(cell)), 0o644) }
	files["cell.lib"] = func(path string) error { return os.WriteFile(path, []byte(daexport.GenerateLiberty(cell)), 0o644) }

	hashes := make(map[string]string, len(files))
	for name, write := range files {
		path := filepath.Join(outputDir, name)
		if err := write(path); err != nil {
			return "", fmt.Errorf("write %s: %w", name, err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		digest := sha256.Sum256(data)
		hashes[name] = hex.EncodeToString(digest[:])
	}
	manifest := exportManifest{
		SchemaVersion: 1,
		RunID:         runID,
		Design:        run.Design,
		Technology:    config.Technology,
		Warning:       fmt.Sprintf("EDA cell/layout assumptions use %s defaults; source run tech_node=%s is not mapped to a foundry PDK.", config.Technology, run.Design.Circuit.TechNode),
		FilesSHA256:   hashes,
	}
	blob, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", err
	}
	blob = append(blob, '\n')
	if err := os.WriteFile(filepath.Join(outputDir, "manifest.json"), blob, 0o644); err != nil {
		return "", err
	}
	complete = true
	return outputDir, nil
}
