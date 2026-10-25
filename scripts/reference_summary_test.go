package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestReferenceSummaryDoesNotCountSyntheticOrExcludedDataAsExperiments(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 required for reference summary")
	}
	root := repoRoot(t)
	script := readFile(t, filepath.Join(root, "scripts", "run_literature_validation.sh"))
	start := strings.Index(script, "python3 - <<'PY'\n")
	if start < 0 {
		t.Fatal("summary Python block not found")
	}
	code := strings.SplitN(script[start+len("python3 - <<'PY'\n"):], "\nPY", 2)[0]
	dir := t.TempDir()
	for _, id := range []string{"park2015_hzo_10nm", "park2015_synthetic_shape", "bto2021_cryst11101192_hysteresis_digitized"} {
		body := `{"material_id":"` + id + `","pr_err_pct":1,"ec_err_pct":2,"rmse_over_ps":0.01,"loop_area_err_pct":3,"pass":true}`
		if err := os.WriteFile(filepath.Join(dir, "module1_pe_loop_"+id+".json"), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.Command(python, "-c", code)
	cmd.Env = append(os.Environ(), "FECIM_LITERATURE_JSON_DIR="+dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("summary failed: %v: %s", err, out)
	}
	for _, want := range []string{"synthetic-regression", "reference_artifacts=1", "experimental_validations=0", "excluded=1", "superseded=1"} {
		if !strings.Contains(string(out), want) {
			t.Errorf("summary missing %q: %s", want, out)
		}
	}
}
