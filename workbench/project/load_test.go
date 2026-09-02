package project

import (
	"crypto/sha256"
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
	if _, err := Load(root, LoadOptions{}); err == nil || !strings.Contains(err.Error(), "escapes project root") {
		t.Fatalf("err=%v, want containment error", err)
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
