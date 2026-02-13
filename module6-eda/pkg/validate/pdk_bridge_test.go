package validate

import "testing"

func TestBuildPDKRealityBridge_ScoreAllPresent(t *testing.T) {
	out := BuildPDKRealityBridge(PDKReadinessInput{
		HasNLDM:        true,
		HasMultiCorner: true,
		HasDRCPass:     true,
		HasLVSPass:     true,
		Envelopes:      []ProcessCornerEnvelope{{Corner: "tt", VoltageV: 1.8, TemperatureC: 25}},
	})
	if out.ReadinessScore != 100 {
		t.Fatalf("expected score 100, got %d", out.ReadinessScore)
	}
	if len(out.Envelopes) != 1 || out.Envelopes[0].Corner != "tt" {
		t.Fatalf("unexpected envelopes: %+v", out.Envelopes)
	}
}

func TestBuildPDKRealityBridge_ScoreMissingItems(t *testing.T) {
	out := BuildPDKRealityBridge(PDKReadinessInput{
		HasNLDM:        true,
		HasMultiCorner: false,
		HasDRCPass:     false,
		HasLVSPass:     true,
	})
	if out.ReadinessScore != 50 {
		t.Fatalf("expected score 50, got %d", out.ReadinessScore)
	}
}
