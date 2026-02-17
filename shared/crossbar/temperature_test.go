package crossbar

import (
	"math"
	"testing"
)

func TestNewTemperatureEffects(t *testing.T) {
	te := NewTemperatureEffects(300.0)
	if te.AmbientK != 300.0 {
		t.Errorf("Expected 300.0, got %f", te.AmbientK)
	}

	teNeg := NewTemperatureEffects(-10.0)
	if teNeg.AmbientK != 300.0 {
		t.Errorf("Expected 300.0 for negative input, got %f", teNeg.AmbientK)
	}
}

func TestAdjustedWireResistance(t *testing.T) {
	const R0 = 100.0
	const tolerance = 1e-9

	tests := []struct {
		tempK float64
		want  float64
	}{
		{300.0, R0},
		{301.0, R0 * (1.0 + 0.00393*1.0)},
		{TempAutomotive, R0 * (1.0 + 0.00393*100.0)},
		{77.0, R0 * (1.0 + 0.00393*(77.0-300.0))},
	}

	for _, tt := range tests {
		te := NewTemperatureEffects(tt.tempK)
		got := te.AdjustedWireResistance(R0)
		if math.Abs(got-tt.want) > tolerance {
			t.Errorf("At %fK: AdjustedWireResistance(%f) = %f, want %f", tt.tempK, R0, got, tt.want)
		}
	}
}

func TestAdjustedConductanceRange(t *testing.T) {
	const gMin, gMax = 10e-6, 100e-6
	const tolerance = 1e-12

	tests := []struct {
		name      string
		tempK     float64
		checkFunc func(gotMin, gotMax float64) bool
	}{
		{
			"Room Temp",
			300.0,
			func(gotMin, gotMax float64) bool {
				return math.Abs(gotMin-gMin) < tolerance && math.Abs(gotMax-gMax) < tolerance
			},
		},
		{
			"Cryogenic 77K",
			77.0,
			func(gotMin, gotMax float64) bool {
				return gotMin < gMin && gotMax > gMax
			},
		},
		{
			"Extreme Cryo 4K",
			4.0,
			func(gotMin, gotMax float64) bool {
				te77 := NewTemperatureEffects(77.0)
				gMin77, gMax77 := te77.AdjustedConductanceRange(gMin, gMax)
				return gotMin < gMin77 && gotMax > gMax77
			},
		},
		{
			"High Temp 400K",
			400.0,
			func(gotMin, gotMax float64) bool {
				return gotMin > gMin && gotMax < gMax
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			te := NewTemperatureEffects(tt.tempK)
			gotMin, gotMax := te.AdjustedConductanceRange(gMin, gMax)
			if !tt.checkFunc(gotMin, gotMax) {
				t.Errorf("At %fK: AdjustedConductanceRange failed: got (%e, %e)", tt.tempK, gotMin, gotMax)
			}
		})
	}
}

func TestAdjustedDriftRate(t *testing.T) {
	const drift0 = 0.001
	const tolerance = 1e-12

	te300 := NewTemperatureEffects(300.0)
	if math.Abs(te300.AdjustedDriftRate(drift0)-drift0) > tolerance {
		t.Errorf("At 300K: expected drift %f, got %f", drift0, te300.AdjustedDriftRate(drift0))
	}

	te400 := NewTemperatureEffects(400.0)
	drift400 := te400.AdjustedDriftRate(drift0)
	if drift400 <= drift0 {
		t.Errorf("At 400K: expected drift > %f, got %f", drift0, drift400)
	}

	te77 := NewTemperatureEffects(77.0)
	drift77 := te77.AdjustedDriftRate(drift0)
	if drift77 >= drift0 {
		t.Errorf("At 77K: expected drift < %f, got %f", drift0, drift77)
	}
}

func TestAdjustedRetention(t *testing.T) {
	te300 := NewTemperatureEffects(300.0)
	if math.Abs(te300.AdjustedRetention()-1.0) > 1e-9 {
		t.Errorf("At 300K: expected retention factor 1.0, got %f", te300.AdjustedRetention())
	}

	te400 := NewTemperatureEffects(400.0)
	if te400.AdjustedRetention() >= 1.0 {
		t.Errorf("At 400K: expected retention factor < 1.0, got %f", te400.AdjustedRetention())
	}

	te77 := NewTemperatureEffects(77.0)
	if te77.AdjustedRetention() <= 1.0 {
		t.Errorf("At 77K: expected retention factor > 1.0, got %f", te77.AdjustedRetention())
	}
}

func TestAdjustedNoise(t *testing.T) {
	te300 := NewTemperatureEffects(300.0)
	if math.Abs(te300.AdjustedNoise()-1.0) > 1e-9 {
		t.Errorf("At 300K: expected noise factor 1.0, got %f", te300.AdjustedNoise())
	}

	te600 := NewTemperatureEffects(600.0)
	if math.Abs(te600.AdjustedNoise()-math.Sqrt(2)) > 1e-9 {
		t.Errorf("At 600K: expected noise factor sqrt(2), got %f", te600.AdjustedNoise())
	}
}

