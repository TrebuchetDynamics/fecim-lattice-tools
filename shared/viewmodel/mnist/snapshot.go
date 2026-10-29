package mnist

import (
	"fecim-lattice-tools/shared/viewmodel"
	"fmt"
)

func buildSnapshot(state MNISTState) viewmodel.ModuleSnapshot {
	return viewmodel.ModuleSnapshot{
		Descriptor: viewmodel.ModuleDescriptor{ID: viewmodel.ModuleMNIST, Title: "FeCIM MNIST Neural Network", Description: "Local CPU inference with explicit model and dataset identity.", Status: viewmodel.StatusFunctional, BoundaryNotice: "EDUCATIONAL SIMULATION — not a validated device measurement. No fixed accuracy baseline; model/data evaluation is required."},
		Metrics:    []viewmodel.Metric{{ID: "accuracy", Label: "Accuracy", Value: "Unavailable — select a model and labelled dataset in local web mode"}, {ID: "levels", Label: "Quantization", Value: fmt.Sprintf("%d levels", state.NumLevels)}, {ID: "correct", Label: "Correct", Value: "Unavailable — no labelled samples evaluated"}},
		Sections: []viewmodel.Section{
			{ID: "pipeline", Title: "Inference Pipeline", Body: fmt.Sprintf("Image → %d-level quantize → crossbar MVM → softmax. Inference unavailable in this summary; use the local web module with actual model bytes and an input.", state.NumLevels), Category: "research"},
			{ID: "nonideality", Title: "Non-Ideality Impact", Body: "Educational IR drop, sneak paths, DAC/ADC and noise effects are configured explicitly in the local CPU inference path.", Category: "research"},
			{ID: "edu_pipeline", Title: "CIM Inference Pipeline", Body: "MNIST pixels → DAC → crossbar matrix-vector multiply → ADC → activation → digit prediction. These are numerical simulations, not measured analog timing.", Category: "education"},
			{ID: "research_benchmark", Title: "Evaluation Evidence", Body: "Educational MNIST evaluation requires actual labelled samples, model and dataset SHA-256 identities, selected subset and configuration. No 80% baseline or synthetic accuracy is supplied.", Category: "research"},
			{ID: "design_tradeoff", Title: "Accuracy vs. Quantization", Body: fmt.Sprintf("Selected %d levels. Accuracy and sweep curves are unavailable until real labelled evaluation completes in local web mode; no illustrative staircase is reported as evidence.", state.NumLevels), Category: "design"},
		},
		Actions: []viewmodel.Action{{ID: "run_inference", Label: "Inference unavailable — use local web mode", Kind: viewmodel.ActionCommand}, {ID: "sweep_levels", Label: "Set levels (no accuracy evaluation)", Kind: viewmodel.ActionCommand}},
	}
}
