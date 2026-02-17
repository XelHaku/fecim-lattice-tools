# Module 3 Automated Testing Plan — Research-Grade MNIST Inference Validation

## 1. Purpose

Build a **research-grade** automated test pipeline for Module 3 (`module3-mnist`) that validates MNIST digit classification accuracy, quantization effects, noise robustness, and inference performance at publication quality. Module 3 demonstrates analog compute-in-memory (CIM) using ferroelectric crossbar arrays for neuromorphic inference.

### Scope

| Grade | Definition | Status |
|-------|------------|--------|
| **Simulation-grade** | Network runs, produces output, basic tests pass | Current (189 tests) |
| **Research-grade** | Accuracy validated, quantization characterized, noise-robust, energy-efficient, reproducible | Target |

### Current State

| Metric | Value |
|--------|-------|
| Source files | 47 (.go) |
| Test files | 28 (*_test.go) |
| Test functions | 189 |
| Packages | 4 (core, mnist, training, gui) |

---

## 2. Architecture Overview

### Inference Pipeline

```
MNIST Test Set (10,000 images, 28×28 grayscale)
    │
    ▼
Input Quantization (8-bit → N-bit)
    │
    ▼
Layer 1: 784 inputs → 128 neurons (W1: 784×128 weights)
    │  └─→ Crossbar MVM (analog CIM physics)
    ▼
ReLU Activation
    │
    ▼
Layer 2: 128 → 10 outputs (W2: 128×10 weights)
    │  └─→ Crossbar MVM (analog CIM physics)
    ▼
Argmax → Predicted Digit (0-9)
    │
    ▼
Accuracy = Correct / Total
```

### Key Physics

- **Quantization**: Weights/activations mapped to N-bit conductance levels
- **CIM Noise**: Device-to-device variation, read disturb, drift
- **Energy Model**: Dynamic (switching) + static (leakage) power
- **Throughput**: Inferences per second (latency-bound or batch-bound)

---

## 3. Test Plan — 6 Phases

### Phase 1: Baseline Accuracy Validation (P0 — foundational)

**Goal**: Verify network achieves ≥80% accuracy on MNIST test set (published baseline).

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M3-ACC-01 | `accuracy_baseline_test.go` | Floating-point baseline | FP32 weights → accuracy ≥95% |
| M3-ACC-02 | `accuracy_baseline_test.go` | 8-bit quantized | 8-bit weights/activations → accuracy ≥80% |
| M3-ACC-03 | `accuracy_per_digit_test.go` | Per-digit accuracy | All digits 0-9 ≥70% individually |
| M3-ACC-04 | `accuracy_confusion_matrix_test.go` | Confusion matrix | No class with >50% misclassification |
| M3-ACC-05 | `accuracy_determinism_test.go` | Same input → same output | Bit-identical results (same seed) |

**Acceptance**: ≥80% overall, ≥70% per digit, deterministic.

### Phase 2: Quantization Impact (P0)

**Goal**: Characterize accuracy vs bit width (2-8 bits).

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M3-QUANT-01 | `quantization_sweep_test.go` | 2-bit, 4-bit, 6-bit, 8-bit | Accuracy vs bits (monotonic) |
| M3-QUANT-02 | `quantization_weight_test.go` | Weight-only quantization | Weights N-bit, activations 8-bit |
| M3-QUANT-03 | `quantization_activation_test.go` | Activation-only quantization | Activations N-bit, weights 8-bit |
| M3-QUANT-04 | `quantization_clipping_test.go` | Saturation effects | Extreme values don't cause NaN/Inf |

**Acceptance**: Accuracy monotonic with bits, 8-bit ≥80%, 4-bit ≥60%, 2-bit ≥30%.

### Phase 3: Noise & Variation Robustness (P1)

**Goal**: Validate inference under CIM non-idealities (noise, drift, variation).

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M3-NOISE-01 | `noise_injection_test.go` | Gaussian weight noise | σ=0.01, 0.05, 0.10 → accuracy degradation |
| M3-NOISE-02 | `noise_read_disturb_test.go` | Read disturb simulation | Conductance shift per read |
| M3-NOISE-03 | `noise_process_variation_test.go` | Device-to-device variation | Monte Carlo N=100 runs |
| M3-NOISE-04 | `noise_drift_test.go` | Temporal drift | Conductance drift over time |

**Acceptance**: Accuracy ≥70% with σ=0.05 noise, ≥60% with σ=0.10.

### Phase 4: Inference Performance (P1)

**Goal**: Benchmark latency and throughput.

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M3-PERF-01 | `performance_latency_test.go` | Single inference latency | Time per image (ms) |
| M3-PERF-02 | `performance_throughput_test.go` | Batch throughput | Images/second for batch=100 |
| M3-PERF-03 | `performance_scaling_test.go` | Batch size scaling | Throughput vs batch size (1, 10, 100, 1000) |
| M3-PERF-04 | `performance_memory_test.go` | Memory footprint | Bytes allocated per inference |

**Acceptance**: Latency < 10ms per image, throughput > 100 img/s for batch=100.

### Phase 5: Energy Model Validation (P1)

**Goal**: Verify energy calculations match physics.

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M3-ENERGY-01 | `energy_dynamic_test.go` | Dynamic energy | E_dyn = C × V² × transitions |
| M3-ENERGY-02 | `energy_static_test.go` | Leakage power | P_leak > 0, scales with array size |
| M3-ENERGY-03 | `energy_inference_test.go` | Energy per inference | Total energy (fJ-pJ range) |
| M3-ENERGY-04 | `energy_efficiency_test.go` | Energy efficiency | Energy/inference/accuracy (ENAC metric) |

