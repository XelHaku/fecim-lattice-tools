// pkg/compiler/types.go
// All other files depend on these data structures

package compiler

// Architecture types for crossbar array
const (
	ArchPassive = "passive" // Passive crossbar (WL, BL only)
	Arch1T1R    = "1T1R"    // 1 Transistor 1 Resistor (WL, BL, SL)
)

// CompileConfig holds all parameters for compilation
type CompileConfig struct {
	ArrayRows int     `json:"array_rows"` // Target array rows
	ArrayCols int     `json:"array_cols"` // Target array cols
	Levels    int     `json:"levels"`     // Quantization levels (2-30)
	GMin      float64 `json:"g_min"`      // Min conductance (μS)
	GMax      float64 `json:"g_max"`      // Max conductance (μS)
	VProgMin  float64 `json:"v_prog_min"` // Min programming voltage (V)
	VProgMax  float64 `json:"v_prog_max"` // Max programming voltage (V)
	TPulse    float64 `json:"t_pulse"`    // Pulse width (ns)

	// Phase 2: Architecture configuration
	Architecture string  `json:"architecture"` // "passive" or "1T1R"
	CellPitch    float64 `json:"cell_pitch"`   // Cell width in microns (default: 0.46)
	RowHeight    float64 `json:"row_height"`   // Cell height in microns (default: 2.72)
}

// DefaultConfig returns standard FeCIM parameters
func DefaultConfig() CompileConfig {
	return CompileConfig{
		ArrayRows:    128,
		ArrayCols:    128,
		Levels:       30,
		GMin:         1.0,
		GMax:         100.0,
		VProgMin:     2.0,
		VProgMax:     5.0,
		TPulse:       50.0,
		Architecture: ArchPassive, // Default: passive crossbar
		CellPitch:    0.46,        // SKY130 compatible: 2 * 0.23um site width
		RowHeight:    2.72,        // SKY130 standard cell row height
	}
}

// Config1T1R returns configuration for 1T1R architecture
// 1T1R uses a select transistor per cell to mitigate sneak paths
func Config1T1R() CompileConfig {
	cfg := DefaultConfig()
	cfg.Architecture = Arch1T1R
	cfg.CellPitch = 0.92  // Larger cell for transistor (4x site width)
	cfg.RowHeight = 2.72  // Same row height
	return cfg
}

// CellAssignment represents one programmed FeFET cell
type CellAssignment struct {
	Row         int     `json:"row"`
	Col         int     `json:"col"`
	WeightValue float64 `json:"weight_value"` // Original weight
	QuantLevel  int     `json:"quant_level"`  // 0 to Levels-1
	Conductance float64 `json:"conductance"`  // μS
	ProgramV    float64 `json:"program_v"`    // V
}

// CrossbarMapping is the complete compilation output
type CrossbarMapping struct {
	Config CompileConfig    `json:"config"`
	Cells  []CellAssignment `json:"cells"` // Flat array for simplicity
	Stats  Stats            `json:"stats"`
}

// Stats holds compilation statistics
type Stats struct {
	TotalCells   int     `json:"total_cells"`
	UsedCells    int     `json:"used_cells"`
	Utilization  float64 `json:"utilization"`   // 0.0 to 1.0
	WeightMin    float64 `json:"weight_min"`
	WeightMax    float64 `json:"weight_max"`
	QuantMSE     float64 `json:"quant_mse"`     // Mean squared error
	QuantPSNR    float64 `json:"quant_psnr_db"` // dB
	UniqueLevels int     `json:"unique_levels"`
}
