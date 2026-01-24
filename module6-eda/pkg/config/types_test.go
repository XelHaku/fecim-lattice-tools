// pkg/config/types_test.go
package config

import (
	"testing"
)

func TestDefaultCellConfig(t *testing.T) {
	cfg := DefaultCellConfig()

	// Test basic properties
	if cfg.Name != "fecim_bitcell" {
		t.Errorf("Expected name 'fecim_bitcell', got '%s'", cfg.Name)
	}

	// Test SKY130 dimensions
	if cfg.Width != 0.46 {
		t.Errorf("Expected width 0.46 μm (SKY130 unithd), got %.3f", cfg.Width)
	}
	if cfg.Height != 2.72 {
		t.Errorf("Expected height 2.72 μm (SKY130 std cell), got %.3f", cfg.Height)
	}

	// Test default cell type
	if cfg.CellType != "passive" {
		t.Errorf("Expected cell type 'passive', got '%s'", cfg.CellType)
	}

	// Test technology
	if cfg.Technology != "sky130" {
		t.Errorf("Expected technology 'sky130', got '%s'", cfg.Technology)
	}

	// Test placeholder timing values exist (not testing exact values as they're placeholders)
	if cfg.RiseTime <= 0 {
		t.Errorf("RiseTime should be positive, got %.3f", cfg.RiseTime)
	}
	if cfg.FallTime <= 0 {
		t.Errorf("FallTime should be positive, got %.3f", cfg.FallTime)
	}
	if cfg.InputCap <= 0 {
		t.Errorf("InputCap should be positive, got %.6f", cfg.InputCap)
	}
	if cfg.LeakagePower <= 0 {
		t.Errorf("LeakagePower should be positive, got %.6f", cfg.LeakagePower)
	}
}

func TestDefaultArrayConfig(t *testing.T) {
	cfg := DefaultArrayConfig()

	// Test default dimensions
	if cfg.Rows != 4 {
		t.Errorf("Expected 4 rows, got %d", cfg.Rows)
	}
	if cfg.Cols != 4 {
		t.Errorf("Expected 4 columns, got %d", cfg.Cols)
	}

	// Test mode
	if cfg.Mode != "storage" {
		t.Errorf("Expected mode 'storage', got '%s'", cfg.Mode)
	}

	// Test architecture
	if cfg.Architecture != "passive" {
		t.Errorf("Expected architecture 'passive', got '%s'", cfg.Architecture)
	}

	// Test technology
	if cfg.Technology != "sky130" {
		t.Errorf("Expected technology 'sky130', got '%s'", cfg.Technology)
	}

	// Test cell dimensions match DefaultCellConfig
	defaultCell := DefaultCellConfig()
	if cfg.CellWidth != defaultCell.Width {
		t.Errorf("CellWidth mismatch: expected %.3f, got %.3f", defaultCell.Width, cfg.CellWidth)
	}
	if cfg.CellHeight != defaultCell.Height {
		t.Errorf("CellHeight mismatch: expected %.3f, got %.3f", defaultCell.Height, cfg.CellHeight)
	}
}

func TestCellConfigCustomization(t *testing.T) {
	cfg := CellConfig{
		Name:         "custom_fecim",
		Width:        1.0,
		Height:       3.0,
		CellType:     "1t1r",
		Technology:   "sky130",
		RiseTime:     0.2,
		FallTime:     0.2,
		InputCap:     0.005,
		LeakagePower: 0.002,
	}

	if cfg.Name != "custom_fecim" {
		t.Errorf("Custom name not set correctly")
	}
	if cfg.CellType != "1t1r" {
		t.Errorf("Custom cell type not set correctly")
	}
}

func TestArrayConfigScaling(t *testing.T) {
	testCases := []struct {
		rows int
		cols int
		name string
	}{
		{4, 4, "4x4 array"},
		{8, 8, "8x8 array"},
		{16, 16, "16x16 array"},
		{32, 32, "32x32 array"},
		{64, 64, "64x64 array"},
		{128, 128, "128x128 array"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ArrayConfig{
				Rows:         tc.rows,
				Cols:         tc.cols,
				Mode:         "compute",
				Architecture: "1t1r",
				Technology:   "sky130",
				CellWidth:    0.46,
				CellHeight:   2.72,
			}

			if cfg.Rows != tc.rows {
				t.Errorf("Expected %d rows, got %d", tc.rows, cfg.Rows)
			}
			if cfg.Cols != tc.cols {
				t.Errorf("Expected %d cols, got %d", tc.cols, cfg.Cols)
			}

			// Calculate expected total cells
			expectedCells := tc.rows * tc.cols
			actualCells := cfg.Rows * cfg.Cols
			if actualCells != expectedCells {
				t.Errorf("Expected %d total cells, got %d", expectedCells, actualCells)
			}
		})
	}
}
