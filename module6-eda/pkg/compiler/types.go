// Package compiler generates FeCIM chip designs for fabrication.
//
// This package supports three distinct operation modes for FeCIM technology:
//
// # Operation Modes
//
// 1. Storage Mode (NAND Replacement):
// High-density non-volatile storage using 30 conductance levels per cell.
// Optimizes for retention time and endurance. No compute functionality.
//
// 2. Memory Mode (DRAM Replacement):
// High-speed, zero-refresh memory with 10ns access times.
// Optimizes for speed and reliability. No compute functionality.
//
// 3. Compute Mode (AI Accelerator):
// Analog compute-in-memory for neural network inference.
// Can optionally initialize cells with pre-trained weights.
//
// # Design Flow
//
// The design process:
//  1. Configure array parameters (mode, size, technology)
//  2. Set peripheral requirements (DAC/ADC bits, TIA gain)
//  3. Generate physical design files (Verilog, DEF, SPICE)
//  4. Validate with simulation
//  5. Export for fabrication (GDS via OpenLane)
//
// # Usage Examples
//
// Storage chip (no weights needed):
//
//	config := compiler.NewArrayConfig(compiler.ModeStorage, 256, 256)
//	design, err := compiler.GenerateDesign(config)
//
// Memory chip (no weights needed):
//
//	config := compiler.NewArrayConfig(compiler.ModeMemory, 128, 128)
//	design, err := compiler.GenerateDesign(config)
//
// Compute chip (weights optional):
//
//	config := compiler.NewArrayConfig(compiler.ModeCompute, 64, 64)
//	config.ComputeConfig.InitialWeights = weights // optional
//	design, err := compiler.GenerateDesign(config)
//
// # OpenLane Integration
//
// Generated DEF files use FIXED placement keywords for OpenLane integration.
// Set PLACEMENT_CURRENT_DEF and PL_SKIP_INITIAL_PLACEMENT=1 in config.
package compiler

// OperationMode defines the target application for the FeCIM array
type OperationMode int

const (
	// ModeStorage creates high-density non-volatile storage (NAND replacement)
	// - Optimizes for data retention and endurance
	// - 30 levels per cell = ~4.9 bits/cell
	// - No compute functionality
	ModeStorage OperationMode = iota

	// ModeMemory creates high-speed memory (DRAM replacement)
	// - Optimizes for access speed (~10ns)
	// - Zero refresh required (non-volatile)
	// - No compute functionality
	ModeMemory

	// ModeCompute creates AI accelerator arrays (GPU/TPU replacement)
	// - Enables analog matrix-vector multiply
	// - Can optionally initialize with trained weights
	// - Supports inference operations
	ModeCompute
)

// String returns the human-readable name of the operation mode
func (m OperationMode) String() string {
	switch m {
	case ModeStorage:
		return "Storage"
	case ModeMemory:
		return "Memory"
	case ModeCompute:
		return "Compute"
	default:
		return "Unknown"
	}
}

// Architecture types for crossbar array
const (
	ArchPassive = "passive" // Passive crossbar (WL, BL only)
	Arch1T1R    = "1t1r"    // 1 Transistor 1 Resistor (WL, BL, SL)
	Arch2T1R    = "2t1r"    // 2 Transistor 1 Resistor (WL, BL, SL, CSL)
)

// Technology nodes supported
const (
	TechSKY130 = "SKY130"     // SkyWater 130nm (open source)
	TechGF180  = "GF180MCU"   // GlobalFoundries 180nm (open source)
	TechIHP    = "IHP_SG13G2" // IHP 130nm SiGe BiCMOS
)

// SKY130 physical constants for cell layout.
// Ref: SkyWater SKY130 PDK — unithd standard cell site.
const (
	sky130CellPitch = 0.46  // µm — met1 pitch / unithd site width
	sky130RowHeight = 2.72  // µm — standard cell row height
	sky130VDD       = 1.8   // V — nominal supply voltage
	sky130ClockMHz  = 100.0 // MHz — default operating frequency

	// 1T1R cell dimensions (wider for transistor)
	cell1T1RPitch  = 0.92 // µm — ~2x passive for selector transistor
	cell1T1RHeight = 3.40 // µm — taller for transistor + FeFET stack

	// 2T1R cell dimensions (widest for dual transistors)
	cell2T1RPitch = 1.38 // µm — ~3x passive for two transistors
)

