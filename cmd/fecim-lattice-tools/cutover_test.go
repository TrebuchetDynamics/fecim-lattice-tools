package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestCanonicalCommandImportsFyne(t *testing.T) {
	cmd := exec.Command("go", "list", "-e", "-deps", "./cmd/fecim-lattice-tools")
	cmd.Dir = repoRoot()
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		t.Fatalf("go list default command deps failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "fyne.io/fyne/v2") {
		t.Fatalf("canonical command must depend on Fyne after the Fyne rollback; deps did not include fyne.io/fyne/v2")
	}
}
