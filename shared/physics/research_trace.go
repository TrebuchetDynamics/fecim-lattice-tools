package physics

// TraceValue carries a value with unit and 1-sigma uncertainty.
type TraceValue struct {
	Value       float64 `json:"value"`
	Unit        string  `json:"unit"`
	Uncertainty float64 `json:"uncertainty"`
}

// DACTrace captures input conversion stage.
type DACTrace struct {
	InputCode         int        `json:"input_code"`
	ReferenceVoltage  TraceValue `json:"reference_voltage"`
	OutputVoltage     TraceValue `json:"output_voltage"`
	SettlingTime      TraceValue `json:"settling_time"`
	QuantizationError TraceValue `json:"quantization_error"`
}

// ArrayTrace captures array compute stage.
type ArrayTrace struct {
	WordlineVoltage TraceValue `json:"wordline_voltage"`
	BitlineCurrent  TraceValue `json:"bitline_current"`
	CellConductance TraceValue `json:"cell_conductance"`
	IRDrop          TraceValue `json:"ir_drop"`
	ArrayOutput     TraceValue `json:"array_output"`
}

// TIATrace captures transimpedance conversion stage.
type TIATrace struct {
	InputCurrent  TraceValue `json:"input_current"`
	FeedbackOhms  TraceValue `json:"feedback_ohms"`
	OutputVoltage TraceValue `json:"output_voltage"`
	InputReferred TraceValue `json:"input_referred_noise"`
}

// ADCTrace captures digitization stage.
type ADCTrace struct {
	InputVoltage      TraceValue `json:"input_voltage"`
	ResolutionBits    int        `json:"resolution_bits"`
	SampleRate        TraceValue `json:"sample_rate"`
	OutputCode        int        `json:"output_code"`
	QuantizationNoise TraceValue `json:"quantization_noise"`
}

// ClassifierTrace captures final classification stage.
type ClassifierTrace struct {
	Logit              TraceValue `json:"logit"`
	Probability        TraceValue `json:"probability"`
	PredictedClass     int        `json:"predicted_class"`
	ConfidenceInterval TraceValue `json:"confidence_interval"`
}

// ResearchTrace captures full inference path DAC→array→TIA→ADC→classifier.
type ResearchTrace struct {
	DAC        DACTrace        `json:"dac"`
	Array      ArrayTrace      `json:"array"`
	TIA        TIATrace        `json:"tia"`
	ADC        ADCTrace        `json:"adc"`
	Classifier ClassifierTrace `json:"classifier"`
}

// BuildResearchTrace creates a consistent, unit-tagged sample path trace.
func BuildResearchTrace(inputCode int, gCellSiemens, tiaFeedbackOhms float64, adcBits int) ResearchTrace {
	if tiaFeedbackOhms <= 0 {
		tiaFeedbackOhms = 50e3
	}
	if adcBits <= 0 {
		adcBits = 10
	}
	vref := 1.0
	maxCode := float64((int64(1) << uint(adcBits)) - 1)
	if inputCode < 0 {
		inputCode = 0
	}
	if float64(inputCode) > maxCode {
		inputCode = int(maxCode)
	}

	dacOut := vref * float64(inputCode) / maxCode
	blCurrent := dacOut * gCellSiemens
	irDrop := 0.02 * dacOut
	arrayOutCurrent := blCurrent * 0.98
	tiaOut := arrayOutCurrent * tiaFeedbackOhms
	adcCode := int((tiaOut / vref) * maxCode)
	if adcCode < 0 {
		adcCode = 0
	}
	if float64(adcCode) > maxCode {
		adcCode = int(maxCode)
	}
	prob := float64(adcCode) / maxCode
	pred := 0
	if prob >= 0.5 {
		pred = 1
	}

	return ResearchTrace{
		DAC: DACTrace{
			InputCode:         inputCode,
			ReferenceVoltage:  TraceValue{Value: vref, Unit: "V", Uncertainty: 1e-3},
			OutputVoltage:     TraceValue{Value: dacOut, Unit: "V", Uncertainty: 2e-3},
			SettlingTime:      TraceValue{Value: 20e-9, Unit: "s", Uncertainty: 2e-9},
			QuantizationError: TraceValue{Value: 1.0 / maxCode, Unit: "LSB", Uncertainty: 0},
		},
		Array: ArrayTrace{
			WordlineVoltage: TraceValue{Value: dacOut, Unit: "V", Uncertainty: 2e-3},
			BitlineCurrent:  TraceValue{Value: blCurrent, Unit: "A", Uncertainty: 0.03 * blCurrent},
			CellConductance: TraceValue{Value: gCellSiemens, Unit: "S", Uncertainty: 0.05 * gCellSiemens},
			IRDrop:          TraceValue{Value: irDrop, Unit: "V", Uncertainty: 0.25 * irDrop},
			ArrayOutput:     TraceValue{Value: arrayOutCurrent, Unit: "A", Uncertainty: 0.03 * arrayOutCurrent},
		},
		TIA: TIATrace{
			InputCurrent:  TraceValue{Value: arrayOutCurrent, Unit: "A", Uncertainty: 0.03 * arrayOutCurrent},
			FeedbackOhms:  TraceValue{Value: tiaFeedbackOhms, Unit: "ohm", Uncertainty: 0.01 * tiaFeedbackOhms},
			OutputVoltage: TraceValue{Value: tiaOut, Unit: "V", Uncertainty: 0.04 * tiaOut},
			InputReferred: TraceValue{Value: 20e-9, Unit: "A/sqrt(Hz)", Uncertainty: 5e-9},
		},
		ADC: ADCTrace{
			InputVoltage:      TraceValue{Value: tiaOut, Unit: "V", Uncertainty: 0.04 * tiaOut},
			ResolutionBits:    adcBits,
			SampleRate:        TraceValue{Value: 10e6, Unit: "Sa/s", Uncertainty: 0.1e6},
			OutputCode:        adcCode,
			QuantizationNoise: TraceValue{Value: 1.0 / maxCode, Unit: "LSB_rms", Uncertainty: 0},
		},
		Classifier: ClassifierTrace{
			Logit:              TraceValue{Value: (prob - 0.5) * 8.0, Unit: "logit", Uncertainty: 0.15},
			Probability:        TraceValue{Value: prob, Unit: "1", Uncertainty: 0.03},
			PredictedClass:     pred,
			ConfidenceInterval: TraceValue{Value: 0.95, Unit: "CI", Uncertainty: 0.02},
		},
	}
}
