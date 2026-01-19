# Demo 1: P-E Hysteresis Curve Visualizer

**Complexity:** ⭐ Beginner (Graphics only)  
**Timeline:** 1-2 weeks  
**Status:** In Development

## Goal

Interactive visualization of ferroelectric P-E hysteresis curve with:
- Real-time curve plotting as voltage changes
- Material parameter comparison (HfO₂, ZrO₂, HZO superlattice)
- 30 discrete analog state demonstration
- Interactive voltage slider control

## Architecture

```
demo1-hysteresis/
├── cmd/hysteresis/main.go     # Entry point
├── pkg/
│   ├── ferroelectric/         # Physics (CPU-based)
│   │   ├── preisach.go        # ✅ Already implemented
│   │   └── material.go        # ✅ HZO parameters
│   ├── render/                # Vulkan graphics
│   │   └── render.go          # Graphics pipeline
│   └── simulation/            # Time-stepping
│       └── engine.go
└── shaders/
    ├── hysteresis.vert        # Vertex shader
    └── hysteresis.frag        # Fragment shader
```

## Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Physics engine | CPU (Go) | Preisach model already works |
| Vulkan compute | **Not needed** | Simple enough for CPU |
| Rendering | 2D line graphics | P-E curve + state bars |
| Windowing | GLFW | Standard approach |

## Physics Model

Uses the **Mayergoyz Preisach hysteresis model** with:
- Bistable hysterons distributed on the Preisach (α, β) plane
- Gaussian distribution centered at ±Ec with 20% spread
- Hysteresis emergent from hysteron memory (state persists between thresholds)
- 30 discrete levels via linear P discretization: `Level = round((P/Ps + 1) × 14.5)`
- HZO material parameters (Ps=30μC/cm², Ec=1.0MV/cm)

**Note:** Switching time τ is defined but not used in real-time visualization (quasistatic approximation — valid at low frequencies).

## Visualization

```
┌─────────────────────────────────────────────────┐
│  P-E Hysteresis Curve        [HZO Superlattice] │
│                                                 │
│       P                                         │
│       ↑     ╭──────╮                           │
│       │    ╱        ╲                          │
│       │   /          \                         │
│  ─────┼──●────────────●───→ E                  │
│       │   \          /                         │
│       │    ╲        ╱                          │
│       │     ╰──────╯                           │
│                                                 │
│  Voltage: ████████░░░░░░░  [+2.5V]             │
│  State:   14/30  [▓▓▓▓▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░░]  │
└─────────────────────────────────────────────────┘
```

## Implementation Phases

- [x] Phase 1: Core Physics - Preisach model ✅
- [ ] Phase 2: GLFW Window + Vulkan Surface
- [ ] Phase 3: Graphics Pipeline (line rendering)
- [ ] Phase 4: Interactive UI (voltage slider)

## Dependencies

```go
require (
    github.com/bbredesen/go-vk    // Vulkan bindings
    github.com/go-gl/glfw/v3.3/glfw // Windowing
    gonum.org/v1/gonum            // Math library
)
```

## Run

```bash
cd demo1-hysteresis
go run cmd/hysteresis/main.go
# or with headless mode
go run cmd/hysteresis/main.go --headless
```
