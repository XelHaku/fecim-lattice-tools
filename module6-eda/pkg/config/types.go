// pkg/config/types.go
// Configuration types for FeCIM Design Suite Module 6
//
// References:
// [1] SkyWater SKY130 PDK: https://skywater-pdk.readthedocs.io/
//
//	Cell dimensions based on SKY130 unithd site (0.46 × 2.72 μm)
package config

// CellConfig defines configuration for a single FeCIM bitcell
// Used for generating LEF, Liberty, and Verilog files
type CellConfig struct {
	Name       string  // Cell name, e.g., "fecim_bitcell"
	Width      float64 // Cell width in μm (SKY130 unithd site width: 0.46 μm)
	Height     float64 // Cell height in μm (SKY130 standard cell height: 2.72 μm)
	CellType   string  // "passive" (0T1R) or "1t1r" (with selector)
	Technology string  // Target PDK, e.g., "sky130"

	// Operating conditions (for Liberty file generation)
	Voltage     float64 // Operating voltage in V (default: 1.8V for SKY130)
	Temperature float64 // Operating temperature in °C (default: 25°C typical corner)
	Process     float64 // Process corner (1.0 = typical, <1.0 = fast, >1.0 = slow)

	// Metal layer parameters (for LEF file generation)
	MetalPitch float64 // Metal1 pitch in μm (default: 0.46 for SKY130)
	MetalWidth float64 // Metal1 minimum width in μm (default: 0.14 for SKY130)

	// Published FeFET timing/power anchors for Liberty defaults
	// - Trentzsch et al., IEDM 2016 (28nm FDSOI FeFET)
	// - Muller et al., IEEE TED 2013 (NC-FinFET low leakage)
	RiseTime     float64 // Rise time in ns (write path typical: 50 ns)
	FallTime     float64 // Fall time in ns (read path typical: 5 ns)
	InputCap     float64 // Input capacitance in pF (typical FeFET mid-range: 0.015 pF)
	LeakagePower float64 // Leakage power in nW (published low-leakage envelope: 0.0003 nW)
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
// Timing defaults use published FeFET anchors (Trentzsch et al., IEDM 2016; Muller et al., IEEE TED 2013)
func DefaultCellConfig() CellConfig {
	return CellConfig{
		Name:       "fecim_bitcell",
		Width:      0.46, // μm (SKY130 unithd site width)
		Height:     2.72, // μm (SKY130 standard cell height)
		CellType:   "passive",
		Technology: "sky130",
		// Operating conditions (SKY130 typical corner)
		Voltage:     1.8,  // V (SKY130 nominal VDD)
		Temperature: 25.0, // °C (typical corner)
		Process:     1.0,  // Typical process corner
		// Metal layer parameters (SKY130 met1)
		MetalPitch: 0.46, // μm (SKY130 met1 pitch)
		MetalWidth: 0.14, // μm (SKY130 met1 minimum width)
		// Published FeFET defaults (Trentzsch 2016 / Muller 2013)
		RiseTime:     50.0,   // ns (write / cell_rise, typical)
		FallTime:     5.0,    // ns (read / cell_fall, typical)
		InputCap:     0.015,  // pF (mid-range FeFET)
		LeakagePower: 0.0003, // nW (published NC-FinFET low-leakage envelope)
	}
}

// DefaultArrayConfig returns a default array configuration
// Starts with minimal 4×4 array for testing
func DefaultArrayConfig() ArrayConfig {
	return ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "storage", // "storage", "memory", or "compute"
		Architecture: "passive", // "passive" (0T1R) or "1t1r"
		Technology:   "sky130",
		CellWidth:    0.46, // μm (from DefaultCellConfig)
		CellHeight:   2.72, // μm (from DefaultCellConfig)
	}
}
