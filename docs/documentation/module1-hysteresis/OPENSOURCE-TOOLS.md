# Module 1: Hysteresis - Open-Source Tools

## When To Use External Tools

- Calibrating model parameters against published measurements.
- Running higher-fidelity hysteresis or domain simulations.
- Producing publication-quality plots for reports.

## Recommended Tools (With Rationale)

- NumPy and SciPy for data fitting and parameter sweeps.
- Matplotlib for high-quality P-E loop plots.
- Jupyter for reproducible, shareable experiments.
- ngspice or Xyce for circuit-level validation of polarization models.

## Integration Notes

- Source parameter ranges from `docs/research-papers/by-topic/01-ferroelectric-materials/`.
- Use `module1-hysteresis/pkg/ferroelectric/material.go` as the local parameter baseline.
- If you export data, keep units explicit and consistent with PHYSICS tables.
