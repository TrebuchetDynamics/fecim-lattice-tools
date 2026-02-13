package physics

import "testing"

func TestBuildResearchTracePopulatesAllFields(t *testing.T) {
	trace := BuildResearchTrace(512, 2e-6, 75e3, 10)

	if trace.DAC.OutputVoltage.Unit == "" || trace.DAC.OutputVoltage.Value == 0 {
		t.Fatalf("DAC output voltage not populated: %+v", trace.DAC.OutputVoltage)
	}
	if trace.Array.ArrayOutput.Unit == "" || trace.Array.ArrayOutput.Value == 0 {
		t.Fatalf("array output not populated: %+v", trace.Array.ArrayOutput)
	}
	if trace.TIA.OutputVoltage.Unit == "" || trace.TIA.OutputVoltage.Value == 0 {
		t.Fatalf("TIA output not populated: %+v", trace.TIA.OutputVoltage)
	}
	if trace.ADC.ResolutionBits == 0 || trace.ADC.SampleRate.Unit == "" {
		t.Fatalf("ADC fields not populated: %+v", trace.ADC)
	}
	if trace.Classifier.Probability.Unit == "" || trace.Classifier.ConfidenceInterval.Unit == "" {
		t.Fatalf("classifier fields not populated: %+v", trace.Classifier)
	}
}
