package validation

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

const gogpuParityLedgerPath = "docs/3-develop/gui/gogpu-parity-ledger.json"

type gogpuParityLedger struct {
	Version int                `json:"version"`
	Entries []gogpuParityEntry `json:"entries"`
}

type gogpuParityEntry struct {
	Module        string   `json:"module"`
	Feature       string   `json:"feature"`
	Status        string   `json:"status"`
	LegacySources []string `json:"legacy_sources"`
	GogpuSurfaces []string `json:"gogpu_surfaces"`
	Evidence      []string `json:"evidence"`
	Notes         string   `json:"notes"`
}

var validGogpuParityStatuses = map[string]struct{}{
	"ported":   {},
	"better":   {},
	"deferred": {},
	"gap":      {},
}

var legacyFyneFeatureRoots = []string{
	"cmd/demo-frames-fyne",
	"cmd/fecim-lattice-tools-fyne",
	"cmd/fecim-screenshotter-fyne",
	"module1-hysteresis/pkg/gui",
	"module2-crossbar/pkg/gui",
	"module3-mnist/pkg/gui",
	"module4-circuits/pkg/gui",
	"module5-comparison/pkg/gui",
	"module6-eda/pkg/gui",
	"module7-docs/pkg/gui",
	"shared/export",
	"shared/keyboard",
	"shared/theme",
	"shared/themes",
	"shared/widgets",
}

func TestGogpuParityLedgerClassifiesLegacyFyneFeatureSources(t *testing.T) {
	root := repoRoot(t)
	ledger := readGogpuParityLedger(t, root)
	sources := collectLegacyFyneFeatureSources(t, root)
	if len(sources) == 0 {
		t.Fatal("expected legacy Fyne feature sources to classify")
	}

	var unclassified []string
	for _, source := range sources {
		if !ledger.classifies(source) {
			unclassified = append(unclassified, source)
		}
	}
	if len(unclassified) > 0 {
		t.Fatalf("gogpu parity ledger has unclassified legacy Fyne feature sources:\n%s", strings.Join(unclassified, "\n"))
	}
}

func TestGogpuParityLedgerEntriesAreActionable(t *testing.T) {
	root := repoRoot(t)
	ledger := readGogpuParityLedger(t, root)
	if ledger.Version != 1 {
		t.Fatalf("ledger version = %d, want 1", ledger.Version)
	}
	if len(ledger.Entries) == 0 {
		t.Fatal("ledger has no entries")
	}

	seenModules := map[string]bool{}
	for i, entry := range ledger.Entries {
		if entry.Module == "" {
			t.Fatalf("entry[%d] missing module", i)
		}
		seenModules[entry.Module] = true
		if entry.Feature == "" {
			t.Fatalf("entry[%d] missing feature", i)
		}
		if _, ok := validGogpuParityStatuses[entry.Status]; !ok {
			t.Fatalf("entry[%d] status = %q, want one of ported|better|deferred|gap", i, entry.Status)
		}
		if len(entry.LegacySources) == 0 {
			t.Fatalf("entry[%d] %s/%s has no legacy_sources", i, entry.Module, entry.Feature)
		}
		if entry.Status == "ported" || entry.Status == "better" {
			if len(entry.GogpuSurfaces) == 0 {
				t.Fatalf("entry[%d] %s/%s status %s requires gogpu_surfaces", i, entry.Module, entry.Feature, entry.Status)
			}
			if len(entry.Evidence) == 0 {
				t.Fatalf("entry[%d] %s/%s status %s requires evidence", i, entry.Module, entry.Feature, entry.Status)
			}
		}
		if (entry.Status == "gap" || entry.Status == "deferred") && entry.Notes == "" {
			t.Fatalf("entry[%d] %s/%s status %s requires notes", i, entry.Module, entry.Feature, entry.Status)
		}
	}

	for _, module := range []string{"hysteresis", "crossbar", "mnist", "circuits", "comparison", "eda", "docs", "shared", "shell"} {
		if !seenModules[module] {
			t.Fatalf("ledger missing module %q", module)
		}
	}
}

func readGogpuParityLedger(t *testing.T, root string) gogpuParityLedger {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(root, gogpuParityLedgerPath))
	if err != nil {
		t.Fatalf("read %s: %v", gogpuParityLedgerPath, err)
	}
	var ledger gogpuParityLedger
	if err := json.Unmarshal(body, &ledger); err != nil {
		t.Fatalf("decode %s: %v", gogpuParityLedgerPath, err)
	}
	return ledger
}

func (l gogpuParityLedger) classifies(source string) bool {
	for _, entry := range l.Entries {
		for _, pattern := range entry.LegacySources {
			if pathPatternMatches(pattern, source) {
				return true
			}
		}
	}
	return false
}

func pathPatternMatches(pattern, source string) bool {
	pattern = filepath.ToSlash(pattern)
	source = filepath.ToSlash(source)
	if strings.HasSuffix(pattern, "/...") {
		prefix := strings.TrimSuffix(pattern, "/...")
		return source == prefix || strings.HasPrefix(source, prefix+"/")
	}
	matched, err := path.Match(pattern, source)
	return err == nil && matched
}

func collectLegacyFyneFeatureSources(t *testing.T, root string) []string {
	t.Helper()
	var sources []string
	for _, relRoot := range legacyFyneFeatureRoots {
		scanRoot := filepath.Join(root, relRoot)
		err := filepath.WalkDir(scanRoot, func(filePath string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				if os.IsNotExist(walkErr) {
					return nil
				}
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			if filepath.Ext(filePath) != ".go" || strings.HasSuffix(filePath, "_test.go") {
				return nil
			}
			rel, err := filepath.Rel(root, filePath)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			source, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("read %s: %w", rel, err)
			}
			if isLegacyFyneFeatureSource(rel, string(source)) {
				sources = append(sources, rel)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", relRoot, err)
		}
	}
	return sources
}

func isLegacyFyneFeatureSource(rel, source string) bool {
	return strings.Contains(rel, "/pkg/gui/") ||
		strings.HasPrefix(rel, "cmd/demo-frames-fyne/") ||
		strings.HasPrefix(rel, "cmd/fecim-lattice-tools-fyne/") ||
		strings.HasPrefix(rel, "cmd/fecim-screenshotter-fyne/") ||
		strings.Contains(source, "legacy_fyne") ||
		strings.Contains(source, "fyne.io/fyne")
}
