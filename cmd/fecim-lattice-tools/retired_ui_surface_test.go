package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRetiredUISurfaceIsFullyDeprecated(t *testing.T) {
	root := repoRoot()

	cmd := exec.Command("go", "list", "-m", "all")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list -m all failed: %v\n%s", err, out)
	}
	for _, mod := range strings.Fields(string(out)) {
		retiredModulePrefix := "github.com/" + "go" + "gpu/"
		if strings.HasPrefix(mod, retiredModulePrefix) {
			t.Fatalf("retired UI module remains in dependency graph: %s", mod)
		}
	}

	cmd = exec.Command("go", "list", "-e", "./...")
	cmd.Dir = root
	cmd.Env = os.Environ()
	out, err = cmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		t.Fatalf("go list ./... failed: %v\n%s", err, out)
	}
	for _, pkg := range strings.Fields(string(out)) {
		retiredInternalFragment := "/internal/" + "go" + "gpu"
		if strings.Contains(pkg, retiredInternalFragment) || pkg == "fecim-lattice-tools/cmd/fecim-screenshotter" {
			t.Fatalf("retired UI package remains in repo package graph: %s", pkg)
		}
	}

	for _, dir := range []string{
		"internal/" + "go" + "gpuapp",
		"internal/" + "go" + "gpuscreenshot",
		"cmd/fecim-screenshotter",
	} {
		if _, err := os.Stat(filepath.Join(root, dir)); !os.IsNotExist(err) {
			t.Fatalf("deprecated UI surface directory remains: %s", dir)
		}
	}
}
