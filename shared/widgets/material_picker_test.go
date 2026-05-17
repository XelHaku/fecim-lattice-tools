//go:build legacy_fyne

package widgets

import (
	"strings"
	"testing"

	"fecim-lattice-tools/config/physics"
)

func TestFormatPolarization(t *testing.T) {
	tests := []struct {
		name     string
		input    float64 // C/m²
		expected string
	}{
		{"typical HZO", 0.25, "25.0 µC/cm²"},
		{"cryogenic enhanced", 0.75, "75.0 µC/cm²"},
		{"AlScN high Pr", 1.20, "120 µC/cm²"},
		{"low Pr", 0.10, "10.0 µC/cm²"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatPolarization(tt.input)
			if got != tt.expected {
				t.Errorf("FormatPolarization(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormatField(t *testing.T) {
	tests := []struct {
		name     string
		input    float64 // V/m
		expected string
	}{
		{"typical HZO", 1.2e8, "1.2 MV/cm"},
		{"low Ec superlattice", 0.85e8, "0.85 MV/cm"},
		{"high Ec AlScN", 5.0e8, "5.0 MV/cm"},
		{"cryogenic", 1.5e8, "1.5 MV/cm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatField(tt.input)
			if got != tt.expected {
				t.Errorf("FormatField(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormatThickness(t *testing.T) {
	tests := []struct {
		name     string
		input    float64 // m
		expected string
	}{
		{"10nm standard", 10e-9, "10 nm"},
		{"4.5nm FTJ", 4.5e-9, "4.5 nm"},
		{"20nm thick", 20e-9, "20 nm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatThickness(tt.input)
			if got != tt.expected {
				t.Errorf("FormatThickness(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name     string
		input    float64 // seconds
		expected string
	}{
		{"nanosecond", 1e-9, "1.0 ns"},
		{"10 nanoseconds", 10e-9, "10.0 ns"},
		{"picosecond", 360e-12, "360 ps"},
		{"100 years", 3.15e9, "100 years"},
		{"10 years", 3.15e8, "10.0 years"},
		{"microsecond", 1e-6, "1.0 µs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTime(tt.input)
			if got != tt.expected {
				t.Errorf("FormatTime(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormatEndurance(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"10^9 cycles", 1e9, "10^9 cycles"},
		{"10^10 cycles", 1e10, "10^10 cycles"},
		{"10^12 cycles", 1e12, "10^12 cycles"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatEndurance(tt.input)
			if got != tt.expected {
				t.Errorf("FormatEndurance(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormatTemperature(t *testing.T) {
	tests := []struct {
		name     string
		input    float64 // K
		expected string
	}{
		{"room temperature", 300, "300 K (27°C)"},
		{"Curie temperature", 723, "723 K (450°C)"},
		{"cryogenic", 4, "4 K (-269°C)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTemperature(tt.input)
			if got != tt.expected {
				t.Errorf("FormatTemperature(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetMaterialProperties(t *testing.T) {
	mat := &physics.Material{
		Name:               "Test HZO",
		Description:        "Test material",
		Reference:          "Test ref",
		AnalogStates:       30,
		PrCM2:              0.25,
		PsCM2:              0.30,
		EcVM:               1.2e8,
		EpsilonHF:          30,
		EpsilonLF:          38,
		LossTangent:        0.02,
		ThicknessM:         10e-9,
		AreaM2:             100e-12,
		TauS:               1e-9,
		Tau0S:              1e-13,
		ActivationEnergyEV: 0.7,
		KAIExponent:        2.0,
		CurieTempK:         723,
		TempCoeffEc:        -2e5,
		TempCoeffPr:        -5e-5,
		EnduranceCycles:    1e10,
		RetentionTimeS:     3.15e9,
	}

	props := GetMaterialProperties(mat)

	if len(props) == 0 {
		t.Fatal("GetMaterialProperties returned no properties")
	}

	categories := make(map[string]int)
	for _, p := range props {
		categories[p.Category]++
	}

	expectedCategories := []string{
		CategoryCore,
		CategoryGeometry,
		CategoryAlpha,
	}

	for _, cat := range expectedCategories {
		if categories[cat] == 0 {
			t.Errorf("No properties found in category %q", cat)
		}
	}

	var prProp *FormattedProperty
	for _, p := range props {
		if p.Name == "Pr (Remanent)" {
			prProp = &p
			break
		}
	}

	if prProp == nil {
		t.Fatal("Pr property not found")
	}

	if prProp.Value != "25.0 µC/cm²" {
		t.Errorf("Pr value = %q, want %q", prProp.Value, "25.0 µC/cm²")
	}
}

func TestGetPropertiesByCategory(t *testing.T) {
	props := []FormattedProperty{
		{Name: "Pr", Category: CategoryCore},
		{Name: "Ps", Category: CategoryCore},
		{Name: "Ec", Category: CategoryCore},
		{Name: "Thickness", Category: CategoryGeometry},
	}

	coreProps := GetPropertiesByCategory(props, CategoryCore)
	if len(coreProps) != 3 {
		t.Errorf("Got %d core properties, want 3", len(coreProps))
	}

	geoProps := GetPropertiesByCategory(props, CategoryGeometry)
	if len(geoProps) != 1 {
		t.Errorf("Got %d geometry properties, want 1", len(geoProps))
	}

	landauProps := GetPropertiesByCategory(props, CategoryLandau)
	if len(landauProps) != 0 {
		t.Errorf("Got %d Landau properties, want 0", len(landauProps))
	}
}

func TestMaterialPickerCreation(t *testing.T) {
	picker := NewMaterialPicker(nil)

	if picker == nil {
		t.Fatal("NewMaterialPicker returned nil")
	}

	if len(picker.materials) == 0 {
		t.Skip("No materials loaded from config - config may not be available in test environment")
	}

	if len(picker.materialIDs) == 0 {
		t.Error("No material IDs populated")
	}
}

func TestMaterialPickerSearch(t *testing.T) {
	picker := NewMaterialPicker(nil)

	if len(picker.materials) == 0 {
		t.Skip("No materials loaded from config")
	}

	picker.filterQuery = "cryo"
	picker.updateFilter()

	foundCryo := false
	for _, id := range picker.filteredIDs {
		if strings.Contains(strings.ToLower(id), "cryo") || strings.Contains(strings.ToLower(picker.materials[id].Name), "cryo") {
			foundCryo = true
			break
		}
	}

	if !foundCryo && len(picker.filteredIDs) > 0 {
		t.Logf("No explicit cryo material found in current config, filtered count=%d", len(picker.filteredIDs))
	}
}

func TestEngineSupportTag(t *testing.T) {
	pOnly := &physics.Material{PrCM2: 0.2, PsCM2: 0.3, EcVM: 1e8}
	if got := engineSupportTag(pOnly); got != "[P]" {
		t.Fatalf("engineSupportTag(pOnly)=%q want [P]", got)
	}

	lkOnly := &physics.Material{Thermodynamics: physics.MaterialThermodynamics{BetaLandau: 1, GammaLandau: 1, RhoViscosity: 1}}
	if got := engineSupportTag(lkOnly); got != "[LK]" {
		t.Fatalf("engineSupportTag(lkOnly)=%q want [LK]", got)
	}

	both := &physics.Material{PrCM2: 0.2, PsCM2: 0.3, EcVM: 1e8, Thermodynamics: physics.MaterialThermodynamics{BetaLandau: 1, GammaLandau: 1, RhoViscosity: 1}}
	if got := engineSupportTag(both); got != "[P,LK]" {
		t.Fatalf("engineSupportTag(both)=%q want [P,LK]", got)
	}
}

func TestMaterialColumnsIncludeEngineSupport(t *testing.T) {
	foundEng := false
	foundBeta := false
	for _, c := range materialColumns {
		if c.Name == "Eng" {
			foundEng = true
		}
		if c.Name == "β" {
			foundBeta = true
		}
	}
	if !foundEng || !foundBeta {
		t.Fatalf("expected columns Eng and β, got foundEng=%v foundBeta=%v", foundEng, foundBeta)
	}
}
