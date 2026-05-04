package mnist

import (
	"fmt"
	"fecim-lattice-tools/shared/viewmodel"
)

func buildSnapshot(state MNISTState) viewmodel.ModuleSnapshot {
	metrics := []viewmodel.Metric{
		{ID: "accuracy", Label: "Accuracy", Value: fmt.Sprintf("%.1f%%", state.Accuracy*100)},
		{ID: "levels", Label: "Quantization", Value: fmt.Sprintf("%d levels", state.NumLevels)},
		{ID: "correct", Label: "Correct", Value: fmt.Sprintf("%d/%d", state.CorrectImages, state.TotalImages)},
	}
	sections := []viewmodel.Section{
		{ID: "pipeline", Title: "Inference Pipeline", Body: fmt.Sprintf("Image → Quantize → %d-level MVM → Softmax → Prediction. Baseline: %.0f%% at %d levels.", state.NumLevels, state.Accuracy*100, state.NumLevels)},
		{ID: "nonideality", Title: "Non-Ideality Impact", Body: "IR drop and conductance drift modeled at array level. Quantization error increases at lower level counts."},
	}
	actions := []viewmodel.Action{
		{ID: "run_inference", Label: "Run Inference", Kind: viewmodel.ActionCommand},
		{ID: "sweep_levels", Label: "Sweep Levels", Kind: viewmodel.ActionCommand},
	}
	return viewmodel.ModuleSnapshot{
		Descriptor: viewmodel.ModuleDescriptor{
			ID: viewmodel.ModuleMNIST, Title: "FeCIM MNIST Neural Network",
			Description: "Educational CIM inference pipeline with quantized weights and reproducible metrics.",
			Status: viewmodel.StatusFunctional,
		},
		Metrics: metrics, Sections: sections, Actions: actions,
	}
}
