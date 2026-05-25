package module2

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	"fecim-lattice-tools/shared/crossbar"
	sharedval "fecim-lattice-tools/shared/validation"
)

const analyticalLimitsThresholdA = 1e-12

type analyticalLimitsReport struct {
	sharedval.ArtifactEnvelope

	Description  string                `json:"description"`
	ThresholdA   float64               `json:"threshold_A"`
	MaxAbsErrorA float64               `json:"max_abs_error_A"`
	Cases        []analyticalLimitCase `json:"cases"`
	Limitations  []string              `json:"limitations"`
	Pass         bool                  `json:"pass"`
}

type analyticalLimitCase struct {
	Name             string      `json:"name"`
	Description      string      `json:"description"`
	Rows             int         `json:"rows"`
	Cols             int         `json:"cols"`
	Parasitics       string      `json:"parasitics"`
	SneakPaths       string      `json:"sneak_paths"`
	ConductancesS    [][]float64 `json:"conductances_S"`
	AppliedVoltagesV []float64   `json:"applied_voltages_V"`
	ExpectedCurrentA []float64   `json:"expected_current_A"`
	ActualCurrentA   []float64   `json:"actual_current_A"`
	MaxAbsErrorA     float64     `json:"max_abs_error_A"`
	MaxOhmErrorA     float64     `json:"max_ohm_error_A"`
	Iterations       int         `json:"iterations"`
	Pass             bool        `json:"pass"`
}

type analyticalLimitFixture struct {
	name        string
	description string
	g           [][]float64
	v           []float64
}

func TestModule2AnalyticalLimits_PublicValidation(t *testing.T) {
	report := runAnalyticalLimitsSuite(t)

	if !report.Pass {
		t.Fatalf("analytical limit validation failed: max error %.3e A, threshold %.3e A", report.MaxAbsErrorA, report.ThresholdA)
	}
	if len(report.Cases) != 2 {
		t.Fatalf("unexpected analytical limit case count: got %d, want 2", len(report.Cases))
	}
}

func runAnalyticalLimitsSuite(t *testing.T) analyticalLimitsReport {
	t.Helper()

	fixtures := []analyticalLimitFixture{
		{
			name:        "single_cell_ohms_law",
			description: "1x1 crossbar has one current path, so I = G×V exactly.",
			g:           [][]float64{{75e-6}},
			v:           []float64{0.20},
		},
		{
			name:        "two_by_two_no_sneak_zero_parasitics",
			description: "2x2 zero-parasitic fixture decouples bit lines; each output current is the analytical column sum Σ_i G_ij×V_j with no sneak-path coupling.",
			g: [][]float64{
				{20e-6, 40e-6},
				{60e-6, 80e-6},
			},
			v: []float64{0.10, 0.25},
		},
	}

	report := analyticalLimitsReport{
		ArtifactEnvelope: sharedval.NewEnvelope("RG-VAL-M2-02", "", false),
		Description:      "Module 2 crossbar analytical-limit validation for single-cell Ohm's law and 2x2 no-sneak zero-parasitic fixtures",
		ThresholdA:       analyticalLimitsThresholdA,
		Limitations: []string{
			"These fixtures prove analytical limits of the deterministic solver path, not agreement with fabricated devices.",
			"The 2x2 no-sneak case disables wire parasitics; external SPICE comparison remains a separate validation layer.",
		},
		Cases: make([]analyticalLimitCase, 0, len(fixtures)),
	}

	for _, fixture := range fixtures {
		c := runAnalyticalLimitFixture(t, fixture)
		report.Cases = append(report.Cases, c)
		report.MaxAbsErrorA = maxAbsFloat(report.MaxAbsErrorA, c.MaxAbsErrorA)
		report.MaxAbsErrorA = maxAbsFloat(report.MaxAbsErrorA, c.MaxOhmErrorA)
	}

	report.Pass = report.MaxAbsErrorA <= report.ThresholdA
	report.ArtifactEnvelope = sharedval.NewEnvelope("RG-VAL-M2-02", "", report.Pass)
	writeAnalyticalLimitsReport(t, report)

	t.Logf("Module 2 analytical limits: cases=%d max_error=%.3e A threshold=%.3e A artifact=output/validation/module2/analytical_limits.json",
		len(report.Cases), report.MaxAbsErrorA, report.ThresholdA)

	return report
}

