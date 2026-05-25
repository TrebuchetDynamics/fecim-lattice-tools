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

func TestNewPolydomainEnsembleDefaultsInvalidSigmaFraction(t *testing.T) {
	mat := LiteratureSuperlattice()
	template := NewLKSolver()
	template.ConfigureFromMaterial(mat)

	got := NewPolydomainEnsemble(template, mat, 8, math.NaN(), 1009)
	want := NewPolydomainEnsemble(template, mat, 8, defaultPolydomainSigmaFrac, 1009)
	if got == nil || want == nil {
		t.Fatalf("expected ensembles for invalid/default sigma: got=%#v want=%#v", got, want)
	}
	if len(got.EcFactor) != len(want.EcFactor) {
		t.Fatalf("EcFactor length = %d, want %d", len(got.EcFactor), len(want.EcFactor))
	}
	for i := range got.EcFactor {
		if math.IsNaN(got.EcFactor[i]) || math.IsInf(got.EcFactor[i], 0) {
			t.Fatalf("invalid sigma produced non-finite EcFactor[%d]=%g", i, got.EcFactor[i])
		}
		if got.EcFactor[i] != want.EcFactor[i] {
			t.Fatalf("invalid sigma EcFactor[%d]=%g, want default sigma factor %g", i, got.EcFactor[i], want.EcFactor[i])
		}
	}
}

func TestNewPolydomainEnsembleRejectsUnrepresentableLandauScalingMaterial(t *testing.T) {
	template := NewLKSolver()
	mat := &HZOMaterial{
		Name:        "invalid-landau-scaling",
		Ec:          1e8,
		BetaLandau:  1e60,
		GammaLandau: 1e60,
		Pr:          1e59,
		Ps:          1e59,
	}

	got := NewPolydomainEnsemble(template, mat, 4, defaultPolydomainSigmaFrac, 909)
	if got != nil {
		t.Fatalf("NewPolydomainEnsemble with unrepresentable Landau scaling material = %#v, want nil", got)
	}
}

func TestLKSolverEnableEnsembleRejectsInvalidMaterialWithoutMutatingState(t *testing.T) {
	control, solver, mat := matchingDeterministicEnsembleSolvers()

	invalid := *mat
	invalid.Ec = math.NaN()
	solver.EnableEnsemble(8, &invalid, 99)

	assertMatchesControlEnsemble(t, control, solver, mat, "after rejected EnableEnsemble material")
}

func TestLKSolverConfigureFromMaterialRejectsInvalidEnsembleMaterialWithoutMutatingState(t *testing.T) {
	control, solver, mat := matchingDeterministicEnsembleSolvers()

	invalid := *mat
	invalid.Ec = math.NaN()
	solver.ConfigureFromMaterial(&invalid)

	assertMatchesControlEnsemble(t, control, solver, mat, "after rejected ConfigureFromMaterial ensemble material")
}

func TestLKSolverConfigureFromMaterialRecoversInvalidPublicStateBeforeEnsembleRebuild(t *testing.T) {
	mat := LiteratureSuperlattice()
	solver := NewLKSolver()
	solver.ConfigureFromMaterial(mat)
	solver.EnableEnsemble(4, mat, 808)
	solver.P = math.MaxFloat64
	solver.PMax = math.MaxFloat64

	solver.ConfigureFromMaterial(mat)

	if !isRepresentableLKPolarization(solver.GetState()) {
		t.Fatalf("ConfigureFromMaterial left unrepresentable ensemble solver polarization %g", solver.GetState())
	}
	if !isValidLKRuntimePMax(solver.PMax) {
		t.Fatalf("ConfigureFromMaterial left invalid ensemble PMax %g", solver.PMax)
	}
	if solver.polydomain == nil || solver.polydomain.DomainCount() != 4 {
		t.Fatalf("ConfigureFromMaterial did not rebuild expected ensemble: %#v", solver.polydomain)
	}
	for i, domain := range solver.polydomain.Domains {
		if domain == nil {
			t.Fatalf("ConfigureFromMaterial left nil domain %d", i)
		}
		if domain.GetState() != solver.GetState() {
			t.Fatalf("ConfigureFromMaterial domain %d polarization = %g, want broadcast solver state %g", i, domain.GetState(), solver.GetState())
		}
	}
}

