package comparisoncli

import (
	"bytes"
	"strings"
	"testing"

	"fecim-lattice-tools/module5-comparison/pkg/comparison"
)

func TestBuildComparisonResult_PopulatesArchitectures(t *testing.T) {
	w := comparison.MNISTWorkload()
	comp := comparison.CompareArchitectures(w, 1, 10000)
	adv := comparison.CalculateAdvantages(comp)

	res := buildComparisonResult(comp, adv, "mnist", 10000)
	if res.Workload != "mnist" {
		t.Fatalf("workload=%q want mnist", res.Workload)
	}
	if len(res.Architectures) == 0 {
		t.Fatal("expected architectures in JSON result")
	}
	for _, a := range res.Architectures {
		if a.Name == "" {
			t.Fatal("architecture name should not be empty")
		}
	}
}

func TestRunComparisonReportsFlagErrorToStderr(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := runComparison([]string{"-definitely-not-a-flag"}, &stdout, &stderr)

	if err == nil {
		t.Fatal("runComparison error = nil, want invalid flag error")
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
	if !strings.Contains(text, "FeCIM Architecture Comparison CLI") {
		t.Fatalf("stderr = %q, want usage", text)
	}
}
