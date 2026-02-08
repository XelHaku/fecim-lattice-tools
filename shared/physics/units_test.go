package physics

import "testing"

func TestFormatEnergy(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0 J"},
		{-1, "0 J"},
		{1e-15, "1.00 fJ"},
		{1.5e-15, "1.50 fJ"},
		{1e-12, "1.00 pJ"},
		{2.3e-12, "2.30 pJ"},
		{1e-9, "1.00 nJ"},
		{4.5e-9, "4.50 nJ"},
		{1e-6, "1.00 µJ"},
		{6.7e-6, "6.70 µJ"},
		{1e-3, "1.00 mJ"},
		{8.9e-3, "8.90 mJ"},
		{1.0, "1.00 J"},
		{1.2, "1.20 J"},
		{1000, "1000.00 J"},
	}

	for _, tt := range tests {
		result := FormatEnergy(tt.input)
		if result != tt.expected {
			t.Errorf("FormatEnergy(%e) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatEnergyMJ(t *testing.T) {
	// 1 mJ = 1e-3 J
	result := FormatEnergyMJ(1.0)
	expected := "1.00 mJ"
	if result != expected {
		t.Errorf("FormatEnergyMJ(1.0) = %q, want %q", result, expected)
	}

	// 0.001 mJ = 1 µJ
	result = FormatEnergyMJ(0.001)
	expected = "1.00 µJ"
	if result != expected {
		t.Errorf("FormatEnergyMJ(0.001) = %q, want %q", result, expected)
	}
}

func TestFormatEnergyUJ(t *testing.T) {
	// 1 µJ = 1e-6 J
	result := FormatEnergyUJ(1.0)
	expected := "1.00 µJ"
	if result != expected {
		t.Errorf("FormatEnergyUJ(1.0) = %q, want %q", result, expected)
	}

	// 1000 µJ = 1 mJ
	result = FormatEnergyUJ(1000)
	expected = "1.00 mJ"
	if result != expected {
		t.Errorf("FormatEnergyUJ(1000) = %q, want %q", result, expected)
	}
}

func TestFormatConductance(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0 S"},
		{1e-9, "1.00 nS"},
		{50e-6, "50.00 µS"},
		{1e-3, "1.00 mS"},
		{1.0, "1.00 S"},
	}

	for _, tt := range tests {
		result := FormatConductance(tt.input)
		if result != tt.expected {
			t.Errorf("FormatConductance(%e) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatCurrent(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0 A"},
		{1e-12, "1.00 pA"},
		{50e-9, "50.00 nA"},
		{1e-6, "1.00 µA"},
		{1e-3, "1.00 mA"},
		{1.0, "1.00 A"},
	}

	for _, tt := range tests {
		result := FormatCurrent(tt.input)
		if result != tt.expected {
			t.Errorf("FormatCurrent(%e) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatVoltage(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0 V"},
		{1e-6, "1.00 µV"},
		{1e-3, "1.00 mV"},
		{1.5, "1.50 V"},
	}

	for _, tt := range tests {
		result := FormatVoltage(tt.input)
		if result != tt.expected {
			t.Errorf("FormatVoltage(%e) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0 s"},
		{1e-12, "1.00 ps"},
		{1e-9, "1.00 ns"},
		{1e-6, "1.00 µs"},
		{1e-3, "1.00 ms"},
		{1.0, "1.00 s"},
	}

	for _, tt := range tests {
		result := FormatTime(tt.input)
		if result != tt.expected {
			t.Errorf("FormatTime(%e) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatFrequency(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0 Hz"},
		{100, "100.00 Hz"},
		{1e3, "1.00 kHz"},
		{1e6, "1.00 MHz"},
		{1e9, "1.00 GHz"},
	}

	for _, tt := range tests {
		result := FormatFrequency(tt.input)
		if result != tt.expected {
			t.Errorf("FormatFrequency(%e) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatResistance(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0 Ω"},
		{-1, "0 Ω"},
		{0.001, "1.00 mΩ"},
		{0.5, "500.00 mΩ"},
		{1, "1.00 Ω"},
		{100, "100.00 Ω"},
		{4700, "4.70 kΩ"},
		{1e6, "1.00 MΩ"},
		{10e6, "10.00 MΩ"},
		{1e9, "1.00 GΩ"},
		{4.7e9, "4.70 GΩ"},
	}

	for _, tt := range tests {
		result := FormatResistance(tt.input)
		if result != tt.expected {
			t.Errorf("FormatResistance(%e) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatCapacitance(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0 F"},
		{-1, "0 F"},
		{1e-18, "1.00 aF"},
		{50e-18, "50.00 aF"},
		{1e-15, "1.00 fF"},
		{100e-15, "100.00 fF"},
		{1e-12, "1.00 pF"},
		{47e-12, "47.00 pF"},
		{1e-9, "1.00 nF"},
		{100e-9, "100.00 nF"},
		{1e-6, "1.00 µF"},
		{10e-6, "10.00 µF"},
		{1e-3, "1.00 mF"},
		{1, "1.00 F"},
	}

	for _, tt := range tests {
		result := FormatCapacitance(tt.input)
		if result != tt.expected {
			t.Errorf("FormatCapacitance(%e) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatPower(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0 W"},
		{-1, "0 W"},
		{1e-15, "1.00 fW"},
		{50e-15, "50.00 fW"},
		{1e-12, "1.00 pW"},
		{1e-9, "1.00 nW"},
		{1e-6, "1.00 µW"},
		{50e-6, "50.00 µW"},
		{1e-3, "1.00 mW"},
		{500e-3, "500.00 mW"},
		{1, "1.00 W"},
		{1.5, "1.50 W"},
		{100, "100.00 W"},
		{1500, "1.50 kW"},
		{10000, "10.00 kW"},
	}

	for _, tt := range tests {
		result := FormatPower(tt.input)
		if result != tt.expected {
			t.Errorf("FormatPower(%e) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatCharge(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0 C"},
		{-1, "0 C"},
		{1e-15, "1.00 fC"},
		{100e-15, "100.00 fC"},
		{1e-12, "1.00 pC"},
		{50e-12, "50.00 pC"},
		{1e-9, "1.00 nC"},
		{1e-6, "1.00 µC"},
		{1e-3, "1.00 mC"},
		{1, "1.00 C"},
	}

	for _, tt := range tests {
		result := FormatCharge(tt.input)
		if result != tt.expected {
			t.Errorf("FormatCharge(%e) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatPolarization(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0 µC/cm²"},
		{-0.1, "0 µC/cm²"},
		{0.001, "0.10 µC/cm²"},    // 0.1 µC/cm²
		{0.05, "5.00 µC/cm²"},     // 5 µC/cm²
		{0.20, "20.0 µC/cm²"},     // 20 µC/cm² (typical HZO Pr)
		{0.35, "35.0 µC/cm²"},     // 35 µC/cm² (high performance)
		{1.0, "100 µC/cm²"},       // 100 µC/cm² (very high)
		{1.5, "150 µC/cm²"},       // 150 µC/cm²
	}

	for _, tt := range tests {
		result := FormatPolarization(tt.input)
		if result != tt.expected {
			t.Errorf("FormatPolarization(%e) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatElectricField(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0 V/m"},
		{-1e8, "0 V/m"},
		{1e2, "1.00 V/cm"},        // 1 V/cm (very low field)
		{1e5, "1.00 kV/cm"},       // 1 kV/cm
		{5e5, "5.00 kV/cm"},       // 5 kV/cm
		{1e8, "1.00 MV/cm"},       // 1 MV/cm (typical Ec for HZO)
		{1.5e8, "1.50 MV/cm"},     // 1.5 MV/cm
		{3e8, "3.00 MV/cm"},       // 3 MV/cm (high field)
	}

	for _, tt := range tests {
		result := FormatElectricField(tt.input)
		if result != tt.expected {
			t.Errorf("FormatElectricField(%e) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
