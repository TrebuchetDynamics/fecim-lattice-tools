# Module 4: Circuits - Features

## What This Module Does

- Peripheral circuits (DAC, ADC, TIA, charge pump)

## Primary Components

- `module4-circuits/pkg/peripherals/dac.go`
- `module4-circuits/pkg/peripherals/adc.go`
- `module4-circuits/pkg/peripherals/tia.go`

## Key Workflows

- Configure DAC/ADC parameters -> compute timing/power tradeoffs.

## Extension Points

- Peripheral models with timing and power analysis helpers.
- Signal flow visualization of data movement.

## Known Limitations

- No transistor-level verification in this module.
- Simplified noise and nonlinearity models.

