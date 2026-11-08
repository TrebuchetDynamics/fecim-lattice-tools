package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func runPkgsum(t *testing.T, jsonl string) (string, error) {
	t.Helper()
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.jsonl")
	if err := os.WriteFile(in, []byte(jsonl), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}
	cmd := exec.Command("go", "run", ".", in)
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestPkgsum_ToleratesNonJSONNoise(t *testing.T) {
	jsonl := strings.Join([]string{
		"not-json-noise-line",
		`{"Action":"run","Package":"fecim-lattice-tools/foo"}`,
		`{"Action":"pass","Package":"fecim-lattice-tools/foo"}`,
	}, "\n") + "\n"

	out, err := runPkgsum(t, jsonl)
	if err != nil {
		t.Fatalf("expected success, got err=%v out=%s", err, out)
	}
	if !strings.Contains(out, "PKG_SUM pass=1 fail=0 skip=0 total=1") {
		t.Fatalf("unexpected summary: %s", out)
	}
}

func TestPkgsum_FailPackageReturnsNonZero(t *testing.T) {
	jsonl := strings.Join([]string{
		`{"Action":"run","Package":"fecim-lattice-tools/foo"}`,
		`{"Action":"fail","Package":"fecim-lattice-tools/foo"}`,
	}, "\n") + "\n"

	out, err := runPkgsum(t, jsonl)
	if err == nil {
		t.Fatalf("expected non-zero exit for failing package, out=%s", out)
	}
	if !strings.Contains(out, "PKG_SUM pass=0 fail=1 skip=0 total=1") {
		t.Fatalf("unexpected summary: %s", out)
	}
}

func TestRunPkgSumPackageTerminals(t *testing.T) {
	for _, tc := range []struct {
		name   string
		jsonl  string
		code   int
		stdout string
		stderr string
	}{
		{
			name:   "test pass without package completion",
			jsonl:  `{"Action":"pass","Package":"partial","Test":"TestOne"}`,
			code:   2,
			stderr: "ERROR: incomplete go test -json stream: packages without terminal summaries: partial\n",
		},
		{
			name:   "subtest skip without package completion",
			jsonl:  `{"Action":"skip","Package":"partial","Test":"TestOne/subtest"}`,
			code:   2,
			stderr: "ERROR: incomplete go test -json stream: packages without terminal summaries: partial\n",
		},
		{
			name: "completed package with unfinished test pass and skip",
			jsonl: `{"Action":"pass","Package":"complete"}
{"Action":"start","Package":"z-partial"}
{"Action":"pass","Package":"z-partial","Test":"TestOne"}
{"Action":"start","Package":"a-partial"}
{"Action":"skip","Package":"a-partial","Test":"TestOne/subtest"}`,
			code:   2,
			stderr: "ERROR: incomplete go test -json stream: packages without terminal summaries: a-partial, z-partial\n",
		},
		{
			name: "started and output only packages",
			jsonl: `{"Action":"start","Package":"started"}
{"Action":"output","Package":"output-only","Output":"interrupted"}`,
			code:   2,
			stderr: "ERROR: incomplete go test -json stream: packages without terminal summaries: output-only, started\n",
		},
		{
			name:   "test failure without package completion",
			jsonl:  `{"Action":"fail","Package":"partial","Test":"TestOne"}`,
			code:   2,
			stderr: "ERROR: incomplete go test -json stream: packages without terminal summaries: partial\n",
		},
		{
			name: "passing package with skipped subtest",
			jsonl: `{"Action":"skip","Package":"passing","Test":"TestOne/subtest"}
{"Action":"pass","Package":"passing"}
{"Action":"skip","Package":"passing","Test":"TestTwo"}`,
			stdout: "PKG_SUM pass=1 fail=0 skip=0 total=1\n",
		},
		{
			name: "complete interleaved packages and failure dominance",
			jsonl: `{"Action":"start","Package":"passing"}
{"Action":"start","Package":"failing"}
{"Action":"skip","Package":"passing","Test":"TestOne/subtest"}
{"Action":"fail","Package":"failing"}
{"Action":"skip","Package":"skipped"}
{"Action":"pass","Package":"failing"}
{"Action":"pass","Package":"passing","Test":"TestOne"}
{"Action":"skip","Package":"failing"}
{"Action":"pass","Package":"passing"}`,
			code:   1,
			stdout: "PKG_SUM pass=1 fail=1 skip=1 total=3\n",
		},
		{
			name: "test failure dominates package pass",
			jsonl: `{"Action":"fail","Package":"failing","Test":"TestOne/subtest"}
{"Action":"pass","Package":"failing"}`,
			code:   1,
			stdout: "PKG_SUM pass=0 fail=1 skip=0 total=1\n",
		},
		{
			name:   "empty input",
			code:   2,
			stderr: "ERROR: no package summaries parsed from go test -json stream\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, mode := range []string{"stdin", "dash", "file"} {
				t.Run(mode, func(t *testing.T) {
					var args []string
					input := tc.jsonl
					if mode == "dash" {
						args = []string{"-"}
					} else if mode == "file" {
						path := filepath.Join(t.TempDir(), "input.jsonl")
						if err := os.WriteFile(path, []byte(input), 0o600); err != nil {
							t.Fatal(err)
						}
						args = []string{path}
						input = "" // File input must not depend on stdin.
					}
					var stdout, stderr bytes.Buffer
					code := runPkgSum(args, strings.NewReader(input), &stdout, &stderr)
					if code != tc.code || stdout.String() != tc.stdout || stderr.String() != tc.stderr {
						t.Fatalf("got exit=%d stdout=%q stderr=%q; want exit=%d stdout=%q stderr=%q",
							code, &stdout, &stderr, tc.code, tc.stdout, tc.stderr)
					}
				})
			}
		})
	}
}

func TestRunPkgSumReportsOpenErrorWithoutExiting(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	missing := filepath.Join(t.TempDir(), "missing.jsonl")

	code := runPkgSum([]string{missing}, strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code=%d, want 2; stderr=%q", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout=%q, want empty output", stdout.String())
	}
	if !strings.Contains(stderr.String(), "open "+missing) {
		t.Fatalf("stderr=%q, want open error context for %s", stderr.String(), missing)
	}
}
