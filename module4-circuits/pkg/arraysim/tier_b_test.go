package arraysim

import (
	"testing"
)

// TestNewTierBSolver_Defaults verifies default initialization.
func TestNewTierBSolver_Defaults(t *testing.T) {
	solver := NewTierBSolver()

	if solver == nil {
		t.Fatal("NewTierBSolver returned nil")
	}

	if solver.MaxDenseSize != 16 {
		t.Errorf("MaxDenseSize = %d, want 16", solver.MaxDenseSize)
	}
}

// TestTierBSolver_maxSize_Default verifies maxSize returns 16 when MaxDenseSize <= 0.
func TestTierBSolver_maxSize_Default(t *testing.T) {
	tests := []struct {
		name         string
		maxDenseSize int
		want         int
	}{
		{"zero", 0, 16},
		{"negative", -5, 16},
		{"unset nil receiver", 0, 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			solver := &TierBSolver{MaxDenseSize: tt.maxDenseSize}
			got := solver.maxSize()
			if got != tt.want {
				t.Errorf("maxSize() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestTierBSolver_maxSize_Custom verifies maxSize returns custom value when set.
func TestTierBSolver_maxSize_Custom(t *testing.T) {
	tests := []struct {
		maxDenseSize int
		want         int
	}{
		{1, 1},
		{4, 4},
		{16, 16},
		{64, 64},
		{256, 256},
	}

	for _, tt := range tests {
		solver := &TierBSolver{MaxDenseSize: tt.maxDenseSize}
		got := solver.maxSize()
		if got != tt.want {
			t.Errorf("maxSize() with MaxDenseSize=%d = %d, want %d", tt.maxDenseSize, got, tt.want)
		}
	}
}

// TestTierBSolver_SolveDC_EmptyInput verifies handling of empty arrays.
func TestTierBSolver_SolveDC_EmptyInput(t *testing.T) {
	solver := NewTierBSolver()

	tests := []struct {
		name   string
		params SolveParams
	}{
		{
			name: "zero rows",
			params: SolveParams{
				Conductance: [][]float64{},
				BLVoltages:  []float64{1.0, 0.0},
			},
		},
		{
			name: "zero cols",
			params: SolveParams{
				Conductance: [][]float64{{}, {}},
				BLVoltages:  []float64{},
			},
		},
		{
			name: "nil conductance",
			params: SolveParams{
				Conductance: nil,
				BLVoltages:  []float64{1.0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := solver.SolveDC(tt.params)
			if err != nil {
				t.Errorf("SolveDC() error = %v, want nil", err)
			}
			// Result should be empty but valid
			if result.CellVoltages != nil && len(result.CellVoltages) > 0 {
				t.Errorf("Expected empty CellVoltages, got %d rows", len(result.CellVoltages))
			}
		})
	}
}

// TestTierBSolver_SolveDC_SmallArray verifies solving a small array.
func TestTierBSolver_SolveDC_SmallArray(t *testing.T) {
	solver := NewTierBSolver()

	// Create a simple 2x2 array
	params := SolveParams{
		WLVoltages: []float64{1.0, 1.0},
		BLVoltages: []float64{0.0, 0.0},
		Conductance: [][]float64{
			{1e-6, 1e-6}, // 1 µS conductance per cell
			{1e-6, 1e-6},
		},
		Geometry: DefaultCellGeometry(),
		Wire:     WireParams{}.WithDefaults(DefaultCellGeometry()),
	}

	result, err := solver.SolveDC(params)
	if err != nil {
		t.Fatalf("SolveDC() error = %v, want nil", err)
	}

	// Verify result structure
	if len(result.CellVoltages) != 2 {
		t.Errorf("CellVoltages rows = %d, want 2", len(result.CellVoltages))
	}
	if len(result.CellVoltages) > 0 && len(result.CellVoltages[0]) != 2 {
		t.Errorf("CellVoltages cols = %d, want 2", len(result.CellVoltages[0]))
	}

	if len(result.CellCurrents) != 2 {
		t.Errorf("CellCurrents rows = %d, want 2", len(result.CellCurrents))
	}
	if len(result.CellCurrents) > 0 && len(result.CellCurrents[0]) != 2 {
		t.Errorf("CellCurrents cols = %d, want 2", len(result.CellCurrents[0]))
	}

	// Verify currents are reasonable (non-zero for non-zero conductance)
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			if result.CellCurrents[i][j] == 0 {
				t.Errorf("CellCurrents[%d][%d] = 0, expected non-zero", i, j)
			}
		}
	}
}

// TestTierBSolver_SolveDC_MaxSizeExceeded verifies error when array exceeds max size.
func TestTierBSolver_SolveDC_MaxSizeExceeded(t *testing.T) {
	solver := &TierBSolver{MaxDenseSize: 4} // Set small max size

	// Create a 3x3 array (9 cells > 4)
	params := SolveParams{
		WLVoltages: []float64{1.0, 1.0, 1.0},
		BLVoltages: []float64{0.0, 0.0, 0.0},
		Conductance: [][]float64{
			{1e-6, 1e-6, 1e-6},
			{1e-6, 1e-6, 1e-6},
			{1e-6, 1e-6, 1e-6},
		},
	}

	_, err := solver.SolveDC(params)
	if err == nil {
		t.Error("SolveDC() error = nil, want error for exceeding max size")
	}
}

// TestTierBSolver_Solve_DelegatesToSolveDC verifies Solve returns SolveResult portion.
func TestTierBSolver_Solve_DelegatesToSolveDC(t *testing.T) {
	solver := NewTierBSolver()

	// Create a simple 2x2 array
	params := SolveParams{
		WLVoltages: []float64{1.0, 1.0},
		BLVoltages: []float64{0.0, 0.0},
		Conductance: [][]float64{
			{1e-6, 1e-6},
			{1e-6, 1e-6},
		},
		Geometry: DefaultCellGeometry(),
		Wire:     WireParams{}.WithDefaults(DefaultCellGeometry()),
	}

	// Call Solve (which should delegate to SolveDC)
	solveResult, err := solver.Solve(params)
	if err != nil {
		t.Fatalf("Solve() error = %v, want nil", err)
	}

	// Call SolveDC directly for comparison
	dcResult, err := solver.SolveDC(params)
	if err != nil {
		t.Fatalf("SolveDC() error = %v, want nil", err)
	}

	// Verify Solve returns the SolveResult portion of DCResult
	if len(solveResult.CellVoltages) != len(dcResult.CellVoltages) {
		t.Errorf("Solve CellVoltages length mismatch")
	}
	if len(solveResult.CellCurrents) != len(dcResult.CellCurrents) {
		t.Errorf("Solve CellCurrents length mismatch")
	}
	if len(solveResult.RowCurrents) != len(dcResult.RowCurrents) {
		t.Errorf("Solve RowCurrents length mismatch")
	}
	if len(solveResult.ColCurrents) != len(dcResult.ColCurrents) {
		t.Errorf("Solve ColCurrents length mismatch")
	}
}

// TestTierBSolver_SolveDC_NodeVoltages verifies DCResult includes node voltages.
func TestTierBSolver_SolveDC_NodeVoltages(t *testing.T) {
	solver := NewTierBSolver()

	// Create a simple 2x2 array
	params := SolveParams{
		WLVoltages: []float64{1.0, 1.0},
		BLVoltages: []float64{0.0, 0.0},
		Conductance: [][]float64{
			{1e-6, 1e-6},
			{1e-6, 1e-6},
		},
		Geometry: DefaultCellGeometry(),
		Wire:     WireParams{}.WithDefaults(DefaultCellGeometry()),
	}

	result, err := solver.SolveDC(params)
	if err != nil {
		t.Fatalf("SolveDC() error = %v, want nil", err)
	}

	// Verify WLNodes structure
	if len(result.WLNodes) != 2 {
		t.Errorf("WLNodes rows = %d, want 2", len(result.WLNodes))
	}
	if len(result.WLNodes) > 0 && len(result.WLNodes[0]) != 2 {
		t.Errorf("WLNodes cols = %d, want 2", len(result.WLNodes[0]))
	}

	// Verify BLNodes structure
	if len(result.BLNodes) != 2 {
		t.Errorf("BLNodes rows = %d, want 2", len(result.BLNodes))
	}
	if len(result.BLNodes) > 0 && len(result.BLNodes[0]) != 2 {
		t.Errorf("BLNodes cols = %d, want 2", len(result.BLNodes[0]))
	}

	// Verify node voltages are in reasonable range [0, 1]
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			wlVolt := result.WLNodes[i][j]
			if wlVolt < 0 || wlVolt > 1.5 {
				t.Errorf("WLNodes[%d][%d] = %f, out of reasonable range [0, 1.5]", i, j, wlVolt)
			}

			blVolt := result.BLNodes[i][j]
			if blVolt < -0.5 || blVolt > 1.0 {
				t.Errorf("BLNodes[%d][%d] = %f, out of reasonable range [-0.5, 1.0]", i, j, blVolt)
			}
		}
	}
}

// TestTierBSolver_SolveDC_SingleCell verifies solving a 1x1 array.
func TestTierBSolver_SolveDC_SingleCell(t *testing.T) {
	solver := NewTierBSolver()

	// Create a 1x1 array
	params := SolveParams{
		WLVoltages:  []float64{1.0},
		BLVoltages:  []float64{0.0},
		Conductance: [][]float64{{1e-6}},
		Geometry:    DefaultCellGeometry(),
		Wire:        WireParams{}.WithDefaults(DefaultCellGeometry()),
	}

	result, err := solver.SolveDC(params)
	if err != nil {
		t.Fatalf("SolveDC() error = %v, want nil", err)
	}

	// Verify result structure
	if len(result.CellVoltages) != 1 || len(result.CellVoltages[0]) != 1 {
		t.Errorf("CellVoltages size mismatch, got %dx%d, want 1x1",
			len(result.CellVoltages), len(result.CellVoltages[0]))
	}

	if len(result.CellCurrents) != 1 || len(result.CellCurrents[0]) != 1 {
		t.Errorf("CellCurrents size mismatch")
	}

	if len(result.WLNodes) != 1 || len(result.WLNodes[0]) != 1 {
		t.Errorf("WLNodes size mismatch")
	}

	if len(result.BLNodes) != 1 || len(result.BLNodes[0]) != 1 {
		t.Errorf("BLNodes size mismatch")
	}

	// Verify current is positive (flowing from WL to BL)
	if result.CellCurrents[0][0] <= 0 {
		t.Errorf("CellCurrents[0][0] = %e, expected positive", result.CellCurrents[0][0])
	}
}

// TestTierBSolver_SolveDC_DifferentSizes verifies arrays of different sizes.
func TestTierBSolver_SolveDC_DifferentSizes(t *testing.T) {
	solver := NewTierBSolver()

	tests := []struct {
		name string
		rows int
		cols int
	}{
		{"1x1", 1, 1},
		{"2x2", 2, 2},
		{"2x3", 2, 3},
		{"3x2", 3, 2},
		{"4x4", 4, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create conductance matrix
			cond := make([][]float64, tt.rows)
			for i := 0; i < tt.rows; i++ {
				cond[i] = make([]float64, tt.cols)
				for j := 0; j < tt.cols; j++ {
					cond[i][j] = 1e-6
				}
			}

			params := SolveParams{
				WLVoltages:  make([]float64, tt.rows),
				BLVoltages:  make([]float64, tt.cols),
				Conductance: cond,
				Geometry:    DefaultCellGeometry(),
				Wire:        WireParams{}.WithDefaults(DefaultCellGeometry()),
			}

			// Set voltages
			for i := 0; i < tt.rows; i++ {
				params.WLVoltages[i] = 1.0
			}

			result, err := solver.SolveDC(params)
			if err != nil {
				t.Fatalf("SolveDC() error = %v", err)
			}

			// Verify result dimensions
			if len(result.CellVoltages) != tt.rows {
				t.Errorf("CellVoltages rows = %d, want %d", len(result.CellVoltages), tt.rows)
			}
			if len(result.CellVoltages) > 0 && len(result.CellVoltages[0]) != tt.cols {
				t.Errorf("CellVoltages cols = %d, want %d", len(result.CellVoltages[0]), tt.cols)
			}
		})
	}
}

// TestTierBSolver_Interface verifies TierBSolver implements DCEngine.
func TestTierBSolver_Interface(t *testing.T) {
	var _ DCEngine = (*TierBSolver)(nil)
	var _ Engine = (*TierBSolver)(nil)
}
