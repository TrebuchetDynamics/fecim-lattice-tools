package edaexport

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/project"
)

func testRun(t *testing.T, configure ...func(*project.Bundle, *experiment.RunOptions)) (project.Bundle, string) {
	t.Helper()
	bundle := project.Bundle{
		Root: t.TempDir(),
		Project: project.Project{
			SchemaVersion: 1,
			ID:            "eda-export-test",
			Hypothesis:    "Export preserves the selected simulated design.",
			ModelVersion:  "test-v1",
			Objectives:    []project.Objective{{Metric: "energy_pj", Direction: project.Minimize}, {Metric: "latency_ns", Direction: project.Minimize}},
		},
		Design: project.Design{
			SchemaVersion: 1,
			Device:        project.Device{Material: "HZO", ConductanceLevels: 30, GMinS: 1e-6, GMaxS: 30e-6},
			Array:         project.Array{Rows: 4, Cols: 4, ReadVoltageV: 0.2},
			Circuit:       project.Circuit{ADCBits: 6, DACBits: 4, TIAGainOhm: 10000, TechNode: "65nm"},
		},
		Sweep: project.Sweep{Seed: 17, MaxPoints: 1, Parameters: []project.Parameter{{Path: "circuit.adc_bits", Values: []float64{6}}}},
	}
	opts := experiment.RunOptions{
		EvaluatorVersion: "test-v1",
		Workers:          1,
		Now:              func() time.Time { return time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC) },
		Evaluator: func(context.Context, project.Design, int64) (experiment.Result, error) {
			return experiment.Result{Status: experiment.StatusSuccess, Metrics: []experiment.Metric{
				{Name: "energy_pj", Value: 1, Unit: "pJ", Model: "fixture", Assumptions: []string{"fixture"}, Evidence: experiment.EvidenceDerived},
				{Name: "latency_ns", Value: 1, Unit: "ns", Model: "fixture", Assumptions: []string{"fixture"}, Evidence: experiment.EvidenceDerived},
			}}, nil
		},
	}
	for _, configure := range configure {
		configure(&bundle, &opts)
	}
	summary, err := experiment.Run(context.Background(), bundle, opts)
	if err != nil {
		t.Fatal(err)
	}
	return bundle, summary.Runs[0].Manifest.RunID
}

func TestExportSelectedRunUsesVerifiedRunAndWritesAllFormats(t *testing.T) {
	bundle, runID := testRun(t)
	dir, err := ExportSelectedRun(bundle.Root, runID)
	if err != nil {
		t.Fatal(err)
	}
	if dir != filepath.Join(bundle.Root, "exports", runID) {
		t.Fatalf("legacy joined destination=%q", dir)
	}
	for _, name := range []string{"design.spice", "design.v", "design.def", "cell.lef", "cell.lib", "manifest.json"} {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil || info.Size() == 0 {
			t.Fatalf("artifact %s size=%v err=%v", name, info, err)
		}
	}
	var manifest struct {
		RunID string            `json:"run_id"`
		Files map[string]string `json:"files_sha256"`
	}
	blob, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(blob, &manifest); err != nil || manifest.RunID != runID || len(manifest.Files) != 5 {
		t.Fatalf("manifest=%+v err=%v", manifest, err)
	}
	if _, err := ExportSelectedRun(bundle.Root, runID); err == nil {
		t.Fatal("repeat export unexpectedly overwrote existing artifacts")
	}
}

