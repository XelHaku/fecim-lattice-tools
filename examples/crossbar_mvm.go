//go:build ignore

// crossbar_mvm.go - Educational demonstration of crossbar matrix-vector multiplication
//
// This program demonstrates how ferroelectric crossbar arrays perform analog
// matrix-vector multiplication (MVM) - the core operation of compute-in-memory.
//
// Key Concepts:
//   - Crossbar arrays store weights as analog conductances in FeCIM cells
//   - Input voltages applied to rows, currents collected at columns
//   - Kirchhoff's laws naturally compute Y = W × X (matrix-vector product)
//   - O(1) time complexity for MVM regardless of matrix size!
//
// Physics:
//   - I_j = Σ_i V_i × G_ij  (sum of currents from all rows)
//   - G_ij = conductance of cell (i,j), proportional to polarization state
//   - This is analog dot product computed in parallel by physics!
//
// Run: go run examples/crossbar_mvm.go
//
// For detailed architecture, see: docs/crossbar-architecture/
package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
)

// CrossbarArray represents a ferroelectric compute-in-memory crossbar
type CrossbarArray struct {
	Rows    int           // Number of word lines (input size)
	Cols    int           // Number of bit lines (output size)
	Weights [][]float64   // Conductance matrix [Rows][Cols]
	Levels  int           // Number of discrete analog levels (e.g., 30)
	Gmin    float64       // Minimum conductance (S)
	Gmax    float64       // Maximum conductance (S)
	Noise   float64       // Conductance variation (σ/μ)
	IRDrop  bool          // Simulate IR drop effects
	Rline   float64       // Wire resistance (Ω)
	Vread   float64       // Read voltage (V)
}

// NewCrossbarArray creates a new crossbar with given dimensions
func NewCrossbarArray(rows, cols, levels int) *CrossbarArray {
	cb := &CrossbarArray{
		Rows:   rows,
		Cols:   cols,
		Levels: levels,
		Gmin:   1e-7,   // 100 nS minimum
		Gmax:   1e-5,   // 10 µS maximum
		Noise:  0.01,   // 1% device variation
		Rline:  5.0,    // 5 Ω wire resistance
		Vread:  0.2,    // 200 mV read voltage
	}

	// Initialize weight matrix
	cb.Weights = make([][]float64, rows)
	for i := range cb.Weights {
		cb.Weights[i] = make([]float64, cols)
	}

	return cb
}

// SetWeightsFromMatrix programs the crossbar with normalized weights [0, 1]
func (cb *CrossbarArray) SetWeightsFromMatrix(W [][]float64) error {
	if len(W) != cb.Rows {
		return fmt.Errorf("weight rows %d != crossbar rows %d", len(W), cb.Rows)
	}
	if len(W[0]) != cb.Cols {
		return fmt.Errorf("weight cols %d != crossbar cols %d", len(W[0]), cb.Cols)
	}

	// Quantize weights to discrete levels and map to conductance
	for i := range W {
		for j := range W[i] {
			// Clamp to [0, 1]
			w := W[i][j]
			if w < 0 {
				w = 0
			}
			if w > 1 {
				w = 1
			}

			// Quantize to discrete levels
			level := int(math.Round(w * float64(cb.Levels-1)))
			quantizedW := float64(level) / float64(cb.Levels-1)

			// Map to conductance: G = Gmin + w × (Gmax - Gmin)
			G := cb.Gmin + quantizedW*(cb.Gmax-cb.Gmin)

			// Add device-to-device variation
			if cb.Noise > 0 {
				G *= 1.0 + cb.Noise*rand.NormFloat64()
			}

			// Clamp to physical limits
			if G < cb.Gmin {
				G = cb.Gmin
			}
			if G > cb.Gmax {
				G = cb.Gmax
			}

			cb.Weights[i][j] = G
		}
	}

	return nil
}

