package physics

import (
	"hash/fnv"
	"math"
	"math/rand"
)

const (
	defaultPolydomainSigmaFrac = 0.15
	minPolydomainSigmaFrac     = 0.10
	maxPolydomainSigmaFrac     = 0.20
	maxPolydomainDomainCount   = 1_000_000
)

// PolydomainEnsemble approximates polycrystalline ferroelectric behavior by
// averaging many independent single-domain LK solvers with distributed coercive
// fields (Ec).
//
// Domain i uses a Gaussian Ec multiplier m_i around 1.0; this is applied as an
// effective field scaling E_i = E/m_i in Step(). A larger m_i means harder
// switching (higher coercive field).
type PolydomainEnsemble struct {
	Domains  []*LKSolver
	EcFactor []float64
	Imprint  []float64
	Seed     uint64
}

func deriveEnsembleSeed(mat *HZOMaterial, n int) uint64 {
	h := fnv.New64a()
	if mat != nil {
		_, _ = h.Write([]byte(mat.Name))
	}
	seed := h.Sum64() ^ uint64(n*0x9e3779b1)
	if seed == 0 {
		seed = 1
	}
	return seed
}

// NewPolydomainEnsemble builds a deterministic domain ensemble.
// sigmaFrac is clamped to [0.10, 0.20] to keep Ec spread physically bounded.
func NewPolydomainEnsemble(template *LKSolver, mat *HZOMaterial, n int, sigmaFrac float64, seed uint64) *PolydomainEnsemble {
	if template == nil || !isValidPolydomainMaterial(mat) || !isValidPolydomainDomainCount(n) || !template.isRepresentableLKMaterialLandauScaling(mat) {
		return nil
	}
	if invalidFloat(sigmaFrac) || sigmaFrac <= 0 {
		sigmaFrac = defaultPolydomainSigmaFrac
	}
	if sigmaFrac < minPolydomainSigmaFrac {
		sigmaFrac = minPolydomainSigmaFrac
	}
	if sigmaFrac > maxPolydomainSigmaFrac {
		sigmaFrac = maxPolydomainSigmaFrac
	}
	if seed == 0 {
		seed = deriveEnsembleSeed(mat, n)
	}

	rng := rand.New(rand.NewSource(int64(seed)))
	domains := make([]*LKSolver, 0, n)
	ecFactor := make([]float64, 0, n)
	imprint := make([]float64, 0, n)

	for i := 0; i < n; i++ {
		d := NewLKSolver()
		d.ConfigureFromMaterial(mat)
		applyPolydomainTemplateRuntimeState(d, template)
		if isValidPolydomainPMax(template.PMax) {
			d.PMax = template.PMax
		}
		d.rng = rand.New(rand.NewSource(int64(seed) + int64(i+1)*0x9e3779b97f4a7c1))

		factor := 1.0 + rng.NormFloat64()*sigmaFrac
		if factor < 0.4 {
			factor = 0.4
		} else if factor > 2.0 {
			factor = 2.0
		}
		bias := rng.NormFloat64() * 0.10 * mat.Ec

		domains = append(domains, d)
		ecFactor = append(ecFactor, factor)
		imprint = append(imprint, bias)
	}

	return &PolydomainEnsemble{Domains: domains, EcFactor: ecFactor, Imprint: imprint, Seed: seed}
}

func isValidPolydomainMaterial(mat *HZOMaterial) bool {
	return mat != nil && isPositiveFiniteLKMaterialValue(mat.Ec)
}

func isValidPolydomainDomainCount(n int) bool {
	return n > 1 && n <= maxPolydomainDomainCount
}

func (p *PolydomainEnsemble) SetState(P float64) {
	if p == nil || !isRepresentableLKPolarization(P) {
		return
	}
	for _, d := range p.Domains {
		if d == nil {
			continue
		}
		d.SetState(P)
		d.Time = 0
	}
}

