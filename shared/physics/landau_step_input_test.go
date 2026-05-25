package physics

import (
	"math"
	"math/rand"
	"testing"
)

func TestLKSolverStepRejectsInvalidInputsWithoutMutatingState(t *testing.T) {
	cases := []struct {
		name string
		E    float64
		dt   float64
	}{
		{name: "nan field", E: math.NaN(), dt: 1e-12},
		{name: "positive infinite field", E: math.Inf(1), dt: 1e-12},
		{name: "negative infinite field", E: math.Inf(-1), dt: 1e-12},
		{name: "finite unrepresentable field", E: math.MaxFloat64, dt: 1e-12},
		{name: "nan timestep", E: 1e8, dt: math.NaN()},
		{name: "positive infinite timestep", E: 1e8, dt: math.Inf(1)},
		{name: "negative timestep", E: 1e8, dt: -1e-12},
		{name: "unrepresentable finite timestep", E: 1e8, dt: math.MaxFloat64},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			control := newDeterministicLKSolverForInputValidation()
			solver := newDeterministicLKSolverForInputValidation()

			control.Step(1e8, 1e-12)
			solver.Step(1e8, 1e-12)

			wantP := solver.GetState()
			wantTime := solver.Time
			got := solver.Step(tc.E, tc.dt)

			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Fatalf("Step(%g, %g) returned non-finite polarization %g", tc.E, tc.dt, got)
			}
			if got != wantP {
				t.Fatalf("Step(%g, %g) polarization = %g, want current polarization %g", tc.E, tc.dt, got, wantP)
			}
			if solver.GetState() != wantP {
				t.Fatalf("Step(%g, %g) mutated solver polarization to %g, want %g", tc.E, tc.dt, solver.GetState(), wantP)
			}
			if solver.Time != wantTime {
				t.Fatalf("Step(%g, %g) mutated solver time to %g, want %g", tc.E, tc.dt, solver.Time, wantTime)
			}

			gotAfter := solver.Step(2e8, 1e-12)
			wantAfter := control.Step(2e8, 1e-12)
			if gotAfter != wantAfter {
				t.Fatalf("valid step after rejected input = %g, want control result %g", gotAfter, wantAfter)
			}
			if solver.Time != control.Time {
				t.Fatalf("valid step after rejected input left time %g, want control time %g", solver.Time, control.Time)
			}
		})
	}
}

func TestLKSolverPublicStateMethodsHandleNilReceiver(t *testing.T) {
	t.Run("get state", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("GetState on nil receiver panicked: %v", r)
			}
		}()
		var solver *LKSolver
		if got := solver.GetState(); got != 0 {
			t.Fatalf("GetState on nil receiver = %g, want 0", got)
		}
	})

	t.Run("set state", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("SetState on nil receiver panicked: %v", r)
			}
		}()
		var solver *LKSolver
		solver.SetState(0.12)
	})

	t.Run("check timestep", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("CheckTimestep on nil receiver panicked: %v", r)
			}
		}()
		var solver *LKSolver
		if warning := solver.CheckTimestep(1e8, 1e-12); warning != nil {
			t.Fatalf("CheckTimestep on nil receiver = %+v, want nil", warning)
		}
	})
}

func TestLKSolverCheckTimestepRejectsInvalidInputsWithoutNonfiniteWarning(t *testing.T) {
	cases := []struct {
		name string
		E    float64
		dt   float64
	}{
		{name: "nan field", E: math.NaN(), dt: 1e-12},
		{name: "positive infinite field", E: math.Inf(1), dt: 1e-12},
		{name: "finite unrepresentable field", E: math.MaxFloat64, dt: 1e-12},
		{name: "nan timestep", E: 1e8, dt: math.NaN()},
		{name: "negative timestep", E: 1e8, dt: -1e-12},
		{name: "unrepresentable finite timestep", E: 1e8, dt: math.MaxFloat64},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			solver := newDeterministicLKSolverForInputValidation()
			solver.ConfigureFromMaterial(LiteratureSuperlattice())
			beforeP := solver.GetState()
			beforeTime := solver.Time

			warning := solver.CheckTimestep(tc.E, tc.dt)
			if warning != nil {
				if math.IsNaN(warning.StepRatio) || math.IsInf(warning.StepRatio, 0) || math.IsNaN(warning.Recommended) || math.IsInf(warning.Recommended, 0) {
					t.Fatalf("CheckTimestep(%g, %g) returned non-finite warning: %+v", tc.E, tc.dt, warning)
				}
				t.Fatalf("CheckTimestep(%g, %g) returned warning %+v, want nil for invalid input", tc.E, tc.dt, warning)
			}
			if solver.GetState() != beforeP {
				t.Fatalf("CheckTimestep(%g, %g) mutated polarization to %g, want %g", tc.E, tc.dt, solver.GetState(), beforeP)
			}
			if solver.Time != beforeTime {
				t.Fatalf("CheckTimestep(%g, %g) mutated time to %g, want %g", tc.E, tc.dt, solver.Time, beforeTime)
			}
		})
	}
}

