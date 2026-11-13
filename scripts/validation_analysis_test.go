package scripts_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
		Test    string `json:"Test,omitempty"`
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
	writeOrphan := func(name string, e event) func(*testing.T, string) {
		return func(t *testing.T, root string) {
			line, err := json.Marshal(e)
			if err != nil {
				t.Fatal(err)
			}
			mustWrite(t, filepath.Join(root, "output/validation", name, "shared_results.json"), string(line)+"\n")
		}
	}
	remove := func(name string) func(*testing.T, string) {
		return func(t *testing.T, root string) {
			if err := os.RemoveAll(filepath.Join(root, "output/validation", name)); err != nil {
				t.Fatal(err)
			}
		}
	}
	changeSummary := func(index int, key string, value any) func(*testing.T, string) {
		return func(t *testing.T, root string) {
			path := filepath.Join(root, "output/validation/history.jsonl")
			lines := strings.Split(strings.TrimSpace(readFile(t, path)), "\n")
			var summary map[string]any
			if err := json.Unmarshal([]byte(lines[index]), &summary); err != nil {
				t.Fatal(err)
			}
			if value == nil {
				delete(summary, key)
			} else {
				summary[key] = value
			}
			line, err := json.Marshal(summary)
			if err != nil {
				t.Fatal(err)
			}
			lines[index] = string(line)
			mustWrite(t, path, strings.Join(lines, "\n")+"\n")
		}
	}
	laneStatuses := func(lane string, previous, current any) func(*testing.T, string) {
		return func(t *testing.T, root string) {
			path := filepath.Join(root, "output/validation/history.jsonl")
			lines := strings.Split(strings.TrimSpace(readFile(t, path)), "\n")
			for i, status := range []any{previous, current} {
				var summary struct {
					Timestamp string                    `json:"timestamp"`
					Modules   map[string]map[string]any `json:"modules"`
				}
				if err := json.Unmarshal([]byte(lines[i]), &summary); err != nil {
					t.Fatal(err)
				}
				counts := summary.Modules["shared"]
				if status != nil {
					counts["exit_code"] = status
				}
				summary.Modules = map[string]map[string]any{lane: counts}
				runDir := filepath.Join(root, "output/validation", summary.Timestamp)
				if lane != "shared" {
					if err := os.Rename(filepath.Join(runDir, "shared_results.json"), filepath.Join(runDir, lane+"_results.json")); err != nil {
						t.Fatal(err)
					}
				}
				line, err := json.Marshal(summary)
				if err != nil {
					t.Fatal(err)
				}
				lines[i] = string(line)
				mustWrite(t, filepath.Join(runDir, "validation_summary.json"), string(line)+"\n")
			}
			mustWrite(t, path, strings.Join(lines, "\n")+"\n")
		}
	}
	type analysisCase struct {
		name        string
		previous    []event
		current     []event
		history     int
		wantExit    int
		regressions []string
		message     string
		timestamps  []string
		setup       func(*testing.T, string)
	}
	cases := []analysisCase{
		{"collision fail then pass", []event{aPass, bPass}, []event{aFail, bPass, extra}, 2, 1, []string{namedRegression}, "", nil, nil},
		{"collision pass then fail", []event{aPass, bPass}, []event{bPass, aFail, extra}, 2, 1, []string{namedRegression}, "", nil, nil},
		{"different package between runs", []event{{pkgA, "TestX", "pass"}}, []event{{pkgB, "TestX", "fail"}, extra}, 2, 0, nil, "No regressions found", nil, nil},
		{"same package", []event{aPass}, []event{aFail, extra}, 2, 1, []string{namedRegression}, "", nil, nil},
		{"module pass drop only", []event{aPass, bPass}, []event{aPass}, 2, 1, []string{"[MODULE REGRESSION] module=shared: pass count 2 -> 1"}, "", nil, nil},
		{"no regression", []event{aPass, bPass}, []event{bPass, aPass}, 2, 0, nil, "No regressions found", nil, nil},
		{"pass to skip", []event{aPass}, []event{{pkgA, "TestClamp01", "skip"}, extra}, 2, 0, nil, "No regressions found", nil, nil},
		{"skip to fail", []event{{pkgA, "TestClamp01", "skip"}}, []event{aFail}, 2, 0, nil, "No regressions found", nil, nil},
		{"legacy regression beside named pass", []event{legacyPass, aPass}, []event{legacyFail, aPass, extra}, 2, 1, []string{legacyRegression}, "", nil, nil},
		{"legacy not named", []event{legacyPass}, []event{aFail, extra}, 2, 0, nil, "No regressions found", nil, nil},
		{"named not legacy", []event{aPass}, []event{legacyFail, extra}, 2, 0, nil, "No regressions found", nil, nil},
		{"missing history", nil, nil, 0, 0, nil, "No history file found", nil, nil},
		{"one entry history", nil, []event{aFail}, 1, 0, nil, "No regression baseline available", nil, nil},
		{"interrupted next run", []event{aPass}, []event{aFail, extra}, 2, 1, []string{namedRegression}, "", nil, writeOrphan("2026-01-01-1202", aPass)},
		{"unrelated later directory", []event{aPass}, []event{aFail, extra}, 2, 1, []string{namedRegression}, "", nil, writeOrphan("zzz", aPass)},
		{"history append order", []event{aPass}, []event{aFail, extra}, 2, 1, []string{namedRegression}, "", []string{"2026-01-01-1201", "2026-01-01-1200"}, nil},
		{"unrelated failure", []event{aPass}, []event{aPass}, 2, 0, nil, "No regressions found", nil, writeOrphan("zzz", aFail)},
		{"missing directory", []event{aPass}, []event{aPass}, 2, 1, nil, "Comparison unavailable", nil, remove("2026-01-01-1201")},
		{"missing previous log", []event{aPass}, []event{aPass}, 2, 1, nil, "Comparison unavailable", nil, remove("2026-01-01-1200/shared_results.json")},
		{"missing current log", []event{aPass}, []event{aPass}, 2, 1, nil, "Comparison unavailable", nil, remove("2026-01-01-1201/shared_results.json")},
		{"duplicate minute", []event{aPass}, []event{aPass}, 2, 1, nil, "Comparison unavailable", []string{"2026-01-01-1200", "2026-01-01-1200"}, nil},
		{"previous-only lane required", []event{aPass}, []event{aPass}, 2, 1, nil, "Comparison unavailable", nil, changeSummary(0, "modules", map[string]any{"shared": map[string]int{"pass": 1}, "m1": map[string]int{"pass": 0}})},
		{"current-only lane required", []event{aPass}, []event{aPass}, 2, 1, nil, "Comparison unavailable", nil, changeSummary(1, "modules", map[string]any{"shared": map[string]int{"pass": 1}, "m4": map[string]int{"pass": 0}})},
		{"missing identity", []event{aPass}, []event{aPass}, 2, 1, nil, "Comparison unavailable", nil, changeSummary(1, "timestamp", nil)},
		{"non-string identity", []event{aPass}, []event{aPass}, 2, 1, nil, "Comparison unavailable", nil, changeSummary(0, "timestamp", 42)},
		{"traversal identity", []event{aPass}, []event{aPass}, 2, 1, nil, "Comparison unavailable", nil, changeSummary(1, "timestamp", "../2026-01-01-1201")},
		{"absolute identity", []event{aPass}, []event{aPass}, 2, 1, nil, "Comparison unavailable", nil, changeSummary(1, "timestamp", "/2026-01-01-1201")},
		{"arbitrary identity", []event{aPass}, []event{aPass}, 2, 1, nil, "Comparison unavailable", nil, changeSummary(1, "timestamp", "zzz")},
		{name: "symlink directory", previous: []event{aPass}, current: []event{aPass}, history: 2, wantExit: 1, message: "Comparison unavailable", setup: func(t *testing.T, root string) {
			remove("2026-01-01-1201")(t, root)
			if err := os.Symlink("2026-01-01-1200", filepath.Join(root, "output/validation/2026-01-01-1201")); err != nil {
				t.Fatal(err)
			}
		}},
	}
	laneRegression := func(lane string, status int) string {
		return fmt.Sprintf("[LANE EXIT REGRESSION] module=%s: exit_code 0 -> %d", lane, status)
	}
	cases = append(cases, analysisCase{
		name: "shared build failure", previous: []event{aPass, {pkgB, "", "pass"}},
		current: []event{aPass, {pkgB, "", "fail"}}, history: 2, wantExit: 1,
		regressions: []string{laneRegression("shared", 1)}, setup: laneStatuses("shared", 0, 1),
	})
	for _, lane := range []string{"m1", "m4", "m6", "shared", "validation"} {
		for _, status := range []int{124, 127} {
			cases = append(cases, analysisCase{
				name: fmt.Sprintf("%s exit %d", lane, status), previous: []event{aPass}, current: []event{aPass},
				history: 2, wantExit: 1, regressions: []string{laneRegression(lane, status)}, setup: laneStatuses(lane, 0, status),
			})
		}
	}
	for _, statuses := range []struct {
		name              string
		previous, current any
	}{
		{"success", 0, 0}, {"recovery", 1, 0}, {"unchanged failure", 124, 124},
		{"legacy previous", nil, 1}, {"legacy current", 0, nil},
		{"invalid previous string", "0", 1}, {"invalid previous boolean", false, 1},
		{"invalid current string", 0, "1"}, {"invalid current boolean", 0, true},
		{"invalid current negative", 0, -1}, {"invalid current range", 0, 256}, {"invalid current fraction", 0, 1.5},
	} {
		cases = append(cases, analysisCase{
			name: statuses.name, previous: []event{aPass}, current: []event{aPass}, history: 2,
			message: "No regressions found", setup: laneStatuses("shared", statuses.previous, statuses.current),
		})
	}
	cases = append(cases,
		analysisCase{name: "lane and test regression", previous: []event{aPass}, current: []event{aFail, extra}, history: 2, wantExit: 1,
			regressions: []string{laneRegression("shared", 1), namedRegression}, setup: laneStatuses("shared", 0, 1)},
		analysisCase{name: "lane and module regression", previous: []event{aPass, bPass}, current: []event{aPass}, history: 2, wantExit: 1,
			regressions: []string{"[MODULE REGRESSION] module=shared: pass count 2 -> 1", laneRegression("shared", 1)}, setup: laneStatuses("shared", 0, 1)},
	)
	historyPath := func(root string) string { return filepath.Join(root, "output/validation/history.jsonl") }
	for _, text := range []string{"", "\n \t\n"} {
		cases = append(cases, analysisCase{name: fmt.Sprintf("empty history %q", text), message: "No history entries present. No regression baseline available.",
			setup: func(t *testing.T, root string) { mustWrite(t, historyPath(root), text) }})
	}
	cases = append(cases,
		analysisCase{name: "sole summary without raw evidence", history: 1, message: "No regression baseline available", setup: func(t *testing.T, root string) {
			remove("2026-01-01-1200")(t, root)
			remove("2026-01-01-1201")(t, root)
		}},
		analysisCase{name: "blank separators and older corrupt record", previous: []event{aPass}, current: []event{aFail, extra}, history: 2, wantExit: 1,
			regressions: []string{namedRegression}, setup: func(t *testing.T, root string) {
				text := readFile(t, historyPath(root))
				mustWrite(t, historyPath(root), "{broken\n\n"+strings.ReplaceAll(text, "\n", "\n \t\n"))
			}},
	)
	invalidRecords := []struct{ name, text, context string }{
		{"truncated JSON", `{"timestamp":`, "invalid JSON"},
		{"null summary", `null`, "summary must be an object"},
		{"scalar summary", `42`, "summary must be an object"},
		{"string summary", `"summary"`, "summary must be an object"},
		{"list summary", `[]`, "summary must be an object"},
		{"missing timestamp", `{"modules":{"shared":{"pass":0}}}`, "timestamp"},
	}
	for _, modules := range []string{"", `,"modules":null`, `,"modules":[]`, `,"modules":{}`} {
		invalidRecords = append(invalidRecords, struct{ name, text, context string }{
			"modules " + modules, `{"timestamp":"2026-01-01-1201"` + modules + `}`, "modules"})
	}
	for _, lane := range []string{`null`, `[]`, `1`, `{}`, `{"pass":null}`, `{"pass":true}`, `{"pass":"1"}`, `{"pass":-1}`, `{"pass":1.5}`, `{"pass":1.0}`} {
		invalidRecords = append(invalidRecords, struct{ name, text, context string }{
			"lane " + lane, `{"timestamp":"2026-01-01-1201","modules":{"shared":` + lane + `}}`, "lane 'shared'"})
	}
	for _, invalid := range invalidRecords {
		for _, selection := range []string{"sole", "previous", "current"} {
			label, count := selection, 2
			if selection == "sole" {
				label, count = "current", 1
			}
			cases = append(cases, analysisCase{name: selection + " invalid " + invalid.name, history: count, wantExit: 1,
				message: "Comparison unavailable / incomplete: " + label + " " + invalid.context, setup: func(t *testing.T, root string) {
					lines := strings.Split(strings.TrimSpace(readFile(t, historyPath(root))), "\n")
					index := len(lines) - 1
					if selection == "previous" {
						index = 0
					}
					lines[index] = invalid.text
					if selection == "previous" {
						lines[index] = strings.ReplaceAll(lines[index], "2026-01-01-1201", "2026-01-01-1200")
					}
					// No trailing newline also models an interrupted final append.
					mustWrite(t, historyPath(root), strings.Join(lines, "\n"))
				}})
		}
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			mustWrite(t, filepath.Join(root, "scripts/analyze_validation.sh"), analyzer)
			history := ""
			timestamps := tc.timestamps
			if timestamps == nil {
				timestamps = []string{"2026-01-01-1200", "2026-01-01-1201"}
			}
			for i, events := range [][]event{tc.previous, tc.current} {
				timestamp := timestamps[i]
				counts := map[string]int{"pass": 0, "fail": 0, "skip": 0}
				raw := ""
				for _, e := range events {
					line, err := json.Marshal(e)
					if err != nil {
						t.Fatal(err)
					}
					raw += string(line) + "\n"
					if e.Test != "" {
						counts[e.Action]++
					}
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
			if tc.setup != nil {
				tc.setup(t, root)
			}
			before, _ := os.ReadFile(historyPath(root))
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
			after, _ := os.ReadFile(historyPath(root))
			if string(before) != string(after) {
				t.Error("analyzer mutated history")
			}
			if strings.Contains(string(out), "Traceback") {
				t.Errorf("uncaught traceback: %s", out)
			}
			if tc.history == 2 && !strings.Contains(tc.message, "Comparison unavailable") {
				for _, label := range []string{"Current timestamp:  " + timestamps[1], "Previous timestamp: " + timestamps[0]} {
					if !strings.Contains(string(out), label) {
						t.Errorf("missing %q: %s", label, out)
					}
				}
			}
			if tc.wantExit != 0 && (strings.Contains(string(out), "Result: No regressions found") || strings.Contains(string(out), "No regression baseline available")) {
				t.Errorf("false clean verdict: %s", out)
			}
			if !strings.Contains(string(out), tc.message) {
				t.Errorf("missing %q: %s", tc.message, out)
			}
		})
	}
}
