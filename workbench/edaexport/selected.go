package edaexport

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
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
	handle, err := os.OpenRoot(root)
	if err != nil {
		return "", err
	}
	defer handle.Close()
	rel, err := ExportSelectedRunRoot(handle, runID)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, rel), nil
}

// ExportSelectedRunRoot exports a verified successful run and returns exports/<runID>
// relative to root. The caller owns root and must keep it open until this returns.
// Writes and failure cleanup use pinned child handles, never root.Name().
func ExportSelectedRunRoot(root *os.Root, runID string) (string, error) {
	run, err := experiment.LoadRunRoot(root, runID)
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

	if err := root.MkdirAll("exports", 0o755); err != nil {
		return "", err
	}
	exports, err := root.OpenRoot("exports")
	if err != nil {
		return "", err
	}
	defer exports.Close()
	if err := exports.Mkdir(runID, 0o755); err != nil {
		return "", fmt.Errorf("create export directory: %w", err)
	}
	dir, err := exports.OpenRoot(runID)
	if err != nil {
		_ = exports.Remove(runID) // Only an empty entry, never recursive cleanup.
		return "", err
	}
	defer dir.Close()
	created := make(map[string]os.FileInfo)
	complete := false
	defer func() {
		if complete {
			return
		}
		// Remove only files we created and still recognize in our pinned directory.
		for name, owned := range created {
			entry, err := dir.Lstat(name)
			if err == nil && os.SameFile(owned, entry) {
				_ = dir.Remove(name)
			}
		}
		owned, ownErr := dir.Stat(".")
		entry, entryErr := exports.Lstat(runID)
		if ownErr == nil && entryErr == nil && os.SameFile(owned, entry) {
			_ = exports.Remove(runID)
		}
	}()
	write := func(name string, data []byte) error {
		info, err := writeExportFile(dir, name, data)
		if info != nil {
			created[name] = info
		}
		if err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
		return nil
	}

	cell := edaconfig.DefaultCellConfig()
	cell.Name = config.Name + "_cell"
	cell.Width = config.CellPitch
	cell.Height = config.RowHeight
	files := []struct {
		name     string
		generate func() string
	}{
		{"design.spice", func() string { return daexport.GenerateSPICE(array, config.Peripherals.VDD) }},
		{"design.v", func() string { return daexport.GenerateVerilogWithDefaults(array) }},
		{"design.def", func() string { return daexport.GenerateDEFWithDefaults(array) }},
		{"cell.lef", func() string { return daexport.GenerateLEF(cell) }},
		{"cell.lib", func() string { return daexport.GenerateLiberty(cell) }},
	}
	hashes := make(map[string]string, len(files))
	for _, file := range files {
		data := []byte(file.generate())
		if err := write(file.name, data); err != nil {
			return "", err
		}
		// A successful exclusive write includes all bytes and a successful close.
		digest := sha256.Sum256(data)
		hashes[file.name] = hex.EncodeToString(digest[:])
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
	// The manifest is the completion marker and is always published last.
	if err := write("manifest.json", blob); err != nil {
		return "", err
	}
	complete = true
	return filepath.Join("exports", runID), nil
}

// Return the new file's identity even on a write/close error so failure cleanup
// can distinguish an incomplete owned file from a substituted directory entry.
func writeExportFile(root *os.Root, name string, data []byte) (os.FileInfo, error) {
	file, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	n, err := file.Write(data)
	closeErr := file.Close()
	if err != nil {
		return info, err
	}
	if n != len(data) {
		return info, io.ErrShortWrite
	}
	return info, closeErr
}
