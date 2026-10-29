package study

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/fecim"
	"fecim-lattice-tools/workbench/project"
	"fecim-lattice-tools/workbench/report"
)

// Reopening metadata paths or accepting a second citation authority must fail
// these checks against real project files and untouched outside directories.
func TestStudyRootOperations(t *testing.T) {
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	open := func(path string) *os.Root {
		t.Helper()
		r, err := os.OpenRoot(path)
		must(err)
		t.Cleanup(func() { _ = r.Close() })
		return r
	}
	fixture := func() (*os.Root, *os.Root) {
		t.Helper()
		r := open(writeProjectFixture(t))
		data, err := r.ReadFile("project.yaml")
		must(err)
		must(r.WriteFile("project.yaml", append(data, []byte("citations: [fixture]\n")...), 0o600))
		c := open(t.TempDir())
		must(c.WriteFile("fixture.md", []byte("citation"), 0o600))
		return r, c
	}
	unchanged := func(path, want string) {
		t.Helper()
		entries, err := os.ReadDir(path)
		must(err)
		data, err := os.ReadFile(filepath.Join(path, "marker"))
		must(err)
		if len(entries) != 1 || string(data) != want {
			t.Fatalf("outside directory changed: %s", path)
		}
	}

	t.Run("accepted bundle survives later configuration changes", func(t *testing.T) {
		root, citations := fixture()
		bundle, count, err := ValidateProjectRoot(root, citations)
		must(err)
		must(root.WriteFile("sweep.yaml", []byte("not a valid sweep\n"), 0600))
		must(root.WriteFile("project.yaml", []byte("not a valid project\n"), 0600))
		out, err := RunBundleRoot(context.Background(), root, citations, bundle, RunOptions{Workers: 1})
		must(err)
		if out.ResultsState != ResultsComplete || !out.AnalysisAvailable || len(out.Summary.Runs) != count || !reflect.DeepEqual(out.Bundle, bundle) {
			t.Fatalf("captured bundle not executed: %+v", out)
		}
		if bad, err := RunBundleRoot(context.Background(), nil, citations, bundle, RunOptions{Workers: 1}); err == nil || bad.AnalysisAvailable {
			t.Fatal("nil root accepted by prepared execution")
		}
	})
	t.Run("renamed handles across evaluation and reuse", func(t *testing.T) {
		root, citations := fixture()
		bundle, count, err := ValidateProjectRoot(root, citations)
		must(err)
		if count != 2 || bundle.Project.ID != "hzo-study" {
			t.Fatalf("validation = %+v, %d", bundle, count)
		}
		// Deterministic evaluator barrier: swap both paths after preflight but
		// before committing results, reconciliation, and automatic reporting.
		moved := false
		evaluate := func(ctx context.Context, design project.Design, seed int64) (experiment.Result, error) {
			if !moved {
				for _, r := range []*os.Root{root, citations} {
					if err := os.Rename(r.Name(), filepath.Join(t.TempDir(), "moved")); err != nil {
						return experiment.Result{}, err
					}
					if err := os.Mkdir(r.Name(), 0o700); err != nil {
						return experiment.Result{}, err
					}
					if err := os.WriteFile(filepath.Join(r.Name(), "marker"), []byte("outside"), 0o600); err != nil {
						return experiment.Result{}, err
					}
				}
				moved = true
			}
			return fecim.Evaluate(ctx, design, seed)
		}
		progress := make(chan experiment.ProgressSnapshot, 8)
		out, err := runRoot(context.Background(), root, citations, RunOptions{Workers: 1, Progress: progress}, evaluate, report.GenerateRoot)
		must(err)
		if !moved || out.ResultsState != ResultsComplete || !out.AnalysisAvailable || len(out.Summary.Runs) != 2 {
			t.Fatalf("outcome = %+v, moved=%v", out, moved)
		}
		close(progress)
		var last experiment.ProgressSnapshot
		for p := range progress {
			last = p
		}
		if last != (experiment.ProgressSnapshot{Total: 2, Completed: 2}) {
			t.Fatalf("progress = %+v", last)
		}
		id := out.Summary.Runs[0].Manifest.RunID
		before := make(map[string]string)
		for _, name := range []string{"manifest.json", "resolved-design.json", "result.json"} {
			data, err := root.ReadFile(filepath.Join("runs", id, name))
			must(err)
			before[name] = string(data)
		}
		_, count, err = ValidateProjectRoot(root, citations)
		must(err)
		if count != 2 {
			t.Fatalf("moved validation count = %d", count)
		}
		unread := make(chan experiment.ProgressSnapshot)
		out, err = RunRoot(context.Background(), root, citations, RunOptions{Workers: 1, Progress: unread})
		must(err)
		close(unread) // Rooted operations do not own the notification channel.
		for _, record := range out.Summary.Runs {
			if !record.Reused {
				t.Fatal("lost reuse flag")
			}
		}
		analysis, err := GenerateReportRoot(root, citations)
		must(err)
		if len(analysis.Runs) != 2 {
			t.Fatalf("report runs = %d", len(analysis.Runs))
		}
		rel, err := ExportRunRoot(root, id)
		must(err)
		if rel != filepath.Join("exports", id) {
			t.Fatalf("export path = %q", rel)
		}
		for _, name := range []string{"reports/report.md", "reports/results.json", "reports/results.csv", filepath.Join(rel, "manifest.json"), filepath.Join(rel, "design.spice")} {
			data, err := root.ReadFile(name)
			must(err)
			if len(data) == 0 {
				t.Fatalf("empty artifact: %s", name)
			}
		}
		for name, want := range before {
			data, err := root.ReadFile(filepath.Join("runs", id, name))
			must(err)
			if string(data) != want {
				t.Fatalf("canonical record changed: %s", name)
			}
		}
		for _, r := range []*os.Root{root, citations} {
			_, err := r.Stat(".")
			must(err)
			unchanged(r.Name(), "outside")
		}
	})

	t.Run("invalid and conflicting authority", func(t *testing.T) {
		root, citations := fixture()
		out, err := RunRoot(context.Background(), root, citations, RunOptions{Workers: 1, CitationDir: citations.Name()})
		if err == nil || out.ResultsState != ResultsUnavailable || out.AnalysisAvailable {
			t.Fatalf("citation conflict = %+v, %v", out, err)
		}
		if _, err := root.Stat("runs"); !os.IsNotExist(err) {
			t.Fatalf("conflict wrote runs: %v", err)
		}
		_, _, err = ValidateProjectRoot(root, nil)
		must(err)
		closed := open(t.TempDir())
		must(closed.Close())
		withoutKeys := open(writeProjectFixture(t))
		for _, pair := range [][2]*os.Root{{nil, citations}, {closed, citations}, {withoutKeys, closed}} {
			if _, _, err := ValidateProjectRoot(pair[0], pair[1]); err == nil {
				t.Fatal("invalid authority validated")
			}
			out, err := RunRoot(context.Background(), pair[0], pair[1], RunOptions{Workers: 1})
			if err == nil || out.ResultsState != ResultsUnavailable || out.AnalysisAvailable {
				t.Fatalf("invalid authority run = %+v, %v", out, err)
			}
			if _, err := GenerateReportRoot(pair[0], pair[1]); err == nil {
				t.Fatal("invalid authority reported")
			}
		}
		for _, r := range []*os.Root{nil, closed} {
			if _, err := ExportRunRoot(r, strings.Repeat("a", 64)); err == nil {
				t.Fatal("invalid authority exported")
			}
		}
		for _, r := range []*os.Root{root, citations} {
			_, err := r.Stat(".")
			must(err)
		}
	})

	for _, name := range []string{"project.yaml", "fixture.md", "runs", "reports", "exports"} {
		t.Run("outside link "+name, func(t *testing.T) {
			root, citations := fixture()
			outside := t.TempDir()
			marker := "outside"
			if name == "project.yaml" {
				// Valid outside YAML ensures an unauthorized read cannot hide
				// behind a parsing/validation failure.
				data, err := root.ReadFile(name)
				must(err)
				marker = string(data)
			}
			must(os.WriteFile(filepath.Join(outside, "marker"), []byte(marker), 0o600))
			var id string
			if name == "reports" || name == "exports" {
				out, err := RunRoot(context.Background(), root, citations, RunOptions{Workers: 1})
				must(err)
				id = out.Summary.Runs[0].Manifest.RunID
				must(root.RemoveAll(name))
			}
			r := root
			if name == "fixture.md" {
				r = citations
			}
			target := outside
			if name == "project.yaml" || name == "fixture.md" {
				must(r.Remove(name))
				target = filepath.Join(outside, "marker")
			}
			must(r.Symlink(target, name))
			if name == "exports" {
				if _, err := ExportRunRoot(root, id); err == nil {
					t.Fatal("escaping export accepted")
				}
			} else {
				out, err := RunRoot(context.Background(), root, citations, RunOptions{Workers: 1})
				if err == nil || (name == "reports" && (out.ResultsState != ResultsComplete || !out.AnalysisAvailable)) {
					t.Fatalf("escaping study = %+v, %v", out, err)
				}
				if _, err := GenerateReportRoot(root, citations); err == nil {
					t.Fatal("escaping report accepted")
				}
			}
			unchanged(outside, marker)
			for _, r := range []*os.Root{root, citations} {
				_, err := r.Stat(".")
				must(err)
			}
		})
	}
}

