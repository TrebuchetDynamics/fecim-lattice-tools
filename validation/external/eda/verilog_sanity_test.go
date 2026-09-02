package external_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	"fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/module6-eda/pkg/export"
	"fecim-lattice-tools/validation/external/internal/testsupport"
)

type verilogPair struct {
	name      string
	cellPath  string
	arrayPath string
}

// TestVerilogSanityCheck validates the legacy planned generator/stub pair and
// the production GUI export pair. Its scope is syntax and elaboration only,
// not synthesis, timing, physical-design, or hardware validation.
func TestVerilogSanityCheck(t *testing.T) {
	dir := t.TempDir()

	design := makeTestDesign(4, 4, compiler.ArchPassive)
	legacyArray := export.GenerateVerilogWithDefaults(design)
	const passiveCellStub = `module fecim_bit #(
    parameter LEVEL = 0
) (
    input wire WL,
    inout wire BL,
    inout wire VPWR,
    inout wire VGND
);
endmodule
`
	assertVerilogPairStructure(t, "legacy", legacyArray, passiveCellStub, "module fecim_crossbar (", "module fecim_bit #(", "fecim_bit #(", 16)
	legacy := writeVerilogPair(t, dir, "legacy", passiveCellStub, legacyArray)

	productionArray := export.GenerateArrayVerilog(config.DefaultArrayConfig())
	productionCell := export.GenerateCellVerilog(config.DefaultCellConfig())
	assertVerilogPairStructure(t, "production", productionArray, productionCell, "module fecim_crossbar_4x4", "module fecim_bitcell", "fecim_bitcell cell_", 16)
	production := writeVerilogPair(t, dir, "production", productionCell, productionArray)

	pairs := []verilogPair{legacy, production}
	switch {
	case testsupport.HasCommand("iverilog"):
		version, err := exec.Command("iverilog", "-V").CombinedOutput()
		if err != nil {
			t.Fatalf("iverilog version failed: %v\n%s", err, version)
		}
		t.Logf("iverilog version: %s", strings.TrimSpace(string(version)))
		for _, pair := range pairs {
			outputPath := filepath.Join(dir, pair.name+"-lint.out")
			cmd := exec.Command("iverilog", "-g2012", "-t", "null", "-o", outputPath, pair.cellPath, pair.arrayPath)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s Verilog pair failed iverilog syntax/elaboration: %v\n%s", pair.name, err, output)
			}
		}
	case testsupport.HasCommand("verilator"):
		version, err := exec.Command("verilator", "--version").CombinedOutput()
		if err != nil {
			t.Fatalf("verilator version failed: %v\n%s", err, version)
		}
		t.Logf("verilator version: %s", strings.TrimSpace(string(version)))
		for _, pair := range pairs {
			outputDir := filepath.Join(dir, pair.name+"-verilator")
			cmd := exec.Command("verilator", "--lint-only", "--Mdir", outputDir, pair.cellPath, pair.arrayPath)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s Verilog pair failed verilator syntax/elaboration: %v\n%s", pair.name, err, output)
			}
		}
	default:
		t.Skip("iverilog or verilator is required for external Verilog validation")
	}
}

func assertVerilogPairStructure(t *testing.T, name, array, cell, topModule, masterModule, instanceToken string, wantInstances int) {
	t.Helper()
	if !strings.Contains(array, topModule) {
		t.Fatalf("%s array missing top module %q", name, topModule)
	}
	if !strings.Contains(cell, masterModule) {
		t.Fatalf("%s cell model missing master module %q", name, masterModule)
	}
	if got := strings.Count(array, instanceToken); got != wantInstances {
		t.Fatalf("%s array has %d instances matching %q, want %d", name, got, instanceToken, wantInstances)
	}
}

func writeVerilogPair(t *testing.T, dir, name, cell, array string) verilogPair {
	t.Helper()
	pair := verilogPair{
		name:      name,
		cellPath:  filepath.Join(dir, name+"-cells.v"),
		arrayPath: filepath.Join(dir, name+"-array.v"),
	}
	if err := os.WriteFile(pair.cellPath, []byte(cell), 0o644); err != nil {
		t.Fatalf("write %s cell Verilog: %v", name, err)
	}
	if err := os.WriteFile(pair.arrayPath, []byte(array), 0o644); err != nil {
		t.Fatalf("write %s array Verilog: %v", name, err)
	}
	return pair
}
