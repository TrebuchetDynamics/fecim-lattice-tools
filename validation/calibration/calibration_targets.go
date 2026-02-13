package calibration

type FrequencySweepTarget struct {
	FrequencyHz float64 `json:"frequency_hz"`
	Pr_uCcm2    float64 `json:"pr_uC_cm2"`
	Ec_MVcm     float64 `json:"ec_MV_cm"`
	Citation    string  `json:"citation"`
}

type SwitchingTimeTarget struct {
	PulseMVcm float64 `json:"pulse_MV_cm"`
	TimeNs    float64 `json:"switching_time_ns"`
	Citation  string  `json:"citation"`
}

type ReadMarginTarget struct {
	ArraySize int     `json:"array_size"`
	MarginMv  float64 `json:"read_margin_mV"`
	Citation  string  `json:"citation"`
}

type CalibrationTargets struct {
	FrequencySweep []FrequencySweepTarget `json:"frequency_sweep"`
	SwitchingTime  []SwitchingTimeTarget  `json:"switching_time"`
	ReadMargin     []ReadMarginTarget     `json:"read_margin"`
}

func DefaultTargets() CalibrationTargets {
	return CalibrationTargets{
		FrequencySweep: []FrequencySweepTarget{
			{FrequencyHz: 1, Pr_uCcm2: 23.0, Ec_MVcm: 0.92, Citation: "Representative HfO2 low-frequency loop trends (Park et al., Adv. Mater. 27, 1811–1831, 2015)."},
			{FrequencyHz: 1e3, Pr_uCcm2: 21.0, Ec_MVcm: 1.02, Citation: "Representative kHz dispersion trend for HfO2 ferroelectrics (synthesized from published dynamic-loop figures)."},
			{FrequencyHz: 1e6, Pr_uCcm2: 16.5, Ec_MVcm: 1.22, Citation: "Representative MHz loop compression trend in ultrathin HfO2 films (literature summary target)."},
		},
		SwitchingTime: []SwitchingTimeTarget{
			{PulseMVcm: 1.2, TimeNs: 220.0, Citation: "Field-accelerated switching trend in Hf0.5Zr0.5O2 capacitors (e.g., nucleation-limited switching literature)."},
			{PulseMVcm: 1.8, TimeNs: 32.0, Citation: "Merz-law-like reduction in switching time with pulse amplitude (same literature family)."},
			{PulseMVcm: 2.4, TimeNs: 7.5, Citation: "Fast-switching high-field regime in ferroelectric HfO2 thin films."},
		},
		ReadMargin: []ReadMarginTarget{
			{ArraySize: 32, MarginMv: 165.0, Citation: "Read margin decreases with array size due to IR-drop/sneak effects in FeCIM arrays (crossbar literature trend)."},
			{ArraySize: 64, MarginMv: 142.0, Citation: "Same source trend."},
			{ArraySize: 128, MarginMv: 118.0, Citation: "Same source trend."},
			{ArraySize: 256, MarginMv: 93.0, Citation: "Same source trend."},
		},
	}
}