func TestRunRetainsPartialOutcome(t *testing.T) {
	for _, mode := range []string{"options", "load", "validation", "cached integrity"} {
		t.Run(mode+" unavailable", func(t *testing.T) {
			root := writeProjectFixture(t)
			opts := RunOptions{Workers: 1}
			switch mode {
			case "options":
				opts.Workers = 0
			case "load":
				root = filepath.Join(root, "missing")
			case "validation":
				if err := os.WriteFile(filepath.Join(root, "sweep.yaml"), []byte("schema_version: 1\nparameters: []\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			case "cached integrity":
				bundle, _, err := ValidateProject(root, "")
				if err != nil {
					t.Fatal(err)
				}
				points, err := experiment.Expand(bundle)
				if err != nil {
					t.Fatal(err)
				}
				id, err := experiment.ID(points[0], fecim.EvaluatorVersion, bundle.Inputs)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(filepath.Join(root, "runs", id), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			outcome, err := Run(context.Background(), root, opts)
			if err == nil || outcome.ResultsState != ResultsUnavailable || outcome.AnalysisAvailable || outcome.Analysis.Counts != nil {
				t.Fatalf("unavailable outcome = %+v, error = %v", outcome, err)
			}
		})
	}

	for _, adapter := range []string{"legacy", "rooted"} {
		t.Run("empty cancellation/"+adapter, func(t *testing.T) {
			root := writeProjectFixture(t)
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			var outcome RunOutcome
			var err error
			if adapter == "rooted" {
				handle, openErr := os.OpenRoot(root)
				if openErr != nil {
					t.Fatal(openErr)
				}
				defer handle.Close()
				outcome, err = RunRoot(ctx, handle, nil, RunOptions{Workers: 1})
			} else {
				outcome, err = Run(ctx, root, RunOptions{Workers: 1})
			}
			if !errors.Is(err, context.Canceled) || outcome.Bundle.Root != root || outcome.ResultsState != ResultsPartial || !outcome.AnalysisAvailable {
				t.Fatalf("empty partial outcome = %+v, error = %v", outcome, err)
			}
			want := map[string]int{"success": 0, "failed": 0, "feasible": 0, "infeasible": 0, "unusable": 0, "pareto": 0}
			if len(outcome.Summary.Runs) != 0 || !reflect.DeepEqual(outcome.Analysis.Counts, want) {
				t.Fatalf("empty partial summary/analysis = %+v / %+v", outcome.Summary, outcome.Analysis)
			}
		})
	}

	for _, tc := range []struct {
		mode      string
		state     OutcomeResultsState
		runs      int
		available bool
	}{
		{"cancelled", ResultsPartial, 2, true},
		{"systemic", ResultsPartial, 2, true},
		{"reconciliation integrity", ResultsPartial, 2, false},
		{"late cancellation", ResultsComplete, 3, true},
		{"report failure", ResultsComplete, 3, true},
	} {
		for _, adapter := range []string{"legacy", "rooted"} {
			t.Run(tc.mode+"/"+adapter, func(t *testing.T) {
				root := writeProjectFixture(t)
				if err := os.WriteFile(filepath.Join(root, "sweep.yaml"), []byte("schema_version: 1\nseed: 17\nmax_points: 8\nparameters:\n  - path: circuit.adc_bits\n    values: [2, 4, 6]\n"), 0o600); err != nil {
					t.Fatal(err)
				}
				bundle, _, err := ValidateProject(root, "")
				if err != nil {
					t.Fatal(err)
				}
				// Seed one reused current point and one unrelated old point.
				commitPoint := func(design project.Design, seed int64, failed bool) (experiment.RunRecord, error) {
					t.Helper()
					single := bundle
					single.Design = design
					single.Sweep = project.Sweep{SchemaVersion: 1, Seed: seed, MaxPoints: 1, Parameters: []project.Parameter{{Path: "circuit.adc_bits", Values: []float64{float64(design.Circuit.ADCBits)}}}}
					summary, err := experiment.Run(context.Background(), single, experiment.RunOptions{Workers: 1, EvaluatorVersion: fecim.EvaluatorVersion, Evaluator: func(ctx context.Context, design project.Design, seed int64) (experiment.Result, error) {
						if failed {
							return experiment.Result{Status: experiment.StatusFailed, Failure: &experiment.Failure{Kind: "domain", Message: "test failure"}}, nil
						}
						return fecim.Evaluate(ctx, design, seed)
					}})
					if err != nil {
						return experiment.RunRecord{}, err
					}
					if len(summary.Runs) != 1 {
						return experiment.RunRecord{}, errors.New("fixture did not commit exactly one point")
					}
					return summary.Runs[0], nil
				}
				first := bundle.Design
				first.Circuit.ADCBits = 2
				cached, err := commitPoint(first, 17, false)
				if err != nil {
					t.Fatal(err)
				}
				old, err := commitPoint(bundle.Design, 999, false)
				if err != nil {
					t.Fatal(err)
				}
				// Unrelated corrupt storage must not poison the active sweep.
				if err := os.WriteFile(filepath.Join(root, "runs", old.Manifest.RunID, "result.json"), []byte("tampered"), 0o600); err != nil {
					t.Fatal(err)
				}
				marker := filepath.Join(root, "reports", "report.md")
				if err := os.MkdirAll(filepath.Dir(marker), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(marker, []byte("previous report"), 0o600); err != nil {
					t.Fatal(err)
				}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				systemErr := errors.New("systemic evaluator failure")
				reportErr := errors.New("report writer failure")
				evaluate := func(ctx context.Context, design project.Design, seed int64) (experiment.Result, error) {
					switch {
					case design.Circuit.ADCBits == 4 && (tc.mode == "cancelled" || tc.mode == "systemic"):
						// Commit via the real storage seam without delivering a runner record.
						if _, err := commitPoint(design, seed, true); err != nil {
							return experiment.Result{}, err
						}
						if tc.mode == "cancelled" {
							cancel()
							return experiment.Result{}, context.Canceled
						}
						return experiment.Result{}, systemErr
					case design.Circuit.ADCBits == 6 && tc.mode == "reconciliation integrity":
						id, err := experiment.ID(experiment.DesignPoint{Design: design, Seed: seed, ModelVersion: bundle.Project.ModelVersion}, fecim.EvaluatorVersion, bundle.Inputs)
						if err != nil {
							return experiment.Result{}, err
						}
						if err := os.MkdirAll(filepath.Join(root, "runs", id), 0o700); err != nil {
							return experiment.Result{}, err
						}
						return experiment.Result{}, systemErr
					case design.Circuit.ADCBits == 6 && tc.mode == "late cancellation":
						if _, err := commitPoint(design, seed, false); err != nil {
							return experiment.Result{}, err
						}
						cancel()
						return experiment.Result{}, context.Canceled
					default:
						return fecim.Evaluate(ctx, design, seed)
					}
				}
				writeReport := report.GenerateRoot
				if tc.mode == "report failure" {
					writeReport = func(*os.Root, project.Bundle, experiment.Analysis) error { return reportErr }
				}
				var outcome RunOutcome
				if adapter == "rooted" {
					handle, openErr := os.OpenRoot(root)
					if openErr != nil {
						t.Fatal(openErr)
					}
					defer handle.Close()
					outcome, err = runRoot(ctx, handle, nil, RunOptions{Workers: 1}, evaluate, writeReport)
				} else {
					outcome, err = run(ctx, root, RunOptions{Workers: 1}, evaluate, writeReport)
				}
				wantErr := systemErr
				if strings.Contains(tc.mode, "cancel") {
					wantErr = context.Canceled
				} else if tc.mode == "report failure" {
					wantErr = reportErr
				}
				if !errors.Is(err, wantErr) || outcome.Bundle.Root != root || outcome.ResultsState != tc.state || outcome.AnalysisAvailable != tc.available || len(outcome.Summary.Runs) != tc.runs {
					t.Fatalf("outcome = %+v, error = %v; want %s, %d runs, analysis=%v, error %v", outcome, err, tc.state, tc.runs, tc.available, wantErr)
				}
				if outcome.Summary.Runs[0].Manifest.RunID != cached.Manifest.RunID || !outcome.Summary.Runs[0].Reused {
					t.Fatalf("lost cached run/reuse flag: %+v", outcome.Summary.Runs[0])
				}
				for _, record := range outcome.Summary.Runs[1:] {
					if record.Reused || record.Manifest.RunID == old.Manifest.RunID {
						t.Fatalf("unexpected reused/unrelated run: %+v", record)
					}
				}
				if tc.available {
					if len(outcome.Analysis.Runs) != tc.runs || len(outcome.Analysis.Counts) != 6 {
						t.Fatalf("missing partial analysis: %+v", outcome.Analysis)
					}
					if tc.runs == 2 && (outcome.Analysis.Counts["success"] != 1 || outcome.Analysis.Counts["failed"] != 1) {
						t.Fatalf("partial counts = %+v, want one success and one failure", outcome.Analysis.Counts)
					}
				} else if !errors.Is(err, experiment.ErrUnverifiableRun) || outcome.Analysis.Counts != nil || len(outcome.Analysis.Runs) != 0 {
					t.Fatalf("integrity failure advertised analysis or lost error: %+v, %v", outcome.Analysis, err)
				}
				if data, err := os.ReadFile(marker); err != nil || string(data) != "previous report" {
					t.Fatalf("failed operation rewrote report: %q, %v", data, err)
				}
			})
		}
	}
}

func TestRunAndGenerateReportUseStructuredStudyOutcome(t *testing.T) {
	root := writeProjectFixture(t)
	outcome, err := Run(context.Background(), root, RunOptions{Workers: 2, RepositoryRevision: "test-revision"})
	if err != nil {
		t.Fatal(err)
	}
	if len(outcome.Summary.Runs) != 2 || len(outcome.Analysis.Runs) != 2 || outcome.ResultsState != ResultsComplete || !outcome.AnalysisAvailable {
		t.Fatalf("want complete outcome with 2 runs and available analysis, got %+v", outcome)
	}
	if outcome.Bundle.Project.ModelVersion != "study-model-v1" {
		t.Fatalf("model version = %q", outcome.Bundle.Project.ModelVersion)
	}
	if _, err := os.Stat(filepath.Join(root, "reports", "report.md")); err != nil {
		t.Fatalf("run did not generate report: %v", err)
	}

	analysis, err := GenerateReport(root, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.Runs) != 2 {
		t.Fatalf("generated report analysis runs=%d, want 2", len(analysis.Runs))
	}
}

func TestRunForwardsProgress(t *testing.T) {
	for _, state := range []struct {
		name string
		want experiment.ProgressSnapshot
		err  error
	}{
		{"success", experiment.ProgressSnapshot{Total: 2, Completed: 2}, nil},
		{"reused", experiment.ProgressSnapshot{Total: 2, Completed: 2, Reused: 2}, nil},
		{"model failure", experiment.ProgressSnapshot{Total: 2, Completed: 2, Failed: 2}, nil},
		{"cancelled", experiment.ProgressSnapshot{Total: 2, Cancelled: 2}, context.Canceled},
		{"invalid cache", experiment.ProgressSnapshot{Total: 2}, experiment.ErrUnverifiableRun},
	} {
		for _, observer := range []string{"buffered", "nil", "full", "unread"} {
			t.Run(state.name+"/"+observer, func(t *testing.T) {
				root := writeProjectFixture(t)
				if state.name == "model failure" {
					path := filepath.Join(root, "design.yaml")
					data, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(path, []byte(strings.Replace(string(data), "material: HZO", "material: BTO", 1)), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				if state.name == "reused" || state.name == "invalid cache" {
					outcome, err := Run(context.Background(), root, RunOptions{Workers: 2})
					if err != nil {
						t.Fatal(err)
					}
					if state.name == "invalid cache" {
						path := filepath.Join(root, "runs", outcome.Summary.Runs[0].Manifest.RunID, "manifest.json")
						if err := os.Remove(path); err != nil {
							t.Fatal(err)
						}
					}
				}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				if state.name == "cancelled" {
					cancel()
				}
				var progress chan experiment.ProgressSnapshot
				sentinel := experiment.ProgressSnapshot{Total: 99}
				switch observer {
				case "buffered":
					progress = make(chan experiment.ProgressSnapshot, 8)
				case "full":
					progress = make(chan experiment.ProgressSnapshot, 1)
					progress <- sentinel
				case "unread":
					progress = make(chan experiment.ProgressSnapshot)
				}
				outcome, err := Run(ctx, root, RunOptions{Workers: 2, Progress: progress})
				if !errors.Is(err, state.err) {
					t.Fatalf("error = %v, want %v", err, state.err)
				}
				if err == nil && (len(outcome.Summary.Runs) != 2 || len(outcome.Analysis.Runs) != 2) {
					t.Fatalf("outcome runs summary=%d analysis=%d, want 2/2", len(outcome.Summary.Runs), len(outcome.Analysis.Runs))
				}
				if progress != nil {
					close(progress) // The caller still owns the channel after Run.
				}
				if observer == "buffered" {
					var last experiment.ProgressSnapshot
					for snapshot := range progress {
						last = snapshot
					}
					if last != state.want {
						t.Fatalf("last progress = %+v, want %+v", last, state.want)
					}
				} else if observer == "full" {
					if got := <-progress; got != sentinel {
						t.Fatalf("full observer changed: %+v", got)
					}
				}
			})
		}
	}
}

func TestCreateStarterProjectRejectsInvalidSweepBeforeFilesystemChanges(t *testing.T) {
	for _, options := range []StarterSweepOptions{
		{ConductanceLevels: []int{8, 8}, Rows: []int{32}, ADCBits: []int{4}},
		{ConductanceLevels: []int{1}, Rows: []int{32}, ADCBits: []int{4}},
		{ConductanceLevels: []int{8}, Rows: []int{0}, ADCBits: []int{4}},
		{ConductanceLevels: []int{8}, Rows: []int{32}, ADCBits: []int{17}},
	} {
		root := filepath.Join(t.TempDir(), "invalid")
		if err := CreateStarterProject(root, options); err == nil {
			t.Fatal("invalid starter sweep was accepted")
		}
		if _, err := os.Stat(root); !os.IsNotExist(err) {
			t.Fatalf("invalid sweep changed target directory: %v", err)
		}
	}
}

func TestCreateStarterProjectUsesRequestedSweepAndPreservesExistingData(t *testing.T) {
	root := filepath.Join(t.TempDir(), "starter")
	options := StarterSweepOptions{ConductanceLevels: []int{8, 16}, Rows: []int{16, 32}, ADCBits: []int{4, 6}}
	if err := CreateStarterProject(root, options); err != nil {
		t.Fatal(err)
	}
	bundle, points, err := ValidateProject(root, "")
	if err != nil {
		t.Fatal(err)
	}
	if points != 8 || len(bundle.Sweep.Parameters) != 3 {
		t.Fatalf("sweep points=%d parameters=%d, want 8 and 3", points, len(bundle.Sweep.Parameters))
	}
	if bundle.Sweep.Parameters[0].Values[0] != 8 || bundle.Sweep.Parameters[1].Values[0] != 16 || bundle.Sweep.Parameters[2].Values[0] != 4 {
		t.Fatalf("starter sweep values = %+v", bundle.Sweep.Parameters)
	}

	marker := filepath.Join(root, "keep.txt")
	if err := os.WriteFile(marker, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := CreateStarterProject(root, options); err == nil {
		t.Fatal("initialized over a non-empty project")
	}
	if got, err := os.ReadFile(marker); err != nil || string(got) != "keep" {
		t.Fatalf("existing file changed: got=%q err=%v", got, err)
	}
}

func writeProjectFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"project.yaml": `schema_version: 1
id: hzo-study
name: HZO Study
hypothesis: ADC resolution changes energy and area.
model_version: study-model-v1
objectives:
  - metric: energy_pj
    direction: minimize
`,
		"design.yaml": `schema_version: 1
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
`,
		"sweep.yaml": `schema_version: 1
seed: 17
max_points: 8
parameters:
  - path: circuit.adc_bits
    values: [4, 6]
`,
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}
