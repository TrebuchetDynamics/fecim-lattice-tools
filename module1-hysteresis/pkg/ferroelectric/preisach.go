// Package ferroelectric provides physics models for ferroelectric materials.
package ferroelectric

import (
	"math"

	"fecim-lattice-tools/shared/logging"
	"fecim-lattice-tools/shared/physics"
)

// Package-level logger
var log *logging.Logger

// Preisach model constants
const (
	// defaultDeltaFrac is the initial Tanh Everett Delta/Ec ratio before calibration.
	// Tuned via tuneDeltaForPr in updateReversibleParams.
	defaultDeltaFrac = 0.25

	// saturationFieldMultiplier is the factor applied to Ec to determine the
	// saturation field used by the Preisach stack. Typical 3-5x; we use 5x to
	// ensure full saturation coverage across all minor loops.
	saturationFieldMultiplier = 5.0

	// defaultStressGPa is the default in-plane stress for Preisach simulation.
	// Typical TiN-capped HZO stack, ~1 GPa compressive.
	defaultStressGPa = 1.0

	// nlsEaEcRatio is the heuristic ratio for NLS activation field from Ec.
	// NLS Ea ≈ 10*Ec is an empirical approximation for HZO-class ferroelectrics.
	nlsEaEcRatio = 10.0

	// quasiStaticDt is the effective dt (seconds) used in Update() to approximate
	// infinite-time (quasi-static) switching via NLS relaxation.
	quasiStaticDt = 1.0

	// roomTemperatureK is the default operating temperature.
	roomTemperatureK = 300.0

	// defaultQ12HZO is the transverse electrostriction coefficient for HZO.
	// Ref: Park et al., J. Appl. Phys. 117, 074103 (2015).
	// See material.go Q12 field comment for calibration note vs DFT values.
	defaultQ12HZO = -0.026

	// maxHysteresisLoopPoints bounds public loop requests. Visualization and
	// validation callers use hundreds of points; this still permits high-resolution
	// offline loops while keeping allocation and integer arithmetic bounded.
	maxHysteresisLoopPoints = 1_000_000

	// maxDiscreteStateCount bounds programmable-state generation. Default use is
	// tens of states; this upper bound keeps accidental allocation attacks from
	// panicking while leaving ample room for offline sweeps.
	maxDiscreteStateCount = 1_000_000
)

func init() {
	log = logging.NewLogger("preisach")
}

// TanhEverett is a compatibility alias for shared Preisach Everett adapter.
type TanhEverett = physics.TanhEverett

// tuneDeltaForPr estimates a Tanh Everett distribution width (Delta) so that
// the remanent polarization after a full saturation-and-return matches targetPr.
// This keeps the Preisach loop consistent with the material Pr/Ps ratio.
func tuneDeltaForPr(ec, saturationE, psIrrev, targetPr float64) float64 {
	if ec <= 0 {
		return 0
	}
	if psIrrev <= 0 || targetPr <= 0 {
		return ec * defaultDeltaFrac
	}

	satE := math.Abs(saturationE)
	if satE <= 0 {
		satE = ec * saturationFieldMultiplier
	}

	targetRatio := targetPr / psIrrev
	if targetRatio <= 0 {
		return ec * 2.0
	}
	if targetRatio > 0.999 {
		targetRatio = 0.999
	}
	if targetRatio < 0.01 {
		targetRatio = 0.01
	}

	ratioFor := func(delta float64) float64 {
		if delta <= 0 {
			return 0
		}
		everett := &TanhEverett{
			Ps:    psIrrev,
			Ec:    ec,
			Delta: delta,
		}
		stack := physics.NewPreisachStack(satE, everett)
		stack.Update(satE)
		pr := stack.Update(0)
		if psIrrev == 0 {
			return 0
		}
		ratio := pr / psIrrev
		if ratio < 0 {
			return 0
		}
		if ratio > 1 {
			return 1
		}
		return ratio
	}

	lo := ec * 0.05
	hi := ec * 2.0
	rLo := ratioFor(lo)
	rHi := ratioFor(hi)
	if rLo < rHi {
		lo, hi = hi, lo
		rLo, rHi = rHi, rLo
	}

	// Expand search bounds if needed to bracket target.
	for rLo < targetRatio && lo > ec*1e-6 {
		lo *= 0.5
		rLo = ratioFor(lo)
	}
	for rHi > targetRatio && hi < ec*10.0 {
		hi *= 1.5
		rHi = ratioFor(hi)
	}

	if targetRatio >= rLo {
		return lo
	}
	if targetRatio <= rHi {
		return hi
	}

	for i := 0; i < 32; i++ {
		mid := 0.5 * (lo + hi)
		rMid := ratioFor(mid)
		if rMid > targetRatio {
			lo = mid
		} else {
			hi = mid
		}
	}
	return 0.5 * (lo + hi)
}

