//go:build ignore

// Build-time presentation catalogue. No scientific defaults live in TypeScript.
package main

import (
	"encoding/json"
	h "fecim-lattice-tools/workbench/hysteresis"
	"os"
)

func main() {
	c := h.DefaultConfig()
	json.NewEncoder(os.Stdout).Encode(map[string]any{
		"Identity":     h.BridgeIdentity(),
		"Program":      h.DefaultProgramConfig(),
		"ProgramModes": []h.ProgramMode{h.TargetWrite, h.WriteRead, h.TimeResolved},
		"Methods":      []h.ISPPMethod{h.LinearISPP, h.LogarithmicISPP},
		"Diagnostics":  []h.AnalysisKind{h.AnalysisPUND, h.AnalysisFORC},
		"Tables":       map[h.AnalysisKind][]h.AnalysisTable{h.AnalysisPUND: {h.PUNDTraces}, h.AnalysisFORC: {h.FORCSweep, h.FORCDensity}},
		"FORC":         h.DefaultFORCParameters(),
		"Views":        []h.View{h.FullHistory, h.LastCycle},
		// Explicit UI work schedule, not new physics defaults. Same cold-loop
		// schedule as the production parity gate; users may set physical DT.
		"LoopBudget":        map[string]any{"Steps": 64, "DT": 1 / (c.Frequency * 64)},
		"ProgramBudget":     map[string]any{"Steps": h.MaxBridgeSteps, "DT": 0.1},
		"AnalysisBudget":    map[string]any{"Steps": 1},
		"PageSize":          h.PageSize,
		"ReferenceBytes":    h.MaxReferenceBytes,
		"Target":            1,
		"PlaybackCadenceMS": 16,
		"CoreTimeLabel":     "Time (s)",
		"AnalysisColumns": map[h.AnalysisTable][]map[string]string{
			h.PUNDTraces:  {{"Key": "Index", "Label": "Index"}, {"Key": "Pulse", "Label": "Pulse"}, {"Key": "TimeValue", "Label": "TimeS (s)"}, {"Key": "CurrentValue", "Label": "CurrentA (A)"}},
			h.FORCSweep:   {{"Key": "Index", "Label": "Index"}, {"Key": "Branch", "Label": "Branch"}, {"Key": "Reversal", "Label": "Reversal field"}, {"Key": "E", "Label": "E"}, {"Key": "P", "Label": "P"}},
			h.FORCDensity: {{"Key": "Index", "Label": "Index"}, {"Key": "Ea", "Label": "Ea"}, {"Key": "Eb", "Label": "Eb"}, {"Key": "DensityValue", "Label": "Density (C/V²)"}},
		},
	})
}
