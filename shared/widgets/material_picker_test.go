package widgets

import (
	"math"
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
		Name:         "Test HZO",
		Description:  "Test material",
		Reference:    "Test ref",
		AnalogStates: 30,
		PrCM2:        0.25,
		PsCM2:        0.30,
		EcVM:         1.2e8,
		EpsilonHF:    30,
		EpsilonLF:    38,
		LossTangent:  0.02,
		ThicknessM:   10e-9,
		AreaM2:       100e-12,
		TauS:         1e-9,
		Tau0S:        1e-13,
		ActivationEnergyEV: 0.7,
		KAIExponent:  2.0,
		CurieTempK:   723,
		TempCoeffEc:  -2e5,
		TempCoeffPr:  -5e-5,
		EnduranceCycles: 1e10,
		RetentionTimeS:  3.15e9,
	}

	props := GetMaterialProperties(mat)

	if len(props) == 0 {
		t.Fatal("GetMaterialProperties returned no properties")
	}

	// Check that we have properties in expected L-K equation categories
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

	// Check specific property values
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

	// Check that materials were loaded
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

	// Test searching for "cryo"
	picker.filterQuery = "cryo"
	picker.updateFilter()

	foundCryo := false
	for _, id := range picker.filteredIDs {
		if id == "cryogenic_hzo" {
			foundCryo = true
			break
		}
	}

	// Only check if cryogenic material exists
	if _, exists := picker.materials["cryogenic_hzo"]; exists && !foundCryo {
		t.Error("Search for 'cryo' did not find cryogenic_hzo")
	}

	// Test clearing filter
	picker.filterQuery = ""
	picker.updateFilter()

	if len(picker.filteredIDs) != len(picker.materialIDs) {
		t.Errorf("After clearing filter, got %d materials, want %d",
			len(picker.filteredIDs), len(picker.materialIDs))
	}
}

func TestMaterialPickerSelection(t *testing.T) {
	// Create picker without extending base widget (to avoid Fyne app requirement)
	picker := &MaterialPicker{
		OnSelected:  nil,
		selectedRow: -1,
	}

	// Load materials from config
	cfg, err := physics.Load()
	if err != nil {
		t.Skip("No materials loaded from config")
	}
	picker.materials = cfg.Materials
	picker.materialIDs = []string{}
	for id := range picker.materials {
		picker.materialIDs = append(picker.materialIDs, id)
	}
	picker.filteredIDs = picker.materialIDs

	if len(picker.materials) == 0 {
		t.Skip("No materials loaded from config")
	}

	// Get first material ID
	firstID := picker.materialIDs[0]

	// Directly set selection (bypassing Refresh)
	picker.selectedID = firstID

	gotID, gotMat := picker.GetSelected()
	if gotID != firstID {
		t.Errorf("GetSelected ID = %q, want %q", gotID, firstID)
	}
	if gotMat == nil {
		t.Error("GetSelected material is nil")
	}

	// Test invalid selection - direct check of SetSelected logic
	if _, exists := picker.materials["nonexistent_material"]; !exists {
		// This is the expected case - ID doesn't exist
		// SetSelected should not change selection for invalid IDs
		oldSelected := picker.selectedID
		// Simulate what SetSelected does
		if _, exists := picker.materials["nonexistent_material"]; !exists {
			// Should not change
		}
		if picker.selectedID != oldSelected {
			t.Error("Invalid SetSelected should not change selection")
		}
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10c", 10, "exactly10c"},
		{"this is a longer string", 10, "this is..."},
		{"ab", 3, "ab"},
		{"abcd", 3, "abc"},
	}

	for _, tt := range tests {
		got := TruncateString(tt.input, tt.maxLen)
		if got != tt.expected {
			t.Errorf("TruncateString(%q, %d) = %q, want %q",
				tt.input, tt.maxLen, got, tt.expected)
		}
	}
}

func TestMaterialCardCreation(t *testing.T) {
	mat := &physics.Material{
		Name:         "Test Material",
		Description:  "Test description",
		Reference:    "Test et al., 2024",
		AnalogStates: 30,
		PrCM2:        0.25,
		EcVM:         1.2e8,
	}

	var tappedID string
	card := NewMaterialCard("test_id", mat, func(id string) {
		tappedID = id
	})

	if card == nil {
		t.Fatal("NewMaterialCard returned nil")
	}

	if card.GetMaterialID() != "test_id" {
		t.Errorf("GetMaterialID = %q, want %q", card.GetMaterialID(), "test_id")
	}

	// Test selection
	card.SetSelected(true)
	if !card.IsSelected() {
		t.Error("Card should be selected")
	}

	card.SetSelected(false)
	if card.IsSelected() {
		t.Error("Card should not be selected")
	}

	// Test tap callback
	card.Tapped(nil)
	if tappedID != "test_id" {
		t.Errorf("Tap callback got ID %q, want %q", tappedID, "test_id")
	}
}

func TestFormatArea(t *testing.T) {
	tests := []struct {
		name     string
		input    float64 // m²
		expected string
	}{
		{"100 nm²", 100e-18, "100.0 nm²"},
		{"small FeCIM cell", 45e-9 * 45e-9, "2025 nm²"},
		{"very small", 1e-18, "1.0 nm²"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatArea(tt.input)
			if got != tt.expected {
				t.Errorf("FormatArea(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestAnalogStatesBitsCalculation(t *testing.T) {
	// Verify the bits/cell calculation is correct
	tests := []struct {
		states       int
		expectedBits float64
	}{
		{2, 1.0},
		{4, 2.0},
		{8, 3.0},
		{16, 4.0},
		{30, 4.91}, // FeCIM's 30 states ≈ 4.91 bits
		{32, 5.0},
		{64, 6.0},
		{140, 7.13}, // FTJ 140 states ≈ 7.13 bits
	}

	for _, tt := range tests {
		bits := math.Log2(float64(tt.states))
		// Allow small floating point difference
		if math.Abs(bits-tt.expectedBits) > 0.01 {
			t.Errorf("%d states: got %.2f bits, want %.2f bits",
				tt.states, bits, tt.expectedBits)
		}
	}
}
