package crossbar

type CrossbarState struct {
	Rows         int         `json:"rows"`
	Cols         int         `json:"cols"`
	Conductances [][]float64 `json:"conductances"`
	IRDrop       float64     `json:"ir_drop"`
	SneakPaths   bool        `json:"sneak_paths"`
	DriftFactor  float64     `json:"drift_factor"`
	InputVector  []float64   `json:"input_vector"`
	OutputVector []float64   `json:"output_vector"`
}