**Acceptance**: Energy in physically realistic range (10-100 pJ/inference), dynamic > static.

### Phase 6: GUI & Integration (P2)

**Goal**: Validate GUI lifecycle and cross-module integration.

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M3-GUI-01 | `gui_lifecycle_test.go` | Tab lifecycle | BuildContent/Start/Stop without panic |
| M3-GUI-02 | `gui_inference_test.go` | Live inference | Feed image → get prediction |
| M3-GUI-03 | `gui_export_test.go` | Result export | CSV/JSON export valid |
| M3-INT-01 | `integration_m2_m3_test.go` | M2 crossbar → M3 inference | Weights programmed → classify |

**Acceptance**: No panics, inference works, exports valid, M2↔M3 integration passes.

---

## 4. Automation Scripts

### Fast Gate (CI — < 30s)

```bash
#!/bin/bash
# scripts/run_module3_fast_gate.sh
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
echo "=== M3 Fast Gate ==="
go build ./module3-mnist/...
go vet ./module3-mnist/...
go test -short -count=1 ./module3-mnist/...
echo "=== M3 Fast Gate PASS ==="
```

### Full Gate (Nightly — < 5min)

```bash
#!/bin/bash
# scripts/run_module3_full_gate.sh
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
echo "=== M3 Full Gate ==="
go build ./module3-mnist/...
go vet ./module3-mnist/...
go test -race -count=1 ./module3-mnist/...
echo "=== M3 Full Gate PASS ==="
```

---

## 5. Evidence Requirements

Every test must produce:
1. **Exact accuracy percentages** (not "it works")
2. **Per-digit breakdown** (confusion matrix or table)
3. **Latency/throughput with units** (ms, img/s)
4. **Energy values with units** (pJ, fJ)
5. **Deterministic** — same seed → same result

### Reporting Format

```
M3-ACC-02: 8-bit quantized accuracy = 82.4% (8240/10000 correct) — PASS
M3-QUANT-01: 2-bit=34.2%, 4-bit=68.1%, 8-bit=82.4% (monotonic ✓) — PASS
M3-PERF-01: Single inference latency = 2.3 ms/image — PASS
M3-ENERGY-03: Energy per inference = 45.7 pJ (dynamic=41.2 pJ, static=4.5 pJ) — PASS
```

---

## 6. Priority & Timeline

| Phase | Priority | Est. Tests | Description |
|-------|----------|-----------|-------------|
| Phase 1 | P0 | 10 | Baseline accuracy — foundational validation |
| Phase 2 | P0 | 8 | Quantization — core CIM characteristic |
| Phase 3 | P1 | 8 | Noise robustness — non-ideality resilience |
| Phase 4 | P1 | 8 | Performance — latency/throughput benchmarks |
| Phase 5 | P1 | 8 | Energy model — power consumption |
| Phase 6 | P2 | 8 | GUI & Integration — end-to-end |
| **Total** | | **50** | |

Phase 1-2 (P0) are mandatory for research claims.
Phase 3-5 (P1) are required for publication.
Phase 6 (P2) is for completeness.

---

## 7. Cross-Module Validation Chain

```
M1 (Hysteresis) → Conductance levels → M3 (MNIST) weight mapping
M2 (Crossbar) → MVM physics → M3 (MNIST) matrix multiplication
M4 (Circuits) → DAC/ADC quantization → M3 (MNIST) bit widths
```

Test M3-INT-01 validates this chain.

---

## 8. Literature Benchmarks

- **MNIST Baseline**: 98-99% (LeNet-5, full precision)
- **4-bit Quantized**: 96-97% (typical CNN)
- **CIM Implementations**: 80-90% (FeCIM, ReRAM, PCM papers 2018-2024)
- **Noise Tolerance**: σ=0.10 → 5-10% accuracy drop (published CIM work)
- **Energy/Inference**: 10-100 pJ (analog CIM state-of-the-art)

Target: Match or exceed published CIM baselines for FeCIM devices.

---

## 9. Dataset Requirements

- **MNIST Test Set**: 10,000 images (standard)
- **Subset for CI**: 1,000 images (for fast gate)
- **Data location**: `data/mnist/` or download on first run
- **Format**: IDX format (ubyte pixel values, uint8 labels)

---

## 10. Known Gaps (Current State)

| Gap | Impact | Remediation |
|-----|--------|-------------|
| No full 10k accuracy test | Accuracy unvalidated | Add M3-ACC-02 with full test set |
| No quantization sweep | Bit width impact unknown | Add M3-QUANT-01 sweep 2-8 bits |
| No noise injection tests | Robustness uncharacterized | Add M3-NOISE-01 Gaussian noise |
| No energy validation | Power claims unverified | Add M3-ENERGY-01 physics check |
| No throughput benchmarks | Performance unknown | Add M3-PERF-02 batch throughput |

All gaps addressed in Phases 1-6.

---

## 11. Acceptance Criteria Summary

✅ **Accuracy**: ≥80% on 10k test set (8-bit quantized)
✅ **Per-digit**: All digits ≥70%
✅ **Quantization**: Monotonic accuracy vs bits
✅ **Noise**: ≥70% accuracy with σ=0.05 weight noise
✅ **Latency**: < 10 ms per inference
✅ **Throughput**: > 100 img/s (batch=100)
✅ **Energy**: 10-100 pJ per inference
✅ **Determinism**: Bit-identical results (same seed)
✅ **No regressions**: M1+M2 test counts stable
✅ **Build/vet**: CLEAN
