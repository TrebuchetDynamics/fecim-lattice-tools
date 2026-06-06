package edacli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunEDACLIE2EExportsWideFormatMatrix(t *testing.T) {
	tests := []struct {
		name       string
		mode       string
		arch       string
		rows       int
		cols       int
		levels     int
		vdd        string
		tech       string
		wantActive int
	}{
		{name: "compute-passive-rectangular", mode: "compute", arch: "passive", rows: 3, cols: 5, levels: 16, vdd: "1.2", tech: "SKY130", wantActive: 0},
		{name: "memory-1t1r-square", mode: "memory", arch: "1T1R", rows: 4, cols: 4, levels: 8, vdd: "1.8", tech: "GF180MCU", wantActive: 0},
		{name: "storage-passive-tall", mode: "storage", arch: "passive", rows: 6, cols: 2, levels: 30, vdd: "2.5", tech: "IHP_SG13G2", wantActive: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			outDir := filepath.Join(t.TempDir(), "eda-out")
			designName := "fecim_e2e_" + strings.ReplaceAll(tc.name, "-", "_")
			t.Setenv("HOME", home)

			var stdout, stderr bytes.Buffer
			err := runEDACLI([]string{
				"-json-output",
				"-quiet",
				"-mode", tc.mode,
				"-arch", tc.arch,
				"-rows", fmt.Sprintf("%d", tc.rows),
				"-cols", fmt.Sprintf("%d", tc.cols),
				"-levels", fmt.Sprintf("%d", tc.levels),
				"-vdd", tc.vdd,
				"-tech", tc.tech,
				"-name", designName,
				"-output", outDir,
			}, &stdout, &stderr)
			if err != nil {
				t.Fatalf("runEDACLI returned error: %v\nstderr:\n%s", err, stderr.String())
			}
			if strings.Contains(stderr.String(), "Error:") {
				t.Fatalf("stderr contains error: %s", stderr.String())
			}

			var result EDAResult
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatalf("stdout is not EDAResult JSON: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
			}
			if result.DesignName != designName || result.Mode != tc.mode || result.Rows != tc.rows || result.Cols != tc.cols || result.Technology != tc.tech {
				t.Fatalf("result metadata = %+v, want %s/%s/%dx%d/%s", result, designName, tc.mode, tc.rows, tc.cols, tc.tech)
			}
			if result.TotalCells != tc.rows*tc.cols || result.ActiveCells != tc.wantActive || result.AreaMM2 <= 0 || result.PowerMW <= 0 {
				t.Fatalf("result stats = %+v, want positive stats and %d/%d cells", result, tc.rows*tc.cols, tc.wantActive)
			}
			if len(result.OutputFiles) != 5 {
				t.Fatalf("output files = %v, want JSON/CSV/SPICE/Verilog/DEF", result.OutputFiles)
			}
			for _, path := range result.OutputFiles {
				if !strings.HasPrefix(path, outDir+string(os.PathSeparator)) {
					t.Fatalf("output path %q escaped output dir %q", path, outDir)
				}
				if info, err := os.Stat(path); err != nil || info.Size() == 0 {
					t.Fatalf("output path %q stat=(%v,%v), want non-empty file", path, info, err)
				}
			}
			assertCLIFileContains(t, filepath.Join(outDir, designName+"_design.json"), "\"name\": \""+designName+"\"", "\"total_cells\": "+fmt.Sprintf("%d", tc.rows*tc.cols))
			assertCLIFileContains(t, filepath.Join(outDir, designName+"_cells.csv"), "row,col", "level")
			assertCLIFileContains(t, filepath.Join(outDir, designName+".sp"), "NOT suitable for signoff", ".subckt", fmt.Sprintf(".param VDD = %.2f", mustParseCLIFloat(t, tc.vdd)))
			assertCLIFileContains(t, filepath.Join(outDir, designName+".v"), "module fecim_crossbar", fmt.Sprintf("parameter ROWS = %d", tc.rows), fmt.Sprintf("parameter COLS = %d", tc.cols), "input", "inout")
			assertCLIFileContains(t, filepath.Join(outDir, designName+".def"), "VERSION", "DESIGN "+designName, fmt.Sprintf("COMPONENTS %d", tc.rows*tc.cols))
		})
	}
}