// ArrayConfig holds all parameters for FeCIM array design
type ArrayConfig struct {
	// Basic array parameters
	Name      string        `json:"name"`       // Design name
	Mode      OperationMode `json:"mode"`       // Storage, Memory, or Compute
	ArrayRows int           `json:"array_rows"` // Number of rows
	ArrayCols int           `json:"array_cols"` // Number of columns

	// Technology selection
	Technology   string `json:"technology"`   // SKY130, GF180MCU, IHP_SG13G2
	Architecture string `json:"architecture"` // "passive" or "1T1R"

	// Physical parameters
	CellPitch float64 `json:"cell_pitch"` // Cell width in microns
	RowHeight float64 `json:"row_height"` // Cell height in microns
	Levels    int     `json:"levels"`     // Conductance levels (default: 30)

	// Electrical parameters
	GMin     float64 `json:"g_min"`      // Min conductance (μS)
	GMax     float64 `json:"g_max"`      // Max conductance (μS)
	VProgMin float64 `json:"v_prog_min"` // Min programming voltage (V)
	VProgMax float64 `json:"v_prog_max"` // Max programming voltage (V)
	TPulse   float64 `json:"t_pulse"`    // Programming pulse width (ns)

	// Peripheral configuration
	Peripherals PeripheralConfig `json:"peripherals"`

	// Mode-specific configurations
	StorageConfig *StorageArrayConfig `json:"storage_config,omitempty"` // Only for ModeStorage
	MemoryConfig  *MemoryArrayConfig  `json:"memory_config,omitempty"`  // Only for ModeMemory
	ComputeConfig *ComputeArrayConfig `json:"compute_config,omitempty"` // Only for ModeCompute
}

// PeripheralConfig defines DAC/ADC/TIA parameters
type PeripheralConfig struct {
	DACBits   int     `json:"dac_bits"`   // DAC resolution (default: 8)
	ADCBits   int     `json:"adc_bits"`   // ADC resolution (default: 8)
	TIAGain   float64 `json:"tia_gain"`   // Transimpedance gain (Ω)
	VDD       float64 `json:"vdd"`        // Supply voltage
	ClockFreq float64 `json:"clock_freq"` // Operating frequency (MHz)
}

// StorageArrayConfig holds storage-mode specific parameters
type StorageArrayConfig struct {
	RetentionYears  float64 `json:"retention_years"`  // Target data retention
	EnduranceCycles int     `json:"endurance_cycles"` // Write endurance
	ErrorCorrection string  `json:"error_correction"` // ECC type: "none", "SECDED", "BCH"
}

// MemoryArrayConfig holds memory-mode specific parameters
type MemoryArrayConfig struct {
	AccessTimeNs  float64 `json:"access_time_ns"`  // Target read access time
	WriteTimeNs   float64 `json:"write_time_ns"`   // Target write time
	BandwidthGBps float64 `json:"bandwidth_gbps"`  // Target bandwidth
	PowerBudgetMW float64 `json:"power_budget_mw"` // Max power consumption
}

// ComputeArrayConfig holds compute-mode specific parameters
type ComputeArrayConfig struct {
	// Optional: Pre-trained weights to initialize the array
	// If nil, array is generated without initial programming
	InitialWeights [][]float64 `json:"initial_weights,omitempty"`

	// Compute parameters
	QuantLevels     int    `json:"quant_levels"`     // Quantization levels for weights
	AccumulatorBits int    `json:"accumulator_bits"` // Bit width for MAC accumulator
	ActivationFunc  string `json:"activation_func"`  // "none", "relu", "sigmoid"
}

// NewArrayConfig creates a new array configuration with sensible defaults
func NewArrayConfig(mode OperationMode, rows, cols int) *ArrayConfig {
	cfg := &ArrayConfig{
		Name:         "fecim_crossbar",
		Mode:         mode,
		ArrayRows:    rows,
		ArrayCols:    cols,
		Technology:   TechSKY130,
		Architecture: ArchPassive,
		CellPitch:    sky130CellPitch,
		RowHeight:    sky130RowHeight,
		Levels:       30,    // FeCIM standard (conference claim, DefaultLevels)
		GMin:         10.0,  // µS
		GMax:         100.0, // µS
		VProgMin:     2.0,   // V
		VProgMax:     5.0,   // V
		TPulse:       50.0,  // ns
		Peripherals: PeripheralConfig{
			DACBits:   8,
			ADCBits:   8,
			TIAGain:   10000.0, // 10 kΩ
			VDD:       sky130VDD,
			ClockFreq: sky130ClockMHz,
		},
	}

	// Initialize mode-specific config
	switch mode {
	case ModeStorage:
		cfg.StorageConfig = &StorageArrayConfig{
			RetentionYears:  10,
			EnduranceCycles: 1000000,
			ErrorCorrection: "SECDED",
		}
	case ModeMemory:
		cfg.MemoryConfig = &MemoryArrayConfig{
			AccessTimeNs:  10.0,
			WriteTimeNs:   50.0,
			BandwidthGBps: 10.0,
			PowerBudgetMW: 100.0,
		}
	case ModeCompute:
		cfg.ComputeConfig = &ComputeArrayConfig{
			InitialWeights:  nil, // No initial weights by default
			QuantLevels:     30,
			AccumulatorBits: 16,
			ActivationFunc:  "none",
		}
	}

	return cfg
}

