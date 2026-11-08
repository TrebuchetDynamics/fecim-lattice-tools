package scripts_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const regressionPEFile = "module1_pe_loop_z_subject.json"

func regressionArtifacts(t *testing.T, module string) (string, map[string]string) {
	t.Helper()
	files := map[string]string{
		regressionPEFile: `{"schema_version":"v1","timestamp_utc":"2026-01-01T00:00:00Z","commit":"fixture","gate":"pr","test_id":"fixture","material":"HZO","dataset":"synthetic","metrics":{"error":0.1},"thresholds":{"error":0.2},"uncertainty":{"method":"opaque-method","confidence":0.95,"sample_size":2},"verdict":"fail","metadata":[null,true,"NaN",{"opaque":[1,2.5]}]}`,
	}
	if module == "module1" {
		// The valid file sorts first: a later invalid artifact must prevent PASS.
		files["module1_pe_loop_a_valid.json"] = files[regressionPEFile]
		files["module1_forc_ignored.json"] = `not JSON`
	} else {
		files = map[string]string{
			"summary.json":            `{"git_hash":"fixture","metadata":[null,true,"Infinity"]}`,
			"matrix.json":             `[{"size":2}]`,
			"thresholds.json":         `[0.2]`,
			"results.json":            `{"verdict":"fail","metrics":{"error":0.1}}`,
			"material_snapshot.json":  `{"material":"HZO"}`,
			"signal_chain_trace.json": `[0,1.5]`,
			"solver_diagnostics.json": `{"kcl_residual_max":0.01}`,
			"thermodynamics.json":     `{"power":0.1}`,
			"confidence_ledger.json":  `{"measured":[],"estimated":[],"placeholder":[]}`,
			"uncertainty.json":        `{"method":"opaque-method","confidence":0.95,"sample_size":2}`,
		}
	}
	return t.TempDir(), files
}