func TestRunEDACLIE2EComputeWeightsPropagateToArtifacts(t *testing.T) {
	home := t.TempDir()
	outDir := filepath.Join(t.TempDir(), "weighted-out")
	weightsPath := filepath.Join(t.TempDir(), "weights.json")
	t.Setenv("HOME", home)
	if err := os.WriteFile(weightsPath, []byte(`{
  "name": "signed-weights",
  "rows": 2,
  "cols": 3,
  "weights": [[-1.0, 0.0, 0.75], [0.25, -0.5, 1.0]]
}`), 0644); err != nil {
		t.Fatalf("write weights: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := runEDACLI([]string{
		"-json-output", "-quiet",
		"-mode", "compute",
		"-rows", "2", "-cols", "3",
		"-levels", "12",
		"-gmin", "5", "-gmax", "55",
		"-name", "fecim_weighted_e2e",
		"-input", weightsPath,
		"-output", outDir,
	}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runEDACLI weighted compute: %v\nstderr:\n%s", err, stderr.String())
	}
	var result EDAResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("decode weighted EDA result: %v\n%s", err, stdout.String())
	}
	if result.Rows != 2 || result.Cols != 3 || result.TotalCells != 6 || result.ActiveCells != 6 || result.PowerMW <= 0 || result.AreaMM2 <= 0 {
		t.Fatalf("weighted result = %+v, want 2x3 active compute design with positive exported stats", result)
	}
	assertCLIFileContains(t, filepath.Join(outDir, "fecim_weighted_e2e_design.json"), "\"initial_weight\": -1", "\"initial_weight\": 1", "\"level\"")
	assertCLIFileContains(t, filepath.Join(outDir, "fecim_weighted_e2e_cells.csv"), "0,0", "1,2", "-1.000000", "1.000000")
	assertCLIFileContains(t, filepath.Join(outDir, "fecim_weighted_e2e.v"), "MODE", "Compute", "LEVEL", "weight=-1.0000", "weight=1.0000")
	assertCLIFileContains(t, filepath.Join(outDir, "fecim_weighted_e2e.sp"), "R_level=200000.00", "R_level=18181.82")
}

func mustParseCLIFloat(t *testing.T, value string) float64 {
	t.Helper()
	var parsed float64
	if _, err := fmt.Sscanf(value, "%f", &parsed); err != nil {
		t.Fatalf("parse float %q: %v", value, err)
	}
	return parsed
}

func assertCLIFileContains(t *testing.T, path string, needles ...string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	text := string(data)
	for _, needle := range needles {
		if !strings.Contains(text, needle) {
			t.Fatalf("%s missing %q\n--- file ---\n%s", path, needle, text)
		}
	}
}

func TestRunEDACLIE2EConfigFileAndSelectiveExports(t *testing.T) {
	home := t.TempDir()
	workDir := t.TempDir()
	configPath := filepath.Join(workDir, "eda-config.json")
	outDir := filepath.Join(workDir, "selective-out")
	t.Setenv("HOME", home)
	if err := os.WriteFile(configPath, []byte(`{
  "mode": "memory",
  "rows": 5,
  "cols": 7,
  "levels": 10,
  "technology": "GF180MCU",
  "vdd": 1.65
}`), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := runEDACLI([]string{
		"-json-output", "-quiet",
		"-config", configPath,
		"-name", "fecim_config_selective_e2e",
		"-output", outDir,
		"-csv=false",
		"-verilog=false",
		"-def=false",
	}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runEDACLI config/selective: %v\nstderr:\n%s", err, stderr.String())
	}
	var result EDAResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("decode config/selective result: %v\nstdout:\n%s", err, stdout.String())
	}
	if result.Mode != "memory" || result.Rows != 5 || result.Cols != 7 || result.TotalCells != 35 || result.Technology != "GF180MCU" {
		t.Fatalf("config-derived result = %+v, want memory 5x7 GF180MCU", result)
	}
	if len(result.OutputFiles) != 2 {
		t.Fatalf("output files = %v, want only design JSON and SPICE", result.OutputFiles)
	}
	assertCLIFileContains(t, filepath.Join(outDir, "fecim_config_selective_e2e_design.json"), "\"array_rows\": 5", "\"array_cols\": 7", "\"levels\": 10", "\"vdd\": 1.65")
	assertCLIFileContains(t, filepath.Join(outDir, "fecim_config_selective_e2e.sp"), ".param VDD = 1.65", "Operation Mode: Memory", "Array: 35 cells")
	for _, unexpected := range []string{"fecim_config_selective_e2e_cells.csv", "fecim_config_selective_e2e.v", "fecim_config_selective_e2e.def"} {
		if _, err := os.Stat(filepath.Join(outDir, unexpected)); !os.IsNotExist(err) {
			t.Fatalf("%s exists/stat error = %v, want selective export to omit it", unexpected, err)
		}
	}
}

