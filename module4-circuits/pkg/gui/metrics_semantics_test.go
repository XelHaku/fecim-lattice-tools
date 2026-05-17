//go:build legacy_fyne

package gui

import (
	"math"
	"strings"
	"testing"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
)

func TestReadModeMetricLabelsContainParentheses(t *testing.T) {
	labels := readModeMetricLabels()
	if len(labels) == 0 {
		t.Fatal("expected read-mode metric labels")
	}
	for _, label := range labels {
		if !strings.Contains(label, "(") || !strings.Contains(label, ")") {
			t.Fatalf("label %q must include parentheses with unit/range", label)
		}
	}
}

func TestMetricsSemanticsTIAOutputEqualsCurrentTimesGain(t *testing.T) {
	sense := arraysim.SenseChain{
		TIA: arraysim.TIAConfig{Rf: 10e3, Vref: 0, Vmin: 0, Vmax: 1.0},
		ADC: arraysim.ADCConfig{Bits: 5, Vmin: 0, Vmax: 1.0},
	}
	currentA := 20e-6
	got := sense.ConvertCurrent(currentA).Vout
	want := currentA * sense.TIA.Rf
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("V_TIA mismatch: got %.12f V, want %.12f V", got, want)
	}
}

func TestMetricsSemanticsADCCodeMatchesFloorOverLSBVoltage(t *testing.T) {
	sense := arraysim.SenseChain{
		TIA: arraysim.TIAConfig{Rf: 10e3, Vref: 0, Vmin: 0, Vmax: 1.0},
		ADC: arraysim.ADCConfig{Bits: 5, Vmin: 0, Vmax: 1.0},
	}
	levels := 1 << sense.ADC.Bits
	lsbVoltage := (sense.ADC.Vmax - sense.ADC.Vmin) / float64(levels-1)
	vTIA := 0.300
	currentA := vTIA / sense.TIA.Rf
	got := sense.ConvertCurrent(currentA).Code
	want := int(math.Floor(vTIA / lsbVoltage))
	if got != want {
		t.Fatalf("ADC code mismatch: got %d, want floor(V_TIA/LSB_voltage)=%d (V_TIA=%.3fV, LSB=%.6fV)", got, want, vTIA, lsbVoltage)
	}
}
