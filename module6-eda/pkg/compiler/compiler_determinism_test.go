// pkg/compiler/compiler_determinism_test.go
package compiler

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"
)

// M6-COMP-05: Determinism test
// Verify same input → bit-identical output
// Run compilation 10 times, hash netlists, verify identical

// TestDeterminism_IRGeneration verifies IR generation is deterministic
func TestDeterminism_IRGeneration(t *testing.T) {
	testCases := []struct {
		name string
		rows int
		cols int
		arch string
	}{
		{"4x4_Passive", 4, 4, ArchPassive},
		{"8x8_1T1R", 8, 8, Arch1T1R},
		{"16x16_2T1R", 16, 16, Arch2T1R},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := NewComputeConfig(tc.rows, tc.cols)

			switch tc.arch {
			case Arch1T1R:
				config.With1T1R()
			case Arch2T1R:
				config.With2T1R()
			case ArchPassive:
				config.Architecture = ArchPassive
			}

			// Run compilation 10 times
			const iterations = 10
			hashes := make([]string, iterations)

			for i := 0; i < iterations; i++ {
				design, err := GenerateDesign(config)
				if err != nil {
					t.Fatalf("Iteration %d: GenerateDesign failed: %v", i, err)
				}

				// Serialize design to JSON for hashing
				jsonData, err := json.Marshal(design)
				if err != nil {
					t.Fatalf("Iteration %d: JSON marshal failed: %v", i, err)
				}

				// Compute SHA256 hash
				hash := sha256.Sum256(jsonData)
				hashes[i] = fmt.Sprintf("%x", hash)
			}

			// Verify all hashes identical
			referenceHash := hashes[0]
			for i := 1; i < iterations; i++ {
				if hashes[i] != referenceHash {
					t.Errorf("Iteration %d: hash mismatch\n  Reference: %s\n  Got:       %s",
						i, referenceHash, hashes[i])
				}
			}

			t.Logf("%s: %d iterations → identical hash: %s", tc.name, iterations, referenceHash[:16])
		})
	}
}

// TestDeterminism_WithWeights verifies determinism with weight mapping
func TestDeterminism_WithWeights(t *testing.T) {
	// Fixed seed weights for reproducibility
	weights := [][]float64{
		{0.1, -0.2, 0.3, -0.4, 0.5},
		{-0.6, 0.7, -0.8, 0.9, -1.0},
		{0.15, -0.25, 0.35, -0.45, 0.55},
		{-0.65, 0.75, -0.85, 0.95, 0.05},
	}

	config := NewComputeConfig(8, 8)
	config.Levels = 30
	config.WithWeights(weights)

	const iterations = 10
	hashes := make([]string, iterations)

	for i := 0; i < iterations; i++ {
		design, err := GenerateDesign(config)
		if err != nil {
			t.Fatalf("Iteration %d: GenerateDesign failed: %v", i, err)
		}

		jsonData, err := json.Marshal(design)
		if err != nil {
			t.Fatalf("Iteration %d: JSON marshal failed: %v", i, err)
		}

		hash := sha256.Sum256(jsonData)
		hashes[i] = fmt.Sprintf("%x", hash)
	}

	// Verify all hashes identical
	referenceHash := hashes[0]
	allIdentical := true
	for i := 1; i < iterations; i++ {
		if hashes[i] != referenceHash {
			t.Errorf("Iteration %d: hash mismatch", i)
			allIdentical = false
		}
	}

	if allIdentical {
		t.Logf("Weight mapping: %d iterations → identical hash: %s", iterations, referenceHash[:16])
	}
}

// Note: Export-level determinism tests (SPICE, Verilog, DEF, LEF, Liberty)
// are located in module6-eda/pkg/export/*_test.go to avoid import cycles.

// TestDeterminism_CellOrdering verifies cell array ordering is deterministic
func TestDeterminism_CellOrdering(t *testing.T) {
	config := NewComputeConfig(8, 8)
	config.Levels = 30

	// Weights to test ordering stability
	weights := make([][]float64, 8)
	for i := range weights {
		weights[i] = make([]float64, 8)
		for j := range weights[i] {
			weights[i][j] = float64(i*8+j)/64.0*2.0 - 1.0 // [-1, 1]
		}
	}
	config.WithWeights(weights)

	const iterations = 10
	cellOrderings := make([]string, iterations)

	for iter := 0; iter < iterations; iter++ {
		design, err := GenerateDesign(config)
		if err != nil {
			t.Fatalf("Iteration %d: GenerateDesign failed: %v", iter, err)
		}

		// Build cell ordering signature
		var ordering string
		for _, cell := range design.Cells {
			ordering += fmt.Sprintf("(%d,%d,%d)", cell.Row, cell.Col, cell.Level)
		}

		cellOrderings[iter] = ordering
	}

	// Verify cell ordering identical across iterations
	referenceOrdering := cellOrderings[0]
	for i := 1; i < iterations; i++ {
		if cellOrderings[i] != referenceOrdering {
			t.Errorf("Iteration %d: cell ordering differs from reference", i)
		}
	}

	t.Logf("Cell ordering: %d iterations → identical ordering", iterations)
}

// TestDeterminism_QuantizationStability verifies quantization is deterministic
func TestDeterminism_QuantizationStability(t *testing.T) {
	weights := [][]float64{
		{0.123456789, -0.987654321, 0.555555555},
		{-0.111111111, 0.999999999, -0.333333333},
	}

	config := NewComputeConfig(4, 4)
	config.Levels = 30
	config.WithWeights(weights)

	const iterations = 10
	quantLevels := make([][]int, iterations)

	for iter := 0; iter < iterations; iter++ {
		design, err := GenerateDesign(config)
		if err != nil {
			t.Fatalf("Iteration %d: GenerateDesign failed: %v", iter, err)
		}

		// Extract quantization levels for weight cells
		levels := make([]int, 6) // 2×3 weights
		idx := 0
		for i := 0; i < 2; i++ {
			for j := 0; j < 3; j++ {
				for _, cell := range design.Cells {
					if cell.Row == i && cell.Col == j {
						levels[idx] = cell.Level
						idx++
						break
					}
				}
			}
		}
		quantLevels[iter] = levels
	}

	// Verify quantization levels identical
	reference := quantLevels[0]
	for iter := 1; iter < iterations; iter++ {
		for i := range reference {
			if quantLevels[iter][i] != reference[i] {
				t.Errorf("Iteration %d: quantization level mismatch at index %d: got %d, expected %d",
					iter, i, quantLevels[iter][i], reference[i])
			}
		}
	}

	t.Logf("Quantization: %d iterations → identical levels: %v", iterations, reference)
}
