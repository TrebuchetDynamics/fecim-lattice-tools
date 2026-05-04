package circuits

import (
	"fmt"
	"fecim-lattice-tools/shared/viewmodel"
)

func buildSnapshot(state CircuitsState) viewmodel.ModuleSnapshot {
	metrics := []viewmodel.Metric{
		{ID: "adc", Label: "ADC", Value: fmt.Sprintf("%d-bit SAR", state.ADCResolution)},
		{ID: "dac", Label: "DAC", Value: fmt.Sprintf("%d-bit R-2R", state.DACResolution)},
		{ID: "tia", Label: "TIA", Value: fmt.Sprintf("%.0f kΩ", state.TIAGain/1e3)},
		{ID: "charge_pump", Label: "Charge Pump", Value: fmt.Sprintf("%d-stage Dickson", state.ChargePumpStages)},
		{ID: "ispp", Label: "ISPP", Value: fmt.Sprintf("%v", state.ISPPEnabled)},
		{ID: "supply", Label: "Vdd", Value: fmt.Sprintf("%.1f V", state.SupplyVoltage)},
	}
	sections := []viewmodel.Section{
		{ID: "read_path", Title: "Read Path", Body: fmt.Sprintf("TIA (%.0f kΩ) → %d-bit SAR ADC. Latency: ~%.1f µs.", state.TIAGain/1e3, state.ADCResolution, float64(state.ADCResolution)*0.5)},
		{ID: "write_path", Title: "Write Path (ISPP)", Body: fmt.Sprintf("%d-stage charge pump → %d-bit DAC → ISPP pulse train.", state.ChargePumpStages, state.DACResolution)},
	}
	actions := []viewmodel.Action{
		{ID: "run_read", Label: "Simulate Read", Kind: viewmodel.ActionCommand},
		{ID: "run_write", Label: "Simulate Write", Kind: viewmodel.ActionCommand},
	}
	return viewmodel.ModuleSnapshot{
		Descriptor: viewmodel.ModuleDescriptor{
			ID: viewmodel.ModuleCircuits, Title: "FeCIM Peripheral Circuits Visualizer",
			Description: "DAC, ADC, TIA, read path, write path, and ISPP circuit behavior.",
			Status: viewmodel.StatusFunctional,
		},
		Metrics: metrics, Sections: sections, Actions: actions,
	}
}
