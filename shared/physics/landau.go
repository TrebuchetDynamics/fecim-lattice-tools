package physics

import (
	"hash/fnv"
	"math"
	"math/rand"

	"fecim-lattice-tools/shared/logging"
)

func estimateLandauEc(alpha, beta, gamma, pr float64) float64 {
	if pr <= 0 {
		return 0
	}
	// Coarse grid search: Ec ~ max |dG/dP| over (-Pr,Pr).
	// With first-order Landau (beta<0,gamma>0), the switching saddle typically
	// occurs on the opposite-polarity branch (e.g. negative P when writing up),
	// so using |dG/dP| is robust.
	const n = 8000
	maxAbs := 0.0
	for i := 1; i < n; i++ {
		p := -pr + 2*pr*float64(i)/float64(n)
		p2 := p * p
		p3 := p2 * p
		p5 := p3 * p2
		dg := (2 * alpha * p) + (4 * beta * p3) + (6 * gamma * p5)
		if v := math.Abs(dg); v > maxAbs {
			maxAbs = v
		}
	}
	return maxAbs
}

// Package-level logger for Landau-Khalatnikov solver diagnostics.
var lkLog *logging.Logger

// LKSolver implements the First-Order Landau-Khalatnikov dynamics
// for ferroelectric polarization evolution.
type LKSolver struct {
	// Static Material Properties (from calibration)
	Beta  float64 // First-order barrier coefficient (Negative)
	Gamma float64 // Stability coefficient (Positive)
	Rho   float64 // Viscosity / Damping (Ohm-meters)

	// Optional LK04 mitigation: allow material-calibrated alpha (instead of Curie-Weiss).
	// When enabled, Alpha is treated as constant and UpdateParams() is skipped.
	UseMaterialAlpha bool

	// Electrostriction & Stress
	Q12    float64 // Electrostriction coefficient (m^4/C^2)
	Stress float64 // In-plane tensile stress (Pa)

	// Circuit parasitics
	SeriesResistance float64 // Series resistance (Ohms)
	Thickness        float64 // Film thickness (m)
	Area             float64 // Active area (m^2)

	// Depolarization (Polycrystalline Analog Behavior)
	K_dep float64 // Depolarization coefficient (V*m/C) - creates "slant" for analog levels

	// NLS Parameters
	UseNLS          bool
	ActivationField float64 // Activation field (V/m) for Merz's Law
	TauInf          float64 // Infinite field switching time (s)
	IncubationEnd   float64 // Time when switching can start (s),

	// Thermodynamic constants
	CurieTemp  float64 // Curie temperature (K)
	CurieConst float64 // Curie constant (K)

	// Noise (Langevin Dynamics)
	EnableNoise bool

	// Deterministic RNG for NLS/noise (per-solver). If nil, uses math/rand global.
	rng *rand.Rand

	// Series-resistance aggregation (ρ_eff = ρ + R_series*A/d)
	UseEffectiveViscosity bool

	// Dynamic State
	Alpha       float64 // Calculated stiffness (Temperature + Stress dependent)
	P           float64 // Current Polarization (C/m^2)
	PMax        float64 // Saturation polarization clamp for numerical stability (C/m^2)
	Temperature float64 // Current Temperature (K)
	Time        float64 // Simulation time

	// Internal logging controls
	logCount int
	logLimit int

	// Numerical stability logging (rate-limited)
	nanCount int
	nanLimit int

	// Polydomain / multi-domain ensemble mode (for multi-level remanent states).
	//
	// A single-domain Landau double-well only supports two stable remanent states at E=0.
	// Multi-level (multi-bit) behavior requires partial switching across many domains with
	// distributed thresholds; ensemble mode approximates this by averaging many LK domains.
	ensemble []*LKSolver
	ensembleImprint []float64 // per-domain imprint / bias field (V/m) added to applied E
	ensembleScale []float64   // per-domain coefficient scale (dimensionless)
	ensembleSeed uint64
}


