// pkg/export/new_generators_test.go
// Tests for newly added script generators:
//   - GenerateOpenLaneTCLConfig
//   - GenerateOpenLaneTCLMacroPlacement
//   - GenerateNetgenLVSScript
//   - GenerateNetgenLVSTCL
//   - GenerateMagicDRCScript
//   - GenerateMagicExtractionScript
//   - GenerateCrossSIMConfig
//   - GenerateCrossSIMRunScript
//   - GeneratePySpiceScript
//   - GenerateOpenVAFVerilogA
package export

import (
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

// ── OpenLane v1 TCL Config ────────────────────────────────────────────────

func TestGenerateOpenLaneTCLConfig_SmokeTest(t *testing.T) {
	cfg := config.DefaultArrayConfig()
	got := GenerateOpenLaneTCLConfig(cfg)

	if got == "" {
		t.Fatal("GenerateOpenLaneTCLConfig returned empty string")
	}

	// Must use OpenLane v1 set ::env() syntax
	if !strings.Contains(got, "set ::env(DESIGN_NAME)") {
		t.Error("missing set ::env(DESIGN_NAME)")
	}
	if !strings.Contains(got, "set ::env(PDK)") {
		t.Error("missing set ::env(PDK)")
	}
	if !strings.Contains(got, "set ::env(FP_SIZING)") {
		t.Error("missing set ::env(FP_SIZING)")
	}
	if !strings.Contains(got, "set ::env(RUN_LVS)") {
		t.Error("missing set ::env(RUN_LVS)")
	}
}

func TestGenerateOpenLaneTCLConfig_DesignName(t *testing.T) {
	cfg := config.ArrayConfig{Rows: 8, Cols: 16, Architecture: "passive", Technology: "sky130"}
	got := GenerateOpenLaneTCLConfig(cfg)
	if !strings.Contains(got, "fecim_crossbar_8x16") {
		t.Errorf("expected design name fecim_crossbar_8x16, got:\n%s", got)
	}
}

func TestGenerateOpenLaneTCLConfig_PDKVariants(t *testing.T) {
	tests := []struct {
		tech    string
		wantPDK string
		wantLib string
	}{
		{"sky130", "sky130A", "sky130_fd_sc_hd"},
		{"IHP_SG13G2", "sg13g2", "sg13g2_stdcell"},
		{"GF180MCU", "gf180mcuD", "gf180mcu_fd_sc_mcu7t5v0"},
	}
	for _, tc := range tests {
		cfg := config.ArrayConfig{Rows: 4, Cols: 4, Technology: tc.tech}
		got := GenerateOpenLaneTCLConfig(cfg)
		if !strings.Contains(got, tc.wantPDK) {
			t.Errorf("tech=%s: expected PDK %q in output", tc.tech, tc.wantPDK)
		}
		if !strings.Contains(got, tc.wantLib) {
			t.Errorf("tech=%s: expected library %q in output", tc.tech, tc.wantLib)
		}
	}
}

func TestGenerateOpenLaneTCLConfig_PassiveSkipsCTS(t *testing.T) {
	cfg := config.ArrayConfig{Rows: 4, Cols: 4, Architecture: "passive"}
	got := GenerateOpenLaneTCLConfig(cfg)
	// Passive array has no clock — CTS should be disabled
	if !strings.Contains(got, "set ::env(RUN_CTS) 0") {
		t.Error("passive array should have RUN_CTS=0")
	}
}

func TestGenerateOpenLaneTCLConfig_1T1REnablesCTS(t *testing.T) {
	cfg := config.ArrayConfig{Rows: 4, Cols: 4, Architecture: "1t1r"}
	got := GenerateOpenLaneTCLConfig(cfg)
	if !strings.Contains(got, "set ::env(RUN_CTS) 1") {
		t.Error("1t1r array should have RUN_CTS=1")
	}
}

func TestGenerateOpenLaneTCLConfig_DieArea(t *testing.T) {
	cfg := config.ArrayConfig{
		Rows: 4, Cols: 4,
		CellWidth: 0.46, CellHeight: 2.72,
	}
	got := GenerateOpenLaneTCLConfig(cfg)
	// Die area should be computed from array footprint + margins
	if !strings.Contains(got, "DIE_AREA") {
		t.Error("missing DIE_AREA in TCL config")
	}
}

func TestGenerateOpenLaneTCLMacroPlacement_Format(t *testing.T) {
	cfg := config.DefaultArrayConfig()
	got := GenerateOpenLaneTCLMacroPlacement(cfg)
	if got == "" {
		t.Fatal("GenerateOpenLaneTCLMacroPlacement returned empty string")
	}
	// Should contain cell name and orientation
	if !strings.Contains(got, "fecim_bitcell") {
		t.Error("expected fecim_bitcell in macro placement")
	}
	// Orientation must be one of N/S/E/W/FN/FS/FE/FW
	validOrientations := []string{" N\n", " S\n", " E\n", " W\n"}
	found := false
	for _, o := range validOrientations {
		if strings.Contains(got, o) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("no valid orientation found in macro placement:\n%s", got)
	}
}

// ── Netgen LVS ───────────────────────────────────────────────────────────

func TestGenerateNetgenLVSScript_Smoke(t *testing.T) {
	cfg := config.DefaultCellConfig()
	got := GenerateNetgenLVSScript(cfg)
	if got == "" {
		t.Fatal("GenerateNetgenLVSScript returned empty string")
	}
	if !strings.Contains(got, "netgen") {
		t.Error("missing netgen command in LVS script")
	}
	if !strings.Contains(got, "lvs") {
		t.Error("missing lvs command in LVS script")
	}
}

func TestGenerateNetgenLVSScript_CellNames(t *testing.T) {
	tests := []struct {
		cellType string
		wantName string
	}{
		{"passive", "fecim_bitcell"},
		{"1t1r", "fecim_1t1r_bitcell"},
		{"2t1r", "fecim_2t1r_bitcell"},
	}
	for _, tc := range tests {
		cfg := config.DefaultCellConfig()
		cfg.CellType = tc.cellType
		got := GenerateNetgenLVSScript(cfg)
		if !strings.Contains(got, tc.wantName) {
			t.Errorf("cellType=%s: expected cell name %q", tc.cellType, tc.wantName)
		}
	}
}

func TestGenerateNetgenLVSScript_PDKSetupTCL(t *testing.T) {
	tests := []struct {
		tech      string
		wantSetup string
	}{
		{"sky130", "sky130A_setup.tcl"},
		{"IHP_SG13G2", "sg13g2_setup.tcl"},
		{"GF180MCU", "gf180mcu_setup.tcl"},
	}
	for _, tc := range tests {
		cfg := config.DefaultCellConfig()
		cfg.Technology = tc.tech
		got := GenerateNetgenLVSScript(cfg)
		if !strings.Contains(got, tc.wantSetup) {
			t.Errorf("tech=%s: expected setup.tcl file %q", tc.tech, tc.wantSetup)
		}
	}
}

func TestGenerateNetgenLVSTCL_Smoke(t *testing.T) {
	cfg := config.DefaultCellConfig()
	got := GenerateNetgenLVSTCL(cfg)
	if got == "" {
		t.Fatal("GenerateNetgenLVSTCL returned empty string")
	}
	if !strings.Contains(got, "readnet spice") {
		t.Error("missing readnet spice command")
	}
	if !strings.Contains(got, "lvs") {
		t.Error("missing lvs command in TCL")
	}
}

// ── Magic DRC ────────────────────────────────────────────────────────────

func TestGenerateMagicDRCScript_Smoke(t *testing.T) {
	cfg := config.DefaultArrayConfig()
	got := GenerateMagicDRCScript(cfg)
	if got == "" {
		t.Fatal("GenerateMagicDRCScript returned empty string")
	}
	if !strings.Contains(got, "magic") {
		t.Error("missing magic command in DRC script")
	}
	if !strings.Contains(got, "drc check") {
		t.Error("missing drc check command")
	}
}

func TestGenerateMagicDRCScript_TechFile(t *testing.T) {
	tests := []struct {
		tech     string
		wantTech string
	}{
		{"sky130", "sky130A.tech"},
		{"IHP_SG13G2", "sg13g2.tech"},
		{"GF180MCU", "gf180mcuD.tech"},
	}
	for _, tc := range tests {
		cfg := config.ArrayConfig{Technology: tc.tech}
		got := GenerateMagicDRCScript(cfg)
		if !strings.Contains(got, tc.wantTech) {
			t.Errorf("tech=%s: expected tech file %q", tc.tech, tc.wantTech)
		}
	}
}

func TestGenerateMagicExtractionScript_Smoke(t *testing.T) {
	cfg := config.DefaultCellConfig()
	got := GenerateMagicExtractionScript(cfg)
	if got == "" {
		t.Fatal("GenerateMagicExtractionScript returned empty string")
	}
	if !strings.Contains(got, "extract all") {
		t.Error("missing extract all command")
	}
	if !strings.Contains(got, "ext2spice") {
		t.Error("missing ext2spice command")
	}
}

// ── CrossSim ──────────────────────────────────────────────────────────────

func TestGenerateCrossSIMConfig_Smoke(t *testing.T) {
	cfg := config.DefaultArrayConfig()
	got := GenerateCrossSIMConfig(cfg)
	if got == "" {
		t.Fatal("GenerateCrossSIMConfig returned empty string")
	}
	if !strings.Contains(got, "rows:") {
		t.Error("missing rows field in CrossSim YAML")
	}
	if !strings.Contains(got, "g_max:") {
		t.Error("missing g_max field in CrossSim YAML")
	}
	if !strings.Contains(got, "g_min:") {
		t.Error("missing g_min field in CrossSim YAML")
	}
}

func TestGenerateCrossSIMConfig_ConductanceRange(t *testing.T) {
	tests := []struct {
		arch       string
		wantHigher bool // 1t1r has higher conductance than passive
	}{
		{"passive", false},
		{"1t1r", true},
	}
	for _, tc := range tests {
		cfg := config.ArrayConfig{Rows: 4, Cols: 4, Architecture: tc.arch}
		got := GenerateCrossSIMConfig(cfg)
		// passive: g_max: 10.0000, 1t1r: g_max: 100.0000
		if tc.wantHigher {
			if !strings.Contains(got, "100.0000") {
				t.Errorf("arch=%s: expected g_max=100µS for 1t1r", tc.arch)
			}
		} else {
			if !strings.Contains(got, "10.0000") {
				t.Errorf("arch=%s: expected g_max=10µS for passive", tc.arch)
			}
		}
	}
}

func TestGenerateCrossSIMConfig_SneakPaths(t *testing.T) {
	passive := config.ArrayConfig{Rows: 4, Cols: 4, Architecture: "passive"}
	got := GenerateCrossSIMConfig(passive)
	if !strings.Contains(got, "enabled: true") {
		t.Error("passive array should have sneak paths enabled")
	}

	active := config.ArrayConfig{Rows: 4, Cols: 4, Architecture: "1t1r"}
	got = GenerateCrossSIMConfig(active)
	if !strings.Contains(got, "enabled: false") {
		t.Error("1t1r array should have sneak paths disabled")
	}
}

// TestGenerateCrossSIMConfig_WireResistanceDimensions verifies that WL resistance
// uses cfg.Cols (WL spans horizontally across columns) and BL resistance uses
// cfg.Rows (BL spans vertically across rows). For non-square arrays this matters.
func TestGenerateCrossSIMConfig_WireResistanceDimensions(t *testing.T) {
	// 3 rows × 8 cols: wlRes=8Ω, blRes=3Ω
	cfg := config.ArrayConfig{Rows: 3, Cols: 8, Architecture: "passive"}
	got := GenerateCrossSIMConfig(cfg)

	// WL spans cols: 8 × 1Ω/cell = 8.00 Ω
	if !strings.Contains(got, "r_wl_ohm: 8.00") {
		t.Errorf("3×8 array: expected r_wl_ohm=8.00 (cols), got:\n%s", extractParasitics(got))
	}
	// BL spans rows: 3 × 1Ω/cell = 3.00 Ω
	if !strings.Contains(got, "r_bl_ohm: 3.00") {
		t.Errorf("3×8 array: expected r_bl_ohm=3.00 (rows), got:\n%s", extractParasitics(got))
	}
	// Comment must say cols for WL
	if !strings.Contains(got, "8 cols") {
		t.Errorf("3×8 array: WL comment should mention cols, got:\n%s", extractParasitics(got))
	}
	// Comment must say rows for BL
	if !strings.Contains(got, "3 rows") {
		t.Errorf("3×8 array: BL comment should mention rows, got:\n%s", extractParasitics(got))
	}
}

func extractParasitics(yaml string) string {
	lines := strings.Split(yaml, "\n")
	for i, l := range lines {
		if strings.Contains(l, "wire_resistance:") {
			end := i + 6
			if end > len(lines) {
				end = len(lines)
			}
			return strings.Join(lines[i:end], "\n")
		}
	}
	return "wire_resistance section not found"
}

func TestGenerateCrossSIMRunScript_Smoke(t *testing.T) {
	cfg := config.DefaultArrayConfig()
	got := GenerateCrossSIMRunScript(cfg)
	if got == "" {
		t.Fatal("GenerateCrossSIMRunScript returned empty string")
	}
	if !strings.Contains(got, "CrossSimParameters") {
		t.Error("missing CrossSimParameters in runner script")
	}
	if !strings.Contains(got, "crosssim.yaml") {
		t.Error("missing config file reference in runner script")
	}
}

// ── PySpice ───────────────────────────────────────────────────────────────

func TestGeneratePySpiceScript_Smoke(t *testing.T) {
	cfg := config.DefaultArrayConfig()
	got := GeneratePySpiceScript(cfg)
	if got == "" {
		t.Fatal("GeneratePySpiceScript returned empty string")
	}
	if !strings.Contains(got, "PySpice") {
		t.Error("missing PySpice import in script")
	}
	if !strings.Contains(got, "Circuit") {
		t.Error("missing Circuit class usage")
	}
}

func TestGeneratePySpiceScript_ArrayDimensions(t *testing.T) {
	cfg := config.ArrayConfig{Rows: 8, Cols: 16, Architecture: "passive"}
	got := GeneratePySpiceScript(cfg)
	if !strings.Contains(got, "ROWS = 8") {
		t.Errorf("missing ROWS=8 in script")
	}
	if !strings.Contains(got, "COLS = 16") {
		t.Error("missing COLS=16 in script")
	}
}

func TestGeneratePySpiceScript_ConductanceParams(t *testing.T) {
	passive := config.ArrayConfig{Rows: 4, Cols: 4, Architecture: "passive"}
	got := GeneratePySpiceScript(passive)
	// passive: G_MAX = 10µS
	if !strings.Contains(got, "10.000000e-6") {
		t.Errorf("passive: expected 10µS G_MAX")
	}

	active := config.ArrayConfig{Rows: 4, Cols: 4, Architecture: "1t1r"}
	got = GeneratePySpiceScript(active)
	// 1t1r: G_MAX = 100µS
	if !strings.Contains(got, "100.000000e-6") {
		t.Errorf("1t1r: expected 100µS G_MAX")
	}
}

// ── OpenVAF Verilog-A ─────────────────────────────────────────────────────

func TestGenerateOpenVAFVerilogA_Smoke(t *testing.T) {
	cfg := config.DefaultCellConfig()
	got := GenerateOpenVAFVerilogA(cfg)
	if got == "" {
		t.Fatal("GenerateOpenVAFVerilogA returned empty string")
	}
	if !strings.Contains(got, "module fecim_lk") {
		t.Error("missing module declaration")
	}
	if !strings.Contains(got, "endmodule") {
		t.Error("missing endmodule")
	}
}

func TestGenerateOpenVAFVerilogA_LKEquation(t *testing.T) {
	cfg := config.DefaultCellConfig()
	got := GenerateOpenVAFVerilogA(cfg)
	// Must contain Landau-Khalatnikov equation components
	if !strings.Contains(got, "T_FE") {
		t.Error("missing T_FE parameter")
	}
	if !strings.Contains(got, "EC") {
		t.Error("missing EC coercive field parameter")
	}
	if !strings.Contains(got, "idt(dPdt") {
		t.Error("missing idt() state variable integration")
	}
}

func TestGenerateOpenVAFVerilogA_VerilogADirectives(t *testing.T) {
	cfg := config.DefaultCellConfig()
	got := GenerateOpenVAFVerilogA(cfg)
	// Verilog-A requires these include files
	if !strings.Contains(got, "constants.vams") {
		t.Error("missing `include constants.vams")
	}
	if !strings.Contains(got, "disciplines.vams") {
		t.Error("missing `include disciplines.vams")
	}
}

// ── Config helpers ────────────────────────────────────────────────────────

func TestDefaultGF180CellConfig(t *testing.T) {
	cfg := config.DefaultGF180CellConfig()
	if cfg.Technology != "GF180MCU" {
		t.Errorf("expected technology GF180MCU, got %s", cfg.Technology)
	}
	if cfg.Height <= 0 {
		t.Error("height must be positive")
	}
	if cfg.Width <= 0 {
		t.Error("width must be positive")
	}
	if cfg.Voltage != 1.8 {
		t.Errorf("expected 1.8V for GF180MCU, got %.1f", cfg.Voltage)
	}
}

func TestDefaultIHPCellConfig(t *testing.T) {
	cfg := config.DefaultIHPCellConfig()
	if cfg.Technology != "IHP_SG13G2" {
		t.Errorf("expected technology IHP_SG13G2, got %s", cfg.Technology)
	}
	// IHP CoreSite: 0.48 × 3.78 µm (from sg13g2_stdcell.lef)
	if cfg.Width != 0.48 {
		t.Errorf("expected IHP width 0.48 µm, got %.2f", cfg.Width)
	}
	if cfg.Height != 3.78 {
		t.Errorf("expected IHP height 3.78 µm, got %.2f", cfg.Height)
	}
	if cfg.Voltage != 1.5 {
		t.Errorf("expected 1.5V for IHP SG13G2, got %.1f", cfg.Voltage)
	}
	// Metal1 from sg13g2_tech.lef
	if cfg.MetalPitch != 0.42 {
		t.Errorf("expected IHP M1 pitch 0.42 µm, got %.2f", cfg.MetalPitch)
	}
	if cfg.MetalWidth != 0.16 {
		t.Errorf("expected IHP M1 width 0.16 µm, got %.2f", cfg.MetalWidth)
	}
}

