# Configuration Files

This directory contains all FeCIM simulation configuration files.

## Overview

| File | Purpose | Key Parameters |
|------|---------|----------------|
| `materials.yaml` | Ferroelectric material presets (HZO, PZT, BTO, AlScN, etc.) | Ec, Ps, Pr, thickness, Gmin/Gmax, TargetLevels |
| `constants.yaml` | Physical constants | k_B, epsilon_0, q_e |
| `benchmarks.yaml` | Performance benchmark targets | timeout, iterations |
| `calibration.yaml` | Calibration settings | method, tolerance |
| `crossbar.yaml` | Crossbar array geometry | rows, columns, architecture |
| `energy.yaml` | Energy model parameters | fJ per MAC, peripheral estimates |
| `mnist.yaml` | MNIST inference settings | batch_size, learning_rate |
| `preisach.yaml` | Preisach model parameters | alpha, beta, gamma, Ec |
| `simulation.yaml` | Simulation defaults | dt, max_steps |
| `timing.yaml` | Timing constraints | read_latency, write_latency |
| `training.yaml` | Training hyperparameters | epochs, optimizer |

## Subdirectories

- `physics/` — Physics-specific parameter overrides

## Usage

Load via `config.Load()`:

```go
cfg, err := config.Load("materials.yaml")
mat := cfg.Materials["fecim_hzo"]
```

All files are YAML 1.1 compliant.
