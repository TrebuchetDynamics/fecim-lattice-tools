// Package circuits is the persistent, CPU-only circuit workbench. Science lives
// in the same arraysim/material/peripheral helpers used by the native GUI.
package circuits

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	"fecim-lattice-tools/shared/peripherals"
	physics "fecim-lattice-tools/shared/physics"
)

const ModelVersion = "circuits-device/1"

var Sizes = []int{1, 2, 4, 8, 16, 32, 64, 128}

type Config struct {
	Rows          int     `json:"rows"`
	Cols          int     `json:"cols"`
	Material      string  `json:"material"`
	Node          int     `json:"node"`
	Architecture  string  `json:"architecture"`
	Coupling      string  `json:"coupling"`
	DACBits       int     `json:"dac_bits"`
	ADCBits       int     `json:"adc_bits"`
	Gain          float64 `json:"gain"`
	Supply        float64 `json:"supply"`
	SenseMin      float64 `json:"sense_min"`
	SenseMax      float64 `json:"sense_max"`
	MaxIterations int     `json:"max_iterations"`
	MaxPulses     int     `json:"max_pulses"`
	Seed          int64   `json:"seed"`
}

func DefaultConfig() Config {
	return Config{8, 8, "FeCIM", 130, "0T1R", "TierA", 5, 6, 1e4, 1.8, 0, 1, 4000, 40, 42}
}
func Material(name string) *physics.HZOMaterial {
	switch name {
	case "FeCIM":
		return physics.FeCIMMaterial()
	case "HZO":
		return physics.DefaultHZO()
	case "BTO":
		return physics.BTO()
	case "PZT":
		return physics.PZT()
	}
	return nil
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func size(n int) bool {
	for _, v := range Sizes {
		if v == n {
			return true
		}
	}
	return false
}
func (c Config) Validate(limit int) error {
	if !size(c.Rows) || !size(c.Cols) || c.Rows > limit || c.Cols > limit {
		return fmt.Errorf("array size requires native powers of two 1–%d; hosted coupled CPU cap is 8", limit)
	}
	if Material(c.Material) == nil {
		return errors.New("unknown material preset")
	}
	if c.Node != 130 && c.Node != 65 && c.Node != 28 && c.Node != 14 {
		return errors.New("node must be 130, 65, 28 or 14 nm")
	}
	if c.Architecture != "0T1R" && c.Architecture != "1T1R" && c.Architecture != "2T1R" {
		return errors.New("unknown architecture")
	}
	if c.Coupling != "Ideal" && c.Coupling != "TierA" && c.Coupling != "TierB" {
		return errors.New("unknown coupling")
	}
	// Tier A is the native dense reference (2*rows*cols nodes), not an
	// iterative approximation. Keep its allocation/cubic work bounded; larger
	// local sizes retain Ideal and sparse, cancellable Tier B.
	if c.Coupling == "TierA" && (c.Rows > 8 || c.Cols > 8) {
		return errors.New("dense TierA is limited to 8 by 8; choose Ideal or sparse TierB for larger local arrays")
	}
	if c.DACBits < 4 || c.DACBits > 8 || c.ADCBits < 5 || c.ADCBits > 8 {
		return errors.New("DAC domain 4–8 bits; ADC domain 5–8 bits")
	}
	if !finite(c.Gain) || c.Gain < 1e3 || c.Gain > 1e6 || !finite(c.Supply) || c.Supply < 1 || c.Supply > 3.3 {
		return errors.New("TIA gain requires 1000–1000000 ohms; supply 1–3.3 volts")
	}
	if !finite(c.SenseMin) || !finite(c.SenseMax) || c.SenseMin < 0 || c.SenseMax > 1.5 || c.SenseMax-c.SenseMin < 0.001 {
		return errors.New("sense references require 0–1.5 volts and at least 0.001 volt span")
	}
	if c.MaxIterations < 1 || c.MaxIterations > 4000 || c.MaxPulses < 1 || c.MaxPulses > 40 {
		return errors.New("solver requires 1–4000 iterations; circuit ISPP requires 1–40 pulses")
	}
	return nil
}

// DeviceSettings follows native material/node geometry, selector and voltage assumptions.
func (c Config) DeviceSettings() arraysim.DeviceSettings {
	mat := Material(c.Material)
	tech := physics.TechnologyNodeFromName(fmt.Sprintf("%dnm", c.Node))
	geom := arraysim.CellGeometry{PitchX: tech.CellPitchX, PitchY: tech.CellRowHeight, WireWidth: tech.MetalWidth, WireThickness: tech.MetalThickness, MetalResistivity: tech.MetalResistivity, Film: physics.GeometryFromMaterial(mat)}
	ron, leak := arraysim.NodeSelectorParams(c.Node)
	tia, adc, dac := peripherals.DefaultTIA(), peripherals.DefaultADC(), peripherals.DefaultDAC()
	tia.Gain = c.Gain
	adc.Bits = c.ADCBits
	adc.VrefLow = c.SenseMin
	adc.VrefHigh = c.SenseMax
	dac.Bits = c.DACBits
	coupling := arraysim.CouplingIdeal
	if c.Coupling == "TierA" {
		coupling = arraysim.CouplingTierA
	}
	if c.Coupling == "TierB" {
		coupling = arraysim.CouplingTierB
	}
	readMax, _ := arraysim.DeviceVoltageLimits(mat)
	return arraysim.DeviceSettings{Material: mat, Geometry: geom, Coupling: coupling, Passive: c.Architecture == "0T1R", TIA: tia, ADC: adc, DAC: dac, Temperature: 300, ReadMax: readMax, SelectorEnabled: true, SelectorRon: ron, SelectorLeakage: leak, MaxIterations: c.MaxIterations}
}

type State struct {
	Weights    [][]int   `json:"weights"`
	Input      []int     `json:"input"`
	Voltages   []float64 `json:"voltages,omitempty"`
	ActiveRows []bool    `json:"active_rows"`
	ReadMask   [][]bool  `json:"read_mask,omitempty"`
	WriteMask  [][]bool  `json:"write_mask,omitempty"`
	Row        int       `json:"row"`
	Col        int       `json:"col"`
}

func Initial(c Config, random bool) State {
	s := State{Weights: make([][]int, c.Rows), Input: make([]int, c.Cols), ActiveRows: make([]bool, c.Rows)}
	rng := rand.New(rand.NewSource(c.Seed))
	levels := Material(c.Material).GetNumLevels()
	for r := range s.Weights {
		s.Weights[r] = make([]int, c.Cols)
		s.ActiveRows[r] = true
		for col := range s.Weights[r] {
			if random {
				s.Weights[r][col] = rng.Intn(levels)
			}
		}
	}
	for i := range s.Input {
		s.Input[i] = 255
	}
	return s
}
func clone(s State) State {
	data, _ := json.Marshal(s)
	var out State
	_ = json.Unmarshal(data, &out)
	return out
}
func validateState(c Config, s State) error {
	if len(s.Weights) != c.Rows || len(s.Input) != c.Cols || len(s.ActiveRows) != c.Rows || s.Row < 0 || s.Row >= c.Rows || s.Col < 0 || s.Col >= c.Cols {
		return errors.New("exact array/vector/mask dimensions and valid selected cell required")
	}
	levels := Material(c.Material).GetNumLevels()
	if s.Voltages != nil {
		readMax, _ := arraysim.DeviceVoltageLimits(Material(c.Material))
		if len(s.Voltages) != c.Cols {
			return errors.New("manual DAC voltages must exactly match columns")
		}
		for _, v := range s.Voltages {
			if !finite(v) || math.Abs(v) > readMax {
				return errors.New("manual DAC voltage exceeds material read range")
			}
		}
	}
	for _, row := range s.Weights {
		if len(row) != c.Cols {
			return errors.New("ragged array")
		}
		for _, v := range row {
			if v < 0 || v >= levels {
				return errors.New("cell level outside material domain")
			}
		}
	}
	if c.Architecture == "0T1R" {
		for _, on := range s.ActiveRows {
			if !on {
				return errors.New("0T1R always enables every word line; use 1T1R or 2T1R for row masks")
			}
		}
	}
	for _, v := range s.Input {
		if v < -255 || v > 255 {
			return errors.New("signed input codes require -255 through 255")
		}
	}
	for _, mask := range [][][]bool{s.ReadMask, s.WriteMask} {
		if mask == nil {
			continue
		}
		if c.Architecture != "2T1R" || len(mask) != c.Rows {
			return errors.New("cell masks require 2T1R and exact dimensions")
		}
		for _, row := range mask {
			if len(row) != c.Cols {
				return errors.New("ragged cell mask")
			}
		}
	}
	return nil
}

type Pulse struct {
	Iteration, Level, Code        int
	Requested, Applied, Effective float64
}
type Result struct {
	Config    Config                `json:"config"`
	Operation string                `json:"operation"`
	State     State                 `json:"state"`
	Device    arraysim.DeviceResult `json:"device"`
	Pulses    []Pulse               `json:"pulses"`
	Verified  bool                  `json:"verified"`
	Reason    string                `json:"reason"`
}

func (r Result) Canonical() string { b, _ := json.Marshal(r); return string(b) }
func setup(c Config, s State, operation string) arraysim.DeviceSettings {
	d := c.DeviceSettings()
	d.Read = operation == "read"
	d.Write = operation == "write"
	d.ActiveRows = append([]bool(nil), s.ActiveRows...)
	d.WL = make([]float64, c.Rows)
	d.BL = make([]float64, c.Cols)
	d.ReadMask, d.WriteMask = s.ReadMask, s.WriteMask
	if operation == "read" {
		for r := range d.ActiveRows {
			d.ActiveRows[r] = d.Passive || r == s.Row
		}
		d.BL[s.Col] = arraysim.SafeDeviceReadVoltage(d.ReadMax)
	}
	if operation == "compute" {
		for col, code := range s.Input {
			d.BL[col] = float64(code) / 255 * d.ReadMax
			if s.Voltages != nil {
				d.BL[col] = s.Voltages[col]
			}
		}
	}
	if operation == "write" {
		for r := range d.ActiveRows {
			d.ActiveRows[r] = d.Passive || r == s.Row
		}
	}
	return d
}

// Execute owns a private state copy. Callers commit only after success and a final
// cancellation check; nonconverged ISPP retains actual pulses but never says verified.
func Execute(ctx context.Context, c Config, state State, operation string, target, limit int, yield func(context.Context) error) (Result, error) {
	if err := c.Validate(limit); err != nil {
		return Result{}, err
	}
	if err := validateState(c, state); err != nil {
		return Result{}, err
	}
	if operation != "read" && operation != "compute" && operation != "write" {
		return Result{}, errors.New("unknown circuit operation")
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	s := clone(state)
	d := setup(c, s, operation)
	levels := d.Material.GetNumLevels()
	out := Result{Config: c, Operation: operation, State: s, Pulses: []Pulse{}}
	if operation == "write" {
		if target < 0 || target >= levels {
			return Result{}, errors.New("target outside material levels")
		}
		if s.WriteMask != nil && !s.WriteMask[s.Row][s.Col] {
			return Result{}, errors.New("selected cell is blocked by write mask")
		}
		current := s.Weights[s.Row][s.Col]
		calc := physics.NewISPPCalculator(d.Material.CoerciveVoltage(), levels)
		direction := physics.GetDirection(current, target)
		_, maxV := arraysim.DeviceVoltageLimits(d.Material)
		calibrated := -maxV + float64(target)*(2*maxV/float64(levels-1))
		voltage := calc.CalculateStartVoltage(calibrated)
		for i := 0; i < c.MaxPulses && current != target; i++ {
			if yield != nil {
				if err := yield(ctx); err != nil {
					return Result{}, err
				}
			}
			if err := ctx.Err(); err != nil {
				return Result{}, err
			}
			applied, code := arraysim.QuantizeDeviceWrite(d, voltage)
			for col := range d.BL {
				d.BL[col] = 0
			}
			if d.Passive {
				d.BL[s.Col] = -applied
			} else {
				d.BL[s.Col] = applied
			}
			result, err := arraysim.EvaluateDevice(ctx, d, s.Weights, levels)
			if err != nil {
				return Result{}, err
			}
			effective := applied
			if result.CellVoltages != nil {
				effective = result.CellVoltages[s.Row][s.Col]
			}
			next := arraysim.ProgramDeviceLevel(d, current, effective, 200e-9, levels)
			next = arraysim.ClampProgramLevel(next, current, target, direction)
			// Native passive programming stresses the entire selected column; do not
			// pretend a 0T1R array provides selected-cell isolation.
			if d.Passive {
				for r := range s.Weights {
					if r == s.Row {
						continue
					}
					// Native applyColumnWrite uses applied column-drive voltage
					// for neighbors; only the selected ISPP cell uses coupled V.
					s.Weights[r][s.Col] = arraysim.ProgramDeviceLevel(d, s.Weights[r][s.Col], applied, 200e-9, levels)
				}
			}
			s.Weights[s.Row][s.Col] = next
			current = next
			out.Pulses = append(out.Pulses, Pulse{i + 1, current, code, voltage, applied, effective})
			status := calc.CheckResult(current, target, direction, i+1)
			if status == physics.ISPPSuccess {
				out.Verified = true
				break
			}
			if status == physics.ISPPMaxPulses {
				break
			}
			voltage = calc.CalculateNextVoltage(voltage, direction)
		}
		out.Verified = out.Verified || current == target
		if !out.Verified {
			out.Reason = "Circuit ISPP pulse limit reached; actual state retained, target not verified"
		}
		d = setup(c, s, "read")
	}
	result, err := arraysim.EvaluateDevice(ctx, d, s.Weights, levels)
	if err != nil {
		return Result{}, err
	}
	if yield != nil {
		if err = yield(ctx); err != nil {
			return Result{}, err
		}
	}
	if err = ctx.Err(); err != nil {
		return Result{}, err
	}
	out.State = s
	out.Device = result
	return out, nil
}
func Yield(ctx context.Context) error { time.Sleep(time.Millisecond); return ctx.Err() }
