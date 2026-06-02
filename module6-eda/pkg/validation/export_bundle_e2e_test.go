package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	"fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/module6-eda/pkg/export"
	"fecim-lattice-tools/module6-eda/pkg/validate"
)

func TestModule6E2ECellAndArrayExportBundleValidationMatrix(t *testing.T) {
	tests := []struct {
		name      string
		cellType  string
		rows      int
		cols      int
		pins      []string
		configure func(*compiler.ArrayConfig)
	}{
		{name: "passive", cellType: "passive", rows: 3, cols: 4, pins: []string{"WL", "BL", "VPWR", "VGND"}},
		{name: "one-transistor-one-resistor", cellType: "1t1r", rows: 4, cols: 3, pins: []string{"WL", "BL", "SL", "VPWR", "VGND"}, configure: func(c *compiler.ArrayConfig) { c.With1T1R() }},
		{name: "two-transistor-one-resistor", cellType: "2t1r", rows: 2, cols: 5, pins: []string{"WL", "BL", "SL", "CSL", "VPWR", "VGND"}, configure: func(c *compiler.ArrayConfig) { c.With2T1R() }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			cellCfg := config.DefaultCellConfig()
			cellCfg.CellType = tc.cellType
			cellName := expectedE2ECellName(cellCfg)

			lefPath := filepath.Join(dir, cellName+".lef")
			libPath := filepath.Join(dir, cellName+".lib")
			cellVerilogPath := filepath.Join(dir, cellName+".v")
			cellSpicePath := filepath.Join(dir, cellName+".sp")
			cellVerilog := export.GenerateCellVerilog(cellCfg)
			if !strings.Contains(cellVerilog, "Placeholder") {
				t.Fatalf("generated cell Verilog for %s lost educational placeholder warning", tc.name)
			}
			writeModule6E2EFile(t, lefPath, export.GenerateLEF(cellCfg))
			writeModule6E2EFile(t, libPath, export.GenerateLiberty(cellCfg))
			writeModule6E2EFile(t, cellVerilogPath, module6E2EStripInlineComments(cellVerilog))
			writeModule6E2EFile(t, cellSpicePath, module6E2ECellSPICE(cellName, tc.pins))

			if err := CrossCheckFiles(lefPath, libPath, cellVerilogPath); err != nil {
				t.Fatalf("CrossCheckFiles(%s): %v", tc.name, err)
			}
			if err := validate.ValidateLVSConsistency(lefPath, cellVerilogPath, cellSpicePath); err != nil {
				t.Fatalf("ValidateLVSConsistency(%s): %v", tc.name, err)
			}
			e2eDRCRules := validate.DRCRules{MinMetalWidth: 0.13, MinMetalSpacing: 0.0, MinViaEnclosure: 0.05}
			if err := validate.ValidateLEFDRCFile(lefPath, e2eDRCRules); err != nil {
				t.Fatalf("ValidateLEFDRCFile(%s): %v", tc.name, err)
			}
			if err := validate.ValidateLEFWithPDKConstraintsFile(lefPath, e2eDRCRules); err != nil {
				t.Fatalf("ValidateLEFWithPDKConstraintsFile(%s): %v", tc.name, err)
			}

			cfg := compiler.NewComputeConfig(tc.rows, tc.cols)
			cfg.Name = "fecim_bundle_" + strings.ReplaceAll(tc.name, "-", "_")
			if tc.configure != nil {
				tc.configure(cfg)
			}
			design, err := compiler.GenerateDesign(cfg)
			if err != nil {
				t.Fatalf("GenerateDesign(%s): %v", tc.name, err)
			}
			defPath := filepath.Join(dir, cfg.Name+".def")
			spicePath := filepath.Join(dir, cfg.Name+".sp")
			verilogPath := filepath.Join(dir, cfg.Name+".v")
			if err := export.ExportDEF(design, defPath); err != nil {
				t.Fatalf("ExportDEF(%s): %v", tc.name, err)
			}
			if err := export.ExportSPICE(design, spicePath, 1.8); err != nil {
				t.Fatalf("ExportSPICE(%s): %v", tc.name, err)
			}
			if err := export.ExportVerilog(design, verilogPath); err != nil {
				t.Fatalf("ExportVerilog(%s): %v", tc.name, err)
			}
			if err := ValidateDEF(defPath); err != nil {
				t.Fatalf("ValidateDEF(%s): %v", tc.name, err)
			}
			stats, err := GetDEFStats(defPath)
			if err != nil {
				t.Fatalf("GetDEFStats(%s): %v", tc.name, err)
			}
			if stats["design_name"] != cfg.Name || stats["component_count"] != tc.rows*tc.cols {
				t.Fatalf("DEF stats(%s) = %#v, want design %s and %d components", tc.name, stats, cfg.Name, tc.rows*tc.cols)
			}

			spice := readModule6E2EFile(t, spicePath)
			verilog := readModule6E2EFile(t, verilogPath)
			for _, pin := range tc.pins {
				if !strings.Contains(spice, pin) {
					t.Fatalf("%s SPICE missing architecture pin %s\n%s", tc.name, pin, spice)
				}
			}
			if tc.cellType == "1t1r" && (!strings.Contains(spice, "fefet_1t1r") || !strings.Contains(verilog, "fecim_1t1r")) {
				t.Fatalf("1T1R bundle missing selector-specific SPICE/Verilog markers")
			}
			if tc.cellType == "2t1r" && (!strings.Contains(spice, "fefet_2t1r") || !strings.Contains(verilog, "fecim_2t1r")) {
				t.Fatalf("2T1R bundle missing selector-specific SPICE/Verilog markers")
			}
		})
	}
}

func expectedE2ECellName(cfg config.CellConfig) string {
	if cfg.CellType == "1t1r" && cfg.Name == "fecim_bitcell" {
		return "fecim_1t1r_bitcell"
	}
	if cfg.CellType == "2t1r" && cfg.Name == "fecim_bitcell" {
		return "fecim_2t1r_bitcell"
	}
	return cfg.Name
}

func module6E2ECellSPICE(cellName string, pins []string) string {
	return fmt.Sprintf("* Module 6 E2E cell interface\n.subckt %s %s\n.ends %s\n", cellName, strings.Join(pins, " "), cellName)
}

func module6E2EStripInlineComments(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if idx := strings.Index(line, "//"); idx >= 0 {
			line = strings.TrimRight(line[:idx], " \t")
		}
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

func writeModule6E2EFile(t *testing.T, path, content string) {
	t.Helper()
	if strings.TrimSpace(content) == "" {
		t.Fatalf("empty content for %s", path)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readModule6E2EFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
