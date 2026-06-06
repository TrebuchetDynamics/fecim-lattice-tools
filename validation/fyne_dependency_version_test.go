package validation

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
)

func TestFyneDependencyTracksPatchedUIToolkitRelease(t *testing.T) {
	modBytes, err := os.ReadFile(filepath.Join("..", "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}

	version := findRequiredFyneVersion(string(modBytes))
	if version == nil {
		t.Fatal("go.mod must require fyne.io/fyne/v2")
	}

	minimum := semanticVersion{major: 2, minor: 7, patch: 4}
	if version.lessThan(minimum) {
		t.Fatalf("fyne.io/fyne/v2 = v%d.%d.%d, want at least v%d.%d.%d for RichText/Tree/Accordion/progress fixes documented in fyne-guide",
			version.major, version.minor, version.patch,
			minimum.major, minimum.minor, minimum.patch)
	}
}

type semanticVersion struct {
	major int
	minor int
	patch int
}

func (v semanticVersion) lessThan(other semanticVersion) bool {
	if v.major != other.major {
		return v.major < other.major
	}
	if v.minor != other.minor {
		return v.minor < other.minor
	}
	return v.patch < other.patch
}

func findRequiredFyneVersion(mod string) *semanticVersion {
	matches := regexp.MustCompile(`(?m)^\s*fyne\.io/fyne/v2\s+v(\d+)\.(\d+)\.(\d+)\b`).FindStringSubmatch(mod)
	if matches == nil {
		return nil
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])
	return &semanticVersion{major: major, minor: minor, patch: patch}
}
