package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleasedCommandSurfaceDoesNotExposeNextWrappers(t *testing.T) {
	cmd := exec.Command("go", "list", "-e", "./cmd/...")
	cmd.Dir = filepath.Clean(filepath.Join("..", ".."))
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		t.Fatalf("go list ./cmd/... failed: %v\n%s", err, out)
	}
	for _, pkg := range strings.Fields(string(out)) {
		if strings.HasSuffix(pkg, "-next") {
			t.Fatalf("released command surface must not expose transition wrapper %s", pkg)
		}
	}
}

func TestReleasedCommandSurfaceDoesNotExposeLegacyFyneCommands(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	for _, pkg := range listCommandPackages(t, root) {
		if strings.Contains(pkg, "-fyne") {
			t.Fatalf("released command surface must not expose legacy Fyne command %s", pkg)
		}
	}
}
