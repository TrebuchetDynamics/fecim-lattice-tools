//go:build legacy_fyne

package reference

import "testing"

func TestParseArraySize_DefaultsInvalidSelection(t *testing.T) {
	if got := ParseArraySize(""); got != 32 {
		t.Fatalf("empty size got %d, want 32", got)
	}
	if got := ParseArraySize("64"); got != 64 {
		t.Fatalf("size got %d, want 64", got)
	}
}

func TestNewSpecSummary_DerivedMetrics(t *testing.T) {
	summary := NewSpecSummary(32, 76)
	if summary.Cells != 1024 {
		t.Fatalf("cells got %d, want 1024", summary.Cells)
	}
	if summary.ThroughputText != "1024 MACs (Ops) / 76ns = 13.5 GOPS" {
		t.Fatalf("throughput text got %q", summary.ThroughputText)
	}
	if summary.EfficiencyGOPSW != 629 {
		t.Fatalf("efficiency got %d, want 629", summary.EfficiencyGOPSW)
	}
	if summary.TotalPowerText != "21.4 mW" || summary.TotalAreaText != "0.09 mm²" || summary.TotalLatencyText != "76 ns" {
		t.Fatalf("unexpected total texts: %#v", summary)
	}
}

func TestNewSpecSummary_DefaultsInvalidInputs(t *testing.T) {
	summary := NewSpecSummary(0, 0)
	if summary.Size != 32 || summary.TotalLatencyNS != 76 {
		t.Fatalf("unexpected defaults: %#v", summary)
	}
}
