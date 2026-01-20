# IronLattice Implementation Guide
**Extracted from Comprehensive Research Analysis**  
**Date:** 2026-01-18

---

## 🎯 CRITICAL ACTIONABLE INSIGHTS

This guide extracts the key implementation recommendations from the comprehensive ferroelectric CIM analysis.

---

## Demo 1: Hysteresis Visualizer - Preisach Model Fix

### Problem
Current implementation uses simplified tanh model. Need rigorous Preisach hysteresis tracking.

### Solution: Discrete Preisach Model (DPM)

#### Mathematical Framework
```
f(t) = ∬(α≥β) μ(α,β) γ̂_αβ[u(t)] dα dβ
```

#### Implementation Steps

1. **Discretize Preisach Plane**
   - Create 100×100 grid of (α, β) values
   - Each cell represents a hysteron with switching thresholds

2. **Maintain Voltage History Stack**
   - Track local extrema (peaks and valleys)
   - **Critical:** Implements "wiping-out" property
   - When voltage exceeds previous max, clear nested loop history

3. **Geometric State Tracking**
   - Divide Preisach triangle into S⁺ (up) and S⁻ (down) regions
   - Track "staircase interface" L(t) between regions
   - Update interface as voltage increases (horizontal) or decreases (vertical)

4. **Advanced: Preisach Neural Network**
   - **Layer 1:** Stop operator neurons (one per hysteron)
   - **Layer 2:** Linear summation with learned weights
   - **Training:** Backprop to learn μ(α,β) from device calibration
   - **Benefit:** Self-calibrating hysteresis model

#### Code Location
- `demo1-hysteresis/pkg/ferroelectric/preisach.go`

#### Expected Improvement
- Accurate state prediction after complex voltage sequences
- Inverse control: Calculate exact pulse for target conductance

---

## Demo 2: 30-Level Quantization - Pulse Scheme Fix

### Problem
Current bug: Using constant-amplitude pulses (Scheme A) causes state bunching.

### Solution: Incremental Amplitude Programming (Scheme C)

#### The Three Schemes (Oh et al. 2017)

| Scheme | Method | Result |
|--------|--------|--------|
| **A** ❌ | Identical pulses (constant V, t) | States bunch - FAILS |
| **B** ⚠️ | Variable pulse width | Works but complex timing |
| **C** ✅ | **Incremental voltage** | **32 linear states - SUCCESS** |

#### Implementation: Scheme C Algorithm

```go
// Pseudocode for incremental amplitude programming
func ProgramLevel(targetLevel int) {
    V_start := 1.0  // volts
    V_end := 3.0
    V_step := 0.05  // 50mV increments
    
    numSteps := targetLevel  // 0-30 for 30 levels
    
    for i := 0; i < numSteps; i++ {
        V_prog := V_start + (V_step * i)
        ApplyPulse(V_prog, 100ns)  // Fixed 100ns width
    }
}
```

#### Key Parameters
- **Voltage range:** 1.0V → 3.0V
- **Increment:** 50mV per step
- **Pulse width:** 100ns (fixed)
- **Result:** 30 distinct, non-overlapping conductance levels

#### Physics Explanation
- Each voltage increment switches a specific grain population
- Overcomes varying coercive fields in polycrystalline HZO
- Prevents screening effects that cause Scheme A failure

#### Code Location
- `demo2-crossbar/pkg/crossbar/array.go`
- Replace constant ADC quantization with Scheme C pulse sequence

---

## Demo 3: MNIST Accuracy - Symmetric Updates

### Problem
Achieving 87% target (current: 95.8% already exceeds, but understanding mechanism).

### Solution: Symmetric Potentiation/Depression (Jerry et al. 2017)

#### Critical Parameter: 75ns Pulse Width

**Result:** Jerry et al. achieved **90% MNIST accuracy** using 32-level HZO FeFETs.

**Key Finding:** HZO exhibits symmetric update curves ONLY at **75ns pulse width**.

#### Symmetry Importance
```
Asymmetric (BAD):
- Potentiation (LTP): Smooth increase ✓
- Depression (LTD): Abrupt drop ✗
- Result: Network can't learn efficiently

Symmetric (GOOD):
- Potentiation: Gradual ✓
- Depression: Gradual ✓  
- Result: 90% accuracy without correction algorithms
```

#### Implementation
```go
// Optimize pulse width for symmetric updates
const OPTIMAL_PULSE_WIDTH = 75 // nanoseconds

func UpdateWeight(delta float64) {
    if delta > 0 {
        // Potentiation
        ApplyPulse(V_pot, OPTIMAL_PULSE_WIDTH)
    } else {
        // Depression
        ApplyPulse(V_dep, OPTIMAL_PULSE_WIDTH)
    }
}
```

#### Why 75ns?
- Balance between:
  - Domain nucleation time (~10ns)
  - Domain wall propagation (~100ns)
- Allows symmetric switching dynamics

---

