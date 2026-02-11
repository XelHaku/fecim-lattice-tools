# Transformer & LLM Accelerators with FeCIM

**Priority:** HIGH (Hottest AI application area)

## Why This Matters

Transformers and Large Language Models (LLMs) are driving AI adoption but require massive compute. FeCIM can accelerate the memory-bound attention mechanism, potentially enabling on-device LLMs.

## Impact on Project

- **Module 3:** Missing transformer/attention demo
- **Market Relevance:** LLMs are the #1 AI application
- **Differentiation:** Most CIM demos only show CNNs

---

## Papers in This Directory

### 1. Analog-Digital Hybrid Attention (2024)
**File:** `analog_digital_hybrid_attention_2024.pdf`

**Description:** Mixed analog-digital architecture for transformer attention mechanism. Performs Q×K^T and Attention×V in analog domain while keeping softmax digital.

**Key Findings:**
- Hybrid approach: Analog MVM + digital softmax
- 50× energy reduction vs all-digital attention
- Accuracy loss <0.5% on BERT and GPT-2 models
- Scalable to 2048 sequence length

**Relevance:** Practical architecture for FeFET transformer accelerators.

---

### 2. FAMOUS: FPGA Attention Accelerator (2024)
**File:** `famous_fpga_attention_accelerator_2024.pdf`

**Description:** High-performance FPGA-based attention accelerator with novel tiling strategy. Optimizes memory bandwidth and compute utilization.

**Key Findings:**
- 4.2 TFLOPS attention throughput on Xilinx VU9P
- Flash attention algorithm reduces memory by 4×
- Dynamic tiling adapts to sequence length
- 10× speedup vs CPU, 2× vs GPU for long sequences

**Relevance:** Reference architecture and optimization techniques for custom accelerators.

---

### 3. Hardware Acceleration for LLMs Survey (2024)
**File:** `hardware_acceleration_llm_survey_2024.pdf`

**Description:** Comprehensive survey of hardware accelerators for large language models. Covers GPU, TPU, CIM, and specialized architectures.

**Key Findings:**
- **Memory bottleneck:** LLM inference is 90% memory-bound
- CIM approaches: 10-100× energy efficiency potential
- Quantization: INT4/INT8 with <1% accuracy loss
- Attention optimization: Flash attention, sparse attention, linear attention
- Emerging technologies: CIM (ReRAM, FeFET), photonics, neuromorphic

**Relevance:** Complete market and technology landscape for LLM acceleration.

---

### 4. Memristor Transformer Self-Attention (2024)
**File:** `memristor_transformer_self_attention_2024.pdf`

**Description:** End-to-end memristor-based transformer implementation with in-memory attention computation. Demonstrates full BERT inference on analog hardware.

**Key Findings:**
- Complete transformer layer in memristor crossbar
- Q×K^T and Attention×V performed in-memory
- 100× energy efficiency vs GPU (5 mJ vs 500 mJ per inference)
- BERT-base on SQuAD: 88.1 F1 score (vs 88.5 FP32)

**Relevance:** Proof-of-concept for full transformer on FeFET-like devices.

---

## Key Findings Across All Papers

### Why LLMs are Perfect for CIM

**Memory-Bound Nature:**
- 90% of LLM inference time is memory access, not compute
- Attention mechanism: O(n²d) memory reads
- Feed-forward network: O(4d²) weight reads
- CIM eliminates data movement → 10-100× efficiency

**Transformer Operations Breakdown:**
```
Per-token inference (GPT-2):
├── Self-Attention (60%):  Q×K^T + Softmax + Attention×V
├── Feed-Forward (35%):    2× Linear layers (MVM)
└── LayerNorm (5%):        Digital peripheral
```

### Performance Metrics

| Model | GPU Latency | CIM Latency (projected) | GPU Energy | CIM Energy (projected) |
|-------|-------------|-------------------------|------------|------------------------|
| GPT-2 (1.5B) | 50 ms/token | 5-10 ms/token | 100 mJ | 1-2 mJ |
| BERT-base (110M) | 10 ms | 1-2 ms | 20 mJ | 0.2-0.5 mJ |
| LLaMA-7B | 200 ms/token | 50-100 ms/token | 500 mJ | 10-20 mJ |
| LLaMA-13B | 400 ms/token | 150-200 ms/token | 1 J | 50-100 mJ |

