package mnistcli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseLevelList covers the comma-separated level parser used by --core-levels
// and --export-levels flags.
func TestParseLevelList(t *testing.T) {
	cases := []struct {
		in      string
		want    []int
		wantErr bool
	}{
		{"8,16,24,31", []int{8, 16, 24, 31}, false},
		{"4", []int{4}, false},
		{"16,8", []int{8, 16}, false},   // must be sorted
		{"8,8,16", []int{8, 16}, false}, // deduplication
		{"", nil, false},
		{" ", nil, false},
		{"bad", nil, true},
		{"8,bad,16", nil, true},
	}
	for _, tc := range cases {
		got, err := parseLevelList(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseLevelList(%q): expected error, got nil", tc.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseLevelList(%q): unexpected error: %v", tc.in, err)
			continue
		}
		if len(got) != len(tc.want) {
			t.Errorf("parseLevelList(%q) = %v, want %v", tc.in, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("parseLevelList(%q)[%d] = %d, want %d", tc.in, i, got[i], tc.want[i])
			}
		}
	}
}

// TestParseDirList covers the comma-separated directory parser.
func TestParseDirList(t *testing.T) {
	cases := []struct {
		in      string
		wantLen int
		wantErr bool
	}{
		{"out1,out2", 2, false},
		{"single", 1, false},
		{"", 0, false},
		{" ", 0, false},
	}
	for _, tc := range cases {
		got, err := parseDirList(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseDirList(%q): expected error", tc.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseDirList(%q): unexpected error: %v", tc.in, err)
			continue
		}
		if len(got) != tc.wantLen {
			t.Errorf("parseDirList(%q): got %d dirs, want %d", tc.in, len(got), tc.wantLen)
		}
	}
}

// TestMNISTConfigDefaults checks that a zero-value MNISTConfig marshals cleanly.
func TestMNISTConfigDefaults(t *testing.T) {
	cfg := MNISTConfig{HiddenSize: 128, NoiseLevel: 0.02, Epochs: 5, Levels: []int{8, 16, 30}}
	b, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("json.Marshal(MNISTConfig): %v", err)
	}
	var rt MNISTConfig
	if err := json.Unmarshal(b, &rt); err != nil {
		t.Fatalf("json.Unmarshal(MNISTConfig): %v", err)
	}
	if rt.HiddenSize != cfg.HiddenSize {
		t.Errorf("HiddenSize round-trip: got %d, want %d", rt.HiddenSize, cfg.HiddenSize)
	}
	if rt.Epochs != cfg.Epochs {
		t.Errorf("Epochs round-trip: got %d, want %d", rt.Epochs, cfg.Epochs)
	}
}

// TestEvaluationResultJSONRoundtrip verifies EvaluationResult serialises and
// deserialises correctly (used for --json CLI output).
func TestEvaluationResultJSONRoundtrip(t *testing.T) {
	orig := EvaluationResult{
		Samples:     1000,
		Accuracy:    0.876,
		FPAccuracy:  0.912,
		CIMAccuracy: 0.876,
		AgreeRate:   0.941,
		AvgKL:       0.023,
		AvgEnergy:   1.234,
		Levels:      30,
	}
	b, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var rt EvaluationResult
	if err := json.Unmarshal(b, &rt); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if rt.Accuracy != orig.Accuracy {
		t.Errorf("Accuracy: got %f, want %f", rt.Accuracy, orig.Accuracy)
	}
	if rt.Samples != orig.Samples {
		t.Errorf("Samples: got %d, want %d", rt.Samples, orig.Samples)
	}
}

// TestResolveWeightsPathExplicit ensures an explicit path is returned unchanged.
func TestResolveWeightsPathExplicit(t *testing.T) {
	got, err := resolveWeightsPath("/tmp/weights.json")
	if err != nil {
		t.Fatalf("resolveWeightsPath(explicit): %v", err)
	}
	if got != "/tmp/weights.json" {
		t.Errorf("resolveWeightsPath: got %q, want %q", got, "/tmp/weights.json")
	}
}

