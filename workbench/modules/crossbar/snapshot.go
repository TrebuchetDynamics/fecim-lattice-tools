package crossbar

import (
	"encoding/json"
	core "fecim-lattice-tools/shared/crossbar"
	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/presentation"
	"fmt"
	"strconv"
)

const evidence = "Simulation only; educational parameters, not measured hardware"

func cell(v float64, unit string) moduleapi.NumericCell {
	return moduleapi.NumericCell{Value: &v, Unit: unit, Evidence: evidence, Presentation: presentation.FormatMetric(experiment.Metric{Value: v, Unit: unit}, "")}
}
func field(id, label, input, unit string) moduleapi.Field {
	return moduleapi.Field{ID: id, Label: label, Input: input, Unit: unit, Presentation: presentation.Value{Display: input, Speakable: input, Availability: "available"}}
}
func (p *Port) snapshot() moduleapi.Snapshot {
	c := p.cfg
	s := moduleapi.Snapshot{Revision: p.revision, Descriptor: p.Descriptor(), Status: "Ready; configure and run explicitly", Evidence: []string{evidence, fmt.Sprintf("CPU cap %d by %d; no GPU/WebGPU; full sneak cap 32 by 32", p.limit, p.limit), "Accuracy view reports numerical error, not classification accuracy or hardware validation. No fabricated latency, throughput or GPU energy comparison."}, Tables: []moduleapi.TablePage{}, Plots: []moduleapi.Plot{}}
	ints := []struct {
		id, label string
		v         int
	}{{"rows", "Rows", c.Rows}, {"cols", "Columns", c.Cols}, {"adc_bits", "ADC bits", c.ADCBits}, {"dac_bits", "DAC bits", c.DACBits}, {"levels", "Programming levels", c.Levels}}
	for _, v := range ints {
		s.Fields = append(s.Fields, field(v.id, v.label, strconv.Itoa(v.v), "count"))
	}
	s.Fields = append(s.Fields, field("profile", "Model profile", c.Profile, ""), field("seed", "Seed", strconv.FormatInt(c.Seed, 10), "count"), field("architecture", "Isolation architecture", c.Architecture, ""), field("ir", "IR drop", strconv.FormatBool(c.IR), ""), field("sneak", "Sneak paths", strconv.FormatBool(c.Sneak), ""))
	for _, v := range []struct {
		id, label, unit string
		value           float64
	}{{"noise", "Device noise", "ratio", c.Noise}, {"temperature", "Temperature", "K", c.Temperature}, {"rp_row", "Row parasitic Rp/Rmin", "ratio", c.RpRow}, {"rp_col", "Column parasitic Rp/Rmin", "ratio", c.RpCol}} {
		f := field(v.id, v.label, strconv.FormatFloat(v.value, 'g', -1, 64), v.unit)
		f.Presentation = presentation.FormatMetric(experiment.Metric{Value: v.value, Unit: v.unit}, "")
		s.Fields = append(s.Fields, f)
	}
	input, _ := json.Marshal(p.input)
	s.Fields = append(s.Fields, field("input", "Input vector (JSON volts)", string(input), "V"), field("selected_row", "Selected row", strconv.Itoa(p.row), "count"), field("selected_col", "Selected column", strconv.Itoa(p.col), "count"), field("mode", "Analysis mode", p.mode, ""), field("layer_count", "Layer count", strconv.Itoa(p.layers), "count"), field("selected_layer", "Selected layer", strconv.Itoa(p.layer), "count"), field("job_id", "Job identifier", p.job.ID, ""), field("job_state", "Job state", p.job.State, ""))
	for _, id := range []string{"configure", "program", "select", "view", "reset", "run", "stack", "import"} {
		s.Actions = append(s.Actions, moduleapi.Action{ID: id, Label: id, Available: !p.active()})
	}
	for _, id := range []string{"load", "save"} {
		s.Actions = append(s.Actions, moduleapi.Action{ID: id, Label: id, Available: p.limit == 128 && !p.active(), Reason: "Requires authenticated local application and explicitly opened rooted project"})
	}
	if p.active() {
		s.Status = p.job.Message
	} else if p.job.State != "idle" {
		s.Status = p.job.Message
	}
	if p.result != nil {
		s.Fields = append(s.Fields, field("retained_profile", "Retained model profile", p.result.Config.Profile, ""))
		s.Evidence = append(s.Evidence, fmt.Sprintf("Retained result revision %d; current inputs must be explicitly rerun after changes", p.resultRevision), p.result.Normalization)
		if p.inputEpoch != p.resultEpoch {
			s.Evidence = append(s.Evidence, "STALE: retained outputs describe earlier inputs; rerun explicitly before interpreting them as current.")
		}
	}
	// Use only core quantization/mapping to present current cell state.
	cfg := &core.Config{Rows: c.Rows, Cols: c.Cols, ADCBits: c.ADCBits, DACBits: c.DACBits, NoiseLevel: c.Noise, QuantizationLevels: c.Levels}
	if c.Profile == "fecap" {
		cfg = core.DefaultFeCAPConfig(c.Cols, c.Rows)
		cfg.NoiseLevel = c.Noise
	}
	a, _ := core.NewArraySeeded(cfg, c.Seed)
	defer a.Destroy()
	matrix := clone(p.matrix)
	var physical [][]float64
	unit := "S"
	if c.Profile == "fecap" {
		a.ProgramCapacitanceMatrix(transpose(matrix))
		matrix, physical = transpose(a.GetCapacitanceMatrix()), transpose(a.GetEffectiveCapacitanceMatrix())
		unit = "F"
	} else {
		a.ProgramWeightMatrix(matrix)
		matrix, physical = a.GetConductanceMatrix(), physicalConductances(a)
	}
	total := (c.Rows*c.Cols + 31) / 32
	table := moduleapi.TablePage{ID: "matrix", Caption: fmt.Sprintf("Programmed matrix — page %d of %d; zero-based row and column", p.page, total), Page: p.page, Total: total, Columns: []string{"Row", "Column", "State (normalized)", "Physical cell (" + unit + ")"}, Rows: [][]moduleapi.NumericCell{}}
	for index := (p.page - 1) * 32; index < c.Rows*c.Cols && index < p.page*32; index++ {
		i, j := index/c.Cols, index%c.Cols
		table.Rows = append(table.Rows, []moduleapi.NumericCell{cell(float64(i), "count"), cell(float64(j), "count"), cell(matrix[i][j], "ratio"), cell(physical[i][j], unit)})
	}
	s.Tables = append(s.Tables, table, moduleapi.TablePage{ID: "selection", Caption: "Selected cell", Page: 1, Total: 1, Columns: table.Columns, Rows: [][]moduleapi.NumericCell{{cell(float64(p.row), "count"), cell(float64(p.col), "count"), cell(matrix[p.row][p.col], "ratio"), cell(physical[p.row][p.col], unit)}}})
	vector := moduleapi.TablePage{ID: "input", Caption: "Current input vector — columns, volts", Page: 1, Total: 1, Columns: []string{"Column", "Input voltage"}, Rows: [][]moduleapi.NumericCell{}}
	for i, v := range p.input {
		vector.Rows = append(vector.Rows, []moduleapi.NumericCell{cell(float64(i), "count"), cell(v, "V")})
	}
	s.Tables = append(s.Tables, vector)
	if p.mode == "stack" {
		s.Tables = append(s.Tables, p.stackTable())
	}
	if p.result == nil {
		s.Evidence = append(s.Evidence, "No computed result yet. IR, sneak, solver and output tables become available only after an explicit real model run.")
		return s
	}
	r := p.result
	output := moduleapi.TablePage{ID: "output", Caption: fmt.Sprintf("Retained ideal and actual output — result revision %d; %s", p.resultRevision, r.Unit), Page: 1, Total: 1, Columns: []string{"Output index (rows for MVM/FeCAP, columns for SOR)", "Ideal", "Actual"}, Rows: [][]moduleapi.NumericCell{}}
	for i, v := range r.Actual {
		output.Rows = append(output.Rows, []moduleapi.NumericCell{cell(float64(i), "count"), cell(r.Ideal[i], r.Unit), cell(v, r.Unit)})
	}
	s.Tables = append(s.Tables, output)
	s.Tables = append(s.Tables, moduleapi.TablePage{ID: "accuracy", Caption: "Numerical output error, not hardware/classification accuracy", Page: 1, Total: 1, Columns: []string{"RMSE"}, Rows: [][]moduleapi.NumericCell{{cell(r.RMSE, r.Unit)}}})
	resultTotal := (r.Config.Rows*r.Config.Cols + 31) / 32
	if r.IR != nil {
		ir := moduleapi.TablePage{ID: "ir", Caption: "Retained IR effective-voltage map (normalized to 1 volt drive)", Page: p.page, Total: resultTotal, Columns: []string{"Row", "Column", "Effective voltage"}, Rows: [][]moduleapi.NumericCell{}}
		for index := (p.page - 1) * 32; index < r.Config.Rows*r.Config.Cols && index < p.page*32; index++ {
			i, j := index/r.Config.Cols, index%r.Config.Cols
			ir.Rows = append(ir.Rows, []moduleapi.NumericCell{cell(float64(i), "count"), cell(float64(j), "count"), cell(r.IR.EffectiveVoltage[i][j], "V")})
		}
		s.Tables = append(s.Tables, ir, moduleapi.TablePage{ID: "ir_solver", Caption: fmt.Sprintf("IR relaxation converged: %t; residual is maximum voltage change", r.IR.Converged), Page: 1, Total: 1, Columns: []string{"Iterations", "Residual", "Max drop"}, Rows: [][]moduleapi.NumericCell{{cell(float64(r.IR.Iterations), "count"), cell(r.IR.Residual, "V"), cell(r.IR.MaxIRDrop, "ratio")}}})
	}
	if r.Sneak != nil {
		s.Tables = append(s.Tables, moduleapi.TablePage{ID: "sneak", Caption: "Retained core sneak analysis for center cell; currents in normalized conductance-times-unit-drive units, not amperes", Page: 1, Total: 1, Columns: []string{"Signal current", "Total sneak current", "Sneak ratio"}, Rows: [][]moduleapi.NumericCell{{cell(r.Sneak.TotalSignal, "ratio"), cell(r.Sneak.TotalSneak, "ratio"), cell(r.Sneak.MaxSneakRatio, "ratio")}}})
	}
	if r.Trace != nil {
		trace := moduleapi.TablePage{ID: "sneak-trace", Caption: "Retained three-cell MVM sneak trace; normalized path conductance/current, top 3 paths per output row", Page: 1, Total: 1, Columns: []string{"Output row", "Source row", "Source column", "Exit column", "Path conductance", "Path current"}, Rows: [][]moduleapi.NumericCell{}}
		for _, row := range r.Trace.Rows {
			for _, path := range row.TopPaths {
				trace.Rows = append(trace.Rows, []moduleapi.NumericCell{cell(float64(row.Row), "count"), cell(float64(path.SourceRow), "count"), cell(float64(path.SourceCol), "count"), cell(float64(path.ExitCol), "count"), cell(path.PathG, "ratio"), cell(path.PathCurrent, "ratio")})
			}
		}
		s.Tables = append(s.Tables, trace)
	}
	if r.Solver != nil {
		devices := moduleapi.TablePage{ID: "solver-cells", Caption: fmt.Sprintf("Retained SOR device voltages and currents — page %d of %d; normalized model", p.page, resultTotal), Page: p.page, Total: resultTotal, Columns: []string{"Row", "Column", "Device voltage", "Device current"}, Rows: [][]moduleapi.NumericCell{}}
		for index := (p.page - 1) * 32; index < r.Config.Rows*r.Config.Cols && index < p.page*32; index++ {
			i, j := index/r.Config.Cols, index%r.Config.Cols
			devices.Rows = append(devices.Rows, []moduleapi.NumericCell{cell(float64(i), "count"), cell(float64(j), "count"), cell(r.Solver.DeviceVoltages[i][j], "V"), cell(r.Solver.DeviceCurrents[i][j], "ratio")})
		}
		s.Tables = append(s.Tables, devices)
		s.Tables = append(s.Tables, moduleapi.TablePage{ID: "solver", Caption: fmt.Sprintf("SOR converged: %t; device voltages/current use normalized model", r.Solver.Converged), Page: 1, Total: 1, Columns: []string{"Iterations", "Residual", "Final omega"}, Rows: [][]moduleapi.NumericCell{{cell(float64(r.Solver.Iterations), "count"), cell(r.Solver.MaxError, "V"), cell(r.Solver.FinalOmega, "ratio")}}})
	}
	return s
}
