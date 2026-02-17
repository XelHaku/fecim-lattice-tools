// Package crossbar implements ferroelectric crossbar array simulation.
package crossbar

import (
	"math"
)

// OptimizedParasiticSolver is a memory-efficient version of ParasiticSolver.
// It pre-allocates all work buffers and reuses them across SolveMVM calls,
// eliminating per-iteration allocations that dominate the hot path.
type OptimizedParasiticSolver struct {
	rows   int
	cols   int
	config *SORConfig

	// Normalized parasitic resistances (Rp/Rmin)
	RpRow float64
	RpCol float64

	// Conductance matrix (normalized 0-1)
	conductances [][]float64

	// Pre-allocated work buffers (reused across calls)
	dV        [][]float64 // Device voltages
	Ires      [][]float64 // Device currents
	IsumCol   [][]float64 // Cumulative column currents
	IsumRow   [][]float64 // Cumulative row currents
	VdropsCol [][]float64 // Column voltage drops
	VdropsRow [][]float64 // Row voltage drops
	VerrMat   [][]float64 // Voltage error matrix

	// Output buffer (reused for result building)
	output []float64
}

// NewOptimizedParasiticSolver creates an allocation-efficient parasitic solver.
// All work buffers are pre-allocated once and reused across SolveMVM calls.
func NewOptimizedParasiticSolver(rows, cols int, config *SORConfig) (*OptimizedParasiticSolver, error) {
	if rows <= 0 || cols <= 0 {
		return nil, ErrInvalidConfiguration
	}
	if config == nil {
		config = DefaultSORConfig()
	}

	s := &OptimizedParasiticSolver{
		rows:         rows,
		cols:         cols,
		config:       config,
		conductances: make([][]float64, rows),
	}

	// Pre-allocate all work matrices
	s.dV = make([][]float64, rows)
	s.Ires = make([][]float64, rows)
	s.IsumCol = make([][]float64, rows)
	s.IsumRow = make([][]float64, rows)
	s.VdropsCol = make([][]float64, rows)
	s.VdropsRow = make([][]float64, rows)
	s.VerrMat = make([][]float64, rows)

	for i := 0; i < rows; i++ {
		s.conductances[i] = make([]float64, cols)
		s.dV[i] = make([]float64, cols)
		s.Ires[i] = make([]float64, cols)
		s.IsumCol[i] = make([]float64, cols)
		s.IsumRow[i] = make([]float64, cols)
		s.VdropsCol[i] = make([]float64, cols)
		s.VdropsRow[i] = make([]float64, cols)
		s.VerrMat[i] = make([]float64, cols)
	}

	s.output = make([]float64, cols)

	return s, nil
}

// SetConductances sets the conductance matrix (normalized 0-1 values).
func (s *OptimizedParasiticSolver) SetConductances(g [][]float64) {
	for i := range g {
		if i >= s.rows {
			break
		}
		copy(s.conductances[i], g[i])
	}
}

// SetParasitics sets the parasitic resistance values.
func (s *OptimizedParasiticSolver) SetParasitics(rpRow, rpCol float64) {
	s.RpRow = rpRow
	s.RpCol = rpCol
}

