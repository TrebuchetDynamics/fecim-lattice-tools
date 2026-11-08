package scripts_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// agentsPackagePathSurfaces are the module AGENTS.md guides whose
// path-qualified *.go references must resolve.
//
// Root AGENTS.md and module6-eda/AGENTS.md are deliberately excluded: they are
// held by the staged-proposals-validation and split-eda-export worktrees, so
// repairing their paths here would conflict with that in-flight work.
var agentsPackagePathSurfaces = []string{
	"module1-hysteresis/AGENTS.md",
	"module2-crossbar/AGENTS.md",
	"module2-crossbar/pkg/gui/AGENTS.md",
	"module3-mnist/AGENTS.md",
	"module7-docs/AGENTS.md",
	"validation/AGENTS.md",
}

// goFileBacktickPattern matches a backtick-quoted token ending in .go.
var goFileBacktickPattern = regexp.MustCompile("`([^`]*\\.go)`")

// TestAgentsPackagePaths fails when a module AGENTS.md guide points at a
// path-qualified *.go file that does not exist relative to the guide's
// directory or the repository root. Bare filenames (no slash) are skipped:
// they name files inside the guide's own package.
func TestAgentsPackagePaths(t *testing.T) {
	root := repoRoot(t)
	for _, name := range agentsPackagePathSurfaces {
		path := filepath.Join(root, filepath.FromSlash(name))
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		dir := filepath.Dir(path)
		for _, match := range goFileBacktickPattern.FindAllStringSubmatch(string(body), -1) {
			ref := strings.TrimSpace(match[1])
			if !strings.Contains(ref, "/") {
				continue
			}
			if !agentsPathExists(filepath.Join(dir, filepath.FromSlash(ref))) &&
				!agentsPathExists(filepath.Join(root, filepath.FromSlash(ref))) {
				t.Errorf("%s references missing Go path %q", name, ref)
			}
		}
	}
}

func agentsPathExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
