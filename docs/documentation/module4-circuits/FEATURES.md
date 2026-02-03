# Module 4: Circuits - Features

## What This Module Does

- Models DAC, ADC, TIA, and charge pump behavior.
- Estimates timing and power for peripheral operations.
- Visualizes signal flow and circuit-level effects.

## Primary Components

- `module4-circuits/pkg/peripherals/dac.go`
- `module4-circuits/pkg/peripherals/adc.go`
- `module4-circuits/pkg/peripherals/tia.go`
- `module4-circuits/pkg/peripherals/analysis.go`

## Key Workflows

- Convert digital inputs to analog voltages for array drive.
- Convert array currents into voltages and digital codes.
- Estimate timing and power breakdown for conversions.

## Extension Points

- Add new ADC/DAC architectures or nonlinearity models.
- Extend power analysis with additional blocks.
- Connect to exported SPICE netlists from module 6.

## Known Limitations

- Behavior is analytic, not SPICE-accurate.
- Parameter defaults are for teaching, not silicon tuning.
- Timing is approximate and does not include full routing effects.
