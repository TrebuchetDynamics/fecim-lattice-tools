package validate

// ProcessCornerEnvelope describes process-voltage-temperature envelopes attached
// to exported artifacts.
type ProcessCornerEnvelope struct {
	Corner       string  `json:"corner"`
	VoltageV     float64 `json:"voltage_v"`
	TemperatureC float64 `json:"temperature_c"`
}

// PDKReadinessInput captures available signoff collateral.
type PDKReadinessInput struct {
	HasNLDM        bool
	HasMultiCorner bool
	HasDRCPass     bool
	HasLVSPass     bool
	Envelopes      []ProcessCornerEnvelope
}

// PDKRealityBridgeOutput annotates exports with reality checks.
type PDKRealityBridgeOutput struct {
	Envelopes      []ProcessCornerEnvelope `json:"envelopes"`
	ReadinessScore int                     `json:"readiness_score"`
}

// BuildPDKRealityBridge returns corner envelopes and a 0-100 readiness score.
// The score is additive with equal weights:
// NLDM (25), multi-corner coverage (25), DRC pass (25), LVS pass (25).
func BuildPDKRealityBridge(in PDKReadinessInput) PDKRealityBridgeOutput {
	score := 0
	if in.HasNLDM {
		score += 25
	}
	if in.HasMultiCorner {
		score += 25
	}
	if in.HasDRCPass {
		score += 25
	}
	if in.HasLVSPass {
		score += 25
	}

	return PDKRealityBridgeOutput{
		Envelopes:      append([]ProcessCornerEnvelope(nil), in.Envelopes...),
		ReadinessScore: score,
	}
}
