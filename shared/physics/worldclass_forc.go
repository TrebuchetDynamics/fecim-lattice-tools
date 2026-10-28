package physics

import (
	"context"
	"fmt"
	"math"
)

// GUIFORCEverett is the simplified weighted Everett function used by the
// native GUI diagnostic, not the headless TanhEverett or an L-K experiment.
type GUIFORCEverett struct{ SaturationE, Ps float64 }

func (e GUIFORCEverett) Calculate(alpha, beta float64) float64 {
	if alpha <= beta || e.SaturationE <= 0 {
		return 0
	}
	x := (alpha - beta) / (2 * e.SaturationE)
	if x < 0 {
		x = 0
	}
	if x > 1 {
		x = 1
	}
	center := (alpha + beta) / (2 * e.SaturationE)
	weight := preisachExp(-(center * center) / (2 * 0.2 * 0.2))
	return e.Ps * x * weight
}

// FORCCurve is one first-order reversal branch at fixed reversal field Er.
type FORCCurve struct {
	ReversalField_Vm float64   `json:"reversal_field_vm"`
	AppliedField_Vm  []float64 `json:"applied_field_vm"`
	Polarization_Cm2 []float64 `json:"polarization_cm2"`
}

// FORCResult stores FORC sweep outputs and derived Preisach density grid.
type FORCResult struct {
	Emax_Vm float64 `json:"emax_vm"`

	// Monotonic grid from -Emax to +Emax used for reversal fields.
	ReversalFields_Vm []float64   `json:"reversal_fields_vm"`
	Curves            []FORCCurve `json:"curves"`

	// Preisach density rho(Ea, Eb) sampled on grid indices [reversal][applied].
	// Entries outside Ea<=Eb region are zero.
	PreisachDensity [][]float64  `json:"preisach_density"`
	ReversalPairs   [][2]float64 `json:"reversal_pairs_vm"`
}

// RunFORCSweep runs first-order reversal curves: for each reversal field Er,
// ascend from -Emax to Er, then descend to -Emax while recording P(E).
func RunFORCSweep(model *PreisachStack, Emax float64, numReversals int) (FORCResult, error) {
	if model == nil {
		return FORCResult{}, fmt.Errorf("model is nil")
	}
	if Emax <= 0 {
		return FORCResult{}, fmt.Errorf("Emax must be > 0")
	}
	if numReversals < 3 {
		return FORCResult{}, fmt.Errorf("numReversals must be >= 3")
	}

	reversalGrid := linspace(-Emax, Emax, numReversals)
	result := FORCResult{
		Emax_Vm:           Emax,
		ReversalFields_Vm: reversalGrid,
		Curves:            make([]FORCCurve, 0, numReversals),
	}

	for i := range reversalGrid {
		curve, err := RunFORCCurveContext(context.Background(), model, reversalGrid, i)
		if err != nil {
			return FORCResult{}, err
		}
		result.Curves = append(result.Curves, curve)
	}

	result.PreisachDensity, result.ReversalPairs = ComputeFORCDensity(result)
	return result, nil
}

// ComputeFORCDensity estimates Preisach distribution rho(Ea,Eb)
// via mixed second derivative:
//
//	rho = -0.5 * d²P/(dEa dEb)
//
// using central finite differences on the FORC grid.
func ComputeFORCDensity(results FORCResult) ([][]float64, [][2]float64) {
	density, pairs, _ := ComputeFORCDensityContext(context.Background(), results)
	return density, pairs
}

// RunFORCCurveContext is one resumable sweep unit: initialize at negative
// saturation, ascend to the reversal, then record the descending branch.
// grid is caller-owned and must remain unchanged during this call.
func RunFORCCurveContext(ctx context.Context, model *PreisachStack, grid []float64, index int) (FORCCurve, error) {
	if model == nil || index < 0 || index >= len(grid) {
		return FORCCurve{}, fmt.Errorf("invalid FORC curve input")
	}
	if err := ctx.Err(); err != nil {
		return FORCCurve{}, err
	}
	ps := NewPreisachStack(model.SaturationE, model.Everett)
	if ps == nil {
		return FORCCurve{}, fmt.Errorf("invalid FORC stack")
	}
	_ = ps.Update(grid[index])
	curve := FORCCurve{ReversalField_Vm: grid[index], AppliedField_Vm: make([]float64, 0, index+1), Polarization_Cm2: make([]float64, 0, index+1)}
	for j := index; j >= 0; j-- {
		if err := ctx.Err(); err != nil {
			return FORCCurve{}, err
		}
		ea := grid[j]
		p := ps.Update(ea)
		curve.AppliedField_Vm = append(curve.AppliedField_Vm, ea)
		curve.Polarization_Cm2 = append(curve.Polarization_Cm2, p)
	}
	return curve, nil
}

// ComputeFORCDensityContext checks cancellation while constructing and
// differentiating the grid. Cancelled matrices are never returned as complete.
func ComputeFORCDensityContext(ctx context.Context, results FORCResult) ([][]float64, [][2]float64, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	n := len(results.ReversalFields_Vm)
	density := make([][]float64, n)
	for i := range density {
		density[i] = make([]float64, n)
	}
	if n < 3 || len(results.Curves) != n {
		return density, nil, nil
	}

	dE := results.ReversalFields_Vm[1] - results.ReversalFields_Vm[0]
	if dE == 0 {
		return density, nil, nil
	}

	gridP := make([][]float64, n)
	for i := 0; i < n; i++ {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		gridP[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			gridP[i][j] = math.NaN()
		}
		curve := results.Curves[i]
		for k, ea := range curve.AppliedField_Vm {
			j := int(math.Round((ea - results.ReversalFields_Vm[0]) / dE))
			if j >= 0 && j < n && k < len(curve.Polarization_Cm2) {
				gridP[i][j] = curve.Polarization_Cm2[k]
			}
		}
	}

	pairs := make([][2]float64, 0, n*n/2)
	for i := 1; i < n-1; i++ {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		eb := results.ReversalFields_Vm[i]
		for j := 1; j < n-1; j++ {
			ea := results.ReversalFields_Vm[j]
			if ea > eb {
				continue
			}
			p11 := gridP[i+1][j+1]
			p10 := gridP[i+1][j-1]
			p01 := gridP[i-1][j+1]
			p00 := gridP[i-1][j-1]
			if math.IsNaN(p11) || math.IsNaN(p10) || math.IsNaN(p01) || math.IsNaN(p00) {
				continue
			}
			d2 := (p11 - p10 - p01 + p00) / (4 * dE * dE)
			rho := -0.5 * d2
			density[i][j] = rho
			pairs = append(pairs, [2]float64{ea, eb})
		}
	}

	return density, pairs, nil
}

func linspace(min, max float64, n int) []float64 {
	if n <= 1 {
		return []float64{min}
	}
	out := make([]float64, n)
	step := (max - min) / float64(n-1)
	for i := range out {
		out[i] = min + float64(i)*step
	}
	return out
}
