package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestNonLegacyCommandsDoNotDependOnLegacyGraphics(t *testing.T) {
	root := repoRoot()
	packages := listCommandPackages(t, root)
	disallowed := []string{
		"fyne.io/" + "fyne",
		"github.com/go-gl/glfw",
		"github.com/vulkan-go/vulkan",
		"fecim-lattice-tools/shared/theme",
		"fecim-lattice-tools/shared/themes",
		"fecim-lattice-tools/shared/widgets",
	}
	for _, pkg := range packages {
		if isLegacyFyneCommand(pkg) {
			continue
		}
		deps := listDeps(t, root, pkg)
		for _, dep := range deps {
			for _, needle := range disallowed {
				if strings.HasPrefix(dep, needle) {
					t.Fatalf("non-legacy command %s must not depend on legacy graphics surface %s", pkg, dep)
				}
			}
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
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		t.Fatalf("go list command packages failed: %v\n%s", err, out)
	}
	return strings.Fields(string(out))
}

func listDeps(t *testing.T, root string, pkg string) []string {
	t.Helper()
	cmd := exec.Command("go", "list", "-e", "-deps", pkg)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		t.Fatalf("go list -deps %s failed: %v\n%s", pkg, err, out)
	}
	return strings.Fields(string(out))
}

func isLegacyFyneCommand(pkg string) bool {
	return strings.Contains(pkg, "-fyne")
}

func repoRoot() string {
	return filepath.Clean(filepath.Join("..", ".."))
}
