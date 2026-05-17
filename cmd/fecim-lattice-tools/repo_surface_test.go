package main

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestNonLegacyPackagesDoNotDependOnLegacyGraphics(t *testing.T) {
	root := repoRootForRepoSurface()
	for _, pkg := range listRepoPackages(t, root) {
		if isLegacyGraphicsPackage(pkg) {
			continue
		}
		for _, dep := range listRepoDeps(t, root, pkg) {
			if isLegacyGraphicsDependency(pkg, dep) {
				t.Errorf("non-legacy package %s must not depend on legacy graphics surface %s", pkg, dep)
			}
		}
	}
}

func TestDefaultRepoGraphDoesNotExposeLegacyFynePackages(t *testing.T) {
	root := repoRootForRepoSurface()
	for _, pkg := range listRepoPackages(t, root) {
		if isLegacyGraphicsPackage(pkg) {
			t.Fatalf("default repo graph must not expose legacy Fyne package %s", pkg)
		}
	}
}

func TestFyneImportsAreLegacyTagged(t *testing.T) {
	root := repoRootForRepoSurface()
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if isSkippedRepoSurfaceDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		source := string(data)
		importsFyne, err := fileImportsFyne(path, data)
		if err != nil {
			return err
		}
		if importsFyne && !hasLegacyFyneBuildTag(source) {
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				rel = path
			}
			t.Errorf("Go file importing Fyne must be tagged legacy_fyne: %s", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk repo source files: %v", err)
	}
}

func TestLivingGuidanceUsesCanonicalGogpuSurface(t *testing.T) {
	root := repoRootForRepoSurface()
	files := []string{
		"CONTRIBUTING.md",
		"tools/fecim-skills/_shared/fecim-context.md",
		"tools/fecim-skills/fecim-builder/SKILL.md",
		"tools/fecim-skills/fecim-gogpu-migrate/SKILL.md",
		"tools/fecim-skills/fecim-labtester/SKILL.md",
	}
	stalePhrases := []string{
		"current default desktop app remains the Fyne shell",
		"future zero-CGO",
		"cmd/fecim-lattice-tools-next",
		"make test-next-ui",
		"Next gogpu/ui shell",
		"Future shell",
		"Legacy Fyne shell: `cmd/fecim-lattice-tools`",
		"placeholder path until it reaches module parity",
	}
	for _, file := range files {
		body, err := os.ReadFile(filepath.Join(root, file))
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(body)
		for _, phrase := range stalePhrases {
			if strings.Contains(text, phrase) {
				t.Errorf("%s contains stale gogpu/Fyne guidance %q", file, phrase)
			}
		}
	}
}

func listRepoPackages(t *testing.T, root string) []string {
	t.Helper()
	cmd := exec.Command("go", "list", "-e", "./...")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		t.Fatalf("go list ./... failed: %v\n%s", err, out)
	}
	return strings.Fields(string(out))
}

func listRepoDeps(t *testing.T, root string, pkg string) []string {
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

func isLegacyGraphicsDependency(pkg string, dep string) bool {
	disallowed := []string{
		"fyne.io/" + "fyne",
		"github.com/go-gl/glfw",
		"fecim-lattice-tools/shared/theme",
		"fecim-lattice-tools/shared/themes",
		"fecim-lattice-tools/shared/widgets",
	}
	for _, needle := range disallowed {
		if strings.HasPrefix(dep, needle) {
			return true
		}
	}
	if strings.HasPrefix(dep, "github.com/vulkan-go/vulkan") && !isAllowedVulkanComputePackage(pkg) {
		return true
	}
	return false
}

func isAllowedVulkanComputePackage(pkg string) bool {
	return pkg == "fecim-lattice-tools/shared/compute" ||
		pkg == "fecim-lattice-tools/module4-circuits/pkg/gpuperiph"
}

func isLegacyGraphicsPackage(pkg string) bool {
	if strings.Contains(pkg, "-fyne") {
		return true
	}
	legacyAreas := []string{
		"/pkg/gui",
		"/shared/theme",
		"/shared/themes",
		"/shared/widgets",
	}
	for _, area := range legacyAreas {
		if strings.Contains(pkg, area) {
			return true
		}
	}
	return false
}

func isSkippedRepoSurfaceDir(name string) bool {
	switch name {
	case ".git", ".worktrees", "artifacts", "tmp":
		return true
	default:
		return false
	}
}

func hasLegacyFyneBuildTag(source string) bool {
	for _, line := range strings.Split(source, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "package ") {
			return false
		}
		if strings.HasPrefix(trimmed, "//go:build") {
			return strings.Contains(trimmed, "legacy_fyne")
		}
	}
	return false
}

func fileImportsFyne(path string, data []byte) (bool, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, data, parser.ImportsOnly)
	if err != nil {
		return false, err
	}
	for _, imported := range file.Imports {
		importPath, err := strconv.Unquote(imported.Path.Value)
		if err != nil {
			return false, err
		}
		if strings.HasPrefix(importPath, "fyne.io/"+"fyne/v2") {
			return true, nil
		}
	}
	return false, nil
}

func repoRootForRepoSurface() string {
	return filepath.Clean(filepath.Join("..", ".."))
}
