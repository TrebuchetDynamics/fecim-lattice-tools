package physics

import (
	"math"
	"testing"
)

func TestPolydomainEnsemble_IntermediateRemanentStates(t *testing.T) {
	mat := LiteratureSuperlattice()
	s := NewLKSolver()
	s.ConfigureFromMaterial(mat)
	s.EnableNoise = false
	s.UseNLS = false
	s.EnableEnsemble(96, mat, 7)

	// Strong positive write pulse, then relax at E=0.
	eMax := 2.5 * mat.Ec
	dt := 5e-6
	for i := 0; i < 400; i++ {
		s.Step(eMax, dt)
	}
	for i := 0; i < 1200; i++ {
		s.Step(0, dt)
	}

	pRem := s.GetState()
	if math.Abs(math.Abs(pRem)-math.Abs(mat.Ps)) < 0.02*math.Abs(mat.Ps) {
		t.Fatalf("remanent state collapsed to saturation: Prem=%.6f Ps=%.6f", pRem, mat.Ps)
	}
	if math.Abs(pRem) <= 0.10*math.Abs(mat.Ps) {
		t.Fatalf("remanent state too close to zero, expected intermediate switched state: Prem=%.6f Ps=%.6f", pRem, mat.Ps)
	}
}

func TestPolydomainEnsemble_30Levels(t *testing.T) {
	mat := LiteratureSuperlattice()
	numLevels := 30
	success := 0

	for lvl := 0; lvl < numLevels; lvl++ {
		solver := NewLKSolver()
		solver.ConfigureFromMaterial(mat)
		solver.EnableNoise = false
		solver.UseNLS = true
		solver.EnableEnsemble(96, mat, 11)

		targetP := -mat.Ps + (2*mat.Ps*float64(lvl))/float64(numLevels-1)
		targetG := PolarizationToConductance(targetP, mat.Ps, mat.Gmin, mat.Gmax)

		wc := NewWriteController(solver, mat)
		wc.MaxVoltage = 2.5 * mat.Ec * mat.Thickness
		wc.MaxIterations = 25
		wc.PulseWidth = 2e-3
		wc.MaxStep = 5e-6
		wc.Tolerance = 0.03
		attempts, ok, _ := wc.WriteTargetWithReset(targetG, true)
		if ok && attempts <= 25 {
			success++
		}
	}

	if success < 25 {
		t.Fatalf("insufficient level convergence: %d/30 (need >=25)", success)
	}
}

func TestPolydomainEnsemble_ReproducibleWithSeed(t *testing.T) {
	mat := LiteratureSuperlattice()
	mk := func() *LKSolver {
		s := NewLKSolver()
		s.ConfigureFromMaterial(mat)
		s.EnableNoise = false
		s.UseNLS = false
		s.EnableEnsemble(96, mat, 12345)
		s.SetState(-math.Abs(mat.Ps))
		return s
	}

	a := mk()
	b := mk()
	wave := []float64{0.3, 0.8, 1.2, 0.6, 0.0, -0.4, 0.9, 0.0}
	dt := 1e-5
	for k := 0; k < 400; k++ {
		e := wave[k%len(wave)] * mat.Ec
		pa := a.Step(e, dt)
		pb := b.Step(e, dt)
		if math.Abs(pa-pb) > 1e-12 {
			t.Fatalf("seed reproducibility mismatch at step %d: pa=%.15f pb=%.15f", k, pa, pb)
		}
	}
}
