# Module 3: MNIST - Physics

## Prerequisites

- Vectors and matrices
- Basic neural network concepts
- Probability and normalization

## Core Model

- A feed-forward network computes logits from inputs.
- ReLU introduces nonlinearity; softmax produces probabilities.
- The CIM path applies quantization and noise to weights and activations.

## Key Equations (Simplified)

```
Hidden = ReLU(W1 * x + b1)
Logits = W2 * Hidden + b2
Probs = softmax(Logits)
QuantizedValue = round(value * (L-1)) / (L-1)
```

## Parameters And Units

| Symbol | Meaning | Units |
|---|---|---|
| x | Input vector (784) | unitless |
| W | Weight matrix | unitless |
| b | Bias vector | unitless |
| L | Quantization levels | levels |
| sigma | Noise coefficient | unitless |

## Assumptions And Limits

- The architecture is a small MLP for visualization speed.
- Quantization is uniform with optional per-layer levels.
- Noise is modeled as additive perturbations.

## Where It Lives In Code

- `module3-mnist/pkg/core/network_inference.go`
- `module3-mnist/pkg/core/quantize.go`
- `module3-mnist/pkg/gui/dualmode.go`

## Sources

- `docs/development/scriptReference.md#demo-3-mnist-module3-mnist`
- `module3-mnist/pkg/core/network_inference.go`
