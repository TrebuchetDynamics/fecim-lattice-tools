package crossbarcli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestModule2CrossbarCLIE2EWideVisualizationModes(t *testing.T) {
	modes := []struct {
		name    string
		flags   []string
		markers []string
	}{
		{name: "array", flags: []string{"--show-array"}, markers: []string{"Crossbar", "Level Legend", "Neural Network Inference"}},
		{name: "mvm", flags: []string{"--show-mvm"}, markers: []string{"MVM", "Input", "Output"}},
		{name: "irdrop", flags: []string{"--show-irdrop"}, markers: []string{"IR Drop", "Wire resistance", "Drop"}},
		{name: "sneak", flags: []string{"--show-sneak"}, markers: []string{"Sneak", "Path", "Selected"}},
		{name: "nonidealities", flags: []string{"--show-nonidealities"}, markers: []string{"MVM Comparison", "Ideal", "Non-Ideal"}},
	}
	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			args := append([]string{"--size", "4", "--layers", "2", "--noise", "0", "--adc", "6", "--seed", "7", "--no-color"}, mode.flags...)
			out, stderr, err := captureCrossbarCLIE2EOutput(args)
			if err != nil {
				t.Fatalf("runInference(%v) error = %v\nstderr=%s\nstdout=%s", args, err, stderr, out)
			}
			if stderr != "" {
				t.Fatalf("stderr = %q, want empty", stderr)
			}
			for _, marker := range append([]string{"FeCIM Demo 2", "Crossbar size: 4 x 4", "Discrete levels: 30"}, mode.markers...) {
				if !strings.Contains(out, marker) {
					t.Fatalf("mode %s output missing %q\n%s", mode.name, marker, out)
				}
			}
		})
	}
}

func TestModule2CrossbarCLIE2EConfigHelpAndInvalidMatrix(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "crossbar.json")
	if err := os.WriteFile(configPath, []byte(`{"array_size":5,"num_layers":2,"noise_level":0.01,"adc_bits":5}`), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	out, stderrText, err := captureCrossbarCLIE2EOutput([]string{"--config", configPath, "--show-mvm", "--seed", "11", "--no-color"})
	if err != nil {
		t.Fatalf("runInference(config) error = %v\nstderr=%s", err, stderrText)
	}
	var stdout, stderr bytes.Buffer
	for _, marker := range []string{"Crossbar size: 5 x 5", "Layers: 2", "Noise level: 1.00%", "ADC bits: 5", "MVM"} {
		if !strings.Contains(out, marker) {
			t.Fatalf("config output missing %q\n%s", marker, out)
		}
	}

	stdout.Reset()
	stderr.Reset()
	if err := runInference([]string{"--help"}, &stdout, &stderr); err != nil {
		t.Fatalf("runInference(help) error = %v", err)
	}
	if stderr.Len() != 0 || !strings.Contains(stdout.String(), "FeCIM Crossbar Inference") || !strings.Contains(stdout.String(), "show-nonidealities") {
		t.Fatalf("help stdout/stderr unexpected\nstdout=%s\nstderr=%s", stdout.String(), stderr.String())
	}

	invalids := [][]string{
		{"--size", "0", "--show-mvm"},
		{"--size", "4", "--adc", "0", "--show-mvm"},
		{"--config", filepath.Join(t.TempDir(), "missing.json"), "--show-mvm"},
	}
	for _, args := range invalids {
		stdout.Reset()
		stderr.Reset()
		err := runInference(args, &stdout, &stderr)
		if err == nil {
			t.Fatalf("runInference(%v) returned nil error\nstdout=%s", args, stdout.String())
		}
	}
}

func TestModule2CrossbarCLIE2EDeterministicSeedOutput(t *testing.T) {
	run := func() string {
		out, stderr, err := captureCrossbarCLIE2EOutput([]string{"--size", "4", "--show-mvm", "--seed", "123", "--noise", "0", "--adc", "6", "--no-color"})
		if err != nil {
			t.Fatalf("runInference deterministic error = %v\nstderr=%s", err, stderr)
		}
		return out
	}
	first := run()
	second := run()
	if first != second {
		t.Fatalf("deterministic seeded output changed\nFIRST:\n%s\nSECOND:\n%s", first, second)
	}
}

func captureCrossbarCLIE2EOutput(args []string) (string, string, error) {
	originalStdout := os.Stdout
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		return "", "", err
	}
	os.Stdout = writePipe

	var unusedStdout, stderr bytes.Buffer
	runErr := runInference(args, &unusedStdout, &stderr)

	_ = writePipe.Close()
	os.Stdout = originalStdout
	captured, readErr := io.ReadAll(readPipe)
	_ = readPipe.Close()
	if runErr != nil {
		return string(captured), stderr.String(), runErr
	}
	if readErr != nil {
		return string(captured), stderr.String(), readErr
	}
	return string(captured), stderr.String(), nil
}