func TestRunEDACLIE2EInvalidConfigurationMatrixIsSideEffectSafe(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "bad design path", args: []string{"-name", "../escape"}, wantErr: "invalid characters"},
		{name: "zero rows", args: []string{"-rows", "0"}, wantErr: "rows and cols must be > 0"},
		{name: "too many cells", args: []string{"-rows", "4096", "-cols", "4096"}, wantErr: "array too large"},
		{name: "bad levels", args: []string{"-levels", "31"}, wantErr: "levels must be in [2,30]"},
		{name: "bad vdd", args: []string{"-vdd", "5.5"}, wantErr: "vdd must be in (0,5]"},
		{name: "bad conductance", args: []string{"-gmin", "100", "-gmax", "10"}, wantErr: "conductance must satisfy"},
		{name: "bad mode", args: []string{"-mode", "alchemy"}, wantErr: "unknown mode"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			outDir := filepath.Join(t.TempDir(), "should-not-exist")
			t.Setenv("HOME", home)
			args := append([]string{"-json-output", "-quiet", "-output", outDir}, tc.args...)
			var stdout, stderr bytes.Buffer
			err := runEDACLI(args, &stdout, &stderr)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error = %v, want containing %q", err, tc.wantErr)
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty on invalid config", stdout.String())
			}
			if _, err := os.Stat(outDir); !os.IsNotExist(err) {
				t.Fatalf("output dir stat = %v, want no output directory", err)
			}
			if _, err := os.Stat(filepath.Join(home, ".fecim")); !os.IsNotExist(err) {
				t.Fatalf("log dir stat = %v, want validation failure before logging side effects", err)
			}
		})
	}
}

func TestRunEDACLIE2ERejectsMalformedWeightsWithoutArtifacts(t *testing.T) {
	home := t.TempDir()
	workDir := t.TempDir()
	weightsPath := filepath.Join(workDir, "ragged-weights.json")
	outDir := filepath.Join(workDir, "bad-weights-out")
	t.Setenv("HOME", home)
	if err := os.WriteFile(weightsPath, []byte(`{"name":"ragged","weights":[[1,2],[3]]}`), 0644); err != nil {
		t.Fatalf("write malformed weights: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := runEDACLI([]string{"-json-output", "-quiet", "-mode", "compute", "-input", weightsPath, "-output", outDir}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "weights row 1 has 1 columns, expected 2") {
		t.Fatalf("error = %v, want ragged weights error", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty on malformed weights", stdout.String())
	}
	if _, err := os.Stat(outDir); !os.IsNotExist(err) {
		t.Fatalf("output dir stat = %v, want no artifact directory for malformed weights", err)
	}
}

func TestRunEDACLIReportsFlagErrorToStderr(t *testing.T) {
	var stdout, stderr bytes.Buffer
	home := t.TempDir()
	t.Setenv("HOME", home)

	err := runEDACLI([]string{"-definitely-not-a-flag"}, &stdout, &stderr)

	if err == nil {
		t.Fatal("runEDACLI error = nil, want invalid flag error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout length = %d, want 0; stdout=%q", stdout.Len(), stdout.String())
	}
	text := stderr.String()
	if !strings.Contains(text, "flag provided but not defined: -definitely-not-a-flag") {
		t.Fatalf("stderr = %q, want invalid flag context", text)
	}
	if !strings.Contains(text, "Error:") {
		t.Fatalf("stderr = %q, want error prefix", text)
	}
	if !strings.Contains(text, "FeCIM EDA CLI") {
		t.Fatalf("stderr = %q, want usage", text)
	}
	if _, err := os.Stat(filepath.Join(home, ".fecim")); !os.IsNotExist(err) {
		t.Fatalf("invalid flag initialized logging directory; stat error = %v", err)
	}
}