func TestLKSolverStepRecoversInvalidPublicTime(t *testing.T) {
	cases := []struct {
		name string
		time float64
	}{
		{name: "nan time", time: math.NaN()},
		{name: "positive infinite time", time: math.Inf(1)},
		{name: "negative time", time: -1},
		{name: "finite unrepresentable time", time: math.MaxFloat64},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			control := newDeterministicLKSolverForInputValidation()
			solver := newDeterministicLKSolverForInputValidation()
			control.ConfigureFromMaterial(LiteratureSuperlattice())
			solver.ConfigureFromMaterial(LiteratureSuperlattice())
			solver.Time = tc.time

			want := control.Step(1e8, 1e-12)
			got := solver.Step(1e8, 1e-12)
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Fatalf("Step with invalid public Time returned non-finite polarization %g", got)
			}
			if got != want {
				t.Fatalf("Step with invalid public Time = %g, want recovered control result %g", got, want)
			}
			if solver.Time != control.Time {
				t.Fatalf("Step with invalid public Time left Time %g, want recovered control time %g", solver.Time, control.Time)
			}
		})
	}
}

func TestLKSolverStepRecoversInvalidPublicRho(t *testing.T) {
	cases := []struct {
		name string
		rho  float64
	}{
		{name: "nan rho", rho: math.NaN()},
		{name: "positive infinite rho", rho: math.Inf(1)},
		{name: "finite unrepresentable rho", rho: math.MaxFloat64},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			control := newLinearLKSolverForRuntimeCoefficientValidation()
			solver := newLinearLKSolverForRuntimeCoefficientValidation()
			solver.Rho = tc.rho

			want := control.Step(1e8, 1e-12)
			got := solver.Step(1e8, 1e-12)
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Fatalf("Step with invalid public Rho returned non-finite polarization %g", got)
			}
			if got != want {
				t.Fatalf("Step with invalid public Rho = %g, want default-viscosity result %g", got, want)
			}
		})
	}
}

func TestLKSolverStepNeutralizesInvalidPublicLandauForceCoefficients(t *testing.T) {
	cases := []struct {
		name    string
		corrupt func(*LKSolver)
	}{
		{name: "nan alpha", corrupt: func(s *LKSolver) { s.Alpha = math.NaN() }},
		{name: "nan beta", corrupt: func(s *LKSolver) { s.Beta = math.NaN() }},
		{name: "positive infinite gamma", corrupt: func(s *LKSolver) { s.Gamma = math.Inf(1) }},
		{name: "nan depolarization", corrupt: func(s *LKSolver) { s.K_dep = math.NaN() }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			control := newLinearLKSolverForRuntimeCoefficientValidation()
			solver := newLinearLKSolverForRuntimeCoefficientValidation()
			tc.corrupt(solver)

			want := control.Step(1e8, 1e-12)
			got := solver.Step(1e8, 1e-12)
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Fatalf("Step with invalid public Landau-force coefficient returned non-finite polarization %g", got)
			}
			if got != want {
				t.Fatalf("Step with invalid public Landau-force coefficient = %g, want neutral-coefficient result %g", got, want)
			}
		})
	}
}

