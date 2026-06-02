package circuitscli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCircuitsReportsFlagErrorToStderr(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := runCircuits([]string{"-definitely-not-a-flag"}, &stdout, &stderr)

	if err == nil {
		t.Fatal("runCircuits error = nil, want invalid flag error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout length = %d, want 0; stdout=%q", stdout.Len(), stdout.String())
	}
	text := stderr.String()
	if !strings.Contains(text, "flag provided but not defined: -definitely-not-a-flag") {
		t.Fatalf("stderr = %q, want invalid flag context", text)
	}
	if !strings.Contains(text, "Error:") {
		t.Fatalf("stderr = %q, want error prefix", text)
	}
	if !strings.Contains(text, "FeCIM Peripheral Circuits CLI") {
		t.Fatalf("stderr = %q, want usage", text)
	}
}

func TestBuildCircuitsResultShowAll(t *testing.T) {
	r := buildCircuitsResult(false, false, false, false, true)
	if r.DAC == nil || r.ADC == nil || r.TIA == nil || r.Pump == nil {
		t.Fatal("showAll should populate all peripheral result sections")
	}
	if r.DAC.Levels <= 0 || r.ADC.Levels <= 0 {
		t.Fatal("invalid DAC/ADC levels in result")
	}
}

func TestCheckMonotonicity(t *testing.T) {
	if !checkMonotonicity([]float64{0.1, -0.9, 0.2}) {
		t.Fatal("expected monotonic pass when all DNL > -1")
	}
	if checkMonotonicity([]float64{0.1, -1.2, 0.2}) {
		t.Fatal("expected monotonic fail when a DNL < -1")
	}
}

func TestRunCircuitsE2EWideJSONPeripheralMatrix(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want []string
	}{
		{name: "all", args: []string{"--json", "--all"}, want: []string{"dac", "adc", "tia", "pump"}},
		{name: "write_path", args: []string{"--json", "--dac", "--pump", "--level", "29"}, want: []string{"dac", "pump"}},
		{name: "read_path", args: []string{"--json", "--adc", "--tia", "--level", "0"}, want: []string{"adc", "tia"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), tc.name+".json")
			args := append(append([]string{}, tc.args...), "--output", out)
			var stdout, stderr bytes.Buffer
			if err := runCircuits(args, &stdout, &stderr); err != nil {
				t.Fatalf("runCircuits(%v): %v stderr=%s", args, err, stderr.String())
			}
			if stdout.Len() != 0 || stderr.Len() != 0 {
				t.Fatalf("JSON output-file run should not write provided buffers stdout=%q stderr=%q", stdout.String(), stderr.String())
			}
			raw, err := os.ReadFile(out)
			if err != nil {
				t.Fatalf("read output: %v", err)
			}
			var result map[string]any
			if err := json.Unmarshal(raw, &result); err != nil {
				t.Fatalf("invalid JSON: %v data=%s", err, raw)
			}
			for _, key := range tc.want {
				section, ok := result[key].(map[string]any)
				if !ok || len(section) == 0 {
					t.Fatalf("missing section %q in %s", key, raw)
				}
			}
			if _, ok := result["dac"]; ok && result["dac"].(map[string]any)["levels"].(float64) <= 0 {
				t.Fatalf("invalid DAC levels: %s", raw)
			}
			if _, ok := result["adc"]; ok && result["adc"].(map[string]any)["enob"].(float64) <= 0 {
				t.Fatalf("invalid ADC ENOB: %s", raw)
			}
		})
	}
}

func TestRunCircuitsE2EConfigOutputHelpAndInvalidMatrix(t *testing.T) {
	tmp := t.TempDir()
	config := filepath.Join(tmp, "circuits.json")
	if err := os.WriteFile(config, []byte(`{"level":22,"show_all":true,"verbosity":3}`), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	out := filepath.Join(tmp, "configured.json")
	var stdout, stderr bytes.Buffer
	if err := runCircuits([]string{"--json", "--config", config, "--output", out}, &stdout, &stderr); err != nil {
		t.Fatalf("configured run: %v stderr=%s", err, stderr.String())
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("configured output missing: %v", err)
	}
	if !strings.Contains(string(raw), `"dac"`) || !strings.Contains(string(raw), `"pump"`) {
		t.Fatalf("configured JSON missing show_all sections: %s", raw)
	}

	stdout.Reset()
	stderr.Reset()
	if err := runCircuits([]string{"--help"}, &stdout, &stderr); err != nil {
		t.Fatalf("help run: %v", err)
	}
	if !strings.Contains(stdout.String(), "FeCIM Peripheral Circuits CLI") || !strings.Contains(stdout.String(), "--json") {
		t.Fatalf("help stdout missing usage: %q", stdout.String())
	}

	badConfig := filepath.Join(tmp, "bad.json")
	badOut := filepath.Join(tmp, "bad-out.json")
	if err := os.WriteFile(badConfig, []byte(`{"level":`), 0644); err != nil {
		t.Fatalf("write bad config: %v", err)
	}
	stdout.Reset()
	stderr.Reset()
	if err := runCircuits([]string{"--json", "--config", badConfig, "--output", badOut}, &stdout, &stderr); err == nil {
		t.Fatal("malformed config should fail")
	}
	if _, err := os.Stat(badOut); !os.IsNotExist(err) {
		t.Fatalf("bad config should not create output, stat err=%v", err)
	}

	stdout.Reset()
	stderr.Reset()
	if err := runCircuits([]string{"--json", "--all", "--output", filepath.Join(tmp, "missing", "out.json")}, &stdout, &stderr); err == nil {
		t.Fatal("missing output directory should fail")
	}
}
