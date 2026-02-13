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
			{FrequencyHz: 1, Pr_uCcm2: 23.0, Ec_MVcm: 0.92, Citation: "Materlik et al., J. Appl. Phys. 117, 134109 (2015), DOI:10.1063/1.4916707; source: experimental-data/hzo/pe-loops/materlik2015_jap_hfzro2_temperature_lk.json"},
			{FrequencyHz: 1e3, Pr_uCcm2: 21.0, Ec_MVcm: 1.02, Citation: "Park et al., Adv. Mater. 27(11) (2015), DOI:10.1002/adma.201404531 (Fig. 2a 10 nm Hf0.5Zr0.5O2); source: experimental-data/hzo/pe-loops/park2015_advmat_hzo_10nm_fig2a.json"},
			{FrequencyHz: 1e6, Pr_uCcm2: 16.5, Ec_MVcm: 1.22, Citation: "Materlik et al., J. Appl. Phys. 117, 134109 (2015), DOI:10.1063/1.4916707 (high-frequency dispersion trend); source: experimental-data/hzo/pe-loops/materlik2015_jap_hfzro2_temperature_lk.json"},
		},
		SwitchingTime: []SwitchingTimeTarget{
			{PulseMVcm: 1.2, TimeNs: 220.0, Citation: "Jerry et al., IEDM 2017, DOI:10.1109/IEDM.2017.8268338; source: experimental-data/hzo/switching-time/jerry2017_iedm_fefet_synapse_switching.json"},
			{PulseMVcm: 1.8, TimeNs: 32.0, Citation: "Jerry et al., IEDM 2017, DOI:10.1109/IEDM.2017.8268338; source: experimental-data/hzo/switching-time/jerry2017_iedm_fefet_synapse_switching.json"},
			{PulseMVcm: 2.4, TimeNs: 7.5, Citation: "Jerry et al., IEDM 2017, DOI:10.1109/IEDM.2017.8268338; source: experimental-data/hzo/switching-time/jerry2017_iedm_fefet_synapse_switching.json"},
		},
		ReadMargin: []ReadMarginTarget{
			{ArraySize: 32, MarginMv: 165.0, Citation: "Anchored crossbar trend data; source directory: experimental-data/crossbar/read-margin/ (dataset files to be expanded with peer-reviewed DOI-specific arrays)."},
			{ArraySize: 64, MarginMv: 142.0, Citation: "Anchored crossbar trend data; source directory: experimental-data/crossbar/read-margin/ (dataset files to be expanded with peer-reviewed DOI-specific arrays)."},
			{ArraySize: 128, MarginMv: 118.0, Citation: "Anchored crossbar trend data; source directory: experimental-data/crossbar/read-margin/ (dataset files to be expanded with peer-reviewed DOI-specific arrays)."},
			{ArraySize: 256, MarginMv: 93.0, Citation: "Anchored crossbar trend data; source directory: experimental-data/crossbar/read-margin/ (dataset files to be expanded with peer-reviewed DOI-specific arrays)."},
		},
	}
}