// SolveMVM performs matrix-vector multiplication with parasitic resistance effects.
// This optimized version reuses pre-allocated buffers, eliminating per-iteration allocations.
func (s *OptimizedParasiticSolver) SolveMVM(appliedVoltages []float64) (*ParasiticMVMResult, error) {
	if len(appliedVoltages) != s.cols {
		return nil, ErrInvalidConfiguration
	}

	rows, cols := s.rows, s.cols
	dV := s.dV
	Ires := s.Ires
	IsumCol := s.IsumCol
	IsumRow := s.IsumRow
	VdropsCol := s.VdropsCol
	VdropsRow := s.VdropsRow
	VerrMat := s.VerrMat

	// Initialize device voltages to applied voltages (initial guess)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			dV[i][j] = appliedVoltages[j]
			Ires[i][j] = 0
		}
	}

	omega := s.config.OmegaInitial
	prevErr := math.MaxFloat64
	var finalErr float64

	for iter := 0; iter < s.config.MaxIterations; iter++ {
		// Step 1: Compute device currents (Ohm's law: I = G × V)
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				g := s.conductances[i][j]
				if g < 1e-10 {
					g = 1e-10
				}
				Ires[i][j] = g * dV[i][j]
			}
		}

		// Step 2: Compute cumulative currents along columns (from bit line ground to row)
		for j := 0; j < cols; j++ {
			IsumCol[0][j] = Ires[0][j]
		}
		for i := 1; i < rows; i++ {
			for j := 0; j < cols; j++ {
				IsumCol[i][j] = IsumCol[i-1][j] + Ires[i][j]
			}
		}

		// Step 3: Compute cumulative currents along rows (from row driver to column)
		for i := 0; i < rows; i++ {
			IsumRow[i][cols-1] = Ires[i][cols-1]
		}
		for i := 0; i < rows; i++ {
			for j := cols - 2; j >= 0; j-- {
				IsumRow[i][j] = IsumRow[i][j+1] + Ires[i][j]
			}
		}

		// Step 4: Compute parasitic voltage drops - Column (bit line)
		for j := 0; j < cols; j++ {
			VdropsCol[0][j] = 0
		}
		for i := 1; i < rows; i++ {
			for j := 0; j < cols; j++ {
				VdropsCol[i][j] = VdropsCol[i-1][j] + s.RpCol*IsumCol[i-1][j]
			}
		}

		// Step 4b: Row drops (word line)
		for i := 0; i < rows; i++ {
			VdropsRow[i][0] = 0
		}
		for i := 0; i < rows; i++ {
			for j := 1; j < cols; j++ {
				VdropsRow[i][j] = VdropsRow[i][j-1] + s.RpRow*IsumRow[i][j]
			}
		}

		// Step 5: Calculate voltage error
		maxErr := 0.0
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				Vexpected := appliedVoltages[j] - VdropsRow[i][j] - VdropsCol[i][j]
				if Vexpected < 0 {
					Vexpected = 0
				}
				VerrMat[i][j] = Vexpected - dV[i][j]
				if absVal := math.Abs(VerrMat[i][j]); absVal > maxErr {
					maxErr = absVal
				}
			}
		}
		finalErr = maxErr

		// Step 6: Check convergence
		if maxErr < s.config.Tolerance {
			return s.buildResult(iter+1, true, maxErr, omega), nil
		}

		// Step 7: Detect divergence and adapt omega
		if s.config.AdaptiveOmega && iter > 5 && maxErr > prevErr {
			omega *= s.config.OmegaDecay
			if omega < s.config.OmegaMin {
				return s.buildResult(iter+1, false, maxErr, omega), ErrConvergenceFailed
			}
		}
		prevErr = maxErr

		// Step 8: SOR update
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				dV[i][j] += omega * VerrMat[i][j]
				if dV[i][j] < 0 {
					dV[i][j] = 0
				}
				if dV[i][j] > appliedVoltages[j] {
					dV[i][j] = appliedVoltages[j]
				}
			}
		}
	}

	return s.buildResult(s.config.MaxIterations, false, finalErr, omega), ErrMaxIterationsExceeded
}

// buildResult constructs the MVMResult from solver state without allocating new matrices.
func (s *OptimizedParasiticSolver) buildResult(iters int, converged bool, maxErr, omega float64) *ParasiticMVMResult {
	// Sum currents along columns to get output (reuse pre-allocated buffer)
	for j := 0; j < s.cols; j++ {
		s.output[j] = 0
		for i := 0; i < s.rows; i++ {
			s.output[j] += s.Ires[i][j]
		}
	}

	// Copy results for return (these allocations are unavoidable for the API contract)
	output := make([]float64, s.cols)
	copy(output, s.output)

	deviceCurrents := make([][]float64, s.rows)
	deviceVoltages := make([][]float64, s.rows)
	for i := 0; i < s.rows; i++ {
		deviceCurrents[i] = make([]float64, s.cols)
		deviceVoltages[i] = make([]float64, s.cols)
		copy(deviceCurrents[i], s.Ires[i])
		copy(deviceVoltages[i], s.dV[i])
	}

	return &ParasiticMVMResult{
		OutputCurrents: output,
		DeviceCurrents: deviceCurrents,
		DeviceVoltages: deviceVoltages,
		Iterations:     iters,
		Converged:      converged,
		MaxError:       maxErr,
		FinalOmega:     omega,
	}
}