// NewLKSolver creates a new solver with default "Golden Set" parameters for 10nm HZO.
func NewLKSolver() *LKSolver {
	return &LKSolver{
		// Default to "Golden Set" (Set I)
		Beta:   -2.160e8,
		Gamma:  1.653e10,
		Rho:    0.05,
		Q12:    -0.026,
		Stress: 1.0e9, // 1 GPa

		// Circuit parasitics (from hysteresis-gemini.md)
		SeriesResistance: 50.0,          // Ohms
		Thickness:        10e-9,         // 10 nm
		Area:             45e-9 * 45e-9, // 45 nm x 45 nm (FeCIM default cell)

		// Depolarization for Polycrystalline Analog Behavior
		K_dep: 2.5e8, // V*m/C - Default value (matches physics.yaml, within recommended 1-5×10⁸ range)

		UseNLS:          true,
		ActivationField: 1.9e9, // 19 MV/cm (Merz activation field)
		TauInf:          1.0e-13,

		CurieTemp:  723.0,
		CurieConst: 1.5e5,

		EnableNoise:           false,
		UseEffectiveViscosity: true,
		rng:                   rand.New(rand.NewSource(1)), // deterministic default

		Temperature: 300.0,

		// CRITICAL: Initialize to negative saturation (-Pr)
		// If P0, then E_dep=K_dep*P=0, and depolarization has no effect!
		// This causes binary switching instead of analog slope.
		P:    -0.30, // C/m² (Approximate -Pr for FeCIM HZO at negative saturation)
		PMax: 0.30,  // C/m² (Default clamp, overridden by material config)

		logLimit: 25,
		nanLimit: 5,
	}
}

// UpdateParams recalculates Alpha based on current Temperature and Stress using
// the Unified Coefficient Formula: alpha = alpha_t(T) - 2*Q12*Stress
func (s *LKSolver) UpdateParams() {
	const (
		Eps0 = 8.854e-12 // Vacuum Permittivity (F/m)
	)

	// Thermodynamic contribution (Curie-Weiss)
	alphaT := (s.Temperature - s.CurieTemp) / (2 * Eps0 * s.CurieConst)

	// Mechanical contribution (Electrostriction)
	// Alpha(T,σ) = (T-Tc)/(2ε0C) - 2*Q12*σ
	alphaMech := 2 * s.Q12 * s.Stress

	s.Alpha = alphaT - alphaMech
}