// MVM performs matrix-vector multiplication: Y = W × X
// Input X is voltages, output Y is currents (normalized)
func (cb *CrossbarArray) MVM(X []float64) ([]float64, error) {
	if len(X) != cb.Rows {
		return nil, fmt.Errorf("input size %d != rows %d", len(X), cb.Rows)
	}

	Y := make([]float64, cb.Cols)

	// Scale input to voltages
	V := make([]float64, cb.Rows)
	for i := range X {
		V[i] = X[i] * cb.Vread // Map [0,1] input to [0, Vread]
	}

	// Compute column currents using Kirchhoff's Current Law
	// I_j = Σ_i V_i × G_ij
	for j := 0; j < cb.Cols; j++ {
		var current float64
		for i := 0; i < cb.Rows; i++ {
			// Ohm's law: I = V × G
			current += V[i] * cb.Weights[i][j]
		}

		// Optional: IR drop correction (simplified model)
		if cb.IRDrop {
			// IR drop reduces effective voltage at far end of array
			irLoss := current * cb.Rline * float64(cb.Rows)
			current *= (1.0 - irLoss/(cb.Vread*float64(cb.Rows)))
		}

		Y[j] = current
	}

	// Normalize output to [0, 1] range
	Imax := float64(cb.Rows) * cb.Vread * cb.Gmax
	for j := range Y {
		Y[j] /= Imax
	}

	return Y, nil
}

