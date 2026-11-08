package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var (
	goTestLinePattern = regexp.MustCompile(`go test\b[^\n]*`)
	goTestSelector    = regexp.MustCompile(`\./[^\s'"]+`)
	guiTestMarkers    = []string{"E2EGUI", "VisualXvfb", "UICrawler", "LayoutAudit"}
)

// TestTestingDocPaths guards the offline test guide against selectors that no
// longer name a Go package, and against GUI selectors that silently match no test.
func TestTestingDocPaths(t *testing.T) {
	root := repoRoot(t)
	body, err := os.ReadFile(filepath.Join(root, "docs", "internals", "testing", "TEST_GUIDE.md"))
	if err != nil {
		t.Fatal(err)
	}

	selectors := make(map[string]struct{})
	for _, line := range goTestLinePattern.FindAllString(string(body), -1) {
		for _, selector := range goTestSelector.FindAllString(line, -1) {
			selectors[selector] = struct{}{}
		}
	}
	if len(selectors) == 0 {
		t.Fatal("no go test selectors found in TEST_GUIDE.md")
	}

	for selector := range selectors {
		cmd := exec.Command("go", "list", selector)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Errorf("go test selector %s does not resolve to a package: %v\n%s", selector, err, out)
		}
	}

	for _, line := range goTestLinePattern.FindAllString(string(body), -1) {
		marker := ""
		for _, candidate := range guiTestMarkers {
			if strings.Contains(line, candidate) {
				marker = candidate
				break
			}
		}
		if marker == "" {
			continue
		}
		for _, selector := range goTestSelector.FindAllString(line, -1) {
			cmd := exec.Command("go", "test", "-list", ".*", selector)
			cmd.Dir = root
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("go test -list for %s selector %s failed: %v\n%s", marker, selector, err, out)
				continue
			}
			if !strings.Contains(string(out), marker) {
				t.Errorf("GUI selector %s for %s lists no matching test:\n%s", selector, marker, out)
			}
		}
	}
}