func TestRunMNISTCLIE2EExportQuantizedWeightsMatrix(t *testing.T) {
	tmp := t.TempDir()
	outA := filepath.Join(tmp, "a")
	outB := filepath.Join(tmp, "b")
	var stdout, stderr bytes.Buffer
	weights16 := filepath.Join("..", "..", "data", "pretrained_weights_16.json")
	global := captureMNISTStdoutE2E(t, func() {
		if err := runMNISTCLI([]string{"--export-levels", "4,8", "--export-dir", outA + "," + outB, "--hidden", "16", "--load", weights16}, &stdout, &stderr); err != nil {
			t.Fatalf("export quantized weights: %v stderr=%s", err, stderr.String())
		}
	})
	if !strings.Contains(global, "Export Quantized Weights") || !strings.Contains(global, "Export complete") {
		t.Fatalf("export stdout missing markers: %q", global)
	}
	for _, dir := range []string{outA, outB} {
		for _, name := range []string{"pretrained_weights.json", "pretrained_weights_4.json", "pretrained_weights_8.json"} {
			raw, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatalf("missing export %s/%s: %v", dir, name, err)
			}
			var payload map[string]any
			if err := json.Unmarshal(raw, &payload); err != nil {
				t.Fatalf("invalid JSON export %s/%s: %v", dir, name, err)
			}
			if strings.Contains(name, "_4") && payload["quant_levels"].(float64) != 4 {
				t.Fatalf("quant_levels for %s = %v", name, payload["quant_levels"])
			}
		}
	}
}

func TestRunMNISTCLIE2EHelpConfigAndInvalidExportMatrix(t *testing.T) {
	tmp := t.TempDir()
	cfg := filepath.Join(tmp, "mnist.json")
	if err := os.WriteFile(cfg, []byte(`{"hidden_size":16,"noise_level":0.01,"epochs":1}`), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	out := filepath.Join(tmp, "help.txt")
	var stdout, stderr bytes.Buffer
	if err := runMNISTCLI([]string{"--help", "--output", out}, &stdout, &stderr); err != nil {
		t.Fatalf("help: %v", err)
	}
	if !strings.Contains(stdout.String(), "FeCIM MNIST CLI") || !strings.Contains(stdout.String(), "--json") {
		t.Fatalf("help stdout missing usage: %q", stdout.String())
	}
	badOut := filepath.Join(tmp, "bad")
	weights16 := filepath.Join("..", "..", "data", "pretrained_weights_16.json")
	if err := runMNISTCLI([]string{"--config", cfg, "--export-levels", "1", "--export-dir", badOut, "--hidden", "16", "--load", weights16}, &stdout, &stderr); err == nil {
		t.Fatal("export level <2 should fail")
	}
	if _, err := os.Stat(filepath.Join(badOut, "pretrained_weights_1.json")); !os.IsNotExist(err) {
		t.Fatalf("invalid export should not write level file, stat=%v", err)
	}
	badConfig := filepath.Join(tmp, "bad.json")
	if err := os.WriteFile(badConfig, []byte(`{"hidden_size":`), 0644); err != nil {
		t.Fatalf("write bad config: %v", err)
	}
	if err := runMNISTCLI([]string{"--config", badConfig, "--export-levels", "4", "--export-dir", filepath.Join(tmp, "badcfg")}, &stdout, &stderr); err == nil {
		t.Fatal("malformed config should fail")
	}
}

func captureMNISTStdoutE2E(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = old
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("read pipe: %v", err)
	}
	_ = r.Close()
	return buf.String()
}

func TestRunMNISTCLIReportsFlagErrorToStderr(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("os.Chdir(temp): %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	err = runMNISTCLI([]string{"-definitely-not-a-flag"}, &stdout, &stderr)

	if err == nil {
		t.Fatal("runMNISTCLI error = nil, want invalid flag error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout length = %d, want 0; stdout=%q", stdout.Len(), stdout.String())
	}
	text := stderr.String()
	if !strings.Contains(text, "flag provided but not defined: -definitely-not-a-flag") {
		t.Fatalf("stderr = %q, want invalid flag context", text)
	}
	if !strings.Contains(text, "Error:") {
		t.Fatalf("stderr = %q, want top-level error prefix", text)
	}
	if !strings.Contains(text, "FeCIM MNIST CLI") {
		t.Fatalf("stderr = %q, want CLI usage", text)
	}
	if _, err := os.Stat(filepath.Join(tmp, "logs")); !os.IsNotExist(err) {
		t.Fatalf("invalid flag created logs directory; stat error = %v", err)
	}
}
