package comparison

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

const ScenarioModelVersion = "scenario-estimates/2"
const ScenarioArtifactVersion = "fecim-comparison/2"
const ScenarioProtocol = "workload-rate-scenario"
const ArchitectureProtocol = "architecture-demonstration"
const HoursPerMonth = 730.0
const ElectricityUSDPerKWh = 0.10
const CO2UnavailableReason = "No physical energy-duration input supports scenario CO2; currency times kg/kWh is dimensionally invalid"
const ScenarioNotice = "Simulation/assumption only. Not validated silicon or measured M2/M3/M4 results. Source confidence is qualitative, not a statistical interval."

// ScenarioSources identifies the actual assumptions, not a claim that a public
// datasheet validates an aggregate per-MAC number.
func ScenarioSources() []string {
	return []string{
		"module5-comparison/pkg/comparison/scenario_profiles.go: bundled educational energy, latency, TCO and emissions-intensity assumptions; unchanged coefficients",
		"module5-comparison/pkg/gui/app.go: CPU + DRAM and GPU + HBM public-datasheet descriptions are context-only, not validated aggregate benchmarks; FeCIM is an assumed simulation baseline",
		"shared/physics/confidence_ledger.go: provenance vocabulary; fallback confidence tags are qualitative bookkeeping, not statistical intervals or calibration",
	}
}

type EnergyUse struct {
	JPerInference   float64 `json:"joules_per_inference"`
	DailyInferences float64 `json:"daily_inferences"`
	DailyKWh        float64 `json:"daily_kwh"`
	PowerW          float64 `json:"power_w"`
	MonthlyUSD      float64 `json:"monthly_usd"`
}

func finitePositive(v float64) bool { return v > 0 && !math.IsNaN(v) && !math.IsInf(v, 0) }
func EnergyAtRate(macs int64, energyFJ, rate float64) (EnergyUse, error) {
	if macs <= 0 || macs > 140_000_000_000_000 || !finitePositive(energyFJ) || !finitePositive(rate) || rate > 1e9 {
		return EnergyUse{}, errors.New("MACs, energy and inference rate require finite positive bounded inputs")
	}
	j := float64(macs) * energyFJ * 1e-15
	u := EnergyUse{JPerInference: j, DailyInferences: rate * 86400, DailyKWh: j * (rate * 86400) / 3.6e6, PowerW: j * rate, MonthlyUSD: j * rate / 1000 * HoursPerMonth * ElectricityUSDPerKWh}
	for _, v := range []float64{u.JPerInference, u.DailyKWh, u.PowerW, u.MonthlyUSD} {
		if !finitePositive(v) {
			return EnergyUse{}, errors.New("energy result is nonfinite or underflowed")
		}
	}
	return u, nil
}

// ScenarioWorkloadMACs preserves the calculator protocol, not the obsolete
// export lookup that substituted parameter counts for MACs.
func ScenarioWorkloadMACs(name string) (int64, error) {
	switch name {
	case "MNIST":
		return 101632, nil
	case "ResNet-50":
		return 4000000000, nil
	case "BERT-Base":
		return 11000000000, nil
	case "GPT-2":
		return 35000000000, nil
	case "LLM-70B":
		return 140000000000000, nil
	}
	return 0, errors.New("required workload MAC denominator is missing")
}
func ScenarioWorkloads() []string {
	return []string{"MNIST", "ResNet-50", "BERT-Base", "GPT-2", "LLM-70B"}
}
func ValidScenario(p ScenarioProfile) bool {
	return p == ScenarioConservative || p == ScenarioBaseline || p == ScenarioOptimistic
}

type ScenarioEstimate struct {
	Protocol                string          `json:"protocol"`
	Workload                string          `json:"workload"`
	MACsPerInference        int64           `json:"macs_per_inference"`
	InferencesPerSecond     float64         `json:"inferences_per_second"`
	Profile                 ScenarioProfile `json:"profile"`
	Config                  ScenarioConfig  `json:"assumptions"`
	Energy                  [3]EnergyUse    `json:"energy_cpu_gpu_fecim"`
	AnnualSavingsMillionUSD float64         `json:"annual_savings_million_usd"`
	Run                     ScenarioRun     `json:"scenario_evidence"`
}