// PreisachModel implements the classical Preisach hysteresis model for
// ferroelectric materials, wrapping shared/physics.PreisachStack.
//
// The total polarization has two contributions:
//
//	P_total(E) = P_irrev(E) + P_rev(E)
//
// where P_irrev comes from the Preisach stack (irreversible domain switching)
// and P_rev = P_sat_rev * tanh(E/Ec) is a nonlinear reversible (dielectric)
// contribution derived from the material's low-frequency permittivity.
//
// NLS (Nucleation-Limited Switching) kinetics from physics.NLSKinetics provide
// time-dependent relaxation: P_final = NLS.Relax(P_start, P_target, E, dt).
//
// dynamicP tracks physical polarization across Reset() calls to prevent
// plot teleportation during PREP phases (see MEMORY.md).
type PreisachModel struct {
	material *HZOMaterial
	stack    *physics.PreisachStack
	everett  *TanhEverett

	Temperature float64
	// NOTE: Preisach.Stress is in GPa (converted to Pa inline at calculation site).
	// LKSolver.Stress stores Pa directly. Be careful when passing values between models.
	Stress float64 // GPa

	// Reversible (nonlinear) contribution derived from permittivity and Ec.
	reversibleChi  float64 // C/(V*m)
	reversiblePSat float64 // Saturating reversible polarization (C/m^2)

	effectivePs float64 // Temperature/stress-adjusted total Ps

	// Kinetics
	nls *physics.NLSKinetics

	// dynamicP tracks the actual physical polarization across Reset() calls.
	// Both Update() and TimeStep() use this as P_start to avoid plot
	// teleportation: Reset() reinitializes the stack to LastE=-saturationE (so
	// Polarization() returns ~-Ps), but the device is still physically at its
	// pre-reset P. By keeping dynamicP alive, the first PREP-phase step starts
	// from the real P and drives smoothly to saturation instead of jumping.
	// Reset() intentionally does NOT clear these fields.
	// lockDynamic prevents Update()/TimeStep() from writing dynamicP during
	// quasi-static loop generation (GetHysteresisLoop) so the active simulation
	// state is unaffected. When locked, P_start falls back to Polarization().
	dynamicP    float64 // last P_final from Update/TimeStep (C/m²)
	hasDynamicP bool    // true once Update/TimeStep has been called at least once
	lockDynamic bool    // when true, skip dynamicP read/write (loop generation)
}

