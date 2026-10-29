package comparison

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"fecim-lattice-tools/shared/physics"
)

type TaggedValue struct {
	Name              string
	Value             float64
	Units             string
	Tag               physics.ConfidenceTag
	Uncertainty       float64 // Deprecated heuristic magnitude, never a statistical interval.
	UnavailableReason string
}

// MarshalJSON makes unavailable scientific values null, never a fabricated zero.
// The deprecated heuristic field is deliberately absent from version 2 artifacts.
func (v TaggedValue) MarshalJSON() ([]byte, error) {
	var value *float64
	if v.UnavailableReason == "" && !math.IsNaN(v.Value) && !math.IsInf(v.Value, 0) {
		value = &v.Value
	}
	reason := v.UnavailableReason
	if value == nil && reason == "" {
		reason = "Required value is nonfinite"
	}
	return json.Marshal(struct {
		Name              string
		Value             *float64
		Units             string
		Tag               physics.ConfidenceTag
		UnavailableReason string `json:",omitempty"`
	}{v.Name, value, v.Units, v.Tag, reason})
}

type ScenarioRun struct {
	Profile ScenarioProfile
	Inputs  map[string]TaggedValue
	Outputs map[string]TaggedValue
}

type SensitivityImpact struct {
	InputName      string
	OutputName     string
	ImpactPct      float64
	DownwardImpact float64
	UpwardImpact   float64
}

type ScenarioDiff struct {
	FromProfile        string
	ToProfile          string
	ChangedAssumptions []string
	OutputDeltas       map[string]float64
	Attribution        []string
}

func BuildScenarioRun(profile ScenarioProfile, cfg ScenarioConfig) ScenarioRun {
	ledger := physics.NewConfidenceLedger()
	inputs := map[string]TaggedValue{}
	put := func(k string, v float64, units string, fallback physics.Provenance, conf float64) {
		tag, ok := ledger.Lookup(k)
		if !ok {
			tag = physics.ConfidenceTag{Provenance: fallback, Confidence: conf}
		}
		inputs[k] = TaggedValue{Name: k, Value: v, Units: units, Tag: tag}
	}

	put("cpu_energy_pj_per_mac", cfg.CPUEnergyPJPerMAC, "pJ/MAC", physics.ProvenanceEstimated, 0.7)
	put("gpu_energy_pj_per_mac", cfg.GPUEnergyPJPerMAC, "pJ/MAC", physics.ProvenanceEstimated, 0.75)
	put("fecim_energy_pj_per_mac", cfg.FeCIMEnergyPJPerMAC, "pJ/MAC", physics.ProvenancePlaceholder, 0.45)
	put("cpu_latency_ns_per_mac", cfg.CPULatencyNSPerMAC, "ns/MAC", physics.ProvenanceEstimated, 0.7)
	put("gpu_latency_ns_per_mac", cfg.GPULatencyNSPerMAC, "ns/MAC", physics.ProvenanceEstimated, 0.75)
	put("fecim_latency_ns_per_mac", cfg.FeCIMLatencyNSPerMAC, "ns/MAC", physics.ProvenancePlaceholder, 0.45)
	put("cpu_tco_usd_per_month", cfg.CPUTCOUSDPerMonth, "USD/month", physics.ProvenanceEstimated, 0.6)
	put("gpu_tco_usd_per_month", cfg.GPUTCOUSDPerMonth, "USD/month", physics.ProvenanceEstimated, 0.7)
	put("fecim_tco_usd_per_month", cfg.FeCIMTCOUSDPerMonth, "USD/month", physics.ProvenancePlaceholder, 0.45)
	put("cpu_co2_kg_per_kwh", cfg.CPUCO2KgPerKWh, "kg/kWh", physics.ProvenanceEstimated, 0.8)
	put("gpu_co2_kg_per_kwh", cfg.GPUCO2KgPerKWh, "kg/kWh", physics.ProvenanceEstimated, 0.8)
	put("fecim_co2_kg_per_kwh", cfg.FeCIMCO2KgPerKWh, "kg/kWh", physics.ProvenanceEstimated, 0.8)
	// The source ledger's historical calibrated labels do not validate these presets.
	for _, key := range []string{"cpu_co2_kg_per_kwh", "gpu_co2_kg_per_kwh", "fecim_co2_kg_per_kwh"} {
		v := inputs[key]
		v.Tag.Provenance = physics.ProvenanceEstimated
		inputs[key] = v
	}

	outputs := map[string]TaggedValue{}
	out := func(k string, value float64, units string, sourceInput string, fallback physics.Provenance, conf float64) {
		tag, ok := ledger.Lookup(sourceInput)
		if !ok {
			tag = physics.ConfidenceTag{Provenance: fallback, Confidence: conf}
		}
		v := TaggedValue{Name: k, Value: value, Units: units, Tag: tag}
		if err := cfg.Validate(); err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
			v.Value = 0
			v.UnavailableReason = "Required scenario assumptions or denominators are missing, nonfinite or nonpositive"
		}
		outputs[k] = v
	}

	out("energy_reduction_pct", 100*(1-(cfg.FeCIMEnergyPJPerMAC/cfg.GPUEnergyPJPerMAC)), "%", "fecim_energy_pj_per_mac", physics.ProvenancePlaceholder, 0.45)
	out("latency_reduction_pct", 100*(1-(cfg.FeCIMLatencyNSPerMAC/cfg.GPULatencyNSPerMAC)), "%", "fecim_latency_ns_per_mac", physics.ProvenancePlaceholder, 0.45)
	out("tco_reduction_pct", 100*(1-(cfg.FeCIMTCOUSDPerMonth/cfg.GPUTCOUSDPerMonth)), "%", "fecim_tco_usd_per_month", physics.ProvenancePlaceholder, 0.45)
	outputs["co2_reduction_pct"] = TaggedValue{Name: "co2_reduction_pct", Units: "%", UnavailableReason: CO2UnavailableReason, Tag: physics.ConfidenceTag{Provenance: physics.ProvenanceEstimated}}

	return ScenarioRun{Profile: profile, Inputs: inputs, Outputs: outputs}
}

