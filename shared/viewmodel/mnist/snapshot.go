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
	// Education layer
	sections = append(sections, viewmodel.Section{
		ID: "edu_pipeline", Title: "📖 CIM Inference Pipeline",
		Body: "Image pixels → quantize to voltage levels → apply to crossbar rows → currents sum at columns (MVM) → softmax activation → digit prediction. The crossbar performs the matrix multiplication in O(1) analog time instead of O(n³) digital.",
	})
	// Research layer
	sections = append(sections, viewmodel.Section{
		ID: "research_benchmark", Title: "🔬 Benchmark Reference",
		Body: "80% baseline on MNIST test set (10,000 images). Educational, not validated device claim. Compare against: HZO FTJ reservoir computing (98.24%, J. Alloys Compounds 2025) — note that this is a different architecture, not FeCIM. Crossbar non-idealities reduce accuracy from ideal baseline.",
	})
	// Design layer
	sections = append(sections, viewmodel.Section{
		ID: "design_tradeoff", Title: "⚙️ Accuracy vs. Quantization",
		Body: fmt.Sprintf("Design sweep: vary quantization levels (8–128). More levels = higher accuracy but harder to program. At %d levels, expect ~%.0f%% accuracy. At 64 levels, expect ~85-90%% (projected, not validated). Cross-reference: Module 2 for array sizing vs. accuracy.", state.NumLevels, state.Accuracy*100),
	})
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
