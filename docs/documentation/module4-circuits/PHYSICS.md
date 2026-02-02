# Module 4: Circuits - Physics

## Prerequisites

- Ohm's law
- Sampling/quantization
- Basic op-amp intuition

## Core Model

- DAC: digital code -> analog voltage steps.
- ADC: analog voltage -> digital code with quantization error.
- TIA: V_out = I_in * R (first-order model).

## Key Equations (Simplified)

```
V_out = I_in * R
Quantization step = V_ref / (2^N - 1)
```

## Parameters and Units

| Symbol | Meaning | Units |
|---|---|---|
| N | Converter resolution | bits |
| V_ref | Reference voltage | Volts |
| R | TIA resistance | Ohms |

## Assumptions and Limits

- Linear converter models without layout parasitics.
- Noise treated as additive for analysis.

## Where It Lives in Code

- `module4-circuits/pkg/peripherals/dac.go`
- `module4-circuits/pkg/peripherals/adc.go`
- `module4-circuits/pkg/peripherals/tia.go`

## Sources

- `docs/development/scriptReference.md#demo-4-circuits-module4-circuits`

