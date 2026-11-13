package scripts_test

import (
	"context"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestValidationAnalysis(t *testing.T) {
	analyzer := readFile(t, filepath.Join(repoRoot(t), "scripts/analyze_validation.sh"))
	type event struct {
		Package string `json:"Package,omitempty"`
		Test    string
		Action  string
	}
	const pkgA = "fecim-lattice-tools/shared/mathutil"
	const pkgB = "fecim-lattice-tools/shared/testutil"
	aPass := event{pkgA, "TestClamp01", "pass"}
	aFail := event{pkgA, "TestClamp01", "fail"}
	bPass := event{pkgB, "TestClamp01", "pass"}
	extra := event{pkgB, "TestNew", "pass"}
	legacyPass := event{"", "TestClamp01", "pass"}
	legacyFail := event{"", "TestClamp01", "fail"}
	namedRegression := "[TEST REGRESSION] module=shared package='" + pkgA + "' test='TestClamp01': PASS -> FAIL"
	legacyRegression := "[TEST REGRESSION] module=shared package='' test='TestClamp01': PASS -> FAIL"
	cases := []struct {
		name        string
		previous    []event
		current     []event
		history     int
		wantExit    int
		regressions []string
		message     string
	}{
		{"collision fail then pass", []event{aPass, bPass}, []event{aFail, bPass, extra}, 2, 1, []string{namedRegression}, ""},
		{"collision pass then fail", []event{aPass, bPass}, []event{bPass, aFail, extra}, 2, 1, []string{namedRegression}, ""},
		{"different package between runs", []event{{pkgA, "TestX", "pass"}}, []event{{pkgB, "TestX", "fail"}, extra}, 2, 0, nil, "No regressions found"},
		{"same package", []event{aPass}, []event{aFail, extra}, 2, 1, []string{namedRegression}, ""},
		{"module pass drop only", []event{aPass, bPass}, []event{aPass}, 2, 1, []string{"[MODULE REGRESSION] module=shared: pass count 2 -> 1"}, ""},
		{"no regression", []event{aPass, bPass}, []event{bPass, aPass}, 2, 0, nil, "No regressions found"},
		{"pass to skip", []event{aPass}, []event{{pkgA, "TestClamp01", "skip"}, extra}, 2, 0, nil, "No regressions found"},
		{"skip to fail", []event{{pkgA, "TestClamp01", "skip"}}, []event{aFail}, 2, 0, nil, "No regressions found"},
		{"legacy regression beside named pass", []event{legacyPass, aPass}, []event{legacyFail, aPass, extra}, 2, 1, []string{legacyRegression}, ""},
		{"legacy not named", []event{legacyPass}, []event{aFail, extra}, 2, 0, nil, "No regressions found"},
		{"named not legacy", []event{aPass}, []event{legacyFail, extra}, 2, 0, nil, "No regressions found"},
		{"missing history", nil, nil, 0, 0, nil, "No history file found"},
		{"one entry history", nil, []event{aFail}, 1, 0, nil, "No regression baseline available"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			mustWrite(t, filepath.Join(root, "scripts/analyze_validation.sh"), analyzer)
			history := ""
			for i, events := range [][]event{tc.previous, tc.current} {
				timestamp := []string{"2026-01-01-1200", "2026-01-01-1201"}[i]
				counts := map[string]int{"pass": 0, "fail": 0, "skip": 0}
				raw := ""
				for _, e := range events {
					line, err := json.Marshal(e)
					if err != nil {
						t.Fatal(err)
					}
					raw += string(line) + "\n"
					counts[e.Action]++
				}
				runDir := filepath.Join(root, "output/validation", timestamp)
				mustWrite(t, filepath.Join(runDir, "shared_results.json"), raw)
				summary, err := json.Marshal(map[string]any{"timestamp": timestamp, "modules": map[string]any{"shared": counts}})
				if err != nil {
					t.Fatal(err)
				}
				mustWrite(t, filepath.Join(runDir, "validation_summary.json"), string(summary)+"\n")
				if tc.history == 2 || (tc.history == 1 && i == 1) {
					history += string(summary) + "\n"
				}
			}
			if tc.history > 0 {
				mustWrite(t, filepath.Join(root, "output/validation/history.jsonl"), history)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, "bash", "scripts/analyze_validation.sh")
			cmd.Dir = root
			out, err := cmd.CombinedOutput()
			if ctx.Err() != nil || cmd.ProcessState == nil {
				t.Fatalf("analyzer did not finish: %v; %s", err, out)
			}
			if got := cmd.ProcessState.ExitCode(); got != tc.wantExit {
				t.Errorf("exit = %d, want %d; %s", got, tc.wantExit, out)
			}
			var gotRegressions []string
			for _, line := range strings.Split(string(out), "\n") {
				if strings.HasPrefix(line, "- [") {
					gotRegressions = append(gotRegressions, strings.TrimPrefix(line, "- "))
				}
			}
			// Exact diagnostics also forbid blaming a passing or absent package.
			if strings.Join(gotRegressions, "\n") != strings.Join(tc.regressions, "\n") {
				t.Errorf("regressions = %q, want %q; %s", gotRegressions, tc.regressions, out)
			}
			if !strings.Contains(string(out), tc.message) {
				t.Errorf("missing %q: %s", tc.message, out)
			}
		})
	}
}