### Attention Optimization Techniques

1. **Flash Attention:** Reduce memory by 4× through tiling
2. **Sparse Attention:** Only compute important attention scores
3. **Linear Attention:** O(n) instead of O(n²) complexity
4. **Multi-Query Attention:** Share K/V across heads
5. **Sliding Window:** Local attention for long sequences

### Quantization for CIM

| Precision | BERT Accuracy | GPT-2 Perplexity | FeFET Suitability |
|-----------|---------------|------------------|-------------------|
| FP32 | 88.5% (SQuAD F1) | 18.5 | N/A |
| INT8 | 88.3% | 18.7 | ✅ Easy |
| **INT4** | **87.9%** | **19.2** | ✅ **FeFET: 4.9-bit** |
| INT2 | 84.1% | 25.3 | ⚠️ Marginal |

---

## Architecture Insights

### In-Memory Attention Flow
```
Input Embeddings (n × d)
    ↓
[Q×K^T in Crossbar] → Scores (n × n)
    ↓
[Digital Softmax] → Attention Weights (n × n)
    ↓
[Attention×V in Crossbar] → Output (n × d)
    ↓
[Feed-Forward in Crossbar] → Final Output
```

### FeFET Advantages for LLMs

1. **Non-volatile KV-cache:** Zero refresh power for stored context
2. **High density:** 30 analog levels (demo baseline; simulation baseline) = 4.9 bits effective precision
3. **In-memory MVM:** Eliminates weight transfer bottleneck
4. **Low energy:** 1-10 fJ per MAC (vs 1-10 pJ for SRAM)
5. **Scalability:** 3D stacking for billion-parameter models

### Critical Challenges

| Challenge | Solution | Status |
|-----------|----------|--------|
| Large attention matrices (n²) | Tiled computation | ✅ Solved |
| Softmax in analog | Digital peripheral circuit | ✅ Solved |
| Variable sequence length | Dynamic resource allocation | ⚠️ Partial |
| KV-cache management | FeFET non-volatile storage | 🔬 Research |
| Multi-head parallelism | Array partitioning | ✅ Solved |
| Quantization accuracy | INT4 with calibration | ✅ Solved |

---

## Related Topics

### Primary Connections
- **[Topic 4: Analog CIM](../04-cim-architectures/)** - Hardware foundation
  - Transformer = stack of MVM operations
  - Attention is perfect for crossbar arrays
  - Requires high-precision ADCs for softmax

- **[Topic 13: In-Memory Training](../13-in-memory-training/)** - LLM fine-tuning
  - On-chip fine-tuning for personalization
  - Federated learning on edge devices
  - LoRA (Low-Rank Adaptation) reduces training memory

### Secondary Connections
- **[Topic 2: HZO Materials](../01-ferroelectric-materials/)** - Scaling to large models
  - Billion-parameter models need high density
  - Retention critical for persistent KV-cache

- **[Topic 6: 3D Integration](../15-3d-stacking-architectures/)** - Memory capacity
  - Vertical stacking enables on-chip LLaMA-7B
  - Reduces inter-layer latency

- **[Topic 16: Photonic Hybrids](../16-photonic-ferroelectric-hybrids/)** - Future direction
  - Photonic attention for 1000× bandwidth
  - Hybrid electronic-photonic transformers

- **[Topic 8: Reliability](../08-reliability-retention/)** - Long-context models
  - KV-cache retention for multi-turn conversations
  - Error correction for critical weights

---

## Market Opportunity

### Edge LLM Applications
- **On-device ChatGPT:** Privacy + low latency
- **Wearable AI assistants:** Energy-efficient inference
- **Autonomous vehicles:** Real-time language understanding
- **IoT edge intelligence:** Local processing without cloud

### Projected Impact
- **Market size:** $50B by 2030 (edge AI inference)
- **Energy savings:** 50-100× vs GPU
- **Latency reduction:** 5-10× for memory-bound models
- **Privacy:** No data sent to cloud

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
