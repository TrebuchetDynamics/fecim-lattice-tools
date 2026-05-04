package eda

type EDAState struct {
	DesignName    string   `json:"design_name"`
	ProcessNode   string   `json:"process_node"`
	ArrayRows     int      `json:"array_rows"`
	ArrayCols     int      `json:"array_cols"`
	ExportFormats []string `json:"export_formats"`
}
