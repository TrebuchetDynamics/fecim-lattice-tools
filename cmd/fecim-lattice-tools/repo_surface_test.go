package main

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultRepoGraphExposesFynePackages(t *testing.T) {
	root := repoRootForRepoSurface()
	pkgs := listRepoPackages(t, root)
	want := []string{
		"fecim-lattice-tools/module1-hysteresis/pkg/gui",
		"fecim-lattice-tools/module2-crossbar/pkg/gui",
		"fecim-lattice-tools/module3-mnist/pkg/gui",
		"fecim-lattice-tools/module4-circuits/pkg/gui",
		"fecim-lattice-tools/module5-comparison/pkg/gui",
		"fecim-lattice-tools/module6-eda/pkg/gui",
		"fecim-lattice-tools/module7-docs/pkg/gui",
		"fecim-lattice-tools/shared/widgets",
		"fecim-lattice-tools/shared/theme",
		"fecim-lattice-tools/shared/themes",
	}
	for _, pkg := range want {
		if !containsExactPackage(pkgs, pkg) {
			t.Fatalf("default repo graph missing restored Fyne package %s", pkg)
		}
	}
}

func TestGoFilesDoNotCarryLegacyFyneBuildTags(t *testing.T) {
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
		for _, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "package ") {
				break
			}
			if (strings.HasPrefix(trimmed, "//go:build") || strings.HasPrefix(trimmed, "// +build")) && strings.Contains(trimmed, "legacy_fyne") {
				rel, relErr := filepath.Rel(root, path)
				if relErr != nil {
					rel = path
				}
				t.Errorf("Go file carries obsolete legacy_fyne build tag: %s", rel)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk repo source files: %v", err)
	}
}

func TestLivingGuidanceDoesNotCallFyneRetired(t *testing.T) {
	root := repoRootForRepoSurface()
	files := []string{
		"README.md",
		"CONTRIBUTING.md",
		"docs/README.md",
		"docs/guides/README.md",
	}
	stalePhrases := []string{
		"removed legacy ui",
		"retired graphics",
		"legacy fyne shell",
		"requires `-tags legacy_fyne`",
		"requires -tags legacy_fyne",
	}
	for _, file := range files {
		body, err := os.ReadFile(filepath.Join(root, file))
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := strings.ToLower(string(body))
		for _, phrase := range stalePhrases {
			if strings.Contains(text, phrase) {
				t.Errorf("%s contains stale Fyne-retirement guidance %q", file, phrase)
			}
		}
	}
}

func listRepoPackages(t *testing.T, root string) []string {
	t.Helper()
	cmd := exec.Command("go", "list", "-e", "./...")
	cmd.Dir = root
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		t.Fatalf("go list ./... failed: %v\n%s", err, out)
	}
	return strings.Fields(string(out))
}

func isSkippedRepoSurfaceDir(name string) bool {
	switch name {
	case ".git", ".worktrees", "artifacts", "tmp", "output":
		return true
	default:
		return false
	}
}

func repoRootForRepoSurface() string {
	return filepath.Clean(filepath.Join("..", ".."))
}