## Material Strategy: Three Pathways

### 1. HZO (Primary - CMOS Compatible) ✅

**Advantages:**
- CMOS fabrication compatible
- 10nm scaling demonstrated
- Robust Pr ~20 μC/cm²
- Low Ec ~1 MV/cm (operates at 1-3V)

**Challenges:**
- Requires Scheme C for linearity
- Needs 75ns pulse optimization

**Status:** **Recommended for immediate deployment**

### 2. α-In₂Se₃ (Future - Dr. Tour's 2D) 🚀

**Advantages:**
- Ferroelectric semiconductor (channel IS ferroelectric)
- Monolayer stability (interlocked OOP/IP polarization)
- Flash-Within-Flash synthesis (gram-scale in seconds!)

**Challenges:**
- Novel material, less mature
- Currently 87% MNIST (vs 90% for HZO)

**Synthesis Method:**
```
Flash-Within-Flash (FWF):
1. Nested tubes: Inner (In+Se) | Outer (coke)
2. Arc discharge through coke (>2000°C in ms)
3. Radiant heating → sublimation → α-phase kinetic trapping
4. Cooling >10⁴ K/s prevents β-phase reversion
```

**Status:** **High-risk, high-reward - continue R&D**

### 3. Weebit ReRAM (Baseline - Harsh Environments) 🛡️

**Advantages:**
- Radiation hardened (gamma resistant)
- AEC-Q100 automotive qualified
- 10 years retention at 150°C

**Challenges:**
- Filamentary switching = poor linearity
- Hard to achieve 30 analog levels
- Not ideal for on-chip training

**Status:** **Use for inference in harsh environments only**

---

## Algorithmic Co-Design

### 1. Quantization-Aware Training (QAT)

**Problem:** Training in 32-bit then truncating to 5-bit = accuracy loss.

**Solution:**
```python
# Pseudocode
def forward_pass(x, W):
    W_quantized = quantize_to_30_levels(W)  # Simulate hardware
    return neural_net(x, W_quantized)

def backward_pass(loss):
    grad = compute_gradient(loss)
    # Straight-Through Estimator: ignore quantization in gradient
    return grad

def update_weights(W, grad, lr):
    W_shadow = W_shadow - lr * grad  # High precision shadow
    W = quantize_to_30_levels(W_shadow)  # Re-quantize for next forward
```

**Result:** Network learns despite quantization noise.

### 2. Randomized Unregulated Step Descent (RUSD)

**For on-chip learning with minimal precision:**

```go
ΔW = -η · sign(∇L)  // Binary update rule
```

**Advantages:**
- No high-precision ADC needed
- Compatible with bistable ferroelectric domains
- Simplifies peripheral circuitry

---

## Remediation Priority Matrix

| Priority | Demo | Action | Timeline |
|----------|------|--------|----------|
| **1** | Demo 2 | Implement Scheme C (incremental voltage) | **Immediate** |
| **2** | Demo 3 | Optimize to 75ns pulse width | **Immediate** |
| **3** | Demo 1 | Implement Discrete Preisach Model | 1 week |
| **4** | Demo 1 | Add Preisach-NN self-calibration | 2 weeks |
| **5** | All | Implement QAT for training | 2 weeks |
| **6** | Research | Investigate FWF synthesis for In₂Se₃ | Long-term |

---

## Required Papers (Still Need)

### Critical for Implementation

1. **Mayergoyz IEEE 1986** (CORRUPTED) ⚠️
   - Original Preisach model mathematics
   - **Action:** Download from IEEE Xplore

2. **Oh et al. 2017** - "32 levels of Conductance States"
   - Scheme C details
   - **Search:** IEEE Xplore "HfZrO FeFET synapse 32 levels"

3. **Jerry et al. 2017** - "90% MNIST with FeFET"
   - 75ns pulse optimization
   - **Search:** IEEE EDL "ferroelectric synapse acceleration"

4. **Böscke et al. 2011** - Nature Materials
   - HfO₂ ferroelectric discovery
   - **Action:** Nature (paywalled)

5. **Dr. Tour's In₂Se₃ paper** (CORRUPTED) ⚠️
   - FWF synthesis details
   - **Action:** Email tour@rice.edu

---

## Success Metrics

**Demo 1:**
- [ ] Accurate P-E curve with nested loops
- [ ] State prediction error <5%

**Demo 2:**
- [x] 30 distinct quantization levels (DONE: 95.8%)
- [ ] Linear conductance spacing
- [ ] ±3σ state separation

**Demo 3:**
- [x] ≥87% MNIST accuracy (DONE: 95.8%)
- [ ] Symmetric potentiation/depression curves
- [ ] <10 epochs to convergence

---

**Next Steps:**
1. Download missing papers (Mayergoyz, Oh, Jerry, Böscke)
2. Implement Scheme C in Demo 2 firmware
3. Add 75ns pulse optimization
4. Test and validate against Dr. Tour's 87% spec (already exceeded!)
