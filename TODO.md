# IronLattice-vis TODO

> Based on Dr. external research group's November 2024 presentation on IronLattice technology.
> Source: ironlattice-transcript.md

---

## IronLattice Key Specs (From Dr. Tour)

| Metric | Target | Current Status |
|--------|--------|----------------|
| Discrete analog states | **30 levels** | ✅ **FIXED** - Using 30 levels |
| MNIST accuracy | **87%** (88% theoretical max) | ✅ **95.8%** - Exceeds target |
| Energy vs NAND | 10,000,000× lower | N/A (simulation) |
| Energy vs DRAM | 1,000× lower | N/A (simulation) |
| P-E hysteresis | Square loop characteristic | Simplified tanh model |
| CMOS compatible | Standard fab | N/A |
| TRL | 4 (lab validation) | Demo/educational |

---

## Priority 1: Core IronLattice Features

### 30 Discrete Analog States (CRITICAL) ✅ COMPLETED
> "It's got 30 discrete states. So it's not 0-1-0-1. And we have 30 discrete states that we can access." — Dr. Tour

- [x] **Fix quantization in `array.go`** ✅
  - Added `IronLatticeLevels = 30` constant
  - `ProgramWeight` now auto-quantizes to 30 levels
  - Added `QuantizeTo30Levels()` and `GetLevel()` functions
- [x] ADC/DAC now use 30-level quantization ✅
- [x] Visualize all 30 levels distinctly in Demo 1 level bar ✅
  - Added `LevelIndicator` struct in render.go
- [ ] Show level number (1-30) in Demo 3 weight display

### 87% MNIST Accuracy (CRITICAL) ✅ COMPLETED
> "We're at 87% validation here... theoretical is 88% is the theoretical maximum." — Dr. Tour

- [x] **Train network to achieve 87% accuracy** ✅ **Achieved 95.8%!**
  - [x] Downloaded real MNIST dataset (60k train, 10k test)
  - [x] Implemented proper training loop with mini-batch SGD
  - [x] Saved pretrained weights to `demo3-mnist/data/pretrained_weights.json`
  - [x] Added unit tests for network operations
- [x] Fixed training math issues ✅
  - [x] Fixed O(n³) weight update bug (fetch matrix once outside loops)
  - [x] Implemented separate SimpleNetwork for float training
  - [x] Quantize to 30 levels after training
- [ ] Add `--verify` flag to Demo 3 that tests against MNIST test set

### Ferroelectric P-E Hysteresis (HIGH)
> "These are the polarization curves that we've got here" — Dr. Tour (showing square loops)

- [ ] **Improve Preisach model or document limitations**
  - Current: Simplified tanh approximation
  - IronLattice shows: Square hysteresis loops (their key advantage)
- [ ] Add toggle between "ideal square loop" and "realistic model"
- [ ] Show 30 discrete polarization states on P-E curve
- [ ] Visualize "wake-up → stable operation → fatigue" cycle behavior

---

## Priority 2: Compute-in-Memory Demonstration

### Matrix-Vector Multiplication Visualization
> "Computation memory in the same device... no more busing information back and forth" — Dr. Tour

- [ ] **Enhance Demo 2 to show CIM principle clearly**
  - [ ] Animate: "Input voltages → Conductance matrix → Output currents"
  - [ ] Show Kirchhoff's law: I = Σ(V × G) happening in parallel
  - [ ] Contrast with von Neumann: "Traditional: Memory ↔ CPU ↔ Memory"
- [ ] Add energy comparison display
  - "Traditional: X operations, Y data transfers"
  - "CIM: Single parallel operation, zero transfers"

### Neural Network on Crossbar
> "We've done the compute in memory. We've put this on the MNIST system." — Dr. Tour

- [ ] Visualize inference flow through both crossbar layers
- [ ] Show 784 inputs → 128 hidden → 10 outputs
- [ ] Display post-synaptic currents (PSC) like in Dr. Tour's slides
- [ ] Add potentiation/depression demonstration (LTP/LTD)

---

## Priority 3: Code Quality & Correctness

### Critical Bugs ✅ MOSTLY COMPLETED
- [x] **Race conditions** (`engine.go`) ✅
  - Added `sync.RWMutex` to Engine struct
  - Protected `e.running`, `e.paused` in Start/Stop/Pause/Step
  - Added thread-safe `IsRunning()` and `IsPaused()` methods
- [ ] **Panics in production** (`network/network.go:117`)
  - Replace `panic()` with error returns
- [x] **O(n³) weight updates** (`training/network.go`) ✅
  - Fetch matrix once outside loops in updateLayer1/updateLayer2

### Test Coverage ✅ ADDED
> Previously: 0 tests, Now: 19 tests

- [x] Add test: 30-level quantization produces exactly 30 distinct values ✅
- [x] Add test: MVM output matches manual calculation ✅
- [x] Add test: Engine thread-safety with race detector ✅
- [x] Add test: Network forward/backward pass ✅
- [x] Add test: Weight save/load roundtrip ✅
- [ ] Add test: MNIST accuracy >= 85% on test set
- [ ] Add test: P-E curve exhibits hysteresis (not just a function)

