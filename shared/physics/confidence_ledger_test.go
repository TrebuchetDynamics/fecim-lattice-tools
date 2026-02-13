package physics

import "testing"

func TestConfidenceLedgerLookupKnownParam(t *testing.T) {
	ledger := NewConfidenceLedger()
	tag, ok := ledger.Lookup("Ec")
	if !ok {
		t.Fatalf("expected Ec in ledger")
	}
	if tag.Provenance != ProvenanceMeasured {
		t.Fatalf("Ec provenance=%s, want %s", tag.Provenance, ProvenanceMeasured)
	}
	if tag.Confidence <= 0 || tag.Confidence > 1 {
		t.Fatalf("Ec confidence=%.3f outside [0,1]", tag.Confidence)
	}
}

func TestConfidenceLedgerUnknownParamDefaultsToPlaceholder(t *testing.T) {
	ledger := NewConfidenceLedger()
	tagged := ledger.TagOutput("not_real", 1.23)
	if tagged.Tag.Provenance != ProvenancePlaceholder {
		t.Fatalf("unknown provenance=%s, want %s", tagged.Tag.Provenance, ProvenancePlaceholder)
	}
}
