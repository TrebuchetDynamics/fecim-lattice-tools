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
	// Education layer
	sections = append(sections, viewmodel.Section{
		ID: "edu_adc", Title: "📖 How SAR ADC Works",
		Body: fmt.Sprintf("Successive Approximation Register ADC: Binary search over %d levels. Each bit is tested: set bit, compare against input, keep or discard. %d clock cycles to complete. INL/DNL characterize deviation from ideal.", 1<<state.ADCResolution, state.ADCResolution),
	})
	sections = append(sections, viewmodel.Section{
		ID: "edu_ispp", Title: "📖 ISPP Write-Verify",
		Body: "Incremental Step Pulse Programming: Apply voltage pulse → Wait for settling → Verify conductance → If not at target, increase pulse amplitude → Repeat. Guard-band pulses prevent overshoot. Binary search accelerates convergence.",
	})
	// Research layer
	sections = append(sections, viewmodel.Section{
		ID: "research_pvt", Title: "🔬 PVT Variation",
		Body: fmt.Sprintf("Process/Voltage/Temperature corners: TT (typical), FF (fast NMOS/PMOS), SS (slow). ADC INL degrades at SS corner. Charge pump output drops at low Vdd (%.1f V min). All values are educational models.", state.SupplyVoltage*0.9),
	})
	// Design layer
	sections = append(sections, viewmodel.Section{
		ID: "design_readpath", Title: "⚙️ Optimizing the Read Path",
		Body: fmt.Sprintf("Latency budget: TIA settling + %d-cycle ADC conversion. Lower resolution = faster but noisier. Design trade: 5-bit ADC for 30-level cells gives 1.7× noise margin. Cross-reference: Module 2 array output feeds this read path.", state.ADCResolution),
	})
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
