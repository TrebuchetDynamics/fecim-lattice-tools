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
	if template == nil || mat == nil || n <= 1 {
		return nil
	}
	if sigmaFrac <= 0 {
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
		d.EnableNoise = template.EnableNoise
		d.UseNLS = template.UseNLS
		d.Temperature = template.Temperature
		d.Stress = template.Stress
		if template.PMax > 0 {
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

func (p *PolydomainEnsemble) SetState(P float64) {
	for _, d := range p.Domains {
		d.SetState(P)
		d.Time = 0
	}
}

func (p *PolydomainEnsemble) Step(template *LKSolver, E, dt float64) float64 {
	if p == nil || len(p.Domains) == 0 {
		return 0
	}
	sum := 0.0
	for i, d := range p.Domains {
		d.Temperature = template.Temperature
		d.Stress = template.Stress
		d.UseNLS = template.UseNLS
		d.EnableNoise = template.EnableNoise

		factor := 1.0
		if i < len(p.EcFactor) {
			factor = p.EcFactor[i]
		}
		if factor == 0 {
			factor = 1
		}
		bias := 0.0
		if i < len(p.Imprint) {
			bias = p.Imprint[i]
		}
		sum += d.Step(E/factor+bias, dt)
	}
	return sum / float64(len(p.Domains))
}

func (p *PolydomainEnsemble) DomainCount() int {
	if p == nil {
		return 0
	}
	return len(p.Domains)
}

func (p *PolydomainEnsemble) RemanentSpread(ps float64) float64 {
	if p == nil || len(p.Domains) == 0 || ps == 0 {
		return 0
	}
	mean := 0.0
	for _, d := range p.Domains {
		mean += d.P
	}
	mean /= float64(len(p.Domains))
	return math.Abs(mean / ps)
}
