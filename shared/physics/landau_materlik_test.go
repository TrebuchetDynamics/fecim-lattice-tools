package physics

import (
	"math"
	"testing"
)

func interpXAtY0(x1, y1, x2, y2 float64) (float64, bool) {
	dy := y2 - y1
	if dy == 0 {
		return 0, false
	}
	f := -y1 / dy
	if f < 0 || f > 1 {
		return 0, false
	}
	return x1 + f*(x2-x1), true
}

func interpYAtX0(x1, y1, x2, y2 float64) (float64, bool) {
	dx := x2 - x1
	if dx == 0 {
		return 0, false
	}
	f := -x1 / dx
	if f < 0 || f > 1 {
		return 0, false
	}
	return y1 + f*(y2-y1), true
}

func TestLandauMaterlik_PELoop_PrEcAt300K10nm(t *testing.T) {
	mat := MaterlikHfO2()

	s := NewLKSolver()
	s.ConfigureFromMaterial(mat)
	s.Temperature = 300
	s.UseNLS = false // deterministic loop extraction
	s.EnableNoise = false
	s.SetState(-math.Abs(mat.Pr))

	const (
		eMax          = 3.0e8 // ±3 MV/cm
		nPtsHalf      = 301
		stepsPerPoint = 3000
		dt            = 2e-12
	)

	fields := make([]float64, 0, 2*nPtsHalf)
	pols := make([]float64, 0, 2*nPtsHalf)

	for i := 0; i < nPtsHalf; i++ {
		E := -eMax + (2*eMax*float64(i))/float64(nPtsHalf-1)
		for k := 0; k < stepsPerPoint; k++ {
			s.Step(E, dt)
		}
		fields = append(fields, E)
		pols = append(pols, s.GetState())
	}
	for i := 0; i < nPtsHalf; i++ {
		E := eMax - (2*eMax*float64(i))/float64(nPtsHalf-1)
		for k := 0; k < stepsPerPoint; k++ {
			s.Step(E, dt)
		}
		fields = append(fields, E)
		pols = append(pols, s.GetState())
	}

	// Pr from both E≈0 crossings (up and down sweep), then average |Pr|.
	var prVals []float64
	for i := 1; i < len(fields); i++ {
		if fields[i-1] == 0 {
			prVals = append(prVals, math.Abs(pols[i-1]))
			continue
		}
		if fields[i-1]*fields[i] <= 0 {
			p0, ok := interpYAtX0(fields[i-1], pols[i-1], fields[i], pols[i])
			if ok {
				prVals = append(prVals, math.Abs(p0))
			}
		}
	}
	if len(prVals) == 0 {
		t.Fatalf("failed to extract Pr from loop")
	}
	pr := 0.0
	for _, v := range prVals {
		pr += v
	}
	pr /= float64(len(prVals))

	// Ec from zero-polarization crossings on each branch.
	var ecVals []float64
	for i := 1; i < len(pols); i++ {
		if pols[i-1]*pols[i] <= 0 {
			ec, ok := interpXAtY0(pols[i-1], fields[i-1], pols[i], fields[i])
			if ok {
				ecVals = append(ecVals, math.Abs(ec))
			}
		}
	}
	if len(ecVals) == 0 {
		bestI := 0
		bestAbsP := math.Abs(pols[0])
		for i := 1; i < len(pols); i++ {
			if v := math.Abs(pols[i]); v < bestAbsP {
				bestAbsP = v
				bestI = i
			}
		}
		ecVals = append(ecVals, math.Abs(fields[bestI]))
	}
	ec := 0.0
	for _, v := range ecVals {
		ec += v
	}
	ec /= float64(len(ecVals))

	prUCcm2 := pr * 100.0 // C/m² -> µC/cm²
	ecMVcm := ec / 1.0e8  // V/m -> MV/cm
	t.Logf("Materlik loop extracted: Pr=%.3f uC/cm^2, Ec=%.3f MV/cm", prUCcm2, ecMVcm)
	if math.Abs(prUCcm2-20.0) > 2.0 {
		t.Fatalf("Pr = %.3f uC/cm^2, expected 20 ± 2", prUCcm2)
	}
	if math.Abs(ecMVcm-1.0) > 0.2 {
		t.Fatalf("Ec = %.3f MV/cm, expected 1.0 ± 0.2", ecMVcm)
	}
}
