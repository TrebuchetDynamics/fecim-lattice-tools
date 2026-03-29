package heracles

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	sharedio "fecim-lattice-tools/shared/io"
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

// WriteCompareReport serializes the report as indented JSON and writes it
// to the given path. Returns an error if the path is empty or invalid.
func WriteCompareReport(path string, report CompareReport) error {
	cleanPath, err := sharedio.ValidatePath(path)
	if err != nil {
		return fmt.Errorf("invalid report path: %w", err)
	}

	blob, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	blob = append(blob, '\n')

	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	return os.WriteFile(cleanPath, blob, 0o644)
}
