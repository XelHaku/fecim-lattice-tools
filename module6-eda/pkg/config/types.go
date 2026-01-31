// pkg/config/types.go
// Configuration types for FeCIM Design Suite Module 6
//
// References:
// [1] SkyWater SKY130 PDK: https://skywater-pdk.readthedocs.io/
//     Cell dimensions based on SKY130 unithd site (0.46 × 2.72 μm)
package config

// CellConfig defines configuration for a single FeCIM bitcell
// Used for generating LEF, Liberty, and Verilog files
type CellConfig struct {
	Name         string  // Cell name, e.g., "fecim_bitcell"
	Width        float64 // Cell width in μm (SKY130 unithd site width: 0.46 μm)
	Height       float64 // Cell height in μm (SKY130 standard cell height: 2.72 μm)
	CellType     string  // "passive" (0T1R) or "1t1r" (with selector)
	Technology   string  // Target PDK, e.g., "sky130"
	
	// ⚠️ PLACEHOLDER TIMING VALUES
	// These are estimates requiring FeFET characterization via SPICE simulation
	// Real values need: SPICE compact model + Liberty characterization
	RiseTime     float64 // Rise time in ns (PLACEHOLDER: 10.0 ns, realistic for HfO₂ FeFET)
	FallTime     float64 // Fall time in ns (PLACEHOLDER: 10.0 ns, realistic for HfO₂ FeFET)
	InputCap     float64 // Input capacitance in pF (PLACEHOLDER: 0.015 pF, mid-range for FeFET cell)
	LeakagePower float64 // Leakage power in nW (PLACEHOLDER: 0.0003 nW, matches published 30nm NC-FinFET)
}

// ArrayConfig defines configuration for a FeCIM crossbar array
type ArrayConfig struct {
	Rows         int     // Number of rows (e.g., 4, 8, 16, 32)
	Cols         int     // Number of columns
	Mode         string  // "storage", "memory", or "compute"
	Architecture string  // "passive" or "1t1r"
	Technology   string  // e.g., "sky130"
	CellWidth    float64 // From CellConfig, in μm
	CellHeight   float64 // From CellConfig, in μm
}

// DefaultCellConfig returns a default cell configuration for FeCIM bitcell
// Dimensions based on SKY130 PDK unithd site [Ref 1]
// All timing values are PLACEHOLDERS requiring characterization
func DefaultCellConfig() CellConfig {
	return CellConfig{
		Name:         "fecim_bitcell",
		Width:        0.46,  // μm (SKY130 unithd site width)
		Height:       2.72,  // μm (SKY130 standard cell height)
		CellType:     "passive",
		Technology:   "sky130",
		// PLACEHOLDER timing values
		RiseTime:     10.0,   // ns (realistic for HfO₂ FeFET at standard voltages)
		FallTime:     10.0,   // ns (realistic for HfO₂ FeFET at standard voltages)
		InputCap:     0.015,  // pF (mid-range for FeFET cell with ferroelectric capacitor)
		LeakagePower: 0.0003, // nW (matches published 30nm NC-FinFET measurements)
	}
}

// DefaultArrayConfig returns a default array configuration
// Starts with minimal 4×4 array for testing
func DefaultArrayConfig() ArrayConfig {
	return ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "storage",    // "storage", "memory", or "compute"
		Architecture: "passive",    // "passive" (0T1R) or "1t1r"
		Technology:   "sky130",
		CellWidth:    0.46,  // μm (from DefaultCellConfig)
		CellHeight:   2.72,  // μm (from DefaultCellConfig)
	}
}