func TestLKSolverEnableEnsembleRecoversInvalidPublicStateBeforeBroadcast(t *testing.T) {
	mat := LiteratureSuperlattice()
	solver := NewLKSolver()
	solver.ConfigureFromMaterial(mat)
	solver.P = math.MaxFloat64
	solver.PMax = math.MaxFloat64

	solver.EnableEnsemble(4, mat, 707)

	if !isRepresentableLKPolarization(solver.GetState()) {
		t.Fatalf("EnableEnsemble left unrepresentable solver polarization %g", solver.GetState())
	}
	if !isValidLKRuntimePMax(solver.PMax) {
		t.Fatalf("EnableEnsemble left invalid PMax %g", solver.PMax)
	}
	if solver.polydomain == nil || solver.polydomain.DomainCount() != 4 {
		t.Fatalf("EnableEnsemble did not install expected ensemble: %#v", solver.polydomain)
	}
	for i, domain := range solver.polydomain.Domains {
		if domain == nil {
			t.Fatalf("EnableEnsemble left nil domain %d", i)
		}
		if domain.GetState() != solver.GetState() {
			t.Fatalf("EnableEnsemble domain %d polarization = %g, want broadcast solver state %g", i, domain.GetState(), solver.GetState())
		}
	}
}

func TestLKSolverEnableEnsembleRejectsUnrepresentableDomainCountWithoutMutatingState(t *testing.T) {
	control, solver, mat := matchingDeterministicEnsembleSolvers()

	maxInt := int(^uint(0) >> 1)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("EnableEnsemble with unrepresentable domain count panicked: %v", r)
		}
	}()
	solver.EnableEnsemble(maxInt, mat, 101)

	assertMatchesControlEnsemble(t, control, solver, mat, "after rejected EnableEnsemble domain count")
}

func TestPolydomainEnsembleStepRejectsInvalidPublicState(t *testing.T) {
	mat := LiteratureSuperlattice()
	template := NewLKSolver()
	template.ConfigureFromMaterial(mat)
	template.EnableNoise = false
	template.UseNLS = false

	ensemble := NewPolydomainEnsemble(template, mat, 4, defaultPolydomainSigmaFrac, 202)
	if ensemble == nil {
		t.Fatal("expected test ensemble")
	}
	ensemble.SetState(-math.Abs(mat.Pr))
	for _, domain := range ensemble.Domains {
		domain.EnableNoise = false
		domain.UseNLS = false
	}

	// EcFactor, Imprint, and Domains are exported public state. Invalid caller
	// mutations should be rejected locally by the ensemble step seam rather than
	// panicking or returning NaN/Inf.
	ensemble.EcFactor[0] = math.NaN()
	ensemble.EcFactor[1] = 0
	ensemble.EcFactor[2] = math.Inf(1)
	ensemble.Imprint[0] = math.NaN()
	ensemble.Imprint[1] = math.Inf(1)
	ensemble.Domains[3] = nil

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("PolydomainEnsemble.Step with invalid public state panicked: %v", r)
		}
	}()

	got := ensemble.Step(nil, 0.25*mat.Ec, 1e-9)
	if math.IsNaN(got) || math.IsInf(got, 0) {
		t.Fatalf("PolydomainEnsemble.Step with invalid public state returned non-finite polarization %g", got)
	}
}

func TestPolydomainEnsembleStepRejectsInvalidInputBeforeTemplatePropagation(t *testing.T) {
	mat := LiteratureSuperlattice()
	base := NewLKSolver()
	base.ConfigureFromMaterial(mat)
	base.EnableNoise = false
	base.UseNLS = false

	ensemble := NewPolydomainEnsemble(base, mat, 3, defaultPolydomainSigmaFrac, 808)
	if ensemble == nil {
		t.Fatal("expected test ensemble")
	}
	states := []float64{0.12, -0.06, 0.03}
	for i, domain := range ensemble.Domains {
		domain.EnableNoise = false
		domain.UseNLS = false
		domain.Temperature = 300
		domain.Stress = 1e9
		domain.Time = float64(i + 1)
		domain.SetState(states[i])
	}

	template := NewLKSolver()
	template.ConfigureFromMaterial(mat)
	template.EnableNoise = true
	template.UseNLS = true
	template.Temperature = 325
	template.Stress = 1.25e9

	got := ensemble.Step(template, math.MaxFloat64, 1e-12)
	want := (states[0] + states[1] + states[2]) / 3
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("Step with invalid input = %g, want preserved mean %g", got, want)
	}
	for i, domain := range ensemble.Domains {
		if domain.GetState() != states[i] {
			t.Fatalf("domain %d state mutated to %g, want %g", i, domain.GetState(), states[i])
		}
		if domain.Time != float64(i+1) {
			t.Fatalf("domain %d time mutated to %g, want %g", i, domain.Time, float64(i+1))
		}
		if domain.EnableNoise || domain.UseNLS {
			t.Fatalf("domain %d template flags propagated before invalid input rejection", i)
		}
		if domain.Temperature != 300 || domain.Stress != 1e9 {
			t.Fatalf("domain %d template runtime state propagated before invalid input rejection: T=%g stress=%g", i, domain.Temperature, domain.Stress)
		}
	}
}

