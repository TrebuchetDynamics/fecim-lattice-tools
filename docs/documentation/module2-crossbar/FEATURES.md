# Module 2: Crossbar - Features

## What This Module Does

- Array physics, MVM, and non-idealities

## Primary Components

- `module2-crossbar/pkg/crossbar/array.go`
- `module2-crossbar/pkg/crossbar/nonidealities.go`
- `module2-crossbar/pkg/gui/tabs/irdrop_tab.go`

## Key Workflows

- Program weights -> run MVM -> compare ideal vs non-ideal.
- Sweep wire parameters to study IR drop sensitivity.

## Extension Points

- Interactive heatmap of conductance values.
- Tabbed analyses for ideal MVM, IR drop, sneak paths, drift.
- DAC/ADC quantization hooks for system-level realism.

## Known Limitations

- No full SPICE-level transient simulation.
- Non-idealities are simplified and not device-calibrated by default.