func TestAdjustedSwitchingEnergy(t *testing.T) {
	const energy0 = 1.0
	te300 := NewTemperatureEffects(300.0)
	if math.Abs(te300.AdjustedSwitchingEnergy(energy0)-energy0) > 1e-9 {
		t.Errorf("At 300K: expected energy %f, got %f", energy0, te300.AdjustedSwitchingEnergy(energy0))
	}

	te77 := NewTemperatureEffects(77.0)
	if te77.AdjustedSwitchingEnergy(energy0) >= energy0 {
		t.Errorf("At 77K: expected energy < %f, got %f", energy0, te77.AdjustedSwitchingEnergy(energy0))
	}

	te400 := NewTemperatureEffects(400.0)
	if te400.AdjustedSwitchingEnergy(energy0) >= energy0 {
		t.Errorf("At 400K: expected energy < %f, got %f", energy0, te400.AdjustedSwitchingEnergy(energy0))
	}
}

func TestRetentionCurve(t *testing.T) {
	rc := DefaultRetentionCurve()
	if len(rc.TemperaturesK) == 0 {
		t.Fatal("Expected non-empty retention curve")
	}

	for i := 1; i < len(rc.TemperaturesK); i++ {
		if rc.RetentionS[i] >= rc.RetentionS[i-1] {
			t.Errorf("Expected decreasing retention, but at index %d: %e >= %e", i, rc.RetentionS[i], rc.RetentionS[i-1])
		}
	}

	r300 := rc.RetentionAt(300.0)
	r400 := rc.RetentionAt(400.0)
	if r400 >= r300 {
		t.Errorf("Expected lower retention at higher temp, got %e >= %e", r400, r300)
	}

	yrs300 := rc.RetentionYearsAt(300.0)
	if yrs300 <= 0 {
		t.Errorf("Expected positive years, got %f", yrs300)
	}
}

func TestMeetsAutomotiveGrade(t *testing.T) {
	rc := DefaultRetentionCurve()

	if rc.RetentionYearsAt(298.0) < 10.0 {
		t.Errorf("Expected FeCIM to meet 10-year retention at 25C, got %f", rc.RetentionYearsAt(298.0))
	}

	if !rc.MeetsAutomotiveGrade(3, 1.0) {
		t.Error("Expected FeCIM to meet Grade 3 (70C) for 1 year")
	}

	if rc.MeetsAutomotiveGrade(0, 1e9) {
		t.Error("Did not expect to meet Grade 0 for 1 billion years")
	}
}

func TestThermalPhysicsModel(t *testing.T) {
	m := DefaultHZOThermalModel()
	const tolerance = 1e-6

	pr300 := m.PrAtTemperature(300.0)
	if math.Abs(pr300-m.RefPr) > tolerance {
		t.Errorf("At 300K: expected Pr %f, got %f", m.RefPr, pr300)
	}

	pr400 := m.PrAtTemperature(400.0)
	if pr400 >= pr300 {
		t.Errorf("Expected lower Pr at higher temp, got %f >= %f", pr400, pr300)
	}

	prCurie := m.PrAtTemperature(m.CurieTempK)
	if prCurie != 0 {
		t.Errorf("Expected 0 Pr at Curie temp, got %f", prCurie)
	}

	ec300 := m.EcAtTemperature(300.0)
	if math.Abs(ec300-m.RefEc) > tolerance {
		t.Errorf("At 300K: expected Ec %f, got %f", m.RefEc, ec300)
	}

	ecCurie := m.EcAtTemperature(m.CurieTempK)
	if ecCurie != 0 {
		t.Errorf("Expected 0 Ec at Curie temp, got %f", ecCurie)
	}

	levels300 := m.EffectiveLevelsAtTemperature(300.0, 30)
	if levels300 != 30 {
		t.Errorf("At 300K: expected 30 levels, got %d", levels300)
	}

	levels500 := m.EffectiveLevelsAtTemperature(500.0, 30)
	if levels500 >= 30 {
		t.Errorf("At 500K: expected fewer than 30 levels, got %d", levels500)
	}
}

func TestGetAutomotiveReport(t *testing.T) {
	m := DefaultHZOThermalModel()
	report := m.GetAutomotiveReport(30)

	if report.Material != "HZO" {
		t.Errorf("Expected material HZO, got %s", report.Material)
	}

	if report.IndustrialLevels > 30 || report.IndustrialLevels < 2 {
		t.Errorf("Unexpected IndustrialLevels: %d", report.IndustrialLevels)
	}

	if report.Grade0LevelsAt150C > report.Grade1LevelsAt125C {
		t.Errorf("Grade 0 levels should be <= Grade 1 levels, got %d > %d", report.Grade0LevelsAt150C, report.Grade1LevelsAt125C)
	}
}
