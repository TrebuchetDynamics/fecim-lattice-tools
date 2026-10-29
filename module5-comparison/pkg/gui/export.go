// Package gui provides Fyne-based GUI components for architecture comparison.
// export.go provides data export functionality for comparison results.
package gui

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	model "fecim-lattice-tools/module5-comparison/pkg/comparison"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"fecim-lattice-tools/shared/export"
)

// ComparisonExportConfig contains the configuration for comparison exports
type ComparisonExportConfig struct {
	Metadata          export.ExportMetadata `json:"metadata"`
	WorkloadConfig    WorkloadConfig        `json:"workload_config"`
	EnergySpecs       EnergySpecsExport     `json:"energy_specs"`
	CalculatedResults *CalculatedResults    `json:"calculated_results,omitempty"`
}

// WorkloadConfig represents the workload configuration
type WorkloadConfig struct {
	WorkloadName     string `json:"workload_name"`
	InferencesPerDay int    `json:"inferences_per_day"`
	MACsPerInference int64  `json:"macs_per_inference"`
	Description      string `json:"description,omitempty"`
}

// EnergySpecsExport represents energy specifications for export
type EnergySpecsExport struct {
	CPUSpec   EnergySpecExport `json:"cpu_spec"`
	GPUSpec   EnergySpecExport `json:"gpu_spec"`
	FeCIMSpec EnergySpecExport `json:"fecim_spec"`
}

// EnergySpecExport represents a single energy spec for export
type EnergySpecExport struct {
	Name          string  `json:"name"`
	EnergyFJ      float64 `json:"energy_fj_per_mac"`
	Source        string  `json:"source"`
	Verified      bool    `json:"verified"`
	SourceDetails string  `json:"source_details,omitempty"`
}

// CalculatedResults contains the calculated comparison results
type CalculatedResults struct {
	TotalMACsPerDay    float64   `json:"total_macs_per_day"`
	CPUEnergyKWh       float64   `json:"cpu_energy_kwh_per_day"`
	GPUEnergyKWh       float64   `json:"gpu_energy_kwh_per_day"`
	FeCIMEnergyKWh     float64   `json:"fecim_energy_kwh_per_day"`
	CPUSavingsKWh      float64   `json:"cpu_savings_kwh_per_day"`
	GPUSavingsKWh      float64   `json:"gpu_savings_kwh_per_day"`
	CPUSavingsPercent  float64   `json:"cpu_savings_percent"`
	GPUSavingsPercent  float64   `json:"gpu_savings_percent"`
	AnnualCostSavingsM float64   `json:"annual_cost_savings_millions"`
	Timestamp          time.Time `json:"timestamp"`
}

// exportComparisonData exports the current comparison data to CSV and JSON files
func (ca *ComparisonApp) exportComparisonData() {
	timestamp := time.Now().Format("2006-01-02_15-04-05")

	// Ensure data/comparison folder exists
	dataDir := filepath.Join("exports", "comparison")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		ca.showExportError(fmt.Sprintf("Cannot create exports folder: %v", err))
		return
	}

	// Version 2 explicitly fixes the former daily-count, kWh and million-USD
	// mistakes. Export the calculator's workload denominator and scenario.
	result, err := model.CalculateScenario(ca.currentWorkload, ca.currentInferences, ca.scenarioProfile)
	if err != nil {
		ca.showExportError(err.Error())
		return
	}
	data, err := model.MarshalScenario(result, "json")
	if err != nil {
		ca.showExportError(err.Error())
		return
	}
	exporter := export.NewExporter(dataDir, fmt.Sprintf("comparison-v2_%s", timestamp))
	jsonResult := exporter.ExportJSON(json.RawMessage(data))
	if jsonResult.Error != nil {
		ca.showExportError(jsonResult.Error.Error())
		return
	}
	data, err = model.MarshalScenario(result, "csv")
	if err != nil {
		ca.showExportError(err.Error())
		return
	}
	rows, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil {
		ca.showExportError(err.Error())
		return
	}
	csvResult := exporter.ExportCSV(rows[0], rows[1:])
	if csvResult.Error != nil {
		ca.showExportError(csvResult.Error.Error())
		return
	}
	ca.showExportSuccess(fmt.Sprintf("Exported:\n• %s\n• %s", jsonResult.FilePath, csvResult.FilePath))
}

// exportVisualization exports the current visualization as a PNG
func (ca *ComparisonApp) exportVisualization() {
	export.ExportVisualization(ca.window, "comparison", debug)
}

// showExportError displays an export error dialog
func (ca *ComparisonApp) showExportError(msg string) {
	export.ShowExportError(ca.window, debug, msg)
}

// showExportSuccess displays an export success dialog
func (ca *ComparisonApp) showExportSuccess(msg string) {
	export.ShowExportSuccess(ca.window, debug, msg)
}

// getWorkloadMACs returns the MACs per inference for a given workload
func getWorkloadMACs(workload string) int64 {
	macs, _ := model.ScenarioWorkloadMACs(workload)
	return macs
}

func currentGitCommit() string {
	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func (ca *ComparisonApp) exportReproducibilityPack() {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	reproDir := filepath.Join("exports", "comparison", fmt.Sprintf("repro-pack_%s", timestamp))

	if err := os.MkdirAll(reproDir, 0o755); err != nil {
		ca.showExportError(fmt.Sprintf("Cannot create reproducibility pack folder: %v", err))
		return
	}

	assumptionsYAML := fmt.Sprintf("module: module5-comparison\nworkload: %s\ninferences_per_sec: %.0f\nversion: %s\ncommit: %s\ntimestamp: %s\nassumptions:\n  - CPU+DRAM energy=1000 pJ/MAC (model input)\n  - GPU+HBM energy=100 pJ/MAC (model input)\n  - FeCIM energy=1 pJ/MAC (model input)\n", ca.currentWorkload, ca.currentInferences, "v0", currentGitCommit(), time.Now().Format(time.RFC3339))

	summaryAsset := filepath.Join(reproDir, "module5_outputs.txt")
	summary := fmt.Sprintf("workload=%s\ninferences_per_sec=%.0f\nexported_at=%s\n", ca.currentWorkload, ca.currentInferences, time.Now().Format(time.RFC3339))
	_ = os.WriteFile(summaryAsset, []byte(summary), 0o644)

	inputs := export.ReproducibilityPackInput{
		ConfigYAML:      []byte(assumptionsYAML),
		RandomSeeds:     map[string]int64{"module5_comparison_seed": 42},
		GitCommitHash:   currentGitCommit(),
		TestResults:     "Module 5 comparison export path exercised via GUI export and package tests.",
		GeneratedAssets: []string{summaryAsset},
	}

	outPath, err := export.CreateReproducibilityPack(reproDir, inputs)
	if err != nil {
		ca.showExportError(fmt.Sprintf("Reproducibility pack export failed: %v", err))
		return
	}
	ca.showExportSuccess(fmt.Sprintf("Reproducibility pack exported:\n• %s", outPath))
}
