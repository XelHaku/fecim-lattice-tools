# Module 3: MNIST - Physics

## Evidence Status (Demonstrated vs Modeled vs Aspirational)

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

## Prerequisites

- Vectors and matrices
- Basic neural network concepts
- Probability and normalization

## Core Model

- A feed-forward network computes logits from inputs.
- ReLU introduces nonlinearity; softmax produces probabilities.
- The CIM path applies quantization and noise to weights and activations.

## Key Equations (Simplified)

```
Hidden = ReLU(W1 * x + b1)
Logits = W2 * Hidden + b2
Probs = softmax(Logits)
QuantizedWeight = -Wmax + round((w + Wmax) / (2*Wmax) * (L-1)) * (2*Wmax/(L-1))
NoisyValue = v + N(0,1) * |v| * sigma_over_mu
```

## Parameters And Units

| Symbol | Meaning | Units |
|---|---|---|
| x | Input vector (784) | unitless |
| W | Weight matrix | unitless |
| b | Bias vector | unitless |
| L | Quantization levels | levels |
| Wmax | Max absolute weight for symmetric quantization | unitless |
| sigma_over_mu | Noise coefficient (sigma/mu) | unitless |
| W1, W2 | Layer weight matrices | unitless |
| b1, b2 | Layer bias vectors | unitless |
| N(0,1) | Standard normal random variable | unitless |

## Assumptions And Limits

- The architecture is a small MLP for visualization speed.
- Quantization is uniform with optional per-layer levels.
- Noise is modeled as multiplicative Gaussian perturbations (sigma/mu).
- Quantization formulas are software-level approximations, not device-physics models.

## Where It Lives In Code

- `module3-mnist/pkg/core/network_inference.go`
- `module3-mnist/pkg/core/quantize.go`
- `module3-mnist/pkg/gui/dualmode.go`

## Sources

- `docs/development/SCRIPT_REFERENCE.md#demo-3-mnist-module3-mnist`
- `module3-mnist/pkg/core/network_inference.go`
