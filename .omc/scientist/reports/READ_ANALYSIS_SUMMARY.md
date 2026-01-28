# READ Operation Analysis - Executive Summary

## Research Stage 3: READ Operation Physics Comparison

**Date**: 2026-01-27  
**Analyst**: Scientist Agent  
**Scope**: Module 2 (Crossbar) vs Module 4 (Peripheral Circuits)

---

## Key Findings

### [FINDING] Kirchhoff's Current Law is the Fundamental Physics of READ
[STAT:implementation_locations] 3 (array.go:147, mvm.comp:131, mvm.comp:149)

Module 2 explicitly implements KCL for passive current summation at bit line nodes. The crossbar performs true analog computation where cell currents physically merge according to I_out = Σ(I_cell), with no active summing circuit required.

[STAT:kcl_evidence_strength] High (direct code comments + shader implementation)

### [FINDING] Peripheral Circuits Dominate READ Latency
[STAT:tia_settling_ns] 11.0  
[STAT:adc_conversion_ns] 50.0  
[STAT:total_read_latency_ns] 61.0  
[STAT:crossbar_stabilization_ns] 1.0

The peripheral sensing chain (TIA + ADC) is 61× slower than crossbar current stabilization. Circuit optimization (faster ADC, higher TIA bandwidth) is critical for throughput.

[STAT:max_throughput_mhz] 16.4

### [FINDING] IR Drop Creates Position-Dependent Errors up to 10%
[STAT:typical_ir_drop_percent] 3-10  
[STAT:worst_cell_location] [Rows-1, Cols-1]  
[STAT:ir_drop_scaling] O(N) with array size

Bottom-right cells experience maximum voltage reduction due to cumulative resistive drops along word lines and bit lines. Large arrays (>64×64) require mitigation strategies (wider metal, tiled architecture).

[STAT:64x64_worst_case_drop_mv] 31.5  
[STAT:64x64_worst_case_drop_percent] 3.15

### [FINDING] 1T1R Architecture Reduces Sneak Currents by 1000×
[STAT:sneak_ratio_0t1r] 1.0  
[STAT:sneak_ratio_1t1r] 0.001  
[STAT:isolation_improvement] 1000×

Passive crossbars (0T1R) suffer from sneak currents equal to signal current. Access transistors (1T1R) provide ~10⁶ ON/OFF ratio, reducing sneak to 0.1% of signal.

### [FINDING] Modules are Complementary, Not Redundant
[STAT:shared_code_interfaces] 0  
[STAT:module2_focus] Array-level physics  
[STAT:module4_focus] Circuit-level sensing

Module 2 answers "what current does the array produce?" while Module 4 answers "how do we measure and digitize it?". No programmatic integration exists - they operate in separate simulation domains.

### [FINDING] 5-Bit Peripherals Match 30-Level Cell Encoding with Margin
[STAT:cell_levels] 30  
[STAT:dac_bits] 5  
[STAT:adc_bits] 5  
[STAT:peripheral_codes] 32

Both DAC and ADC use 5 bits (32 codes), providing 2 extra levels beyond the 30-state ferroelectric encoding. This margin accommodates noise and nonlinearity.

[STAT:dac_inl_lsb] 0.5  
[STAT:dac_dnl_lsb] 0.25  
[STAT:adc_inl_lsb] 0.5  
[STAT:adc_dnl_lsb] 0.25  
[STAT:adc_enob] 4.77

---

## Signal Path Summary

### Physical Current Range
[STAT:g_min_microsiemens] 10  
[STAT:g_max_microsiemens] 100  
[STAT:i_min_microamps] 10  
[STAT:i_max_microamps] 100  
[STAT:v_read_volts] 0.1

### TIA Characteristics
[STAT:tia_gain_kohms] 10  
[STAT:tia_bandwidth_mhz] 100  
[STAT:tia_input_noise_pa_per_sqrt_hz] 1  
[STAT:tia_output_offset_mv] 5  
[STAT:tia_dynamic_range_db] 140  
[STAT:tia_snr_max_db] 80

### ADC Characteristics
[STAT:adc_resolution_bits] 5  
[STAT:adc_lsb_mv] 32.26  
[STAT:adc_conversion_time_ns] 50  
[STAT:adc_sqnr_db] 31.9

---

## Error Budget

| Error Source | Module | Typical Magnitude | Impact |
|--------------|--------|-------------------|--------|
| Cell quantization | 2 | 3.3% (1/30) | Fundamental limit |
| Device variation | 2 | 1-5% | Manufacturing |
| IR drop | 2 | 3-10% | Array size dependent |
| Sneak paths (0T1R) | 2 | 50-100% | Architecture critical |
| Sneak paths (1T1R) | 2 | 0.1% | Transistor isolation |
| Conductance drift | 2 | 1-5%/year | Temporal degradation |
| TIA noise | 4 | 0.01% | Negligible |
| TIA offset | 4 | 0.5% | Calibratable |
| ADC quantization | 4 | 3.1% (1/32) | Matched to cell |
| ADC nonlinearity | 4 | 1.5% (INL+DNL) | Circuit design |

