package edalattice

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunLatticeGenE2EWideGeometryMatrix(t *testing.T) {
	tests := []struct {
		name string
		rows int
		cols int
	}{
		{name: "single-cell", rows: 1, cols: 1},
		{name: "wide", rows: 2, cols: 7},
		{name: "tall", rows: 6, cols: 3},
		{name: "square", rows: 8, cols: 8},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			outDir := filepath.Join(t.TempDir(), "lattice-out")
			t.Setenv("HOME", home)
			var stdout, stderr bytes.Buffer

			err := runLatticeGen([]string{"-rows", fmt.Sprintf("%d", tc.rows), "-cols", fmt.Sprintf("%d", tc.cols), "-output", outDir}, &stdout, &stderr)
			if err != nil {
				t.Fatalf("runLatticeGen(%dx%d): %v\nstderr:\n%s", tc.rows, tc.cols, err, stderr.String())
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want lattice-gen to report through logger only", stdout.String())
			}

			verilogPath := filepath.Join(outDir, fmt.Sprintf("lattice_%dx%d.v", tc.rows, tc.cols))
			defPath := filepath.Join(outDir, fmt.Sprintf("lattice_%dx%d.def", tc.rows, tc.cols))
			verilog := readLatticeE2EFile(t, verilogPath)
			def := readLatticeE2EFile(t, defPath)
			wantCells := tc.rows * tc.cols

			assertLatticeContains(t, verilogPath, verilog,
				"FeCIM Lattice Verilog Netlist",
				fmt.Sprintf("module lattice_%dx%d", tc.rows, tc.cols),
				fmt.Sprintf("input  wire [%d:0] WL", tc.rows-1),
				fmt.Sprintf("inout  wire [%d:0] BL", tc.cols-1),
				"endmodule",
			)
			if got := strings.Count(verilog, "fecim_bit cell_"); got != wantCells {
				t.Fatalf("%s cell instance count = %d, want %d", verilogPath, got, wantCells)
			}
			assertLatticeContains(t, verilogPath, verilog,
				"cell_0_0",
				fmt.Sprintf("cell_%d_%d", tc.rows-1, tc.cols-1),
				fmt.Sprintf(".WL  (WL[%d])", tc.rows-1),
				fmt.Sprintf(".BL  (BL[%d])", tc.cols-1),
			)

			assertLatticeContains(t, defPath, def,
				"VERSION 5.8 ;",
				fmt.Sprintf("DESIGN lattice_%dx%d ;", tc.rows, tc.cols),
				fmt.Sprintf("COMPONENTS %d ;", wantCells),
				fmt.Sprintf("PINS %d ;", tc.rows+tc.cols+2),
				fmt.Sprintf("NETS %d ;", tc.rows+tc.cols+2),
				"END DESIGN",
			)
			if got := strings.Count(def, " fecim_bit + FIXED "); got != wantCells {
				t.Fatalf("%s DEF component count = %d, want %d", defPath, got, wantCells)
			}
			assertLatticeContains(t, defPath, def,
				"- WL[0]",
				fmt.Sprintf("- WL[%d]", tc.rows-1),
				"- BL[0]",
				fmt.Sprintf("- BL[%d]", tc.cols-1),
				"- VPWR",
				"- VGND",
			)
		})
	}
}

func TestRunLatticeGenE2EInvalidDimensionsAreSideEffectSafe(t *testing.T) {
	tests := []struct {
		name string
		rows int
		cols int
		want string
	}{
		{name: "zero row", rows: 0, cols: 4, want: "rows and cols must be > 0"},
		{name: "negative col", rows: 4, cols: -1, want: "rows and cols must be > 0"},
		{name: "dimension too large", rows: 2049, cols: 1, want: "rows/cols exceed max dimension"},
		{name: "too many cells", rows: 1000, cols: 1001, want: "array too large"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			outDir := filepath.Join(t.TempDir(), "bad-lattice-out")
			t.Setenv("HOME", home)
			var stdout, stderr bytes.Buffer

			err := runLatticeGen([]string{"-rows", fmt.Sprintf("%d", tc.rows), "-cols", fmt.Sprintf("%d", tc.cols), "-output", outDir}, &stdout, &stderr)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want containing %q", err, tc.want)
			}
			if stdout.Len() != 0 || stderr.Len() != 0 {
				t.Fatalf("stdout/stderr = %q/%q, want validation error returned without CLI noise", stdout.String(), stderr.String())
			}
			if _, err := os.Stat(outDir); !os.IsNotExist(err) {
				t.Fatalf("output dir stat = %v, want no output dir", err)
			}
			if _, err := os.Stat(filepath.Join(home, ".fecim")); !os.IsNotExist(err) {
				t.Fatalf("log dir stat = %v, want no logging side effect", err)
			}
		})
	}
}

func readLatticeE2EFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if len(data) == 0 {
		t.Fatalf("%s is empty", path)
	}
	return string(data)
}

func assertLatticeContains(t *testing.T, path, text string, needles ...string) {
	t.Helper()
	for _, needle := range needles {
		if !strings.Contains(text, needle) {
			t.Fatalf("%s missing %q\n--- content ---\n%s", path, needle, text)
		}
	}
}

func TestRunRejectsInvalidDimensions(t *testing.T) {
	err := Run([]string{"-rows", "0", "-cols", "4"})
	if err == nil {
		t.Fatal("expected error for invalid dimensions")
	}
	if !strings.Contains(err.Error(), "invalid lattice dimensions") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunLatticeGenReportsFlagErrorToStderr(t *testing.T) {
	var stdout, stderr bytes.Buffer
	home := t.TempDir()
	t.Setenv("HOME", home)

	err := runLatticeGen([]string{"-definitely-not-a-flag"}, &stdout, &stderr)

	if err == nil {
		t.Fatal("runLatticeGen error = nil, want invalid flag error")
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
	if !strings.Contains(text, "FeCIM Lattice Generator") {
		t.Fatalf("stderr = %q, want usage", text)
	}
	if _, err := os.Stat(filepath.Join(home, ".fecim")); !os.IsNotExist(err) {
		t.Fatalf("invalid flag initialized logging directory; stat error = %v", err)
	}
}
