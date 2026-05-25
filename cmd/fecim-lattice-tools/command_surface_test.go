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
		"fecim-lattice-tools/internal/legacycommand",
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

func TestNonLegacyCommandsDoNotAdvertiseLegacyFyneEntrypoints(t *testing.T) {
	root := repoRoot()
	for _, pkg := range listCommandPackages(t, root) {
		if isLegacyFyneCommand(pkg) {
			continue
		}
		dir := packageDir(t, root, pkg)
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			body, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			text := string(body)
			if strings.Contains(text, "-fyne") || strings.Contains(text, "legacy Fyne") {
				t.Fatalf("non-legacy command %s must not advertise legacy Fyne entrypoints in %s", pkg, path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
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

func packageDir(t *testing.T, root string, pkg string) string {
	t.Helper()
	cmd := exec.Command("go", "list", "-e", "-f", "{{.Dir}}", pkg)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
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
