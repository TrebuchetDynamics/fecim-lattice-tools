package physics

// TestPolydomainEnsemble_DomainCountConvergence documents how many
// polydomain ensemble members are needed for convergent results.
// This is essential for choosing ensemble size in research simulations.
//
// Methodology: For each domain count N in {5, 10, 20, 50, 100, 200},
// run a full P-E hysteresis loop (LK engine in ensemble mode), extract
// Pr and Ec, and check that the result converges as N increases.
//
// Convergence criterion: |Pr(N=100) - Pr(N=200)| < 2% of |Pr(N=200)|.
//
// Results are logged as a table suitable for supplementary material and
// written to output/validation/physics/polydomain_convergence.json.

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// polyConvergenceResult holds extracted loop metrics for a single domain count.
type polyConvergenceResult struct {
	DomainCount int     `json:"domain_count"`
	Pr          float64 `json:"pr_c_m2"`
	Ec          float64 `json:"ec_v_m"`
}

// polyConvergenceReport is the JSON artifact for domain-count convergence.
type polyConvergenceReport struct {
	TestID    string                  `json:"test_id"`
	Material  string                  `json:"material"`
	Seed      uint64                  `json:"seed"`
	Results   []polyConvergenceResult `json:"results"`
	Converged bool                    `json:"converged"`
	DeltaPr   float64                 `json:"delta_pr_frac"` // |Pr(100)-Pr(200)|/|Pr(200)|
}

func TestPolydomainEnsemble_DomainCountConvergence(t *testing.T) {
	mat := DefaultHZO()
	const seed uint64 = 99
	domainCounts := []int{5, 10, 20, 50, 100, 200}

	results := make([]polyConvergenceResult, 0, len(domainCounts))

	t.Logf("%-12s %12s %12s", "DomainCount", "Pr (µC/cm²)", "Ec (MV/cm)")
	t.Logf("%-12s %12s %12s", "———————————", "———————————", "———————————")

	for _, n := range domainCounts {
		pr, ec := runPolydomainPELoop(mat, n, seed)
		results = append(results, polyConvergenceResult{
			DomainCount: n,
			Pr:          pr,
			Ec:          ec,
		})
		t.Logf("%-12d %12.4f %12.4f", n, pr*1e6, ec/1e8)
	}

	// Convergence check: |Pr(N=100) - Pr(N=200)| < 2% of |Pr(N=200)|.
	var pr100, pr200 float64
	for _, r := range results {
		if r.DomainCount == 100 {
			pr100 = r.Pr
		}
		if r.DomainCount == 200 {
			pr200 = r.Pr
		}
	}

	deltaPr := 0.0
	if pr200 != 0 {
		deltaPr = math.Abs(pr100-pr200) / math.Abs(pr200)
	}
	converged := deltaPr < 0.02
	if !converged {
		t.Errorf("Pr not converged: |Pr(100)-Pr(200)|/|Pr(200)| = %.4f (>2%%)", deltaPr)
	}
	t.Logf("Convergence: delta_Pr = %.4f%% (threshold: 2%%)", deltaPr*100)

	// Write JSON artifact.
	report := polyConvergenceReport{
		TestID:    "polydomain_domain_count_convergence",
		Material:  mat.Name,
		Seed:      seed,
		Results:   results,
		Converged: converged,
		DeltaPr:   deltaPr,
	}

	outDir := os.Getenv("FECIM_LITERATURE_JSON_DIR")
	if outDir == "" {
		outDir = filepath.Join("output", "validation", "physics")
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Logf("warn: could not create output dir %s: %v", outDir, err)
		return
	}
	outPath := filepath.Join(outDir, "polydomain_convergence.json")
	b, _ := json.MarshalIndent(report, "", "  ")
	if err := os.WriteFile(outPath, b, 0o644); err != nil {
		t.Logf("warn: could not write artifact %s: %v", outPath, err)
	} else {
		t.Logf("artifact: %s", outPath)
	}
}

// runPolydomainPELoop runs a full P-E hysteresis loop using the LK solver in
// polydomain ensemble mode with the given domain count, and returns (Pr, Ec).
func runPolydomainPELoop(mat *HZOMaterial, domainCount int, seed uint64) (float64, float64) {
	solver := NewLKSolver()
	solver.ConfigureFromMaterial(mat)
	solver.EnableNoise = false
	solver.UseNLS = false
	if domainCount > 1 {
		solver.EnableEnsemble(domainCount, mat, seed)
	}
	solver.SetState(-math.Abs(mat.Ps))

	eMax := 2.0 * mat.Ec
	const nPoints = 201
	const dt = 5e-6
	const settleSteps = 50

	// Build E waveform: ascending -Emax → +Emax, then descending +Emax → -Emax.
	totalPoints := 2 * nPoints
	eWave := make([]float64, totalPoints)
	for i := 0; i < nPoints; i++ {
		eWave[i] = -eMax + 2*eMax*float64(i)/float64(nPoints-1)
	}
	for i := 0; i < nPoints; i++ {
		eWave[nPoints+i] = eMax - 2*eMax*float64(i)/float64(nPoints-1)
	}

	// Step through the waveform, settling at each E-field value.
	pTrace := make([]float64, totalPoints)
	for i, e := range eWave {
		for j := 0; j < settleSteps; j++ {
			solver.Step(e, dt)
		}
		pTrace[i] = solver.GetState()
	}

	// Extract Pr: P at E≈0 on descending branch (second half).
	pr := 0.0
	for i := nPoints; i < totalPoints-1; i++ {
		if eWave[i] >= 0 && eWave[i+1] < 0 {
			frac := eWave[i] / (eWave[i] - eWave[i+1])
			pr = pTrace[i] + frac*(pTrace[i+1]-pTrace[i])
			break
		}
	}

	// Extract Ec: E at P≈0 on ascending branch (first half).
	ec := 0.0
	for i := 0; i < nPoints-1; i++ {
		if pTrace[i] <= 0 && pTrace[i+1] > 0 {
			frac := -pTrace[i] / (pTrace[i+1] - pTrace[i])
			ec = eWave[i] + frac*(eWave[i+1]-eWave[i])
			break
		}
	}

	return pr, ec
}

func init() {
	// Ensure the package-level variable reference resolves for test-only logging.
	_ = fmt.Sprintf
}
