package validation

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type staleGogpuMigrationCodeFinding struct {
	Path   string
	Line   int
	Phrase string
	Text   string
}

func (f staleGogpuMigrationCodeFinding) String() string {
	return fmt.Sprintf("%s:%d contains %q: %s", f.Path, f.Line, f.Phrase, f.Text)
}

var staleGogpuMigrationCodePhrases = []string{
	"cmd/fecim-lattice-tools-next",
	"current Fyne",
	"Fyne remains",
	"future `gogpu/ui`",
	"future default",
	"future gogpu",
	"future zero",
	"future-default",
	"next shell",
	"reaches parity",
	"stable Fyne",
}

func TestStaleGogpuMigrationCodeWordingScannerScopesDefaultSurfaces(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "go.mod", "module example.com/codeguard\n")
	writeTestFile(t, root, filepath.Join("cmd", "fecim-lattice-tools", "surface_test.go"), "package main\n\nconst label = \"future default gogpu shell\"\n")
	writeTestFile(t, root, filepath.Join("cmd", "fecim-lattice-tools-fyne", "legacy_test.go"), "package main\n\nconst label = \"future default gogpu shell\"\n")
	writeTestFile(t, root, filepath.Join("validation", "guard_test.go"), "package validation\n\nconst fixture = \"future gogpu shell\"\n")

	findings, err := staleGogpuMigrationCodeWordingFindings(root)
	if err != nil {
		t.Fatalf("staleGogpuMigrationCodeWordingFindings error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("finding count = %d, want 1: %+v", len(findings), findings)
	}
	if findings[0].Path != filepath.ToSlash(filepath.Join("cmd", "fecim-lattice-tools", "surface_test.go")) {
		t.Fatalf("finding path = %q, want cmd/fecim-lattice-tools/surface_test.go", findings[0].Path)
	}
}

func TestDefaultGoSourcesAvoidStaleGogpuMigrationWording(t *testing.T) {
	root := repoRoot(t)
	findings, err := staleGogpuMigrationCodeWordingFindings(root)
	if err != nil {
		t.Fatalf("staleGogpuMigrationCodeWordingFindings error: %v", err)
	}
	if len(findings) == 0 {
		return
	}

	var report strings.Builder
	for _, finding := range findings {
		report.WriteString("\n")
		report.WriteString(finding.String())
	}
	t.Fatalf("default Go sources contain stale gogpu migration wording:%s", report.String())
}

func staleGogpuMigrationCodeWordingFindings(root string) ([]staleGogpuMigrationCodeFinding, error) {
	var findings []staleGogpuMigrationCodeFinding
	for _, scanRoot := range []string{
		filepath.Join(root, "cmd", "fecim-lattice-tools"),
		filepath.Join(root, "internal"),
		filepath.Join(root, "shared", "viewmodel"),
	} {
		rootFindings, err := staleGogpuMigrationCodeWordingFindingsInTree(root, scanRoot)
		if err != nil {
			return nil, err
		}
		findings = append(findings, rootFindings...)
	}
	return findings, nil
}

func staleGogpuMigrationCodeWordingFindingsInTree(root, scanRoot string) ([]staleGogpuMigrationCodeFinding, error) {
	var findings []staleGogpuMigrationCodeFinding
	err := filepath.WalkDir(scanRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if os.IsNotExist(walkErr) {
				return nil
			}
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		fileFindings, err := staleGogpuMigrationCodeFindingsInFile(path, rel)
		if err != nil {
			return err
		}
		findings = append(findings, fileFindings...)
		return nil
	})
	return findings, err
}

func staleGogpuMigrationCodeFindingsInFile(path, rel string) ([]staleGogpuMigrationCodeFinding, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", rel, err)
	}
	defer file.Close()

	var findings []staleGogpuMigrationCodeFinding
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := scanner.Text()
		lowerLine := strings.ToLower(line)
		for _, phrase := range staleGogpuMigrationCodePhrases {
			if strings.Contains(lowerLine, strings.ToLower(phrase)) {
				findings = append(findings, staleGogpuMigrationCodeFinding{
					Path:   rel,
					Line:   lineNumber,
					Phrase: phrase,
					Text:   strings.TrimSpace(line),
				})
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", rel, err)
	}
	return findings, nil
}
