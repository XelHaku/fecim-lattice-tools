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

// ── Energy constants ─────────────────────────────────────────────────────────

const (
	// EnergyPerBitPerMACJ is compute energy per MAC per effective bit of precision (J).
	// Source: Jerry et al. (IEDM 2017) ~50 fJ/MAC at ~4.9 bits → ~10 fJ/bit/MAC.
	EnergyPerBitPerMACJ = 10e-15

	// DefaultEnergyPerDACJ is the default DAC conversion energy (J).
	// Typical 4-bit R-2R ladder DAC at 65 nm node.
	DefaultEnergyPerDACJ = 0.1e-12 // 100 fJ

	// DefaultEnergyPerADCJ is the default ADC conversion energy (J).
	// Typical 4-bit SAR ADC; ADC dominates 40-60% of system energy.
	DefaultEnergyPerADCJ = 0.5e-12 // 500 fJ
)

// ── MAC energy ───────────────────────────────────────────────────────────────

// MACEnergyJ returns the compute energy per single MAC operation (J) for a crossbar
// array running at the given quantization level count.
//
// Effective bits are clamped to at least 1 to prevent degenerate results for
// binary (2-level) arrays.
func MACEnergyJ(levels int) float64 {
	if levels <= 0 {
		levels = 1
	}
	bits := math.Log2(float64(levels))
	if bits < 1 {
		bits = 1
	}
	return EnergyPerBitPerMACJ * bits
}

// ── Inference energy estimate ─────────────────────────────────────────────────

// InferenceEnergy holds a per-inference energy breakdown.
// All energy fields are in Joules (J). MAC counts are integers.
type InferenceEnergy struct {
	TotalJ    float64 // total energy = ComputeJ + ADCDACJ
	ComputeJ  float64 // crossbar MAC energy only
	ADCDACJ   float64 // ADC + DAC peripheral energy
	TotalMACs int     // total multiply-accumulate operations

	// Per-layer detail (populated when multiple layers are provided)
	LayerJ    []float64 // ComputeJ per layer
	LayerMACs []int     // MAC count per layer
}

// MLPEnergyConfig parameterises an MLP inference energy estimate.
type MLPEnergyConfig struct {
	// LayerSizes lists neuron counts: [input, hidden1, hidden2, ..., output].
	// Must have at least 2 elements.
	LayerSizes []int

	// LevelsPerLayer lists the quantisation level count for each layer's crossbar.
	// Length must equal len(LayerSizes)-1.
	// If nil, DefaultLevels is used for all layers.
	LevelsPerLayer []int

	// DefaultLevels is used when LevelsPerLayer is nil (default: 16).
	DefaultLevels int

	// EnergyPerDACJ is the DAC energy per conversion (default: DefaultEnergyPerDACJ).
	EnergyPerDACJ float64

	// EnergyPerADCJ is the ADC energy per conversion (default: DefaultEnergyPerADCJ).
	EnergyPerADCJ float64
}

func (c *MLPEnergyConfig) defaults() {
	if c.DefaultLevels <= 0 {
		c.DefaultLevels = 16
	}
	if c.EnergyPerDACJ == 0 {
		c.EnergyPerDACJ = DefaultEnergyPerDACJ
	}
	if c.EnergyPerADCJ == 0 {
		c.EnergyPerADCJ = DefaultEnergyPerADCJ
	}
}

// EstimateMLPEnergyJ estimates the total energy for one forward pass of an MLP.
//
// Layer sizes are given as [inputDim, hidden1Dim, ..., outputDim]. Each
// adjacent pair forms one crossbar layer. The model accounts for:
//   - Compute energy: MACs × MACEnergyJ(levels)
//   - DAC energy:     inputDim conversions (first layer column drivers)
//   - ADC energy:     all hidden + output neurons (one ADC per output row)
func EstimateMLPEnergyJ(cfg MLPEnergyConfig) InferenceEnergy {
	cfg.defaults()
	nLayers := len(cfg.LayerSizes) - 1
	if nLayers <= 0 {
		return InferenceEnergy{}
	}

	est := InferenceEnergy{
		LayerJ:    make([]float64, nLayers),
		LayerMACs: make([]int, nLayers),
	}

	// DAC energy: drive the input layer columns
	est.ADCDACJ += float64(cfg.LayerSizes[0]) * cfg.EnergyPerDACJ

	for l := 0; l < nLayers; l++ {
		in := cfg.LayerSizes[l]
		out := cfg.LayerSizes[l+1]

		levels := cfg.DefaultLevels
		if cfg.LevelsPerLayer != nil && l < len(cfg.LevelsPerLayer) {
			levels = cfg.LevelsPerLayer[l]
		}

		macs := in * out
		layerE := float64(macs) * MACEnergyJ(levels)

		est.LayerMACs[l] = macs
		est.LayerJ[l] = layerE
		est.TotalMACs += macs
		est.ComputeJ += layerE

		// ADC energy: one conversion per output neuron
		est.ADCDACJ += float64(out) * cfg.EnergyPerADCJ
	}

	est.TotalJ = est.ComputeJ + est.ADCDACJ
	return est
}
