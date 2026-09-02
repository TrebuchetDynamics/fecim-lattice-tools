package project

type Direction string

const (
	Minimize Direction = "minimize"
	Maximize Direction = "maximize"
)

type Objective struct {
	Metric    string    `yaml:"metric" json:"metric"`
	Direction Direction `yaml:"direction" json:"direction"`
}

type Constraint struct {
	Metric   string  `yaml:"metric" json:"metric"`
	Operator string  `yaml:"operator" json:"operator"`
	Value    float64 `yaml:"value" json:"value"`
	Unit     string  `yaml:"unit" json:"unit"`
}

type InputRef struct {
	Path     string `yaml:"path" json:"path"`
	SHA256   string `yaml:"sha256" json:"sha256"`
	Citation string `yaml:"citation" json:"citation"`
	Evidence string `yaml:"evidence" json:"evidence"`
}

type Project struct {
	SchemaVersion int          `yaml:"schema_version" json:"schema_version"`
	ID            string       `yaml:"id" json:"id"`
	Name          string       `yaml:"name" json:"name"`
	Hypothesis    string       `yaml:"hypothesis" json:"hypothesis"`
	ModelVersion  string       `yaml:"model_version" json:"model_version"`
	Objectives    []Objective  `yaml:"objectives" json:"objectives"`
	Constraints   []Constraint `yaml:"constraints" json:"constraints"`
	Citations     []string     `yaml:"citations" json:"citations"`
	Inputs        []InputRef   `yaml:"inputs" json:"inputs"`
}

type Device struct {
	Material          string  `yaml:"material" json:"material"`
	ConductanceLevels int     `yaml:"conductance_levels" json:"conductance_levels"`
	GMinS             float64 `yaml:"g_min_s" json:"g_min_s"`
	GMaxS             float64 `yaml:"g_max_s" json:"g_max_s"`
}

type Array struct {
	Rows         int     `yaml:"rows" json:"rows"`
	Cols         int     `yaml:"cols" json:"cols"`
	ReadVoltageV float64 `yaml:"read_voltage_v" json:"read_voltage_v"`
}

type Circuit struct {
	ADCBits    int     `yaml:"adc_bits" json:"adc_bits"`
	DACBits    int     `yaml:"dac_bits" json:"dac_bits"`
	TIAGainOhm float64 `yaml:"tia_gain_ohm" json:"tia_gain_ohm"`
	TechNode   string  `yaml:"tech_node" json:"tech_node"`
}

type Design struct {
	SchemaVersion int     `yaml:"schema_version" json:"schema_version"`
	Device        Device  `yaml:"device" json:"device"`
	Array         Array   `yaml:"array" json:"array"`
	Circuit       Circuit `yaml:"circuit" json:"circuit"`
}

type LinearRange struct {
	Start float64 `yaml:"start" json:"start"`
	Stop  float64 `yaml:"stop" json:"stop"`
	Count int     `yaml:"count" json:"count"`
}

type Parameter struct {
	Path   string       `yaml:"path" json:"path"`
	Values []float64    `yaml:"values" json:"values"`
	Range  *LinearRange `yaml:"range" json:"range,omitempty"`
}

type Sweep struct {
	SchemaVersion int         `yaml:"schema_version" json:"schema_version"`
	Seed          int64       `yaml:"seed" json:"seed"`
	MaxPoints     int         `yaml:"max_points" json:"max_points"`
	Parameters    []Parameter `yaml:"parameters" json:"parameters"`
}

type ResolvedInput struct {
	Path     string `json:"path"`
	SHA256   string `json:"sha256"`
	Citation string `json:"citation"`
	Evidence string `json:"evidence"`
}

type Bundle struct {
	Root    string
	Project Project
	Design  Design
	Sweep   Sweep
	Inputs  []ResolvedInput
}

type LoadOptions struct {
	CitationDir string
}