[STAT:dominant_error_sources] 3 (IR drop, Sneak paths, Quantization)

---

## Timing Analysis

| Phase | Duration | Bottleneck |
|-------|----------|------------|
| Crossbar stabilization | 1 ns | RC time constant |
| TIA settling | 11 ns | Bandwidth-limited |
| ADC conversion | 50 ns | SAR architecture |
| **Total** | **61 ns** | ADC dominates |

[STAT:read_speedup_potential] 5.5× (if ADC improved to 10 ns)

---

## Kirchhoff's Laws in Action

### KCL (Current Law)
**Location**: Bit line nodes (column junctions)  
**Equation**: I_out[j] = Σᵢ I_cell[i,j]  
**Implementation**: Passive wire junction (no active circuit)  
**Purpose**: Enables analog MVM computation

### KVL (Voltage Law)
**Location**: Word line + cell + bit line loop  
**Equation**: V_driver - Σ(I×R_wl) - V_cell - Σ(I×R_bl) = 0  
**Implementation**: Iterative relaxation solver (`irdrop.go`)  
**Purpose**: Models resistive voltage drops

---

## Recommendations

### For Large-Scale Arrays (>64×64)
1. **Use 1T1R architecture** to eliminate sneak paths (1000× improvement)
2. **Widen metal lines** or use better conductors (Cu vs Al) to reduce IR drop
3. **Tile arrays** into smaller blocks to shorten current paths
4. **Implement IR-drop-aware training** to pre-compensate for position-dependent errors

### For High-Throughput Applications
1. **Upgrade ADC**: Use flash or pipeline architecture (10 ns vs 50 ns)
2. **Increase TIA bandwidth**: 500 MHz target (settling < 3 ns)
3. **Parallelize read channels**: 256+ columns simultaneously

### For Integration Work
1. **Create unified simulation**: Combine Module 2 crossbar output with Module 4 peripheral chain
2. **Add validation tests**: Verify crossbar currents → TIA → ADC produces expected codes
3. **Develop co-optimization**: Jointly tune array size, peripheral specs, and SNR

---

## Files Analyzed

### Module 2 (Crossbar)
- `/module2-crossbar/pkg/crossbar/array.go` (237 lines)
- `/module2-crossbar/pkg/crossbar/reference.go` (200 lines)
- `/module2-crossbar/pkg/crossbar/irdrop.go` (276 lines)
- `/module2-crossbar/pkg/crossbar/nonidealities.go` (414 lines)
- `/module2-crossbar/shaders/mvm.comp` (200+ lines, partial)

### Module 4 (Peripherals)
- `/module4-circuits/pkg/peripherals/tia.go` (101 lines)
- `/module4-circuits/pkg/peripherals/adc.go` (123 lines)
- `/module4-circuits/pkg/peripherals/dac.go` (90 lines)
- `/module4-circuits/pkg/peripherals/chargepump.go` (127 lines)
- `/module4-circuits/pkg/peripherals/analysis.go` (264 lines)

[STAT:total_lines_analyzed] 1833

---

## Limitations

[LIMITATION] No end-to-end simulation: Cannot trace a single READ operation from cell state through TIA and ADC in unified framework

[LIMITATION] TIA model simplified: DC gain + noise only, no frequency-domain transfer function or settling transients

[LIMITATION] ADC model idealized: SAR assumed perfect, no comparator offset or metastability effects

[LIMITATION] Modules operate independently: No shared test cases to validate Module 2 currents match Module 4 input assumptions

[LIMITATION] Analysis based on code review: Actual hardware may differ from simulation models

---

## Conclusion

The READ operation in FeCIM systems is a **multi-physics process** governed by:
1. **Ohm's Law** (cell current generation)
2. **Kirchhoff's Current Law** (passive summation enabling MVM)
3. **Transimpedance amplification** (current-to-voltage sensing)
4. **Quantization** (analog-to-digital conversion)

Module 2 and Module 4 are **complementary tools** representing different abstraction levels:
- **Module 2**: What current does the array produce? (includes non-idealities)
- **Module 4**: How do we sense and digitize it? (circuit-level details)

**Primary bottleneck**: ADC conversion time (50 ns) dominates total READ latency (61 ns).

**Primary accuracy limiter**: IR drop in large arrays (3-10%) and sneak paths in 0T1R (50-100%).

**Design recommendation**: Use 1T1R architecture with optimized peripherals (faster ADC, higher-BW TIA) for production systems.

---

**Report Generated**: 2026-01-27  
**Analysis Tool**: FeCIM Lattice Tools (Go + GPU compute)  
**Methodology**: Comparative code analysis + physics modeling