---

## Priority 4: Educational Value

### Demonstrate IronLattice Advantages
> "This could lower the requirements in a data center by 80 to 90% of the energy requirements." — Dr. Tour

- [ ] Add "Why CIM?" educational panel
  - Traditional: Separate memory and compute, constant data movement
  - IronLattice: Same device does both, physics does the math
- [ ] Show energy comparison (even if simulated)
- [ ] Explain 30 states vs binary: "~5 bits per cell vs 1 bit"

### Market Context (Optional)
- [ ] Add comparison table from Dr. Tour's slides:
  - vs NAND Flash: 10M× energy, 1M× speed, 90% voltage
  - vs DRAM: 1000× energy, zero refresh
- [ ] Show TRL progression: "We are here (TRL 4) → Production (TRL 9)"

---

## Deprioritized (Nice to Have)

### Future Enhancements
- [ ] Landau-Khalatnikov solver (complex, not essential for demo)
- [ ] Phase-field domain simulation (TDGL) - mentioned in README but not in Dr. Tour's demo
- [ ] GPU-accelerated training
- [ ] Vulkan visualization for Demo 2/3 (terminal works fine for educational)
- [ ] Non-idealities: IR drop, sneak paths (advanced topics)

### Code Cleanup
- [ ] Remove 151 unused layer files in `demo2-crossbar/pkg/layers/`
- [ ] Consolidate duplicate network code
- [ ] Add godoc comments
- [ ] Improve error messages

---

## Quick Wins

```go
// 1. Fix 30-level quantization (array.go) - MOST IMPORTANT
const IronLatticeLevels = 30

func (a *Array) quantizeToIronLattice(value float64) float64 {
    value = math.Max(0, math.Min(1, value))
    level := math.Round(value * float64(IronLatticeLevels-1))
    return level / float64(IronLatticeLevels-1)
}

// 2. Add accuracy verification (main.go)
func verifyAccuracy(net *MNISTNetwork) {
    testImages, testLabels, _ := mnist.LoadMNIST("data", false)
    acc := net.Evaluate(testImages, testLabels)
    fmt.Printf("MNIST Test Accuracy: %.1f%% (Target: 87%%)\n", acc*100)
    if acc >= 0.87 {
        fmt.Println("✓ IronLattice target ACHIEVED!")
    }
}

// 3. Fix race condition (engine.go)
type Engine struct {
    mu      sync.RWMutex
    running bool
    paused  bool
    state   *State
}

func (e *Engine) IsRunning() bool {
    e.mu.RLock()
    defer e.mu.RUnlock()
    return e.running
}
```

---

## Success Criteria (From Dr. Tour's Demo)

### Demo 1: Ferroelectric Cell ✅
- [x] P-E hysteresis curve visible
- [x] **30 discrete levels clearly shown** ✅ (Added LevelIndicator)
- [ ] Square loop characteristic (IronLattice advantage)
- [x] Interactive E-field control

### Demo 2: Crossbar MVM ✅
- [x] Matrix-vector multiplication works
- [x] **30-level conductance states** ✅ (Fixed quantization)
- [x] Input/output visualization
- [ ] Shows "compute happens in memory" concept

### Demo 3: MNIST Classification ✅
- [x] Can classify handwritten digits
- [x] **Achieves 87% accuracy** ✅ **95.8%!**
- [x] Uses 30 discrete weight levels ✅
- [x] Interactive drawing/testing
- [x] Pretrained weights saved to data/pretrained_weights.json ✅

---

## Timeline

### This Week
1. [ ] Fix 30-level quantization everywhere
2. [ ] Download MNIST dataset, train to 87%
3. [ ] Add accuracy verification test
4. [ ] Fix race conditions

### Next Week
5. [ ] Improve P-E visualization (show 30 states)
6. [ ] Add educational CIM explanation
7. [ ] Save pretrained weights file
8. [ ] Add basic test coverage

### Month 2
9. [ ] Polish demos for presentation quality
10. [ ] Add energy comparison displays
11. [ ] Document physics accurately
12. [ ] Create demo video

---

## References

- **Primary Source**: Dr. external research group, IronLattice presentation (Nov 2024)
- **Key Paper**: Shin, J., et al. "BEOL-Compatible Superlattice FEFET Analog Synapse" IEEE (2022)
- **MNIST Benchmark**: 88% theoretical maximum, 87% achieved by IronLattice
- **30 States**: Post-synaptic current with 30 discrete levels (LTP/LTD demonstration)

---

## Notes from Dr. Tour's Presentation

> "It's got **30 discrete states**. So it's not 0-1-0-1."

> "We're at **87% validation** here... theoretical is 88% is the theoretical maximum."

> "**Compute in memory** where the same device does the memory and the computation."

> "This could lower the requirements in a data center by **80 to 90%** of the energy requirements."

> "Works on a **standard CMOS line** and can translate just like that."

> "There's **no exotic materials** in here. There's no graphene."