func TestPolydomainEnsembleStepRejectsFiniteOverflowingDomainPolarization(t *testing.T) {
	mat := LiteratureSuperlattice()
	template := NewLKSolver()
	template.ConfigureFromMaterial(mat)
	template.EnableNoise = false
	template.UseNLS = false

	ensemble := NewPolydomainEnsemble(template, mat, 3, defaultPolydomainSigmaFrac, 404)
	if ensemble == nil {
		t.Fatal("expected test ensemble")
	}
	for _, domain := range ensemble.Domains {
		domain.EnableNoise = false
		domain.UseNLS = false
		domain.PMax = 0
		domain.P = math.MaxFloat64
	}

	got := ensemble.Step(template, 0, 1e-12)
	if math.IsNaN(got) || math.IsInf(got, 0) {
		t.Fatalf("PolydomainEnsemble.Step with finite overflowing domain polarization returned non-finite value %g", got)
	}
}

func TestPolydomainEnsembleSetStateAndRemanentSpreadRejectInvalidPublicState(t *testing.T) {
	mat := LiteratureSuperlattice()
	template := NewLKSolver()
	template.ConfigureFromMaterial(mat)
	ensemble := NewPolydomainEnsemble(template, mat, 4, defaultPolydomainSigmaFrac, 303)
	if ensemble == nil {
		t.Fatal("expected test ensemble")
	}

	for _, domain := range ensemble.Domains {
		domain.Time = 7
	}
	ensemble.Domains[0] = nil

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SetState/RemanentSpread with invalid public state panicked: %v", r)
		}
	}()

	ensemble.SetState(0.12)
	for i, domain := range ensemble.Domains {
		if domain == nil {
			continue
		}
		if domain.GetState() != 0.12 {
			t.Fatalf("SetState skipped valid domain %d: got %g", i, domain.GetState())
		}
		if domain.Time != 0 {
			t.Fatalf("SetState did not reset valid domain %d time: got %g", i, domain.Time)
		}
	}

	ensemble.Domains[1].P = math.NaN()
	ensemble.Domains[2].P = math.Inf(1)
	spread := ensemble.RemanentSpread(mat.Ps)
	if math.IsNaN(spread) || math.IsInf(spread, 0) {
		t.Fatalf("RemanentSpread with invalid public state returned non-finite value %g", spread)
	}
	if got := ensemble.RemanentSpread(math.NaN()); got != 0 {
		t.Fatalf("RemanentSpread with invalid Ps = %g, want 0", got)
	}
}

func TestPolydomainEnsembleSetStateRejectsFiniteUnrepresentableInputWithoutMutation(t *testing.T) {
	mat := LiteratureSuperlattice()
	template := NewLKSolver()
	template.ConfigureFromMaterial(mat)
	ensemble := NewPolydomainEnsemble(template, mat, 3, defaultPolydomainSigmaFrac, 1001)
	if ensemble == nil {
		t.Fatal("expected test ensemble")
	}

	const state = 0.12
	ensemble.SetState(state)
	beforeState := make([]float64, len(ensemble.Domains))
	beforeTime := make([]float64, len(ensemble.Domains))
	for i, domain := range ensemble.Domains {
		domain.Time = float64(i + 1)
		beforeState[i] = domain.GetState()
		beforeTime[i] = domain.Time
	}

	ensemble.SetState(math.MaxFloat64)
	for i, domain := range ensemble.Domains {
		if domain.GetState() != beforeState[i] {
			t.Fatalf("SetState with finite unrepresentable input mutated domain %d state to %g, want %g", i, domain.GetState(), beforeState[i])
		}
		if domain.Time != beforeTime[i] {
			t.Fatalf("SetState with finite unrepresentable input mutated domain %d time to %g, want %g", i, domain.Time, beforeTime[i])
		}
	}
}

func TestPolydomainEnsembleRemanentSpreadRejectsFiniteOverflowingDomainPolarization(t *testing.T) {
	mat := LiteratureSuperlattice()
	template := NewLKSolver()
	template.ConfigureFromMaterial(mat)
	ensemble := NewPolydomainEnsemble(template, mat, 3, defaultPolydomainSigmaFrac, 505)
	if ensemble == nil {
		t.Fatal("expected test ensemble")
	}

	const validPolarization = 0.12
	ensemble.Domains[0].P = validPolarization
	ensemble.Domains[1].P = math.MaxFloat64
	ensemble.Domains[2].P = math.MaxFloat64

	got := ensemble.RemanentSpread(mat.Ps)
	want := math.Abs(validPolarization / mat.Ps)
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("RemanentSpread with finite overflowing domain polarization = %g, want valid-domain spread %g", got, want)
	}
}

