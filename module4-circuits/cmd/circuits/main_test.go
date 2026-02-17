package circuitscli

import "testing"

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
