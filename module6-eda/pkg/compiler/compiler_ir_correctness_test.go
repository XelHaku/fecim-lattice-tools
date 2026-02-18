// pkg/compiler/compiler_ir_correctness_test.go
package compiler

import (
	"testing"
)

// M6-COMP-01: Net conservation test
// Verify every cell in N×M array is present with valid row/col coordinates
// No missing cells, no duplicate cells, no out-of-bounds indices
func TestIRCorrectness_NetConservation(t *testing.T) {
	testCases := []struct {
		name string
		rows int
		cols int
		arch string
	}{
		{"4x4_Passive", 4, 4, ArchPassive},
		{"8x8_Passive", 8, 8, ArchPassive},
		{"16x16_Passive", 16, 16, ArchPassive},
		{"4x4_1T1R", 4, 4, Arch1T1R},
		{"8x8_1T1R", 8, 8, Arch1T1R},
		{"16x16_1T1R", 16, 16, Arch1T1R},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := NewComputeConfig(tc.rows, tc.cols)
			if tc.arch == Arch1T1R {
				config.With1T1R()
			}

			design, err := GenerateDesign(config)
			if err != nil {
				t.Fatalf("GenerateDesign failed: %v", err)
			}

			// Net conservation check: all cells present
			expectedCells := tc.rows * tc.cols
			if len(design.Cells) != expectedCells {
				t.Errorf("Expected %d cells, got %d", expectedCells, len(design.Cells))
			}

			// Track cell coverage with 2D map
			cellMap := make(map[[2]int]bool)
			for _, cell := range design.Cells {
				// Check bounds
				if cell.Row < 0 || cell.Row >= tc.rows {
					t.Errorf("Cell row %d out of bounds [0, %d)", cell.Row, tc.rows)
				}
				if cell.Col < 0 || cell.Col >= tc.cols {
					t.Errorf("Cell col %d out of bounds [0, %d)", cell.Col, tc.cols)
				}

				// Check for duplicates
				key := [2]int{cell.Row, cell.Col}
				if cellMap[key] {
					t.Errorf("Duplicate cell at (%d, %d)", cell.Row, cell.Col)
				}
				cellMap[key] = true
			}

			// Check all cells covered (no gaps)
			for i := 0; i < tc.rows; i++ {
				for j := 0; j < tc.cols; j++ {
					key := [2]int{i, j}
					if !cellMap[key] {
						t.Errorf("Missing cell at (%d, %d)", i, j)
					}
				}
			}

			// Verify stats consistency
			if design.Stats.TotalCells != expectedCells {
				t.Errorf("Stats.TotalCells = %d, expected %d", design.Stats.TotalCells, expectedCells)
			}

			t.Logf("%s: %d cells verified, net conservation PASS", tc.name, len(design.Cells))
		})
	}
}

// M6-COMP-02: Instance count test
// N×M array → exactly N×M cell instances (parasitic R/C modeled in export, not IR)
func TestIRCorrectness_InstanceCount(t *testing.T) {
	testCases := []struct {
		name string
		rows int
		cols int
		arch string
	}{
		{"4x4_Passive", 4, 4, ArchPassive},
		{"8x8_Passive", 8, 8, ArchPassive},
		{"16x16_Passive", 16, 16, ArchPassive},
		{"4x4_1T1R", 4, 4, Arch1T1R},
		{"8x8_1T1R", 8, 8, Arch1T1R},
		{"16x16_1T1R", 16, 16, Arch1T1R},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := NewComputeConfig(tc.rows, tc.cols)
			if tc.arch == Arch1T1R {
				config.With1T1R()
			}

			design, err := GenerateDesign(config)
			if err != nil {
				t.Fatalf("GenerateDesign failed: %v", err)
			}

			expectedInstances := tc.rows * tc.cols
			actualInstances := len(design.Cells)

			if actualInstances != expectedInstances {
				t.Errorf("Expected %d instances, got %d", expectedInstances, actualInstances)
			}

			// Verify all instances have valid electrical parameters
			for idx, cell := range design.Cells {
				if cell.Conductance <= 0 {
					t.Errorf("Cell %d: invalid conductance %.6e", idx, cell.Conductance)
				}
				if cell.Resistance <= 0 {
					t.Errorf("Cell %d: invalid resistance %.6e", idx, cell.Resistance)
				}
				if cell.ProgramV < 0 {
					t.Errorf("Cell %d: invalid programming voltage %.3f V", idx, cell.ProgramV)
				}
				if cell.Level < 0 || cell.Level >= config.Levels {
					t.Errorf("Cell %d: level %d out of range [0, %d)", idx, cell.Level, config.Levels)
				}
			}

			t.Logf("%s: %d instances, %d parasitic elements (modeled in export)",
				tc.name, actualInstances, 0)
		})
	}
}

