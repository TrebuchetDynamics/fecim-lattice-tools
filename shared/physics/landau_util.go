// landau_util.go contains LKSolver helper methods: NLS switching statistics,
// effective viscosity, noise, clamping, logging, state management, and ensemble
// configuration. The core integration (Step, dPdT, RK4, implicit) lives in landau.go.
package physics

import (
	"math"
	"math/rand"

	"fecim-lattice-tools/shared/logging"
)

func (s *LKSolver) updateNLSState(E, dt float64) {
	const minField = 1.0e6 // 0.01 MV/cm threshold
	if dt <= 0 {
		return
	}
	if math.Abs(E) < minField {
		s.nlsState.CumulativeTime = 0
		s.nlsState.SwitchedFraction = 0
		return
	}
	s.nlsState.CumulativeTime += dt
	s.nlsState.SwitchedFraction = s.nlsSwitchedFraction(E, s.nlsState.CumulativeTime)
}

// nlsSwitchedFraction returns deterministic cumulative switched fraction under
// field E and total stress time, using a log-normal distribution of switching
// times (Guo et al., APL 112, 262903, 2018).
func (s *LKSolver) nlsSwitchedFraction(E, totalTime float64) float64 {
	E_mag := math.Abs(E)
	if E_mag < 1e6 || totalTime <= 0 {
		return 0
	}
	// Fallback defaults mirror the NewLKSolver constructor values so that a
	// zero-value solver still produces physically sensible NLS behaviour.
	tauInf := s.TauInf
	if tauInf <= 0 {
		tauInf = 1e-10 // matches NewLKSolver default (100 ps); Guo et al. APL 112, 262903 (2018)
	}
	activationField := s.ActivationField
	if activationField <= 0 {
		activationField = 1.9e9 // matches NewLKSolver default (19 MV/cm)
	}
	sigma := s.NLSSigma
	if sigma <= 0 {
		sigma = 1.5 // matches NewLKSolver default; Guo et al. APL 112, 262903 (2018)
	}

	// Gauss-Hermite-style quadrature over log-normal switching-time distribution.
	// nlsQuadN: number of quadrature points (20 balances accuracy vs. speed;
	//   convergence tests show <0.5% error vs. N=100 for typical ISPP parameters).
	// nlsQuadSpan: integration range in multiples of sigma (±3σ covers 99.7%).
	const (
		nlsQuadN    = 20
		nlsQuadSpan = 6.0 // total span = ±3σ
	)

	lnTauMean := math.Log(tauInf) + activationField/E_mag
	f := 0.0
	norm := 0.0
	for i := 0; i < nlsQuadN; i++ {
		x := lnTauMean + sigma*(float64(i)-float64(nlsQuadN-1)/2.0)*nlsQuadSpan/float64(nlsQuadN)
		tau := math.Exp(x)
		weight := math.Exp(-0.5 * math.Pow((x-lnTauMean)/sigma, 2))
		f += weight * (1.0 - math.Exp(-totalTime/tau))
		norm += weight
	}
	if norm > 0 {
		f /= norm
	}
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 1
	}
	return f
}

func (s *LKSolver) effectiveRho() float64 {
	rhoEff := s.Rho
	if invalidFloat(rhoEff) || rhoEff > maxLKViscosity {
		rhoEff = defaultLKViscosity
	}
	if s.UseEffectiveViscosity && s.SeriesResistance > 0 && s.Thickness > 0 && s.Area > 0 {
		seriesTerm := s.SeriesResistance * s.Area / s.Thickness
		if isValidLKRuntimeViscosity(seriesTerm) {
			next := rhoEff + seriesTerm
			if isValidLKRuntimeViscosity(next) {
				rhoEff = next
			}
		}
	}
	// Guard: viscosity must be strictly positive for dP/dt = (...)/rhoEff.
	// A zero, negative, non-finite, or unrepresentably large value would produce
	// Inf/NaN or effectively stall corrupted public state; fall back to the
	// literature default for 10 nm HfO2 (Materlik 2015).
	if !isValidLKRuntimeViscosity(rhoEff) {
		rhoEff = defaultLKViscosity
	}
	return rhoEff
}

