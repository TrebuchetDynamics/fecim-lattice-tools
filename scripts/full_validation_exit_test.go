package scripts_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestFullValidationExit(t *testing.T) {
	runner := readFile(t, filepath.Join(repoRoot(t), "scripts/run_full_validation.sh"))
	lanes := []struct{ path, key string }{
		{"module1-hysteresis", "m1"}, {"module4-circuits", "m4"},
		{"module6-eda", "m6"}, {"shared", "shared"}, {"validation", "validation"},
	}
	type counts struct {
		Pass     int  `json:"pass"`
		Fail     int  `json:"fail"`
		Skip     int  `json:"skip"`
		ExitCode *int `json:"exit_code"`
	}
	passSkip := "{\"Action\":\"pass\",\"Test\":\"TestPass\"}\n{\"Action\":\"skip\",\"Test\":\"TestSkip\"}\n"
	type testCase struct {
		name   string
		lane   int
		status int
		raw    string
		want   counts
	}
	cases := []testCase{
		{"success", -1, 0, "", counts{}},
		{"test failure", 1, 1, passSkip + "{\"Action\":\"fail\",\"Test\":\"TestFail\"}\n", counts{Pass: 1, Fail: 1, Skip: 1}},
		{"build failure", 2, 1, "{\"Action\":\"output\",\"Output\":\"build failed\"}\n{\"Action\":\"fail\",\"Package\":\"fixture\"}\n", counts{}},
		{"setup error", 3, 127, "go: command/setup error\n", counts{}},
	}
	for i, lane := range lanes {
		cases = append(cases, testCase{"timeout " + lane.key, i, 124, passSkip + "timeout diagnostic\n", counts{Pass: 1, Skip: 1}})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			mustWrite(t, filepath.Join(root, "scripts/run_full_validation.sh"), runner)
			mustWrite(t, filepath.Join(root, "bin/timeout"), `#!/bin/bash
printf '%s\n' "$*" >> "$FIXTURE/invocations"
shift
exec "$@"
`)
			mustWrite(t, filepath.Join(root, "bin/go"), `#!/bin/bash
lane="${@: -1}"
lane="${lane#./}"
lane="${lane%/...}"
cat "$FIXTURE/$lane.raw"
printf 'stderr for %s\n' "$lane" >&2
exit "$(cat "$FIXTURE/$lane.status")"
`)
			mustWrite(t, filepath.Join(root, "bin/git"), "#!/bin/bash\nprintf 'fixture-commit\\n'\n")
			wantModules := make(map[string]counts)
			wantLogs := make(map[string]string)
			wantInvocations := ""
			wantPass, wantFail := 0, 0
			for i, lane := range lanes {
				raw, status, want := passSkip, 0, counts{Pass: 1, Skip: 1}
				if i == tc.lane {
					raw, status, want = tc.raw, tc.status, tc.want
				}
				// An exit-zero lane with fail events still uses command status for the verdict.
				if tc.lane == -1 && i == 2 {
					raw += "{\"Action\":\"fail\",\"Test\":\"TestFixtureFail\"}\n"
					want.Fail++
				}
				want.ExitCode = new(int)
				*want.ExitCode = status
				wantModules[lane.key] = want
				wantPass += want.Pass
				wantFail += want.Fail
				mustWrite(t, filepath.Join(root, lane.path+".raw"), raw)
				mustWrite(t, filepath.Join(root, lane.path+".status"), fmt.Sprint(status))
				wantLogs[lane.key] = raw + "stderr for " + lane.path + "\n"
				wantInvocations += "37s go test -short -count=1 -json ./" + lane.path + "/...\n"
			}
			// Existing history must be retained, not replaced.
			mustWrite(t, filepath.Join(root, "output/validation/history.jsonl"), "{\"previous\":true}\n")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, "bash", filepath.Join(root, "scripts/run_full_validation.sh"))
			cmd.Dir = root
			for _, env := range os.Environ() {
				if !strings.HasPrefix(env, "PATH=") && !strings.HasPrefix(env, "VALIDATION_TEST_TIMEOUT=") && !strings.HasPrefix(env, "FIXTURE=") {
					cmd.Env = append(cmd.Env, env)
				}
			}
			cmd.Env = append(cmd.Env, "PATH="+filepath.Join(root, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"), "FIXTURE="+root, "VALIDATION_TEST_TIMEOUT=37s")
			out, err := cmd.CombinedOutput()
			if ctx.Err() != nil || cmd.ProcessState == nil {
				t.Fatalf("runner did not finish: %v; %s", err, out)
			}
			wantExit := 0
			if tc.status != 0 {
				wantExit = 1
			}
			if got := cmd.ProcessState.ExitCode(); got != wantExit {
				t.Errorf("exit = %d, want %d; %s", got, wantExit, out)
			}
			if got := readFile(t, filepath.Join(root, "invocations")); got != wantInvocations {
				t.Errorf("lane commands = %q, want %q", got, wantInvocations)
			}
			if tc.status == 124 && !strings.Contains(string(out), lanes[tc.lane].path+" timed out after 37s") {
				t.Errorf("missing accurate timeout warning: %s", out)
			}
			files, err := filepath.Glob(filepath.Join(root, "output/validation/*/validation_summary.json"))
			if err != nil || len(files) != 1 {
				t.Fatalf("summary files = %v: %v", files, err)
			}
			summary := readFile(t, files[0])
			var got struct {
				Timestamp string            `json:"timestamp"`
				Commit    string            `json:"commit"`
				Modules   map[string]counts `json:"modules"`
				TotalPass int               `json:"total_pass"`
				TotalFail int               `json:"total_fail"`
			}
			if err := json.Unmarshal([]byte(summary), &got); err != nil {
				t.Fatal(err)
			}
			if _, err := time.Parse("2006-01-02-1504", got.Timestamp); err != nil {
				t.Errorf("timestamp: %v", err)
			}
			if got.Timestamp != filepath.Base(filepath.Dir(files[0])) || got.Commit != "fixture-commit" || got.TotalPass != wantPass || got.TotalFail != wantFail || !reflect.DeepEqual(got.Modules, wantModules) {
				t.Errorf("unexpected summary: %s; want modules %+v", summary, wantModules)
			}
			for key, raw := range wantLogs {
				if got := readFile(t, filepath.Join(filepath.Dir(files[0]), key+"_results.json")); got != raw {
					t.Errorf("%s raw log = %q, want %q", key, got, raw)
				}
			}
			if got := readFile(t, filepath.Join(root, "output/validation/history.jsonl")); got != "{\"previous\":true}\n"+summary+"\n" {
				t.Errorf("history differs from appended summary: %q", got)
			}
		})
	}
}
