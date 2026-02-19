// Package system provides system-level cost models for FeCIM crossbar arrays.
//
// Inspired by MNSIM 2.0's 9-module architecture (Accuracy_Model, Area_Model,
// Energy_Model, Latency_Model, Power_Model, Mapping_Model, Hardware_Model, NoC,
// Interface). This package provides the energy and area estimation primitives
// needed for design-space exploration before fabrication.
//
// Physics basis:
//   - Compute energy: Jerry et al. IEDM 2017 — ~50 fJ/MAC at ~30 levels (~4.9 bits)
//     → ~10 fJ/bit/MAC (EnergyPerBitPerMACJ).
//   - ADC overhead: dominates 40–60% of system energy per the 2025 literature survey
//     (see docs/research/crossbar-circuits-literature-review-2025.md).
package system

import "math"

// TechnologyNode identifies a CMOS fabrication node.
type TechnologyNode string

const (
	Node130nm TechnologyNode = "130nm"
	Node65nm  TechnologyNode = "65nm"
	Node28nm  TechnologyNode = "28nm"
	Node22nm  TechnologyNode = "22nm"
	Node14nm  TechnologyNode = "14nm"
)

// CellType identifies the non-volatile memory device used in each crossbar cell.
type CellType string

const (
	CellFeFET CellType = "FeFET"
	CellRRAM  CellType = "RRAM"
	CellPCM   CellType = "PCM"
	CellFeCAP CellType = "FeCAP"
)

// CrossbarAreaModel estimates the physical area of a crossbar array and its
// peripheral circuits at a given technology node.
//
// Cell area values are approximate simulation defaults derived from published
// array layouts; calibrate against measured die photos for quantitative claims.
type CrossbarAreaModel struct {
	Rows int
	Cols int
	Node TechnologyNode
	Cell CellType
}

// NewCrossbarAreaModel creates a CrossbarAreaModel for the given array dimensions
// and technology/device combination.
func NewCrossbarAreaModel(rows, cols int, node TechnologyNode, cell CellType) *CrossbarAreaModel {
	return &CrossbarAreaModel{Rows: rows, Cols: cols, Node: node, Cell: cell}
}

// CellAreaUM2 returns the estimated area per cell in µm² for the configured
// technology node and device type.
//
// Values are drawn from published array layouts:
//   - FeFET: ~4F² at the given node (literature range 4–8F²)
//   - RRAM:  ~4F² (1T1R cell at the given node)
//   - PCM:   ~6F² (slightly larger due to heater structure)
//   - FeCAP: ~8F² (MIM stack with access transistor)
//
// F (feature size) is taken as half the node pitch.
func (m *CrossbarAreaModel) CellAreaUM2() float64 {
	// Feature size in nm (half-pitch approximation)
	var fNM float64
	switch m.Node {
	case Node130nm:
		fNM = 65
	case Node65nm:
		fNM = 32.5
	case Node28nm:
		fNM = 14
	case Node22nm:
		fNM = 11
	case Node14nm:
		fNM = 7
	default:
		fNM = 32.5 // default to 65 nm node
	}

	// F² in µm²
	f2 := (fNM * 1e-3) * (fNM * 1e-3)

	var factorF2 float64
	switch m.Cell {
	case CellFeFET:
		factorF2 = 4.0
	case CellRRAM:
		factorF2 = 4.0
	case CellPCM:
		factorF2 = 6.0
	case CellFeCAP:
		factorF2 = 8.0
	default:
		factorF2 = 4.0
	}
	return factorF2 * f2
}

// ArrayAreaUM2 returns the total cell array area in µm²:
//
//	Rows × Cols × CellAreaUM2()
func (m *CrossbarAreaModel) ArrayAreaUM2() float64 {
	return float64(m.Rows) * float64(m.Cols) * m.CellAreaUM2()
}

// PeripheralAreaUM2 estimates the combined ADC + DAC peripheral circuit area in µm².
//
// The estimate uses an empirical scaling:
//   - One ADC per column: area ∝ 2^adcBits (SAR topology scales exponentially with bits)
//   - One DAC per row: similar scaling, but DAC is smaller by ~0.5×
//
// Baseline at 4-bit, 65 nm: ~20 µm² per ADC, ~10 µm² per DAC.
// Area scales as 2^(bits-4) and as (65 nm / node_F) for node scaling.
func (m *CrossbarAreaModel) PeripheralAreaUM2(adcBits int) float64 {
	if adcBits <= 0 {
		adcBits = 4
	}

	// Node scaling factor relative to 65 nm baseline
	var fNM float64
	switch m.Node {
	case Node130nm:
		fNM = 65
	case Node65nm:
		fNM = 32.5
	case Node28nm:
		fNM = 14
	case Node22nm:
		fNM = 11
	case Node14nm:
		fNM = 7
	default:
		fNM = 32.5
	}
	// Peripheral area scales roughly as F² (transistor count × F²)
	nodeScale := (fNM / 32.5) * (fNM / 32.5) // normalised to 65 nm

	// Bit-resolution scaling relative to 4-bit baseline
	bitScale := math.Pow(2, float64(adcBits-4))

	const baselineADCUM2 = 20.0 // µm² per ADC at 4-bit, 65 nm
	const baselineDACUM2 = 10.0 // µm² per DAC at 4-bit, 65 nm

	adcTotal := float64(m.Cols) * baselineADCUM2 * bitScale * nodeScale
	dacTotal := float64(m.Rows) * baselineDACUM2 * bitScale * nodeScale
	return adcTotal + dacTotal
}

// TotalAreaUM2 returns ArrayAreaUM2() + PeripheralAreaUM2(adcBits).
func (m *CrossbarAreaModel) TotalAreaUM2(adcBits int) float64 {
	return m.ArrayAreaUM2() + m.PeripheralAreaUM2(adcBits)
}
