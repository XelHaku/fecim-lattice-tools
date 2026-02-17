// Package crossbar implements ferroelectric crossbar array simulation.
// This file defines interface abstractions inspired by CrossSim's ICore/IDevice hierarchy.
package crossbar

// CrossbarCore is the primary abstraction for crossbar array simulators.
//
// Inspired by CrossSim's ICore hierarchy (AnalogCore, BalancedCore, BitSlicedCore,
// OffsetCore, NumericCore). All callers in shared/neural and module GUIs accept
// CrossbarCore — never the concrete *Array type — enabling plug-in implementations.
//
// Current implementations:
//   - *Array  (shared/crossbar/array.go) — behavioural MVM with optional IR-drop/noise
//
// Future implementations (Phase 0b+):
//   - *ExactKCLCore — badcrossbar-style nodal analysis (ground-truth oracle)
//   - *NoisyCore    — decorator that injects read/programming noise (Pattern E from plan)
type CrossbarCore interface {
	// MVM performs a matrix-vector multiply: output[i] = Σ_j G[i][j] · input[j].
	// input[j] is a normalised column voltage in [0, 1].
	// Returns row currents in physical units (A), or an error if dimensions mismatch.
	MVM(input []float64) ([]float64, error)

	// VMM performs a vector-matrix multiply: output[j] = Σ_i G[i][j] · input[i].
	// input[i] is a normalised row voltage in [0, 1].
	// Returns column currents in physical units (A), or an error if dimensions mismatch.
	VMM(input []float64) ([]float64, error)

	// ProgramWeightMatrix sets the conductance of every cell in one call.
	// weights[i][j] is a normalised weight in [0, 1]; the core maps it to a physical
	// conductance in [G_min, G_max] according to its device model.
	ProgramWeightMatrix(weights [][]float64) error

	// GetConductanceMatrix returns a snapshot of the current physical conductances (S).
	GetConductanceMatrix() [][]float64

	// Rows returns the number of rows (word-line count) in the array.
	Rows() int

	// Cols returns the number of columns (bit-line count) in the array.
	Cols() int
}

// DeviceModel abstracts per-cell error physics.
//
// Inspired by CrossSim's IDevice hierarchy which separates device-level error
// physics (programming error, read noise, drift) from array-level simulation.
// This enables swapping FeFET models for RRAM or PCM without changing array code.
//
// All error methods return a modified conductance (S).
type DeviceModel interface {
	// ApplyWriteError models programming inaccuracy: returns the conductance that
	// is actually stored after a write targeting targetG.
	ApplyWriteError(targetG float64) float64

	// ReadNoise models cycle-to-cycle read variation: returns the effective
	// conductance seen during a single read of a cell nominally at storedG.
	ReadNoise(storedG float64) float64

	// DriftError models time-dependent conductance decay: returns the conductance
	// after elapsedSeconds since last programming.
	DriftError(storedG, elapsedSeconds float64) float64

	// Levels returns the number of discrete conductance states.
	Levels() int

	// ConductanceRange returns the minimum and maximum physical conductances (S).
	ConductanceRange() (min, max float64)
}

// Compile-time check: *Array must satisfy CrossbarCore.
var _ CrossbarCore = (*Array)(nil)
