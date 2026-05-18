package hysteresis

const (
	EventSelectMaterial   = "select_material"
	EventSetFieldRange    = "set_field_range"
	EventSetWaveform      = "set_waveform"
	EventToggleSimulation = "toggle_simulation"
	EventExportCSV        = "export_csv"
	EventRunPUND          = "run_pund"
	EventRunFORC          = "run_forc"
	EventExportPUNDCSV    = "export_pund_csv"
	EventExportFORCSweep  = "export_forc_sweep_csv"
	EventExportFORCMatrix = "export_forc_matrix_csv"
	EventExportFORCMeta   = "export_forc_metadata_json"
)

const (
	WaveformSine     = "sine"
	WaveformTriangle = "triangle"
	WaveformSquare   = "square"
	WaveformManual   = "manual"
)
