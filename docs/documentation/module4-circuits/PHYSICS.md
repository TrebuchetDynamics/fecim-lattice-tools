# Module 4: Circuits - Physics

## Prerequisites

- Ohm's law
- Basic sampling and quantization
- Simple RC concepts

## Core Model

- DAC maps digital codes to analog voltages.
- TIA maps analog currents to voltages.
- ADC maps analog voltages back to digital codes.

## Key Equations (Simplified)

```
LSB = Vref / (2^N)
V_out = I_in * R_tia
QuantizationError ≈ ±0.5 LSB
```

## Parameters And Units

| Symbol | Meaning | Units |
|---|---|---|
| Vref | Reference voltage | Volts |
| N | Converter resolution | bits |
| I_in | Input current | Amps |
| R_tia | TIA resistance | Ohms |

## Assumptions And Limits

- Idealized converter behavior by default.
- No full transistor-level modeling.
- Nonlinearity and noise are simplified.

## Where It Lives In Code

- `module4-circuits/pkg/peripherals/dac.go`
- `module4-circuits/pkg/peripherals/adc.go`
- `module4-circuits/pkg/peripherals/tia.go`
- `module4-circuits/pkg/gui/app.go`

## Sources

- `docs/development/scriptReference.md#demo-4-circuits-module4-circuits`
- `module4-circuits/pkg/peripherals/analysis.go`