// NewStorageConfig creates a storage-optimized configuration
func NewStorageConfig(rows, cols int) *ArrayConfig {
	return NewArrayConfig(ModeStorage, rows, cols)
}

// NewMemoryConfig creates a memory-optimized configuration
func NewMemoryConfig(rows, cols int) *ArrayConfig {
	return NewArrayConfig(ModeMemory, rows, cols)
}

// NewComputeConfig creates a compute-optimized configuration
func NewComputeConfig(rows, cols int) *ArrayConfig {
	return NewArrayConfig(ModeCompute, rows, cols)
}

// With1T1R switches the configuration to 1T1R architecture
func (c *ArrayConfig) With1T1R() *ArrayConfig {
	c.Architecture = Arch1T1R
	c.CellPitch = cell1T1RPitch
	c.RowHeight = cell1T1RHeight
	return c
}

// With2T1R switches the configuration to 2T1R architecture
// 2T1R uses dual transistors (row + column select) for individual cell addressing
func (c *ArrayConfig) With2T1R() *ArrayConfig {
	c.Architecture = Arch2T1R
	c.CellPitch = cell2T1RPitch
	c.RowHeight = cell1T1RHeight // Same height as 1T1R
	return c
}

// WithWeights sets initial weights for compute mode
// Returns error if not in compute mode
func (c *ArrayConfig) WithWeights(weights [][]float64) *ArrayConfig {
	if c.Mode != ModeCompute {
		return c // Ignore for non-compute modes
	}
	if c.ComputeConfig == nil {
		c.ComputeConfig = &ComputeArrayConfig{QuantLevels: 30}
	}
	c.ComputeConfig.InitialWeights = weights
	return c
}

// CellAssignment represents one programmed FeFET cell
type CellAssignment struct {
	Row         int     `json:"row"`
	Col         int     `json:"col"`
	Level       int     `json:"level"`       // Programmed level (0 to Levels-1)
	Conductance float64 `json:"conductance"` // μS
	Resistance  float64 `json:"resistance"`  // Ω (1e6/Conductance)
	ProgramV    float64 `json:"program_v"`   // Programming voltage

	// Only populated for compute mode with initial weights
	InitialWeight float64 `json:"initial_weight,omitempty"`

	// Legacy fields for backward compatibility
	WeightValue float64 `json:"weight_value,omitempty"` // Deprecated: use InitialWeight
	QuantLevel  int     `json:"quant_level,omitempty"`  // Deprecated: use Level
}

// ArrayDesign is the complete design output
type ArrayDesign struct {
	Config *ArrayConfig     `json:"config"`
	Cells  []CellAssignment `json:"cells"`
	Stats  DesignStats      `json:"stats"`
}

// DesignStats holds design statistics
type DesignStats struct {
	TotalCells     int     `json:"total_cells"`
	ActiveCells    int     `json:"active_cells"`
	AreaMM2        float64 `json:"area_mm2"`
	PowerMW        float64 `json:"power_mw"`        // Estimated power
	ThroughputGOPS float64 `json:"throughput_gops"` // For compute mode

	// Compute mode only (when weights provided)
	QuantMSE  float64 `json:"quant_mse,omitempty"`     // Mean squared error
	QuantPSNR float64 `json:"quant_psnr_db,omitempty"` // Peak signal-to-noise
	WeightMin float64 `json:"weight_min,omitempty"`
	WeightMax float64 `json:"weight_max,omitempty"`

	// Legacy fields for backward compatibility
	UsedCells    int     `json:"used_cells,omitempty"`    // Deprecated: use ActiveCells
	Utilization  float64 `json:"utilization,omitempty"`   // Deprecated: calculate from ActiveCells/TotalCells
	UniqueLevels int     `json:"unique_levels,omitempty"` // Deprecated: compute-mode specific
}

// ============================================================================
// LEGACY SUPPORT - For backward compatibility with existing code
// These types wrap the new architecture while maintaining the old interface
// ============================================================================

// CompileConfig is the legacy configuration type
// Deprecated: Use ArrayConfig instead
type CompileConfig = ArrayConfig

// DefaultConfig returns a legacy-compatible default configuration
// Deprecated: Use NewComputeConfig instead
func DefaultConfig() CompileConfig {
	cfg := NewComputeConfig(128, 128)
	return *cfg
}

// Config1T1R returns a legacy-compatible 1T1R configuration
// Deprecated: Use NewComputeConfig().With1T1R() instead
func Config1T1R() CompileConfig {
	cfg := NewComputeConfig(128, 128).With1T1R()
	return *cfg
}

// Config2T1R returns a CompileConfig pre-configured for 2T1R architecture
func Config2T1R() CompileConfig {
	cfg := NewComputeConfig(128, 128).With2T1R()
	return *cfg
}

// CrossbarMapping is the legacy output type
// Deprecated: Use ArrayDesign instead
type CrossbarMapping = ArrayDesign

// Stats is the legacy statistics type
// Deprecated: Use DesignStats instead
type Stats = DesignStats
