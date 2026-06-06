package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommandSurfaceExposesFyneEntrypoints(t *testing.T) {
	root := repoRoot()
	pkgs := listCommandPackages(t, root)
	want := []string{
		"fecim-lattice-tools/cmd/fecim-lattice-tools",
		"fecim-lattice-tools/cmd/fecim-lattice-tools-fyne",
		"fecim-lattice-tools/module1-hysteresis/cmd/hysteresis-fyne",
		"fecim-lattice-tools/module4-circuits/cmd/circuits-gui-fyne",
	}
	for _, pkg := range want {
		if !containsExactPackage(pkgs, pkg) {
			t.Fatalf("command surface missing Fyne entrypoint %s; packages=%v", pkg, pkgs)
		}
	}
}

func TestDefaultCommandDoesNotDependOnGogpuApp(t *testing.T) {
	root := repoRoot()
	for _, dep := range listDeps(t, root, "fecim-lattice-tools/cmd/fecim-lattice-tools") {
		if strings.HasPrefix(dep, "fecim-lattice-tools/internal/"+"go"+"gpuapp") {
			t.Fatalf("default command must not depend on retired UI app surface after Fyne rollback; found %s", dep)
		}
	}
}

func listCommandPackages(t *testing.T, root string) []string {
	t.Helper()
	args := []string{
		"list",
		"-e",
		"./cmd/...",
		"./module1-hysteresis/cmd/...",
		"./module2-crossbar/cmd/...",
		"./module3-mnist/cmd/...",
		"./module4-circuits/cmd/...",
		"./module5-comparison/cmd/...",
		"./module6-eda/cmd/...",
	}
	cmd := exec.Command("go", args...)
	cmd.Dir = root
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		t.Fatalf("go list command packages failed: %v\n%s", err, out)
	}
	return strings.Fields(string(out))
}

func packageDir(t *testing.T, root string, pkg string) string {
	t.Helper()
	cmd := exec.Command("go", "list", "-e", "-f", "{{.Dir}}", pkg)
	cmd.Dir = root
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		t.Fatalf("go list -f .Dir %s failed: %v\n%s", pkg, err, out)
	}
	return strings.TrimSpace(string(out))
}

func listDeps(t *testing.T, root string, pkg string) []string {
	t.Helper()
	cmd := exec.Command("go", "list", "-e", "-deps", pkg)
	cmd.Dir = root
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		t.Fatalf("go list -deps %s failed: %v\n%s", pkg, err, out)
	}
	return strings.Fields(string(out))
}

func containsExactPackage(pkgs []string, want string) bool {
	for _, pkg := range pkgs {
		if pkg == want {
			return true
		}
	}
	return false
}

func repoRoot() string {
	return filepath.Clean(filepath.Join("..", ".."))
}