// ConfigureFromMaterial updates solver parameters from HZOMaterial.
// This should be called after NewLKSolver() to override defaults with material-specific values.
// Critical for ensuring the depolarization field (K_dep) matches the material configuration.
func (s *LKSolver) ConfigureFromMaterial(mat *HZOMaterial) {
	if mat == nil {
		return
	}
	// If configured previously in ensemble mode, rebuild domain configs from the new material.
	if len(s.ensemble) > 0 {
		n := len(s.ensemble)
		seed := s.ensembleSeed
		s.ensemble = nil
		s.ensembleImprint = nil
		s.ensembleScale = nil
		s.ensembleSeed = seed
		s.EnableEnsemble(n, mat, seed)
		return
	}

	if mat.BetaLandau != 0 {
		s.Beta = mat.BetaLandau
	}
	if mat.GammaLandau != 0 {
		s.Gamma = mat.GammaLandau
	}
	if mat.RhoViscosity != 0 {
		s.Rho = mat.RhoViscosity
	}
	if mat.Q12 != 0 {
		s.Q12 = mat.Q12
	}
	if mat.StressGPa != 0 {
		s.Stress = mat.StressGPa * 1e9
	}
	if mat.K_dep > 0 {
		s.K_dep = mat.K_dep
	}
	if mat.Thickness > 0 {
		s.Thickness = mat.Thickness
	}
	if mat.Area > 0 {
		s.Area = mat.Area
	}
	if mat.CurieTemp > 0 {
		s.CurieTemp = mat.CurieTemp
	}
	if mat.CurieConst > 0 {
		s.CurieConst = mat.CurieConst
	}
	if mat.SeriesResistanceOhm > 0 {
		s.SeriesResistance = mat.SeriesResistanceOhm
	}
	if mat.Tau0NLS > 0 {
		s.TauInf = mat.Tau0NLS
	}
	if mat.EaNLS > 0 {
		s.ActivationField = mat.EaNLS
	}

	// Initialize P to negative remanent polarization if provided
	if mat.Pr != 0 {
		s.P = -math.Abs(mat.Pr)
	}

	// LK04 mitigation: choose Alpha so that the zero-field equilibrium satisfies
	// dG/dP=0 at |P|=Pr for the configured (Beta,Gamma).
	//
	// Given dG/dP = 2αP + 4βP^3 + 6γP^5, enforcing dG/dP(P=Pr)=0 yields:
	//   α = -2βPr^2 - 3γPr^4
	// This improves consistency between the advertised Pr and the Landau potential.
	if mat.Pr != 0 && s.Gamma != 0 {
		pr := math.Abs(mat.Pr)
		s.Alpha = -2.0*s.Beta*pr*pr - 3.0*s.Gamma*math.Pow(pr, 4)
		s.UseMaterialAlpha = true

		// Optional LK04: scale Landau coefficients to match the material's advertised Ec.
		// The raw (alpha,beta,gamma) sets often imply coercive fields far from mat.Ec,
		// which makes the LK engine effectively unswitchable within the controller's MaxField.
		//
		// We compute a coarse theoretical coercive field from the Landau polynomial
		// E_L(P) = dG/dP (without depolarization) and scale (alpha,beta,gamma) by a
		// single factor so that Ec_theory ≈ mat.Ec while preserving Pr.
		if mat.Ec > 0 {
			ecTheory := estimateLandauEc(s.Alpha, s.Beta, s.Gamma, pr)
			if ecTheory > 0 {
				scale := mat.Ec / ecTheory
				// Clamp to a sane range to avoid pathological configs.
				if scale < 1e-3 {
					scale = 1e-3
				} else if scale > 1e3 {
					scale = 1e3
				}
				s.Beta *= scale
				s.Gamma *= scale
				s.Alpha *= scale
			}
		}
	}
	// Configure saturation clamp using material Ps/Pr when available.
	pMax := math.Max(math.Abs(mat.Ps), math.Abs(mat.Pr))
	if pMax > 0 {
		s.PMax = pMax
	}
	if s.PMax <= 0 {
		s.PMax = math.Abs(s.P)
	}

	// Debug logging to confirm configuration
	matLog.Input("ConfigureFromMaterial", map[string]interface{}{
		"Beta":             s.Beta,
		"Gamma":            s.Gamma,
		"Rho":              s.Rho,
		"Q12":              s.Q12,
		"Stress_Pa":        s.Stress,
		"K_dep":            s.K_dep,
		"Thickness":        s.Thickness,
		"Area":             s.Area,
		"CurieTemp":        s.CurieTemp,
		"CurieConst":       s.CurieConst,
		"SeriesResistance": s.SeriesResistance,
		"ActivationField":  s.ActivationField,
		"TauInf":           s.TauInf,
		"UseMaterialAlpha": s.UseMaterialAlpha,
		"Alpha":            s.Alpha,
		"InitPolarization": s.P,
	})
}

// dPdT is the Master Differential Equation with Depolarization:
// rho * dP/dt = E_eff - dG/dP
// E_eff = E_applied - E_depolarization
// E_depolarization = K_dep * P
// dG/dP = 2*alpha*P + 4*beta*P^3 + 6*gamma*P^5
func (s *LKSolver) dPdT(t, P, E_applied, noise, rhoEff float64) float64 {
	// Depolarization Field: creates "slant" for 30-level analog operation
	// This term opposes polarization, braking the switching to enable intermediate states
	E_depolarization := s.K_dep * P

	// Effective Field
	E_eff := E_applied - E_depolarization

	// Deterministic Force (Gradient of Free Energy)
	P2 := P * P
	P3 := P2 * P
	P5 := P3 * P2
	dG_dP := (2 * s.Alpha * P) + (4 * s.Beta * P3) + (6 * s.Gamma * P5)

	return (E_eff + noise - dG_dP) / rhoEff
}

func (s *LKSolver) dFdP(P, rhoEff float64) float64 {
	if rhoEff == 0 {
		return 0
	}
	P2 := P * P
	P4 := P2 * P2
	d2G := (2 * s.Alpha) + (12 * s.Beta * P2) + (30 * s.Gamma * P4)
	return -(s.K_dep + d2G) / rhoEff
}