// TestIRCorrectness_WithWeights verifies net conservation with weight mapping
func TestIRCorrectness_WithWeights(t *testing.T) {
	testCases := []struct {
		name       string
		weightRows int
		weightCols int
		arrayRows  int
		arrayCols  int
		arch       string
	}{
		{"4x4_weights_in_4x4", 4, 4, 4, 4, ArchPassive},
		{"2x3_weights_in_8x8", 2, 3, 8, 8, ArchPassive},
		{"8x8_weights_in_16x16", 8, 8, 16, 16, Arch1T1R},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate test weights
			weights := make([][]float64, tc.weightRows)
			for i := range weights {
				weights[i] = make([]float64, tc.weightCols)
				for j := range weights[i] {
					weights[i][j] = float64(i*tc.weightCols+j) / float64(tc.weightRows*tc.weightCols)
				}
			}

			config := NewComputeConfig(tc.arrayRows, tc.arrayCols)
			if tc.arch == Arch1T1R {
				config.With1T1R()
			}
			config.WithWeights(weights)

			design, err := GenerateDesign(config)
			if err != nil {
				t.Fatalf("GenerateDesign failed: %v", err)
			}

			// Net conservation: full array populated
			expectedCells := tc.arrayRows * tc.arrayCols
			if len(design.Cells) != expectedCells {
				t.Errorf("Expected %d cells, got %d", expectedCells, len(design.Cells))
			}

			// Verify weight-mapped cells have non-zero initial weights
			weightCellsFound := 0
			for i := 0; i < tc.weightRows; i++ {
				for j := 0; j < tc.weightCols; j++ {
					// Find cell at (i, j)
					found := false
					for _, cell := range design.Cells {
						if cell.Row == i && cell.Col == j {
							found = true
							if cell.InitialWeight != weights[i][j] {
								t.Errorf("Cell (%d, %d): weight %.6f != expected %.6f",
									i, j, cell.InitialWeight, weights[i][j])
							}
							weightCellsFound++
							break
						}
					}
					if !found {
						t.Errorf("Weight cell (%d, %d) not found in design", i, j)
					}
				}
			}

			expectedWeightCells := tc.weightRows * tc.weightCols
			if weightCellsFound != expectedWeightCells {
				t.Errorf("Expected %d weight cells, found %d", expectedWeightCells, weightCellsFound)
			}

			// Verify stats
			if design.Stats.ActiveCells != expectedWeightCells {
				t.Errorf("Stats.ActiveCells = %d, expected %d", design.Stats.ActiveCells, expectedWeightCells)
			}

			t.Logf("%s: %d total cells, %d active weights, net conservation PASS",
				tc.name, len(design.Cells), design.Stats.ActiveCells)
		})
	}
}

// TestIRCorrectness_AreaCalculation verifies area consistency
func TestIRCorrectness_AreaCalculation(t *testing.T) {
	testCases := []struct {
		name      string
		rows      int
		cols      int
		cellPitch float64 // μm
		rowHeight float64 // μm
	}{
		{"4x4_SKY130", 4, 4, 0.46, 2.72},
		{"8x8_SKY130", 8, 8, 0.46, 2.72},
		{"16x16_1T1R", 16, 16, 0.92, 3.40},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := NewComputeConfig(tc.rows, tc.cols)
			config.CellPitch = tc.cellPitch
			config.RowHeight = tc.rowHeight

			design, err := GenerateDesign(config)
			if err != nil {
				t.Fatalf("GenerateDesign failed: %v", err)
			}

			// Expected area: (pitch × height) × N × M in mm²
			cellAreaMM2 := (tc.cellPitch * tc.rowHeight) * 1e-6
			expectedAreaMM2 := cellAreaMM2 * float64(tc.rows*tc.cols)

			if design.Stats.AreaMM2 != expectedAreaMM2 {
				t.Errorf("Area mismatch: got %.6e mm², expected %.6e mm²",
					design.Stats.AreaMM2, expectedAreaMM2)
			}

			t.Logf("%s: Area = %.6e mm² (cell: %.2f×%.2f μm)",
				tc.name, design.Stats.AreaMM2, tc.cellPitch, tc.rowHeight)
		})
	}
}
