//go:build legacy_fyne

package widgets

import (
	"testing"

	"fecim-lattice-tools/config/physics"
)

func TestModelUsageString(t *testing.T) {
	tests := []struct {
		name     string
		usage    ModelUsage
		expected string
	}{
		{
			name:     "Preisach only",
			usage:    ModelUsage{Preisach: true},
			expected: "[P]",
		},
		{
			name:     "No models",
			usage:    ModelUsage{Preisach: false},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.usage.String()
			if result != tt.expected {
				t.Errorf("ModelUsage.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetMaterialPropertiesModelIndicators(t *testing.T) {
	// Create a test material with L-K relevant parameters
	mat := &physics.Material{
		Name:       "Test Material",
		PrCM2:      0.25,
		PsCM2:      0.30,
		EcVM:       1e8,
		ThicknessM: 10e-9,
		CurieTempK: 700,
		Thermodynamics: physics.MaterialThermodynamics{
			BetaLandau:   -2.16e8,
			GammaLandau:  1.653e10,
			RhoViscosity: 0.05,
			CurieConstK:  1.5e5,
		},
		Conductance: physics.MaterialConductance{
			GminS: 1e-6,
			GmaxS: 100e-6,
		},
	}

	props := GetMaterialProperties(mat)

	// Map of property names that should have L-K indicator
	expectedLK := map[string]bool{
		"d (Thickness)":   true,
		"A (Area)":        true,
		"β (Landau)":      true,
		"γ (Landau)":      true,
		"ρ (Viscosity)":   true,
		"Tc (Curie)":      true,
		"C (Curie Const)": true,
	}

	for _, prop := range props {
		shouldHaveLK := expectedLK[prop.Name]
		hasLK := prop.Models.LandauKh

		if shouldHaveLK && !hasLK {
			t.Errorf("Property %q should have L-K indicator but doesn't", prop.Name)
		}
	}
}

func TestGetMaterialPropertiesDescriptions(t *testing.T) {
	mat := &physics.Material{
		Name:       "Test Material",
		PrCM2:      0.25,
		PsCM2:      0.30,
		EcVM:       1e8,
		ThicknessM: 10e-9,
		CurieTempK: 700,
		Thermodynamics: physics.MaterialThermodynamics{
			BetaLandau:  -2.16e8,
			GammaLandau: 1.653e10,
		},
	}

	props := GetMaterialProperties(mat)

	// Check that all properties have descriptions
	for _, prop := range props {
		if prop.Description == "" {
			t.Errorf("Property %q has empty description", prop.Name)
		}
	}
}

func TestPreisachModelVariable(t *testing.T) {
	// Ensure preisachModel helper is set correctly
	if !preisachModel.Preisach {
		t.Error("preisachModel should have Preisach=true")
	}
	if preisachModel.String() != "[P]" {
		t.Errorf("preisachModel.String() = %q, want %q", preisachModel.String(), "[P]")
	}
}

func TestFormattedPropertyModelField(t *testing.T) {
	// Test that FormattedProperty correctly stores Models
	prop := FormattedProperty{
		Name:        "Test Property",
		Value:       "123",
		RawValue:    123.0,
		Category:    CategoryCore,
		Description: "Test description",
		Models:      ModelUsage{LandauKh: true, Preisach: true},
	}

	if !prop.Models.LandauKh {
		t.Error("Property should have LandauKh=true")
	}
	if !prop.Models.Preisach {
		t.Error("Property should have Preisach=true")
	}
	if prop.Models.String() != "[L+P]" {
		t.Errorf("Models.String() = %q, want %q", prop.Models.String(), "[L+P]")
	}
}
