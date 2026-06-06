package openlane

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	edaconfig "fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/module6-eda/pkg/export"
)

func TestModule6E2EOpenLaneProjectBundleMatrix(t *testing.T) {
	tests := []struct {
		name        string
		rows        int
		cols        int
		arch        string
		cellWidth   float64
		cellHeight  float64
		wantDieArea string
	}{
		{name: "passive-rectangular", rows: 4, cols: 6, arch: "passive", cellWidth: 0.46, cellHeight: 2.72, wantDieArea: "0 0 4.760 12.880"},
		{name: "one-transistor-square", rows: 8, cols: 8, arch: "1t1r", cellWidth: 0.92, cellHeight: 3.40, wantDieArea: "0 0 9.360 29.200"},
		{name: "wide-compute", rows: 3, cols: 9, arch: "passive", cellWidth: 0.46, cellHeight: 2.72, wantDieArea: "0 0 6.140 10.160"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			projectDir := t.TempDir()
			outputDir := filepath.Join(projectDir, "output")
			cellDir := filepath.Join(projectDir, "cells", "fecim_bitcell")
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				t.Fatalf("mkdir output: %v", err)
			}
			if err := os.MkdirAll(cellDir, 0755); err != nil {
				t.Fatalf("mkdir cells: %v", err)
			}

			arrayCfg := edaconfig.ArrayConfig{Rows: tc.rows, Cols: tc.cols, Mode: "compute", Architecture: tc.arch, Technology: "sky130", CellWidth: tc.cellWidth, CellHeight: tc.cellHeight}
			openlaneJSON := export.GenerateOpenLaneConfig(arrayCfg)
			configPath := filepath.Join(projectDir, "config.json")
			writeOpenLaneE2EFile(t, configPath, openlaneJSON)

			var flow map[string]any
			if err := json.Unmarshal([]byte(openlaneJSON), &flow); err != nil {
				t.Fatalf("OpenLane JSON invalid: %v\n%s", err, openlaneJSON)
			}
			designName := fmt.Sprintf("fecim_crossbar_%dx%d", tc.rows, tc.cols)
			if flow["DESIGN_NAME"] != designName || flow["DIE_AREA"] != tc.wantDieArea {
				t.Fatalf("OpenLane metadata = %#v, want design %s die %s", flow, designName, tc.wantDieArea)
			}
			for key, want := range map[string]float64{"RUN_CTS": 0, "DESIGN_IS_CORE": 0, "SYNTH_ELABORATE_ONLY": 1, "PL_SKIP_INITIAL_PLACEMENT": 1} {
				if got, ok := flow[key].(float64); !ok || got != want {
					t.Fatalf("OpenLane key %s = %#v, want %.0f", key, flow[key], want)
				}
			}

			compilerCfg := compiler.NewComputeConfig(tc.rows, tc.cols)
			compilerCfg.Name = designName
			if strings.EqualFold(tc.arch, "1t1r") {
				compilerCfg.With1T1R()
			}
			design, err := compiler.GenerateDesign(compilerCfg)
			if err != nil {
				t.Fatalf("GenerateDesign: %v", err)
			}
			writeOpenLaneE2EGeneratedBundle(t, design, outputDir, cellDir)

			for _, key := range []string{"VERILOG_FILES", "VERILOG_FILES_BLACKBOX", "FP_DEF_TEMPLATE", "EXTRA_LEFS", "EXTRA_LIBS"} {
				logical, ok := flow[key].(string)
				if !ok || logical == "" {
					t.Fatalf("OpenLane key %s = %#v, want non-empty string", key, flow[key])
				}
				path := resolveOpenLaneE2EDirReference(t, projectDir, logical)
				if info, err := os.Stat(path); err != nil || info.Size() == 0 {
					t.Fatalf("OpenLane %s reference %q -> %q stat=(%v,%v), want non-empty file", key, logical, path, info, err)
				}
			}
			def := readOpenLaneE2EFile(t, filepath.Join(outputDir, designName+".def"))
			for _, marker := range []string{"VERSION", "DESIGN " + designName, fmt.Sprintf("COMPONENTS %d", tc.rows*tc.cols), "END COMPONENTS", "END DESIGN"} {
				if !strings.Contains(def, marker) {
					t.Fatalf("DEF for %s missing %q\n%s", tc.name, marker, def)
				}
			}
			assertOpenLaneE2ECellLibraryConsistent(t, cellDir)
		})
	}
}

