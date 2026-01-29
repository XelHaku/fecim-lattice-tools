package widgets

import (
	"strings"
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

func TestGetMaterialPropertiesPreisachIndicators(t *testing.T) {
	// Create a test material with all relevant parameters
	mat := &physics.Material{
		Name:               "Test Material",
		PrCM2:              0.25,
		PsCM2:              0.30,
		EcVM:               1e8,
		ThicknessM:         10e-9,
		TauS:               1e-9,
		Tau0S:              1e-12,
		ActivationEnergyEV: 0.5,
		KAIExponent:        2.0,
		CurieTempK:         700,
		GmaxGminRatio:      1000,
	}

	props := GetMaterialProperties(mat)

	// Map of property names that should have Preisach indicator
	expectedPreisach := map[string]bool{
		"Remanent Polarization (Pr)":   true,
		"Saturation Polarization (Ps)": true,
		"Coercive Field (Ec)":          true,
		"Film Thickness":               true,
		"Switching Time (τ)":           true,
		"Attempt Time (τ₀)":            true,
		"Activation Energy":            true,
		"KAI Exponent":                 true,
		"Curie Temperature":            true,
		"Gmax/Gmin Ratio":              true,
	}

	for _, prop := range props {
		shouldHavePreisach := expectedPreisach[prop.Name]
		hasPreisach := prop.Models.Preisach

		if shouldHavePreisach && !hasPreisach {
			t.Errorf("Property %q should have Preisach indicator but doesn't", prop.Name)
		}
		if !shouldHavePreisach && hasPreisach {
			t.Errorf("Property %q should NOT have Preisach indicator but does", prop.Name)
		}
	}
}

func TestGetMaterialPropertiesDescriptions(t *testing.T) {
	mat := &physics.Material{
		Name:        "Test Material",
		PrCM2:       0.25,
		PsCM2:       0.30,
		EcVM:        1e8,
		ThicknessM:  10e-9,
		TauS:        1e-9,
		CurieTempK:  700,
		KAIExponent: 2.0,
	}

	props := GetMaterialProperties(mat)

	// Check that all properties have descriptions
	for _, prop := range props {
		if prop.Description == "" {
			t.Errorf("Property %q has empty description", prop.Name)
		}

		// Preisach-model properties should mention Preisach in their description
		if prop.Models.Preisach {
			if !strings.Contains(strings.ToLower(prop.Description), "preisach") {
				t.Errorf("Property %q is marked as Preisach parameter but description doesn't mention Preisach: %s",
					prop.Name, prop.Description)
			}
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
		Category:    CategoryPolarization,
		Description: "Test description",
		Models:      ModelUsage{Preisach: true},
	}

	if !prop.Models.Preisach {
		t.Error("Property should have Preisach=true")
	}
	if prop.Models.String() != "[P]" {
		t.Errorf("Models.String() = %q, want %q", prop.Models.String(), "[P]")
	}
}