func (s *LKSolver) stepImplicit(prevP, E, dt, noise, rhoEff float64) (float64, bool) {
	if dt <= 0 {
		return prevP, true
	}
	guess := prevP + dt*s.dPdT(0, prevP, E, noise, rhoEff)
	if s.PMax > 0 {
		guess = s.clampP(guess)
	}
	tol := 1e-6
	if s.PMax > 0 {
		tol = 1e-6 * s.PMax
	}
	const maxIter = 6
	for i := 0; i < maxIter; i++ {
		f := s.dPdT(0, guess, E, noise, rhoEff)
		g := guess - prevP - dt*f
		if math.Abs(g) < tol {
			return guess, true
		}
		dfdp := s.dFdP(guess, rhoEff)
		denom := 1 - dt*dfdp
		if invalidFloat(denom) || denom == 0 {
			return guess, false
		}
		step := g / denom
		if invalidFloat(step) {
			return guess, false
		}
		guess -= step
		if s.PMax > 0 {
			guess = s.clampP(guess)
		}
		if math.Abs(step) < tol {
			return guess, true
		}
	}
	return guess, !invalidFloat(guess)
}

// Step performs one Runge-Kutta 4 (RK4) integration step.
// Returns the new Polarization P.
func (s *LKSolver) Step(E, dt float64) float64 {
	// Ensemble mode: average many LK domains with per-domain imprint biases.
	if len(s.ensemble) > 0 {
		if dt <= 0 {
			return s.P
		}
		sum := 0.0
		for i, d := range s.ensemble {
			// Keep shared external state in sync.
			d.Temperature = s.Temperature
			d.Stress = s.Stress
			d.UseNLS = s.UseNLS
			d.EnableNoise = s.EnableNoise
				bias := 0.0
			if i < len(s.ensembleImprint) {
				bias = s.ensembleImprint[i]
			}
			sum += d.Step(E+bias, dt)
		}
		s.P = sum / float64(len(s.ensemble))
		s.Time += dt
		return s.P
	}

	if !s.UseMaterialAlpha {
		s.UpdateParams() // Ensure Alpha is current
	}

	rhoEff := s.effectiveRho()
	noise := s.noiseTerm(dt, rhoEff)
	if invalidFloat(s.P) {
		s.logNumericalIssue("state", E, dt, rhoEff, noise, s.P)
		if s.PMax > 0 {
			s.P = -math.Abs(s.PMax)
		} else {
			s.P = 0
		}
	}
	if s.PMax > 0 {
		s.P = s.clampP(s.P)
	}

	// Nucleation-Limited Switching (NLS) Check
	if s.UseNLS {
		if !s.checkIncubation(E, dt) {
			// If still incubating, P does not change significantly (only dielectric response)
			// For simplicity, we define dP/dt = E / Rho (linear response) or just 0
			// A better approx is P = P_prev + epsilon * E
			return s.P
		}
	}

	// RK4 Integration with stability guards.
	prevP := s.P

	// Implicit step for stiff regimes (improves stability with larger dt).
	stiffness := math.Abs(s.dFdP(prevP, rhoEff)) * dt
	const stiffThreshold = 0.5
	if stiffness > stiffThreshold {
		nextP, ok := s.stepImplicit(prevP, E, dt, noise, rhoEff)
		if ok && !invalidFloat(nextP) {
			nextP = s.clampP(nextP)
			s.P = nextP
			s.Time += dt
			s.logStep(E, dt, rhoEff, noise, s.dPdT(0, prevP, E, noise, rhoEff))
			return s.P
		}
	}

	// Rate limiter: cap |dP/dt| with a fixed ceiling to avoid overflow without
	// canceling the RK4 step (dt-scaled clamps can cause k1/k2 sign flipping).
	const maxAbsRate = 1e12
	clampRate := func(rate float64) float64 {
		if rate > maxAbsRate {
			return maxAbsRate
		}
		if rate < -maxAbsRate {
			return -maxAbsRate
		}
		return rate
	}

	// Clamp intermediate polarization states to keep RK4 evaluation stable.
	clampState := func(p float64) float64 {
		return s.clampP(p)
	}

	k1 := clampRate(s.dPdT(0, prevP, E, noise, rhoEff))
	p2 := clampState(prevP + 0.5*dt*k1)
	k2 := clampRate(s.dPdT(dt/2, p2, E, noise, rhoEff))
	p3 := clampState(prevP + 0.5*dt*k2)
	k3 := clampRate(s.dPdT(dt/2, p3, E, noise, rhoEff))
	p4 := clampState(prevP + dt*k3)
	k4 := clampRate(s.dPdT(dt, p4, E, noise, rhoEff))

	if invalidFloat(k1) || invalidFloat(k2) || invalidFloat(k3) || invalidFloat(k4) {
		s.logNumericalIssue("k", E, dt, rhoEff, noise, prevP)
		s.Time += dt
		return s.P
	}

	dP := (dt / 6.0) * (k1 + 2*k2 + 2*k3 + k4)
	// Prevent pathological single-step jumps while preserving direction.
	if s.PMax > 0 {
		maxDelta := 2.0 * s.PMax
		if dP > maxDelta {
			dP = maxDelta
		} else if dP < -maxDelta {
			dP = -maxDelta
		}
	}
	if invalidFloat(dP) {
		s.logNumericalIssue("dP", E, dt, rhoEff, noise, prevP)
		s.Time += dt
		return s.P
	}
	nextP := prevP + dP
	if invalidFloat(nextP) {
		s.logNumericalIssue("P", E, dt, rhoEff, noise, prevP)
		s.Time += dt
		return s.P
	}
	nextP = s.clampP(nextP)

	s.P = nextP
	s.Time += dt

	s.logStep(E, dt, rhoEff, noise, k1)

	return s.P
}

