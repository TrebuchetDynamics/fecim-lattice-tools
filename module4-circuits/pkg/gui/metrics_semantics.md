# Module4 Read-Mode Metrics Semantics

This table documents the numeric metrics shown in **READ mode** (sense panel + selected-cell info line).

| UI label | Symbol | Physical quantity | Unit | Formula | Source in code |
|---|---|---|---|---|---|
| `I_cell (µA)` | \(I_{cell}\) | Sensed cell/row current shown in panel value | µA | \(I = G_{cell} \times V_{read}\) (panel value currently from selected row current) | `tab_unified.go:updateSensePanel` (`GetRowCurrent`), model current from `DeviceState.Compute` |
| `V_TIA (V)` | \(V_{TIA}\) | TIA output voltage | V | \(V_{TIA}=I_{cell}\times R_f + V_{ref}\), then clipped by rails | `arraysim/sensechain.go:ConvertCurrent`, `shared/peripherals/tia.go:Convert` |
| `ADC Code (0–2^N-1)` | \(Code\) | Quantized ADC output code | code (dimensionless) | \(Code \approx \lfloor V_{TIA}/LSB_V \rfloor\) for non-edge values; implementation uses nearest-integer quantization with clamp | `arraysim/sensechain.go:ConvertCurrent`, `shared/peripherals/adc.go:Convert` |
| `SNR (dB)` | \(SNR\) | Signal-to-noise ratio | dB | \(20\log_{10}(I_{signal}/I_{noise})\) | `tab_unified.go:composedSenseSNRdB` |
| `Status (region)` | — | Sense-chain operating region | linear/clipped | Linear or saturated depending on TIA/ADC clipping flags | `tab_unified.go:updateSensePanel`, saturation from `DeviceState.IsSaturated` |
| `I_range (µA)` | \(I_{min}..I_{max}\) | Measurable input current window | µA | \(I=(V-V_{ref})/R_f\) over effective voltage window | `arraysim/sensechain.go:CurrentRange` |
| `I_LSB (µA/code)` | \(I_{LSB}\) | Current represented by one code step | µA/code | \(I_{LSB}=I_{range}/2^N\) (UI semantic); implementation uses effective span over \((2^N-1)\) codes | `arraysim/sensechain.go:CurrentLSB` |
| `Cell [r,c]: State L/(Lmax)` | \(L\) | Discrete ferroelectric state index | level | integer state index | `tab_unified.go:updateCellInfo` |
| `G=...uS` | \(G_{cell}\) | Cell conductance | µS | material discrete-level mapping | `tab_unified.go:updateCellInfo`, `material.DiscreteLevel` |
| `BL=...V` and `WL=...V` | \(V_{BL},V_{WL}\) | Bitline/wordline bias voltages | V | selected mode/routing values | `tab_unified.go:updateCellInfo`, `DeviceState.GetDACVoltage/GetWLVoltage` |
| `Vcell=...V` | \(V_{cell}\) | Effective cell voltage | V | bias-dependent effective drop | `tab_unified.go:updateCellInfo`, `DeviceState.GetEffectiveCellVoltage` |
| `Icell=...uA` | \(I_{cell}\) | Expected cell current in info line | µA | \(I=G_{cell}\times V_{cell}\) | `tab_unified.go:updateCellInfo` |
| `TIA=...V` | \(V_{TIA}\) | Row TIA output in info line | V | TIA conversion from row current | `tab_unified.go:updateCellInfo`, `DeviceState.GetRowVoltage` |
| `ADC=...` | \(Code\) | Row ADC code in info line | code | ADC quantization of TIA output | `tab_unified.go:updateCellInfo`, `DeviceState.GetRowLevel` |

## Notes
- No `R0` label is used in read-mode sense panel after this update.
- Label semantics prioritize unambiguous symbol + unit/range in parentheses.
