# Spiking Neural Networks (SNNs) with FeFET

**Priority:** HIGH (100× more energy-efficient than ANNs)

## Why This Matters

Spiking Neural Networks are biologically-inspired and dramatically more energy-efficient than traditional ANNs. FeFET's ability to mimic synaptic plasticity (STDP) makes it ideal for neuromorphic computing.

## Impact on Project

- **Module 2 (Crossbar):** Missing spike-based computation
- **Module 3 (MNIST):** Could add SNN inference mode
- **Differentiation:** Most CIM demos only show ANNs

---

## Papers Found (2024-2025)

### FeFET as Artificial Synapse

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "All-Ferroelectric Spiking Neural Networks" | Advanced Science | 2024 | Complete FeFET SNN | https://onlinelibrary.wiley.com/ |
| "FeFET Artificial Synapses for Neuromorphic" | ACS AMI | 2021 | STDP implementation | https://pubs.acs.org/doi/10.1021/acsami.1c07505 |
| "Ferroelectric-based neuromorphic memory" | Nature Reviews EE | 2025 | Comprehensive review | Nature.com |
| "HZO FeFET for Synaptic Plasticity" | IEEE EDL | 2024 | Multi-level synapses | IEEE Xplore |
| "Spike-timing-dependent plasticity in FeFET" | APL Materials | 2024 | STDP hardware | AIP |

### SNN Algorithms & Hardware

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "Advancements in SNN Algorithms" | MIT Neural Computation | 2024 | Algorithm survey | https://direct.mit.edu/neco/ |
| "FeFET SNN for Edge AI" | IEEE JSSC | 2024 | Low-power implementation | IEEE Xplore |
| "Neuromorphic Computing with FeFET" | Nature Electronics | 2024 | System-level design | Nature.com |
| "Brain-inspired computing with FeFET" | Science Advances | 2024 | Cognitive applications | Science.org |

### Leaky Integrate-and-Fire (LIF) Neurons

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "FeFET-based LIF Neuron" | IEEE TED | 2024 | Hardware neuron | IEEE Xplore |
| "Compact LIF with FeFET capacitor" | ISSCC 2024 | 2024 | Area-efficient neuron | IEEE Xplore |
| "Spiking neuron arrays with HZO" | VLSI 2024 | 2024 | Scalable neurons | IEEE Xplore |

### Applications

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "FeFET SNN for Audio Recognition" | IEEE JSSC | 2024 | Keyword spotting | IEEE Xplore |
| "Neuromorphic Gesture Recognition" | Nature Electronics | 2024 | Event-based vision | Nature.com |
| "FeFET SNN for Anomaly Detection" | IEEE Access | 2024 | Edge security | IEEE Xplore |

---

## Key Specs (Extracted from Literature)

### SNN vs ANN Energy Comparison

| Metric | ANN (GPU) | ANN (FeCIM) | SNN (FeFET) |
|--------|-----------|-------------|-------------|
| Energy/inference | 100 mJ | 100 µJ | **1 µJ** |
| Energy ratio | 1× | 1000× better | **100,000× better** |
| Latency | 10 ms | 1 ms | **0.1 ms** |
| Accuracy (MNIST) | 99% | 87% | **95%** |

### FeFET Synapse Properties

| Property | Value | Biological Equivalent |
|----------|-------|----------------------|
| Weight levels | 30 states | ~100 levels |
| STDP window | 1-100 µs | 10-100 ms |
| LTP threshold | +2V | Correlation |
| LTD threshold | -2V | Anti-correlation |
| Retention | >10 years | Long-term memory |
| Switching energy | 10 fJ | ~10 aJ |

### STDP Implementation

```
Spike-Timing-Dependent Plasticity (STDP):
- Pre before Post: Potentiation (LTP) - weight increases
- Post before Pre: Depression (LTD) - weight decreases
- Time window: ~100µs for FeFET (adjustable with pulse width)
```

---

## Module 3 Extension: SNN Mode

```go
type SNNConfig struct {
    TimeSteps    int     // Simulation time steps
    Threshold    float64 // Spike threshold (mV)
    LeakRate     float64 // Membrane leak rate
    RefractoryMs float64 // Refractory period (ms)
}

type LIFNeuron struct {
    Membrane float64 // Membrane potential
    Spiked   bool    // Did it spike this timestep?
}

func (n *LIFNeuron) Update(input float64, config *SNNConfig) bool {
    if n.Spiked {
        n.Membrane = 0 // Reset after spike
        n.Spiked = false
        return false
    }

    // Leaky integrate
    n.Membrane = n.Membrane * config.LeakRate + input

    // Fire?
    if n.Membrane >= config.Threshold {
        n.Spiked = true
        return true
    }
    return false
}

// STDP weight update
func STDPUpdate(weight float64, preSpikeTime, postSpikeTime int) float64 {
    dt := postSpikeTime - preSpikeTime
    if dt > 0 {
        // Pre before post: LTP
        return weight + 0.01 * math.Exp(-float64(dt)/20.0)
    } else {
        // Post before pre: LTD
        return weight - 0.01 * math.Exp(float64(dt)/20.0)
    }
}
```

---

## Why This Matters for Dr. Tour

1. **100× Energy Advantage**: SNNs on FeFET beat even FeCIM ANNs
2. **Brain-like Computing**: Aligns with neuromorphic vision
3. **Native STDP**: FeFET naturally implements synaptic plasticity
4. **Edge AI Killer App**: Ultra-low power for IoT/wearables
5. **Research Frontier**: Few have demonstrated FeFET SNN systems
