// Package crossbar is the CPU-only, memory-first crossbar web port. Scientific
// algorithms and device mappings remain in shared/crossbar.
package crossbar

import (
	"context"
	"encoding/json"
	"errors"
	core "fecim-lattice-tools/shared/crossbar"
	"fmt"
	"math"
	"math/rand"
)

const ModelVersion = "crossbar-cpu/1"

type Config struct {
	Rows         int     `json:"rows"`
	Cols         int     `json:"cols"`
	Profile      string  `json:"profile"`
	Levels       int     `json:"levels"`
	Seed         int64   `json:"seed"`
	Noise        float64 `json:"noise"`
	ADCBits      int     `json:"adc_bits"`
	DACBits      int     `json:"dac_bits"`
	Temperature  float64 `json:"temperature"`
	Architecture string  `json:"architecture"`
	IR           bool    `json:"ir"`
	Sneak        bool    `json:"sneak"`
	RpRow        float64 `json:"rp_row"`
	RpCol        float64 `json:"rp_col"`
}

func DefaultConfig() Config {
	return Config{Rows: 8, Cols: 8, Profile: "enhanced", Levels: 30, Seed: 42, ADCBits: 16, DACBits: 16, Temperature: 300, Architecture: "0T1R", IR: true, RpRow: 0.001, RpCol: 0.001}
}
func (c Config) validate(limit int) error {
	if limit != 32 && limit != 128 {
		return fmt.Errorf("invalid execution profile limit")
	}
	if c.Rows < 1 || c.Cols < 1 || c.Rows > limit || c.Cols > limit {
		return fmt.Errorf("array dimensions must be 1 through %d before allocation", limit)
	}
	if c.Profile != "basic" && c.Profile != "enhanced" && c.Profile != "parasitic" && c.Profile != "fecap" {
		return fmt.Errorf("unknown CPU model profile")
	}
	if c.Levels < 2 || c.Levels > 256 {
		return fmt.Errorf("programming levels must be 2 through 256")
	}
	if c.ADCBits < 1 || c.ADCBits > 24 || c.DACBits < 1 || c.DACBits > 24 {
		return fmt.Errorf("converter resolution must be 1 through 24 bits")
	}
	if !finite(c.Noise) || c.Noise < 0 || c.Noise > 0.2 || !finite(c.Temperature) || c.Temperature < 200 || c.Temperature > 500 {
		return fmt.Errorf("noise must be [0,0.2], temperature [200,500] K and finite")
	}
	if c.Architecture != "0T1R" && c.Architecture != "1T1R" && c.Architecture != "2T1R" {
		return fmt.Errorf("unknown isolation architecture")
	}
	if !finite(c.RpRow) || !finite(c.RpCol) || c.RpRow < 0 || c.RpCol < 0 || c.RpRow > 0.01 || c.RpCol > 0.01 {
		return fmt.Errorf("parasitic ratios must be finite in [0,0.01]")
	}
	// Full three-cell sneak enumeration has a stricter measured work envelope.
	if c.Sneak && (c.Rows > 32 || c.Cols > 32) {
		return fmt.Errorf("sneak analysis is capped at 32 by 32; disable sneak for larger local arrays")
	}
	return nil
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func validateData(c Config, matrix [][]float64, input []float64) error {
	if len(matrix) != c.Rows || len(input) != c.Cols {
		return fmt.Errorf("exact rows by columns matrix and columns-element input required")
	}
	for _, row := range matrix {
		if len(row) != c.Cols {
			return fmt.Errorf("ragged matrix")
		}
		for _, v := range row {
			if !finite(v) || v < 0 || v > 1 {
				return fmt.Errorf("matrix states must be finite in [0,1]")
			}
		}
	}
	for _, v := range input {
		if !finite(v) || v < 0 || v > 1 {
			return fmt.Errorf("input voltages must be finite in [0,1] V")
		}
	}
	return nil
}
func SeedMatrix(c Config) [][]float64 {
	rng := rand.New(rand.NewSource(c.Seed))
	m := make([][]float64, c.Rows)
	for i := range m {
		m[i] = make([]float64, c.Cols)
		for j := range m[i] {
			m[i][j] = rng.Float64()
		}
	}
	return m
}
func SeedInput(c Config) []float64 {
	rng := rand.New(rand.NewSource(c.Seed + 1))
	v := make([]float64, c.Cols)
	for i := range v {
		v[i] = rng.Float64()
	}
	return v
}

type Result struct {
	Version       string                    `json:"version"`
	Config        Config                    `json:"config"`
	Matrix        [][]float64               `json:"matrix"`
	Physical      [][]float64               `json:"physical"`
	PhysicalUnit  string                    `json:"physical_unit"`
	Input         []float64                 `json:"input"`
	Ideal         []float64                 `json:"ideal"`
	Actual        []float64                 `json:"actual"`
	Unit          string                    `json:"unit"`
	Normalization string                    `json:"normalization"`
	IR            *core.IRDropAnalysis      `json:"ir,omitempty"`
	Sneak         *core.SneakPathAnalysis   `json:"sneak,omitempty"`
	Trace         *core.MVMSneakTraceReport `json:"trace,omitempty"`
	Solver        *core.ParasiticMVMResult  `json:"solver,omitempty"`
	RMSE          float64                   `json:"rmse"`
}

func (r Result) Canonical() ([]byte, error) { return json.Marshal(r) }
func transpose(m [][]float64) [][]float64 {
	out := make([][]float64, len(m[0]))
	for j := range out {
		out[j] = make([]float64, len(m))
		for i := range m {
			out[j][i] = m[i][j]
		}
	}
	return out
}
func clone(m [][]float64) [][]float64 {
	out := make([][]float64, len(m))
	for i := range m {
		out[i] = append([]float64(nil), m[i]...)
	}
	return out
}

// Evaluate admits dimensions/work before allocating the core. Every profile uses
// columns as inputs; normalized/charge profiles return rows and the Kirchhoff
// solver returns native column sums. FeCAP explicitly transposes into its
// native row-drive/column-charge orientation, never silently padding a vector.
func Evaluate(ctx context.Context, c Config, matrix [][]float64, input []float64, limit int, yield func(context.Context) error) (Result, error) {
	var out Result
	if err := ctx.Err(); err != nil {
		return out, err
	}
	if err := c.validate(limit); err != nil {
		return out, err
	}
	if err := validateData(c, matrix, input); err != nil {
		return out, err
	}
	step := func() error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if yield != nil {
			return yield(ctx)
		}
		return nil
	}
	if err := step(); err != nil {
		return out, err
	}
	cfg := &core.Config{Rows: c.Rows, Cols: c.Cols, ADCBits: c.ADCBits, DACBits: c.DACBits, NoiseLevel: c.Noise, QuantizationLevels: c.Levels}
	if c.Profile == "fecap" {
		cfg = core.DefaultFeCAPConfig(c.Cols, c.Rows)
		cfg.ADCBits = c.ADCBits
		cfg.DACBits = c.DACBits
		cfg.NoiseLevel = c.Noise
	}
	a, err := core.NewArraySeeded(cfg, c.Seed)
	if err != nil {
		return out, err
	}
	defer a.Destroy()
	out = Result{Version: ModelVersion, Config: c, Input: append([]float64(nil), input...), Unit: "ratio", PhysicalUnit: "S", Normalization: "Sum of normalized cell state times DAC-quantized input, divided by columns, ADC-quantized; not amperes."}
	if c.Profile == "fecap" {
		if err = a.ProgramCapacitanceMatrix(transpose(matrix)); err != nil {
			return out, err
		}
		out.Matrix = transpose(a.GetCapacitanceMatrix())
		out.Physical = transpose(a.GetEffectiveCapacitanceMatrix())
		out.PhysicalUnit = "F"
		out.Actual, err = a.MVMCharge(input)
		idealCfg := *cfg
		idealCfg.NoiseLevel = 0
		ideal, e := core.NewArraySeeded(&idealCfg, c.Seed)
		if e != nil {
			return Result{}, e
		}
		defer ideal.Destroy()
		if e = ideal.ProgramCapacitanceMatrix(transpose(matrix)); e != nil {
			return Result{}, e
		}
		if err == nil {
			out.Ideal, err = ideal.MVMCharge(input)
		}
		out.Unit = "C"
		out.Normalization = "Raw charge Q = sum(C physical times DAC-quantized row-drive voltage), coulombs. Core transposed to preserve web column-input orientation; no DC IR/sneak model."
	} else {
		if err = a.ProgramWeightMatrix(matrix); err != nil {
			return out, err
		}
		out.Matrix = a.GetConductanceMatrix()
		out.Physical = physicalConductances(a)
		switch c.Profile {
		case "basic":
			idealCfg := *cfg
			idealCfg.NoiseLevel = 0
			ideal, _ := core.NewArraySeeded(&idealCfg, c.Seed)
			defer ideal.Destroy()
			ideal.ProgramWeightMatrix(matrix)
			out.Ideal, err = ideal.MVM(input)
			if err == nil {
				out.Actual, err = a.MVM(input)
			}
		case "enhanced":
			r, e := a.MVMWithNonIdealities(input, &core.MVMOptions{EnableIRDrop: c.IR, EnableSneakPaths: c.Sneak, EnableVariation: true, Temperature: c.Temperature, Architecture: c.Architecture, Yield: step})
			err = e
			if r != nil {
				out.Ideal = r.IdealOutput
				out.Actual = r.ActualOutput
				out.IR = r.IRDropAnalysis
				out.Sneak = r.SneakPathAnalysis
				out.Trace = r.SneakTrace
			}
			out.Normalization = "Enhanced core: ideal normalized dot product; actual includes selected IR/sneak/variation, DAC and ADC quantization, divided by columns. Not amperes."
		case "parasitic":
			solver, e := core.NewParasiticSolver(c.Rows, c.Cols, nil)
			if e != nil {
				return out, e
			}
			solver.SetConductances(out.Matrix)
			solver.SetParasitics(c.RpRow, c.RpCol)
			out.Ideal = solver.ComputeIdealMVM(input)
			out.Solver, err = solver.SolveMVMContext(ctx, input, step)
			// Nonconvergence is scientific evidence, not a transport failure.
			if out.Solver != nil && (errors.Is(err, core.ErrMaxIterationsExceeded) || errors.Is(err, core.ErrConvergenceFailed)) {
				err = nil
			}
			if out.Solver != nil {
				out.Actual = out.Solver.OutputCurrents
			}
			out.Normalization = "SOR Kirchhoff network with normalized conductance and Rp/Rmin wire ratios; native column current sums (one output per column) in normalized units, no ADC/DAC or physical-ampere claim. Noise/temperature/isolation toggles do not apply."
		}
	}
	if err != nil {
		return Result{}, err
	}
	if err = step(); err != nil {
		return Result{}, err
	}
	for i, v := range out.Actual {
		if !finite(v) || !finite(out.Ideal[i]) {
			return Result{}, fmt.Errorf("core produced nonfinite output")
		}
		d := v - out.Ideal[i]
		out.RMSE += d * d
	}
	out.RMSE = math.Sqrt(out.RMSE / float64(len(out.Actual)))
	// Marshal is also a finite-state guard for all nested cell/convergence values.
	if _, err = out.Canonical(); err != nil {
		return Result{}, err
	}
	return out, nil
}
