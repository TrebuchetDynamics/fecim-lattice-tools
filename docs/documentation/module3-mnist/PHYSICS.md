# Module 3: MNIST - Physics

## Prerequisites

- Linear algebra basics
- Activation functions
- Quantization

## Core Model

- Forward pass is a sequence of matrix multiplies and nonlinearities.
- CIM path quantizes weights/activations to discrete levels and injects noise.

## Key Equations (Simplified)

```
a_{l+1} = f(W_l a_l + b_l)
Quantize(x) -> nearest discrete level
```

## Parameters and Units

| Symbol | Meaning | Units |
|---|---|---|
| NumLevels | Quantization levels | count |
| NoiseLevel | Additive noise std | unitless |
| ADCBits | ADC resolution | bits |

## Assumptions and Limits

- Feedforward network only (no training loop in the GUI).
- Noise modeled as simple Gaussian perturbation.

## Where It Lives in Code

- `module3-mnist/pkg/core/network.go`
- `module3-mnist/pkg/core/quantize.go`
- `module3-mnist/pkg/gui/dualmode.go`

## Sources

- `docs/development/scriptReference.md#demo-3-mnist-module3-mnist`
- `docs/ELI5.md`

