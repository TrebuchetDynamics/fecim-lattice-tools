package heracles

import (
	"encoding/json"
	"os"
)

type CurvePair struct {
	Reference []PEPoint `json:"reference"`
	Model     []PEPoint `json:"model"`
}

type CompareMetrics struct {
	RMSE_uCcm2          float64 `json:"rmse_uC_cm2"`
	EcRef_MVcm          float64 `json:"ec_reference_MV_cm"`
	EcModel_MVcm        float64 `json:"ec_model_MV_cm"`
	EcMismatchPct       float64 `json:"ec_mismatch_percent"`
	PrRef_uCcm2         float64 `json:"pr_reference_uC_cm2"`
	PrModel_uCcm2       float64 `json:"pr_model_uC_cm2"`
	PrMismatchPct       float64 `json:"pr_mismatch_percent"`
	LoopAreaRef_Jm3     float64 `json:"loop_area_reference_J_m3"`
	LoopAreaModel_Jm3   float64 `json:"loop_area_model_J_m3"`
	LoopAreaMismatchPct float64 `json:"loop_area_mismatch_percent"`
}

type CompareReport struct {
	Title      string         `json:"title"`
	Reference  string         `json:"reference"`
	Dataset    string         `json:"dataset"`
	Parameters map[string]any `json:"lk_parameters"`
	Ascending  CurvePair      `json:"ascending_branch"`
	Descending CurvePair      `json:"descending_branch"`
	Metrics    CompareMetrics `json:"metrics"`
}

func WriteCompareReport(path string, report CompareReport) error {
	blob, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	blob = append(blob, '\n')
	return os.WriteFile(path, blob, 0o644)
}
