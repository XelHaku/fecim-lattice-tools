package peripherals

import (
	"math"
	"testing"
)

func TestDACConstants(t *testing.T) {
	if DefaultBits != 6 {
		t.Errorf("DefaultBits = %d, want 6 (ADC default for 30-level resolution)", DefaultBits)
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

func TestDefaultDAC(t *testing.T) {
	dac := DefaultDAC()

	if dac.Bits != 4 {
		t.Errorf("DAC Bits = %d, want 4 (literature optimal)", dac.Bits)
	}
	if dac.VrefHigh != 1.5 {
		t.Errorf("DAC VrefHigh = %v, want 1.5", dac.VrefHigh)
	}
	if dac.VrefLow != -1.5 {
		t.Errorf("DAC VrefLow = %v, want -1.5", dac.VrefLow)
	}
	if dac.INL != 0.5 {
		t.Errorf("DAC INL = %v, want 0.5", dac.INL)
	}
	if dac.DNL != 0.25 {
		t.Errorf("DAC DNL = %v, want 0.25", dac.DNL)
	}
	if dac.SettleTime != 10.0 {
		t.Errorf("DAC SettleTime = %v, want 10.0", dac.SettleTime)
	}
}

func TestDefaultADC(t *testing.T) {
	adc := DefaultADC()

	if adc.Bits != 6 {
		t.Errorf("ADC Bits = %d, want 6 (64 codes for 30 conductance levels)", adc.Bits)
	}
	if adc.VrefHigh != 1.0 {
		t.Errorf("ADC VrefHigh = %v, want 1.0", adc.VrefHigh)
	}
	if adc.VrefLow != 0.0 {
		t.Errorf("ADC VrefLow = %v, want 0.0", adc.VrefLow)
	}
	if adc.INL != 0.5 {
		t.Errorf("ADC INL = %v, want 0.5", adc.INL)
	}
	if adc.DNL != 0.25 {
		t.Errorf("ADC DNL = %v, want 0.25", adc.DNL)
	}
	if adc.ConversionTime != 50.0 {
		t.Errorf("ADC ConversionTime = %v, want 50.0", adc.ConversionTime)
	}
	if adc.Type != ADCTypeSAR {
		t.Errorf("ADC Type = %v, want SAR", adc.Type)
	}
}

func TestDACResolution(t *testing.T) {
	dac := DefaultDAC()
	// 4-bit: 16 levels, range = 3V, LSB = 3/15 = 200 mV
	expected := 3.0 / 15.0
	if math.Abs(dac.Resolution()-expected) > 1e-10 {
		t.Errorf("DAC Resolution = %v, want %v", dac.Resolution(), expected)
	}
}

func TestADCResolution(t *testing.T) {
	adc := DefaultADC()
	// 6-bit: 64 levels, range = 1V, LSB = 1/63 ≈ 15.87 mV
	expected := 1.0 / 63.0
	if math.Abs(adc.Resolution()-expected) > 1e-10 {
		t.Errorf("ADC Resolution = %v, want %v", adc.Resolution(), expected)
	}
}

func TestDefaultDACLevels(t *testing.T) {
	dac := DefaultDAC()
	if dac.Levels() != 16 {
		t.Errorf("DAC Levels = %d, want 16 (4-bit)", dac.Levels())
	}
}

func TestADCLevels(t *testing.T) {
	adc := DefaultADC()
	if adc.Levels() != 64 {
		t.Errorf("ADC Levels = %d, want 64 (6-bit)", adc.Levels())
	}
}

func TestDefaultTIA(t *testing.T) {
	tia := DefaultTIA()

	if tia.Gain != 10e3 {
		t.Errorf("TIA Gain = %v, want 10e3", tia.Gain)
	}
	if tia.Bandwidth != 100e6 {
		t.Errorf("TIA Bandwidth = %v, want 100e6", tia.Bandwidth)
	}
	// TIA uses InputNoiseRMS (1e-12) not Noise field
	if tia.InputNoiseRMS != 1e-12 {
		t.Errorf("TIA InputNoiseRMS = %v, want 1e-12", tia.InputNoiseRMS)
	}
	// TIA uses OutputOffset (5mV) not Offset field
	if tia.OutputOffset != 5e-3 {
		t.Errorf("TIA OutputOffset = %v, want 0.005", tia.OutputOffset)
	}
}

func TestTIASettlingTime(t *testing.T) {
	tia := DefaultTIA()
	// SettlingTime uses ln(1/0.001) / (2*pi*BW)
	expected := tia.SettlingTime()
	// Just verify it returns a positive value
	if expected <= 0 {
		t.Errorf("TIA SettlingTime = %v, want positive value", expected)
	}
}

func TestFeCIMLevelsConstant(t *testing.T) {
	if FeCIMLevels != 30 {
		t.Errorf("FeCIMLevels = %d, want 30", FeCIMLevels)
	}
}
