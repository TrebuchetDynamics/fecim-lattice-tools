# Demo 2: Crossbar Array MVM Visualizer

**Complexity:** ⭐⭐ Intermediate (Compute + Graphics)  
**Timeline:** 2-3 weeks  
**Status:** Structure Ready

## Goal

Animated visualization of Matrix-Vector Multiplication in a ferroelectric crossbar array:
- Watch currents flow through the crossbar during computation
- Toggle non-idealities (IR drop, sneak paths, device variation)
- Click cells to program conductance values
- Input pulse animation showing voltage propagation

## Architecture

```
demo2-crossbar/
├── cmd/crossbar/main.go       # Entry point
├── pkg/
│   ├── crossbar/              # Array modeling
│   │   ├── array.go           # Crossbar structure
│   │   ├── cell.go            # FeFET/FTJ cell
│   │   └── wire.go            # Wire resistance
│   ├── compute/               # Vulkan compute
│   │   ├── mvm.go             # MVM kernel
│   │   └── nonideal.go        # Non-ideality injection
│   └── layers/                # Neural network layers
└── shaders/
    ├── mvm.comp               # MVM compute shader
    ├── crossbar.vert          # Grid vertex shader
    └── crossbar.frag          # Cell color shader
```

## Key Features

### Matrix-Vector Multiply (MVM)

```
Input Vector (Voltages)
    V₀  V₁  V₂  V₃  V₄  V₅  V₆  V₇
    ↓   ↓   ↓   ↓   ↓   ↓   ↓   ↓
   ┌───┬───┬───┬───┬───┬───┬───┬───┐
I₀ │▓▓▓│░░░│▓░░│░▓░│▓▓▓│░░░│▓░░│░▓░│→ Σ (output)
I₁ │░▓░│▓░░│▓▓▓│░░░│░▓░│▓░░│░░░│▓▓▓│→ Σ
I₂ │▓░░│░░░│░▓░│▓▓▓│░░░│▓▓▓│▓░░│░▓░│→ Σ
   └───┴───┴───┴───┴───┴───┴───┴───┘

Cell color = conductance (▓=high, ░=low)
Animation: Current flows row-by-row
```

### Non-Idealities Modeled

| Effect | Description | Toggle |
|--------|-------------|--------|
| **IR Drop** | Voltage attenuation along wires | [✓] |
| **Sneak Path** | Parasitic current in passive arrays | [✓] |
| **Device Variation** | Cell-to-cell conductance spread | [✓] |
| **ADC Quantization** | Limited output precision | [✓] |

## Vulkan Compute Pipeline

This demo **introduces compute shaders** for parallel MVM:

```glsl
// mvm.comp
layout(local_size_x = 64) in;

layout(set = 0, binding = 0) readonly buffer Weights { float G[]; };
layout(set = 0, binding = 1) readonly buffer Input { float V[]; };
layout(set = 0, binding = 2) writeonly buffer Output { float I[]; };

void main() {
    uint row = gl_GlobalInvocationID.x;
    float sum = 0.0;
    for (uint col = 0; col < numCols; col++) {
        sum += G[row * numCols + col] * V[col];
    }
    I[row] = sum;
}
```

## Implementation Phases

- [ ] Phase 1: Crossbar data structure + MVM logic
- [ ] Phase 2: Vulkan compute pipeline setup
- [ ] Phase 3: 2D grid visualization with cell colors
- [ ] Phase 4: Current flow animation
- [ ] Phase 5: Non-ideality toggles + interactive programming

## Benchmarks (from Literature)

| Architecture | MNIST Accuracy | Source |
|--------------|----------------|--------|
| 24×24 FE Memristor | 98.78% | ScienceDirect 2025 |
| Multi-Level FeFET 28nm | 96.6% | Nature Comms 2023 |
| FTJ Crossbar | 92% | SemiEngineering 2024 |
| Ferroelectric CIM Target | 87% | Dr. Tour presentation |

## Dependencies

```go
require (
    github.com/bbredesen/go-vk
    github.com/go-gl/glfw/v3.3/glfw
    gonum.org/v1/gonum
)
```

## Run

```bash
cd demo2-crossbar
go run cmd/crossbar/main.go
```
