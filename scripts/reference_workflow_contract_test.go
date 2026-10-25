package scripts_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReferenceWorkflowFiltersSelectCurrentRegressionTest(t *testing.T) {
	for _, name := range []string{"ci.yml", "nightly.yml"} {
		body, err := os.ReadFile(filepath.Join(repoRoot(t), ".github", "workflows", name))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(body), "TestModule1_PELoop_LiteratureBacked") || !strings.Contains(string(body), "TestModule1_PELoop_ReferenceRegression") {
			t.Errorf("%s contains a stale/no-match reference test filter", name)
		}
	}
}