func isValidPreisachMaterial(material *HZOMaterial) bool {
	if material == nil {
		return false
	}
	return material.Ps > 0 && isFinite(material.Ps) &&
		material.Pr > 0 && isFinite(material.Pr) && material.Pr <= material.Ps &&
		isRepresentableCoerciveField(material.Ec)
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func isRepresentableCoerciveField(ec float64) bool {
	return ec > 0 && isFinite(ec) &&
		isFinite(ec*saturationFieldMultiplier) &&
		isFinite(ec*nlsEaEcRatio)
}

func isRepresentableLoopField(maxField float64) bool {
	return maxField > 0 && isFinite(maxField) && isFinite(2*maxField)
}

func isValidLoopPointCount(points int) bool {
	return points > 0 && points <= maxHysteresisLoopPoints
}

func isValidDiscreteStateCount(n int) bool {
	return n >= 2 && n <= maxDiscreteStateCount
}

func snapshotMaterial(material *HZOMaterial) *HZOMaterial {
	if material == nil {
		return nil
	}
	snapshot := *material
	return &snapshot
}

func newNLSKineticsForMaterial(material *HZOMaterial) *physics.NLSKinetics {
	nls := physics.NewNLSKinetics()
	if material == nil {
		return nls
	}
	if material.Tau0NLS > 0 && !math.IsNaN(material.Tau0NLS) && !math.IsInf(material.Tau0NLS, 0) {
		nls.Tau0 = material.Tau0NLS
	}
	if material.EaNLS > 0 && !math.IsNaN(material.EaNLS) && !math.IsInf(material.EaNLS, 0) {
		nls.Ea = material.EaNLS
	} else if material.Ec > 0 {
		// Fallback heuristic: Ea ≈ 10*Ec for materials without explicit NLS data.
		nls.Ea = material.Ec * nlsEaEcRatio
	}
	return nls
}

func newQuasiStaticNLSForMaterial(material *HZOMaterial) *physics.NLSKinetics {
	nls := physics.NewNLSKinetics()
	if material != nil && material.Ec > 0 && !math.IsNaN(material.Ec) && !math.IsInf(material.Ec, 0) {
		// Keep the quasi-static Update path independent of material time constants;
		// TimeStep is the public interface for material-specific switching kinetics.
		nls.Ea = material.Ec * nlsEaEcRatio
	}
	return nls
}

// NewPreisachModel creates a new Preisach model with the given material.
func NewPreisachModel(material *HZOMaterial) *PreisachModel {
	materialSnapshot := snapshotMaterial(material)
	if !isValidPreisachMaterial(materialSnapshot) {
		return nil
	}

	log.Input("NewPreisachModel", map[string]interface{}{
		"material_name": materialSnapshot.Name,
		"Ec":            materialSnapshot.Ec,
		"Ps":            materialSnapshot.Ps,
	})

	// Configure Everett function based on material
	everett := &TanhEverett{
		Ps:    materialSnapshot.Ps,
		Ec:    materialSnapshot.Ec,
		Delta: materialSnapshot.Ec * defaultDeltaFrac, // Initial guess; tuned to match Pr in updateReversibleParams
	}

	// E_saturation should be > Ec. typically 3-5x Ec.
	E_sat := materialSnapshot.Ec * saturationFieldMultiplier

	nls := newNLSKineticsForMaterial(materialSnapshot)

	model := &PreisachModel{
		material:    materialSnapshot,
		stack:       physics.NewPreisachStack(E_sat, everett),
		everett:     everett,
		Temperature: roomTemperatureK,
		Stress:      defaultStressGPa,
		effectivePs: materialSnapshot.Ps,
		nls:         nls,
	}
	model.updateReversibleParams()

	return model
}

// DiscreteState represents a single programmable state.
type DiscreteState struct {
	Level        int
	Polarization float64
	NormalizedP  float64
	Voltage      float64
	Conductance  float64
}

// DiscreteStates returns n evenly spaced polarization states from -Ps to +Ps.
// Used for testing and visualization of programmable level distributions.
func (p *PreisachModel) DiscreteStates(n int) []DiscreteState {
	if p == nil || p.material == nil || p.material.Ps <= 0 || math.IsNaN(p.material.Ps) || math.IsInf(p.material.Ps, 0) || !isValidDiscreteStateCount(n) {
		return nil
	}

	states := make([]DiscreteState, n)
	step := 2.0 * p.material.Ps / float64(n-1)
	for i := 0; i < n; i++ {
		pol := -p.material.Ps + float64(i)*step
		states[i] = DiscreteState{
			Level:        i + 1,
			Polarization: pol,
			NormalizedP:  pol / p.material.Ps,
			Voltage:      0, // Placeholder
			Conductance:  0, // Placeholder
		}
	}
	return states
}

// Reset clears the history and sets the model to negative saturation
// (including the reversible dielectric contribution at -E_sat).
func (p *PreisachModel) Reset() {
	if p == nil || p.material == nil || p.stack == nil || p.stack.Everett == nil || p.material.Ec <= 0 || math.IsNaN(p.material.Ec) || math.IsInf(p.material.Ec, 0) {
		return
	}

	// Re-initialize stack
	E_sat := p.saturationField()
	everett := p.stack.Everett
	newStack := physics.NewPreisachStack(E_sat, everett)
	if newStack == nil {
		return
	}
	p.stack = newStack
}

func (p *PreisachModel) saturationField() float64 {
	if p == nil {
		return 0
	}
	ec := 0.0
	if p.everett != nil && p.everett.Ec > 0 && !math.IsNaN(p.everett.Ec) && !math.IsInf(p.everett.Ec, 0) {
		ec = p.everett.Ec
	} else if p.material != nil && p.material.Ec > 0 && !math.IsNaN(p.material.Ec) && !math.IsInf(p.material.Ec, 0) {
		ec = p.material.Ec
	}
	return ec * saturationFieldMultiplier
}

type polarizationAdvanceMode int

const (
	advanceQuasiStatic polarizationAdvanceMode = iota
	advanceTimed
)

// Update applies a new electric field and returns the resulting polarization.
// This uses the quasi-static Preisach path: material-specific switching kinetics
// are reserved for TimeStep(), while Update() remains stable for loop generation
// and GUI field scrubbing.
func (p *PreisachModel) Update(E float64) float64 {
	return p.advancePolarization(E, quasiStaticDt, advanceQuasiStatic)
}

// TimeStep applies a constant electric field E for duration dt (seconds).
// Returns the resulting polarization after material-specific NLS relaxation.
func (p *PreisachModel) TimeStep(E, dt float64) float64 {
	return p.advancePolarization(E, dt, advanceTimed)
}

func (p *PreisachModel) advancePolarization(E, dt float64, mode polarizationAdvanceMode) float64 {
	if p == nil {
		return 0
	}
	if !p.canAdvancePolarization(E, dt, mode) {
		return p.fallbackPolarization()
	}

	P_start := p.advanceStartPolarization()
	Pirrev_target := p.stack.Update(E)
	P_target := Pirrev_target + p.reversiblePolarization(E)

	P_final := P_target
	if mode == advanceTimed {
		if p.nls == nil {
			p.nls = newNLSKineticsForMaterial(p.material)
		}
		P_final = p.nls.Relax(P_start, P_target, E, dt)
		if logging.IsVerbose(logging.VerbosityTrace) {
			log.Calculation("TimeStep", map[string]interface{}{
				"E": E, "dt": dt, "P_start": P_start, "P_target": P_target, "P_final": P_final,
			}, P_final)
		}
	} else {
		P_final = newQuasiStaticNLSForMaterial(p.material).Relax(P_start, P_target, E, quasiStaticDt)
	}

	if !p.lockDynamic {
		p.dynamicP = P_final
		p.hasDynamicP = true
	}
	return P_final
}

func (p *PreisachModel) canAdvancePolarization(E, dt float64, mode polarizationAdvanceMode) bool {
	if math.IsNaN(E) || math.IsInf(E, 0) {
		return false
	}
	if mode == advanceTimed && (dt <= 0 || math.IsNaN(dt) || math.IsInf(dt, 0)) {
		return false
	}
	return isValidPreisachMaterial(p.material) &&
		p.stack != nil &&
		p.stack.Everett != nil &&
		p.stack.SaturationE > 0 && !math.IsNaN(p.stack.SaturationE) && !math.IsInf(p.stack.SaturationE, 0) &&
		!math.IsNaN(p.stack.LastE) && !math.IsInf(p.stack.LastE, 0)
}

func (p *PreisachModel) advanceStartPolarization() float64 {
	if p.hasDynamicP && !p.lockDynamic {
		return p.dynamicP
	}
	return p.Polarization()
}

func (p *PreisachModel) fallbackPolarization() float64 {
	if p.hasDynamicP && !math.IsNaN(p.dynamicP) && !math.IsInf(p.dynamicP, 0) {
		return p.dynamicP
	}
	pol := p.Polarization()
	if math.IsNaN(pol) || math.IsInf(pol, 0) {
		return 0
	}
	return pol
}

// Polarization returns the current polarization state.
func (p *PreisachModel) Polarization() float64 {
	if p == nil || p.stack == nil || p.stack.Everett == nil || len(p.stack.Stack) == 0 || p.stack.SaturationE <= 0 || math.IsNaN(p.stack.SaturationE) || math.IsInf(p.stack.SaturationE, 0) || math.IsNaN(p.stack.LastE) || math.IsInf(p.stack.LastE, 0) {
		return 0
	}

	// Compute polarization at current field without mutating history.
	Pirrev := p.stack.ComputePolarization(p.stack.LastE)
	return Pirrev + p.reversiblePolarization(p.stack.LastE)
}

// NormalizedPolarization returns polarization as fraction of Ps.
func (p *PreisachModel) NormalizedPolarization() float64 {
	if p == nil || p.material == nil {
		return 0
	}

	denom := p.effectivePs
	if denom == 0 {
		denom = p.material.Ps
	}
	if denom <= 0 || math.IsNaN(denom) || math.IsInf(denom, 0) {
		return 0
	}
	if p.hasDynamicP {
		if math.IsNaN(p.dynamicP) || math.IsInf(p.dynamicP, 0) {
			return 0
		}
		return p.dynamicP / denom
	}
	pol := p.Polarization()
	if math.IsNaN(pol) || math.IsInf(pol, 0) {
		return 0
	}
	return pol / denom
}

// reversiblePolarization returns the nonlinear reversible (dielectric) contribution.
//
//	P_rev(E) = P_sat_rev * tanh(E / Ec)
//
// where P_sat_rev = chi * Ec (chi from eps0 * (epsilon_LF - 1)).
// The small-signal slope dP_rev/dE|_{E=0} = P_sat_rev / Ec = chi matches
// the material's low-frequency permittivity.
func (p *PreisachModel) reversiblePolarization(E float64) float64 {
	if p.reversiblePSat == 0 || p.everett == nil || p.everett.Ec <= 0 {
		return 0
	}
	return p.reversiblePSat * math.Tanh(E/p.everett.Ec)
}

func (p *PreisachModel) updateReversibleParams() {
	if p.material == nil || p.everett == nil {
		return
	}

	// Linear susceptibility from permittivity (low frequency preferred).
	const epsilon0 = 8.854e-12
	if p.material.EpsilonLF > 1 {
		p.reversibleChi = epsilon0 * (p.material.EpsilonLF - 1.0)
	} else if p.material.Epsilon > 1 {
		p.reversibleChi = epsilon0 * (p.material.Epsilon - 1.0)
	} else {
		p.reversibleChi = 0
	}

	if p.reversibleChi > 0 && p.everett.Ec > 0 {
		p.reversiblePSat = p.reversibleChi * p.everett.Ec
	} else {
		p.reversiblePSat = 0
	}

	// Split saturation into irreversible + reversible components so total Ps is preserved.
	totalPs := p.effectivePs
	if totalPs == 0 {
		totalPs = p.material.Ps
	}
	psIrrev := totalPs - p.reversiblePSat
	if psIrrev < 0 {
		psIrrev = 0
		p.reversiblePSat = totalPs
	}
	p.everett.Ps = psIrrev

	// Tune Delta so that remanent polarization matches material Pr.
	p.everett.Delta = tuneDeltaForPr(p.everett.Ec, p.saturationField(), p.everett.Ps, p.material.Pr)
}

// GetHysteresisLoop generates a full P-E hysteresis loop.
func (p *PreisachModel) GetHysteresisLoop(Emax float64, points int) ([]float64, []float64) {
	if p == nil || !isValidLoopPointCount(points) || !isRepresentableLoopField(Emax) || !isValidPreisachMaterial(p.material) || p.stack == nil || p.stack.Everett == nil || p.stack.SaturationE <= 0 || math.IsNaN(p.stack.SaturationE) || math.IsInf(p.stack.SaturationE, 0) || math.IsNaN(p.stack.LastE) || math.IsInf(p.stack.LastE, 0) {
		return nil, nil
	}

	loopModel := p.newLoopModel()
	if loopModel == nil {
		return nil, nil
	}

	E := make([]float64, 0, points*4)
	PVal := make([]float64, 0, points*4) // renamed to avoid collision with P() method

	// Saturation start
	loopModel.Update(-Emax)

	// Ascending
	for i := 0; i <= points*2; i++ {
		e := -Emax + 2*Emax*float64(i)/float64(points*2)
		pol := loopModel.Update(e)
		E = append(E, e)
		PVal = append(PVal, pol)
	}

	// Descending
	for i := 1; i <= points*2; i++ {
		e := Emax - 2*Emax*float64(i)/float64(points*2)
		pol := loopModel.Update(e)
		E = append(E, e)
		PVal = append(PVal, pol)
	}

	return E, PVal
}

func (p *PreisachModel) newLoopModel() *PreisachModel {
	loopEverett := &TanhEverett{
		Ps:    p.everett.Ps,
		Ec:    p.everett.Ec,
		Delta: p.everett.Delta,
	}
	loopStack := physics.NewPreisachStack(p.saturationField(), loopEverett)
	if loopStack == nil {
		return nil
	}
	loopNLS := physics.NewNLSKinetics()
	if p.nls != nil {
		*loopNLS = *p.nls
	}
	return &PreisachModel{
		material:       p.material,
		stack:          loopStack,
		everett:        loopEverett,
		Temperature:    p.Temperature,
		Stress:         p.Stress,
		reversibleChi:  p.reversibleChi,
		reversiblePSat: p.reversiblePSat,
		effectivePs:    p.effectivePs,
		nls:            loopNLS,
		lockDynamic:    true,
		hasDynamicP:    false,
	}
}

// SetTemperature updates the simulation temperature and scales material parameters.
func (p *PreisachModel) SetTemperature(tempK float64) {
	if p == nil {
		return
	}
	if _, _, ok := p.effectiveParametersFor(tempK, p.Stress); !ok {
		return
	}
	p.Temperature = tempK
	p.updateEffectiveParameters()
}

// GetEffectiveEc returns the current temperature-scaled Coercive Field.
func (p *PreisachModel) GetEffectiveEc() float64 {
	if p == nil || p.everett == nil || p.everett.Ec <= 0 || math.IsNaN(p.everett.Ec) || math.IsInf(p.everett.Ec, 0) {
		return 0
	}
	return p.everett.Ec
}

// SetStress updates the mechanical stress and scales Ec accordingly.
// Stress is in GPa.
// Scaling Logic: Ec ~ sqrt(|Alpha|)
// Alpha = AlphaT - 2*Q12*Stress
func (p *PreisachModel) SetStress(stressGPa float64) {
	if p == nil {
		return
	}
	if _, _, ok := p.effectiveParametersFor(p.Temperature, stressGPa); !ok {
		return
	}
	p.Stress = stressGPa

	// Recalculate everything (Temperature and Stress)
	p.updateEffectiveParameters()
}

// updateEffectiveParameters recalculates Ec and Ps based on Curie-Weiss
// temperature scaling and electrostriction stress coupling.
//
// Temperature: Ps(T) = Ps(300K) + TempCoeffPr * (T - 300)
// Coercive field: Ec(T,sigma) scales as |alpha(T,sigma)/alpha_ref|^1.5
//
//	where alpha(T) = (T - T_C) / (2*eps0*C)    (Curie-Weiss)
//	and   alpha(T,sigma) = alpha(T) - 2*Q12*sigma  (electrostriction)
//
// Q12 is the transverse electrostriction coefficient (default -0.026 for HZO).
func (p *PreisachModel) updateEffectiveParameters() {
	if p == nil || p.everett == nil {
		return
	}

	newEc, newPs, ok := p.effectiveParametersFor(p.Temperature, p.Stress)
	if !ok {
		return
	}

	p.everett.Ec = newEc
	p.effectivePs = newPs
	p.updateReversibleParams()
}

func (p *PreisachModel) effectiveParametersFor(tempK, stressGPa float64) (float64, float64, bool) {
	if p == nil || p.material == nil || tempK <= 0 || !isFinite(tempK) || !isFinite(stressGPa) {
		return 0, 0, false
	}

	const epsilon0 = 8.854e-12

	newEc := p.material.Ec
	newPs := p.material.Ps + p.material.TempCoeffPr*(tempK-roomTemperatureK)
	if !isFinite(newEc) || !isFinite(newPs) {
		return 0, 0, false
	}

	if p.material.CurieConst != 0 {
		denom := 2 * epsilon0 * p.material.CurieConst
		if denom == 0 || !isFinite(denom) {
			return 0, 0, false
		}

		alphaT := (tempK - p.material.CurieTemp) / denom
		alphaRef := (roomTemperatureK - p.material.CurieTemp) / denom
		if !isFinite(alphaT) || !isFinite(alphaRef) {
			return 0, 0, false
		}

		q12 := p.material.Q12
		if q12 == 0 {
			q12 = defaultQ12HZO
		}
		alphaRefStress := alphaRef - 2*q12*defaultStressGPa*1e9
		if !isFinite(alphaRefStress) {
			return 0, 0, false
		}
		if alphaRefStress != 0 {
			stressTerm := 2 * q12 * stressGPa * 1e9
			if !isFinite(stressTerm) {
				return 0, 0, false
			}
			alphaStress := alphaT - stressTerm // alpha(σ) = alpha(T) - 2*Q12*σ
			if !isFinite(alphaStress) {
				return 0, 0, false
			}
			ecRatio := math.Pow(math.Abs(alphaStress/alphaRefStress), 1.5)
			newEc = p.material.Ec * ecRatio
			if !isFinite(newEc) {
				return 0, 0, false
			}
		}
	}

	if newEc < 1e5 {
		newEc = 1e5
	}
	if newPs < 1e-6 {
		newPs = 1e-6
	}
	if !isRepresentableCoerciveField(newEc) || newPs <= 0 || !isFinite(newPs) {
		return 0, 0, false
	}

	return newEc, newPs, true
}
