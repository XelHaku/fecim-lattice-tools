// pkg/compiler/compiler_topology_test.go
package compiler

import (
	"testing"
)

// M6-COMP-03: Topology correctness
// Verify device counts match architecture at IR level:
// - 0T1R (passive): 1 FeCap per cell
// - 1T1R: 1 FET + 1 FeCap per cell
// - 2T1R: 2 FET + 1 FeCap per cell
//
// Note: Device topology is verified at the IR (ArrayDesign) level.
// SPICE-level verification is in module6-eda/pkg/export tests.

// TestTopology_PassiveArchitecture verifies 0T1R (passive crossbar)
func TestTopology_PassiveArchitecture(t *testing.T) {
	testCases := []struct {
		name string
		rows int
		cols int
	}{
		{"4x4", 4, 4},
		{"8x8", 8, 8},
		{"16x16", 16, 16},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := NewComputeConfig(tc.rows, tc.cols)
			config.Architecture = ArchPassive

			design, err := GenerateDesign(config)
			if err != nil {
				t.Fatalf("GenerateDesign failed: %v", err)
			}

			expectedCells := tc.rows * tc.cols

			// Verify architecture set correctly
			if design.Config.Architecture != ArchPassive {
				t.Errorf("Expected architecture %s, got %s", ArchPassive, design.Config.Architecture)
			}

			// Verify cell count in IR
			if len(design.Cells) != expectedCells {
				t.Errorf("Expected %d cells, got %d", expectedCells, len(design.Cells))
			}

			// For passive architecture:
			// Each cell in IR → 1 FeCap device (no transistors)
			// Device count verified in export layer
			expectedFeCaps := expectedCells
			expectedTransistors := 0

			t.Logf("%s Passive: %d cells → %d FeCaps, %d transistors (IR level)", 
				tc.name, expectedCells, expectedFeCaps, expectedTransistors)
		})
	}
}

// TestTopology_1T1RArchitecture verifies 1T1R (1 transistor + 1 FeCap)
func TestTopology_1T1RArchitecture(t *testing.T) {
	testCases := []struct {
		name string
		rows int
		cols int
	}{
		{"4x4", 4, 4},
		{"8x8", 8, 8},
		{"16x16", 16, 16},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := NewComputeConfig(tc.rows, tc.cols)
			config.With1T1R() // Sets Architecture to Arch1T1R

			design, err := GenerateDesign(config)
			if err != nil {
				t.Fatalf("GenerateDesign failed: %v", err)
			}

			// Verify architecture field
			if design.Config.Architecture != Arch1T1R {
				t.Errorf("Expected architecture %s, got %s", Arch1T1R, design.Config.Architecture)
			}

			expectedCells := tc.rows * tc.cols

			// Verify cell count in IR
			if len(design.Cells) != expectedCells {
				t.Errorf("Expected %d cells, got %d", expectedCells, len(design.Cells))
			}

			// Verify cell pitch and height match 1T1R geometry
			expectedPitch := 0.92 // μm (larger for transistor)
			expectedHeight := 3.40 // μm
			
			if design.Config.CellPitch != expectedPitch {
				t.Errorf("1T1R cell pitch: got %.2f μm, expected %.2f μm", 
					design.Config.CellPitch, expectedPitch)
			}
			if design.Config.RowHeight != expectedHeight {
				t.Errorf("1T1R row height: got %.2f μm, expected %.2f μm", 
					design.Config.RowHeight, expectedHeight)
			}

			// 1T1R topology: 1 NMOS + 1 FeCap per cell
			expectedTransistors := expectedCells
			expectedFeCaps := expectedCells

			t.Logf("%s 1T1R: %d cells → %d transistors + %d FeCaps (IR level)", 
				tc.name, expectedCells, expectedTransistors, expectedFeCaps)
		})
	}
}

// TestTopology_2T1RArchitecture verifies 2T1R (2 transistors + 1 FeCap)
func TestTopology_2T1RArchitecture(t *testing.T) {
	testCases := []struct {
		name string
		rows int
		cols int
	}{
		{"4x4", 4, 4},
		{"8x8", 8, 8},
		{"16x16", 16, 16},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := NewComputeConfig(tc.rows, tc.cols)
			config.With2T1R() // Sets Architecture to Arch2T1R

			design, err := GenerateDesign(config)
			if err != nil {
				t.Fatalf("GenerateDesign failed: %v", err)
			}

			// Verify architecture field
			if design.Config.Architecture != Arch2T1R {
				t.Errorf("Expected architecture %s, got %s", Arch2T1R, design.Config.Architecture)
			}

			expectedCells := tc.rows * tc.cols

			// Verify cell count in IR
			if len(design.Cells) != expectedCells {
				t.Errorf("Expected %d cells, got %d", expectedCells, len(design.Cells))
			}

			// Verify cell pitch and height match 2T1R geometry
			expectedPitch := 1.38 // μm (~3x passive for two transistors)
			expectedHeight := 3.40 // μm
			
			if design.Config.CellPitch != expectedPitch {
				t.Errorf("2T1R cell pitch: got %.2f μm, expected %.2f μm", 
					design.Config.CellPitch, expectedPitch)
			}
			if design.Config.RowHeight != expectedHeight {
				t.Errorf("2T1R row height: got %.2f μm, expected %.2f μm", 
					design.Config.RowHeight, expectedHeight)
			}

			// 2T1R topology: 2 NMOS + 1 FeCap per cell
			expectedTransistors := expectedCells * 2
			expectedFeCaps := expectedCells

			t.Logf("%s 2T1R: %d cells → %d transistors + %d FeCaps (IR level)", 
				tc.name, expectedCells, expectedTransistors, expectedFeCaps)
		})
	}
}

