# Module 4: Circuits - Open-Source Tools

## When To Use External Tools

- Validating circuit blocks with SPICE-level detail.
- Exploring alternative ADC/DAC architectures.
- Building layout-aware timing and power models.

## Recommended Tools (With Rationale)

- ngspice or Xyce for circuit simulation.
- Qucs-S for schematic-level exploration.
- KiCad for block-level schematics and documentation.

## Integration Notes

- Export SPICE netlists from `module6-eda/pkg/export/spice.go`.
- Peripheral parameters live in `module4-circuits/pkg/peripherals/`.
- Use the circuits GUI to sanity check behavior before SPICE runs.
