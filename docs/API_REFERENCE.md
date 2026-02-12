# FeCIM Lattice Tools API Reference

> Comprehensive API guide for the core simulation packages.
>
> Scope covered:
> - `shared/physics`
> - `shared/peripherals`
> - `module1-hysteresis/pkg/ferroelectric`
> - `module2-crossbar/pkg/crossbar`
> - `module3-mnist/pkg/core`

---

## Contents

1. [shared/physics](#sharedphysics)
2. [shared/peripherals](#sharedperipherals)
3. [module1-hysteresis/pkg/ferroelectric](#module1-hysteresispkgferroelectric)
4. [module2-crossbar/pkg/crossbar](#module2-crossbarpkgcrossbar)
5. [module3-mnist/pkg/core](#module3-mnistpkgcore)

---

## `shared/physics`

Import path:

```go
import "fecim-lattice-tools/shared/physics"
```

### Key types

- `type HZOMaterial struct { ... }`  
  Canonical ferroelectric material model (Pr/Ps/Ec, LK params, temperature and reliability parameters).

- `type LKSolver struct { ... }`  
  Landau-Khalatnikov dynamic solver for polarization state evolution.

- `type WriteController struct { ... }`  
  Closed-loop write engine that iteratively drives target conductance/polarization.

- `type WriteEvent struct { ... }`  
  Event payload for write-loop instrumentation/hooks.

- `type PreisachStack struct { ... }`  
  Generic Preisach hysteresis memory stack with wipe-out/turning-point logic.

- `type EverettFunction interface { Calculate(alpha, beta float64) float64 }`  
  Abstraction for Everett kernel used by Preisach stack.

- `type TurningPoint struct { E float64; IsRising bool }`  
  Stored turning point in Preisach history.

- `type ConductanceModel int`  
  Mapping mode for normalized→physical conductance (`Linear`, `Exponential`, `Lookup`).

- `type AdaptiveISPP struct { ... }`  
  Adaptive binary-search style ISPP controller.

- `type ISPPCalculator struct { ... }`  
  Legacy/analytic ISPP voltage planner and convergence evaluator.

- `type ISPPConfig struct { ... }`  
  ISPP tuning knobs (step sizes, limits, thresholds).

- `type HysteresisDirection int`  
  Enum for program direction (`Up`, `Down`, `None`).

- `type ISPPResult int`  
  Enum for write convergence outcomes.

- `type Calibrator struct { ... }`  
  Level-wise calibration engine with monotonic constraints and verify policy.

- `type VerifyResult struct { ... }`  
  Structured verify outcome from calibration checks.

- `type WriteVerifyStats struct { ... }`  
  Aggregated write/verify metrics (success rate, pulse histograms, overshoot).

- `type DeviceVariationEngine struct { ... }`  
  Spatially-correlated device variation generator for Ec/Pr drift.

- `type DeviceVariationConfig struct { ... }`  
  Config for random variation magnitudes/correlation.

- `type DeviceVariation struct { ... }`  
  Per-cell sampled variation factors.

- `type VariationStats struct { ... }`  
  Array-level variation statistics/yield summary.

- `type CellGeometry struct { ... }`  
  Geometric helper for field/charge/current conversions.

### Key functions and methods

#### Material presets and material methods

- `func DefaultHZO() *HZOMaterial`  
  Baseline HZO preset.

- `func FeCIMMaterial() *HZOMaterial`  
  FeCIM-oriented practical preset.

- `func FeCIMMaterialTarget() *HZOMaterial`  
  Target/aspirational FeCIM material preset.

- `func LiteratureSuperlattice() *HZOMaterial`  
  Literature superlattice preset.

- `func CryogenicHZO() *HZOMaterial`  
  Cryogenic operating-point preset.

- `func HZOStandard32() *HZOMaterial`  
  32-level style preset.

- `func HZOFJT140() *HZOMaterial`  
  140-state oriented preset.

- `func HZOCustom14() *HZOMaterial`  
  14-level compact preset.

- `func AlScN() *HZOMaterial`  
  AlScN material preset.

- `func AllMaterials() []*HZOMaterial`  
  Returns all built-in presets.

- `func (m *HZOMaterial) GetNumLevels() int`  
  Effective level count (fallback aware).

- `func (m *HZOMaterial) CoerciveVoltage() float64`  
  Coercive voltage from field × thickness.

- `func (m *HZOMaterial) SwitchingEnergy() float64`  
  Approximate switching energy per event.

- `func (m *HZOMaterial) SwitchingTime(T float64) float64`  
  Temperature-dependent switching time.

- `func (m *HZOMaterial) CoerciveFieldAtTemp(T float64) float64`  
  Ec(T) scaling helper.

- `func (m *HZOMaterial) PolarizationAtTemp(T float64) float64`  
  Pr(T) scaling helper.

- `func (m *HZOMaterial) EnduranceAtCycles(N float64) float64`  
  Endurance degradation model.

- `func (m *HZOMaterial) RetentionAtTime(t, T float64) float64`  
  Retention-vs-time/temperature estimate.

- `func (m *HZOMaterial) DiscreteLevel(level int, totalLevels int) float64`  
  Maps level index to normalized conductance target.

#### LK solver and write control

- `func NewLKSolver() *LKSolver`  
  Constructs solver with robust defaults.

- `func (s *LKSolver) ConfigureFromMaterial(mat *HZOMaterial)`  
  Loads LK coefficients from a material preset.

- `func (s *LKSolver) UpdateParams()`  
  Recomputes derived coefficients after edits.

- `func (s *LKSolver) Step(E, dt float64) float64`  
  Integrates one time-step under applied field.

- `func (s *LKSolver) SetState(P float64)` / `func (s *LKSolver) GetState() float64`  
  Explicit state management.

- `func NewWriteController(solver *LKSolver, material *HZOMaterial) *WriteController`  
  Creates closed-loop write controller.

- `func (c *WriteController) WriteTarget(targetG float64) (attempts int, success bool, overshootCount int)`  
  Drives to target conductance.

- `func (c *WriteController) WriteTargetWithReset(targetG float64, reset bool) (attempts int, success bool, overshootCount int)`  
  Optional pre-reset before write.

#### Preisach modeling

- `func NewPreisachStack(saturationE float64, everett EverettFunction) *PreisachStack`  
  Creates Preisach stack using user Everett kernel.

- `func (ps *PreisachStack) Update(E float64) float64`  
  Updates history and returns polarization.

- `func (ps *PreisachStack) ComputePolarization(currentE float64) float64`  
  Polarization at specific field using current history.

#### Quantization and transfer mappings

- `func QuantizeToLevels(value float64, levels int) float64`  
  Generic level quantizer.

- `func QuantizeTo30Levels(value float64) float64`  
  30-level shortcut.

- `func GetLevel(conductance float64, levels int) int`  
  Conductance→level index.

- `func LevelSpacing(levels int) float64`  
  Uniform spacing size.

- `func QuantizationError(levels int) float64`  
  Max quantization error estimate.

- `func PolarizationToConductance(P, Ps, Gmin, Gmax float64) float64`  
  Physics transfer function P→G.

- `func ConductanceToPolarization(G, Gmin, Gmax, Ps float64) float64`  
  Inverse transfer function G→P.

- `func NormalizedToPhysical(gNorm float64, model ConductanceModel) float64`  
  Converts normalized conductance to physical units.

- `func ConductanceToLevel(gPhys float64, levels int) int`  
  Physical conductance→level index.

- `func LevelToConductance(level, levels int, model ConductanceModel) float64`  
  Level→physical conductance.

#### ISPP and calibration helpers

- `func NewAdaptiveISPP(solver *LKSolver, mat *HZOMaterial) *AdaptiveISPP`  
  Adaptive write planner.

- `func (c *AdaptiveISPP) BinarySearchWrite(targetP float64) (float64, int, bool)`  
  Searches voltage to hit target polarization.

- `func NewISPPCalculator(ec float64, numLevels int) *ISPPCalculator`  
  Creates analytic ISPP calculator.

- `func NewISPPCalculatorWithConfig(ec float64, numLevels int, config ISPPConfig) *ISPPCalculator`  
  Creates calculator with custom config.

- `func (c *ISPPCalculator) CalculateNextVoltage(currentVoltage float64, direction HysteresisDirection) float64`  
  Next pulse voltage proposal.

- `func (c *ISPPCalculator) CheckResult(currentLevel, targetLevel int, direction HysteresisDirection, pulseCount int) ISPPResult`  
  Evaluates convergence/overshoot.

- `func NewCalibrator(numLevels int, Ec float64) *Calibrator`  
  Creates calibration table manager.

- `func (c *Calibrator) CheckVerify(targetLevel, readLevel, retryCount int) VerifyResult`  
  Verify policy decision.

#### Variation, geometry, units, and statistics

- `func DefaultDeviceVariationConfig() *DeviceVariationConfig`
- `func NewDeviceVariationEngine(config *DeviceVariationConfig) *DeviceVariationEngine`
- `func (e *DeviceVariationEngine) ApplyToMaterial(base *HZOMaterial, row, col int) *HZOMaterial`
- `func (e *DeviceVariationEngine) EstimateYield(rows, cols int, maxDeviation float64) float64`

- `func DefaultCellGeometry() CellGeometry`
- `func GeometryFromMaterial(mat *HZOMaterial) CellGeometry`
- `func (g CellGeometry) ElectricField(voltageV float64) float64`
- `func (g CellGeometry) ChargeFromPolarization(polarizationCPerM2 float64) float64`

- `func NewWriteVerifyStats() *WriteVerifyStats`
- `func (s *WriteVerifyStats) RecordWrite(targetLevel int, pulsesUsed int, success bool, hadOvershoot bool)`
- `func (s *WriteVerifyStats) GetSummary() string`
- `func SimulateFailureRateProgression(cycles int, enduranceLimit float64) float64`

- `func VPerMToMVPerCm(vPerM float64) float64`
- `func MVPerCmToVPerM(mvPerCm float64) float64`
- `func FormatEnergy(joules float64) string`
- `func FormatConductance(siemens float64) string`
- `func FormatElectricField(vm float64) string`

### Usage examples

#### Example 1: Material + LK solver + write control

```go
package main

import (
	"fmt"
	"fecim-lattice-tools/shared/physics"
)

func main() {
	mat := physics.FeCIMMaterial()
	solver := physics.NewLKSolver()
	solver.ConfigureFromMaterial(mat)

	wc := physics.NewWriteController(solver, mat)
	attempts, ok, overshoot := wc.WriteTarget(60e-6) // target 60 µS

	fmt.Printf("write ok=%v attempts=%d overshoot=%d\n", ok, attempts, overshoot)
}
```

#### Example 2: Preisach stack with custom Everett function

```go
package main

import "fecim-lattice-tools/shared/physics"

type linearEverett struct{}

func (linearEverett) Calculate(alpha, beta float64) float64 {
	if alpha <= beta {
		return 0
	}
	return (alpha - beta) / (2 * 2e8) // simple normalized placeholder
}

func main() {
	ps := physics.NewPreisachStack(2e8, linearEverett{})
	_ = ps.Update(-1e8)
	_ = ps.Update(0)
	_ = ps.Update(1e8)
}
```

#### Example 3: Quantization and transfer mapping

```go
package main

import (
	"fmt"
	"fecim-lattice-tools/shared/physics"
)

func main() {
	p := 0.18 // C/m²
	g := physics.PolarizationToConductance(p, 0.30, physics.GMin, physics.GMax)
	lvl := physics.ConductanceToLevel(g, 30)

	fmt.Printf("g=%s level=%d\n", physics.FormatConductance(g), lvl)
}
```

---

## `shared/peripherals`

Import path:

```go
import "fecim-lattice-tools/shared/peripherals"
```

### Key types

- `type DAC struct { ... }`  
  Write-path digital-to-analog converter with INL/DNL and energy models.

- `type ADC struct { ... }`  
  Read-path analog-to-digital converter (with optional SAR-specific noise model).

- `type SARNoiseConfig struct { ... }`  
  Configuration for thermal noise, metastability, and reference drift.

- `type ADCType int`  
  ADC architecture enum (`SAR`, `Flash`, `SigmaDelta`).

- `type TIA struct { ... }`  
  Transimpedance amplifier model for current-to-voltage conversion.

- `type ChargePump struct { ... }`  
  Voltage booster model used for write pulses.

- `type SampleAndHold struct { ... }`  
  Front-end sample/hold model with settling and droop.

- `type VoltageRegulator struct { ... }`  
  Supply regulator model for load/ripple handling.

- `type ProcessCorner string`  
  Process corner enum used in PVT-aware conversion.

- `type INLDNLAnalysis struct { ... }`  
  Code-level linearity report.

- `type PVTINLDNLAnalysis struct { ... }`  
  Linearity report under given PVT condition.

- `type ProcessCornerAnalysis struct { ... }`  
  Cross-corner linearity comparison.

- `type TimingAnalysis struct { ... }`  
  Latency/throughput breakdown for peripheral chain.

- `type PowerBreakdown struct { ... }`  
  Energy and power split across blocks.

- `type TransferFunction struct { ... }`  
  End-to-end DAC→TIA→ADC transfer behavior.

### Key functions and methods

#### Constructors and defaults

- `func DefaultDAC() *DAC`
- `func DefaultADC() *ADC`
- `func DefaultTIA() *TIA`
- `func DefaultChargePump() *ChargePump`
- `func NegativePump() *ChargePump`
- `func DefaultSampleAndHold() *SampleAndHold`
- `func DefaultVoltageRegulator() *VoltageRegulator`
- `func DefaultSARNoiseConfig() *SARNoiseConfig`

#### DAC API

- `func (d *DAC) Levels() int`  
  Number of output codes.

- `func (d *DAC) Convert(level int) float64`  
  Ideal code→voltage conversion.

- `func (d *DAC) ConvertWithNonlinearity(level int) float64`  
  Includes INL/DNL effects.

- `func (d *DAC) ConvertWithCondition(level int, tempK float64, corner ProcessCorner) float64`  
  PVT-aware conversion.

- `func (d *DAC) Resolution() float64`  
  LSB size.

- `func (d *DAC) VoltageRange() (min, max float64)`  
  DAC full-scale range.

- `func (d *DAC) EnergyPerConversion() float64`  
  Energy estimate per conversion.

- `func (d *DAC) AnalyzeINLDNL() *INLDNLAnalysis`  
  Detailed linearity report.

#### ADC API

- `func (a *ADC) Levels() int`
- `func (a *ADC) Convert(voltage float64) int`
- `func (a *ADC) ConvertWithNonlinearity(voltage float64) int`
- `func (a *ADC) ConvertWithCondition(voltage float64, tempK float64, corner ProcessCorner) int`
- `func (a *ADC) Resolution() float64`
- `func (a *ADC) ENOB() float64`
- `func (a *ADC) TheoreticalSNR() float64`
- `func (a *ADC) EffectiveSNR() float64`
- `func (a *ADC) EnergyPerConversion() float64`
- `func (a *ADC) AnalyzeINLDNL() *INLDNLAnalysis`

SAR-noise extensions:

- `func (a *ADC) EnableSARNoise()`
- `func (a *ADC) DisableSARNoise()`
- `func (a *ADC) SetTemperature(tempK float64)`
- `func (a *ADC) GetEffectiveVref() (vrefLow, vrefHigh float64)`
- `func (a *ADC) GetThermalNoiseVoltage() float64`
- `func (a *ADC) GetMetastabilityErrorRate(inputVoltage float64, thresholdVoltage float64) float64`
- `func (a *ADC) ConvertWithSARNoise(voltage float64, seed int64) int`
- `func (a *ADC) GetSARNoiseReport() map[string]float64`

#### TIA / charge-pump / utility APIs

- `func (t *TIA) Convert(current float64) float64`
- `func (t *TIA) ConvertWithNoise(current float64) float64`
- `func (t *TIA) SNR(current float64) float64`
- `func (t *TIA) SettlingTime() float64`
- `func (t *TIA) PowerConsumption() float64`

- `func (c *ChargePump) IdealOutputVoltage() float64`
- `func (c *ChargePump) ActualOutputVoltage() float64`
- `func (c *ChargePump) OutputRipple() float64`
- `func (c *ChargePump) EnergyPerOperation(pulseDuration float64) float64`
- `func (c *ChargePump) MaxCurrentCapability() float64`

- `func (s *SampleAndHold) SettledFraction(tSeconds float64) float64`
- `func (s *SampleAndHold) HoldDroop(tSeconds float64) float64`

- `func (r *VoltageRegulator) Regulate(vin, loadCurrent float64) float64`
- `func (r *VoltageRegulator) SupplyNoiseTransfer(vinRipple float64) float64`

#### System-level analysis APIs

- `func EffectiveINLDNL(inl, dnl, tempK float64, corner ProcessCorner) (float64, float64)`
- `func AnalyzeINLDNLAtCondition(dac *DAC, adc *ADC, temperatureK float64, corner ProcessCorner) *PVTINLDNLAnalysis`
- `func AnalyzeProcessCorners(dac *DAC, adc *ADC, temperatureK float64) *ProcessCornerAnalysis`
- `func AnalyzeTiming(dac *DAC, adc *ADC, tia *TIA, pump *ChargePump) *TimingAnalysis`
- `func AnalyzePower(dac *DAC, adc *ADC, tia *TIA, pump *ChargePump, timing *TimingAnalysis) *PowerBreakdown`
- `func ComputeTransferFunction(dac *DAC, adc *ADC, tia *TIA, pump *ChargePump) *TransferFunction`
- `func BuildBehavioralSpiceSubcircuits(dac *DAC, adc *ADC, tia *TIA, sh *SampleAndHold, vr *VoltageRegulator) string`

### Usage examples

#### Example 1: DAC + ADC round-trip

```go
package main

import (
	"fmt"
	"fecim-lattice-tools/shared/peripherals"
)

func main() {
	dac := peripherals.DefaultDAC()
	adc := peripherals.DefaultADC()

	v := dac.Convert(12)
	code := adc.Convert(v)
	fmt.Printf("level=12 -> %.4f V -> code=%d\n", v, code)
}
```

#### Example 2: PVT and timing/power analysis

```go
package main

import (
	"fmt"
	"fecim-lattice-tools/shared/peripherals"
)

func main() {
	dac := peripherals.DefaultDAC()
	adc := peripherals.DefaultADC()
	tia := peripherals.DefaultTIA()
	pump := peripherals.DefaultChargePump()

	timing := peripherals.AnalyzeTiming(dac, adc, tia, pump)
	power := peripherals.AnalyzePower(dac, adc, tia, pump, timing)
	fmt.Printf("cycle(ns)=%.2f total_energy(J)=%.3e\n", timing.CycleTime, power.TotalEnergy)
}
```

#### Example 3: Generate behavioral SPICE stubs

```go
package main

import (
	"fmt"
	"fecim-lattice-tools/shared/peripherals"
)

func main() {
	netlist := peripherals.BuildBehavioralSpiceSubcircuits(
		peripherals.DefaultDAC(),
		peripherals.DefaultADC(),
		peripherals.DefaultTIA(),
		peripherals.DefaultSampleAndHold(),
		peripherals.DefaultVoltageRegulator(),
	)
	fmt.Println(netlist)
}
```

---

## `module1-hysteresis/pkg/ferroelectric`

Import path:

```go
import "fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
```

### Key types

- `type HZOMaterial = sharedphysics.HZOMaterial`  
  Backward-compatible alias of shared material model.

- `type PreisachModel struct { ... }`  
  High-level hysteresis model wrapper around shared Preisach engine.

- `type TanhEverett struct { Ps, Ec, Delta float64 }`  
  Everett kernel implementation using tanh major-loop approximation.

- `type DiscreteState struct { ... }`  
  One programmable state with level, polarization, voltage, and conductance.

- `type LevelBins struct { ... }`  
  Quantization bin helper with guard-band awareness.

- `type PERenderer struct { Width, Height int; Color bool }`  
  ASCII renderer for loops/domain/switching plots.

### Key functions and methods

#### Material wrappers (re-exported)

- `func DefaultHZO() *HZOMaterial`
- `func FeCIMMaterial() *HZOMaterial`
- `func FeCIMMaterialTarget() *HZOMaterial`
- `func LiteratureSuperlattice() *HZOMaterial`
- `func CryogenicHZO() *HZOMaterial`
- `func HZOStandard32() *HZOMaterial`
- `func HZOFJT140() *HZOMaterial`
- `func HZOCustom14() *HZOMaterial`
- `func AlScN() *HZOMaterial`
- `func AllMaterials() []*HZOMaterial`

#### Preisach model API

- `func NewPreisachModel(material *HZOMaterial) *PreisachModel`  
  Creates hysteresis model with material-derived settings.

- `func (p *PreisachModel) Update(E float64) float64`  
  Applies field step and returns polarization.

- `func (p *PreisachModel) Polarization() float64`  
  Current polarization.

- `func (p *PreisachModel) NormalizedPolarization() float64`  
  Polarization normalized by saturation.

- `func (p *PreisachModel) Reset()`  
  Resets to negative saturation state.

- `func (p *PreisachModel) GetHysteresisLoop(Emax float64, points int) ([]float64, []float64)`  
  Generates full major loop arrays.

- `func (p *PreisachModel) SetTemperature(tempK float64)`  
  Applies thermal parameter scaling.

- `func (p *PreisachModel) SetStress(stressGPa float64)`  
  Applies mechanical stress coupling.

- `func (p *PreisachModel) GetEffectiveEc() float64`  
  Returns current effective coercive field.

- `func (p *PreisachModel) DiscreteStates(n int) []DiscreteState`  
  Samples n evenly spaced programmable states.

#### Binning and rendering API

- `func NewLevelBins(ps float64, numLevels int, rangeFrac float64, guardFrac float64) LevelBins`
- `func (b LevelBins) EffectivePs() float64`
- `func (b LevelBins) Step() float64`
- `func (b LevelBins) LevelForP(P float64) (level int, inError bool, delta float64)`

- `func NewPERenderer() *PERenderer`
- `func (r *PERenderer) RenderPELoop(E, P []float64, material *HZOMaterial) string`
- `func (r *PERenderer) RenderDomainStates(alphas, betas []float64, states []int) string`
- `func (r *PERenderer) RenderDiscreteStates(states []DiscreteState) string`
- `func (r *PERenderer) RenderSwitchingDynamics(times, pols []float64, switched []int, material *HZOMaterial) string`
- `func (r *PERenderer) RenderTemperatureDependence(material *HZOMaterial) string`
- `func (r *PERenderer) RenderMaterialComparison() string`

### Usage examples

#### Example 1: Generate and inspect a P-E loop

```go
package main

import (
	"fmt"
	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

func main() {
	m := ferroelectric.NewPreisachModel(ferroelectric.FeCIMMaterial())
	E, P := m.GetHysteresisLoop(2e8, 200)
	fmt.Printf("points=%d firstP=%g lastP=%g\n", len(E), P[0], P[len(P)-1])
}
```

#### Example 2: Quantize polarization into guarded bins

```go
package main

import (
	"fmt"
	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

func main() {
	bins := ferroelectric.NewLevelBins(0.30, 30, 0.90, 0.20)
	level, inError, delta := bins.LevelForP(0.12)
	fmt.Printf("level=%d guard=%v delta=%g\n", level, inError, delta)
}
```

#### Example 3: Render loop as ASCII

```go
package main

import (
	"fmt"
	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

func main() {
	model := ferroelectric.NewPreisachModel(ferroelectric.DefaultHZO())
	E, P := model.GetHysteresisLoop(2e8, 120)
	plot := ferroelectric.NewPERenderer().RenderPELoop(E, P, ferroelectric.DefaultHZO())
	fmt.Println(plot)
}
```

---

## `module2-crossbar/pkg/crossbar`

Import path:

```go
import "fecim-lattice-tools/module2-crossbar/pkg/crossbar"
```

### Key types

- `type Config struct { ... }`  
  Top-level array config (dimensions, quantization, non-idealities, etc).

- `type Array struct { ... }`  
  Main crossbar array model (programming, MVM/VMM, statistics, exports).

- `type Cell struct { ... }`  
  Per-cell state and metadata.

- `type CellStats struct { ... }`  
  Per-cell derived write/disturb stats.

- `type EnduranceConfig struct { ... }`  
  Endurance-fatigue knobs.

- `type ProcessVariationConfig struct { ... }`  
  Process-gradient/edge variation controls.

- `type HalfSelectConfig struct { ... }`  
  Half-select disturb coupling config.

- `type MVMOptions struct { ... }`  
  Advanced MVM pipeline options (IR drop, sneak, temperature profile, etc).

- `type MVMResult struct { ... }`  
  Rich MVM output incl. error and energy metrics.

- `type AnalysisReport struct { ... }`  
  High-level report snapshot for a run.

- `type AccuracyDegradation struct { ... }` / `type DegradationStep struct { ... }`  
  Accuracy drop progression under cumulative non-idealities.

- `type SORConfig struct { ... }`  
  Iterative parasitic solver settings.

- `type ParasiticSolver struct { ... }`  
  SOR-based IR-drop/parasitic MVM solver.

- `type OptimizedParasiticSolver struct { ... }`  
  Allocation-optimized SOR solver variant.

- `type ParasiticMVMResult struct { ... }`  
  Detailed solve result (iterations, convergence, effective outputs).

- `type IRDropSimulator struct { ... }`  
  Time-domain IR drop simulator.

- `type SneakPathAnalyzer struct { ... }`  
  Dedicated sneak-path analysis engine.

- `type DriftSimulator struct { ... }`  
  Conductance drift evolution simulator.

- `type DeviceErrorEngine struct { ... }`  
  Programming/read error injector.

- `type WriteDisturbEngine struct { ... }`  
  Half-select stress/disturb tracker.

- `type GPUAccelerator struct { ... }`  
  Optional GPU MVM backend.

- `type TemperatureEffects struct { ... }` / `type ThermalPhysicsModel struct { ... }`  
  Thermal scaling and reliability modeling utilities.

### Key functions and methods

#### Array lifecycle and core ops

- `func NewArray(cfg *Config) (*Array, error)`
- `func (a *Array) Destroy()`
- `func (a *Array) Rows() int`
- `func (a *Array) Cols() int`
- `func (a *Array) GetConfig() Config`
- `func (a *Array) GetStats() (reads, writes int64)`

- `func (a *Array) ProgramWeight(row, col int, weight float64) error`
- `func (a *Array) ProgramWeightMatrix(weights [][]float64) error`
- `func (a *Array) ProgramWeightWithDisturb(row, col int, weight float64, isPassive bool) error`
- `func (a *Array) ProgramWeightWithVariation(row, col int, targetLevel int, stats *WriteStatistics) (int, error)`

- `func (a *Array) MVM(input []float64) ([]float64, error)`
- `func (a *Array) VMM(input []float64) ([]float64, error)`
- `func (a *Array) MVMWithNonIdealities(input []float64, opts *MVMOptions) (*MVMResult, error)`

- `func (a *Array) GetConductanceMatrix() [][]float64`
- `func (a *Array) GetEffectiveConductanceMatrix() [][]float64`
- `func (a *Array) GetPhysicalConductance(gNorm float64) float64`
- `func (a *Array) GetPhysicalConductanceForCell(row, col int) float64`

#### Quantization and compatibility helpers

- `func QuantizeToLevels(value float64) float64`  
  30-level normalized quantizer.

- `func GetLevel(conductance float64) int`  
  30-level index conversion.

#### Non-ideal analysis and reporting

- `func (a *Array) AnalyzeRCDelay(params *WireParams, inputVoltage float64) *RCDelayAnalysis`
- `func (a *Array) AnalyzeIRDrop(input []float64, params *WireParams) *IRDropAnalysis`
- `func (a *Array) AnalyzeIRDropIterative(input []float64, params *WireParams, config *IRDropSolverConfig) *IRDropAnalysis`
- `func (a *Array) AnalyzeSneakPaths(selectedRow, selectedCol int) *SneakPathAnalysis`
- `func (a *Array) AnalyzeSneakPathsWithArch(selectedRow, selectedCol int, is1T1R bool) *SneakPathAnalysis`

- `func (a *Array) GenerateMVMSneakTrace(input []float64, opts *MVMOptions, maxPaths int) *MVMSneakTraceReport`
- `func (r *MVMSneakTraceReport) FormatText(maxRows, maxPaths int) string`
- `func (a *Array) AnalyzeSneakContributions(targetRow int, input []float64, maxPaths int) []SneakPathContribution`

- `func (a *Array) GenerateReport(mvmResult *MVMResult) *AnalysisReport`
- `func (a *Array) ExportWeightsCSV(path string) error`
- `func (a *Array) ExportAnalysisJSON(path string, mvmResult *MVMResult) error`

- `func (a *Array) ComputeAccuracyDegradation(input []float64, baselineAccuracy float64) (*AccuracyDegradation, error)`
- `func (a *Array) ComputeAccuracyDegradationWithOptions(input []float64, baselineAccuracy float64, opts *MVMOptions) (*AccuracyDegradation, error)`

#### Parasitic solver API

- `func DefaultSORConfig() *SORConfig`
- `func NewParasiticSolver(rows, cols int, config *SORConfig) (*ParasiticSolver, error)`
- `func (s *ParasiticSolver) SetConductances(g [][]float64)`
- `func (s *ParasiticSolver) SetParasitics(rpRow, rpCol float64)`
- `func (s *ParasiticSolver) SolveMVM(appliedVoltages []float64) (*ParasiticMVMResult, error)`
- `func (s *ParasiticSolver) SolveMVMWithFallback(appliedVoltages []float64) (*ParasiticMVMResult, error)`
- `func (s *ParasiticSolver) ComputeIdealMVM(appliedVoltages []float64) []float64`
- `func (s *ParasiticSolver) AnalyzeParasiticImpact(appliedVoltages []float64) (*ParasiticImpact, error)`

- `func NewOptimizedParasiticSolver(rows, cols int, config *SORConfig) (*OptimizedParasiticSolver, error)`
- `func (s *OptimizedParasiticSolver) SolveMVM(appliedVoltages []float64) (*ParasiticMVMResult, error)`
- `func (s *OptimizedParasiticSolver) SolveMVMFast(appliedVoltages []float64) ([]float64, int, error)`

#### IR drop / sneak path / drift / error engines

- `func NewIRDropSimulator(rows, cols int) *IRDropSimulator`
- `func (ir *IRDropSimulator) SetConductance(row, col int, g float64)`
- `func (ir *IRDropSimulator) SetAllInputs(voltages []float64)`
- `func (ir *IRDropSimulator) Simulate(iterations int)`
- `func (ir *IRDropSimulator) GetOutputCurrents() []float64`
- `func (ir *IRDropSimulator) GetStats() IRDropStats`

- `func NewSneakPathAnalyzer(rows, cols int) *SneakPathAnalyzer`
- `func (sp *SneakPathAnalyzer) AnalyzeTarget(targetRow, targetCol int, voltage float64)`
- `func (sp *SneakPathAnalyzer) GetStats(voltage float64) SneakPathStats`

- `func NewDriftSimulator(rows, cols int, levels int) *DriftSimulator`
- `func NewDriftSimulatorWithModel(rows, cols int, levels int, model DriftModel) *DriftSimulator`
- `func (d *DriftSimulator) SimulateTimeStep(dt float64)`
- `func (d *DriftSimulator) GetStats() DriftStats`
- `func CompareTechnologies(rows, cols int, simulationTime float64) map[string]DriftStats`

- `func NewDeviceErrorEngine(progConfig *ProgrammingErrorConfig, readConfig *ReadNoiseConfig) *DeviceErrorEngine`
- `func (e *DeviceErrorEngine) ApplyProgrammingError(gTarget float64) float64`
- `func (e *DeviceErrorEngine) ApplyReadNoise(gProgrammed float64, row, col int) float64`
- `func ComputeErrorStatistics(target, actual [][]float64) *ErrorStatistics`

- `func NewWriteDisturbEngine(rows, cols int, config *WriteDisturbConfig) *WriteDisturbEngine`
- `func (e *WriteDisturbEngine) RecordWrite(targetRow, targetCol int)`
- `func (e *WriteDisturbEngine) ApplyDisturbEffects(conductances [][]float64, levels int) int`
- `func (e *WriteDisturbEngine) GetStressStats() WriteDisturbStats`

#### Standalone utility functions

- `func ComputeError(ideal, actual []float64) float64`
- `func ComputeAccuracyLoss(ideal, actual []float64) float64`
- `func SimulateAccuracyDegradation(progSigma, readSigma float64, arraySize int) float64`
- `func RecommendErrorBudget(targetAccuracy float64, arraySize int) (progSigma, readSigma float64)`
- `func EstimateDisturbRate(writesPerCell float64, config *WriteDisturbConfig) float64`
- `func CompareArchitectures(writesPerCell float64) (passiveRate, activeRate float64)`
- `func HalfSelectVoltage(writeVoltage float64, scheme string) float64`
- `func IsDisturbCritical(halfSelectV, coerciveV, safetyMargin float64) bool`
- `func NewGPUAccelerator(maxRows, maxCols int) (*GPUAccelerator, error)`

### Usage examples

#### Example 1: Program matrix and run MVM

```go
package main

import (
	"fmt"
	"fecim-lattice-tools/module2-crossbar/pkg/crossbar"
)

func main() {
	cfg := &crossbar.Config{Rows: 2, Cols: 3}
	arr, err := crossbar.NewArray(cfg)
	if err != nil {
		panic(err)
	}
	defer arr.Destroy()

	_ = arr.ProgramWeightMatrix([][]float64{
		{0.2, 0.7, 0.1},
		{0.9, 0.4, 0.6},
	})

	y, _ := arr.MVM([]float64{1.0, 0.5})
	fmt.Println(y)
}
```

#### Example 2: Non-ideality-aware MVM

```go
package main

import (
	"fmt"
	"fecim-lattice-tools/module2-crossbar/pkg/crossbar"
)

func main() {
	arr, _ := crossbar.NewArray(&crossbar.Config{Rows: 4, Cols: 4})
	defer arr.Destroy()

	opts := crossbar.DefaultMVMOptions()
	res, _ := arr.MVMWithNonIdealities([]float64{1, 0, 1, 0}, opts)
	fmt.Printf("rmse=%.4f energy=%.3e\n", res.RMSE, res.EnergyJ)
}
```

#### Example 3: Parasitic solver use

```go
package main

import (
	"fmt"
	"fecim-lattice-tools/module2-crossbar/pkg/crossbar"
)

func main() {
	s := crossbar.DefaultSORConfig()
	solver, _ := crossbar.NewParasiticSolver(8, 8, s)
	solver.SetParasitics(2.5, 2.5)

	result, err := solver.SolveMVM(make([]float64, 8))
	if err != nil {
		panic(err)
	}
	fmt.Printf("iters=%d converged=%v\n", result.Iterations, result.Converged)
}
```

---

## `module3-mnist/pkg/core`

Import path:

```go
import "fecim-lattice-tools/module3-mnist/pkg/core"
```

### Key types

- `type NetworkConfig struct { ... }`  
  Runtime configuration for quantization, noise, ADC/DAC bits, and mode options.

- `type DualModeNetwork struct { ... }`  
  Main MNIST inference model exposing FP and CIM paths.

- `type InferenceResult struct { ... }`  
  Per-sample FP vs CIM outputs, predictions, agreement, and energy metrics.

- `type WeightsFile struct { ... }`  
  JSON schema for stored weights (+ optional quant metadata).

- `type EnergyEstimate struct { ... }`  
  Detailed energy accounting for one inference.

- `type QuantizationStats struct { ... }`  
  Statistical quality metrics comparing original vs quantized weights.

- `type CIMNoiseComponents struct { ... }`  
  Decomposed physical noise components for CIM path.

- `type TIAModel struct { ... }`  
  Lightweight TIA transfer/bandwidth helper used by CIM physics path.

- `type ModeMetrics struct { ... }`  
  Confusion-matrix + aggregate metrics for one inference mode.

- `type DualModeDatasetMetrics struct { ... }`  
  Dataset-level FP/CIM metrics and agreement summary.

- `type RandomSource struct { ... }`  
  Reproducible RNG wrapper for quantization/noise.

- Interfaces:
  - `type Inferer interface { ... }`
  - `type WeightLoader interface { ... }`
  - `type WeightProvider interface { ... }`
  - `type NetworkConfigurer interface { ... }`
  - `type DataLoader interface { ... }`
  - `type Network interface { ... }`

### Key functions and methods

#### Network creation and configuration

- `func DefaultNetworkConfig() *NetworkConfig`
- `func NewDualModeNetwork(inputSize, hiddenSize, outputSize int) *DualModeNetwork`

- `func (net *DualModeNetwork) SetNumLevels(levels int)`
- `func (net *DualModeNetwork) GetNumLevels() int`
- `func (net *DualModeNetwork) SetPerLayerQuant(enabled bool)`
- `func (net *DualModeNetwork) IsPerLayerQuant() bool`
- `func (net *DualModeNetwork) SetPerLayerLevels(layer1, layer2 int)`
- `func (net *DualModeNetwork) GetPerLayerQuantInfo() (enabled bool, l1Levels, l2Levels int)`

- `func (net *DualModeNetwork) SetNoiseLevel(noise float64)`
- `func (net *DualModeNetwork) SetADCBits(bits int)`
- `func (net *DualModeNetwork) SetDACBits(bits int)`
- `func (net *DualModeNetwork) SetSingleLayer(enabled bool)`
- `func (net *DualModeNetwork) IsSingleLayer() bool`

#### Inference API

- `func (net *DualModeNetwork) Infer(input []float64) *InferenceResult`  
  Runs both FP and CIM paths.

- `func (net *DualModeNetwork) InferFPOnly(input []float64) (prediction int, confidence float64, probs []float64)`
- `func (net *DualModeNetwork) InferCIMOnly(input []float64) (prediction int, confidence float64, probs []float64)`

- `func EvaluateDualModeDataset(net *DualModeNetwork, images [][]float64, labels []int) DualModeDatasetMetrics`

#### Weights and quantization

- `func (net *DualModeNetwork) LoadWeights(filename string) error`
- `func (net *DualModeNetwork) LoadWeightsForLevel(dataDir string, levels int) error`
- `func (net *DualModeNetwork) RequantizeWeights()`

- `func ScanAvailableQATLevels(dataDir string) []int`
- `func GetWeightsFilename(dataDir string, levels int) string`
- `func GetBestMatchingWeightsLevel(dataDir string, targetLevels int) int`

- `func QuantizeWeights(fpWeights [][]float64, levels int) ([][]float64, error)`
- `func QuantizeBias(fpBias []float64, levels int) ([]float64, error)`
- `func ComputeQuantizationStats(original, quantized [][]float64) QuantizationStats`

- `func (net *DualModeNetwork) GetQuantizationStats() (layer1Stats, layer2Stats QuantizationStats)`
- `func (net *DualModeNetwork) GetFPWeights() (w1, w2 [][]float64, b1, b2 []float64)`
- `func (net *DualModeNetwork) GetQuantWeights() (w1, w2 [][]float64, b1, b2 []float64)`

#### Noise and energy modeling

- `func AddGaussianNoise(values []float64, noiseLevel float64, rng *RandomSource) []float64`
- `func AddGaussianNoiseInPlace(values []float64, noiseLevel float64, rng *RandomSource)`
- `func NewRandomSource(seed uint64) *RandomSource`

- `func (n CIMNoiseComponents) TotalSigma() float64`
- `func EnergyPerMACJ(levels int) float64`
- `func EstimateInferenceEnergyJ(cfg *NetworkConfig, inputSize, hiddenSize, outputSize int) EnergyEstimate`
- `func EstimateInferenceEnergyMicroJ(cfg *NetworkConfig, inputSize, hiddenSize, outputSize int) float64`

#### GPU hooks and notifications

- `func InitGPU()`
- `func IsGPUAvailable() bool`
- `func DestroyGPU()`
- `func (net *DualModeNetwork) SetUseGPU(use bool)`
- `func (net *DualModeNetwork) UseGPU() bool`

- `func (net *DualModeNetwork) SetNotificationHandler(handler func(message string))`

### Usage examples

#### Example 1: Run dual-mode inference

```go
package main

import (
	"fmt"
	"fecim-lattice-tools/module3-mnist/pkg/core"
)

func main() {
	net := core.NewDualModeNetwork(784, 128, 10)
	net.SetNumLevels(30)
	net.SetNoiseLevel(0.03)

	input := make([]float64, 784) // replace with normalized MNIST sample
	res := net.Infer(input)
	if res == nil {
		panic("invalid input")
	}
	fmt.Printf("fp=%d cim=%d agree=%v\n", res.FPPrediction, res.CIMPrediction, res.Agreement)
}
```

#### Example 2: Load level-matched weights and evaluate energy

```go
package main

import (
	"fmt"
	"fecim-lattice-tools/module3-mnist/pkg/core"
)

func main() {
	cfg := core.DefaultNetworkConfig()
	cfg.NumLevels = 30

	uJ := core.EstimateInferenceEnergyMicroJ(cfg, 784, 128, 10)
	fmt.Printf("estimated energy = %.4f µJ\n", uJ)
}
```

#### Example 3: Quantize custom weight tensor

```go
package main

import (
	"fmt"
	"fecim-lattice-tools/module3-mnist/pkg/core"
)

func main() {
	w := [][]float64{{-0.5, 0.0, 0.8}, {0.3, -0.2, 0.1}}
	q, err := core.QuantizeWeights(w, 30)
	if err != nil {
		panic(err)
	}
	stats := core.ComputeQuantizationStats(w, q)
	fmt.Printf("mse=%g max_abs=%g\n", stats.MSE, stats.MaxAbsError)
}
```

---

## Notes and conventions

- This document intentionally focuses on **public, high-value API surfaces** and omits internal/private helpers.
- All snippets are minimal and intended as starting points; production code should include validation, error handling, and deterministic seeds where applicable.
- For package internals and implementation details, inspect source files directly in the corresponding package directories.
