<!-- Category: Physics | Module: module3-mnist | Reading time: ~6 min -->
# Module 3 Physics: Neural Network Inference on Crossbar Hardware

> Equations for the feed-forward network, quantization model, and
> noise injection used in the MNIST digit recognition module.

## Prerequisites

- Vectors and matrices
- Basic neural network concepts (layers, activations)
- Probability and normalization

---

## 1. Network Architecture

A two-layer feed-forward network (multi-layer perceptron):

```
Input (784) --> Hidden (128) --> Output (10)
                  ReLU            Softmax
```

### Forward Pass

```
Hidden = ReLU(W1 * x + b1)
Logits = W2 * Hidden + b2
Probs  = softmax(Logits)
Prediction = argmax(Probs)
```

| Symbol | Meaning | Shape |
|--------|---------|-------|
| x | Input vector (flattened 28x28 image) | 784 x 1 |
| W1 | First layer weight matrix | 128 x 784 |
| b1 | First layer bias | 128 x 1 |
| W2 | Second layer weight matrix | 10 x 128 |
| b2 | Second layer bias | 10 x 1 |

### Activation Functions

```
ReLU(z) = max(0, z)

softmax(z_i) = exp(z_i) / sum_j(exp(z_j))
```

---

## 2. Quantization Model

The CIM path quantizes weights to L discrete levels (default L = 30):

```
QuantizedWeight = -W_max + round((w + W_max) / (2 * W_max) * (L - 1)) * (2 * W_max / (L - 1))
```

| Symbol | Meaning |
|--------|---------|
| w | Original continuous weight |
| W_max | Maximum absolute weight (symmetric range) |
| L | Number of quantization levels (30) |

This is uniform symmetric quantization: the weight range [-W_max, +W_max]
is divided into L equally-spaced bins.

### Quantization Error

```
Error per weight = |w - QuantizedWeight|
Max quantization step = 2 * W_max / (L - 1)

For L = 30 and W_max = 1.0:
  Step size = 2.0 / 29 = 0.069
```

---

## 3. Noise Model

Device variation is modeled as multiplicative Gaussian noise:

```
NoisyValue = v + N(0, 1) * |v| * sigma_over_mu
```

| Symbol | Meaning | Typical Value |
|--------|---------|---------------|
| v | Clean value (weight or activation) | varies |
| N(0,1) | Standard normal random variable | -- |
| sigma_over_mu | Noise coefficient | 0.00 - 0.20 |

The noise is proportional to the signal magnitude (multiplicative). Larger
weights get proportionally larger noise, matching the physics of device
variation where larger conductances have proportionally larger absolute
deviation.

---

## 4. Parameters and Units

| Symbol | Meaning | Units |
|--------|---------|-------|
| x | Input vector (784 pixels) | unitless [0, 1] |
| W | Weight matrix | unitless |
| b | Bias vector | unitless |
| L | Quantization levels | count |
| W_max | Symmetric weight range | unitless |
| sigma_over_mu | Noise coefficient | unitless |

All values are unitless because the network operates on normalized inputs
and the crossbar conductance mapping is handled separately in Module 2.

---

## 5. Dual-Path Comparison

The module runs both paths on the same input:

```
                    +-- [Full precision W] -- ReLU -- [Full precision W] -- Softmax --> Prediction A
                   /
Input image ------+
                   \
                    +-- [Quantized W + noise] -- ReLU -- [Quantized W + noise] -- Softmax --> Prediction B
```

Accuracy is measured over the test set (10,000 images):

```
Accuracy = (correct predictions / total images) * 100%
Gap = Accuracy_full - Accuracy_CIM
```

---

## 6. Assumptions and Limits

- Architecture is a small MLP for visualization speed, not SOTA accuracy
- Quantization is uniform; per-layer or non-uniform schemes are possible
  but not implemented
- Noise is modeled as multiplicative Gaussian, not from device physics
- Training is done offline; the GUI focuses on inference
- No ADC/DAC quantization modeled at this level (handled in Module 2)

---

## Where It Lives in Code

| Component | File |
|-----------|------|
| Network inference | `module3-mnist/pkg/core/network_inference.go` |
| Quantization | `module3-mnist/pkg/core/quantize.go` |
| Dual-mode UI | `module3-mnist/pkg/gui/dualmode.go` |
| MNIST data loader | `module3-mnist/pkg/mnist/` |
| Accuracy sweep analysis | `shared/neural/accuracy_sweep.go` |

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*