func CalculateScenario(workload string, rate float64, profile ScenarioProfile) (ScenarioEstimate, error) {
	if !ValidScenario(profile) {
		return ScenarioEstimate{}, errors.New("unknown scenario profile")
	}
	macs, err := ScenarioWorkloadMACs(workload)
	if err != nil {
		return ScenarioEstimate{}, err
	}
	cfg := PresetScenarioConfig(profile)
	if err = cfg.Validate(); err != nil {
		return ScenarioEstimate{}, err
	}
	r := ScenarioEstimate{Protocol: ScenarioProtocol, Workload: workload, MACsPerInference: macs, InferencesPerSecond: rate, Profile: profile, Config: cfg, Run: BuildScenarioRun(profile, cfg)}
	for i, pj := range []float64{cfg.CPUEnergyPJPerMAC, cfg.GPUEnergyPJPerMAC, cfg.FeCIMEnergyPJPerMAC} {
		r.Energy[i], err = EnergyAtRate(macs, pj*1000, rate)
		if err != nil {
			return ScenarioEstimate{}, err
		}
	}
	r.AnnualSavingsMillionUSD = (r.Energy[1].DailyKWh - r.Energy[2].DailyKWh) * 365 * ElectricityUSDPerKWh / 1e6
	return r, nil
}
func MarshalScenario(result ScenarioEstimate, format string) ([]byte, error) {
	// Validate even native callers: serializers never turn NaN or absent sources into zero.
	if !ValidScenario(result.Profile) || result.Protocol != ScenarioProtocol {
		return nil, errors.New("unknown scenario protocol")
	}
	if err := result.Config.Validate(); err != nil {
		return nil, err
	}
	if _, err := ScenarioWorkloadMACs(result.Workload); err != nil {
		return nil, err
	}
	// In particular CSV must reject nonfinite canonical quantities instead of
	// emitting textual NaN/Inf cells that appear to be valid numeric exports.
	if _, err := json.Marshal(result); err != nil {
		return nil, err
	}
	switch format {
	case "json":
		return json.Marshal(struct {
			SchemaVersion string           `json:"schema_version"`
			ModelVersion  string           `json:"model_version"`
			Notice        string           `json:"notice"`
			Sources       []string         `json:"sources"`
			HoursPerMonth float64          `json:"monthly_protocol_hours"`
			USDPerKWh     float64          `json:"assumed_usd_per_kwh"`
			Result        ScenarioEstimate `json:"result"`
		}{ScenarioArtifactVersion, ScenarioModelVersion, ScenarioNotice, ScenarioSources(), HoursPerMonth, ElectricityUSDPerKWh, result})
	case "csv":
		var b bytes.Buffer
		w := csv.NewWriter(&b)
		w.Write([]string{"schema_version", "model_version", "protocol", "workload", "scenario", "architecture", "macs_per_inference", "inferences_per_second", "daily_inferences", "joules_per_inference", "daily_kwh", "power_w", "monthly_usd", "annual_savings_million_usd", "monthly_protocol_hours", "assumed_usd_per_kwh", "evidence", "co2_status", "sources"})
		f := func(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }
		for i, name := range []string{"CPU + DRAM", "GPU + HBM", "FeCIM"} {
			u := result.Energy[i]
			w.Write([]string{ScenarioArtifactVersion, ScenarioModelVersion, ScenarioProtocol, result.Workload, string(result.Profile), name, fmt.Sprint(result.MACsPerInference), f(result.InferencesPerSecond), f(u.DailyInferences), f(u.JPerInference), f(u.DailyKWh), f(u.PowerW), f(u.MonthlyUSD), f(result.AnnualSavingsMillionUSD), f(HoursPerMonth), f(ElectricityUSDPerKWh), ScenarioNotice, CO2UnavailableReason, strings.Join(ScenarioSources(), "; ")})
		}
		w.Flush()
		return b.Bytes(), w.Error()
	default:
		return nil, errors.New("enumerated formats are json and csv")
	}
}
