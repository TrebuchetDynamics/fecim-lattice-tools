package main

import (
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
