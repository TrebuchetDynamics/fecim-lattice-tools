package scripts_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// raceLanePackages extracts the quoted entries of the RACE_PKGS bash array
// from the documented headless race lane script.
func raceLanePackages(t *testing.T, script string) []string {
	t.Helper()
	var pkgs []string
	inArray := false
	for _, line := range strings.Split(script, "\n") {
		trimmed := strings.TrimSpace(line)
		if !inArray {
			if strings.HasPrefix(trimmed, "RACE_PKGS=(") {
				inArray = true
			}
			continue
		}
		if trimmed == ")" {
			break
		}
		if entry := strings.Trim(strings.TrimSpace(trimmed), `"`); entry != "" {
			pkgs = append(pkgs, entry)
		}
	}
	if !inArray {
		t.Fatal("scripts/ci/go-test-race.sh no longer defines a RACE_PKGS array")
	}
	return pkgs
}

func TestRaceLanePackagesResolve(t *testing.T) {
	root := repoRoot(t)
	script := readFile(t, filepath.Join(root, "scripts", "ci", "go-test-race.sh"))
	pkgs := raceLanePackages(t, script)

	if len(pkgs) == 0 {
		t.Fatal("RACE_PKGS is empty; the race lane would run no packages")
	}
	seen := make(map[string]bool, len(pkgs))
	for _, pkg := range pkgs {
		if seen[pkg] {
			t.Fatalf("RACE_PKGS contains duplicate entry %q", pkg)
		}
		seen[pkg] = true
	}

	args := append([]string{"list", "-tags=ci"}, pkgs...)
	cmd := exec.Command("go", args...)
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("RACE_PKGS entry does not resolve with go list:\n%s", out)
	}
}