// GetEnergyPerMVM returns estimated energy in picojoules
func (cb *CrossbarArray) GetEnergyPerMVM() float64 {
	// Energy model from FeCIM literature:
	// E ≈ 50 fJ per MAC at 30 levels (Jerry et al. IEDM 2017)
	// Scales with log2(levels): E = 10 fJ × log2(levels) per MAC
	macs := cb.Rows * cb.Cols
	bitsPerCell := math.Log2(float64(cb.Levels))
	energyPerMAC := 10.0 * bitsPerCell // femtojoules
	return float64(macs) * energyPerMAC / 1000 // picojoules
}

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║     FeCIM Crossbar Matrix-Vector Multiplication Demo             ║")
	fmt.Println("║     Analog Compute-in-Memory Architecture                        ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// =========================================================================
	// 1. CROSSBAR ARCHITECTURE
	// =========================================================================
	fmt.Println("1. CROSSBAR ARRAY ARCHITECTURE")
	fmt.Println("   └─ Word lines (rows) carry input voltages")
	fmt.Println("   └─ Bit lines (columns) collect output currents")
	fmt.Println("   └─ Each junction is a ferroelectric memory cell")
	fmt.Println()

	printCrossbarDiagram()

	// =========================================================================
	// 2. SIMPLE EXAMPLE: 3×3 Matrix
	// =========================================================================
	fmt.Println("\n2. EXAMPLE: 3×3 Matrix-Vector Multiplication")
	fmt.Println()

	// Create a small crossbar for demonstration
	cb := NewCrossbarArray(3, 3, 30)
	cb.Noise = 0 // Disable noise for clean demo

	// Define a simple weight matrix
	W := [][]float64{
		{0.5, 0.0, 1.0},
		{0.0, 1.0, 0.5},
		{1.0, 0.5, 0.0},
	}

	fmt.Println("   Weight Matrix W (normalized [0,1]):")
	printMatrix(W)

	cb.SetWeightsFromMatrix(W)

	// Input vector
	X := []float64{1.0, 0.5, 0.0}
	fmt.Printf("   Input Vector X: [%.1f, %.1f, %.1f]\n", X[0], X[1], X[2])

	// Perform MVM
	Y, _ := cb.MVM(X)

	// Compare with ideal result
	Yideal := matVecMul(W, X)

	fmt.Println()
	fmt.Println("   ═══════════════════════════════════════════════════")
	fmt.Println("   Computing: Y = W × X  (in single clock cycle!)")
	fmt.Println("   ═══════════════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("   Crossbar Result:  [%.3f, %.3f, %.3f]\n", Y[0], Y[1], Y[2])
	fmt.Printf("   Ideal Result:     [%.3f, %.3f, %.3f]\n", Yideal[0], Yideal[1], Yideal[2])

	// Error analysis
	mse := computeMSE(Y, Yideal)
	fmt.Printf("\n   Mean Squared Error: %.6f (quantization effect)\n", mse)

	// =========================================================================
	// 3. NEURAL NETWORK LAYER EXAMPLE
	// =========================================================================
	fmt.Println("\n\n3. NEURAL NETWORK LAYER SIMULATION")
	fmt.Println("   └─ Simulating first layer of MNIST classifier")
	fmt.Println("   └─ Input: 784 pixels → Output: 128 hidden neurons")
	fmt.Println()

	// Create a larger crossbar for a realistic NN layer
	cbNN := NewCrossbarArray(784, 128, 30)
	cbNN.Noise = 0.01 // 1% device variation

	// Initialize with random weights (simulating trained network)
	WNN := make([][]float64, 784)
	for i := range WNN {
		WNN[i] = make([]float64, 128)
		for j := range WNN[i] {
			// Xavier-like initialization, normalized to [0,1]
			WNN[i][j] = 0.5 + 0.3*rand.NormFloat64()
			if WNN[i][j] < 0 {
				WNN[i][j] = 0
			}
			if WNN[i][j] > 1 {
				WNN[i][j] = 1
			}
		}
	}
	cbNN.SetWeightsFromMatrix(WNN)

	// Simulate input (random "image")
	inputImage := make([]float64, 784)
	for i := range inputImage {
		inputImage[i] = rand.Float64()
	}

	// Perform inference
	hiddenActivations, _ := cbNN.MVM(inputImage)

	fmt.Printf("   Crossbar size: %d × %d = %d cells\n", cbNN.Rows, cbNN.Cols, cbNN.Rows*cbNN.Cols)
	fmt.Printf("   Analog levels: %d (%.1f bits per cell)\n", cbNN.Levels, math.Log2(float64(cbNN.Levels)))
	fmt.Printf("   Total MACs per inference: %d\n", cbNN.Rows*cbNN.Cols)
	fmt.Printf("   Energy per MVM: %.2f pJ\n", cbNN.GetEnergyPerMVM())
	fmt.Printf("   Output activations: %d neurons\n", len(hiddenActivations))

	// Show some output values
	fmt.Println()
	fmt.Println("   Sample outputs (first 8 neurons):")
	fmt.Print("   ")
	for i := 0; i < 8 && i < len(hiddenActivations); i++ {
		fmt.Printf("%.3f  ", hiddenActivations[i])
	}
	fmt.Println("...")

	// =========================================================================
	// 4. COMPARISON: ANALOG VS DIGITAL
	// =========================================================================
	fmt.Println("\n\n4. ANALOG VS DIGITAL COMPARISON")
	fmt.Println()

	compareAnalogDigital()

	// =========================================================================
	// 5. QUANTIZATION EFFECTS
	// =========================================================================
	fmt.Println("\n\n5. QUANTIZATION LEVEL ANALYSIS")
	fmt.Println("   └─ How does precision affect accuracy?")
	fmt.Println()

	analyzeQuantizationLevels()

	// =========================================================================
	// 6. PHYSICS OF ANALOG COMPUTE
	// =========================================================================
	fmt.Println("\n\n6. PHYSICS BEHIND ANALOG COMPUTE")
	explainPhysics()

	fmt.Println("\n═══════════════════════════════════════════════════════════════════")
	fmt.Println("Demo complete! For interactive crossbar visualization:")
	fmt.Println("  fecim-lattice-tools crossbar")
}

func printCrossbarDiagram() {
	diagram := `
      Bit Lines (Output Currents)
           ↓      ↓      ↓
         Col 0  Col 1  Col 2
           │      │      │
    Row 0 ─┼──────┼──────┼── Word Line 0 ← V₀
           │      │      │
    Row 1 ─┼──────┼──────┼── Word Line 1 ← V₁
           │      │      │
    Row 2 ─┼──────┼──────┼── Word Line 2 ← V₂
           │      │      │
           ↓      ↓      ↓
          I₀     I₁     I₂   (Output currents = weighted sums)

    At each junction: FeCIM cell with conductance G_ij
    Output: I_j = Σᵢ Vᵢ × G_ij  (Kirchhoff's Current Law)
`
	fmt.Println(diagram)
}

func printMatrix(M [][]float64) {
	fmt.Printf("      ┌")
	for j := 0; j < len(M[0]); j++ {
		fmt.Printf("      ")
	}
	fmt.Printf("┐\n")

	for i := range M {
		fmt.Printf("      │")
		for j := range M[i] {
			fmt.Printf(" %4.2f ", M[i][j])
		}
		fmt.Printf("│\n")
	}

	fmt.Printf("      └")
	for j := 0; j < len(M[0]); j++ {
		fmt.Printf("      ")
	}
	fmt.Printf("┘\n")
}

func matVecMul(M [][]float64, V []float64) []float64 {
	result := make([]float64, len(M))
	for i := range M {
		for j := range V {
			result[i] += M[i][j] * V[j]
		}
	}
	return result
}

func computeMSE(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.Inf(1)
	}
	var sum float64
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return sum / float64(len(a))
}

