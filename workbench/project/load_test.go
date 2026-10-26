package project

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeBundle(t *testing.T, root, projectYAML, designYAML, sweepYAML string) {
	t.Helper()
	for name, body := range map[string]string{
		"project.yaml": projectYAML,
		"design.yaml":  designYAML,
		"sweep.yaml":   sweepYAML,
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

const validProjectYAML = `schema_version: 1
id: hzo-study
name: HZO study
hypothesis: Increasing ADC resolution trades area for energy fidelity.
model_version: fecim-system-v1
objectives:
  - metric: energy_pj
    direction: minimize
  - metric: latency_ns
    direction: minimize
constraints:
  - metric: area_um2
    operator: <=
    value: 5000
    unit: um2
citations: [park2015_advmat_hzo]
`

const validDesignYAML = `schema_version: 1
device:
  material: HZO
  conductance_levels: 30
  g_min_s: 0.000001
  g_max_s: 0.00003
array:
  rows: 32
  cols: 32
  read_voltage_v: 0.2
circuit:
  adc_bits: 6
  dac_bits: 4
  tia_gain_ohm: 10000
  tech_node: 65nm
`

const validSweepYAML = `schema_version: 1
seed: 17
max_points: 32
parameters:
  - path: device.conductance_levels
    values: [16, 30]
  - path: array.rows
    range: {start: 16, stop: 32, count: 2}
  - path: circuit.adc_bits
    values: [4, 6]
`

// Reopening Root.Name(), caching a checked path, or using path-based file reads
// would let these fixtures escape the caller's authority.
func TestLoadRootConfinement(t *testing.T) {
	openRoot := func(t *testing.T, path string) *os.Root {
		t.Helper()
		root, err := os.OpenRoot(path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { root.Close() })
		return root
	}
	put := func(t *testing.T, path, body string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	link := func(t *testing.T, target, path string) {
		t.Helper()
		if err := os.Symlink(target, path); err != nil {
			t.Fatal(err)
		}
	}

	for _, name := range []string{"project.yaml", "design.yaml", "sweep.yaml", "input.csv", "park2015_advmat_hzo.md"} {
		t.Run(name+" retarget", func(t *testing.T) {
			base := t.TempDir()
			projectDir, citationDir := filepath.Join(base, "project"), filepath.Join(base, "citations")
			for _, dir := range []string{projectDir, citationDir} {
				if err := os.MkdirAll(filepath.Join(dir, "data", "inner"), 0o755); err != nil {
					t.Fatal(err)
				}
				link(t, "data/inner", filepath.Join(dir, "nested"))
			}
			const marker = "PRIVATE-OUTSIDE-MARKER\n"
			outside := filepath.Join(base, "outside")
			put(t, outside, marker)
			digest := fmt.Sprintf("%x", sha256.Sum256([]byte(marker)))
			p := validProjectYAML + fmt.Sprintf("inputs:\n  - path: input.csv\n    sha256: %s\n    citation: park2015_advmat_hzo\n    evidence: experiment-calibrated\n", digest)
			writeBundle(t, filepath.Join(projectDir, "data", "inner"), p, validDesignYAML, validSweepYAML)
			put(t, filepath.Join(projectDir, "data", "inner", "input.csv"), marker)
			put(t, filepath.Join(citationDir, "data", "inner", "park2015_advmat_hzo.md"), "# Park\n")
			for _, file := range []string{"project.yaml", "design.yaml", "sweep.yaml", "input.csv"} {
				link(t, "nested/"+file, filepath.Join(projectDir, file))
			}
			link(t, "nested/park2015_advmat_hzo.md", filepath.Join(citationDir, "park2015_advmat_hzo.md"))
			root, citations := openRoot(t, projectDir), openRoot(t, citationDir)
			got, err := LoadRoot(root, citations)
			if err != nil || len(got.Inputs) != 1 || got.Inputs[0].SHA256 != digest {
				t.Fatalf("nested in-root links: bundle=%+v, err=%v", got, err)
			}

			// Retarget after a successful load, without a scheduling race or sleep.
			dir := projectDir
			if strings.HasSuffix(name, ".md") {
				dir = citationDir
			}
			path := filepath.Join(dir, name)
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			link(t, "../outside", path)
			for _, load := range []func() (Bundle, error){
				func() (Bundle, error) { return LoadRoot(root, citations) },
				func() (Bundle, error) { return Load(projectDir, LoadOptions{CitationDir: citationDir}) },
			} {
				_, err := load()
				var pathErr *os.PathError
				// A traversal error, not a YAML parse/hash error containing marker
				// bytes, proves the outside file was rejected before consumption.
				if !errors.As(err, &pathErr) || !strings.Contains(pathErr.Err.Error(), "escapes") || strings.Contains(err.Error(), marker) {
					t.Fatalf("err=%v, want traversal rejection before reading outside bytes", err)
				}
			}
			if strings.HasSuffix(name, ".md") {
				if _, err := LoadRoot(root, nil); err != nil {
					t.Fatalf("nil citations should skip existence checks: %v", err)
				}
			}
			if body, err := os.ReadFile(outside); err != nil || string(body) != marker {
				t.Fatalf("outside fixture changed: %q, %v", body, err)
			}
			for _, handle := range []*os.Root{root, citations} {
				if _, err := handle.Stat("."); err != nil {
					t.Fatalf("caller handle closed after rejection: %v", err)
				}
			}
		})
	}

	t.Run("root names are metadata", func(t *testing.T) {
		base := t.TempDir()
		projectDir, citationDir := filepath.Join(base, "project"), filepath.Join(base, "citations")
		for _, dir := range []string{projectDir, citationDir} {
			if err := os.Mkdir(dir, 0o755); err != nil {
				t.Fatal(err)
			}
		}
		writeBundle(t, projectDir, validProjectYAML, validDesignYAML, validSweepYAML)
		put(t, filepath.Join(citationDir, "park2015_advmat_hzo.md"), "# Park")
		root, citations := openRoot(t, projectDir), openRoot(t, citationDir)
		for _, dir := range []string{projectDir, citationDir} {
			if err := os.Rename(dir, dir+"-moved"); err != nil {
				t.Fatal(err)
			}
			link(t, t.TempDir(), dir)
		}
		got, err := LoadRoot(root, citations)
		if err != nil || got.Root != root.Name() || got.Project.ID != "hzo-study" {
			t.Fatalf("renamed roots: bundle=%+v, err=%v", got, err)
		}
		if _, err := LoadRoot(root, nil); err != nil {
			t.Fatalf("optional citation checks / caller handle lifetime: %v", err)
		}
		if _, err := citations.Stat("park2015_advmat_hzo.md"); err != nil {
			t.Fatalf("caller citation handle closed: %v", err)
		}
	})

	for _, tc := range []struct{ name, suffix, want string }{
		{"multiple documents", "---\n{}\n", "multiple YAML documents"},
		{"oversized document", "#" + strings.Repeat(" ", 1<<20), "exceeds"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeBundle(t, dir, validProjectYAML+tc.suffix, validDesignYAML, validSweepYAML)
			_, err := LoadRoot(openRoot(t, dir), nil)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err=%v, want %s", err, tc.want)
			}
		})
	}
}

func TestLoadStrictBundle(t *testing.T) {
	root := t.TempDir()
	citations := filepath.Join(root, "citations")
	if err := os.Mkdir(citations, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(citations, "park2015_advmat_hzo.md"), []byte("# Park"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeBundle(t, root, validProjectYAML, validDesignYAML, validSweepYAML)

	got, err := Load(root, LoadOptions{CitationDir: citations})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Project.ID != "hzo-study" || got.Design.Array.Rows != 32 || got.Sweep.Seed != 17 {
		t.Fatalf("unexpected bundle: %+v", got)
	}
}

func TestLoadRejectsUnknownYAMLField(t *testing.T) {
	root := t.TempDir()
	writeBundle(t, root, validProjectYAML+"mystery: true\n", validDesignYAML, validSweepYAML)
	_, err := Load(root, LoadOptions{})
	if err == nil || !strings.Contains(err.Error(), "field mystery not found") {
		t.Fatalf("err=%v, want unknown-field error", err)
	}
}

func TestLoadValidatesInputHashAndContainment(t *testing.T) {
	root := t.TempDir()
	data := []byte("measured values\n")
	if err := os.Mkdir(filepath.Join(root, "inputs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "inputs", "sample.csv"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(data))
	p := validProjectYAML + fmt.Sprintf("inputs:\n  - path: inputs/sample.csv\n    sha256: %s\n    citation: park2015_advmat_hzo\n    evidence: experiment-calibrated\n", digest)
	writeBundle(t, root, p, validDesignYAML, validSweepYAML)

	got, err := Load(root, LoadOptions{})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got.Inputs) != 1 || got.Inputs[0].SHA256 != digest {
		t.Fatalf("inputs=%+v", got.Inputs)
	}

	writeBundle(t, root, strings.Replace(p, digest, strings.Repeat("0", 64), 1), validDesignYAML, validSweepYAML)
	if _, err := Load(root, LoadOptions{}); err == nil || !strings.Contains(err.Error(), "sha256") {
		t.Fatalf("err=%v, want digest mismatch", err)
	}
}

func TestLoadRejectsInputSymlinkEscapingProject(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.csv")
	if err := os.WriteFile(outside, []byte("private\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "inputs"), 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "inputs", "outside.csv")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte("private\n")))
	p := validProjectYAML + fmt.Sprintf("inputs:\n  - path: inputs/outside.csv\n    sha256: %s\n    citation: park2015_advmat_hzo\n    evidence: experiment-calibrated\n", digest)
	writeBundle(t, root, p, validDesignYAML, validSweepYAML)
	if _, err := Load(root, LoadOptions{}); err == nil || !strings.Contains(err.Error(), "escapes") {
		t.Fatalf("err=%v, want containment error", err)
	}
}

func TestLoadRejectsUnexpandableSweep(t *testing.T) {
	root := t.TempDir()
	badSweep := strings.Replace(validSweepYAML, "device.conductance_levels", "device.unsupported", 1)
	writeBundle(t, root, validProjectYAML, validDesignYAML, badSweep)
	if _, err := Load(root, LoadOptions{}); err == nil || !strings.Contains(err.Error(), "unsupported sweep path") {
		t.Fatalf("Load error=%v, want unsupported sweep path", err)
	}
}

func TestLoadRejectsInvalidDesignTrustBoundary(t *testing.T) {
	root := t.TempDir()
	bad := strings.Replace(validDesignYAML, "g_max_s: 0.00003", "g_max_s: .nan", 1)
	writeBundle(t, root, validProjectYAML, bad, validSweepYAML)
	if _, err := Load(root, LoadOptions{}); err == nil {
		t.Fatal("Load succeeded with non-finite conductance")
	}
}