func TestModule6E2EOpenLaneConfigRoundTripWithCustomToolchain(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cfg := &Config{
		PDKRoot:          filepath.Join(home, "pdk-root"),
		PDKVariant:       "sky130B",
		SCLibrary:        "sky130_fd_sc_hs",
		PreferredMode:    ModeNative,
		TimeoutPlacement: 17 * time.Second,
		TimeoutSynthesis: 29 * time.Second,
		TimeoutRouting:   41 * time.Second,
		DockerImage:      "example/openlane:e2e",
	}
	path := filepath.Join(t.TempDir(), "nested", "openlane-config.json")
	if err := SaveConfig(cfg, path); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if loaded.PDKRoot != cfg.PDKRoot || loaded.PDKVariant != cfg.PDKVariant || loaded.SCLibrary != cfg.SCLibrary || loaded.PreferredMode != ModeNative || loaded.DockerImage != cfg.DockerImage {
		t.Fatalf("loaded config = %+v, want %+v", loaded, cfg)
	}
	if loaded.TimeoutPlacement != 17*time.Second || loaded.TimeoutSynthesis != 29*time.Second || loaded.TimeoutRouting != 41*time.Second {
		t.Fatalf("loaded timeouts = %s/%s/%s", loaded.TimeoutPlacement, loaded.TimeoutSynthesis, loaded.TimeoutRouting)
	}
	for _, path := range []string{loaded.GetTechLEFPath(), loaded.GetCellLEFPath(), loaded.GetLibertyPath()} {
		if !strings.HasPrefix(path, cfg.PDKRoot+string(os.PathSeparator)) || !strings.Contains(path, cfg.PDKVariant) || !strings.Contains(path, cfg.SCLibrary) {
			t.Fatalf("derived PDK path %q does not include custom root/variant/library", path)
		}
	}
}

func assertOpenLaneE2ECellLibraryConsistent(t *testing.T, cellDir string) {
	t.Helper()
	lef := readOpenLaneE2EFile(t, filepath.Join(cellDir, "fecim_bitcell.lef"))
	lib := readOpenLaneE2EFile(t, filepath.Join(cellDir, "fecim_bitcell.lib"))
	verilog := readOpenLaneE2EFile(t, filepath.Join(cellDir, "fecim_bitcell.v"))
	for _, marker := range []string{"fecim_bitcell", "WL", "BL", "VPWR", "VGND"} {
		if !strings.Contains(lef, marker) || !strings.Contains(lib, marker) || !strings.Contains(verilog, marker) {
			t.Fatalf("cell library marker %q missing from LEF/Liberty/Verilog", marker)
		}
	}
}

func writeOpenLaneE2EGeneratedBundle(t *testing.T, design *compiler.ArrayDesign, outputDir, cellDir string) {
	t.Helper()
	if err := export.ExportVerilog(design, filepath.Join(outputDir, design.Config.Name+".v")); err != nil {
		t.Fatalf("ExportVerilog: %v", err)
	}
	if err := export.ExportDEF(design, filepath.Join(outputDir, design.Config.Name+".def")); err != nil {
		t.Fatalf("ExportDEF: %v", err)
	}
	cellCfg := edaconfig.DefaultCellConfig()
	writeOpenLaneE2EFile(t, filepath.Join(cellDir, "fecim_bitcell.lef"), export.GenerateLEF(cellCfg))
	writeOpenLaneE2EFile(t, filepath.Join(cellDir, "fecim_bitcell.lib"), export.GenerateLiberty(cellCfg))
	writeOpenLaneE2EFile(t, filepath.Join(cellDir, "fecim_bitcell.v"), stripOpenLaneE2EInlineComments(export.GenerateCellVerilog(cellCfg)))
	writeOpenLaneE2EFile(t, filepath.Join(cellDir, "fecim_bitcell.gds"), "GDS_PLACEHOLDER_FOR_OPENLANE_REFERENCE\n")
}

func resolveOpenLaneE2EDirReference(t *testing.T, projectDir, logical string) string {
	t.Helper()
	if !strings.HasPrefix(logical, "dir::") {
		t.Fatalf("logical path %q missing dir:: prefix", logical)
	}
	return filepath.Join(projectDir, strings.TrimPrefix(logical, "dir::"))
}

func readOpenLaneE2EFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func writeOpenLaneE2EFile(t *testing.T, path, content string) {
	t.Helper()
	if strings.TrimSpace(content) == "" {
		t.Fatalf("empty OpenLane E2E content for %s", path)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func stripOpenLaneE2EInlineComments(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if idx := strings.Index(line, "//"); idx >= 0 {
			line = strings.TrimRight(line[:idx], " \t")
		}
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}
