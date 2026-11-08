package scripts_test

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestToolchainDriftGoPinMatchesGoMod keeps the external-tool inventory Go pin
// in step with go.mod, so scripts/toolchain/drift_detector.sh reports a
// conforming Go toolchain as match instead of drift.
func TestToolchainDriftGoPinMatchesGoMod(t *testing.T) {
	root := repoRoot(t)

	pin := ""
	for _, line := range strings.Split(readFile(t, filepath.Join(root, "tools", "external", "README.md")), "\n") {
		if !strings.HasPrefix(line, "| Go toolchain |") {
			continue
		}
		fields := strings.Split(line, "|")
		if len(fields) < 3 {
			t.Fatalf("unparsable Go toolchain inventory row: %q", line)
		}
		pin = strings.Trim(strings.TrimSpace(fields[2]), "`")
		break
	}
	if pin == "" {
		t.Fatal("tools/external/README.md has no Go toolchain pin row")
	}
	// The drift detector allows a trailing .x wildcard; compare the prefix.
	pin = strings.TrimSuffix(pin, ".x")

	directive := ""
	for _, line := range strings.Split(readFile(t, filepath.Join(root, "go.mod")), "\n") {
		if fields := strings.Fields(line); len(fields) == 2 && fields[0] == "go" {
			directive = fields[1]
			break
		}
	}
	if directive == "" {
		t.Fatal("go.mod has no go directive")
	}

	if directive != pin && !strings.HasPrefix(directive, pin+".") {
		t.Fatalf("Go pin %q in tools/external/README.md does not prefix-match go.mod go directive %q", pin, directive)
	}
}
