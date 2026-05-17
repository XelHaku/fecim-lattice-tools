//go:build legacy_fyne

package gui

import (
	"testing"
)

// TestEnergyConstants verifies CPUEnergy=1000, GPUEnergy=100, FeCIMEnergy=1
func TestEnergyConstants(t *testing.T) {
	if cpuEnergyPJ != 1000.0 {
		t.Errorf("cpuEnergyPJ = %f, want 1000.0", cpuEnergyPJ)
	}
	if gpuEnergyPJ != 100.0 {
		t.Errorf("gpuEnergyPJ = %f, want 100.0", gpuEnergyPJ)
	}
	if fecimEnergyPJ != 1.0 {
		t.Errorf("fecimEnergyPJ = %f, want 1.0", fecimEnergyPJ)
	}
}

// TestEnergyRatios verifies CPUEnergy/FeCIMEnergy=1000, GPUEnergy/FeCIMEnergy=100
func TestEnergyRatios(t *testing.T) {
	cpuRatio := cpuEnergyPJ / fecimEnergyPJ
	if cpuRatio != 1000.0 {
		t.Errorf("CPU/FeCIM energy ratio = %f, want 1000.0", cpuRatio)
	}

	gpuRatio := gpuEnergyPJ / fecimEnergyPJ
	if gpuRatio != 100.0 {
		t.Errorf("GPU/FeCIM energy ratio = %f, want 100.0", gpuRatio)
	}
}

// TestNewAnimatedEnergyRace verifies NewAnimatedEnergyRace returns non-nil
func TestNewAnimatedEnergyRace(t *testing.T) {
	race := NewAnimatedEnergyRace()
	if race == nil {
		t.Error("NewAnimatedEnergyRace returned nil")
	}
}

// TestEstimatedColorDefined verifies estimatedColor is defined and is amber (R=255, G~191, B=0)
func TestEstimatedColorDefined(t *testing.T) {
	// Check RGB values for amber
	if estimatedColor.R != 255 {
		t.Errorf("estimatedColor.R = %d, want 255", estimatedColor.R)
	}
	if estimatedColor.G != 191 {
		t.Errorf("estimatedColor.G = %d, want 191", estimatedColor.G)
	}
	if estimatedColor.B != 0 {
		t.Errorf("estimatedColor.B = %d, want 0", estimatedColor.B)
	}
	if estimatedColor.A != 255 {
		t.Errorf("estimatedColor.A = %d, want 255 (fully opaque)", estimatedColor.A)
	}
}
