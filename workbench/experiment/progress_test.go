package experiment

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"fecim-lattice-tools/workbench/project"
)

// Missing/double-counted commits, blocking observers, or progress leaking into
// scientific artifacts must fail at the Run boundary.
func TestRunProgressCounts(t *testing.T) {
	checkProgress := func(t *testing.T, ch chan ProgressSnapshot, want ProgressSnapshot) {
		t.Helper()
		var last ProgressSnapshot
		seen := false
		for {
			select {
			case p, ok := <-ch:
				if !ok {
					t.Fatal("Run closed caller-owned progress channel")
				}
				if p.Total != want.Total || p.Completed < 0 || p.Reused < 0 || p.Failed < 0 || p.Cancelled < 0 ||
					p.Reused > p.Completed || p.Failed > p.Completed || p.Completed+p.Cancelled > p.Total ||
					p.Completed < last.Completed || p.Reused < last.Reused || p.Failed < last.Failed || p.Cancelled < last.Cancelled {
					t.Fatalf("invalid progress transition: %+v -> %+v", last, p)
				}
				last, seen = p, true
			default:
				if !seen || last != want {
					t.Fatalf("last progress=%+v (seen=%v), want %+v", last, seen, want)
				}
				return
			}
		}
	}
	eval := func(_ context.Context, _ project.Design, seed int64) (Result, error) {
		if seed == fixtureBundle().Sweep.Seed+2 {
			return Result{Status: StatusFailed, Failure: &Failure{Kind: "domain", Message: "test domain failure"}}, nil
		}
		return successResult(float64(seed)), nil
	}

	t.Run("observers preserve outcomes and artifact bytes", func(t *testing.T) {
		baseline := runnerBundle(t, 3)
		want, err := Run(context.Background(), baseline, deterministicRunOptions(eval))
		if err != nil {
			t.Fatal(err)
		}
		for _, test := range []struct {
			name    string
			workers int
			mode    string
		}{
			{"buffered sequential", 1, "buffered"},
			{"buffered concurrent", 3, "buffered"},
			{"nil concurrent", 3, "nil"},
			{"unread unbuffered", 3, "unread"},
			{"full buffer", 3, "full"},
		} {
			t.Run(test.name, func(t *testing.T) {
				bundle := baseline
				bundle.Root = t.TempDir()
				opts := deterministicRunOptions(eval)
				opts.Workers = test.workers
				var ch chan ProgressSnapshot
				switch test.mode {
				case "buffered":
					ch = make(chan ProgressSnapshot, 16)
				case "unread":
					ch = make(chan ProgressSnapshot)
				case "full":
					ch = make(chan ProgressSnapshot, 1)
					ch <- ProgressSnapshot{Total: 99}
				}
				opts.Progress = ch
				type outcome struct {
					summary Summary
					err     error
				}
				done := make(chan outcome, 1)
				go func() {
					summary, err := Run(context.Background(), bundle, opts)
					done <- outcome{summary, err}
				}()
				var got Summary
				select {
				case result := <-done:
					if result.err != nil {
						t.Fatal(result.err)
					}
					got = result.summary
				case <-time.After(5 * time.Second):
					t.Fatal("progress observer blocked Run")
				}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("progress/schedule changed ordered outcomes: got %+v want %+v", got, want)
				}
				for _, run := range want.Runs {
					id := run.Manifest.RunID
					if !reflect.DeepEqual(snapshotRunFiles(t, filepath.Join(bundle.Root, "runs", id)), snapshotRunFiles(t, filepath.Join(baseline.Root, "runs", id))) {
						t.Fatalf("progress/schedule changed artifact bytes for %s", id)
					}
				}
				if test.mode == "buffered" {
					checkProgress(t, ch, ProgressSnapshot{Total: 3, Completed: 3, Failed: 1})
				} else if test.mode == "full" {
					if p, ok := <-ch; !ok || p != (ProgressSnapshot{Total: 99}) {
						t.Fatalf("full observer was closed or overwritten: %+v, %v", p, ok)
					}
				}
				if ch != nil {
					close(ch) // Caller owns closure, after Run returns.
				}
			})
		}
	})

	t.Run("settled cancellation and reuse", func(t *testing.T) {
		bundle := runnerBundle(t, 4)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		started := make(chan struct{}, 3)
		release := make(chan struct{})
		releaseWorkers := sync.OnceFunc(func() { cancel(); close(release) })
		opts := deterministicRunOptions(func(_ context.Context, design project.Design, seed int64) (Result, error) {
			started <- struct{}{}
			<-release
			// Deliberately finish a commit even though cancellation has occurred.
			return eval(ctx, design, seed)
		})
		opts.Workers = 3
		ch := make(chan ProgressSnapshot, 32)
		opts.Progress = ch
		done := make(chan struct{})
		var partial Summary
		var runErr error
		go func() {
			partial, runErr = Run(ctx, bundle, opts)
			close(done)
		}()
		defer func() { releaseWorkers(); <-done }()
		for range 3 {
			select {
			case <-started:
			case <-time.After(5 * time.Second):
				t.Fatal("workers did not reach cancellation barrier")
			}
		}
		// Initial progress must be visible while simulation is still blocked.
		select {
		case p := <-ch:
			if p != (ProgressSnapshot{Total: 4}) {
				t.Fatalf("initial progress=%+v", p)
			}
		default:
			t.Fatal("no progress while simulation is running")
		}
		releaseWorkers()
		<-done
		if !errors.Is(runErr, context.Canceled) || len(partial.Runs) != 3 {
			t.Fatalf("settled partial=%+v err=%v, want 3 committed runs and cancellation", partial, runErr)
		}
		checkProgress(t, ch, ProgressSnapshot{Total: 4, Completed: 3, Failed: 1, Cancelled: 1})
		stored, err := LoadRuns(bundle.Root)
		if err != nil || !reflect.DeepEqual(stored, partial.Runs) {
			t.Fatalf("settled commits missing from return: stored=%+v err=%v partial=%+v", stored, err, partial)
		}
		before := make(map[string]map[string][]byte)
		for _, run := range partial.Runs {
			before[run.Manifest.RunID] = snapshotRunFiles(t, filepath.Join(bundle.Root, "runs", run.Manifest.RunID))
		}
		var calls atomic.Int32
		resume := deterministicRunOptions(func(ctx context.Context, design project.Design, seed int64) (Result, error) {
			calls.Add(1)
			return eval(ctx, design, seed)
		})
		resume.Progress = ch
		for _, want := range []ProgressSnapshot{
			{Total: 4, Completed: 4, Reused: 3, Failed: 1},
			{Total: 4, Completed: 4, Reused: 4, Failed: 1},
		} {
			summary, err := Run(context.Background(), bundle, resume)
			if err != nil || len(summary.Runs) != 4 || calls.Load() != 1 {
				t.Fatalf("resume/all-reused: runs=%d calls=%d err=%v", len(summary.Runs), calls.Load(), err)
			}
			checkProgress(t, ch, want)
		}
		for id, files := range before {
			if !reflect.DeepEqual(files, snapshotRunFiles(t, filepath.Join(bundle.Root, "runs", id))) {
				t.Fatalf("reuse changed committed artifacts for %s", id)
			}
		}
		close(ch)
		data, err := json.Marshal(ProgressSnapshot{Total: 4, Completed: 4, Reused: 4, Failed: 1})
		if err != nil || string(data) != `{"total":4,"completed":4,"reused":4,"failed":1,"cancelled":0}` {
			t.Fatalf("progress JSON=%s err=%v", data, err)
		}
	})

	t.Run("cancellation differs from systemic error", func(t *testing.T) {
		boom := errors.New("evaluator unavailable")
		for _, cancelled := range []bool{false, true} {
			bundle := runnerBundle(t, 4)
			ctx, cancel := context.WithCancel(context.Background())
			if cancelled {
				cancel()
			}
			ch := make(chan ProgressSnapshot, 16)
			opts := deterministicRunOptions(func(_ context.Context, _ project.Design, seed int64) (Result, error) {
				if seed == bundle.Sweep.Seed+1 {
					return Result{}, boom
				}
				return successResult(float64(seed)), nil
			})
			opts.Progress = ch
			summary, err := Run(ctx, bundle, opts)
			cancel()
			want := ProgressSnapshot{Total: 4, Completed: 1}
			wantErr := boom
			if cancelled {
				want = ProgressSnapshot{Total: 4, Cancelled: 4}
				wantErr = context.Canceled
			}
			if !errors.Is(err, wantErr) || len(summary.Runs) != want.Completed {
				t.Fatalf("runs=%d err=%v, want %+v and %v", len(summary.Runs), err, want, wantErr)
			}
			checkProgress(t, ch, want)
			close(ch)
		}
	})
}
