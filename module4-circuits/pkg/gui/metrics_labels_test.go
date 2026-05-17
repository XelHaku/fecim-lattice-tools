//go:build legacy_fyne

package gui

import (
	"strings"
	"testing"
)

func TestReadModeMetricLabelsContainUnits(t *testing.T) {
	labels := readModeMetricLabels()
	joined := strings.Join(labels, " | ")

	for _, want := range []string{"I_cell (µA)", "V_TIA (V)", "ADC Code", "I_LSB (µA/code)"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing label %q in %q", want, joined)
		}
	}
}

func TestMetricFormattingPrecision(t *testing.T) {
	if got := formatMetricVTIAMV(0.01234); got != "+0.01 V" {
		t.Fatalf("V_TIA format: got %q, want %q", got, "+0.01 V")
	}
	if got := formatMetricICellUA(-1.236); got != "-1.24 µA" {
		t.Fatalf("I_cell format: got %q, want %q", got, "-1.24 µA")
	}
	if got := formatMetricADCCode(127); got != "127" {
		t.Fatalf("ADC code format: got %q, want %q", got, "127")
	}
	if got := formatMetricConductanceUS(14.44); got != "14.4 µS" {
		t.Fatalf("conductance format: got %q, want %q", got, "14.4 µS")
	}
	if got := formatOverlayBottomValue("Vcell", -0.1234); got != "V: -0.12 V" {
		t.Fatalf("overlay V format: got %q, want %q", got, "V: -0.12 V")
	}
	if got := formatOverlayBottomValue("Icell", -1.234e-6); got != "I: -1.23 µA" {
		t.Fatalf("overlay I format: got %q, want %q", got, "I: -1.23 µA")
	}
}
