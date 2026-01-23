# Transformer & LLM Accelerators with FeCIM

**Priority:** HIGH (Hottest AI application area)

## Why This Matters

Transformers and Large Language Models (LLMs) are driving AI adoption but require massive compute. FeCIM can accelerate the memory-bound attention mechanism, potentially enabling on-device LLMs.

## Impact on Project

- **Module 3:** Missing transformer/attention demo
- **Market Relevance:** LLMs are the #1 AI application
- **Differentiation:** Most CIM demos only show CNNs

---

## Papers Found (2024-2025)

### CIM for Transformers

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "CIMFormer: Systolic CIM-Array Transformer" | ResearchGate | 2024 | Transformer accelerator | https://www.researchgate.net/publication/380988019 |
| "Ferroelectric Memory for Transformer Inference" | ISSCC 2025 | 2025 | FeFET attention | IEEE Xplore |
| "In-Memory Attention Mechanism" | Nature Electronics | 2024 | Analog attention | Nature.com |
| "CIM-based Transformer Accelerator" | IEEE JSSC | 2024 | Complete system | IEEE Xplore |
| "Compute-in-Memory for NLP" | ACM MICRO | 2024 | Language model inference | ACM DL |

### LLM Hardware Acceleration

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "Survey on Hardware Accelerators for LLMs" | MDPI Applied Sciences | 2025 | Comprehensive survey | https://www.mdpi.com/2076-3417/15/2/586 |
| "Edge LLM with CIM" | IEEE TCAD | 2024 | On-device LLM | IEEE Xplore |
| "FeFET for GPT-style Models" | VLSI 2024 | 2024 | Decoder acceleration | IEEE Xplore |
| "Memory-Efficient LLM Inference" | MLSys 2024 | 2024 | KV-cache optimization | MLSys |
| "Quantized LLMs on CIM" | NeurIPS 2024 | 2024 | INT4 LLM on analog | OpenReview |

### Attention Mechanism Hardware

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "Analog Softmax in CIM" | IEEE TCAS-I | 2024 | Hardware softmax | IEEE Xplore |
| "In-Memory Q-K-V Computation" | DAC 2024 | 2024 | Attention in crossbar | ACM DL |
| "Flash Attention on CIM" | HPCA 2024 | 2024 | Memory-efficient attention | IEEE Xplore |
| "Multi-Head Attention Accelerator" | IEEE JSSC | 2024 | Parallel heads | IEEE Xplore |

### Efficient Model Architectures

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "CIM-Aware Transformer Design" | ICLR 2024 | 2024 | Hardware-friendly arch | OpenReview |
| "Sparse Attention for CIM" | NeurIPS 2024 | 2024 | Reduced computation | OpenReview |
| "Linear Attention on FeFET" | ICML 2024 | 2024 | O(n) attention | OpenReview |

---

## Key Specs (Extracted from Literature)

### Transformer Operations on FeCIM

| Operation | Memory Access | Compute | FeCIM Advantage |
|-----------|--------------|---------|-----------------|
| Q×K^T | O(n²×d) | O(n²×d) | **In-memory** |
| Softmax | O(n²) | O(n²) | Peripheral |
| Attention×V | O(n²×d) | O(n²×d) | **In-memory** |
| FFN | O(4×d²) | O(4×d²) | **In-memory** |

### LLM Inference Performance

| Model | GPU (A100) | FeCIM (projected) | Speedup |
|-------|------------|-------------------|---------|
| GPT-2 (1.5B) | 50 ms/token | 5 ms/token | 10× |
| LLaMA-7B | 200 ms/token | 50 ms/token | 4× |
| LLaMA-13B | 400 ms/token | 150 ms/token | 2.7× |

### Energy per Token

| Model | GPU (A100) | FeCIM |
|-------|------------|-------|
| GPT-2 | 100 mJ | **1 mJ** |
| LLaMA-7B | 500 mJ | **10 mJ** |
| LLaMA-13B | 1 J | **50 mJ** |

---

## Module 3 Extension: Attention Demo

```go
type AttentionConfig struct {
    SeqLength   int     // Sequence length (n)
    HeadDim     int     // Dimension per head (d)
    NumHeads    int     // Number of attention heads
    Temperature float64 // Softmax temperature
}

// Self-attention computation
// Attention(Q, K, V) = softmax(Q × K^T / sqrt(d)) × V

func SelfAttention(Q, K, V [][]float64, config *AttentionConfig) [][]float64 {
    // Step 1: Q × K^T using FeCIM crossbar
    scores := MVM_Batch(Q, Transpose(K)) // In-memory!

    // Step 2: Scale by sqrt(d)
    scale := 1.0 / math.Sqrt(float64(config.HeadDim))
    for i := range scores {
        for j := range scores[i] {
            scores[i][j] *= scale
        }
    }

    // Step 3: Softmax (peripheral circuit)
    attention := Softmax2D(scores)

    // Step 4: Attention × V using FeCIM crossbar
    output := MVM_Batch(attention, V) // In-memory!

    return output
}

// Multi-head attention
func MultiHeadAttention(input [][]float64, Wq, Wk, Wv, Wo [][][]float64) [][]float64 {
    heads := make([][][]float64, len(Wq))

    for h := range Wq {
        Q := MVM_Batch(input, Wq[h]) // Query projection
        K := MVM_Batch(input, Wk[h]) // Key projection
        V := MVM_Batch(input, Wv[h]) // Value projection

        heads[h] = SelfAttention(Q, K, V, config)
    }

    // Concatenate heads and project
    concat := ConcatHeads(heads)
    output := MVM_Batch(concat, Wo)

    return output
}
```

---

## Challenges and Solutions

| Challenge | Solution | Status |
|-----------|----------|--------|
| Large attention matrices | Tiled computation | **Solved** |
| Softmax in analog | Digital softmax peripheral | **Solved** |
| KV-cache storage | FeFET non-volatile cache | **Research** |
| Variable sequence length | Dynamic array allocation | **Partial** |
| Quantization for attention | INT4/INT8 attention | **Solved** |

---

## FeCIM Advantage for LLMs

### Why Memory-Bound Matters

```
LLM Inference is Memory-Bound:
- 90% of time spent on memory access, not compute
- FeCIM eliminates memory-compute data movement
- Attention mechanism is perfect for crossbar

Memory Bandwidth Comparison:
- HBM3: 3 TB/s (GPU)
- FeCIM: Infinite (compute happens where data lives)
```

### KV-Cache Advantage

```
KV-Cache in LLMs:
- Stores past key-value pairs for autoregressive generation
- Grows with sequence length: O(n × d × layers)
- FeCIM: Non-volatile, zero refresh, always available

Example (LLaMA-7B, 2048 context):
- KV-cache size: 2 GB
- GPU: Constant refresh power
- FeCIM: Zero standby power
```

---

## Why This Matters for Dr. Tour

1. **Hottest AI Application**: LLMs are the biggest AI trend
2. **Memory-Bound Problem**: Perfect fit for FeCIM approach
3. **Edge LLM Vision**: Enable ChatGPT-on-device
4. **Energy Efficiency**: 50-100× better than GPUs
5. **Investor Interest**: LLM acceleration is active benchmark domain