// SolveMVMFast performs MVM and returns only the output currents (no device matrices).
// This is the fastest path when only the aggregate outputs are needed.
func (s *OptimizedParasiticSolver) SolveMVMFast(appliedVoltages []float64) ([]float64, int, error) {
	if len(appliedVoltages) != s.cols {
		return nil, 0, ErrInvalidConfiguration
	}

	rows, cols := s.rows, s.cols
	dV := s.dV
	Ires := s.Ires
	IsumCol := s.IsumCol
	IsumRow := s.IsumRow
	VdropsCol := s.VdropsCol
	VdropsRow := s.VdropsRow
	VerrMat := s.VerrMat

	// Initialize
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			dV[i][j] = appliedVoltages[j]
		}
	}

	omega := s.config.OmegaInitial
	prevErr := math.MaxFloat64

	for iter := 0; iter < s.config.MaxIterations; iter++ {
		// Compute device currents
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				g := s.conductances[i][j]
				if g < 1e-10 {
					g = 1e-10
				}
				Ires[i][j] = g * dV[i][j]
			}
		}

		// Cumulative column currents
		for j := 0; j < cols; j++ {
			IsumCol[0][j] = Ires[0][j]
		}
		for i := 1; i < rows; i++ {
			for j := 0; j < cols; j++ {
				IsumCol[i][j] = IsumCol[i-1][j] + Ires[i][j]
			}
		}

		// Cumulative row currents
		for i := 0; i < rows; i++ {
			IsumRow[i][cols-1] = Ires[i][cols-1]
			for j := cols - 2; j >= 0; j-- {
				IsumRow[i][j] = IsumRow[i][j+1] + Ires[i][j]
			}
		}

		// Column voltage drops
		for j := 0; j < cols; j++ {
			VdropsCol[0][j] = 0
		}
		for i := 1; i < rows; i++ {
			for j := 0; j < cols; j++ {
				VdropsCol[i][j] = VdropsCol[i-1][j] + s.RpCol*IsumCol[i-1][j]
			}
		}

		// Row voltage drops
		for i := 0; i < rows; i++ {
			VdropsRow[i][0] = 0
			for j := 1; j < cols; j++ {
				VdropsRow[i][j] = VdropsRow[i][j-1] + s.RpRow*IsumRow[i][j]
			}
		}

		// Voltage error and convergence check
		maxErr := 0.0
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				Vexpected := appliedVoltages[j] - VdropsRow[i][j] - VdropsCol[i][j]
				if Vexpected < 0 {
					Vexpected = 0
				}
				VerrMat[i][j] = Vexpected - dV[i][j]
				if absVal := math.Abs(VerrMat[i][j]); absVal > maxErr {
					maxErr = absVal
				}
			}
		}

		if maxErr < s.config.Tolerance {
			// Compute output currents
			for j := 0; j < cols; j++ {
				s.output[j] = 0
				for i := 0; i < rows; i++ {
					s.output[j] += Ires[i][j]
				}
			}
			result := make([]float64, cols)
			copy(result, s.output)
			return result, iter + 1, nil
		}

		// Adaptive omega
		if s.config.AdaptiveOmega && iter > 5 && maxErr > prevErr {
			omega *= s.config.OmegaDecay
			if omega < s.config.OmegaMin {
				return nil, iter + 1, ErrConvergenceFailed
			}
		}
		prevErr = maxErr

		// SOR update
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				dV[i][j] += omega * VerrMat[i][j]
				if dV[i][j] < 0 {
					dV[i][j] = 0
				}
				if dV[i][j] > appliedVoltages[j] {
					dV[i][j] = appliedVoltages[j]
				}
			}
		}
	}

	return nil, s.config.MaxIterations, ErrMaxIterationsExceeded
}

// ComputeIdealMVM computes ideal MVM without parasitic effects for comparison.
func (s *OptimizedParasiticSolver) ComputeIdealMVM(appliedVoltages []float64) []float64 {
	output := make([]float64, s.cols)
	for j := 0; j < s.cols; j++ {
		for i := 0; i < s.rows; i++ {
			output[j] += s.conductances[i][j] * appliedVoltages[j]
		}
	}
	return output
}