func runRegressionArtifacts(t *testing.T, module, dir string, files map[string]string, filename, field string) {
	t.Helper()
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.Command("python3", "scripts/ci/validate_regression_artifacts.py", "--module", module, "--root", dir)
	if module == "module4" {
		cmd = exec.Command("bash", "scripts/ci/validate-regression-artifacts.sh", module, dir)
	}
	cmd.Dir = repoRoot(t)
	out, err := cmd.CombinedOutput()
	output := string(out)
	if filename == "" {
		if err != nil || !strings.Contains(output, "[artifact-validate] PASS: "+module+" artifacts valid at "+dir) {
			t.Fatalf("valid artifacts: %v\n%s", err, output)
		}
		if module == "module1" && !strings.Contains(output, "module1 checked 2 artifact files") {
			t.Fatalf("PE-loop selection/count changed: %s", output)
		}
		return
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 {
		t.Errorf("invalid artifacts: exit = %v, want 1\n%s", err, output)
	}
	for _, want := range []string{"[artifact-validate] FAIL:", filename, field} {
		if !strings.Contains(output, want) {
			t.Errorf("missing diagnostic context %q: %s", want, output)
		}
	}
	if strings.Contains(output, "PASS") || strings.Contains(output, "Traceback") {
		t.Errorf("invalid artifacts emitted PASS or traceback: %s", output)
	}
}

func TestRegressionArtifactsNonfinite(t *testing.T) {
	for _, module := range []string{"module1", "module4"} {
		for _, token := range []string{"NaN", "Infinity", "-Infinity", "1e999"} {
			for _, field := range []string{"metrics", "thresholds", "uncertainty", "trace"} {
				t.Run(module+"/"+token+"/"+field, func(t *testing.T) {
					dir, files := regressionArtifacts(t, module)
					filename := regressionPEFile
					if module == "module4" {
						filename = map[string]string{"metrics": "results.json", "thresholds": "thresholds.json", "uncertainty": "uncertainty.json", "trace": "signal_chain_trace.json"}[field]
					}
					// Put the number behind both object and array traversal.
					value := `{"nested":[0,{"value":` + token + `}]}`
					if module == "module1" {
						if field == "trace" {
							files[filename] = strings.TrimSuffix(files[filename], "}") + `,"trace":` + value + `}`
						} else {
							files[filename] = strings.Replace(files[filename], `"`+field+`":{`, `"`+field+`":{"extra":`+value+`,`, 1)
						}
					} else if field == "thresholds" || field == "trace" {
						files[filename] = `[` + value + `]`
					} else {
						files[filename] = strings.TrimSuffix(files[filename], "}") + `,"extra":` + value + `}`
					}
					runRegressionArtifacts(t, module, dir, files, filename, "nested[1].value")
				})
			}
		}
	}
}

func TestRegressionArtifactsSampleSize(t *testing.T) {
	for _, module := range []string{"module1", "module4"} {
		for _, value := range []string{"true", "false", "1.5", "2.0", `"2"`, `"bad"`, "null", "[]", "{}", "0", "-1"} {
			t.Run(module+"/"+value, func(t *testing.T) {
				dir, files := regressionArtifacts(t, module)
				filename := regressionPEFile
				if module == "module4" {
					filename = "uncertainty.json"
				}
				files[filename] = strings.Replace(files[filename], `"sample_size":2`, `"sample_size":`+value, 1)
				runRegressionArtifacts(t, module, dir, files, filename, "sample_size")
			})
		}
	}
}

func TestRegressionArtifactsRequiredObjects(t *testing.T) {
	for _, module := range []string{"module1", "module4"} {
		fields := []string{regressionPEFile, "uncertainty"}
		if module == "module4" {
			fields = []string{"summary.json", "results.json", "material_snapshot.json", "solver_diagnostics.json", "thermodynamics.json", "confidence_ledger.json", "uncertainty.json"}
		}
		for _, field := range fields {
			// Key-bearing strings/arrays must not masquerade as objects.
			for _, value := range []string{"null", "true", "2", `["git_hash","verdict","material","kcl_residual_max","power","measured","estimated","placeholder","method","confidence","sample_size"]`, `"git_hash verdict material kcl_residual_max power measured estimated placeholder method confidence sample_size"`} {
				t.Run(module+"/"+field+"/"+value, func(t *testing.T) {
					dir, files := regressionArtifacts(t, module)
					filename := field
					if field == "uncertainty" {
						filename = regressionPEFile
						files[filename] = strings.Replace(files[filename], `{"method":"opaque-method","confidence":0.95,"sample_size":2}`, value, 1)
					} else {
						files[filename] = value
					}
					runRegressionArtifacts(t, module, dir, files, filename, field)
				})
			}
		}
	}
}

func TestRegressionArtifactsCompatibility(t *testing.T) {
	for _, module := range []string{"module1", "module4"} {
		for _, scenario := range []string{"valid", "missing-key", "invalid-json", "missing-file"} {
			t.Run(module+"/"+scenario, func(t *testing.T) {
				dir, files := regressionArtifacts(t, module)
				filename, field := regressionPEFile, "uncertainty"
				if module == "module4" {
					filename, field = "uncertainty.json", "sample_size"
				}
				switch scenario {
				case "valid":
					filename, field = "", ""
				case "missing-key":
					files[filename] = strings.Replace(files[filename], `"sample_size":2`, `"other":2`, 1)
				case "invalid-json":
					files[filename], field = `{"unfinished":`, "invalid json"
				case "missing-file":
					if module == "module1" {
						delete(files, "module1_pe_loop_a_valid.json")
						field = "no module1 PE-loop artifacts"
						filename = dir
						delete(files, regressionPEFile)
					} else {
						delete(files, filename)
						field = "missing file"
					}
				}
				runRegressionArtifacts(t, module, dir, files, filename, field)
			})
		}
	}
}