func runAnalyticalLimitFixture(t *testing.T, fixture analyticalLimitFixture) analyticalLimitCase {
	t.Helper()

	rows := len(fixture.g)
	cols := len(fixture.v)
	cfg := crossbar.DefaultSORConfig()
	cfg.MaxIterations = 10
	cfg.Tolerance = 1e-15

	solver, err := crossbar.NewParasiticSolver(rows, cols, cfg)
	if err != nil {
		t.Fatalf("%s NewParasiticSolver: %v", fixture.name, err)
	}
	solver.SetParasitics(0, 0)
	solver.SetConductances(fixture.g)

	result, err := solver.SolveMVMWithFallback(fixture.v)
	if err != nil {
		t.Fatalf("%s SolveMVMWithFallback: %v", fixture.name, err)
	}
	if !result.Converged {
		t.Fatalf("%s solver did not converge", fixture.name)
	}

	expected := expectedColumnCurrents(fixture.g, fixture.v)
	caseError := maxVectorAbsDiff(expected, result.OutputCurrents)
	ohmError := maxDeviceOhmError(fixture.g, fixture.v, result.DeviceCurrents)

	return analyticalLimitCase{
		Name:             fixture.name,
		Description:      fixture.description,
		Rows:             rows,
		Cols:             cols,
		Parasitics:       "disabled: RpRow=0, RpCol=0",
		SneakPaths:       "none in deterministic zero-parasitic analytical fixture",
		ConductancesS:    copyMatrix(fixture.g),
		AppliedVoltagesV: append([]float64(nil), fixture.v...),
		ExpectedCurrentA: expected,
		ActualCurrentA:   append([]float64(nil), result.OutputCurrents...),
		MaxAbsErrorA:     caseError,
		MaxOhmErrorA:     ohmError,
		Iterations:       result.Iterations,
		Pass:             caseError <= analyticalLimitsThresholdA && ohmError <= analyticalLimitsThresholdA,
	}
}

func expectedColumnCurrents(g [][]float64, applied []float64) []float64 {
	out := make([]float64, len(applied))
	for row := range g {
		for col := range applied {
			out[col] += g[row][col] * applied[col]
		}
	}
	return out
}

func maxDeviceOhmError(g [][]float64, applied []float64, actual [][]float64) float64 {
	maxErr := 0.0
	for row := range g {
		for col := range applied {
			expected := g[row][col] * applied[col]
			maxErr = maxAbsFloat(maxErr, actual[row][col]-expected)
		}
	}
	return maxErr
}

func maxVectorAbsDiff(expected, actual []float64) float64 {
	maxErr := 0.0
	for i := range expected {
		maxErr = maxAbsFloat(maxErr, actual[i]-expected[i])
	}
	return maxErr
}

func maxAbsFloat(current, candidate float64) float64 {
	if math.Abs(candidate) > current {
		return math.Abs(candidate)
	}
	return current
}

func copyMatrix(in [][]float64) [][]float64 {
	out := make([][]float64, len(in))
	for i := range in {
		out[i] = append([]float64(nil), in[i]...)
	}
	return out
}

func writeAnalyticalLimitsReport(t *testing.T, report analyticalLimitsReport) {
	t.Helper()

	outDir := filepath.Join(repoRoot(t), "output", "validation", "module2")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("create validation output directory: %v", err)
	}

	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("marshal analytical limits report: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "analytical_limits.json"), append(b, '\n'), 0o644); err != nil {
		t.Fatalf("write analytical limits report: %v", err)
	}
}
