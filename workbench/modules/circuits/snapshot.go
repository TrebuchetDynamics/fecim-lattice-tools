package circuits

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	"fecim-lattice-tools/module4-circuits/pkg/gui/writeseq"
	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/presentation"
)

const evidence = "Educational simulation, not measured hardware; native material/device/arraysim equations"

func cell(v float64, unit string) moduleapi.NumericCell {
	out := moduleapi.NumericCell{Unit: unit, Evidence: evidence, Presentation: presentation.FormatMetric(experiment.Metric{Value: v, Unit: unit}, "")}
	if finite(v) {
		out.Value = &v
	}
	return out
}
func field(id, label, value, unit string) moduleapi.Field {
	return moduleapi.Field{ID: id, Label: label, Input: value, Unit: unit, Presentation: presentation.Value{Display: value, Speakable: value, Availability: "available"}}
}
func table(id, caption string, columns ...string) moduleapi.TablePage {
	return moduleapi.TablePage{ID: id, Caption: caption, Page: 1, Total: 1, Columns: columns, Rows: [][]moduleapi.NumericCell{}}
}
func (p *Port) snapshot() moduleapi.Snapshot {
	c, s := p.cfg, p.state
	d := c.DeviceSettings()
	levels := d.Material.GetNumLevels()
	snap := moduleapi.Snapshot{Revision: p.revision, Descriptor: p.Descriptor(), Status: p.status, Evidence: []string{evidence, "READ uses material-native levels. Coupled row sensing uses native magnitude polarity; ideal row currents and all cell currents retain sign.", "0T1R WRITE programs the entire selected column, not an isolated cell. Native neighbor programming uses applied column-drive voltage; selected-cell programming uses coupled voltage. 2T1R masks gate cell connections.", "Circuit level ISPP uses the native deterministic coupled-voltage LK pulse update and shared circuit calculator, not Hysteresis waveform ISPP. Manual DAC voltages, when present, override byte-code vector inputs for COMPUTE.", "Supply is a peripheral reference setting, not a new array equation. GPU and Hysteresis waveform/time-resolved programming are unavailable.", fmt.Sprintf("CPU profile cap %d by %d; solver cap 4000 iterations; ISPP cap 40 pulses; logs retain latest 64 actions", p.limit, p.limit)}, Tables: []moduleapi.TablePage{}, Plots: []moduleapi.Plot{}}
	for _, v := range []struct {
		id, label string
		n         int
	}{{"rows", "Rows", c.Rows}, {"cols", "Columns", c.Cols}, {"node", "Technology node (nm)", c.Node}, {"dac_bits", "DAC bits", c.DACBits}, {"adc_bits", "ADC bits", c.ADCBits}, {"max_iterations", "Solver iteration cap", c.MaxIterations}, {"max_pulses", "ISPP pulse cap", c.MaxPulses}, {"levels", "Material levels", levels}, {"selected_row", "Selected row", s.Row}, {"selected_col", "Selected column", s.Col}} {
		snap.Fields = append(snap.Fields, field(v.id, v.label, strconv.Itoa(v.n), "count"))
	}
	for _, v := range []struct{ id, label, value string }{{"material", "Material", c.Material}, {"architecture", "Architecture", c.Architecture}, {"coupling", "Coupling", c.Coupling}, {"job_id", "Job identifier", p.job.ID}, {"job_state", "Job state", p.job.State}, {"seed", "Random seed", strconv.FormatInt(c.Seed, 10)}} {
		snap.Fields = append(snap.Fields, field(v.id, v.label, v.value, ""))
	}
	for _, v := range []struct {
		id, label, unit string
		n               float64
	}{{"gain", "TIA gain", "Ω", c.Gain}, {"supply", "Supply reference", "V", c.Supply}, {"sense_min", "Sense reference minimum", "V", c.SenseMin}, {"sense_max", "Sense reference maximum", "V", c.SenseMax}} {
		f := field(v.id, v.label, strconv.FormatFloat(v.n, 'g', -1, 64), v.unit)
		f.Presentation = cell(v.n, v.unit).Presentation
		snap.Fields = append(snap.Fields, f)
	}
	for _, v := range []struct {
		id, label string
		value     any
	}{{"input", "Signed input codes (JSON)", s.Input}, {"voltages", "Manual DAC voltages (JSON or null)", s.Voltages}, {"active_rows", "Active rows (JSON)", s.ActiveRows}, {"read_mask", "2T1R read mask (JSON or null)", s.ReadMask}, {"write_mask", "2T1R write mask (JSON or null)", s.WriteMask}} {
		data, _ := json.Marshal(v.value)
		snap.Fields = append(snap.Fields, field(v.id, v.label, string(data), ""))
	}
	options := map[string][]string{"material": {"FeCIM", "HZO", "BTO", "PZT"}, "node": {"130", "65", "28", "14"}, "architecture": {"0T1R", "1T1R", "2T1R"}, "coupling": {"Ideal", "TierA", "TierB"}}
	for _, n := range Sizes {
		if n <= p.limit {
			options["rows"] = append(options["rows"], strconv.Itoa(n))
			options["cols"] = append(options["cols"], strconv.Itoa(n))
		}
	}
	for n := 4; n <= 8; n++ {
		options["dac_bits"] = append(options["dac_bits"], strconv.Itoa(n))
	}
	for n := 5; n <= 8; n++ {
		options["adc_bits"] = append(options["adc_bits"], strconv.Itoa(n))
	}
	presetName := "Custom"
	for _, v := range arraysim.NativeSensePresets() {
		options["sense_preset"] = append(options["sense_preset"], v.Name)
		if c.Gain == v.Rf*1e3 && c.SenseMin == v.Vmin && c.SenseMax == v.Vmax {
			presetName = v.Name
		}
	}
	options["sense_preset"] = append(options["sense_preset"], "Custom")
	snap.Fields = append(snap.Fields, field("sense_preset", "Sense preset", presetName, ""))
	catalogue, _ := json.Marshal(options)
	snap.Fields = append(snap.Fields, field("choices", "Go-owned control choices", string(catalogue), ""), field("write_target", "Write target level", strconv.Itoa(levels/2), "count"))
	for _, id := range []string{"configure", "select", "write", "read", "compute", "undo", "reset", "randomize", "view", "timing"} {
		snap.Actions = append(snap.Actions, moduleapi.Action{ID: id, Label: id, Available: !p.active() && (id != "undo" || p.undo != nil)})
	}
	snap.Actions = append(snap.Actions, moduleapi.Action{ID: "save", Label: "Export circuit state and evidence", Available: p.limit == 128 && !p.active(), Reason: "Requires authenticated local profile and explicitly opened rooted project"})
	geometry := table("reference", "Material and geometry reference — canonical SI values", "Read maximum", "Write absolute maximum", "Film thickness", "Cell area", "Pitch X", "Pitch Y", "Selector resistance", "Selector leakage", "Supply reference")
	_, writeMax := arraysim.DeviceVoltageLimits(d.Material)
	geometry.Rows = append(geometry.Rows, []moduleapi.NumericCell{cell(d.ReadMax, "V"), cell(writeMax, "V"), cell(d.Geometry.Film.Thickness, "m"), cell(d.Geometry.Film.Area, "m²"), cell(d.Geometry.PitchX, "m"), cell(d.Geometry.PitchY, "m"), cell(d.SelectorRon, "Ω"), cell(d.SelectorLeakage, "S"), cell(c.Supply, "V")})
	snap.Tables = append(snap.Tables, geometry)
	cells := table("cells", fmt.Sprintf("Persistent cells — page %d of %d; zero-based coordinates", p.page, (c.Rows*c.Cols+31)/32), "Row", "Column", "Programmed level", "Conductance")
	cells.Page = p.page
	cells.Total = (c.Rows*c.Cols + 31) / 32
	for index := (p.page - 1) * 32; index < c.Rows*c.Cols && index < p.page*32; index++ {
		r, col := index/c.Cols, index%c.Cols
		cells.Rows = append(cells.Rows, []moduleapi.NumericCell{cell(float64(r), "count"), cell(float64(col), "count"), cell(float64(s.Weights[r][col]), "count"), cell(arraysim.DeviceConductance(d, s.Weights[r][col], levels), "S")})
	}
	selected := table("selection", "Selected persistent cell", "Row", "Column", "Programmed level", "Conductance")
	selected.Rows = append(selected.Rows, []moduleapi.NumericCell{cell(float64(s.Row), "count"), cell(float64(s.Col), "count"), cell(float64(s.Weights[s.Row][s.Col]), "count"), cell(arraysim.DeviceConductance(d, s.Weights[s.Row][s.Col], levels), "S")})
	inputs := table("input", "Current vector and DAC command voltages — signed byte codes", "Column", "Input code", "Voltage")
	for col, v := range s.Input {
		voltage := float64(v) / 255 * d.ReadMax
		if s.Voltages != nil {
			voltage = s.Voltages[col]
		}
		inputs.Rows = append(inputs.Rows, []moduleapi.NumericCell{cell(float64(col), "count"), cell(float64(v), "count"), cell(voltage, "V")})
	}
	snap.Tables = append(snap.Tables, cells, selected, inputs)
	timing := table("timing", "Native WRITE phase playback reference — behavioral display timings, not measured latency", "Phase", "Duration", "Playback selected")
	for i := 1; i <= 5; i++ {
		chosen := 0.
		if p.timing == i-1 {
			chosen = 1
		}
		timing.Rows = append(timing.Rows, []moduleapi.NumericCell{cell(float64(i), "count"), cell(float64(writeseq.PhaseDuration(i))*1e-9, "s"), cell(chosen, "count")})
	}
	snap.Evidence = append(snap.Evidence, "WRITE timing phases: RESET, HOLD, WRITE, HOLD, VERIFY. Playback steps do not alter physical simulation time or cell state.")
	peripheralTiming := table("peripheral_timing", "Native peripheral timing estimates — not calibrated end-to-end latency", "DAC settling", "TIA settling", "ADC conversion")
	peripheralTiming.Rows = append(peripheralTiming.Rows, []moduleapi.NumericCell{cell(d.DAC.SettleTime*1e-9, "s"), cell(d.TIA.SettlingTime(), "s"), cell(d.ADC.ConversionTime*1e-9, "s")})
	snap.Tables = append(snap.Tables, timing, peripheralTiming)
	if p.result != nil {
		r := p.result
		snap.Evidence = append(snap.Evidence, fmt.Sprintf("Retained operation %s; material %s, coupling %s, architecture %s, selected cell [%d,%d]; result is immutable until explicitly rerun", r.Operation, r.Config.Material, r.Config.Coupling, r.Config.Architecture, r.State.Row, r.State.Col))
		if p.inputEpoch != p.resultEpoch {
			snap.Evidence = append(snap.Evidence, "STALE: retained results belong to earlier inputs; rerun before treating them as current")
		}
		rows := table("sense", "Retained row sense results — native magnitude polarity for coupled solves", "Row", "Current", "TIA voltage", "ADC code", "Saturated", "Analytical SNR")
		points := [][2]float64{}
		retainedSettings := r.Config.DeviceSettings()
		for row, current := range r.Device.RowCurrents {
			saturated := 0.
			if r.Device.Saturated[row] {
				saturated = 1
			}
			snr := arraysim.DeviceSenseSNR(retainedSettings, current*1e-6)
			rows.Rows = append(rows.Rows, []moduleapi.NumericCell{cell(float64(row), "count"), cell(current*1e-6, "A"), cell(r.Device.RowVoltages[row], "V"), cell(float64(r.Device.RowLevels[row]), "count"), cell(saturated, "count"), cell(snr, "dB")})
			points = append(points, [2]float64{float64(row), current * 1e-6})
		}
		snap.Plots = append(snap.Plots, moduleapi.Plot{ID: "sense", Title: "Retained row currents — all rows; inspect complete sense table", XUnit: "count", YUnit: "A", Points: points})
		outputs := table("cell_results", fmt.Sprintf("Retained cell voltage and current — page %d; retained dimensions %d by %d", p.resultPage, r.Config.Rows, r.Config.Cols), "Row", "Column", "Effective voltage", "Cell current")
		outputs.Page = p.resultPage
		outputs.Total = (r.Config.Rows*r.Config.Cols + 31) / 32
		for index := (p.resultPage - 1) * 32; index < r.Config.Rows*r.Config.Cols && index < p.resultPage*32; index++ {
			row, col := index/r.Config.Cols, index%r.Config.Cols
			outputs.Rows = append(outputs.Rows, []moduleapi.NumericCell{cell(float64(row), "count"), cell(float64(col), "count"), cell(r.Device.CellVoltages[row][col], "V"), cell(r.Device.CellCurrents[row][col], "A")})
		}
		pulses := table("pulses", "Circuit WRITE pulse/verify trace — no fabricated convergence", "Iteration", "Level", "DAC code", "Requested voltage", "Applied voltage", "Effective voltage")
		for _, v := range r.Pulses {
			pulses.Rows = append(pulses.Rows, []moduleapi.NumericCell{cell(float64(v.Iteration), "count"), cell(float64(v.Level), "count"), cell(float64(v.Code), "count"), cell(v.Requested, "V"), cell(v.Applied, "V"), cell(v.Effective, "V")})
		}
		columns := table("columns", "Retained column current and spatial voltage stress — not an endurance or damage prediction", "Column", "Total signed cell current", "Maximum absolute cell voltage", "Voltage/coercive ratio")
		for col := 0; col < r.Config.Cols; col++ {
			current, maxVoltage := 0., 0.
			for row := 0; row < r.Config.Rows; row++ {
				current += r.Device.CellCurrents[row][col]
				maxVoltage = math.Max(maxVoltage, math.Abs(r.Device.CellVoltages[row][col]))
			}
			columns.Rows = append(columns.Rows, []moduleapi.NumericCell{cell(float64(col), "count"), cell(current, "A"), cell(maxVoltage, "V"), cell(maxVoltage/retainedSettings.Material.CoerciveVoltage(), "ratio")})
		}
		snap.Tables = append(snap.Tables, rows, outputs, pulses, columns)
	}
	snap.Evidence = append(snap.Evidence, p.logs...)
	return snap
}
