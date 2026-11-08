package scripts_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// shippedDocPathSurfaces are shipped user-facing text and maintainer comment
// surfaces whose docs/... references must resolve to existing repository paths.
// The exported metadata disclaimers, the comparison CLI banner, the Fyne
// info/glossary panels and the surrounding comments all ship this text.
var shippedDocPathSurfaces = []string{
	"shared/export/metadata/metadata.go",
	"module5-comparison/cmd/comparison/main.go",
	"module1-hysteresis/pkg/gui/info.go",
	"shared/widgets/help/glossary.go",
	"module5-comparison/pkg/comparison/architecture.go",
	"module5-comparison/pkg/gui/app.go",
	"module5-comparison/pkg/gui/embedded.go",
	"module5-comparison/pkg/gui/hero.go",
	"shared/physics/material_presets.go",
	"shared/widgets/validation_dashboard.go",
	"shared/widgets/display/validation_dashboard.go",
	"config/mnist.yaml",
	"config/physics/defaults/mnist.yaml",
	"module3-mnist/FEATURES.md",
}

// shippedDocPathSkipPrefixes are reference paths owned by other subsystems,
// not tracked documentation. The research PDF-collection path is still owned
// by the papers subsystem.
var shippedDocPathSkipPrefixes = []string{
	"docs/4-research/papers/",
}

// TestShippedDocPaths fails when a curated shipped/comment surface points at a
// docs/... path that does not exist in the tree.
func TestShippedDocPaths(t *testing.T) {
	root := repoRoot(t)
	for _, name := range shippedDocPathSurfaces {
		body, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for _, match := range docsReferencePattern.FindAllStringSubmatch(string(body), -1) {
			ref := strings.TrimRight(match[1], "/.")
			if ref == "" || hasSkipPrefix(ref) {
				continue
			}
			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(ref))); err != nil {
				t.Errorf("%s references missing docs path %q: %v", name, ref, err)
			}
		}
	}
}

func hasSkipPrefix(ref string) bool {
	for _, prefix := range shippedDocPathSkipPrefixes {
		if strings.HasPrefix(ref, prefix) {
			return true
		}
	}
	return false
}
