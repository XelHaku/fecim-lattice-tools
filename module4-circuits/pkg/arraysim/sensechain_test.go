package arraysim

import (
	"math"
	"testing"

	"fecim-lattice-tools/shared/peripherals"
)

const senseEpsilon = 1e-12

func TestSenseChain_CurrentRangeAndLSB(t *testing.T) {
	sense := SenseChain{
		TIA: TIAConfig{
			Rf:   10e3,
			Vref: 0.1,
			Vmin: 0.0,
			Vmax: 1.0,
		},
		ADC: ADCConfig{
			Bits: 5,
			Vmin: 0.0,
			Vmax: 1.0,
		},
	}

	imin, imax := sense.CurrentRange()
	wantImin := -10e-6
	wantImax := 90e-6
	if math.Abs(imin-wantImin) > senseEpsilon {
		t.Fatalf("Imin: got %.12e, want %.12e", imin, wantImin)
	}
	if math.Abs(imax-wantImax) > senseEpsilon {
		t.Fatalf("Imax: got %.12e, want %.12e", imax, wantImax)
	}

	lsb := sense.CurrentLSB()
	wantLSB := (1.0 - 0.0) / 31.0 / 10e3
	if math.Abs(lsb-wantLSB) > senseEpsilon {
		t.Fatalf("LSB: got %.12e, want %.12e", lsb, wantLSB)
	}
}

func TestSenseChain_CurrentRangeClampsToADC(t *testing.T) {
	sense := SenseChain{
		TIA: TIAConfig{
			Rf:   20e3,
			Vref: 0.5,
			Vmin: -0.2,
			Vmax: 1.2,
		},
		ADC: ADCConfig{
			Bits: 6,
			Vmin: 0.2,
			Vmax: 0.8,
		},
	}

	imin, imax := sense.CurrentRange()
	wantImin := (0.2 - 0.5) / 20e3
	wantImax := (0.8 - 0.5) / 20e3
	if math.Abs(imin-wantImin) > senseEpsilon {
		t.Fatalf("Imin: got %.12e, want %.12e", imin, wantImin)
	}
	if math.Abs(imax-wantImax) > senseEpsilon {
		t.Fatalf("Imax: got %.12e, want %.12e", imax, wantImax)
	}

	lsb := sense.CurrentLSB()
	wantLSB := (0.8 - 0.2) / 63.0 / 20e3
	if math.Abs(lsb-wantLSB) > senseEpsilon {
		t.Fatalf("LSB: got %.12e, want %.12e", lsb, wantLSB)
	}
}

func TestSenseChain_CurrentRangeSaturation(t *testing.T) {
	sense := SenseChain{
		TIA: TIAConfig{
			Rf:   10e3,
			Vref: 0.1,
			Vmin: 0.0,
			Vmax: 1.0,
		},
		ADC: ADCConfig{
			Bits: 5,
			Vmin: 0.0,
			Vmax: 1.0,
		},
	}

	imin, imax := sense.CurrentRange()
	lsb := sense.CurrentLSB()

	low := sense.ConvertCurrent(imin - lsb)
	if !low.TIASaturated || !low.ADCSaturated {
		t.Fatalf("expected low saturation, got TIA=%v ADC=%v", low.TIASaturated, low.ADCSaturated)
	}

	high := sense.ConvertCurrent(imax + lsb)
	if !high.TIASaturated || !high.ADCSaturated {
		t.Fatalf("expected high saturation, got TIA=%v ADC=%v", high.TIASaturated, high.ADCSaturated)
	}
}

func TestSenseChain_ADCQuantization_RoundToNearestHalfUp(t *testing.T) {
	sense := SenseChain{
		TIA: TIAConfig{
			Rf:   1.0,
			Vref: 0.0,
			Vmin: -10.0,
			Vmax: 10.0,
		},
		ADC: ADCConfig{
			Bits: 3, // 8 levels, codes 0..7
			Vmin: 0.0,
			Vmax: 1.0,
		},
	}

	// For an N-bit ADC with codes 0..maxCode, the ideal LSB is span/maxCode.
	levels := 1 << uint(sense.ADC.Bits)
	maxCode := float64(levels - 1)
	lsbV := (sense.ADC.Vmax - sense.ADC.Vmin) / maxCode

	// Exactly half an LSB above code 0 should round up to code 1 (ties half-up).
	res := sense.ConvertCurrent(0.5 * lsbV)
	if res.Code != 1 {
		t.Fatalf("half-LSB (%.9f V): got code=%d, want code=1", 0.5*lsbV, res.Code)
	}

	// Slightly below the half-LSB threshold should stay at code 0.
	res = sense.ConvertCurrent(0.5*lsbV - 1e-12)
	if res.Code != 0 {
		t.Fatalf("just-below-half-LSB (%.12f V): got code=%d, want code=0", 0.5*lsbV-1e-12, res.Code)
	}

	// Midpoint between codes 3 and 4 should round up to 4.
	res = sense.ConvertCurrent((3.5) * lsbV)
	if res.Code != 4 {
		t.Fatalf("midpoint 3.5 LSB (%.9f V): got code=%d, want code=4", 3.5*lsbV, res.Code)
	}
}

func TestSenseChain_MatchesPeripheralReadPath_Default5Bit(t *testing.T) {
	tia := peripherals.DefaultTIA()
	adc := peripherals.DefaultADC()

	sense := SenseChain{
		TIA: TIAConfig{
			Rf:   tia.Gain,
			Vref: tia.OutputOffset,
			Vmin: 0,
			Vmax: tia.MaxOutputVoltage,
		},
		ADC: ADCConfig{
			Bits: adc.Bits,
			Vmin: adc.VrefLow,
			Vmax: adc.VrefHigh,
		},
	}

	currentsA := []float64{-20e-6, 0, 10e-6, 50e-6, 100e-6, 150e-6}
	const vTolV = 1e-12
	for _, currentA := range currentsA {
		got := sense.ConvertCurrent(currentA)
		wantV := tia.Convert(currentA)
		wantCode := adc.Convert(wantV)

		if math.Abs(got.Vout-wantV) > vTolV {
			t.Fatalf("current %.3f µA: Vout got %.9f V, want %.9f V (|Δ|=%.3e V, tol=%.1e V)",
				currentA*1e6, got.Vout, wantV, math.Abs(got.Vout-wantV), vTolV)
		}
		if got.Code != wantCode {
			t.Fatalf("current %.3f µA: code got %d, want %d", currentA*1e6, got.Code, wantCode)
		}
	}
}

func TestSenseChain_ReadSignConvention_PositiveBLCurrentToHigherCode(t *testing.T) {
	sense := SenseChain{
		TIA: TIAConfig{Rf: 10e3, Vref: 5e-3, Vmin: 0, Vmax: 1.0},
		ADC: ADCConfig{Bits: 5, Vmin: 0, Vmax: 1.0},
	}

	low := sense.ConvertCurrent(-10e-6)
	high := sense.ConvertCurrent(+10e-6)

	if high.Vout <= low.Vout {
		t.Fatalf("sign convention error: +10 µA should increase Vout, got V(+10µA)=%.6f V <= V(-10µA)=%.6f V", high.Vout, low.Vout)
	}
	if high.Code <= low.Code {
		t.Fatalf("sign convention error: +10 µA should increase ADC code, got code(+10µA)=%d <= code(-10µA)=%d", high.Code, low.Code)
	}
}