func (p *PolydomainEnsemble) Step(template *LKSolver, E, dt float64) float64 {
	if p == nil || len(p.Domains) == 0 {
		return 0
	}
	if !p.isValidStepInput(E, dt) {
		return p.currentPolarizationMean()
	}
	sum := 0.0
	count := 0
	for i, d := range p.Domains {
		if d == nil {
			continue
		}
		applyPolydomainTemplateRuntimeState(d, template)

		polarization := d.Step(p.domainField(i, E), dt)
		if invalidFloat(polarization) {
			polarization = d.GetState()
		}
		if invalidFloat(polarization) || !isRepresentableLKPolarization(polarization) {
			continue
		}
		nextSum := sum + polarization
		if invalidFloat(nextSum) || !isRepresentableLKPolarization(nextSum) {
			continue
		}
		sum = nextSum
		count++
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

func (p *PolydomainEnsemble) isValidStepInput(E, dt float64) bool {
	if invalidFloat(E) || invalidFloat(dt) || dt <= 0 || dt > maxLKTimestep {
		return false
	}
	sawDomain := false
	for i, d := range p.Domains {
		if d == nil {
			continue
		}
		sawDomain = true
		field := p.domainField(i, E)
		if invalidFloat(field) || !isRepresentableLKRuntimeForceTerm(field, d.effectiveRho()) {
			return false
		}
	}
	return sawDomain
}

func (p *PolydomainEnsemble) domainField(index int, E float64) float64 {
	factor := 1.0
	if index < len(p.EcFactor) && isValidPolydomainScale(p.EcFactor[index]) {
		factor = p.EcFactor[index]
	}
	bias := 0.0
	if index < len(p.Imprint) && !invalidFloat(p.Imprint[index]) {
		bias = p.Imprint[index]
	}
	return E/factor + bias
}

func (p *PolydomainEnsemble) currentPolarizationMean() float64 {
	sum := 0.0
	count := 0
	for _, d := range p.Domains {
		if d == nil || invalidFloat(d.P) || !isRepresentableLKPolarization(d.P) {
			continue
		}
		nextSum := sum + d.P
		if invalidFloat(nextSum) || !isRepresentableLKPolarization(nextSum) {
			continue
		}
		sum = nextSum
		count++
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

func applyPolydomainTemplateRuntimeState(domain, template *LKSolver) {
	if domain == nil || template == nil {
		return
	}
	domain.EnableNoise = template.EnableNoise
	domain.UseNLS = template.UseNLS
	if isValidPolydomainRuntimeTemperature(domain, template.Temperature) {
		domain.Temperature = template.Temperature
	}
	if isValidPolydomainRuntimeStress(domain, template.Stress) {
		domain.Stress = template.Stress
	}
}

func isValidPolydomainRuntimeTemperature(domain *LKSolver, value float64) bool {
	if domain == nil {
		return !invalidFloat(value)
	}
	_, ok := domain.runtimeAlphaFor(value, domain.Stress)
	return ok
}

func isValidPolydomainRuntimeStress(domain *LKSolver, value float64) bool {
	if domain == nil {
		return !invalidFloat(value)
	}
	_, ok := domain.runtimeAlphaFor(domain.Temperature, value)
	return ok
}

func isValidPolydomainPMax(value float64) bool {
	return value > 0 && isRepresentableLKPolarization(value)
}

func isValidPolydomainScale(value float64) bool {
	return value > 0 && !invalidFloat(value)
}

func (p *PolydomainEnsemble) DomainCount() int {
	if p == nil {
		return 0
	}
	return len(p.Domains)
}

func (p *PolydomainEnsemble) RemanentSpread(ps float64) float64 {
	if p == nil || len(p.Domains) == 0 || ps == 0 || invalidFloat(ps) {
		return 0
	}
	mean := 0.0
	count := 0
	for _, d := range p.Domains {
		if d == nil || invalidFloat(d.P) || !isRepresentableLKPolarization(d.P) {
			continue
		}
		nextMean := mean + d.P
		if invalidFloat(nextMean) || !isRepresentableLKPolarization(nextMean) {
			continue
		}
		mean = nextMean
		count++
	}
	if count == 0 {
		return 0
	}
	mean /= float64(count)
	spread := math.Abs(mean / ps)
	if invalidFloat(spread) {
		return 0
	}
	return spread
}
