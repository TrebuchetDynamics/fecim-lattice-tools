package scripts_test

import (
	"debug/elf"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepositoryRootKeepsCollateralUnderPurposeDirectories(t *testing.T) {
	root := repoRoot(t)

	tracked := trackedPaths(t, root)
	forbiddenRoots := map[string]string{
		"agent-test-loop.sh":  "scripts/agent-test-loop.sh",
		"commit-push.sh":      "scripts/commit-push.sh",
		"crucible":            "tools/crucible",
		"notebook":            "docs/notebook",
		"opensource":          "tools/opensource",
		"paper":               "docs/paper",
		"presenter-script.md": "docs/presentations/presenter-script.md",
		"prompts":             "tools/prompts",
		"screenshots":         "docs/assets/reference-screenshots",
	}

	for path := range tracked {
		first := path
		if i := strings.Index(path, "/"); i >= 0 {
			first = path[:i]
		}
		if want, ok := forbiddenRoots[first]; ok {
			t.Fatalf("tracked root collateral %q should live under %s", path, want)
		}
	}

	for _, want := range forbiddenRoots {
		if _, ok := tracked[want]; ok {
			continue
		}
		prefix := strings.TrimSuffix(want, "/") + "/"
		found := false
		for path := range tracked {
			if strings.HasPrefix(path, prefix) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected tracked collateral under %s", want)
		}
	}
}

func TestRepositoryRootDoesNotTrackGeneratedELFBinaries(t *testing.T) {
	root := repoRoot(t)
	for path := range trackedPaths(t, root) {
		if strings.Contains(path, "/") {
			continue
		}
		file, err := elf.Open(filepath.Join(root, path))
		if err == nil {
			file.Close()
			t.Fatalf("tracked root ELF binary %q should be rebuilt into ignored artifacts, not committed", path)
		}
		// elf.Open returns format errors for normal text/config files.
		continue
	}
}

func TestGoSourceUsesCanonicalCollateralPaths(t *testing.T) {
	root := repoRoot(t)
	for path := range trackedPaths(t, root) {
		if !strings.HasSuffix(path, ".go") {
			continue
		}
		body, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		text := string(body)
		stale := map[string]string{
			"opensource": "tools/opensource/",
			"crucible":   "tools/crucible/",
			"prompts":    "tools/prompts/",
		}
		for rootDir, canonical := range stale {
			old := rootDir + "/"
			staleLiterals := []string{"\"" + old, "`" + old, "'" + old, " at " + old}
			for _, literal := range staleLiterals {
				if strings.Contains(text, literal) {
					t.Fatalf("%s references relocated root collateral %q; use %q", path, old, canonical)
				}
			}
		}
	}
}

func trackedPaths(t *testing.T, root string) map[string]struct{} {
	t.Helper()
	cmd := exec.Command("git", "ls-files", "--cached", "--others", "--exclude-standard")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git ls-files: %v", err)
	}
	paths := make(map[string]struct{})
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		path := filepath.Join(root, filepath.FromSlash(line))
		if _, err := os.Stat(path); err != nil {
			continue
		}
		paths[filepath.ToSlash(line)] = struct{}{}
	}
	return paths
}
