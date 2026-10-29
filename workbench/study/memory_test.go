package study

import (
	"bytes"
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

// The checked-in example includes 64-row arrays, which remain native-only. This
// parity fixture explicitly substitutes [16,32] for [32,64] in BOTH targets; it
// preserves all eight points and uses the original native results serializer.
func TestWebResearchInMemory(t *testing.T) {
	docs := map[string][]byte{}
	nativeDir := t.TempDir()
	for _, name := range []string{"project.yaml", "design.yaml", "sweep.yaml"} {
		data, err := os.ReadFile(filepath.Join("..", "..", "examples", "research-design-workbench", name))
		if err != nil {
			t.Fatal(err)
		}
		if name == "sweep.yaml" {
			data = bytes.Replace(data, []byte("[32, 64]"), []byte("[16, 32]"), 1)
		}
		docs[name] = data
		if err := os.WriteFile(filepath.Join(nativeDir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	bundle, err := project.LoadBytes(docs)
	if err != nil {
		t.Fatal(err)
	}
	native, err := Run(context.Background(), nativeDir, RunOptions{Workers: 1, RepositoryRevision: "parity"})
	if err != nil {
		t.Fatal(err)
	}
	progress := make(chan experiment.ProgressSnapshot, 16)
	memory, err := RunMemory(context.Background(), bundle, RunOptions{Workers: 1, RepositoryRevision: "parity", Progress: progress})
	if err != nil {
		t.Fatal(err)
	}
	if len(memory.Summary.Runs) != 8 || memory.ResultsState != ResultsComplete || !memory.AnalysisAvailable {
		t.Fatalf("memory outcome = %+v", memory)
	}
	want, err := os.ReadFile(filepath.Join(nativeDir, "reports", "results.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(memory.ResultsJSON, want) {
		t.Fatalf("original results.json parity mismatch\n%s\n%s", memory.ResultsJSON, want)
	}
	for i, run := range memory.Summary.Runs {
		if run.Manifest.RunID != native.Summary.Runs[i].Manifest.RunID || !reflect.DeepEqual(run.Result, native.Summary.Runs[i].Result) {
			t.Fatalf("record %d differs", i)
		}
	}
	var previous experiment.ProgressSnapshot
	close(progress)
	snapshots := 0
	for snapshot := range progress {
		if snapshot.Total != 8 || snapshot.Completed < previous.Completed {
			t.Fatalf("progress %+v after %+v", snapshot, previous)
		}
		previous = snapshot
		snapshots++
	}
	if snapshots < 9 || previous.Completed != 8 {
		t.Fatalf("missing incremental progress: count %d, %+v", snapshots, previous)
	}
	t.Run("point and dimension admission before expansion", func(t *testing.T) {
		for _, tt := range []struct {
			name       string
			sweep      project.Sweep
			rows, cols int
			accepted   bool
		}{
			{"boundary", project.Sweep{SchemaVersion: 1, MaxPoints: 10000, Parameters: []project.Parameter{{Path: "array.rows", Range: &project.LinearRange{Start: 1, Stop: 32, Count: 256}}}}, 32, 32, true},
			{"one too many", project.Sweep{SchemaVersion: 1, MaxPoints: 10000, Parameters: []project.Parameter{{Path: "array.rows", Range: &project.LinearRange{Start: 1, Stop: 32, Count: 257}}}}, 32, 32, false},
			{"huge allocation", project.Sweep{SchemaVersion: 1, MaxPoints: 10000, Parameters: []project.Parameter{{Path: "array.rows", Range: &project.LinearRange{Start: 1, Stop: 32, Count: 1 << 30}}}}, 32, 32, false},
			{"rows", bundle.Sweep, 33, 32, false}, {"cols", bundle.Sweep, 32, 33, false},
			{"swept dimension", project.Sweep{SchemaVersion: 1, MaxPoints: 10000, Parameters: []project.Parameter{{Path: "array.rows", Values: []float64{32, 33}}}}, 32, 32, false},
			{"huge integer", project.Sweep{SchemaVersion: 1, MaxPoints: 10000, Parameters: []project.Parameter{{Path: "device.conductance_levels", Values: []float64{1e100}}}}, 32, 32, false},
		} {
			t.Run(tt.name, func(t *testing.T) {
				b := bundle
				b.Sweep = tt.sweep
				b.Design.Array.Rows = tt.rows
				b.Design.Array.Cols = tt.cols
				// Count=256 with a fractional array coordinate is intentionally invalid for
				// Expand. Use duplicate integral coordinates to exercise the admission edge.
				if tt.accepted {
					b.Sweep.Parameters[0].Range = nil
					b.Sweep.Parameters[0].Values = make([]float64, 256)
					for i := range b.Sweep.Parameters[0].Values {
						b.Sweep.Parameters[0].Values[i] = 32
					}
				}
				out, err := RunMemory(context.Background(), b, RunOptions{Workers: 1})
				if tt.accepted {
					if err != nil || len(out.Summary.Runs) != 256 {
						t.Fatalf("boundary: %d %v", len(out.Summary.Runs), err)
					}
				} else if err == nil || out.ResultsState != ResultsUnavailable || len(out.Summary.Runs) != 0 {
					t.Fatalf("invalid admitted: %+v %v", out, err)
				}
			})
		}
	})
	t.Run("native result requirements", func(t *testing.T) {
		for _, modify := range []func(*project.Bundle){
			func(b *project.Bundle) { b.Project.Objectives[0].Metric = "not_produced" },
			func(b *project.Bundle) { b.Project.Constraints[0].Unit = "mm2" },
		} {
			b := bundle
			b.Project.Objectives = append([]project.Objective(nil), bundle.Project.Objectives...)
			b.Project.Constraints = append([]project.Constraint(nil), bundle.Project.Constraints...)
			modify(&b)
			out, err := RunMemory(context.Background(), b, RunOptions{Workers: 1})
			if !errors.Is(err, experiment.ErrInvalidResult) || len(out.Summary.Runs) != 0 {
				t.Fatalf("invalid evaluator output accepted: %+v %v", out, err)
			}
		}
	})
	t.Run("no filesystem authority", func(t *testing.T) {
		b := bundle
		b.Root = filepath.Join(t.TempDir(), "must-not-exist")
		out, err := RunMemory(context.Background(), b, RunOptions{Workers: 1})
		if err != nil || len(out.Summary.Runs) != 8 {
			t.Fatalf("%+v %v", out, err)
		}
		if _, err := os.Stat(b.Root); !os.IsNotExist(err) {
			t.Fatalf("disk touched: %v", err)
		}
		if _, err := RunMemory(context.Background(), b, RunOptions{Workers: 1, CitationDir: "forbidden"}); err == nil {
			t.Fatal("disk citation authority admitted")
		}
		if _, err := RunMemory(context.Background(), b, RunOptions{}); err == nil {
			t.Fatal("invalid worker count admitted")
		}
	})
	t.Run("retained partial on cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		calls := 0
		evaluate := func(ctx context.Context, d project.Design, seed int64) (experiment.Result, error) {
			calls++
			result, err := fecim.Evaluate(ctx, d, seed)
			cancel()
			return result, err
		}
		ch := make(chan experiment.ProgressSnapshot, 16)
		out, err := runMemory(ctx, bundle, RunOptions{Workers: 1, Progress: ch}, evaluate)
		if !errors.Is(err, context.Canceled) || calls != 1 || len(out.Summary.Runs) != 1 || out.ResultsState != ResultsPartial || !out.AnalysisAvailable {
			t.Fatalf("partial: %+v %v calls=%d", out, err, calls)
		}
		var expected bytes.Buffer
		if err := report.WriteJSON(&expected, bundle, out.Analysis); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(expected.Bytes(), out.ResultsJSON) {
			t.Fatal("partial serialization changed")
		}
		close(ch)
		var last experiment.ProgressSnapshot
		for p := range ch {
			last = p
		}
		if last.Completed != 1 || last.Cancelled != 7 {
			t.Fatalf("terminal progress=%+v", last)
		}
		preCancelled, stop := context.WithCancel(context.Background())
		stop()
		out, err = RunMemory(preCancelled, bundle, RunOptions{Workers: 1, Progress: make(chan experiment.ProgressSnapshot)})
		if !errors.Is(err, context.Canceled) || len(out.Summary.Runs) != 0 || out.ResultsState != ResultsPartial || !strings.Contains(string(out.ResultsJSON), `"runs": []`) {
			t.Fatalf("pre-cancel: %+v %v", out, err)
		}
	})
}