func TestLKSolverStepNeutralizesFiniteUnrepresentablePublicLandauForceTerms(t *testing.T) {
	cases := []struct {
		name    string
		corrupt func(*LKSolver)
	}{
		{name: "alpha force overflows rate", corrupt: func(s *LKSolver) { s.Alpha = math.MaxFloat64 }},
		{name: "beta force overflows rate", corrupt: func(s *LKSolver) { s.Beta = math.MaxFloat64 }},
		{name: "gamma force overflows rate", corrupt: func(s *LKSolver) { s.Gamma = math.MaxFloat64 }},
		{name: "depolarization force overflows rate", corrupt: func(s *LKSolver) { s.K_dep = math.MaxFloat64 }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			control := newLinearLKSolverForRuntimeCoefficientValidation()
			solver := newLinearLKSolverForRuntimeCoefficientValidation()
			control.SetState(0.5)
			solver.SetState(0.5)
			tc.corrupt(solver)

			want := control.Step(1e8, 1e-12)
			got := solver.Step(1e8, 1e-12)
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Fatalf("Step with finite unrepresentable public Landau-force term returned non-finite polarization %g", got)
			}
			if got != want {
				t.Fatalf("Step with finite unrepresentable public Landau-force term = %g, want neutral-coefficient result %g", got, want)
			}
		})
	}
}

func TestLKSolverStepPreservesZeroPMaxAsClampDisabled(t *testing.T) {
	solver := NewLKSolver()
	solver.UseNLS = false
	solver.EnableNoise = false
	solver.UseEffectiveViscosity = false
	solver.UseMaterialAlpha = true
	solver.Alpha = 0
	solver.Beta = 0
	solver.Gamma = 0
	solver.K_dep = 0
	solver.Rho = 1
	solver.PMax = 0
	solver.SetState(0.1)

	got := solver.Step(10, 1)
	const want = 10.1
	if got != want {
		t.Fatalf("Step with PMax=0 clamp disabled = %g, want unclamped %g", got, want)
	}
	if solver.PMax != 0 {
		t.Fatalf("Step with PMax=0 mutated PMax to %g", solver.PMax)
	}
}

func TestLKSolverSetStateRejectsFiniteUnrepresentablePolarizationWithoutMutation(t *testing.T) {
	solver := NewLKSolver()
	solver.SetState(0.12)
	before := solver.GetState()

	solver.SetState(math.MaxFloat64)
	if solver.GetState() != before {
		t.Fatalf("SetState with finite unrepresentable polarization mutated state to %g, want %g", solver.GetState(), before)
	}

	mat := LiteratureSuperlattice()
	ensembleSolver := NewLKSolver()
	ensembleSolver.ConfigureFromMaterial(mat)
	ensembleSolver.EnableNoise = false
	ensembleSolver.UseNLS = false
	ensembleSolver.EnableEnsemble(4, mat, 909)
	ensembleSolver.SetState(0.12)
	beforeSolver := ensembleSolver.GetState()
	beforeDomains := make([]float64, len(ensembleSolver.polydomain.Domains))
	for i, domain := range ensembleSolver.polydomain.Domains {
		beforeDomains[i] = domain.GetState()
	}

	ensembleSolver.SetState(math.MaxFloat64)
	if ensembleSolver.GetState() != beforeSolver {
		t.Fatalf("ensemble SetState with finite unrepresentable polarization mutated solver state to %g, want %g", ensembleSolver.GetState(), beforeSolver)
	}
	for i, domain := range ensembleSolver.polydomain.Domains {
		if domain.GetState() != beforeDomains[i] {
			t.Fatalf("ensemble SetState with finite unrepresentable polarization mutated domain %d to %g, want %g", i, domain.GetState(), beforeDomains[i])
		}
	}
}