func SensitivityRanking(run ScenarioRun, outputName string) []SensitivityImpact {
	base, ok := run.Outputs[outputName]
	if !ok || base.UnavailableReason != "" || base.Value == 0 || math.IsNaN(base.Value) || math.IsInf(base.Value, 0) {
		return nil
	}
	required := map[string][]string{
		"energy_reduction_pct":  {"gpu_energy_pj_per_mac", "fecim_energy_pj_per_mac"},
		"latency_reduction_pct": {"gpu_latency_ns_per_mac", "fecim_latency_ns_per_mac"},
		"tco_reduction_pct":     {"gpu_tco_usd_per_month", "fecim_tco_usd_per_month"},
	}[outputName]
	if len(required) == 0 {
		return nil
	}
	for _, key := range required {
		v, ok := run.Inputs[key]
		if !ok || v.UnavailableReason != "" || !finitePositive(v.Value) {
			return nil
		}
	}
	impacts := make([]SensitivityImpact, 0, len(run.Inputs))
	for key := range run.Inputs {
		down := recomputeWithPerturbation(run, key, -0.10, outputName)
		up := recomputeWithPerturbation(run, key, 0.10, outputName)
		impact := math.Max(math.Abs((down-base.Value)/base.Value), math.Abs((up-base.Value)/base.Value)) * 100
		impacts = append(impacts, SensitivityImpact{
			InputName: key, OutputName: outputName, ImpactPct: impact,
			DownwardImpact: down - base.Value, UpwardImpact: up - base.Value,
		})
	}
	sort.Slice(impacts, func(i, j int) bool {
		if impacts[i].ImpactPct == impacts[j].ImpactPct {
			return impacts[i].InputName < impacts[j].InputName
		}
		return impacts[i].ImpactPct > impacts[j].ImpactPct
	})
	return impacts
}

func recomputeWithPerturbation(run ScenarioRun, inputName string, perturb float64, outputName string) float64 {
	vals := map[string]float64{}
	for k, v := range run.Inputs {
		vals[k] = v.Value
	}
	vals[inputName] = vals[inputName] * (1 + perturb)
	switch outputName {
	case "energy_reduction_pct":
		return 100 * (1 - (vals["fecim_energy_pj_per_mac"] / vals["gpu_energy_pj_per_mac"]))
	case "latency_reduction_pct":
		return 100 * (1 - (vals["fecim_latency_ns_per_mac"] / vals["gpu_latency_ns_per_mac"]))
	case "tco_reduction_pct":
		return 100 * (1 - (vals["fecim_tco_usd_per_month"] / vals["gpu_tco_usd_per_month"]))
	case "co2_reduction_pct":
		return 0 // Unavailable; currency is not physical energy.
	default:
		return 0
	}
}

func DiffScenarios(a, b ScenarioRun) ScenarioDiff {
	d := ScenarioDiff{FromProfile: string(a.Profile), ToProfile: string(b.Profile), OutputDeltas: map[string]float64{}}
	for k, av := range a.Inputs {
		if bv, ok := b.Inputs[k]; ok && math.Abs(av.Value-bv.Value) > 1e-9 {
			d.ChangedAssumptions = append(d.ChangedAssumptions, fmt.Sprintf("%s: %.4g -> %.4g %s", k, av.Value, bv.Value, av.Units))
		}
	}
	sort.Strings(d.ChangedAssumptions)
	for k, av := range a.Outputs {
		if bv, ok := b.Outputs[k]; ok && av.UnavailableReason == "" && bv.UnavailableReason == "" {
			delta := bv.Value - av.Value
			d.OutputDeltas[k] = delta
			d.Attribution = append(d.Attribution, fmt.Sprintf("%s delta %.2f is driven by changed assumptions in %s profile", k, delta, b.Profile))
		}
	}
	sort.Strings(d.Attribution)
	return d
}

func PlainTextEvidence(title string, run ScenarioRun) string {
	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\nInputs:\n")
	keys := make([]string, 0, len(run.Inputs))
	for k := range run.Inputs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := run.Inputs[k]
		b.WriteString(fmt.Sprintf("- %s = %.4g %s [%s, c=%.2f]\n", v.Name, v.Value, v.Units, v.Tag.Provenance, v.Tag.Confidence))
	}
	b.WriteString("Outputs:\n")
	keys = keys[:0]
	for k := range run.Outputs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := run.Outputs[k]
		if v.UnavailableReason != "" {
			b.WriteString(fmt.Sprintf("- %s: Not available: %s\n", v.Name, v.UnavailableReason))
			continue
		}
		b.WriteString(fmt.Sprintf("- %s = %.2f %s [%s; qualitative source confidence only, not a statistical interval]\n", v.Name, v.Value, v.Units, v.Tag.Provenance))
	}
	return b.String()
}
