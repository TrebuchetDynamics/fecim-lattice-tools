package project

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// This table catches loss of byte identity, typed-edit validation and preservation
// at the preparation seam; preparation must never mutate the source files.
func TestPrepareConfiguration(t *testing.T) {
	type testCase struct {
		name, document         string
		project, design, sweep string
		changes                []FieldChange
		wantError, changed     bool
		check                  func(*testing.T, PreparedConfiguration)
	}
	cases := []testCase{
		{name: "empty CRLF quoted no final newline", project: strings.ReplaceAll(strings.TrimSuffix(strings.Replace(validProjectYAML, "name: HZO study", "name: 'HZO study' # keep inline", 1), "\n"), "\n", "\r\n")},
		{name: "same string CRLF quoted no final newline", project: strings.ReplaceAll(strings.TrimSuffix(strings.Replace(validProjectYAML, "name: HZO study", "name: 'HZO study' # keep inline", 1), "\n"), "\n", "\r\n"), changes: []FieldChange{{"/name", "HZO study"}}},
		{name: "same integer", document: "design.yaml", design: strings.Replace(validDesignYAML, "rows: 32", "rows: 0x20 # hex stays", 1), changes: []FieldChange{{"/array/rows", "032"}}},
		{name: "same float", document: "design.yaml", changes: []FieldChange{{"/array/read_voltage_v", "2e-1"}}},
		{name: "literal string and preservation", project: "# header\n" + strings.Replace(validProjectYAML, "name: HZO study", "name: 'HZO study' # keep inline", 1), changes: []FieldChange{{"/name", "false: [not YAML]"}, {"/hypothesis", "A literal hypothesis"}}, changed: true, check: func(t *testing.T, p PreparedConfiguration) {
			if p.bundle.Project.Name != "false: [not YAML]" || p.bundle.Project.Hypothesis != "A literal hypothesis" || p.bundle.Project.ModelVersion != "fecim-system-v1" || p.bundle.Project.Constraints[0].Unit != "um2" || !strings.Contains(p.snapshot.documents[0], "# header") || !strings.Contains(p.snapshot.documents[0], "# keep inline") {
				t.Fatalf("lost literal string, comments or unexposed fields: %+v / %s", p.bundle.Project, p.snapshot.documents[0])
			}
		}},
		{name: "constraint scalars", changes: []FieldChange{{"/constraints/0/value", "4e3"}, {"/constraints/0/operator", ">="}}, changed: true, check: func(t *testing.T, p PreparedConfiguration) {
			if c := p.bundle.Project.Constraints[0]; c.Value != 4000 || c.Operator != ">=" {
				t.Fatalf("constraint=%+v", c)
			}
		}},
		{name: "design scalars", document: "design.yaml", changes: []FieldChange{{"/device/material", "BTO"}, {"/device/conductance_levels", "31"}, {"/device/g_min_s", "1e-7"}, {"/device/g_max_s", "4e-5"}, {"/array/rows", "64"}, {"/array/cols", "64"}, {"/array/read_voltage_v", "0.3"}, {"/circuit/adc_bits", "8"}, {"/circuit/dac_bits", "8"}, {"/circuit/tia_gain_ohm", "20000"}, {"/circuit/tech_node", "28nm"}}, changed: true, check: func(t *testing.T, p PreparedConfiguration) {
			want := Design{1, Device{"BTO", 31, 1e-7, 4e-5}, Array{64, 64, 0.3}, Circuit{8, 8, 20000, "28nm"}}
			if p.bundle.Design != want {
				t.Fatalf("design=%+v", p.bundle.Design)
			}
		}},
		{name: "sweep scalars", document: "sweep.yaml", changes: []FieldChange{{"/seed", "-9223372036854775808"}, {"/max_points", "64"}, {"/parameters/0/values/1", "32"}, {"/parameters/1/range/start", "8"}, {"/parameters/1/range/stop", "64"}, {"/parameters/1/range/count", "3"}}, changed: true, check: func(t *testing.T, p PreparedConfiguration) {
			if s := p.bundle.Sweep; s.Seed != -9223372036854775808 || s.MaxPoints != 64 || s.Parameters[0].Values[1] != 32 || *s.Parameters[1].Range != (LinearRange{8, 64, 3}) {
				t.Fatalf("sweep=%+v", s)
			}
		}},
		{name: "alias empty", project: strings.Replace(validProjectYAML, "name: HZO study\nhypothesis: Increasing ADC resolution trades area for energy fidelity.", "name: &title HZO study\nhypothesis: *title", 1)},
		{name: "alias nonempty", project: strings.Replace(validProjectYAML, "name: HZO study\nhypothesis: Increasing ADC resolution trades area for energy fidelity.", "name: &title HZO study\nhypothesis: *title", 1), changes: []FieldChange{{"/name", "HZO study"}}, wantError: true},
		{name: "merge nonempty", design: strings.Replace(validDesignYAML, "  cols: 32", "  <<: {cols: 32}", 1), changes: []FieldChange{{"/name", "new"}}, wantError: true},
		{name: "duplicate keys", project: validProjectYAML + "name: duplicate\n", wantError: true},
		{name: "unknown schema field", design: validDesignYAML + "mystery: 1\n", wantError: true},
		{name: "invalid full bundle", design: strings.Replace(validDesignYAML, "g_max_s: 0.00003", "g_max_s: 0", 1), wantError: true},
		{name: "multiple documents", project: validProjectYAML + "---\n{}\n", wantError: true},
		{name: "invalid YAML", project: "name: [", wantError: true},
		{name: "unknown document", document: "../project.yaml", wantError: true},
		{name: "duplicate targets", changes: []FieldChange{{"/name", "a"}, {"/name", "b"}}, wantError: true},
		{name: "too many changes", changes: make([]FieldChange, 65), wantError: true},
		{name: "oversized scalar", changes: []FieldChange{{"/name", strings.Repeat("é", 2049)}}, wantError: true},
		{name: "invalid UTF8", changes: []FieldChange{{"/name", string([]byte{0xff})}}, wantError: true},
		{name: "nonfinite", document: "design.yaml", changes: []FieldChange{{"/array/read_voltage_v", "NaN"}}, wantError: true},
		{name: "float overflow", document: "design.yaml", changes: []FieldChange{{"/array/read_voltage_v", "1e999"}}, wantError: true},
		{name: "integer overflow", document: "sweep.yaml", changes: []FieldChange{{"/seed", "9223372036854775808"}}, wantError: true},
		{name: "sweep integer value overflow", document: "sweep.yaml", changes: []FieldChange{{"/parameters/0/values/0", "1e20"}}, wantError: true},
		{name: "sweep integer range overflow", document: "sweep.yaml", changes: []FieldChange{{"/parameters/1/range/stop", "9223372036854775808"}}, wantError: true},
		{name: "sweep integer range underflow", document: "sweep.yaml", changes: []FieldChange{{"/parameters/1/range/start", "-1e20"}}, wantError: true},
		{name: "float sweep is not int bounded", document: "sweep.yaml", sweep: strings.Replace(validSweepYAML, "device.conductance_levels", "device.g_max_s", 1), changes: []FieldChange{{"/parameters/0/values/0", "1e20"}}, changed: true},
		{name: "integer YAML input", document: "design.yaml", changes: []FieldChange{{"/array/rows", "0x20"}}, wantError: true},
		{name: "integer fractional input", document: "design.yaml", changes: []FieldChange{{"/array/rows", "32.5"}}, wantError: true},
		{name: "invalid candidate", document: "design.yaml", changes: []FieldChange{{"/device/g_min_s", "1"}}, wantError: true},
		{name: "range allocation preflight", sweep: strings.Replace(validSweepYAML, "count: 2", "count: 9223372036854775807", 1), wantError: true},
		{name: "values allocation preflight", sweep: strings.Replace(validSweepYAML, "max_points: 32", "max_points: 1", 1), wantError: true},
		{name: "range nonfinite", sweep: strings.Replace(validSweepYAML, "start: 16", "start: .inf", 1), wantError: true},
		{name: "range arithmetic overflow", sweep: strings.Replace(validSweepYAML, "start: 16, stop: 32", "start: -1e308, stop: 1e308", 1), wantError: true},
		{name: "range zero count", sweep: strings.Replace(validSweepYAML, "count: 2", "count: 0", 1), wantError: true},
		{name: "default max points", document: "sweep.yaml", changes: []FieldChange{{"/max_points", "0"}}, changed: true},
		{name: "output byte limit", project: validProjectYAML + "#" + strings.Repeat("x", (1<<20)-len(validProjectYAML)-2) + "\n", changes: []FieldChange{{"/name", strings.Repeat("x", 4096)}}, wantError: true},
	}
	for _, path := range []string{"name", "/id", "/model_version", "/schema_version", "/objectives/0/metric", "/constraints/0/unit", "/constraints/01/value", "/constraints/-1/value", "/constraints/-/value", "/constraints/99/value", "/constraints/0", "/name/child", "/na~2me", "/na~", "/na~1me", "/na~0me"} {
		cases = append(cases, testCase{name: "reject path " + path, changes: []FieldChange{{path, "1"}}, wantError: true})
	}
	cases = append(cases, testCase{name: "missing existing scalar", project: strings.Replace(validProjectYAML, "name: HZO study\n", "", 1), changes: []FieldChange{{"/name", "new"}}, wantError: true})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.project == "" {
				tc.project = validProjectYAML
			}
			if tc.design == "" {
				tc.design = validDesignYAML
			}
			if tc.sweep == "" {
				tc.sweep = validSweepYAML
			}
			if tc.document == "" {
				tc.document = "project.yaml"
			}
			dir := t.TempDir()
			writeBundle(t, dir, tc.project, tc.design, tc.sweep)
			root, err := os.OpenRoot(dir)
			if err != nil {
				t.Fatal(err)
			}
			defer root.Close()
			base, err := ReadConfiguration(root)
			if err != nil {
				t.Fatal(err)
			}
			original := base
			got, err := PrepareConfiguration(root, nil, base, tc.document, tc.changes)
			if (err != nil) != tc.wantError {
				t.Fatalf("PrepareConfiguration err=%v, wantError=%v", err, tc.wantError)
			}
			if base != original {
				t.Fatal("snapshot mutated")
			}
			for i, name := range []string{"project.yaml", "design.yaml", "sweep.yaml"} {
				data, err := root.ReadFile(name)
				if err != nil || string(data) != base.documents[i] {
					t.Fatalf("disk changed: %s / %v", name, err)
				}
				if !tc.wantError && (name != tc.document || !tc.changed) && got.snapshot.documents[i] != string(data) {
					t.Fatalf("untouched bytes changed: %s", name)
				}
			}
			if _, err := root.Stat("."); err != nil {
				t.Fatalf("caller root closed: %v", err)
			}
			if tc.wantError {
				return
			}
			if got.Changed() != tc.changed || (got.snapshot.Revision() != base.Revision()) != tc.changed {
				t.Fatalf("changed=%v, base=%s prepared=%s", got.Changed(), base.Revision(), got.snapshot.Revision())
			}
			if tc.check != nil {
				tc.check(t, got)
			}
		})
	}

	t.Run("framed original byte revision", func(t *testing.T) {
		dir := t.TempDir()
		writeBundle(t, dir, validProjectYAML, validDesignYAML, validSweepYAML)
		root, err := os.OpenRoot(dir)
		if err != nil {
			t.Fatal(err)
		}
		defer root.Close()
		var frame bytes.Buffer
		frame.WriteString("fecim-config-v1\x00")
		for i, name := range []string{"project.yaml", "design.yaml", "sweep.yaml"} {
			body := []string{validProjectYAML, validDesignYAML, validSweepYAML}[i]
			binary.Write(&frame, binary.BigEndian, uint64(len(name)))
			frame.WriteString(name)
			binary.Write(&frame, binary.BigEndian, uint64(len(body)))
			frame.WriteString(body)
		}
		base, err := ReadConfiguration(root)
		if err != nil {
			t.Fatal(err)
		}
		if want := fmt.Sprintf(`"cfg-v1-%x"`, sha256.Sum256(frame.Bytes())); base.Revision() != want {
			t.Fatalf("revision=%s want=%s", base.Revision(), want)
		}
		for _, name := range []string{"project.yaml", "design.yaml", "sweep.yaml"} {
			body, _ := root.ReadFile(name)
			if err := os.WriteFile(filepath.Join(dir, name), append(body, []byte("# byte-only change\n")...), 0o600); err != nil {
				t.Fatal(err)
			}
			next, err := ReadConfiguration(root)
			if err != nil || next.Revision() == base.Revision() {
				t.Fatalf("revision unchanged for %s: %v", name, err)
			}
			if err := os.WriteFile(filepath.Join(dir, name), body, 0o600); err != nil {
				t.Fatal(err)
			}
		}
	})

	t.Run("overlay retains input and citation root authority", func(t *testing.T) {
		parent := t.TempDir()
		dir, citationDir := filepath.Join(parent, "project"), filepath.Join(parent, "citations")
		for _, path := range []string{dir, citationDir} {
			if err := os.Mkdir(path, 0o700); err != nil {
				t.Fatal(err)
			}
		}
		data := []byte("calibration\n")
		project := validProjectYAML + fmt.Sprintf("inputs:\n  - path: input.csv\n    sha256: %x\n    evidence: experiment-calibrated\n", sha256.Sum256(data))
		writeBundle(t, dir, project, validDesignYAML, validSweepYAML)
		if err := os.WriteFile(filepath.Join(dir, "input.csv"), data, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(citationDir, "park2015_advmat_hzo.md"), []byte("# Park"), 0o600); err != nil {
			t.Fatal(err)
		}
		root, err := os.OpenRoot(dir)
		if err != nil {
			t.Fatal(err)
		}
		defer root.Close()
		citations, err := os.OpenRoot(citationDir)
		if err != nil {
			t.Fatal(err)
		}
		defer citations.Close()
		base, err := ReadConfiguration(root)
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range []string{dir, citationDir} {
			if err := os.Rename(path, path+"-moved"); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(path, 0o700); err != nil {
				t.Fatal(err)
			}
		}
		patch := []FieldChange{{"/name", "Changed"}}
		got, err := PrepareConfiguration(root, citations, base, "project.yaml", patch)
		if err != nil || len(got.bundle.Inputs) != 1 {
			t.Fatalf("renamed roots: inputs=%v, err=%v", got.bundle.Inputs, err)
		}
		if err := os.WriteFile(filepath.Join(dir+"-moved", "input.csv"), []byte("tampered"), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := PrepareConfiguration(root, citations, base, "project.yaml", nil); err == nil || !strings.Contains(err.Error(), "sha256") {
			t.Fatalf("input hash not revalidated: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir+"-moved", "input.csv"), data, 0o600); err != nil {
			t.Fatal(err)
		}
		for _, target := range []struct{ directory, name string }{{dir + "-moved", "input.csv"}, {citationDir + "-moved", "park2015_advmat_hzo.md"}} {
			path := filepath.Join(target.directory, target.name)
			if err := os.Rename(path, path+".real"); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink("../outside", path); err != nil {
				t.Fatal(err)
			}
			if _, err := PrepareConfiguration(root, citations, base, "project.yaml", patch); err == nil || !strings.Contains(err.Error(), "escapes") {
				t.Fatalf("root escape accepted: %v", err)
			}
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if err := os.Rename(path+".real", path); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.Remove(filepath.Join(citationDir+"-moved", "park2015_advmat_hzo.md")); err != nil {
			t.Fatal(err)
		}
		if _, err := PrepareConfiguration(root, citations, base, "project.yaml", patch); err == nil {
			t.Fatal("missing citation accepted")
		}
		if _, err := PrepareConfiguration(root, nil, base, "project.yaml", patch); err != nil {
			t.Fatalf("nil citations: %v", err)
		}
		for _, handle := range []*os.Root{root, citations} {
			if _, err := handle.Stat("."); err != nil {
				t.Fatalf("caller handle closed: %v", err)
			}
		}
	})

	for _, kind := range []string{"symlink", "directory", "oversized", "missing", "nil root", "zero snapshot"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			writeBundle(t, dir, validProjectYAML, validDesignYAML, validSweepYAML)
			root, err := os.OpenRoot(dir)
			if err != nil {
				t.Fatal(err)
			}
			defer root.Close()
			path := filepath.Join(dir, "project.yaml")
			switch kind {
			case "symlink":
				if err := os.Rename(path, path+".real"); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink("project.yaml.real", path); err != nil {
					t.Fatal(err)
				}
			case "directory":
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(path, 0o700); err != nil {
					t.Fatal(err)
				}
			case "oversized":
				if err := os.WriteFile(path, []byte(strings.Repeat("x", (1<<20)+1)), 0o600); err != nil {
					t.Fatal(err)
				}
			case "missing":
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			case "nil root":
				if _, err := ReadConfiguration(nil); err == nil {
					t.Fatal("nil root accepted")
				}
				return
			case "zero snapshot":
				if _, err := PrepareConfiguration(root, nil, ConfigurationSnapshot{}, "project.yaml", nil); err == nil {
					t.Fatal("zero snapshot accepted")
				}
				return
			}
			if _, err := ReadConfiguration(root); err == nil {
				t.Fatal("invalid configuration file accepted")
			}
		})
	}
}
