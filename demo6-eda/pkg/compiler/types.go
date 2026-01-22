// pkg/compiler/types.go
// All other files depend on these data structures

package compiler

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
}

// DefaultConfig returns standard FeCIM parameters
func DefaultConfig() CompileConfig {
	return CompileConfig{
		ArrayRows: 128,
		ArrayCols: 128,
		Levels:    30,
		GMin:      1.0,
		GMax:      100.0,
		VProgMin:  2.0,
		VProgMax:  5.0,
		TPulse:    50.0,
	}
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
