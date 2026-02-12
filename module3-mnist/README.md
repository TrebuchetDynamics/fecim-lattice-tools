# Module 3: MNIST

Hardware-aware MNIST inference and training for FeCIM crossbar arrays. Compares floating-point baselines against quantized CIM inference with configurable noise, conductance levels, and non-idealities.

## Overview

Module 3 bridges neural network workloads with the ferroelectric crossbar simulator. It loads MNIST digit images, runs inference through multi-layer networks mapped onto crossbar arrays, and quantifies accuracy degradation from quantization, device noise, and analog non-idealities. Supports both full-precision (software) and CIM (hardware-aware) inference modes side-by-side.

## Package Structure

### `pkg/core/` — Inference and Quantization Engine

- **network.go** — Multi-layer neural network definition and forward pass
- **network_inference.go** — Inference pipeline (single image and batch)
- **network_quantization.go** — Weight quantization to discrete conductance levels
- **network_config.go** — Network topology configuration
- **network_notifications.go** — Progress/status callbacks
- **network_gpu.go** — GPU-accelerated inference path
- **quantize.go** — Quantization utilities (uniform, per-layer, symmetric)
- **cim_physics.go** — CIM-specific physics: noise injection, IR drop, drift models
- **energy_model.go** — Energy-per-inference estimation (read/compute/ADC)
- **dualmode_metrics.go** — Side-by-side accuracy comparison metrics
- **constants.go** — Physical and model constants
- **interfaces.go** — Shared interfaces for pluggable backends

### `pkg/mnist/` — Dataset

- **loader.go** — MNIST IDX file parser and image loading

### `pkg/training/` — Training Utilities

- **network.go** — Backpropagation training loop
- **single_layer.go** — Single-layer perceptron trainer
- **trainer_foundation.go** — Foundation trainer with learning rate scheduling
- **seed.go** — Reproducible random seeding

### `pkg/gui/` — Fyne GUI

- **embedded.go** — Embeddable app for unified launcher
- **dualmode.go** — Dual-mode (FP vs CIM) comparison view
- **dualmode_controls.go** — Parameter controls for dual-mode
- **dualmode_inference.go** — Inference execution and display
- **dualmode_weights.go** — Weight visualization (FP vs quantized)
- **dualmode_demo.go** — Demo/walkthrough mode
- **canvas.go** — Digit drawing canvas for live inference
- **energy_widget.go** — Energy breakdown visualization
- **quantization_widget.go** — Quantization sweep controls
- **metrics.go** — Accuracy/loss display
- **activations.go** — Layer activation visualization
- **comparison_card.go** — Architecture comparison cards
- **weight_comparison.go** — Weight distribution comparison
- **network_controller.go** — GUI↔network bridge
- **preprocess.go** — Image preprocessing (center, normalize)
- **export.go** — Result export utilities

### `cmd/` — Entry Points

- **mnist/main.go** — CLI inference runner
- **mnist-gui/main.go** — Standalone GUI launcher
- **train-network/main.go** — Full network training
- **train-single-layer/main.go** — Single-layer training
- **train-ptq/main.go** — Post-training quantization workflow

## Key Types and Functions

| Type / Function | Package | Description |
|---|---|---|
| `Network` | `pkg/core` | Multi-layer network with forward pass and quantization |
| `CIMPhysics` | `pkg/core` | Noise/drift/IR-drop models for hardware-aware inference |
| `EnergyModel` | `pkg/core` | Per-inference energy estimation |
| `DualModeMetrics` | `pkg/core` | FP vs CIM accuracy comparison |
| `Quantize` | `pkg/core` | Weight quantization to N discrete levels |
| `Loader` | `pkg/mnist` | MNIST dataset parser |
| `Trainer` | `pkg/training` | Backpropagation training with LR scheduling |

## Testing

```bash
# Run all module 3 tests
go test ./module3-mnist/...

# With race detector
go test -race ./module3-mnist/...

# Specific packages
go test -v ./module3-mnist/pkg/core/...
go test -v ./module3-mnist/pkg/gui/...
```

Key test suites:
- `pkg/core/` — Quantization accuracy, energy model, CIM physics
- `pkg/gui/` — Widget rendering, preprocessing, energy widget, network controller notifications

## Physics Context

**Quantization:** Weights are mapped from floating-point to N discrete conductance levels (default 30), mimicking the finite states of a ferroelectric memory cell. Quantization error grows as levels decrease.

**CIM Inference:** Matrix-vector multiplication is performed in the analog domain: input voltages × stored conductances = output currents. Non-idealities (noise, IR drop, sneak paths, drift) degrade the analog computation.

**Energy Model:** Energy per inference accounts for DAC conversions (input encoding), crossbar read currents, TIA amplification, and ADC quantization. CIM inference is typically orders of magnitude more energy-efficient than digital MACs for the same operation.

**Key metrics:** Top-1 accuracy (%), accuracy delta (FP − CIM), energy per inference (pJ), throughput (inferences/s).

## Related Documentation

- `docs/documentation/module3-mnist/` — ELI5, features, physics, open-source tools
- `docs/neural-network/` — Architecture, API reference, development notes
- `docs/CLI.md` — Command-line interface reference