func TestExportSelectedRunRootConfinement(t *testing.T) {
	open := func(t *testing.T, path string) *os.Root {
		t.Helper()
		root, err := os.OpenRoot(path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = root.Close() })
		return root
	}
	must := func(t *testing.T, err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	// Only the child has a zero file-size limit; fixtures are persisted beforehand.
	if path := os.Getenv("FECIM_EXPORT_WRITE_FAILURE_ROOT"); path != "" {
		root := open(t, path)
		dir, err := ExportSelectedRunRoot(root, os.Getenv("FECIM_EXPORT_WRITE_FAILURE_ID"))
		if err == nil || dir != "" || !strings.Contains(err.Error(), "write design.spice") {
			t.Fatalf("expected artifact write failure: dir=%q err=%v", dir, err)
		}
		return
	}

	t.Run("verified bytes through moved handle", func(t *testing.T) {
		bundle, id := testRun(t)
		root := open(t, bundle.Root)
		before := make(map[string]string)
		for _, name := range []string{"manifest.json", "resolved-design.json", "result.json"} {
			data, err := root.ReadFile(filepath.Join("runs", id, name))
			must(t, err)
			before[name] = string(data)
		}
		moved := filepath.Join(t.TempDir(), "moved")
		must(t, os.Rename(bundle.Root, moved))
		must(t, os.Mkdir(bundle.Root, 0o755)) // A reopened root.Name() would hit this decoy.
		rel, err := ExportSelectedRunRoot(root, id)
		must(t, err)
		if rel != filepath.Join("exports", id) {
			t.Fatalf("relative destination=%q", rel)
		}
		var manifest struct {
			SchemaVersion int               `json:"schema_version"`
			RunID         string            `json:"run_id"`
			Design        project.Design    `json:"resolved_design"`
			Technology    string            `json:"eda_technology"`
			Warning       string            `json:"warning"`
			Hashes        map[string]string `json:"files_sha256"`
		}
		blob, err := root.ReadFile(filepath.Join(rel, "manifest.json"))
		must(t, err)
		must(t, json.Unmarshal(blob, &manifest))
		if manifest.SchemaVersion != 1 || manifest.RunID != id || !reflect.DeepEqual(manifest.Design, bundle.Design) ||
			manifest.Technology != "SKY130" || !strings.Contains(manifest.Warning, "tech_node=65nm is not mapped to a foundry PDK") || len(manifest.Hashes) != 5 {
			t.Fatalf("export provenance=%+v", manifest)
		}
		for name, fragments := range map[string][]string{
			"design.spice": {"* Array: 16 cells", "X_3_3", "R_level=1000000.00"},
			"design.v":     {"parameter ROWS = 4;", "parameter COLS = 4;", "parameter LEVELS = 30;"},
			"design.def":   {"DESIGN fecim_run_" + id[:12], "COMPONENTS 16"},
			"cell.lef":     {"MACRO fecim_run_" + id[:12] + "_cell"},
			"cell.lib":     {"cell(fecim_run_" + id[:12] + "_cell)"},
		} {
			data, err := root.ReadFile(filepath.Join(rel, name))
			must(t, err)
			digest := sha256.Sum256(data)
			if manifest.Hashes[name] != hex.EncodeToString(digest[:]) {
				t.Errorf("%s hash does not cover actual bytes", name)
			}
			for _, fragment := range fragments {
				if !strings.Contains(string(data), fragment) {
					t.Errorf("%s missing %q", name, fragment)
				}
			}
		}
		for name, want := range before {
			data, err := root.ReadFile(filepath.Join("runs", id, name))
			must(t, err)
			if string(data) != want {
				t.Errorf("export changed source %s", name)
			}
		}
		entries, err := os.ReadDir(bundle.Root)
		must(t, err)
		if len(entries) != 0 {
			t.Fatalf("export wrote to decoy: %v", entries)
		}
	})

	for _, which := range []string{"exports", "runs", "run directory", "manifest.json", "resolved-design.json", "result.json"} {
		t.Run("outside link "+which, func(t *testing.T) {
			bundle, id := testRun(t)
			outside := t.TempDir()
			marker := filepath.Join(outside, "marker")
			must(t, os.WriteFile(marker, []byte("keep"), 0o644))
			link := filepath.Join(bundle.Root, which)
			switch which {
			case "run directory":
				link = filepath.Join(bundle.Root, "runs", id)
			case "manifest.json", "resolved-design.json", "result.json":
				link = filepath.Join(bundle.Root, "runs", id, which)
				data, err := os.ReadFile(link)
				must(t, err)
				must(t, os.WriteFile(filepath.Join(outside, which), data, 0o644))
			}
			must(t, os.RemoveAll(link))
			target := outside
			if strings.HasSuffix(which, ".json") {
				target = filepath.Join(outside, which)
			}
			must(t, os.Symlink(target, link))
			if dir, err := ExportSelectedRunRoot(open(t, bundle.Root), id); err == nil || dir != "" {
				t.Fatalf("outside link accepted: %q, %v", dir, err)
			}
			entries, err := os.ReadDir(outside)
			must(t, err)
			wantEntries := 1
			if strings.HasSuffix(which, ".json") {
				wantEntries++
			}
			data, err := os.ReadFile(marker)
			must(t, err)
			if len(entries) != wantEntries || string(data) != "keep" {
				t.Fatalf("outside changed: %v marker=%q", entries, data)
			}
		})
	}

	for _, which := range []string{"invalid ID", "missing run", "malformed", "tampered", "failed", "zero rows", "zero cols", "oversized", "nil root", "closed root"} {
		t.Run("reject "+which, func(t *testing.T) {
			bundle, id := testRun(t, func(bundle *project.Bundle, opts *experiment.RunOptions) {
				switch which {
				case "zero rows":
					bundle.Design.Array.Rows = 0
				case "zero cols":
					bundle.Design.Array.Cols = 0
				case "oversized":
					bundle.Design.Array.Rows, bundle.Design.Array.Cols = 129, 128
				case "failed":
					opts.Evaluator = func(context.Context, project.Design, int64) (experiment.Result, error) {
						return experiment.Result{Status: experiment.StatusFailed, Failure: &experiment.Failure{Kind: "model-domain", Message: "fixture"}}, nil
					}
				}
			})
			root := open(t, bundle.Root)
			switch which {
			case "invalid ID":
				id = "../outside"
			case "missing run":
				id = strings.Repeat("0", 64)
			case "malformed":
				must(t, os.WriteFile(filepath.Join(bundle.Root, "runs", id, "manifest.json"), []byte("{"), 0o644))
			case "tampered":
				must(t, os.WriteFile(filepath.Join(bundle.Root, "runs", id, "result.json"), []byte(`{"status":"success"}`), 0o644))
			case "nil root":
				root = nil
			case "closed root":
				must(t, root.Close())
			}
			if dir, err := ExportSelectedRunRoot(root, id); err == nil || dir != "" {
				t.Fatalf("invalid input exported: %q, %v", dir, err)
			}
			if _, err := os.Lstat(filepath.Join(bundle.Root, "exports")); !os.IsNotExist(err) {
				t.Fatalf("invalid input created exports: %v", err)
			}
		})
	}

	for _, kind := range []string{"directory", "file", "link"} {
		t.Run("existing destination "+kind, func(t *testing.T) {
			bundle, id := testRun(t)
			dest := filepath.Join(bundle.Root, "exports", id)
			must(t, os.MkdirAll(filepath.Dir(dest), 0o755))
			marker := dest
			switch kind {
			case "directory":
				must(t, os.Mkdir(dest, 0o755))
				marker = filepath.Join(dest, "marker")
			case "link":
				marker = filepath.Join(t.TempDir(), "marker")
				must(t, os.Symlink(marker, dest))
			}
			must(t, os.WriteFile(marker, []byte("keep"), 0o644))
			info, err := os.Lstat(dest)
			must(t, err)
			if _, err := ExportSelectedRunRoot(open(t, bundle.Root), id); err == nil {
				t.Fatal("existing destination overwritten")
			}
			after, err := os.Lstat(dest)
			must(t, err)
			data, err := os.ReadFile(marker)
			must(t, err)
			if !os.SameFile(info, after) || string(data) != "keep" {
				t.Fatal("existing destination or marker changed")
			}
		})
	}

	t.Run("write failure cleans only own destination", func(t *testing.T) {
		if runtime.GOOS != "linux" {
			t.Skip("file-size fault is qualified on Linux")
		}
		bundle, id := testRun(t)
		exports := filepath.Join(bundle.Root, "exports")
		must(t, os.Mkdir(exports, 0o755))
		marker := filepath.Join(exports, "unrelated")
		must(t, os.WriteFile(marker, []byte("keep"), 0o644))
		binary, err := os.Executable()
		must(t, err)
		cmd := exec.Command("bash", "-c", `ulimit -f 0; exec "$1" -test.run '^TestExportSelectedRunRootConfinement$'`, "export-write-failure", binary)
		cmd.Env = append(os.Environ(), "FECIM_EXPORT_WRITE_FAILURE_ROOT="+bundle.Root, "FECIM_EXPORT_WRITE_FAILURE_ID="+id)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("fault child: %v\n%s", err, output)
		}
		entries, err := os.ReadDir(exports)
		must(t, err)
		data, err := os.ReadFile(marker)
		must(t, err)
		if len(entries) != 1 || entries[0].Name() != "unrelated" || string(data) != "keep" {
			t.Fatalf("incomplete destination or damaged marker: %v, %q", entries, data)
		}
		_, err = ExportSelectedRunRoot(open(t, bundle.Root), id)
		must(t, err) // Failure must leave this ID available for a clean retry.
	})
}

func TestExportSelectedRunRejectsTamperedRun(t *testing.T) {
	bundle, runID := testRun(t)
	path := filepath.Join(bundle.Root, "runs", runID, "result.json")
	if err := os.WriteFile(path, []byte(`{"status":"success"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ExportSelectedRun(bundle.Root, runID); err == nil {
		t.Fatal("tampered run unexpectedly exported")
	}
	if _, err := os.Stat(filepath.Join(bundle.Root, "exports")); !os.IsNotExist(err) {
		t.Fatalf("failed export created output directory: %v", err)
	}
}
