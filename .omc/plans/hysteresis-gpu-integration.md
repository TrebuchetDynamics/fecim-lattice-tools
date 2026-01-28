# Hysteresis Module GPU Integration Plan (v2 - Revised)

## Overview

Integrate module1-hysteresis with the shared Vulkan compute infrastructure (`shared/compute/`) to GPU-accelerate the Preisach hysteresis model.

## Current State Analysis

### What Exists

| Component | Location | Status |
|-----------|----------|--------|
| preisach.comp | module1-hysteresis/shaders/ | EXISTS but NOT integrated (different physics model) |
| MayergoyzPreisach | pkg/ferroelectric/preisach_advanced.go | CPU-only, 100% serial |
| shared/compute | shared/compute/*.go | Production-ready |

### Hysteron Count Formula

For a gridSize of N, the number of hysterons is approximately:
```
numHysterons ≈ N * (N - 1) / 2
```
- gridSize=50 → ~1225 hysterons
- gridSize=100 → ~4950 hysterons
- gridSize=200 → ~19900 hysterons

## Design Decisions

### 1. Precision Handling Strategy

**Decision: Accept float32 precision loss with documented tolerance.**

| Aspect | CPU | GPU | Strategy |
|--------|-----|-----|----------|
| Hysteron α, β | float64 | float32 | Convert, accept ~1e-7 relative error |
| Polarization sum | float64 | float32 | Use Kahan summation, accept 1e-4 tolerance |
| Test tolerance | - | - | Use 1e-4 (not 1e-6) due to float32 |

**Rationale:** Double-precision GLSL requires VK_KHR_shader_float64 extension which isn't universally available. The physics doesn't require 1e-6 precision - ferroelectric measurements typically have ~1% uncertainty.

### 2. Temperature Dependence

**Decision: GPU path ignores temperature (operates at room temperature 300K).**

Temperature-dependent calculations (`temperatureCorrectedEc()`) remain CPU-only. For temperature sweeps, use CPU path. Document this limitation.

### 3. Wake-up Factor

**Decision: GPU path uses fixed wake-up factor of 1.0 (fully woken up).**

Wake-up simulation requires per-cycle state tracking which adds complexity. Users needing wake-up modeling should use CPU path.

## Proposed Architecture

### GPU Hysteron Struct (16 bytes, std430 aligned)

```glsl
struct GPUHysteron {
    float alpha;    // 4 bytes - positive switching threshold (V/m)
    float beta;     // 4 bytes - negative switching threshold (V/m)
    int state;      // 4 bytes - current state (+1 or -1)
    float weight;   // 4 bytes - distribution weight μ(α,β)
};
```

### Go-to-GPU Serialization Code

```go
// GPUHysteron matches shader struct layout (16 bytes, std430)
type GPUHysteron struct {
    Alpha  float32
    Beta   float32
    State  int32
    Weight float32
}

// serializeHysterons converts CPU model to GPU format
func serializeHysterons(m *MayergoyzPreisach) []GPUHysteron {
    gpuHyst := make([]GPUHysteron, len(m.hysterons))
    for i, h := range m.hysterons {
        gpuHyst[i] = GPUHysteron{
            Alpha:  float32(h.Alpha),
            Beta:   float32(h.Beta),
            State:  int32(h.State),
            Weight: float32(m.distribution[i][0]),
        }
    }
    return gpuHyst
}

// deserializeStates updates CPU model from GPU results
func deserializeStates(m *MayergoyzPreisach, gpuHyst []GPUHysteron) {
    for i := range m.hysterons {
        m.hysterons[i].State = int(gpuHyst[i].State)
    }
}
```

## Implementation Tasks

### Phase 1: Shader Development

#### 1. hysteron_update.comp - Parallel State Update (Complete)

```glsl
#version 450

layout(local_size_x = 256) in;

struct GPUHysteron {
    float alpha;
    float beta;
    int state;
    float weight;
};

layout(std140, binding = 0) uniform UpdateParams {
    float E;              // Applied electric field (V/m)
    uint numHysterons;    // Total hysteron count
    float padding[2];     // std140 alignment
};

layout(std430, binding = 1) buffer HysteronBuffer {
    GPUHysteron hysterons[];
};

void main() {
    uint idx = gl_GlobalInvocationID.x;
    if (idx >= numHysterons) return;

    GPUHysteron h = hysterons[idx];

    // Classical Preisach switching: compare E against thresholds
    if (E >= h.alpha) {
        hysterons[idx].state = 1;   // Switch UP
    } else if (E <= h.beta) {
        hysterons[idx].state = -1;  // Switch DOWN
    }
    // Else: memory effect - stay in current state
}
```

#### 2. hysteron_reduce.comp - Parallel Reduction (Complete with Kahan)

```glsl
#version 450

// Two-phase parallel reduction for computing P = Σ(weight_i * state_i)
// Phase 1: Each workgroup computes partial sum into shared memory
// Phase 2: Final workgroup sums all partial results

layout(local_size_x = 256) in;

struct GPUHysteron {
    float alpha;
    float beta;
    int state;
    float weight;
};

layout(std140, binding = 0) uniform ReduceParams {
    uint numHysterons;
    uint numWorkgroups;   // For multi-pass reduction
    float padding[2];
};

layout(std430, binding = 1) buffer HysteronBuffer {
    GPUHysteron hysterons[];
};

layout(std430, binding = 2) buffer PartialSums {
    float partials[];     // One per workgroup
};

layout(std430, binding = 3) buffer OutputBuffer {
    float polarization;   // Final result
};

shared float sharedSums[256];
shared float sharedComp[256];  // Kahan compensation terms

void main() {
    uint localIdx = gl_LocalInvocationID.x;
    uint globalIdx = gl_GlobalInvocationID.x;
    uint groupIdx = gl_WorkGroupID.x;

    // Initialize with Kahan summation
    float sum = 0.0;
    float comp = 0.0;  // Compensation for lost low-order bits

    // Each thread processes multiple elements (grid-stride loop)
    for (uint i = globalIdx; i < numHysterons; i += gl_NumWorkGroups.x * 256) {
        GPUHysteron h = hysterons[i];
        float contribution = h.weight * float(h.state);

        // Kahan summation step
        float y = contribution - comp;
        float t = sum + y;
        comp = (t - sum) - y;
        sum = t;
    }

    sharedSums[localIdx] = sum;
    sharedComp[localIdx] = comp;
    barrier();
    memoryBarrierShared();

    // Tree-based reduction within workgroup
    for (uint stride = 128; stride > 0; stride >>= 1) {
        if (localIdx < stride) {
            // Combine with Kahan
            float y = sharedSums[localIdx + stride] - sharedComp[localIdx];
            float t = sharedSums[localIdx] + y;
            sharedComp[localIdx] = (t - sharedSums[localIdx]) - y;
            sharedSums[localIdx] = t;
        }
        barrier();
        memoryBarrierShared();
    }

    // First thread writes workgroup result
    if (localIdx == 0) {
        partials[groupIdx] = sharedSums[0];
    }

    // Final reduction by first workgroup (for small number of workgroups)
    barrier();
    if (groupIdx == 0 && localIdx == 0) {
        float finalSum = 0.0;
        float finalComp = 0.0;
        for (uint g = 0; g < numWorkgroups; g++) {
            float y = partials[g] - finalComp;
            float t = finalSum + y;
            finalComp = (t - finalSum) - y;
            finalSum = t;
        }
        polarization = finalSum;
    }
}
```

#### 3. distribution.comp - Parallel Gaussian Init

```glsl
#version 450

layout(local_size_x = 256) in;

struct GPUHysteron {
    float alpha;
    float beta;
    int state;
    float weight;
};

layout(std140, binding = 0) uniform DistParams {
    float alphaM;     // Mean of alpha distribution
    float betaM;      // Mean of beta distribution
    float sigmaA;     // Std dev of alpha
    float sigmaB;     // Std dev of beta
    float rho;        // Correlation coefficient
    float Ps;         // Saturation polarization (for normalization)
    uint numHysterons;
    float padding;
};

layout(std430, binding = 1) buffer HysteronBuffer {
    GPUHysteron hysterons[];
};

layout(std430, binding = 2) buffer TotalWeight {
    float totalWeight;  // For normalization (atomic add)
};

void main() {
    uint idx = gl_GlobalInvocationID.x;
    if (idx >= numHysterons) return;

    GPUHysteron h = hysterons[idx];

    // Bivariate Gaussian: μ(α,β) = exp(-(da² - 2ρ·da·db + db²) / (2(1-ρ²)))
    float da = (h.alpha - alphaM) / sigmaA;
    float db = (h.beta - betaM) / sigmaB;

    float denom = 2.0 * (1.0 - rho * rho);
    float exponent = -(da * da - 2.0 * rho * da * db + db * db) / denom;
    float weight = exp(exponent);

    hysterons[idx].weight = weight;

    // Atomic add for total (for later normalization pass)
    atomicAdd(totalWeight, weight);
}
```

### Phase 2: Go Integration

#### gpu_preisach.go Structure

```go
// GPUPreisach provides GPU-accelerated Preisach hysteresis
type GPUPreisach struct {
    ctx            *compute.VulkanContext
    updatePipeline *compute.ComputePipeline
    reducePipeline *compute.ComputePipeline
    distPipeline   *compute.ComputePipeline

    hystBuffer     *compute.GPUBuffer  // GPUHysteron array
    partialsBuffer *compute.GPUBuffer  // Workgroup partial sums
    outputBuffer   *compute.GPUBuffer  // Final polarization

    numHysterons   int
    numWorkgroups  int
    cpuModel       *MayergoyzPreisach  // For sync and fallback
}
```

### Phase 3: Update preisach_advanced.go

Add to `MayergoyzPreisach`:
```go
type MayergoyzPreisach struct {
    // ... existing fields ...

    UseGPU        bool          // Enable GPU acceleration
    gpuAccel      *GPUPreisach  // Lazy-initialized
    gpuInitDone   bool          // Tracks initialization attempt
}

func (m *MayergoyzPreisach) Update(E float64) float64 {
    m.initGPU()  // Lazy init

    if m.gpuAccel != nil && m.gpuAccel.IsAvailable() {
        return m.gpuAccel.Update(E)
    }
    return m.updateCPU(E)  // Fallback
}
```

## Test Vectors

### Test Case 1: Basic Switching
```
Input:
  - material: DefaultHZOMaterial() (Ec=1.0 MV/cm, Ps=30 µC/cm²)
  - gridSize: 10 (45 hysterons)
  - E sequence: [0, +2*Ec, 0, -2*Ec, 0]

Expected CPU Output:
  - P[0] ≈ -15 µC/cm² (initial negative saturation)
  - P[1] ≈ +30 µC/cm² (positive saturation after +2Ec)
  - P[2] ≈ +15 µC/cm² (remanent at E=0)
  - P[3] ≈ -30 µC/cm² (negative saturation after -2Ec)
  - P[4] ≈ -15 µC/cm² (remanent at E=0)

GPU Tolerance: ±1e-4 (0.01%) relative to CPU values
```

### Test Case 2: Hysteresis Loop Closure
```
Input:
  - gridSize: 50 (~1225 hysterons)
  - Emax: 1.5 * Ec
  - points: 100

Expected:
  - Loop closes (P_start ≈ P_end within 1%)
  - Symmetric about origin (|P(E)| ≈ |P(-E)|)
  - Pr (remanent) between 0.3*Ps and 0.9*Ps
```

### Speedup Measurement Methodology
```go
func BenchmarkUpdate(b *testing.B) {
    model := NewMayergoyzPreisach(DefaultHZOMaterial(), 100)

    b.Run("CPU", func(b *testing.B) {
        model.UseGPU = false
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            model.Update(1.0e6)  // 1 MV/cm
        }
    })

    b.Run("GPU", func(b *testing.B) {
        model.UseGPU = true
        model.Update(0)  // Warm up, initialize GPU
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            model.Update(1.0e6)
        }
    })
}
```

**Speedup = CPU_time / GPU_time** (measured after GPU warm-up, includes buffer transfer)

## File Changes Summary

### New Files

| File | Lines (est) | Purpose |
|------|-------------|---------|
| module1-hysteresis/pkg/ferroelectric/gpu_preisach.go | 350 | GPU wrapper with serialization |
| module1-hysteresis/shaders/hysteron_update.comp | 40 | State update shader |
| module1-hysteresis/shaders/hysteron_reduce.comp | 90 | Kahan reduction shader |
| module1-hysteresis/shaders/distribution.comp | 50 | Gaussian init shader |
| module1-hysteresis/pkg/ferroelectric/gpu_preisach_test.go | 150 | GPU-specific tests |

### Modified Files

| File | Changes |
|------|---------|
| module1-hysteresis/pkg/ferroelectric/preisach_advanced.go | Add UseGPU field, lazy GPU init, delegation |
| module1-hysteresis/shaders/compile.sh | Add new shaders |

## Acceptance Criteria

1. [ ] All 3 shaders compile with glslc
2. [ ] hysteron_update.comp correctly switches states
3. [ ] hysteron_reduce.comp produces polarization matching CPU within 1e-4
4. [ ] GPUPreisach.Update() matches CPU within 1e-4 relative tolerance
5. [ ] Test Case 1 (Basic Switching) passes on both CPU and GPU
6. [ ] Test Case 2 (Loop Closure) produces valid hysteresis loop
7. [ ] GPU fallback works when Vulkan unavailable
8. [ ] Existing tests pass (`go test ./module1-hysteresis/...`)
9. [ ] 5x+ speedup for gridSize >= 50 (measured per methodology above)

## Limitations (Documented)

1. **Temperature:** GPU path operates at 300K only. Use CPU for temperature sweeps.
2. **Wake-up:** GPU assumes fully woken material. Use CPU for wake-up simulation.
3. **Precision:** GPU uses float32, results differ from CPU float64 by up to 1e-4.
4. **Max gridSize:** Limited by GPU memory. Recommend gridSize ≤ 500 (~125K hysterons).

## Dependencies

- `shared/compute` package (already implemented)
- Vulkan SDK with glslc
- No new Go dependencies
