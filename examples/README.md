# FeCIM Educational Examples

This directory contains standalone example programs demonstrating key concepts in Ferroelectric Compute-in-Memory (FeCIM).

Each example is designed to be:
- **Self-contained**: Run independently without additional setup
- **Educational**: Heavy comments explaining the physics and algorithms
- **Progressive**: Build understanding from fundamentals to applications

## Examples

### 1. `hysteresis_demo.go` - Ferroelectric Hysteresis

**Core Concept**: Ferroelectric materials exhibit bistable polarization states with hysteresis.

```bash
go run examples/hysteresis_demo.go
```

Demonstrates:
- HfO₂-ZrO₂ (HZO) material properties
- P-E hysteresis loop generation using Preisach model
- 30-level analog state encoding
- Temperature dependence of ferroelectric properties
- Non-volatile memory via remanent polarization

**Key Physics**:
- Coercive field (Ec): Electric field required to switch polarization
- Remanent polarization (Pr): Polarization retained at zero field
- Hysteresis: Path-dependent behavior enabling memory

### 2. `crossbar_mvm.go` - Crossbar Matrix-Vector Multiplication

**Core Concept**: Physics naturally computes matrix operations via Kirchhoff's laws.

```bash
go run examples/crossbar_mvm.go
```

Demonstrates:
- Crossbar array architecture (word lines, bit lines)
- Analog MVM via Ohm's law and current summation
- Quantization effects on accuracy
- Energy comparison: analog vs digital
- Neural network layer simulation

**Key Physics**:
- Ohm's Law: I = V × G (current = voltage × conductance)
- Kirchhoff's Current Law: I_col = Σ I_row (current conservation)
- Combined: Y = W × X computed in O(1) time!

### 3. `mnist_inference.go` - Neural Network Inference

**Core Concept**: FeCIM enables efficient neural network inference in memory.

```bash
go run examples/mnist_inference.go
```

Demonstrates:
- MNIST network architecture (784→128→10)
- Quantization-aware inference
- Softmax probability output
- Energy analysis per inference
- Hardware mapping to crossbar arrays

**Key Concepts**:
- 30 analog levels ≈ 5 bits per weight
- ~50 fJ per MAC (vs ~10 pJ digital)
- ~200× energy reduction for matrix operations

## Running the Examples

From the repository root:

```bash
# Run individual examples
go run examples/hysteresis_demo.go
go run examples/crossbar_mvm.go
go run examples/mnist_inference.go

# Get help
go run examples/hysteresis_demo.go --help
```

## Related Modules

For interactive visualization and deeper exploration:

| Module | Command | Description |
|--------|---------|-------------|
| Hysteresis | `fecim-lattice-tools hysteresis` | Interactive P-E loop visualization |
| Crossbar | `fecim-lattice-tools crossbar` | Crossbar array simulation GUI |
| MNIST | `fecim-lattice-tools mnist` | MNIST inference with live demo |

## Learning Path

1. **Start here**: `hysteresis_demo.go` - Understand the physics
2. **Next**: `crossbar_mvm.go` - See how physics enables computation
3. **Finally**: `mnist_inference.go` - Apply to real neural networks

## Physics References

- **Preisach Model**: Classical hysteresis model using distribution of hysterons
- **Landau-Khalatnikov**: Dynamic polarization switching equation
- **Kirchhoff's Laws**: Foundation of analog matrix computation
- **Quantization**: Maps continuous weights to discrete FeCIM levels

## Code Structure

Each example follows a consistent pattern:

```go
func main() {
    // 1. Introduction and key concepts
    // 2. Material/architecture setup
    // 3. Core demonstration
    // 4. Analysis and comparison
    // 5. Summary and next steps
}
```

Helper functions provide ASCII visualizations, making examples terminal-friendly without requiring graphics libraries.

## Contributing

To add a new example:

1. Create `your_example.go` in this directory
2. Include comprehensive comments explaining the physics
3. Add ASCII visualizations for terminal output
4. Update this README with the new example
5. Test with `go run examples/your_example.go`
