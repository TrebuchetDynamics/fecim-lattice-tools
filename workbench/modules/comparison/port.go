// Package comparison adapts the canonical Go scenario model to bounded,
// memory-only web sessions. No file or tool authority is acquired by New.
package comparison

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"

	model "fecim-lattice-tools/module5-comparison/pkg/comparison"
	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/presentation"
)

type Port struct {
	revision                     uint64
	workload, rate, mode, output string
	profile                      model.ScenarioProfile
	result                       model.ScenarioEstimate
	detail, status               string
	closed                       bool
}

func New() *Port {
	p := &Port{}
	p.reset()
	return p
}
func (p *Port) reset() {
	p.workload = "GPT-2"
	p.rate = "10000"
	p.mode = "Technical Review"
	p.profile = model.ScenarioBaseline
	p.output = "tco_reduction_pct"
	p.result, _ = model.CalculateScenario(p.workload, 10000, p.profile)
	p.detail = ""
	p.status = "Ready — simulation assumptions, not measured performance"
}
func (*Port) Descriptor() moduleapi.ModuleDescriptor {
	return moduleapi.ModuleDescriptor{ID: "comparison", Title: "Comparison", Available: true, SchemaVersion: moduleapi.SchemaVersion, ModelVersion: model.ScenarioModelVersion, Features: []string{"workload-rate", "scenarios", "evidence", "diff", "sensitivity", "architecture-demonstration", "local-export"}, Capabilities: []moduleapi.Capability{{Profile: "hosted", Available: true}, {Profile: "local", Available: true}}}
}
func (p *Port) Close() { p.closed = true }
func (p *Port) Apply(ctx context.Context, c moduleapi.Command) (moduleapi.Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return moduleapi.Snapshot{}, err
	}
	if p.closed || c.Revision != p.revision {
		return moduleapi.Snapshot{}, errors.New("closed session or stale comparison revision")
	}
	if len(c.Payload) != 0 {
		return moduleapi.Snapshot{}, errors.New("comparison accepts typed input strings, not payload bytes")
	}
	if c.Action != "export" && len(c.Resources) != 0 {
		return moduleapi.Snapshot{}, errors.New("unexpected resource references")
	}
	switch c.Action {
	case "calculate":
		workload, rate, mode, profile := p.workload, p.rate, p.mode, p.profile
		for k, v := range c.Input {
			if len(v) > 80 {
				return moduleapi.Snapshot{}, errors.New("comparison field exceeds 80 bytes")
			}
			switch k {
			case "workload":
				workload = v
			case "rate":
				rate = v
			case "mode":
				mode = v
			case "profile":
				profile = model.ScenarioProfile(v)
			default:
				return moduleapi.Snapshot{}, errors.New("unknown comparison field")
			}
		}
		if mode != "Technical Review" && mode != "Presentation" {
			return moduleapi.Snapshot{}, errors.New("unknown review mode")
		}
		n, err := strconv.ParseFloat(rate, 64)
		if err != nil {
			return moduleapi.Snapshot{}, errors.New("rate must be finite positive inferences per second (at most 1e9)")
		}
		result, err := model.CalculateScenario(workload, n, profile)
		if err != nil {
			return moduleapi.Snapshot{}, err
		}
		p.workload, p.rate, p.mode, p.profile, p.result = workload, rate, mode, profile, result
		p.detail = ""
		p.status = "Calculated scenario; architecture demonstration remains a separate protocol"
	case "reset":
		if len(c.Input) != 0 {
			return moduleapi.Snapshot{}, errors.New("reset has no inputs")
		}
		p.reset()
	case "presentation":
		if len(c.Input) != 0 {
			return moduleapi.Snapshot{}, errors.New("presentation has no inputs")
		}
		if p.mode == "Technical Review" {
			p.mode = "Presentation"
		} else {
			p.mode = "Technical Review"
		}
		p.status = p.mode + " — evidence and assumptions remain available"
	case "diff":
		if len(c.Input) != 0 {
			return moduleapi.Snapshot{}, errors.New("diff compares baseline to selected scenario")
		}
		p.detail = "diff"
		p.status = "Baseline to selected scenario; only available output deltas are comparable"
	case "sensitivity":
		output := p.output
		for k, v := range c.Input {
			if k != "output" {
				return moduleapi.Snapshot{}, errors.New("unknown sensitivity input")
			}
			output = v
		}
		if _, ok := p.result.Run.Outputs[output]; !ok {
			return moduleapi.Snapshot{}, errors.New("unknown sensitivity output")
		}
		p.output = output
		p.detail = "sensitivity"
		p.status = "Actual minus/plus 10% one-input perturbations, not a statistical interval"
	case "export":
		if len(c.Input) != 1 || (c.Input["format"] != "json" && c.Input["format"] != "csv") {
			return moduleapi.Snapshot{}, errors.New("choose json or csv")
		}
		if len(c.Resources) != 1 || c.Resources[0].Kind != "project" {
			return moduleapi.Snapshot{}, errors.New("export requires one opened, granted local project")
		}
		name, err := exportRooted(ctx, c.Resources[0], p.result, c.Input["format"])
		if err != nil {
			return moduleapi.Snapshot{}, err
		}
		p.status = "Exported " + name + " beneath the granted project; no science was recomputed"
	default:
		return moduleapi.Snapshot{}, errors.New("unsupported comparison action")
	}
	p.revision++
	return p.Snapshot(ctx)
}
func textCell(s string) moduleapi.NumericCell {
	return moduleapi.NumericCell{Presentation: presentation.Value{Display: s, Speakable: s, Availability: "available"}, Evidence: "simulation/assumption"}
}
func number(v float64, unit string) moduleapi.NumericCell {
	value := presentation.FormatMetric(experiment.Metric{Value: v, Unit: unit}, "")
	// Protocol-specific unit vocabulary stays in Go, never a browser conversion.
	if spoken, ok := map[string]string{"inf/s": "inferences per second", "inf/day": "inferences per day", "USD/month": "US dollars per month", "million USD/year": "million US dollars per year", "pJ/MAC": "picojoules per multiply accumulate operation", "ns/MAC": "nanoseconds per multiply accumulate operation", "kg/kWh": "kilograms per kilowatt hour", "percentage points": "percentage points"}[unit]; ok {
		base := presentation.FormatMetric(experiment.Metric{Value: v, Unit: "ratio"}, "")
		value.Display = strings.TrimSuffix(base.Display, " ratio") + " " + unit
		value.Speakable = strings.TrimSuffix(base.Speakable, " ratio") + " " + spoken
		value.Availability = base.Availability
		value.Reason = base.Reason
	}
	return moduleapi.NumericCell{Value: &v, Unit: unit, Presentation: value, Evidence: "simulation/assumption"}
}
func missing(unit, reason string) moduleapi.NumericCell {
	return moduleapi.NumericCell{Unit: unit, Presentation: presentation.FormatMetric(experiment.Metric{}, reason), Evidence: "unavailable — not calibrated"}
}
func page(id, caption string, columns []string, rows [][]moduleapi.NumericCell) moduleapi.TablePage {
	return moduleapi.TablePage{ID: id, Caption: caption + " — page 1 of 1; complete bounded set, canonical Go order", Page: 1, Total: len(rows), Columns: columns, Rows: rows}
}
func choices(id string, values []string) moduleapi.TablePage {
	rows := [][]moduleapi.NumericCell{}
	for _, v := range values {
		rows = append(rows, []moduleapi.NumericCell{textCell(v)})
	}
	return page(id, id, []string{"Option"}, rows)
}
func (p *Port) Snapshot(ctx context.Context) (moduleapi.Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return moduleapi.Snapshot{}, err
	}
	if p.closed {
		return moduleapi.Snapshot{}, errors.New("comparison session closed")
	}
	s := moduleapi.Snapshot{Revision: p.revision, Descriptor: p.Descriptor(), Status: p.status, Plots: []moduleapi.Plot{}, Evidence: []string{model.ScenarioNotice, "Scenario protocol: " + model.ScenarioProtocol + ". Rate is inferences/second; daily count is rate × 86400; kWh is joules / 3.6e6. Monthly electricity-only protocol assumes 730 hours/month and $0.10/kWh; TCO inputs are separate assumptions, not electricity bills.", "Sources: module5-comparison/pkg/comparison/scenario_profiles.go; architecture.go; shared/physics/confidence_ledger.go. CPU/DRAM and GPU/HBM public-datasheet descriptions are context, not validated benchmarks. FeCIM energy is an assumed simulation baseline. No scenario coefficient is calibrated here.", "Market context (legacy market.go inputs, not verified forecasts): NAND Flash 72/85/98, DRAM 130/165/220, AI Semiconductor 140/220/403 billion USD for 2024/2026/2030. Segments can overlap; their sum is not a proven addressable market.", "Fabrication context (legacy fabrication_reality.go model inputs): design 6–9 months; first tape-out 3–4; processing 3–4; testing 3–6; revision +6–12. Illustrative total 18–36 months, $5–50M+; not a quotation or qualification. Analog multilevel yield and novel-memory integration remain risks.", model.PlainTextEvidence("Scenario source evidence", p.result.Run)}}
	for _, f := range [][3]string{{"workload", "Workload", p.workload}, {"profile", "Scenario", string(p.profile)}, {"rate", "Inference rate", p.rate}, {"mode", "Review mode", p.mode}, {"output", "Sensitivity output", p.output}} {
		cell := textCell(f[2])
		unit := ""
		if f[0] == "rate" {
			cell = number(p.result.InferencesPerSecond, "inf/s")
			unit = "inf/s"
		}
		s.Fields = append(s.Fields, moduleapi.Field{ID: f[0], Label: f[1], Input: f[2], Unit: unit, Presentation: cell.Presentation})
	}
	for _, a := range [][2]string{{"calculate", "Calculate"}, {"reset", "Reset"}, {"presentation", "Toggle presentation"}, {"diff", "Compare with baseline"}, {"sensitivity", "Rank sensitivity"}} {
		s.Actions = append(s.Actions, moduleapi.Action{ID: a[0], Label: a[1], Available: true})
	}
	export := moduleapi.Action{ID: "export", Label: "Export JSON/CSV", Available: localExportAvailable(ctx)}
	if !export.Available {
		export.Reason = "Requires authenticated local application and an opened granted project"
	}
	s.Actions = append(s.Actions, export)
	s.Tables = []moduleapi.TablePage{choices("workloads", model.ScenarioWorkloads()), choices("profiles", []string{"conservative", "baseline", "optimistic"}), choices("modes", []string{"Technical Review", "Presentation"}), choices("outputs", []string{"energy_reduction_pct", "latency_reduction_pct", "tco_reduction_pct", "co2_reduction_pct"})}
	rows := [][]moduleapi.NumericCell{}
	for i, name := range []string{"CPU + DRAM", "GPU + HBM", "FeCIM"} {
		u := p.result.Energy[i]
		rows = append(rows, []moduleapi.NumericCell{textCell(name), number(u.JPerInference, "J"), number(u.DailyInferences, "inf/day"), number(u.DailyKWh, "kWh"), number(u.PowerW, "W"), number(u.MonthlyUSD, "USD/month")})
	}
	s.Tables = append(s.Tables, page("energy", "Workload-rate scenario energy and electricity-only cost", []string{"Architecture", "Energy per inference", "Daily inferences", "Daily energy", "Power", "Monthly electricity"}, rows))
	rows = [][]moduleapi.NumericCell{{textCell("Annual GPU minus FeCIM electricity savings"), number(p.result.AnnualSavingsMillionUSD, "million USD/year")}}
	for _, v := range [][2]string{{"energy_reduction_pct", "Energy reduction"}, {"latency_reduction_pct", "Latency reduction"}, {"tco_reduction_pct", "TCO reduction"}, {"co2_reduction_pct", "CO2 reduction"}} {
		o := p.result.Run.Outputs[v[0]]
		cell := number(o.Value, o.Units)
		if o.UnavailableReason != "" {
			cell = missing(o.Units, o.UnavailableReason)
		}
		rows = append(rows, []moduleapi.NumericCell{textCell(v[1]), cell})
	}
	s.Tables = append(s.Tables, page("summary", "Scenario summary — assumed ratios, not validated savings or investment ROI", []string{"Metric", "Value"}, rows))
	rows = nil
	keys := make([]string, 0, len(p.result.Run.Inputs))
	for k := range p.result.Run.Inputs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := p.result.Run.Inputs[k]
		rows = append(rows, []moduleapi.NumericCell{textCell(k), number(v.Value, v.Units), textCell(string(v.Tag.Provenance) + " — qualitative source confidence, not a statistical interval")})
	}
	if p.mode == "Technical Review" {
		s.Tables = append(s.Tables, page("sources", "Scenario assumptions and source evidence", []string{"Input", "Value", "Evidence"}, rows))
	}
	// Architecture specifications are a distinct demonstration, never substituted
	// for the scenario's per-MAC assumptions or passed off as new screen results.
	rows = nil
	for _, a := range model.Architectures() {
		r := a.RunInference(int(p.result.MACsPerInference), 1)
		rows = append(rows, []moduleapi.NumericCell{textCell(a.Name), number(a.PeakTOPS, "TOPS"), number(a.TDP, "W"), number(a.ChipArea, "mm2"), number(r.Latency, "ms"), number(r.Energy*1e-3, "J")})
	}
	s.Tables = append(s.Tables, page("architectures", "Separate architecture-demonstration protocol — TOPS/TDP/memory assumptions, batch 1", []string{"Architecture", "Peak throughput", "TDP", "Chip area", "Latency", "Energy per inference"}, rows))
	if p.detail == "diff" {
		base := model.BuildScenarioRun(model.ScenarioBaseline, model.PresetScenarioConfig(model.ScenarioBaseline))
		diff := model.DiffScenarios(base, p.result.Run)
		s.Evidence = append(s.Evidence, diff.ChangedAssumptions...)
		s.Evidence = append(s.Evidence, diff.Attribution...)
		rows = nil
		keys = keys[:0]
		for k := range diff.OutputDeltas {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			rows = append(rows, []moduleapi.NumericCell{textCell(k), number(diff.OutputDeltas[k], "percentage points")})
		}
		rows = append(rows, []moduleapi.NumericCell{textCell("CO2 reduction"), missing("%", model.CO2UnavailableReason)})
		s.Tables = append(s.Tables, page("diff", "Baseline to selected scenario differences", []string{"Output", "Selected minus baseline"}, rows))
	}
	if p.detail == "sensitivity" {
		rows = nil
		for _, r := range model.SensitivityRanking(p.result.Run, p.output) {
			rows = append(rows, []moduleapi.NumericCell{textCell(r.InputName), number(r.ImpactPct, "%"), number(r.DownwardImpact, "percentage points"), number(r.UpwardImpact, "percentage points")})
		}
		caption := "Sensitivity — " + p.output + "; descending impact then input name; ±10% input perturbation"
		if len(rows) == 0 {
			caption += "; Not available: missing/zero output denominator or " + model.CO2UnavailableReason
		}
		s.Tables = append(s.Tables, page("sensitivity", caption, []string{"Input", "Maximum relative output change", "Minus 10% delta", "Plus 10% delta"}, rows))
	}
	// No clocks, presentation modes, request IDs or browser reserialization enter
	// this canonical scientific identity. Native/Node/Firefox compare these bytes.
	return s, nil
}
func (p *Port) Canonical(format string) ([]byte, error) {
	return model.MarshalScenario(p.result, format)
}

var _ moduleapi.Port = (*Port)(nil)
