package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBaseMaterial(t *testing.T) {
	if baseMaterial("default_hzo") == nil {
		t.Fatal("default_hzo should resolve")
	}
	if baseMaterial("literature_superlattice") == nil {
		t.Fatal("literature_superlattice should resolve")
	}
	if baseMaterial("unknown") != nil {
		t.Fatal("unknown preset should return nil")
	}
}

func TestLoadCSVLoop(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "loop.csv")
	content := "E_MV_cm,P_uC_cm2\n-1.0,-10.0\n0.0,0.0\n1.0,10.0\n"
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	E, P, err := loadCSVLoop(p)
	if err != nil {
		t.Fatalf("loadCSVLoop error: %v", err)
	}
	if len(E) != 3 || len(P) != 3 {
		t.Fatalf("unexpected lengths E=%d P=%d", len(E), len(P))
	}
	if E[0] != -1.0 || P[2] != 10.0 {
		t.Fatalf("unexpected parsed values E0=%v P2=%v", E[0], P[2])
	}
}

func TestLoadCSVLoopRejectsEmptyInput(t *testing.T) {
	for _, content := range []string{"", "E_MV_cm,P_uC_cm2\n"} {
		t.Run(fmt.Sprintf("bytes_%d", len(content)), func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "empty.csv")
			if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, _, err := loadCSVLoop(p); err == nil {
				t.Fatal("expected empty-loop error")
			}
		})
	}
}

func TestLoadCSVLoopRejectsNonFiniteValues(t *testing.T) {
	for _, value := range []string{"NaN", "+Inf", "-Inf"} {
		t.Run(value, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "non-finite.csv")
			content := "E_MV_cm,P_uC_cm2\n" + value + ",1\n"
			if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, _, err := loadCSVLoop(p); err == nil {
				t.Fatal("expected non-finite-value error")
			}
		})
	}
}

func TestLoadCSVLoopRejectsMalformedRows(t *testing.T) {
	t.Run("missing column", func(t *testing.T) {
		d := t.TempDir()
		p := filepath.Join(d, "loop.csv")
		content := "E_MV_cm,P_uC_cm2\n-1.0\n"
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("write csv: %v", err)
		}
		if _, _, err := loadCSVLoop(p); err == nil || !strings.Contains(err.Error(), "expected at least 2 columns") {
			t.Fatalf("expected missing-column error, got %v", err)
		}
	})

	t.Run("bad float", func(t *testing.T) {
		d := t.TempDir()
		p := filepath.Join(d, "loop.csv")
		content := "E_MV_cm,P_uC_cm2\nabc,1.0\n"
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("write csv: %v", err)
		}
		if _, _, err := loadCSVLoop(p); err == nil || !strings.Contains(err.Error(), "parse E_MV_cm") {
			t.Fatalf("expected parse error, got %v", err)
		}
	})
}

func TestRunCalibratePreisachWritesProvenancedJSONArtifact(t *testing.T) {
	input := filepath.Join("..", "..", "validation", "literature", "data", "park2015_synthetic_shape.csv")
	output := filepath.Join(t.TempDir(), "calibration.json")
	var stdout, stderr bytes.Buffer
	code := runCalibratePreisach([]string{"-csv", input, "-preset", "default_hzo", "-out", output}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	blob, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	var artifact struct {
		SchemaVersion int    `json:"schema_version"`
		Model         string `json:"model"`
		Preset        string `json:"preset"`
		Input         struct {
			File   string `json:"file"`
			SHA256 string `json:"sha256"`
		} `json:"input"`
		Fit struct {
			RMSEFS float64 `json:"rmse_fs"`
		} `json:"fit"`
	}
	if err := json.Unmarshal(blob, &artifact); err != nil {
		t.Fatalf("decode calibration artifact: %v", err)
	}
	inputBytes, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	wantHash := fmt.Sprintf("%x", sha256.Sum256(inputBytes))
	if artifact.SchemaVersion != 1 || artifact.Model != "Preisach" || artifact.Preset != "default_hzo" || artifact.Input.SHA256 != wantHash || artifact.Fit.RMSEFS < 0 {
		t.Fatalf("unexpected calibration artifact: %+v", artifact)
	}
	if !strings.Contains(stdout.String(), "best Ec=") {
		t.Fatalf("stdout=%q lost human-readable summary", stdout.String())
	}
}

func TestCalibrationArtifactHashesTheFitInputSnapshot(t *testing.T) {
	input := filepath.Join(t.TempDir(), "loop.csv")
	snapshot := []byte("E_MV_cm,P_uC_cm2\n-1,-10\n0,0\n1,10\n")
	if err := os.WriteFile(input, snapshot, 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(input, []byte("changed after fitting\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "calibration.json")
	if err := writeCalibrationArtifact(output, input, "default_hzo", calibrationFit{}, loaded); err != nil {
		t.Fatal(err)
	}
	blob, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	var artifact calibrationArtifact
	if err := json.Unmarshal(blob, &artifact); err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("%x", sha256.Sum256(snapshot))
	if artifact.Input.SHA256 != want {
		t.Fatalf("artifact input hash = %s, want snapshot hash %s", artifact.Input.SHA256, want)
	}
}

func TestCalibrationArtifactDoesNotOverwriteExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "calibration.json")
	original := []byte("keep this file")
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}
	input := filepath.Join("..", "..", "validation", "literature", "data", "park2015_synthetic_shape.csv")
	if err := writeCalibrationArtifact(path, input, "default_hzo", calibrationFit{}, []byte("input")); err == nil {
		t.Fatal("writing over an existing artifact unexpectedly succeeded")
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != string(original) {
		t.Fatalf("existing file changed: bytes=%q err=%v", got, err)
	}
}

func TestRunCalibratePreisachUnknownPresetWithoutExiting(t *testing.T) {
	d := t.TempDir()
	csvPath := filepath.Join(d, "loop.csv")
	content := "E_MV_cm,P_uC_cm2\n-1.0,-10.0\n0.0,0.0\n1.0,10.0\n"
	if err := os.WriteFile(csvPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write csv: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCalibratePreisach([]string{"-csv", csvPath, "-preset", "unknown"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("exit code=%d, want 1; stderr=%q", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout=%q, want empty output", stdout.String())
	}
	if !strings.Contains(stderr.String(), "unknown preset \"unknown\"") {
		t.Fatalf("stderr=%q, want unknown preset context", stderr.String())
	}
}