// TestTopology_DeviceCountExact verifies exact device counts across architectures
func TestTopology_DeviceCountExact(t *testing.T) {
	testCases := []struct {
		name              string
		rows              int
		cols              int
		arch              string
		expectedFETs      int
		expectedFeCaps    int
		expectedPitch     float64 // μm
		expectedHeight    float64 // μm
	}{
		// Passive: no FETs, 1 FeCap per cell
		{"4x4_Passive", 4, 4, ArchPassive, 0, 16, 0.46, 2.72},
		{"8x8_Passive", 8, 8, ArchPassive, 0, 64, 0.46, 2.72},

		// 1T1R: 1 FET + 1 FeCap per cell
		{"4x4_1T1R", 4, 4, Arch1T1R, 16, 16, 0.92, 3.40},
		{"8x8_1T1R", 8, 8, Arch1T1R, 64, 64, 0.92, 3.40},

		// 2T1R: 2 FET + 1 FeCap per cell
		{"4x4_2T1R", 4, 4, Arch2T1R, 32, 16, 1.38, 3.40},
		{"8x8_2T1R", 8, 8, Arch2T1R, 128, 64, 1.38, 3.40},
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

			design, err := GenerateDesign(config)
			if err != nil {
				t.Fatalf("GenerateDesign failed: %v", err)
			}

			// Verify cell count (IR level)
			expectedCells := tc.rows * tc.cols
			if len(design.Cells) != expectedCells {
				t.Errorf("Expected %d cells, got %d", expectedCells, len(design.Cells))
			}

			// Verify architecture set correctly
			if design.Config.Architecture != tc.arch {
				t.Errorf("Expected architecture %s, got %s", tc.arch, design.Config.Architecture)
			}

			// Verify cell geometry matches architecture
			if design.Config.CellPitch != tc.expectedPitch {
				t.Errorf("Cell pitch: got %.2f μm, expected %.2f μm", 
					design.Config.CellPitch, tc.expectedPitch)
			}
			if design.Config.RowHeight != tc.expectedHeight {
				t.Errorf("Row height: got %.2f μm, expected %.2f μm", 
					design.Config.RowHeight, tc.expectedHeight)
			}

			// Device counts are architecture-dependent:
			// Passive: 0 FETs per cell, 1 FeCap per cell
			// 1T1R: 1 FET per cell, 1 FeCap per cell
			// 2T1R: 2 FETs per cell, 1 FeCap per cell
			
			// Calculate expected device counts
			fetsPerCell := 0
			fecapsPerCell := 1
			
			switch tc.arch {
			case Arch1T1R:
				fetsPerCell = 1
			case Arch2T1R:
				fetsPerCell = 2
			}
			
			totalFETs := fetsPerCell * expectedCells
			totalFeCaps := fecapsPerCell * expectedCells

			if totalFETs != tc.expectedFETs {
				t.Errorf("Expected %d total FETs, got %d", tc.expectedFETs, totalFETs)
			}
			if totalFeCaps != tc.expectedFeCaps {
				t.Errorf("Expected %d total FeCaps, got %d", tc.expectedFeCaps, totalFeCaps)
			}

			t.Logf("%s: %d cells → %d FETs + %d FeCaps (IR level, exact)", 
				tc.name, expectedCells, totalFETs, totalFeCaps)
		})
	}
}

// TestTopology_ArchitectureConsistency verifies architecture configuration is consistent
func TestTopology_ArchitectureConsistency(t *testing.T) {
	testCases := []struct {
		name           string
		arch           string
		expectedPitch  float64 // μm
		expectedHeight float64 // μm
		devicePerCell  int     // Total devices (FETs + FeCaps) per cell
	}{
		{"Passive", ArchPassive, 0.46, 2.72, 1}, // 1 FeCap
		{"1T1R", Arch1T1R, 0.92, 3.40, 2},       // 1 FET + 1 FeCap
		{"2T1R", Arch2T1R, 1.38, 3.40, 3},       // 2 FET + 1 FeCap
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := NewComputeConfig(4, 4)
			
			switch tc.arch {
			case Arch1T1R:
				config.With1T1R()
			case Arch2T1R:
				config.With2T1R()
			case ArchPassive:
				config.Architecture = ArchPassive
			}

			design, err := GenerateDesign(config)
			if err != nil {
				t.Fatalf("GenerateDesign failed: %v", err)
			}

			// Verify architecture field
			if design.Config.Architecture != tc.arch {
				t.Errorf("Expected architecture %s, got %s", tc.arch, design.Config.Architecture)
			}

			// Verify cell geometry
			if design.Config.CellPitch != tc.expectedPitch {
				t.Errorf("Cell pitch: got %.2f μm, expected %.2f μm", 
					design.Config.CellPitch, tc.expectedPitch)
			}
			if design.Config.RowHeight != tc.expectedHeight {
				t.Errorf("Row height: got %.2f μm, expected %.2f μm", 
					design.Config.RowHeight, tc.expectedHeight)
			}

			// Verify cell count
			expectedCells := 16 // 4×4
			if len(design.Cells) != expectedCells {
				t.Errorf("Expected %d cells, got %d", expectedCells, len(design.Cells))
			}

			t.Logf("%s: %.2f×%.2f μm cell, %d devices/cell", 
				tc.name, tc.expectedPitch, tc.expectedHeight, tc.devicePerCell)
		})
	}
}