// checkIncubation determines if the domain has finished incubating based on Merz's Law.
// Returns true if switching can proceed.
func (s *LKSolver) checkIncubation(E, dt float64) bool {
	// Simplified NLS:
	// Calculate Incubation Time t_inc = tau_inf * exp(Ea / E)
	// If cumulative time under field E > t_inc, then switch.

	E_mag := math.Abs(E)
	const MinField = 1.0e6 // 0.01 MV/cm threshold

	if E_mag < MinField {
		return false // Field too small to drive nucleation
	}

	// Merz's Law for Incubation Time
	// E should be in V/m. ActivationField is in V/m.
	activationField := s.ActivationField
	if activationField <= 0 {
		activationField = 1.9e9
	}

	tNum := s.TauInf * math.Exp(activationField/E_mag)

	// Add to incubation time tracker (simplified for now)
	// Real NLS requires tracking state for each domain.
	// For a single solver instance (mean field), we can just use a delay counter.

	// For this Phase 1 implementation, we will perform a probabilistic check
	// Probability of nucleation in dt: P_nuc = 1 - exp(-dt / t_inc)

	prob := 1.0 - math.Exp(-dt/tNum)
	if s.rng != nil {
		return s.rng.Float64() < prob
	}
	return rand.Float64() < prob
}

func (s *LKSolver) effectiveRho() float64 {
	rhoEff := s.Rho
	if s.UseEffectiveViscosity && s.SeriesResistance > 0 && s.Thickness > 0 && s.Area > 0 {
		rhoEff += (s.SeriesResistance * s.Area / s.Thickness)
	}
	return rhoEff
}

func (s *LKSolver) noiseTerm(dt, rhoEff float64) float64 {
	if !s.EnableNoise || dt <= 0 {
		return 0
	}

	const kB = 1.380649e-23 // J/K
	sigma := math.Sqrt(2 * kB * s.Temperature * rhoEff / dt)
	if s.rng != nil {
		return s.rng.NormFloat64() * sigma
	}
	return rand.NormFloat64() * sigma
}

func (s *LKSolver) logStep(E, dt, rhoEff, noise, dPdt float64) {
	if !logging.IsVerbose(logging.VerbosityTrace) {
		return
	}
	if lkLog == nil {
		lkLog = logging.NewLogger("lk-solver")
	}
	if lkLog == nil {
		return
	}
	if s.logLimit > 0 && s.logCount >= s.logLimit {
		return
	}
	s.logCount++

	E_dep := s.K_dep * s.P
	E_eff := E - E_dep
	dG_dP := (2 * s.Alpha * s.P) + (4 * s.Beta * math.Pow(s.P, 3)) + (6 * s.Gamma * math.Pow(s.P, 5))

	lkLog.Calculation("LKStep", map[string]interface{}{
		"E_applied":   E,
		"E_dep":       E_dep,
		"E_eff":       E_eff,
		"Alpha":       s.Alpha,
		"Beta":        s.Beta,
		"Gamma":       s.Gamma,
		"K_dep":       s.K_dep,
		"P":           s.P,
		"dG_dP":       dG_dP,
		"rho_eff":     rhoEff,
		"noise":       noise,
		"dt":          dt,
		"Temperature": s.Temperature,
		"Stress_Pa":   s.Stress,
	}, dPdt)
}

func invalidFloat(v float64) bool {
	return math.IsNaN(v) || math.IsInf(v, 0)
}

