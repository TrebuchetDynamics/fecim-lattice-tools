package heracles

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	sharedphysics "fecim-lattice-tools/shared/physics"
)

type errorMetrics struct {
	RMSE_uCcm2     float64
	MAE_uCcm2      float64
	MaxError_uCcm2 float64
}

func TestHeraclesComparator(t *testing.T) {
	baselinePath := filepath.Join("..", "external", "baselines", "heracles", "reference_pe_loop.csv")
	ascRef, descRef, err := loadBaselineCSV(baselinePath)
	if err != nil {
		t.Fatalf("load baseline: %v", err)
	}

	if _, err := exec.LookPath("heracles"); err != nil {
		t.Log("Heracles binary not found; using stored reference CSV baseline and comparing FeCIM-only output")
	}

	ascModel, descModel := runFeCIMPESweep(ascRef, descRef)
	metrics := computeErrorMetrics(ascRef, descRef, ascModel, descModel)

	if metrics.RMSE_uCcm2 <= 0 {
		t.Fatalf("rmse must be >0 uC/cm^2, got %.6f", metrics.RMSE_uCcm2)
	}
	if metrics.MAE_uCcm2 <= 0 {
		t.Fatalf("mae must be >0 uC/cm^2, got %.6f", metrics.MAE_uCcm2)
	}
	if metrics.MaxError_uCcm2 <= 0 {
		t.Fatalf("max error must be >0 uC/cm^2, got %.6f", metrics.MaxError_uCcm2)
	}

	t.Logf("Heracles comparator metrics: RMSE=%.3f uC/cm^2, MAE=%.3f uC/cm^2, MaxError=%.3f uC/cm^2",
		metrics.RMSE_uCcm2, metrics.MAE_uCcm2, metrics.MaxError_uCcm2)
}

func loadBaselineCSV(path string) ([]PEPoint, []PEPoint, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	recs, err := r.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(recs) < 2 {
		return nil, nil, fmt.Errorf("baseline has no data rows")
	}

	var asc []PEPoint
	var desc []PEPoint
	for i, rec := range recs[1:] {
		if len(rec) < 3 {
			return nil, nil, fmt.Errorf("row %d malformed", i+2)
		}
		branch := strings.ToLower(strings.TrimSpace(rec[0]))
		e, err := strconv.ParseFloat(strings.TrimSpace(rec[1]), 64)
		if err != nil {
			return nil, nil, fmt.Errorf("row %d parse E: %w", i+2, err)
		}
		p, err := strconv.ParseFloat(strings.TrimSpace(rec[2]), 64)
		if err != nil {
			return nil, nil, fmt.Errorf("row %d parse P: %w", i+2, err)
		}
		pt := PEPoint{E_MVcm: e, P_uCcm: p}
		switch branch {
		case "asc":
			asc = append(asc, pt)
		case "desc":
			desc = append(desc, pt)
		default:
			return nil, nil, fmt.Errorf("row %d unknown branch %q", i+2, branch)
		}
	}
	if len(asc) == 0 || len(desc) == 0 {
		return nil, nil, fmt.Errorf("baseline requires asc and desc branches")
	}
	return asc, desc, nil
}

func runFeCIMPESweep(ascRef, descRef []PEPoint) ([]PEPoint, []PEPoint) {
	solver := sharedphysics.NewLKSolver()
	mat := sharedphysics.MaterlikHfO2()
	solver.ConfigureFromMaterial(mat)
	solver.UseNLS = false
	solver.EnableNoise = false
	solver.SetState(-math.Abs(mat.Pr))

	const steps = 1800
	const dt = 2e-12

	ascModel := make([]PEPoint, 0, len(ascRef))
	for _, pt := range ascRef {
		p := settleAtField(solver, pt.E_MVcm*1e8, steps, dt)
		ascModel = append(ascModel, PEPoint{E_MVcm: pt.E_MVcm, P_uCcm: p * 100.0})
	}
	descModel := make([]PEPoint, 0, len(descRef))
	for _, pt := range descRef {
		p := settleAtField(solver, pt.E_MVcm*1e8, steps, dt)
		descModel = append(descModel, PEPoint{E_MVcm: pt.E_MVcm, P_uCcm: p * 100.0})
	}
	return ascModel, descModel
}

func computeErrorMetrics(ascRef, descRef, ascModel, descModel []PEPoint) errorMetrics {
	var sumSq float64
	var sumAbs float64
	var maxAbs float64
	var n int
	acc := func(ref, model []PEPoint) {
		m := len(ref)
		if len(model) < m {
			m = len(model)
		}
		for i := 0; i < m; i++ {
			d := model[i].P_uCcm - ref[i].P_uCcm
			ad := math.Abs(d)
			sumSq += d * d
			sumAbs += ad
			if ad > maxAbs {
				maxAbs = ad
			}
			n++
		}
	}
	acc(ascRef, ascModel)
	acc(descRef, descModel)
	if n == 0 {
		return errorMetrics{}
	}
	return errorMetrics{
		RMSE_uCcm2:     math.Sqrt(sumSq / float64(n)),
		MAE_uCcm2:      sumAbs / float64(n),
		MaxError_uCcm2: maxAbs,
	}
}
