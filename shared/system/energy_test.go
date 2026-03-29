package system

import (
	"math"
	"testing"
)

func TestMACEnergyJ_BasicScaling(t *testing.T) {
	// 2-level (1 bit) → 10 fJ
	e1 := MACEnergyJ(2)
	if math.Abs(e1-10e-15) > 1e-18 {
		t.Errorf("2 levels: got %.3e J, want 10e-15 J", e1)
	}

	// 4-level (2 bits) → 20 fJ
	e2 := MACEnergyJ(4)
	if math.Abs(e2-20e-15) > 1e-18 {
		t.Errorf("4 levels: got %.3e J, want 20e-15 J", e2)
	}

	// 16-level (4 bits) → 40 fJ
	e4 := MACEnergyJ(16)
	if math.Abs(e4-40e-15) > 1e-18 {
		t.Errorf("16 levels: got %.3e J, want 40e-15 J", e4)
	}
}

func TestMACEnergyJ_ZeroAndNegative(t *testing.T) {
	// Zero and negative levels should clamp to 1-bit minimum
	e0 := MACEnergyJ(0)
	eN := MACEnergyJ(-5)
	want := EnergyPerBitPerMACJ // 1 bit minimum

	if math.Abs(e0-want) > 1e-20 {
		t.Errorf("levels=0: got %.3e, want %.3e", e0, want)
	}
	if math.Abs(eN-want) > 1e-20 {
		t.Errorf("levels=-5: got %.3e, want %.3e", eN, want)
	}
}

func TestEstimateMLPEnergyJ_SingleLayer(t *testing.T) {
	// 784 → 10 (MNIST single-layer)
	cfg := MLPEnergyConfig{
		LayerSizes:    []int{784, 10},
		DefaultLevels: 30,
	}
	est := EstimateMLPEnergyJ(cfg)

	expectedMACs := 784 * 10
	if est.TotalMACs != expectedMACs {
		t.Errorf("TotalMACs: got %d, want %d", est.TotalMACs, expectedMACs)
	}
	if est.TotalJ <= 0 {
		t.Errorf("TotalJ must be positive, got %.3e", est.TotalJ)
	}
	// ADC + DAC must be included
	if est.ADCDACJ <= 0 {
		t.Errorf("ADCDACJ must be positive, got %.3e", est.ADCDACJ)
	}
	// DAC: 784 inputs; ADC: 10 outputs
	wantADCDAC := float64(784)*DefaultEnergyPerDACJ + float64(10)*DefaultEnergyPerADCJ
	if math.Abs(est.ADCDACJ-wantADCDAC) > 1e-18 {
		t.Errorf("ADCDACJ: got %.4e, want %.4e", est.ADCDACJ, wantADCDAC)
	}
}

func TestEstimateMLPEnergyJ_TwoLayer(t *testing.T) {
	// 784 → 128 → 10
	cfg := MLPEnergyConfig{
		LayerSizes:    []int{784, 128, 10},
		DefaultLevels: 30,
	}
	est := EstimateMLPEnergyJ(cfg)

	wantMACs := 784*128 + 128*10
	if est.TotalMACs != wantMACs {
		t.Errorf("TotalMACs: got %d, want %d", est.TotalMACs, wantMACs)
	}
	if len(est.LayerMACs) != 2 {
		t.Errorf("LayerMACs length: got %d, want 2", len(est.LayerMACs))
	}
	if est.LayerMACs[0] != 784*128 {
		t.Errorf("Layer0 MACs: got %d, want %d", est.LayerMACs[0], 784*128)
	}
}

func TestEstimateMLPEnergyJ_PerLayerLevels(t *testing.T) {
	cfg := MLPEnergyConfig{
		LayerSizes:     []int{100, 50, 10},
		LevelsPerLayer: []int{16, 4},
	}
	est := EstimateMLPEnergyJ(cfg)

	// Layer 0: 16 levels → 4 bits × 10 fJ/bit
	wantL0 := float64(100*50) * MACEnergyJ(16)
	if math.Abs(est.LayerJ[0]-wantL0) > 1e-18 {
		t.Errorf("Layer0 energy: got %.4e, want %.4e", est.LayerJ[0], wantL0)
	}

	// Layer 1: 4 levels → 2 bits × 10 fJ/bit
	wantL1 := float64(50*10) * MACEnergyJ(4)
	if math.Abs(est.LayerJ[1]-wantL1) > 1e-18 {
		t.Errorf("Layer1 energy: got %.4e, want %.4e", est.LayerJ[1], wantL1)
	}
}

func TestEstimateMLPEnergyJ_TotalConsistency(t *testing.T) {
	cfg := MLPEnergyConfig{
		LayerSizes:    []int{256, 64, 16},
		DefaultLevels: 30,
	}
	est := EstimateMLPEnergyJ(cfg)

	// TotalJ must equal ComputeJ + ADCDACJ
	if math.Abs(est.TotalJ-(est.ComputeJ+est.ADCDACJ)) > 1e-20 {
		t.Errorf("TotalJ (%.4e) != ComputeJ (%.4e) + ADCDACJ (%.4e)",
			est.TotalJ, est.ComputeJ, est.ADCDACJ)
	}

	// ComputeJ must equal sum of LayerJ
	sumLayerJ := 0.0
	for _, lj := range est.LayerJ {
		sumLayerJ += lj
	}
	if math.Abs(est.ComputeJ-sumLayerJ) > 1e-20 {
		t.Errorf("ComputeJ (%.4e) != sum(LayerJ) (%.4e)", est.ComputeJ, sumLayerJ)
	}
}
