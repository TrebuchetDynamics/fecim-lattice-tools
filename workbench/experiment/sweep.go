package experiment

import (
	"fmt"
	"math"

	"fecim-lattice-tools/workbench/project"
)

func Expand(bundle project.Bundle) ([]DesignPoint, error) {
	parameters := bundle.Sweep.Parameters
	values := make([][]float64, len(parameters))
	seen := make(map[string]struct{}, len(parameters))
	points := 1
	for i, parameter := range parameters {
		if _, ok := seen[parameter.Path]; ok {
			return nil, fmt.Errorf("duplicate sweep path %q", parameter.Path)
		}
		seen[parameter.Path] = struct{}{}
		resolved, err := parameterValues(parameter)
		if err != nil {
			return nil, fmt.Errorf("parameter %s: %w", parameter.Path, err)
		}
		if _, err := apply(bundle.Design, parameter.Path, resolved[0]); err != nil {
			return nil, err
		}
		if points > bundle.Sweep.MaxPoints/len(resolved) {
			return nil, fmt.Errorf("sweep exceeds max_points %d", bundle.Sweep.MaxPoints)
		}
		points *= len(resolved)
		values[i] = resolved
	}

	out := make([]DesignPoint, 0, points)
	var walk func(int, project.Design) error
	walk = func(depth int, design project.Design) error {
		if depth == len(parameters) {
			index := len(out)
			out = append(out, DesignPoint{Index: index, Design: design, Seed: bundle.Sweep.Seed + int64(index)})
			return nil
		}
		for _, value := range values[depth] {
			next, err := apply(design, parameters[depth].Path, value)
			if err != nil {
				return err
			}
			if err := walk(depth+1, next); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(0, bundle.Design); err != nil {
		return nil, err
	}
	return out, nil
}

func parameterValues(parameter project.Parameter) ([]float64, error) {
	hasValues := len(parameter.Values) > 0
	hasRange := parameter.Range != nil
	if hasValues == hasRange {
		return nil, fmt.Errorf("provide exactly one of values or range")
	}
	if hasValues {
		out := append([]float64(nil), parameter.Values...)
		for _, value := range out {
			if !finite(value) {
				return nil, fmt.Errorf("values must be finite")
			}
		}
		return out, nil
	}
	r := parameter.Range
	if r.Count < 1 || !finite(r.Start) || !finite(r.Stop) {
		return nil, fmt.Errorf("range requires finite endpoints and count >= 1")
	}
	if r.Count == 1 {
		return []float64{r.Start}, nil
	}
	out := make([]float64, r.Count)
	step := (r.Stop - r.Start) / float64(r.Count-1)
	for i := range out {
		out[i] = r.Start + float64(i)*step
	}
	out[len(out)-1] = r.Stop
	return out, nil
}

func apply(design project.Design, path string, value float64) (project.Design, error) {
	if !finite(value) {
		return project.Design{}, fmt.Errorf("sweep value for %s must be finite", path)
	}
	integer := func() (int, error) {
		if math.Trunc(value) != value {
			return 0, fmt.Errorf("sweep value for %s must be an integer", path)
		}
		return int(value), nil
	}
	switch path {
	case "device.conductance_levels":
		v, err := integer()
		if err != nil {
			return project.Design{}, err
		}
		design.Device.ConductanceLevels = v
	case "device.g_min_s":
		design.Device.GMinS = value
	case "device.g_max_s":
		design.Device.GMaxS = value
	case "array.rows":
		v, err := integer()
		if err != nil {
			return project.Design{}, err
		}
		design.Array.Rows = v
	case "array.cols":
		v, err := integer()
		if err != nil {
			return project.Design{}, err
		}
		design.Array.Cols = v
	case "array.read_voltage_v":
		design.Array.ReadVoltageV = value
	case "circuit.adc_bits":
		v, err := integer()
		if err != nil {
			return project.Design{}, err
		}
		design.Circuit.ADCBits = v
	case "circuit.dac_bits":
		v, err := integer()
		if err != nil {
			return project.Design{}, err
		}
		design.Circuit.DACBits = v
	case "circuit.tia_gain_ohm":
		design.Circuit.TIAGainOhm = value
	default:
		return project.Design{}, fmt.Errorf("unsupported sweep path %q", path)
	}
	return design, nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
