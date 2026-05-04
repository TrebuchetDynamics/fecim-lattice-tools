package circuits

type CircuitsState struct {
	ADCResolution    int     `json:"adc_resolution"`
	DACResolution    int     `json:"dac_resolution"`
	TIAGain          float64 `json:"tia_gain"`
	ChargePumpStages int     `json:"charge_pump_stages"`
	SupplyVoltage    float64 `json:"supply_voltage"`
	ISPPEnabled      bool    `json:"ispp_enabled"`
}
