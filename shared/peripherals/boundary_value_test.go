package peripherals

import (
	"math"
	"testing"
)

func TestBoundaryValue_ADC_InputVoltagePoints(t *testing.T) {
	adc := DefaultADC()
	vmin := adc.VrefLow
	vmax := adc.VrefHigh
	vmid := (vmin + vmax) / 2
	epsilon := adc.Resolution() / 10

	tests := []struct {
		name string
		v    float64
		want int
	}{
		{name: "Vmin", v: vmin, want: 0},
		{name: "Vmax", v: vmax, want: adc.Levels() - 1},
		{name: "midpoint", v: vmid, want: adc.Levels() / 2},
		{name: "Vmin-epsilon", v: vmin - epsilon, want: 0},
		{name: "Vmax+epsilon", v: vmax + epsilon, want: adc.Levels() - 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := adc.Convert(tc.v)
			if got != tc.want {
				t.Fatalf("ADC.Convert(%g) = %d, want %d", tc.v, got, tc.want)
			}
		})
	}
}

func TestBoundaryValue_DAC_CodePoints(t *testing.T) {
	dac := DefaultDAC()
	maxCode := dac.Levels() - 1
	midCode := maxCode / 2

	tests := []struct {
		name string
		code int
		want float64
	}{
		{name: "code-0", code: 0, want: dac.VrefLow},
		{name: "code-max", code: maxCode, want: dac.VrefHigh},
		{name: "code-midpoint", code: midCode, want: dac.VrefLow + float64(midCode)/float64(maxCode)*(dac.VrefHigh-dac.VrefLow)},
	}

	const tol = 1e-15
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := dac.Convert(tc.code)
			if math.Abs(got-tc.want) > tol {
				t.Fatalf("DAC.Convert(%d) = %.15f, want %.15f", tc.code, got, tc.want)
			}
		})
	}
}