func (s *LKSolver) noiseTerm(dt, rhoEff float64) float64 {
	if !s.EnableNoise || dt <= 0 || invalidFloat(dt) || invalidFloat(rhoEff) {
		return 0
	}

	const kB = 1.380649e-23 // J/K
	// Fluctuation-dissipation theorem for intensive polarization dynamics.
	// sigma = sqrt(2*kB*T*rho / (dt * V_cell)) gives correct 1/sqrt(V) Landauer scaling.
	vCell := s.Area * s.Thickness
	if vCell <= 0 || invalidFloat(vCell) {
		vCell = 45e-9 * 45e-9 * 10e-9 // fallback: default FeCIM cell
	}
	// Guard: the sqrt argument must be non-negative. Negative temperature or
	// negative rhoEff would produce NaN; clamp to zero noise in that case.
	arg := 2 * kB * s.Temperature * rhoEff / (dt * vCell)
	if arg <= 0 || invalidFloat(arg) {
		return 0
	}
	sigma := math.Sqrt(arg)
	if invalidFloat(sigma) {
		return 0
	}
	if s.rng != nil {
		noise := s.rng.NormFloat64() * sigma
		if invalidFloat(noise) {
			return 0
		}
		return noise
	}
	noise := rand.NormFloat64() * sigma
	if invalidFloat(noise) {
		return 0
	}
	return noise
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

	P2 := s.P * s.P
	P3 := P2 * s.P
	P5 := P3 * P2
	E_dep := lkRuntimeTerm(s.K_dep * s.P)
	E_eff := E - E_dep
	dG_dP := sumLKRuntimeTerms(
		lkRuntimeTerm(2*s.Alpha*s.P),
		lkRuntimeTerm(4*s.Beta*P3),
		lkRuntimeTerm(6*s.Gamma*P5),
	)

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

func lkRuntimeTerm(term float64) float64 {
	if invalidFloat(term) {
		return 0
	}
	return term
}

func isRepresentableLKRuntimeForceTerm(term, rhoEff float64) bool {
	if invalidFloat(term) || rhoEff <= 0 || invalidFloat(rhoEff) {
		return false
	}
	return !invalidFloat(term / rhoEff)
}

func lkRuntimeForceTerm(term, rhoEff float64) float64 {
	if !isRepresentableLKRuntimeForceTerm(term, rhoEff) {
		return 0
	}
	return term
}

func sumLKRuntimeTerms(terms ...float64) float64 {
	total := 0.0
	for _, term := range terms {
		if invalidFloat(term) {
			continue
		}
		next := total + term
		if invalidFloat(next) {
			return 0
		}
		total = next
	}
	return total
}

func (s *LKSolver) runtimeAlphaFor(temperature, stress float64) (float64, bool) {
	if s == nil || invalidFloat(temperature) || invalidFloat(stress) || invalidFloat(s.Q12) || invalidFloat(s.CurieConst) {
		return 0, false
	}

	alphaMech := 2 * s.Q12 * stress
	if invalidFloat(alphaMech) {
		return 0, false
	}

	alpha := -alphaMech
	if s.CurieConst > 0 {
		if invalidFloat(s.CurieTemp) {
			return 0, false
		}
		const eps0 = 8.854e-12 // Vacuum Permittivity (F/m)
		denom := 2 * eps0 * s.CurieConst
		if denom == 0 || invalidFloat(denom) {
			return 0, false
		}
		thermalNumerator := temperature - s.CurieTemp
		if invalidFloat(thermalNumerator) {
			return 0, false
		}
		alphaT := thermalNumerator / denom
		if invalidFloat(alphaT) {
			return 0, false
		}
		alpha = alphaT - alphaMech
	}
	if !s.isRepresentableRuntimeAlpha(alpha) {
		return 0, false
	}
	return alpha, true
}

func (s *LKSolver) isRepresentableRuntimeAlpha(alpha float64) bool {
	if s == nil || invalidFloat(alpha) {
		return false
	}
	pScale := s.PMax
	if pScale <= 0 || !isRepresentableLKPolarization(pScale) {
		pScale = 1
	}
	force := 2 * alpha * pScale
	if invalidFloat(force) {
		return false
	}
	rhoEff := s.effectiveRho()
	if rhoEff <= 0 || invalidFloat(rhoEff) {
		return false
	}
	rate := force / rhoEff
	return !invalidFloat(rate) && math.Abs(rate) <= maxAbsLKRate
}

func isValidLKStepInput(E, dt, currentTime, rhoEff float64) bool {
	if invalidFloat(E) || invalidFloat(dt) || invalidFloat(currentTime) || currentTime < 0 || currentTime > maxLKSimulationTime || dt <= 0 || dt > maxLKTimestep {
		return false
	}
	if !isRepresentableLKRuntimeForceTerm(E, rhoEff) {
		return false
	}
	return !invalidFloat(currentTime + dt)
}

func isRepresentableLKMaterialScalar(value float64) bool {
	return !invalidFloat(value) && math.Abs(value) <= maxLKMaterialScalarMagnitude
}

func isNonZeroFiniteLKMaterialValue(value float64) bool {
	return value != 0 && isRepresentableLKMaterialScalar(value)
}

func isPositiveFiniteLKMaterialValue(value float64) bool {
	return value > 0 && isRepresentableLKMaterialScalar(value)
}

func finiteScaledLKMaterialValue(value, scale float64) (float64, bool) {
	if !isRepresentableLKMaterialScalar(value) || !isRepresentableLKMaterialScalar(scale) {
		return 0, false
	}
	out := value * scale
	if !isRepresentableLKMaterialScalar(out) {
		return 0, false
	}
	return out, true
}

func (s *LKSolver) isRepresentableLKMaterialLandauScaling(mat *HZOMaterial) bool {
	if s == nil || mat == nil {
		return false
	}
	if mat.Pr == 0 || !isRepresentableLKPolarization(mat.Pr) {
		return true
	}

	beta := s.Beta
	if isNonZeroFiniteLKMaterialValue(mat.BetaLandau) {
		beta = mat.BetaLandau
	}
	gamma := s.Gamma
	if isNonZeroFiniteLKMaterialValue(mat.GammaLandau) {
		gamma = mat.GammaLandau
	}
	if gamma == 0 {
		return true
	}

	pr := math.Abs(mat.Pr)
	alpha, ok := landauAlphaForPr(beta, gamma, pr)
	if !ok || !isRepresentableLKMaterialScalar(alpha) {
		return false
	}
	if !isPositiveFiniteLKMaterialValue(mat.Ec) {
		return true
	}

	ecTheory := estimateLandauEc(alpha, beta, gamma, pr)
	if invalidFloat(ecTheory) || math.Abs(ecTheory) > maxLKMaterialScalarMagnitude {
		return false
	}
	if ecTheory <= 0 {
		return true
	}
	scale := mat.Ec / ecTheory
	if invalidFloat(scale) {
		return false
	}
	if scale < 1e-3 {
		scale = 1e-3
	} else if scale > 1e3 {
		scale = 1e3
	}
	return isRepresentableLKMaterialScalar(alpha*scale) &&
		isRepresentableLKMaterialScalar(beta*scale) &&
		isRepresentableLKMaterialScalar(gamma*scale)
}

func isRepresentableLKPolarization(value float64) bool {
	return !invalidFloat(value) && math.Abs(value) <= maxLKPolarizationMagnitude
}

func isValidLKRuntimePMax(value float64) bool {
	return value == 0 || (value > 0 && isRepresentableLKPolarization(value))
}

func isValidLKRuntimeViscosity(value float64) bool {
	return value > 0 && !invalidFloat(value) && value <= maxLKViscosity
}

func recoveredLKRuntimePMax(currentP float64) float64 {
	if isRepresentableLKPolarization(currentP) && currentP != 0 {
		return math.Abs(currentP)
	}
	return defaultLKRuntimePMax
}

func landauAlphaForPr(beta, gamma, pr float64) (float64, bool) {
	if invalidFloat(beta) || invalidFloat(gamma) || !isRepresentableLKPolarization(pr) {
		return 0, false
	}
	p2 := pr * pr
	p4 := p2 * p2
	if invalidFloat(p2) || invalidFloat(p4) {
		return 0, false
	}
	alpha := -2.0*beta*p2 - 3.0*gamma*p4
	if invalidFloat(alpha) {
		return 0, false
	}
	return alpha, true
}

// pClampOvershootFactor allows 20% overshoot above PMax before hard-clamping.
// This headroom lets the RK4 integrator make small excursions past saturation
// without immediately hitting the hard wall, which would cause step-size hunting.
const pClampOvershootFactor = 1.2

func (s *LKSolver) clampP(P float64) float64 {
	if s.PMax <= 0 {
		return P
	}
	limit := s.PMax * pClampOvershootFactor
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

// SetState forcibly sets polarization state P (C/m^2), with NaN/Inf rejection
// and solver clamp rules for numerical safety.
//
// In ensemble mode the requested state is broadcast to all domains.
func (s *LKSolver) SetState(P float64) {
	if s == nil || !isRepresentableLKPolarization(P) {
		return
	}
	if s.polydomain != nil && s.polydomain.DomainCount() > 0 {
		s.polydomain.SetState(P)
		s.P = s.clampP(P)
		return
	}
	s.P = s.clampP(P)
}

// GetState returns the current solver polarization P in C/m^2.
func (s *LKSolver) GetState() float64 {
	if s == nil {
		return 0
	}
	return s.P
}

// StepWarning reports when a timestep may be too large for the current dynamics.
// The Courant-like condition for LK is dt < rho / |d²G/dP²|.
// We use a simplified heuristic: if |dP| > 0.1*PMax, the step is suspicious.
type StepWarning struct {
	LargeStep   bool    // dt may be too large
	StepRatio   float64 // |dP| / PMax — should be < 0.1
	Recommended float64 // suggested dt (current dt * 0.1 / ratio)
}

// CheckTimestep evaluates whether the given timestep dt is safe for the current
// solver state under applied field E. Returns nil if the timestep is acceptable,
// or a StepWarning with a recommended smaller dt if the step is suspiciously large.
//
// Call this before Step() to get an advisory warning without modifying solver state.
func (s *LKSolver) CheckTimestep(E, dt float64) *StepWarning {
	if s == nil || s.PMax <= 0 || !isRepresentableLKPolarization(s.PMax) || !isRepresentableLKPolarization(s.P) {
		return nil
	}
	rhoEff := s.effectiveRho()
	if !isValidLKStepInput(E, dt, s.Time, rhoEff) {
		return nil
	}

	// Estimate |dP| from the current dP/dt rate.
	rate := s.dPdT(0, s.P, E, 0, rhoEff)
	if invalidFloat(rate) {
		return nil
	}
	absDeltaP := math.Abs(rate * dt)
	if invalidFloat(absDeltaP) {
		return nil
	}
	ratio := absDeltaP / s.PMax
	if invalidFloat(ratio) {
		return nil
	}

	const safeRatio = 0.1
	if ratio <= safeRatio {
		return nil
	}

	recommended := dt * safeRatio / ratio
	if invalidFloat(recommended) {
		return nil
	}
	if recommended < 1e-15 {
		recommended = 1e-15
	}
	return &StepWarning{
		LargeStep:   true,
		StepRatio:   ratio,
		Recommended: recommended,
	}
}

// EnableEnsemble switches this solver into polydomain mode.
func (s *LKSolver) EnableEnsemble(numDomains int, mat *HZOMaterial, seed uint64) {
	if s == nil {
		return
	}
	if numDomains <= 1 {
		s.polydomain = nil
		s.ensembleSeed = 0
		return
	}
	if !isValidPolydomainMaterial(mat) || !isValidPolydomainDomainCount(numDomains) {
		return
	}
	if !isValidLKRuntimePMax(s.PMax) {
		s.PMax = recoveredLKRuntimePMax(s.P)
	}
	initP := s.P
	if !isRepresentableLKPolarization(initP) || initP == 0 {
		initP = recoveredLKEnsembleInitialPolarization(mat, s.PMax)
	}

	ensemble := NewPolydomainEnsemble(s, mat, numDomains, defaultPolydomainSigmaFrac, seed)
	if ensemble == nil {
		return
	}
	s.polydomain = ensemble
	s.ensembleSeed = s.polydomain.Seed
	s.SetState(initP)
}

func recoveredLKEnsembleInitialPolarization(mat *HZOMaterial, pMax float64) float64 {
	if mat != nil {
		if mat.Pr != 0 && isRepresentableLKPolarization(mat.Pr) {
			return -math.Abs(mat.Pr)
		}
		if mat.Ps != 0 && isRepresentableLKPolarization(mat.Ps) {
			return -math.Abs(mat.Ps)
		}
	}
	if pMax != 0 && isRepresentableLKPolarization(pMax) {
		return -math.Abs(pMax)
	}
	return -defaultLKRuntimePMax
}
