package physics

import (
	"math"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// precisionStressEverett intentionally introduces a large common-mode offset so
// polarization evaluation must repeatedly cancel large terms to recover a small
// residual. This is a numerically sensitive setup by design.
type precisionStressEverett struct {
	offset float64
}

func (e precisionStressEverett) Calculate(alpha, beta float64) float64 {
	return e.offset + alpha - beta
}

type linearSymmetricEverett struct{}

func (linearSymmetricEverett) Calculate(alpha, beta float64) float64 {
	return 0.5 * (alpha - beta)
}

func computePolarizationNaive(ps *PreisachStack, currentE float64) float64 {
	sum := 0.0
	n := len(ps.Stack)
	if n == 1 {
		sum += ps.Everett.Calculate(currentE, ps.Stack[0].E)
		return -ps.Everett.Calculate(ps.SaturationE, -ps.SaturationE) + 2.0*sum
	}
	for i := 1; i < n; i += 2 {
		maxVal := ps.Stack[i].E
		minPrev := ps.Stack[i-1].E
		sum += ps.Everett.Calculate(maxVal, minPrev)
		if i+1 < n {
			sum -= ps.Everett.Calculate(maxVal, ps.Stack[i+1].E)
		} else {
			sum -= ps.Everett.Calculate(maxVal, currentE)
		}
	}
	return -ps.Everett.Calculate(ps.SaturationE, -ps.SaturationE) + 2.0*sum
}

func computePolarizationBig(ps *PreisachStack, currentE float64) float64 {
	prec := uint(256)
	newF := func(v float64) *big.Float { return new(big.Float).SetPrec(prec).SetFloat64(v) }

	sum := new(big.Float).SetPrec(prec).SetFloat64(0)
	offset := newF(ps.Everett.(precisionStressEverett).offset)
	calc := func(alpha, beta float64) *big.Float {
		v := new(big.Float).SetPrec(prec).Set(offset)
		v.Add(v, newF(alpha))
		v.Sub(v, newF(beta))
		return v
	}

	n := len(ps.Stack)
	if n == 1 {
		sum.Add(sum, calc(currentE, ps.Stack[0].E))
	} else {
		for i := 1; i < n; i += 2 {
			maxVal := ps.Stack[i].E
			minPrev := ps.Stack[i-1].E
			sum.Add(sum, calc(maxVal, minPrev))
			if i+1 < n {
				sum.Sub(sum, calc(maxVal, ps.Stack[i+1].E))
			} else {
				sum.Sub(sum, calc(maxVal, currentE))
			}
		}
	}

	psat := calc(ps.SaturationE, -ps.SaturationE)
	p := new(big.Float).SetPrec(prec).Neg(psat)
	twoSum := new(big.Float).SetPrec(prec).Mul(newF(2), sum)
	p.Add(p, twoSum)
	out, _ := p.Float64()
	return out
}

func buildLongCyclingStack(t *testing.T, updates int) *PreisachStack {
	t.Helper()
	ps := NewPreisachStack(1.0, precisionStressEverett{offset: 1e15})

	// Nested shrinking excursions create long turning-point histories without
	// repeatedly wiping out prior minor loops.
	for i := 0; i < updates; i++ {
		a := 1.0 - float64(i+1)/float64(updates+2)
		if i%2 == 0 {
			ps.Update(+a)
		} else {
			ps.Update(-a)
		}
	}
	if len(ps.Stack) < 100 {
		t.Fatalf("expected long history stack, got len=%d", len(ps.Stack))
	}
	return ps
}

func TestPreisachPrecision_LongCyclingNoCatastrophicCancellation(t *testing.T) {
	ps := buildLongCyclingStack(t, 2000)
	currentE := 0.123456789

	p := ps.ComputePolarization(currentE)
	pref := computePolarizationBig(ps, currentE)
	err := math.Abs(p - pref)

	if math.IsNaN(p) || math.IsInf(p, 0) {
		t.Fatalf("polarization became non-finite after long cycling: P=%v", p)
	}
	if err > 1e-9 {
		t.Fatalf("catastrophic cancellation suspected after long cycling: |P-ref|=%e, P=%0.17g, ref=%0.17g, stack=%d", err, p, pref, len(ps.Stack))
	}
}

func TestPreisachPrecision_CompensatedSummationUsed(t *testing.T) {
	srcPath := filepath.Join("preisach.go")
	b, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("failed reading %s: %v", srcPath, err)
	}
	s := string(b)

	if !strings.Contains(s, "compensation") {
		t.Fatalf("expected compensated summation variable in ComputePolarization")
	}
	if !strings.Contains(s, "y := v - compensation") || !strings.Contains(s, "compensation = (t - sum) - y") {
		t.Fatalf("expected Kahan/compensated update steps in ComputePolarization")
	}
}

func TestPreisachPrecision_SymmetricPolarization(t *testing.T) {
	everett := linearSymmetricEverett{}
	for _, e := range []float64{0.1, 0.25, 0.5, 0.75, 0.95} {
		psPos := NewPreisachStack(1.0, everett)
		psNeg := NewPreisachStack(1.0, everett)

		pPos := psPos.Update(e)
		pNeg := psNeg.Update(-e)
		if math.Abs(pPos+pNeg) > 1e-10 {
			t.Fatalf("symmetry violated at E=%0.3f: P(E)+P(-E)=%e (P(E)=%0.17g, P(-E)=%0.17g)", e, math.Abs(pPos+pNeg), pPos, pNeg)
		}
	}
}
