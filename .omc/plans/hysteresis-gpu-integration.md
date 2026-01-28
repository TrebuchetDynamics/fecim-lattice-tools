# Hysteresis Module GPU Integration Plan

## Overview

Integrate module1-hysteresis with the shared Vulkan compute infrastructure (`shared/compute/`) to GPU-accelerate the Preisach hysteresis model. The existing `preisach.comp` shader will be enhanced and connected to the shared compute pipeline.

## Current State Analysis

### What Exists

| Component | Location | Status |
|-----------|----------|--------|
| preisach.comp | module1-hysteresis/shaders/ | EXISTS but NOT integrated |
| preisach.comp.spv | module1-hysteresis/shaders/ | Compiled, unused |
| MayergoyzPreisach | pkg/ferroelectric/preisach_advanced.go | CPU-only, 100% serial |
| shared/compute | shared/compute/*.go | Production-ready |

### Key Finding: Orphaned Shader

The `preisach.comp` shader (88 lines) implements basic tanh-based Preisach switching for multiple cells in parallel, but:
- Has NO Go code calling it
- Physics model doesn't match `MayergoyzPreisach` (hysteron-based)
- Uses simple cell struct, not hysteron grid

### CPU Bottlenecks in preisach_advanced.go

| Function | Complexity | Typical Size | GPU Benefit |
|----------|------------|--------------|-------------|
| `Update()` | O(N) | 5000 hysterons | HIGH - parallel state updates |
| `initializeDistribution()` | O(N) | 5000 exp() calls | HIGH - parallel Gaussian |
| `GetHysteresisLoop()` | O(N × points) | 1M operations | HIGH - batch processing |
| `SimulateDomainSwitching()` | O(steps × N) | 5M operations | MEDIUM - timestepping |

## Proposed Architecture

### New Files

```
module1-hysteresis/
├── pkg/ferroelectric/
│   └── gpu_preisach.go      # GPUPreisach wrapper (NEW)
├── shaders/
│   ├── preisach.comp        # Basic cell shader (EXISTS - keep for simple mode)
│   ├── hysteron_update.comp # Hysteron state update (NEW)
│   ├── hysteron_reduce.comp # Polarization reduction (NEW)
│   └── distribution.comp    # Gaussian distribution init (NEW)
```

### Core Integration

```go
// module1-hysteresis/pkg/ferroelectric/gpu_preisach.go

import "fecim-lattice-tools/shared/compute"

// GPUPreisach provides GPU-accelerated Preisach hysteresis model
type GPUPreisach struct {
    ctx            *compute.VulkanContext
    updatePipeline *compute.ComputePipeline
    reducePipeline *compute.ComputePipeline

    // GPU buffers
    hystBuffer     *compute.GPUBuffer  // Hysteron states (alpha, beta, state)
    distBuffer     *compute.GPUBuffer  // Distribution weights
    outputBuffer   *compute.GPUBuffer  // Reduction output

    numHysterons   int
    cpuFallback    *MayergoyzPreisach  // Fallback when GPU unavailable
}

func NewGPUPreisach(material *HZOMaterial, gridSize int) (*GPUPreisach, error)
func (g *GPUPreisach) Update(E float64) float64
func (g *GPUPreisach) GetHysteresisLoop(Emax float64, points int) ([]float64, []float64)
func (g *GPUPreisach) Destroy()
```

## Implementation Tasks

### Phase 1: Shader Development

1. **hysteron_update.comp** - Parallel hysteron state update
   ```glsl
   #version 450
   layout(local_size_x = 256) in;

   struct Hysteron {
       float alpha;    // Positive switching threshold
       float beta;     // Negative switching threshold
       int state;      // +1 or -1
       float weight;   // Distribution weight
   };

   layout(std140, binding = 0) uniform Params {
       float E;           // Applied electric field
       int numHysterons;
       float padding[2];
   };

   layout(std430, binding = 1) buffer HysteronBuffer {
       Hysteron hysterons[];
   };

   void main() {
       uint idx = gl_GlobalInvocationID.x;
       if (idx >= numHysterons) return;

       Hysteron h = hysterons[idx];

       // Classical Preisach switching logic
       if (E >= h.alpha) {
           hysterons[idx].state = 1;   // Switch UP
       } else if (E <= h.beta) {
           hysterons[idx].state = -1;  // Switch DOWN
       }
       // Else: stay in current state (memory effect)
   }
   ```

2. **hysteron_reduce.comp** - Parallel reduction for polarization
   ```glsl
   #version 450
   layout(local_size_x = 256) in;

   // Computes: P = Σ(weight_i * state_i)
   // Uses workgroup shared memory for efficient reduction

   shared float partialSums[256];

   layout(std430, binding = 1) buffer HysteronBuffer { ... };
   layout(std430, binding = 2) buffer OutputBuffer {
       float polarization;
   };
   ```

3. **distribution.comp** - Parallel Gaussian distribution initialization
   ```glsl
   #version 450
   layout(local_size_x = 256) in;

   // Computes bivariate Gaussian for each hysteron:
   // weight = exp(-(da² - 2ρ*da*db + db²) / (2(1-ρ²)))
   ```

### Phase 2: Go Integration

1. **gpu_preisach.go** - Main GPU wrapper
   - Initialize VulkanContext from shared/compute
   - Create compute pipelines for each shader
   - Manage hysteron buffer (upload once, reuse)
   - Implement Update() with GPU dispatch
   - CPU fallback when GPU unavailable

2. **Update preisach_advanced.go**
   - Add `UseGPU bool` option to model
   - Lazy-initialize GPUPreisach on first Update()
   - Delegate to GPU or CPU based on availability

### Phase 3: Move Shader to Shared (Optional)

Consider moving `preisach.comp` to `shared/compute/shaders/` if other modules could use ferroelectric cell simulation.

## Shader Binding Layout

### hysteron_update.comp
| Binding | Type | Content |
|---------|------|---------|
| 0 | Uniform | Params (E, numHysterons) |
| 1 | Storage | HysteronBuffer (read/write) |

### hysteron_reduce.comp
| Binding | Type | Content |
|---------|------|---------|
| 0 | Uniform | Params (numHysterons) |
| 1 | Storage | HysteronBuffer (read) |
| 2 | Storage | OutputBuffer (write - single float) |

### distribution.comp
| Binding | Type | Content |
|---------|------|---------|
| 0 | Uniform | DistParams (means, sigmas, rho, Ec) |
| 1 | Storage | HysteronBuffer (read alpha/beta, write weight) |

## File Changes Summary

### New Files

| File | Lines (est) | Purpose |
|------|-------------|---------|
| module1-hysteresis/pkg/ferroelectric/gpu_preisach.go | 300 | GPU wrapper |
| module1-hysteresis/shaders/hysteron_update.comp | 50 | State update shader |
| module1-hysteresis/shaders/hysteron_reduce.comp | 80 | Reduction shader |
| module1-hysteresis/shaders/distribution.comp | 60 | Gaussian init shader |

### Modified Files

| File | Changes |
|------|---------|
| module1-hysteresis/pkg/ferroelectric/preisach_advanced.go | Add UseGPU option, GPU delegation |
| module1-hysteresis/shaders/compile.sh | Add new shaders to compilation |

## Verification Steps

1. **Unit Tests**
   - `go test ./module1-hysteresis/pkg/ferroelectric/...`
   - Test GPU vs CPU results match within tolerance
   - Test fallback when Vulkan unavailable

2. **Integration Tests**
   - GetHysteresisLoop produces valid P-E curves
   - GPU and CPU paths produce identical curves
   - No memory leaks (run with -race)

3. **Performance Benchmarks**
   - `BenchmarkUpdate_CPU` vs `BenchmarkUpdate_GPU`
   - Target: 10-100x speedup for gridSize=100 (5000 hysterons)

4. **Build Verification**
   - `go build ./module1-hysteresis/...`
   - All shaders compile with glslc

## Acceptance Criteria

1. [ ] hysteron_update.comp compiles and executes
2. [ ] hysteron_reduce.comp produces correct polarization sum
3. [ ] GPUPreisach.Update() matches CPU within 1e-6 tolerance
4. [ ] GPU fallback works when Vulkan unavailable
5. [ ] GetHysteresisLoop produces valid P-E curves
6. [ ] Existing tests pass
7. [ ] 10x+ speedup demonstrated for gridSize >= 50

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Reduction precision | Use Kahan summation in shader |
| Floating point mismatch | Accept 1e-6 tolerance, document |
| GPU memory limit | Limit gridSize, warn user |
| No GPU available | Transparent CPU fallback |

## Dependencies

- `shared/compute` package (already implemented)
- Vulkan SDK with glslc (for shader compilation)
- No new Go dependencies
