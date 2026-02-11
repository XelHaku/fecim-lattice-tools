package core

import "math"

// EnergyEstimate provides a simple, transparent energy breakdown for one inference.
//
// Units:
//   - All energy fields are in Joules (J).
//   - MAC counts are integer counts.
//
// The model is intentionally lightweight and is used by both core inference and the GUI.
// This avoids silent drift between the two.
type EnergyEstimate struct {
	TotalJ   float64
	ComputeJ float64
	ADCDACJ  float64

	TotalMACs int
	MACs1     int
	MACs2     int

	Layer1Levels int
	Layer2Levels int

	EnergyPerMAC1J float64
	EnergyPerMAC2J float64
}

const (
	// EnergyPerBitPerMACJ is the assumed compute energy per MAC per effective bit of precision.
	// Rationale: Jerry et al. (IEDM 2017) reports ~50 fJ/MAC around ~30 levels (~4.9 bits),
	// which maps to ~10 fJ/bit/MAC.
	EnergyPerBitPerMACJ = 10e-15

	// Peripheral overhead (simple constants used by the project so far).
	// These are kept here so GUI and core stay aligned.
	EnergyPerDACJ = 0.1e-12 // 0.1 pJ = 100 fJ per conversion
	EnergyPerADCJ = 0.5e-12 // 0.5 pJ = 500 fJ per conversion
)

// EnergyPerMACJ returns the compute energy per MAC in Joules for a given quantization level count.
// levels are interpreted as effective discrete levels per cell.
//
// The effective bits are clamped to at least 1.
func EnergyPerMACJ(levels int) float64 {
	if levels <= 0 {
		levels = 1
	}
	bits := math.Log2(float64(levels))
	if bits < 1 {
		bits = 1
	}
	return EnergyPerBitPerMACJ * bits
}

// EstimateInferenceEnergyJ estimates FeCIM energy for one inference using the shared model.
//
// This model includes:
//   - Compute energy (MACs × energy_per_MAC(levels))
//   - ADC/DAC overhead energy (fixed per conversion)
//
// If cfg.SingleLayer is true, only one layer is counted (input→output). Otherwise a 2-layer MLP
// (input→hidden→output) is assumed.
func EstimateInferenceEnergyJ(cfg *NetworkConfig, inputSize, hiddenSize, outputSize int) EnergyEstimate {
	est := EnergyEstimate{}
	if cfg == nil {
		cfg = DefaultNetworkConfig()
	}

	// Peripheral overhead
	dacJ := float64(inputSize) * EnergyPerDACJ
	adcCount := hiddenSize + outputSize
	if cfg.SingleLayer {
		adcCount = outputSize
	}
	adcJ := float64(adcCount) * EnergyPerADCJ
	est.ADCDACJ = dacJ + adcJ

	// MAC counts
	if cfg.SingleLayer {
		est.MACs1 = inputSize * outputSize
		est.MACs2 = 0
		est.TotalMACs = est.MACs1

		levels := cfg.NumLevels
		if cfg.PerLayerQuant {
			levels = cfg.Layer1Levels
		}
		est.Layer1Levels = levels
		est.Layer2Levels = 0
		est.EnergyPerMAC1J = EnergyPerMACJ(levels)
		est.EnergyPerMAC2J = 0

		est.ComputeJ = float64(est.MACs1) * est.EnergyPerMAC1J
		est.TotalJ = est.ComputeJ + est.ADCDACJ
		return est
	}

	// Two-layer network
	est.MACs1 = inputSize * hiddenSize
	est.MACs2 = hiddenSize * outputSize
	est.TotalMACs = est.MACs1 + est.MACs2

	l1Levels := cfg.NumLevels
	l2Levels := cfg.NumLevels
	if cfg.PerLayerQuant {
		l1Levels = cfg.Layer1Levels
		l2Levels = cfg.Layer2Levels
	}
	est.Layer1Levels = l1Levels
	est.Layer2Levels = l2Levels

	est.EnergyPerMAC1J = EnergyPerMACJ(l1Levels)
	est.EnergyPerMAC2J = EnergyPerMACJ(l2Levels)

	compute1 := float64(est.MACs1) * est.EnergyPerMAC1J
	compute2 := float64(est.MACs2) * est.EnergyPerMAC2J
	est.ComputeJ = compute1 + compute2
	est.TotalJ = est.ComputeJ + est.ADCDACJ
	return est
}

// EstimateInferenceEnergyMicroJ is a convenience wrapper returning total energy in µJ.
func EstimateInferenceEnergyMicroJ(cfg *NetworkConfig, inputSize, hiddenSize, outputSize int) float64 {
	return EstimateInferenceEnergyJ(cfg, inputSize, hiddenSize, outputSize).TotalJ * 1e6
}
