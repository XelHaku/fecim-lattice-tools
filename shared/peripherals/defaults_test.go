package peripherals

import (
	"math"
	"testing"
)

func TestDACConstants(t *testing.T) {
	if DefaultBits != 5 {
		t.Errorf("DefaultBits = %d, want 5", DefaultBits)
	}
	if DACVrefHigh != 1.5 {
		t.Errorf("DACVrefHigh = %v, want 1.5", DACVrefHigh)
	}
	if DACVrefLow != -1.5 {
		t.Errorf("DACVrefLow = %v, want -1.5", DACVrefLow)
	}
	if DACSettleTime != 10.0 {
		t.Errorf("DACSettleTime = %v, want 10.0", DACSettleTime)
	}
}

func TestADCConstants(t *testing.T) {
	if ADCVrefHigh != 1.0 {
		t.Errorf("ADCVrefHigh = %v, want 1.0", ADCVrefHigh)
	}
	if ADCVrefLow != 0.0 {
		t.Errorf("ADCVrefLow = %v, want 0.0", ADCVrefLow)
	}
	if ADCConversionTime != 50.0 {
		t.Errorf("ADCConversionTime = %v, want 50.0", ADCConversionTime)
	}
}

func TestDefaultDACConfig(t *testing.T) {
	cfg := DefaultDACConfig()

	if cfg.Bits != 5 {
		t.Errorf("DAC Bits = %d, want 5", cfg.Bits)
	}
	if cfg.VrefHigh != 1.5 {
		t.Errorf("DAC VrefHigh = %v, want 1.5", cfg.VrefHigh)
	}
	if cfg.VrefLow != -1.5 {
		t.Errorf("DAC VrefLow = %v, want -1.5", cfg.VrefLow)
	}
	if cfg.INL != 0.5 {
		t.Errorf("DAC INL = %v, want 0.5", cfg.INL)
	}
	if cfg.DNL != 0.25 {
		t.Errorf("DAC DNL = %v, want 0.25", cfg.DNL)
	}
	if cfg.SettleTime != 10.0 {
		t.Errorf("DAC SettleTime = %v, want 10.0", cfg.SettleTime)
	}
}

func TestDefaultADCConfig(t *testing.T) {
	cfg := DefaultADCConfig()

	if cfg.Bits != 5 {
		t.Errorf("ADC Bits = %d, want 5", cfg.Bits)
	}
	if cfg.VrefHigh != 1.0 {
		t.Errorf("ADC VrefHigh = %v, want 1.0", cfg.VrefHigh)
	}
	if cfg.VrefLow != 0.0 {
		t.Errorf("ADC VrefLow = %v, want 0.0", cfg.VrefLow)
	}
	if cfg.INL != 0.5 {
		t.Errorf("ADC INL = %v, want 0.5", cfg.INL)
	}
	if cfg.DNL != 0.25 {
		t.Errorf("ADC DNL = %v, want 0.25", cfg.DNL)
	}
	if cfg.ConversionTime != 50.0 {
		t.Errorf("ADC ConversionTime = %v, want 50.0", cfg.ConversionTime)
	}
	if cfg.Type != ADCTypeSAR {
		t.Errorf("ADC Type = %v, want SAR", cfg.Type)
	}
}

func TestADCTypeString(t *testing.T) {
	tests := []struct {
		adcType  ADCType
		expected string
	}{
		{ADCTypeSAR, "SAR"},
		{ADCTypeFlash, "Flash"},
		{ADCTypeSigmaDelta, "Sigma-Delta"},
		{ADCType(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.adcType.String(); got != tt.expected {
			t.Errorf("ADCType(%d).String() = %q, want %q", tt.adcType, got, tt.expected)
		}
	}
}

func TestDACConfigResolution(t *testing.T) {
	cfg := DefaultDACConfig()
	// 5-bit: 32 levels, range = 3V, LSB = 3/31 ≈ 96.77 mV
	expected := 3.0 / 31.0
	if math.Abs(cfg.Resolution()-expected) > 1e-10 {
		t.Errorf("DAC Resolution = %v, want %v", cfg.Resolution(), expected)
	}
}

func TestADCConfigResolution(t *testing.T) {
	cfg := DefaultADCConfig()
	// 5-bit: 32 levels, range = 1V, LSB = 1/31 ≈ 32.26 mV
	expected := 1.0 / 31.0
	if math.Abs(cfg.Resolution()-expected) > 1e-10 {
		t.Errorf("ADC Resolution = %v, want %v", cfg.Resolution(), expected)
	}
}

func TestDACConfigLevels(t *testing.T) {
	cfg := DefaultDACConfig()
	if cfg.Levels() != 32 {
		t.Errorf("DAC Levels = %d, want 32", cfg.Levels())
	}
}

func TestADCConfigLevels(t *testing.T) {
	cfg := DefaultADCConfig()
	if cfg.Levels() != 32 {
		t.Errorf("ADC Levels = %d, want 32", cfg.Levels())
	}
}

func TestDefaultTIAConfig(t *testing.T) {
	cfg := DefaultTIAConfig()

	if cfg.Gain != 10e3 {
		t.Errorf("TIA Gain = %v, want 10e3", cfg.Gain)
	}
	if cfg.Bandwidth != 100e6 {
		t.Errorf("TIA Bandwidth = %v, want 100e6", cfg.Bandwidth)
	}
	if cfg.Noise != 100e-6 {
		t.Errorf("TIA Noise = %v, want 100e-6", cfg.Noise)
	}
	if cfg.Offset != 5e-3 {
		t.Errorf("TIA Offset = %v, want 5e-3", cfg.Offset)
	}
}

func TestTIASettlingTime(t *testing.T) {
	cfg := DefaultTIAConfig()
	// 0.35 / 100MHz = 3.5 ns
	expected := 0.35 / 100e6
	if math.Abs(cfg.SettlingTime()-expected) > 1e-15 {
		t.Errorf("TIA SettlingTime = %v, want %v", cfg.SettlingTime(), expected)
	}
}

func TestFeCIMLevelsConstant(t *testing.T) {
	if FeCIMLevels != 30 {
		t.Errorf("FeCIMLevels = %d, want 30", FeCIMLevels)
	}
}