func TestPolydomainEnsembleRejectsInvalidTemplateStatePropagation(t *testing.T) {
	mat := LiteratureSuperlattice()
	template := NewLKSolver()
	template.ConfigureFromMaterial(mat)
	template.EnableNoise = true
	template.UseNLS = true
	template.Temperature = math.MaxFloat64
	template.Stress = math.MaxFloat64
	template.PMax = math.MaxFloat64

	ensemble := NewPolydomainEnsemble(template, mat, 3, defaultPolydomainSigmaFrac, 606)
	if ensemble == nil {
		t.Fatal("expected test ensemble")
	}
	for i, domain := range ensemble.Domains {
		if math.IsNaN(domain.Temperature) || math.IsInf(domain.Temperature, 0) || domain.Temperature == math.MaxFloat64 {
			t.Fatalf("domain %d inherited invalid template Temperature %g", i, domain.Temperature)
		}
		if math.IsNaN(domain.Stress) || math.IsInf(domain.Stress, 0) || domain.Stress == math.MaxFloat64 {
			t.Fatalf("domain %d inherited invalid template Stress %g", i, domain.Stress)
		}
		if domain.PMax == math.MaxFloat64 || math.IsInf(domain.PMax, 0) || math.IsNaN(domain.PMax) {
			t.Fatalf("domain %d inherited invalid template PMax %g", i, domain.PMax)
		}
	}

	validTemperature := ensemble.Domains[0].Temperature
	validStress := ensemble.Domains[0].Stress
	template.Temperature = math.MaxFloat64
	template.Stress = -math.MaxFloat64
	ensemble.Step(template, 0.2*mat.Ec, 1e-9)
	for i, domain := range ensemble.Domains {
		if domain.Temperature != validTemperature {
			t.Fatalf("domain %d runtime Temperature = %g after invalid template step, want preserved %g", i, domain.Temperature, validTemperature)
		}
		if domain.Stress != validStress {
			t.Fatalf("domain %d runtime Stress = %g after invalid template step, want preserved %g", i, domain.Stress, validStress)
		}
	}
}

func TestPolydomainEnsembleRejectsRuntimePropagationWhenDomainConstantsInvalid(t *testing.T) {
	mat := LiteratureSuperlattice()
	base := NewLKSolver()
	base.ConfigureFromMaterial(mat)
	base.EnableNoise = false
	base.UseNLS = false

	ensemble := NewPolydomainEnsemble(base, mat, 3, defaultPolydomainSigmaFrac, 707)
	if ensemble == nil {
		t.Fatal("expected test ensemble")
	}
	domain := ensemble.Domains[0]
	preservedTemperature := domain.Temperature
	preservedStress := domain.Stress
	domain.CurieConst = math.NaN()

	template := NewLKSolver()
	template.ConfigureFromMaterial(mat)
	template.EnableNoise = false
	template.UseNLS = false
	template.Temperature = preservedTemperature + 25
	template.Stress = preservedStress + 1e8

	ensemble.Step(template, 0, 1e-12)
	if domain.Temperature != preservedTemperature {
		t.Fatalf("domain with invalid CurieConst accepted template Temperature %g, want preserved %g", domain.Temperature, preservedTemperature)
	}
	if domain.Stress != preservedStress {
		t.Fatalf("domain with invalid CurieConst accepted template Stress %g, want preserved %g", domain.Stress, preservedStress)
	}
}

func matchingDeterministicEnsembleSolvers() (*LKSolver, *LKSolver, *HZOMaterial) {
	mat := LiteratureSuperlattice()
	mk := func() *LKSolver {
		s := NewLKSolver()
		s.ConfigureFromMaterial(mat)
		s.EnableNoise = false
		s.UseNLS = false
		s.EnableEnsemble(16, mat, 77)
		s.SetState(-math.Abs(mat.Pr))
		return s
	}
	return mk(), mk(), mat
}

func assertMatchesControlEnsemble(t *testing.T, control, solver *LKSolver, mat *HZOMaterial, context string) {
	t.Helper()
	wave := []float64{0.5 * mat.Ec, 0, -0.4 * mat.Ec, 0.2 * mat.Ec}
	const dt = 1e-9
	for i, field := range wave {
		want := control.Step(field, dt)
		got := solver.Step(field, dt)
		if math.IsNaN(got) || math.IsInf(got, 0) {
			t.Fatalf("step %d %s returned non-finite polarization %g", i, context, got)
		}
		if got != want {
			t.Fatalf("step %d %s = %g, want preserved ensemble result %g", i, context, got, want)
		}
		if solver.Time != control.Time {
			t.Fatalf("step %d %s left time %g, want %g", i, context, solver.Time, control.Time)
		}
	}
}