func TestLKSolverStepRecoversInvalidPublicPMaxWhenStateInvalid(t *testing.T) {
	cases := []struct {
		name string
		pmax float64
	}{
		{name: "nan pmax", pmax: math.NaN()},
		{name: "positive infinite pmax", pmax: math.Inf(1)},
		{name: "negative pmax", pmax: -1},
		{name: "finite unrepresentable pmax", pmax: math.MaxFloat64},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			control := newDeterministicLKSolverForInputValidation()
			solver := newDeterministicLKSolverForInputValidation()
			mat := LiteratureSuperlattice()
			control.ConfigureFromMaterial(mat)
			solver.ConfigureFromMaterial(mat)

			solver.P = math.NaN()
			solver.PMax = tc.pmax

			_ = control.Step(1e8, 1e-12)
			got := solver.Step(1e8, 1e-12)
			if math.IsNaN(got) || math.IsInf(got, 0) || !isRepresentableLKPolarization(got) {
				t.Fatalf("Step with invalid public PMax returned unrepresentable polarization %g", got)
			}
			if solver.P != got {
				t.Fatalf("Step with invalid public PMax left state %g, want returned state %g", solver.P, got)
			}
			if !isValidLKRuntimePMax(solver.PMax) {
				t.Fatalf("Step with invalid public PMax left invalid PMax %g", solver.PMax)
			}
		})
	}
}

func TestLKSolverConfigureFromMaterialRejectsInvalidNumbersWithoutMutatingState(t *testing.T) {
	cases := []struct {
		name string
		mat  HZOMaterial
	}{
		{
			name: "non-finite material parameters",
			mat: HZOMaterial{
				BetaLandau:          math.NaN(),
				GammaLandau:         math.Inf(1),
				RhoViscosity:        math.NaN(),
				Q12:                 math.Inf(-1),
				StressGPa:           math.NaN(),
				K_dep:               math.Inf(1),
				Thickness:           math.Inf(1),
				Area:                math.Inf(1),
				CurieTemp:           math.Inf(1),
				CurieConst:          math.Inf(1),
				SeriesResistanceOhm: math.Inf(1),
				Tau0NLS:             math.Inf(1),
				EaNLS:               math.Inf(1),
				NLSSigma:            math.Inf(1),
				Pr:                  math.NaN(),
				Ps:                  math.Inf(1),
				Ec:                  math.Inf(1),
			},
		},
		{
			name: "finite values that overflow derived LK arithmetic",
			mat: HZOMaterial{
				StressGPa: math.MaxFloat64,
				Pr:        math.MaxFloat64,
				Ps:        math.MaxFloat64,
				Ec:        math.MaxFloat64,
			},
		},
		{
			name: "finite material scalars beyond runtime bounds",
			mat: HZOMaterial{
				BetaLandau:          math.MaxFloat64,
				GammaLandau:         math.MaxFloat64,
				RhoViscosity:        math.MaxFloat64,
				Q12:                 math.MaxFloat64,
				K_dep:               math.MaxFloat64,
				Thickness:           math.MaxFloat64,
				Area:                math.MaxFloat64,
				CurieTemp:           math.MaxFloat64,
				CurieConst:          math.MaxFloat64,
				SeriesResistanceOhm: math.MaxFloat64,
				Tau0NLS:             math.MaxFloat64,
				EaNLS:               math.MaxFloat64,
				NLSSigma:            math.MaxFloat64,
			},
		},
		{
			name: "finite coercive estimate overflow",
			mat: HZOMaterial{
				BetaLandau:  1e60,
				GammaLandau: 1e60,
				Pr:          1e59,
				Ps:          1e59,
				Ec:          1e8,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			solver := newDeterministicLKSolverForInputValidation()
			before := snapshotLKConfiguration(solver)

			solver.ConfigureFromMaterial(&tc.mat)

			after := snapshotLKConfiguration(solver)
			if after != before {
				t.Fatalf("invalid material mutated LK solver configuration\nbefore: %+v\nafter:  %+v", before, after)
			}
			got := solver.Step(1e8, 1e-12)
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Fatalf("valid step after rejected material returned non-finite polarization %g", got)
			}
		})
	}
}

