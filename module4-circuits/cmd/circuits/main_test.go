package circuitscli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunCircuitsReportsFlagErrorToStderr(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := runCircuits([]string{"-definitely-not-a-flag"}, &stdout, &stderr)

	if err == nil {
		t.Fatal("runCircuits error = nil, want invalid flag error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout length = %d, want 0; stdout=%q", stdout.Len(), stdout.String())
	}
	text := stderr.String()
	if !strings.Contains(text, "flag provided but not defined: -definitely-not-a-flag") {
		t.Fatalf("stderr = %q, want invalid flag context", text)
	}
	if !strings.Contains(text, "Error:") {
		t.Fatalf("stderr = %q, want error prefix", text)
	}
	if !strings.Contains(text, "FeCIM Peripheral Circuits CLI") {
		t.Fatalf("stderr = %q, want usage", text)
	}
}

func TestBuildCircuitsResultShowAll(t *testing.T) {
	r := buildCircuitsResult(false, false, false, false, true)
	if r.DAC == nil || r.ADC == nil || r.TIA == nil || r.Pump == nil {
		t.Fatal("showAll should populate all peripheral result sections")
	}
	if r.DAC.Levels <= 0 || r.ADC.Levels <= 0 {
		t.Fatal("invalid DAC/ADC levels in result")
	}
}

func TestCheckMonotonicity(t *testing.T) {
	if !checkMonotonicity([]float64{0.1, -0.9, 0.2}) {
		t.Fatal("expected monotonic pass when all DNL > -1")
	}
	if checkMonotonicity([]float64{0.1, -1.2, 0.2}) {
		t.Fatal("expected monotonic fail when a DNL < -1")
	}
}
