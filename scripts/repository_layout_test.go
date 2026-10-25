package scripts_test

import (
	"debug/elf"
	"encoding/json"
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
		"agent-test-loop.sh":      "scripts/agent-test-loop.sh",
		"commit-push.sh":          "scripts/commit-push.sh",
		"crucible":                "tools/crucible",
		"opensource":              "tools/opensource",
		"paper":                   "docs/research/paper",
		"presenter-script.md":     "docs/guides/presentations/presenter-script.md",
		"prompts":                 "tools/prompts",
		"screenshots":             "docs/guides/assets/",
		"CONTEXT.md":              "docs/internals/CONTEXT.md",
		"TODO.md":                 "docs/internals/TODO.md",
		"status.md":               "docs/internals/status.md",
		"shared-refactor-plan.md": "docs/internals/architecture/shared-refactor-plan.md",
		"fyne-guide":              "docs/internals/gui/fyne-guide/",
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

// Parse the ingestion manifests without importing Cognee or making network calls.
// A relocation must not silently turn these documents into skipped input.
func TestKnowledgeIngestionIncludesRelocatedInternalNotes(t *testing.T) {
	root := repoRoot(t)
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 is required to inspect optional Cognee manifests")
	}
	for _, tc := range []struct {
		script, variable string
		want             []string
	}{
		{"cognee-ingest.py", "DOCS", []string{"docs/internals/status.md"}},
		{"cognee-ingest-full.py", "DOCS_ROOT", []string{"docs/internals/TODO.md", "docs/internals/status.md"}},
	} {
		t.Run(tc.script, func(t *testing.T) {
			cmd := exec.Command(python, "-c", `
import ast, json, pathlib, sys
tree = ast.parse(pathlib.Path(sys.argv[1]).read_text())
for node in tree.body:
    if isinstance(node, ast.Assign) and any(isinstance(t, ast.Name) and t.id == sys.argv[2] for t in node.targets):
        print(json.dumps(ast.literal_eval(node.value)))
        break
else:
    raise SystemExit("manifest not found")
`, filepath.Join(root, "scripts", tc.script), tc.variable)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("parse manifest: %v: %s", err, out)
			}
			var paths []string
			if err := json.Unmarshal(out, &paths); err != nil {
				t.Fatal(err)
			}
			for _, want := range tc.want {
				found := false
				for _, path := range paths {
					if path == "TODO.md" || path == "status.md" {
						t.Errorf("manifest still references relocated root note %s", path)
					}
					found = found || path == want
				}
				if !found {
					t.Errorf("manifest missing %s", want)
				}
				if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(want))); err != nil {
					t.Errorf("ingestion input unavailable: %v", err)
				}
			}
		})
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