func (s *LKSolver) clampP(P float64) float64 {
	if s.PMax <= 0 {
		return P
	}
	limit := s.PMax * 1.2
	if limit <= 0 {
		return P
	}
	if P > limit {
		return limit
	}
	if P < -limit {
		return -limit
	}
	return P
}

func (s *LKSolver) logNumericalIssue(stage string, E, dt, rhoEff, noise, prevP float64) {
	if !logging.IsVerbose(logging.VerbosityDebug) {
		return
	}
	if s.nanLimit > 0 && s.nanCount >= s.nanLimit {
		return
	}
	s.nanCount++
	if lkLog == nil {
		lkLog = logging.NewLogger("lk-solver")
	}
	if lkLog == nil {
		return
	}
	lkLog.Debug("LK numerical issue (%s): E=%.3e dt=%.3e P=%.3e rho=%.3e noise=%.3e alpha=%.3e beta=%.3e gamma=%.3e",
		stage, E, dt, prevP, rhoEff, noise, s.Alpha, s.Beta, s.Gamma)
}

func (s *LKSolver) SetState(P float64) {
	if invalidFloat(P) {
		return
	}
	if len(s.ensemble) > 0 {
		for _, d := range s.ensemble {
			d.SetState(P)
			d.Time = 0
		}
		s.P = s.clampP(P)
		return
	}
	s.P = s.clampP(P)
}

func (s *LKSolver) GetState() float64 {
	return s.P
}

// EnableEnsemble switches this solver into a polydomain ensemble mode.
//
// The ensemble approximates multi-level remanent states by averaging many LK domains
// with slightly different switching thresholds (modeled here as an imprint/bias field).
//
// Determinism: if seed==0, a seed is derived from material name + domain count.
func (s *LKSolver) EnableEnsemble(numDomains int, mat *HZOMaterial, seed uint64) {
	if numDomains <= 1 {
		s.ensemble = nil
		s.ensembleImprint = nil
		s.ensembleScale = nil
		s.ensembleSeed = 0
		return
	}
	if mat == nil {
		return
	}
	if seed == 0 {
		h := fnv.New64a()
		_, _ = h.Write([]byte(mat.Name))
		// mix in domain count
		_ = h.Sum64()
		seed = h.Sum64() ^ uint64(numDomains*0x9e3779b1)
	}
	// Build domains.
	rng := rand.New(rand.NewSource(int64(seed)))
	domains := make([]*LKSolver, 0, numDomains)
	imprint := make([]float64, 0, numDomains)
	scales := make([]float64, 0, numDomains)

	// Domain dispersion:
	// - imprint/bias field shifts switching threshold (proxy for imprint/defects)
	// - coefficient scale perturbs (alpha,beta,gamma) together (proxy for grain-to-grain variation)
	imprintSigma := 0.10 * mat.Ec
	scaleSigma := 0.20 // 20% 1-sigma
	for i := 0; i < numDomains; i++ {
		d := NewLKSolver()
		// Per-domain deterministic RNG
		d.rng = rand.New(rand.NewSource(int64(seed) + int64(i+1)*0x9e3779b97f4a7c1))
		d.ConfigureFromMaterial(mat)
		// Keep settings aligned.
		d.EnableNoise = s.EnableNoise
		d.UseNLS = s.UseNLS
		d.Temperature = s.Temperature
		d.Stress = s.Stress

		// Seeded per-domain variations.
		scale := 1.0 + rng.NormFloat64()*scaleSigma
		if scale < 0.5 {
			scale = 0.5
		} else if scale > 1.8 {
			scale = 1.8
		}
		// Scale Landau coefficients together to perturb coercive response while preserving Pr relation.
		d.Alpha *= scale
		d.Beta *= scale
		d.Gamma *= scale

		bias := rng.NormFloat64() * imprintSigma
		domains = append(domains, d)
		imprint = append(imprint, bias)
		scales = append(scales, scale)
	}

	s.ensemble = domains
	s.ensembleImprint = imprint
	s.ensembleScale = scales
	s.ensembleSeed = seed

	// Initialise ensemble to current state (or sensible default).
	initP := s.P
	if initP == 0 {
		initP = -math.Abs(mat.Pr)
		if initP == 0 {
			initP = -math.Abs(mat.Ps)
		}
	}
	s.SetState(initP)
}