func TestLKSolverUpdateParamsRejectsInvalidRuntimeThermodynamicInputs(t *testing.T) {
	cases := []struct {
		name        string
		temperature float64
		stress      float64
	}{
		{name: "nan temperature", temperature: math.NaN(), stress: 1e9},
		{name: "infinite stress", temperature: 300, stress: math.Inf(1)},
		{name: "finite overflowing temperature", temperature: math.MaxFloat64, stress: 1e9},
		{name: "finite overflowing stress", temperature: 300, stress: math.MaxFloat64},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			solver := newDeterministicLKSolverForInputValidation()
			solver.ConfigureFromMaterial(LiteratureSuperlattice())
			solver.UseMaterialAlpha = false
			solver.Temperature = 300
			solver.Stress = 1e9
			solver.UpdateParams()
			wantAlpha := solver.Alpha

			solver.Temperature = tc.temperature
			solver.Stress = tc.stress
			solver.UpdateParams()

			if math.IsNaN(solver.Alpha) || math.IsInf(solver.Alpha, 0) {
				t.Fatalf("UpdateParams with invalid runtime inputs produced non-finite Alpha %g", solver.Alpha)
			}
			if solver.Alpha != wantAlpha {
				t.Fatalf("UpdateParams with invalid runtime inputs mutated Alpha to %g, want preserved %g", solver.Alpha, wantAlpha)
			}
		})
	}
}

func TestLKSolverStepTreatsInvalidThermalNoiseAsZero(t *testing.T) {
	const dt = 1e-12
	mat := LiteratureSuperlattice()

	baseline := newDeterministicLKSolverForInputValidation()
	baseline.ConfigureFromMaterial(mat)
	baseline.EnableNoise = false
	baseline.Temperature = math.MaxFloat64
	baseline.UseNLS = false

	noisy := newDeterministicLKSolverForInputValidation()
	noisy.ConfigureFromMaterial(mat)
	noisy.EnableNoise = true
	noisy.rng = rand.New(rand.NewSource(1))
	noisy.Temperature = math.MaxFloat64
	noisy.UseNLS = false

	want := baseline.Step(0, dt)
	got := noisy.Step(0, dt)
	if math.IsNaN(got) || math.IsInf(got, 0) {
		t.Fatalf("Step with invalid thermal noise returned non-finite polarization %g", got)
	}
	if got != want {
		t.Fatalf("Step with invalid thermal noise = %g, want zero-noise result %g", got, want)
	}
	if noisy.Time != baseline.Time {
		t.Fatalf("Step with invalid thermal noise left time %g, want %g", noisy.Time, baseline.Time)
	}
}

func newDeterministicLKSolverForInputValidation() *LKSolver {
	s := NewLKSolver()
	s.EnableNoise = false
	return s
}

func newLinearLKSolverForRuntimeCoefficientValidation() *LKSolver {
	s := NewLKSolver()
	s.UseNLS = false
	s.EnableNoise = false
	s.UseEffectiveViscosity = false
	s.UseMaterialAlpha = true
	s.Alpha = 0
	s.Beta = 0
	s.Gamma = 0
	s.K_dep = 0
	s.Rho = defaultLKViscosity
	s.PMax = 0
	s.SetState(0.1)
	return s
}

type lkConfigurationSnapshot struct {
	Beta             float64
	Gamma            float64
	Rho              float64
	Q12              float64
	Stress           float64
	KDep             float64
	Thickness        float64
	Area             float64
	CurieTemp        float64
	CurieConst       float64
	SeriesResistance float64
	TauInf           float64
	ActivationField  float64
	NLSSigma         float64
	P                float64
	PMax             float64
	Alpha            float64
	UseMaterialAlpha bool
}

func snapshotLKConfiguration(s *LKSolver) lkConfigurationSnapshot {
	return lkConfigurationSnapshot{
		Beta:             s.Beta,
		Gamma:            s.Gamma,
		Rho:              s.Rho,
		Q12:              s.Q12,
		Stress:           s.Stress,
		KDep:             s.K_dep,
		Thickness:        s.Thickness,
		Area:             s.Area,
		CurieTemp:        s.CurieTemp,
		CurieConst:       s.CurieConst,
		SeriesResistance: s.SeriesResistance,
		TauInf:           s.TauInf,
		ActivationField:  s.ActivationField,
		NLSSigma:         s.NLSSigma,
		P:                s.P,
		PMax:             s.PMax,
		Alpha:            s.Alpha,
		UseMaterialAlpha: s.UseMaterialAlpha,
	}
}