func compareAnalogDigital() {
	fmt.Println("   ┌───────────────────┬─────────────────┬─────────────────┐")
	fmt.Println("   │     Metric        │  Digital CMOS   │  FeCIM Crossbar │")
	fmt.Println("   ├───────────────────┼─────────────────┼─────────────────┤")
	fmt.Println("   │ Time Complexity   │     O(N²)       │      O(1)       │")
	fmt.Println("   │ Data Movement     │     High        │      Zero       │")
	fmt.Println("   │ Energy/MAC        │   ~10 pJ        │    ~50 fJ       │")
	fmt.Println("   │ Parallelism       │   Limited       │   Inherent      │")
	fmt.Println("   │ Memory Bandwidth  │   Bottleneck    │   N/A           │")
	fmt.Println("   │ Precision         │   32/64-bit     │   5 bits (30L)  │")
	fmt.Println("   └───────────────────┴─────────────────┴─────────────────┘")
	fmt.Println()
	fmt.Println("   Key Insight: FeCIM eliminates the von Neumann bottleneck!")
	fmt.Println("   → Weights stay in memory, compute happens at storage location")
	fmt.Println("   → ~200× energy reduction for matrix operations")
}

func analyzeQuantizationLevels() {
	levels := []int{2, 4, 8, 16, 30, 64, 128}

	fmt.Println("   Levels │ Bits/Cell │ Ideal MSE │ Energy/MAC (fJ)")
	fmt.Println("   ───────┼───────────┼───────────┼─────────────────")

	for _, L := range levels {
		bits := math.Log2(float64(L))
		// Quantization MSE ≈ 1/(12 × L²) for uniform quantization
		mse := 1.0 / (12.0 * float64(L*L))
		energy := 10.0 * bits // 10 fJ per bit

		marker := ""
		if L == 30 {
			marker = "  ← FeCIM target"
		}

		fmt.Printf("   %4d   │   %5.2f   │  %.2e  │     %6.1f%s\n",
			L, bits, mse, energy, marker)
	}

	fmt.Println()
	fmt.Println("   Trade-off: More levels = better accuracy, but higher energy")
	fmt.Println("   FeCIM's 30 levels balance precision with efficiency")
}

func explainPhysics() {
	fmt.Println()
	fmt.Println("   ┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("   │                    PHYSICS OF ANALOG MVM                        │")
	fmt.Println("   ├─────────────────────────────────────────────────────────────────┤")
	fmt.Println("   │                                                                 │")
	fmt.Println("   │  Ohm's Law at each cell:     I = V × G                          │")
	fmt.Println("   │                                                                 │")
	fmt.Println("   │  Kirchhoff's Current Law:    I_col = Σ I_row                    │")
	fmt.Println("   │                                                                 │")
	fmt.Println("   │  Combining:                  I_j = Σᵢ Vᵢ × G_ij                 │")
	fmt.Println("   │                                                                 │")
	fmt.Println("   │  This IS a dot product!      I_j = V⃗ · G⃗_j                       │")
	fmt.Println("   │                                                                 │")
	fmt.Println("   │  For all columns in parallel: I⃗ = G^T × V⃗  = W × X              │")
	fmt.Println("   │                                                                 │")
	fmt.Println("   │  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━   │")
	fmt.Println("   │  Physics naturally computes matrix-vector multiplication!       │")
	fmt.Println("   │  All operations happen simultaneously via analog signals.       │")
	fmt.Println("   └─────────────────────────────────────────────────────────────────┘")
}

// Helper for init check
func init() {
	for _, arg := range os.Args[1:] {
		if arg == "-h" || arg == "--help" {
			fmt.Println("FeCIM Crossbar MVM Demo - Analog compute-in-memory demonstration")
			fmt.Println()
			fmt.Println("Usage: go run examples/crossbar_mvm.go")
			fmt.Println()
			fmt.Println("This program demonstrates:")
			fmt.Println("  • Crossbar array architecture")
			fmt.Println("  • Matrix-vector multiplication via Kirchhoff's laws")
			fmt.Println("  • Energy efficiency comparison")
			fmt.Println("  • Quantization effects on accuracy")
			fmt.Println()
			fmt.Println("For interactive GUI, run: fecim-lattice-tools crossbar")
			os.Exit(0)
		}
	}
}

func bar(value, max float64, width int) string {
	filled := int(value / max * float64(width))
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}
