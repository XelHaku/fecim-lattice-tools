//go:build legacy_fyne

package gui

import (
	"math"
	"testing"

	"fyne.io/fyne/v2/test"

	"fecim-lattice-tools/module3-mnist/pkg/core"
)

func TestEnergyWidget_RecordInference_UsesCoreEnergyModel(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	ew := NewEnergyWidget(784, 128, 10)
	cfg := core.DefaultNetworkConfig()
	cfg.NumLevels = 30

	ew.RecordInference(cfg)

	est := core.EstimateInferenceEnergyJ(cfg, 784, 128, 10)
	if math.Abs(ew.totalEnergyFeCIM-est.TotalJ) > 1e-21 {
		t.Fatalf("FeCIM total energy mismatch: got %.18g J want %.18g J", ew.totalEnergyFeCIM, est.TotalJ)
	}

	expectedGPU := float64(est.TotalMACs) * ew.config.EnergyPerMAC_GPU
	if math.Abs(ew.totalEnergyGPU-expectedGPU) > 1e-18 {
		t.Fatalf("GPU total energy mismatch: got %.18g J want %.18g J", ew.totalEnergyGPU, expectedGPU)
	}
}

func TestEnergyWidget_RecordInference_SingleLayerMACs(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	ew := NewEnergyWidget(784, 128, 10)
	cfg := core.DefaultNetworkConfig()
	cfg.SingleLayer = true

	ew.RecordInference(cfg)
	est := core.EstimateInferenceEnergyJ(cfg, 784, 128, 10)
	expectedGPU := float64(est.TotalMACs) * ew.config.EnergyPerMAC_GPU

	if est.TotalMACs != 784*10 {
		t.Fatalf("single-layer MACs=%d want %d", est.TotalMACs, 784*10)
	}
	if math.Abs(ew.totalEnergyGPU-expectedGPU) > 1e-18 {
		t.Fatalf("single-layer GPU mismatch: got %.18g want %.18g", ew.totalEnergyGPU, expectedGPU)
	}
}
