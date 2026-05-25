package hysteresis

const (
	EventSelectMaterial                 = "select_material"
	EventSetFieldRange                  = "set_field_range"
	EventSetWaveform                    = "set_waveform"
	EventToggleSimulation               = "toggle_simulation"
	EventExportCSV                      = "export_csv"
	EventRunPUND                        = "run_pund"
	EventRunFORC                        = "run_forc"
	EventRunLevelCalibration            = "run_level_calibration"
	EventSetLevelCalibrationLevelCount  = "set_level_calibration_level_count"
	EventSetLevelCalibrationTargetRange = "set_level_calibration_target_range"
	EventSetLevelCalibrationTemperature = "set_level_calibration_temperature"
	EventExportLevelCalibration         = "export_level_calibration_json"
	EventExportPUNDCSV                  = "export_pund_csv"
	EventExportFORCSweep                = "export_forc_sweep_csv"
	EventExportFORCMatrix               = "export_forc_matrix_csv"
	EventExportFORCMeta                 = "export_forc_metadata_json"
)

const (
	WaveformSine     = "sine"
	WaveformTriangle = "triangle"
	WaveformSquare   = "square"
	WaveformManual   = "manual"
)
