package experiment

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"fecim-lattice-tools/workbench/project"
)

func TestRootedRunStore(t *testing.T) {
	eval := func(context.Context, project.Design, int64) (Result, error) { return successResult(1), nil }
	open := func(t *testing.T, path string) *os.Root {
		t.Helper()
		root, err := os.OpenRoot(path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = root.Close() })
		return root
	}
	for _, outcome := range []string{"success", "invalid result", "invalid sweep", "cancelled"} {
		t.Run("legacy missing root "+outcome, func(t *testing.T) {
			bundle := runnerBundle(t, 1)
			bundle.Root = filepath.Join(bundle.Root, "missing", "workspace")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			opts := deterministicRunOptions(eval)
			switch outcome {
			case "invalid result":
				opts.Evaluator = func(context.Context, project.Design, int64) (Result, error) {
					return Result{Status: StatusSuccess}, nil
				}
			case "invalid sweep":
				bundle.Sweep.MaxPoints = 0
			case "cancelled":
				cancel()
			}
			summary, err := Run(ctx, bundle, opts)
			if outcome == "success" {
				if err != nil || len(summary.Runs) != 1 {
					t.Fatalf("lazy workspace creation=%+v, %v", summary, err)
				}
				return
			}
			if err == nil || errors.Is(err, os.ErrNotExist) {
				t.Fatalf("precommit error=%v; must not be a missing workspace error", err)
			}
			if _, err := os.Stat(bundle.Root); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("created workspace before commit: %v", err)
			}
		})
	}
	t.Run("metadata and canonical compatibility", func(t *testing.T) {
		bundle := runnerBundle(t, 2)
		legacyRoot := bundle.Root
		mixed := func(_ context.Context, design project.Design, _ int64) (Result, error) {
			if design.Circuit.ADCBits == 2 {
				return Result{Status: StatusFailed, Failure: &Failure{Kind: "domain", Message: "invalid design"}}, nil
			}
			return successResult(1), nil
		}
		legacy, err := Run(context.Background(), bundle, deterministicRunOptions(mixed))
		if err != nil {
			t.Fatal(err)
		}
		parent := t.TempDir()
		path := filepath.Join(parent, "original")
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatal(err)
		}
		root := open(t, path)
		retained := filepath.Join(parent, "retained")
		if err := os.Rename(path, retained); err != nil {
			t.Fatal(err)
		}
		outside := t.TempDir()
		if err := os.Symlink(outside, path); err != nil {
			t.Fatal(err)
		}
		bundle.Root = outside // Neither this nor root.Name() grants I/O authority.
		opts := deterministicRunOptions(mixed)
		opts.Workers = 3
		progress := make(chan ProgressSnapshot, 20)
		opts.Progress = progress
		got, err := RunRoot(context.Background(), root, bundle, opts)
		if err != nil || !reflect.DeepEqual(got, legacy) {
			t.Fatalf("rooted=%+v, %v; legacy=%+v", got, err, legacy)
		}
		for _, record := range got.Runs {
			for _, name := range []string{"manifest.json", "resolved-design.json", "result.json"} {
				actual, err := os.ReadFile(filepath.Join(retained, "runs", record.Manifest.RunID, name))
				if err != nil {
					t.Fatal(err)
				}
				want, err := os.ReadFile(filepath.Join(legacyRoot, "runs", record.Manifest.RunID, name))
				if err != nil || !bytes.Equal(actual, want) {
					t.Fatalf("canonical %s differs: %v", name, err)
				}
			}
			loaded, ok, err := LoadRunRecordRoot(root, record.Manifest.RunID)
			if err != nil || !ok || !reflect.DeepEqual(loaded, record) {
				t.Fatalf("record=%+v, %v, %v", loaded, ok, err)
			}
			_, err = LoadRunRoot(root, record.Manifest.RunID)
			if (err != nil) != (record.Result.Status == StatusFailed) {
				t.Fatalf("export status guard: %v", err)
			}
		}
		listed, err := LoadRunsRoot(root)
		if err != nil || !reflect.DeepEqual(listed, got.Runs) {
			t.Fatalf("listing=%+v, %v", listed, err)
		}
		opts.Evaluator = func(context.Context, project.Design, int64) (Result, error) {
			t.Fatal("reevaluated cached run")
			return Result{}, nil
		}
		reused, err := RunRoot(context.Background(), root, bundle, opts)
		if err != nil || len(reused.Runs) != 2 || !reused.Runs[0].Reused || !reused.Runs[1].Reused {
			t.Fatalf("reuse=%+v, %v", reused, err)
		}
		var last ProgressSnapshot
		for len(progress) > 0 {
			last = <-progress
		}
		if last.Completed != 2 || last.Reused != 2 || last.Failed != 1 {
			t.Fatalf("progress=%+v", last)
		}
		entries, err := os.ReadDir(outside)
		if err != nil || len(entries) != 0 {
			t.Fatalf("metadata path touched: %v, %v", entries, err)
		}
		if _, err := root.Stat("runs"); err != nil {
			t.Fatalf("caller handle closed: %v", err)
		}
	})
	for _, link := range []string{"runs", "run", "artifact"} {
		t.Run("outside "+link+" link", func(t *testing.T) {
			outside := runnerBundle(t, 1)
			summary, err := Run(context.Background(), outside, deterministicRunOptions(eval))
			if err != nil {
				t.Fatal(err)
			}
			id := summary.Runs[0].Manifest.RunID
			bundle := outside
			bundle.Root = t.TempDir()
			root := open(t, bundle.Root)
			if err := root.MkdirAll(filepath.Join("runs", id), 0o755); err != nil {
				t.Fatal(err)
			}
			from, to := "runs", filepath.Join(outside.Root, "runs")
			if link != "runs" {
				from, to = filepath.Join("runs", id), filepath.Join(to, id)
			}
			if link == "artifact" {
				for _, name := range []string{"manifest.json", "resolved-design.json", "result.json"} {
					data, err := os.ReadFile(filepath.Join(to, name))
					if err != nil {
						t.Fatal(err)
					}
					if err := root.WriteFile(filepath.Join(from, name), data, 0o600); err != nil {
						t.Fatal(err)
					}
				}
				from, to = filepath.Join(from, "result.json"), filepath.Join(to, "result.json")
			}
			if err := root.RemoveAll(from); err != nil {
				t.Fatal(err)
			}
			if err := root.Symlink(to, from); err != nil {
				t.Fatal(err)
			}
			if _, _, err := LoadRunRecordRoot(root, id); err == nil {
				t.Fatal("record loader followed outside link")
			}
			if _, err := LoadRunRoot(root, id); err == nil {
				t.Fatal("export loader followed outside link")
			}
			if _, err := LoadRunsRoot(root); err == nil {
				t.Fatal("listing followed outside link")
			}
			opts := deterministicRunOptions(func(context.Context, project.Design, int64) (Result, error) {
				t.Fatal("evaluated linked cache")
				return Result{}, nil
			})
			if _, err := RunRoot(context.Background(), root, bundle, opts); err == nil {
				t.Fatal("runner followed outside link")
			}
			unchanged, err := LoadRun(outside.Root, id)
			if err != nil || !reflect.DeepEqual(unchanged, summary.Runs[0]) {
				t.Fatalf("outside record changed: %+v, %v", unchanged, err)
			}
			if _, err := root.Stat("."); err != nil {
				t.Fatalf("error path closed caller handle: %v", err)
			}
		})
	}
	t.Run("pin runs across retarget", func(t *testing.T) {
		bundle := runnerBundle(t, 1)
		root := open(t, bundle.Root)
		if err := root.Mkdir("runs", 0o755); err != nil {
			t.Fatal(err)
		}
		outside := t.TempDir()
		if err := os.WriteFile(filepath.Join(outside, ".tmp-unowned"), []byte("keep"), 0o600); err != nil {
			t.Fatal(err)
		}
		opts := deterministicRunOptions(func(context.Context, project.Design, int64) (Result, error) {
			if err := root.Rename("runs", "retained"); err != nil {
				t.Fatal(err)
			}
			if err := root.Symlink(outside, "runs"); err != nil {
				t.Fatal(err)
			}
			return successResult(1), nil
		})
		summary, err := RunRoot(context.Background(), root, bundle, opts)
		if err != nil || len(summary.Runs) != 1 {
			t.Fatalf("retargeted run=%+v, %v", summary, err)
		}
		if err := root.Remove("runs"); err != nil {
			t.Fatal(err)
		}
		if err := root.Rename("retained", "runs"); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadRunRoot(root, summary.Runs[0].Manifest.RunID); err != nil {
			t.Fatal(err)
		}
		assertNoTempRuns(t, bundle.Root)
		entries, err := os.ReadDir(outside)
		if err != nil || len(entries) != 1 || entries[0].Name() != ".tmp-unowned" {
			t.Fatalf("outside touched: %v, %v", entries, err)
		}
	})
	for _, failure := range []string{"namespace retarget", "scratch retarget", "verified winner"} {
		t.Run("publication failure "+failure, func(t *testing.T) {
			outside := runnerBundle(t, 1)
			seed, err := Run(context.Background(), outside, deterministicRunOptions(eval))
			if err != nil {
				t.Fatal(err)
			}
			bundle := outside
			bundle.Root = t.TempDir()
			root := open(t, bundle.Root)
			storage, err := openStore(root)
			if err != nil {
				t.Fatal(err)
			}
			defer storage.close()
			renameErr := errors.New("publication failed")
			var scratchName string
			storage.rename = func(temp, id string) error {
				scratchName = temp
				switch failure {
				case "namespace retarget":
					if err := root.Rename("runs", "retained"); err != nil {
						t.Fatal(err)
					}
					if err := root.Symlink(filepath.Join(outside.Root, "runs"), "runs"); err != nil {
						t.Fatal(err)
					}
				case "scratch retarget":
					if err := storage.runs.Rename(temp, "owned"); err != nil {
						t.Fatal(err)
					}
					if err := storage.runs.Mkdir(temp, 0o700); err != nil {
						t.Fatal(err)
					}
					if err := storage.runs.WriteFile(filepath.Join(temp, "result.json"), []byte("unowned"), 0o600); err != nil {
						t.Fatal(err)
					}
				case "verified winner":
					// A second writer publishes a fully verified record first.
					other, err := openStore(root)
					if err != nil {
						t.Fatal(err)
					}
					defer other.close()
					if _, err := other.commit(seed.Runs[0]); err != nil {
						t.Fatal(err)
					}
				}
				return renameErr
			}
			got, err := storage.commit(seed.Runs[0])
			if failure == "verified winner" {
				if err != nil || !got.Reused {
					t.Fatalf("verified winner=%+v, %v", got, err)
				}
				if _, err := LoadRunRoot(root, got.Manifest.RunID); err != nil {
					t.Fatalf("winner damaged by cleanup: %v", err)
				}
			} else if !errors.Is(err, renameErr) {
				t.Fatalf("publication failure=%v; must not reuse an outside winner", err)
			}
			if failure == "scratch retarget" {
				data, err := storage.runs.ReadFile(filepath.Join(scratchName, "result.json"))
				if err != nil || string(data) != "unowned" {
					t.Fatalf("cleaned replacement scratch: %q, %v", data, err)
				}
				entries, err := os.ReadDir(filepath.Join(bundle.Root, "runs", "owned"))
				if err != nil || len(entries) != 0 {
					t.Fatalf("owned scratch not cleaned: %v, %v", entries, err)
				}
			} else if _, err := storage.runs.Stat(scratchName); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("owned scratch remains: %v", err)
			}
			unchanged, err := LoadRun(outside.Root, seed.Runs[0].Manifest.RunID)
			if err != nil || !reflect.DeepEqual(unchanged, seed.Runs[0]) {
				t.Fatalf("outside winner changed: %+v, %v", unchanged, err)
			}
		})
	}
	t.Run("invalid IDs and corrupt cache", func(t *testing.T) {
		bundle := runnerBundle(t, 1)
		root := open(t, bundle.Root)
		for _, id := range []string{"../escape", "", strings.Repeat("A", 64)} {
			if _, _, err := LoadRunRecordRoot(root, id); err == nil {
				t.Fatalf("accepted ID %q", id)
			}
			if _, err := LoadRunRoot(root, id); err == nil {
				t.Fatalf("export accepted ID %q", id)
			}
		}
		if _, ok, err := LoadRunRecordRoot(root, strings.Repeat("0", 64)); err != nil || ok {
			t.Fatalf("missing record=%v, %v", ok, err)
		}
		if _, err := LoadRunRoot(root, strings.Repeat("0", 64)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("missing export=%v", err)
		}
		if got, err := LoadRunsRoot(root); err != nil || len(got) != 0 {
			t.Fatalf("empty listing=%v, %v", got, err)
		}
		summary, err := RunRoot(context.Background(), root, bundle, deterministicRunOptions(eval))
		if err != nil {
			t.Fatal(err)
		}
		id := summary.Runs[0].Manifest.RunID
		if err := root.WriteFile(filepath.Join("runs", id, "result.json"), []byte("tampered"), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, _, err := LoadRunRecordRoot(root, id); !errors.Is(err, ErrUnverifiableRun) {
			t.Fatalf("corrupt record=%v", err)
		}
		if _, err := LoadRunsRoot(root); !errors.Is(err, ErrUnverifiableRun) {
			t.Fatalf("corrupt listing=%v", err)
		}
		if _, err := RunRoot(context.Background(), root, bundle, deterministicRunOptions(eval)); !errors.Is(err, ErrUnverifiableRun) {
			t.Fatalf("corrupt reuse=%v", err)
		}
	})
	t.Run("cancelled worker settles without closing caller root", func(t *testing.T) {
		bundle := runnerBundle(t, 2)
		root := open(t, bundle.Root)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		opts := deterministicRunOptions(func(context.Context, project.Design, int64) (Result, error) { cancel(); return successResult(1), nil })
		summary, err := RunRoot(ctx, root, bundle, opts)
		if !errors.Is(err, context.Canceled) || len(summary.Runs) != 1 {
			t.Fatalf("cancelled=%+v, %v", summary, err)
		}
		if _, err := LoadRunRoot(root, summary.Runs[0].Manifest.RunID); err != nil {
			t.Fatal(err)
		}
		assertNoTempRuns(t, bundle.Root)
	})
	for _, invalid := range []string{"nil", "closed"} {
		t.Run(invalid+" root", func(t *testing.T) {
			var root *os.Root
			if invalid == "closed" {
				root = open(t, t.TempDir())
				_ = root.Close()
			}
			bundle := runnerBundle(t, 1)
			if _, err := RunRoot(context.Background(), root, bundle, deterministicRunOptions(eval)); err == nil {
				t.Fatal("accepted invalid root")
			}
			if _, _, err := LoadRunRecordRoot(root, strings.Repeat("0", 64)); err == nil {
				t.Fatal("record accepted invalid root")
			}
			if _, err := LoadRunRoot(root, strings.Repeat("0", 64)); err == nil {
				t.Fatal("export accepted invalid root")
			}
			if _, err := LoadRunsRoot(root); err == nil {
				t.Fatal("listing accepted invalid root")
			}
		})
	}
}
