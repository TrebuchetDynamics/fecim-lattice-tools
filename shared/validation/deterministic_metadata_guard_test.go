package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeterministicMetadata_NoRuntimeTimestampsInRegressionTests(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	roots := []string{
		"validation",
		"module1-hysteresis",
		"module2-crossbar",
		"module3-mnist",
		"module4-circuits/pkg",
		"module5-comparison",
		"module6-eda/pkg",
		"module7-docs",
		"shared/widgets",
	}
	bad := []string{"time.Now().UTC().Format(time.RFC3339)", "time.Now().Unix()"}

	for _, rel := range roots {
		root := filepath.Join(repoRoot, rel)
		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				t.Fatalf("walk %s: %v", root, err)
			}
			if d.IsDir() || !strings.HasSuffix(path, "_test.go") {
				return nil
			}
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			s := string(b)
			for _, needle := range bad {
				if strings.Contains(s, needle) {
					t.Fatalf("runtime timestamp pattern %q found in %s", needle, path)
				}
			}
			return nil
		})
	}
}
