package arraysim

import (
	"math"
	"testing"
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
