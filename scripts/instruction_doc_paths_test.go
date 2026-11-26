package scripts_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// instructionDocSurfaces are the developer and AI instruction files whose docs/...
// references must resolve to existing repository paths.
var instructionDocSurfaces = []string{
	"AGENTS.md",
	".codex/AGENTS.md",
	".opencode/command/fecim-citation.md",
	".opencode/command/fecim-honesty-audit.md",
	".opencode/command/fecim-researcher.md",
	"config/AGENTS.md",
	"module5-comparison/AGENTS.md",
	"module7-docs/AGENTS.md",
}

// runtimeDocPaths are live output or loader-fallback locations the instructions
// describe, not links to tracked documentation. The module7-docs viewer loads
// "docs/documentation" and falls back to "docs" when it is absent.
var runtimeDocPaths = map[string]bool{
	"docs":               true,
	"docs/documentation": true,
}

// Match docs/... only as a standalone path, not the "docs/" suffix of a
// module directory such as module7-docs/pkg/gui.
var docsReferencePattern = regexp.MustCompile(`(?:^|[^A-Za-z0-9_.-])(docs/[A-Za-z0-9._/-]+)`)

// TestInstructionDocPaths keeps the canonical documentation paths readable by
// agents: a docs/... reference in an instruction file must point at a real path.
func TestInstructionDocPaths(t *testing.T) {
	root := repoRoot(t)
	for _, name := range instructionDocSurfaces {
		body, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for _, match := range docsReferencePattern.FindAllStringSubmatch(string(body), -1) {
			ref := strings.TrimRight(match[1], "/.")
			if ref == "" || runtimeDocPaths[ref] {
				continue
			}
			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(ref))); err != nil {
				t.Errorf("%s references missing docs path %q: %v", name, ref, err)
			}
		}
	}
}
