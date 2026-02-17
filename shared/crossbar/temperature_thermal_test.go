package crossbar

import (
	"testing"
)

func TestGenerateRetentionCurve(t *testing.T) {
	// FeCIM: 10^7 seconds at 85°C, Ea = 1.1 eV
	rc := GenerateRetentionCurve(1e7, 1.1)

	if len(rc.TemperaturesK) != 8 {
		t.Errorf("Expected 8 temperature points, got %d", len(rc.TemperaturesK))
	}

	// At 85°C (358K), should be exactly reference
	ret85C := rc.RetentionS[4] // Index 4 = 85°C
	if ret85C != 1e7 {
		t.Errorf("Expected 1e7s at 85°C, got %e", ret85C)
	}

	// At 25°C, should be LONGER (Arrhenius: lower temp = longer retention)
	ret25C := rc.RetentionS[0]
	if ret25C <= ret85C {
		t.Error("Retention at 25°C should be longer than at 85°C")
	}

	// At 150°C, should be SHORTER
	ret150C := rc.RetentionS[7]
	if ret150C >= ret85C {
		t.Error("Retention at 150°C should be shorter than at 85°C")
	}
}

func TestRetentionCurve_RetentionAt(t *testing.T) {
	rc := DefaultRetentionCurve()

	// Room temperature (25°C = 298K) should have very long retention
	ret25C := rc.RetentionYearsAt(298)
	if ret25C < 100 {
		t.Errorf("Expected >100 years retention at 25°C, got %.1f years", ret25C)
	}

	// Industrial (85°C = 358K) - reference is 10^7 s ≈ 0.32 years
	ret85C := rc.RetentionYearsAt(358)
	expectedYears := 1e7 / (365.25 * 24 * 3600) // ~0.32 years
	if ret85C < expectedYears*0.9 || ret85C > expectedYears*1.1 {
		t.Errorf("Expected ~%.2f years at 85°C, got %.2f years", expectedYears, ret85C)
	}
}

func TestRetentionCurve_AutomotiveGrades(t *testing.T) {
	// Default FeCIM curve (10^7 s at 85°C, Ea=1.1)
	rc := DefaultRetentionCurve()

	// With 10^7 s (~0.32 years) at 85°C reference, retention is limited
	// But at lower temperatures, it should be LONGER (Arrhenius)

	ret85C := rc.RetentionYearsAt(358) // 85°C reference
	ret70C := rc.RetentionYearsAt(343) // 70°C
	ret25C := rc.RetentionYearsAt(298) // 25°C

	// Lower temperature should have longer retention
	if ret70C <= ret85C {
		t.Errorf("70°C retention (%.2f yr) should be > 85°C (%.2f yr)", ret70C, ret85C)
	}
	if ret25C <= ret70C {
		t.Errorf("25°C retention (%.2f yr) should be > 70°C (%.2f yr)", ret25C, ret70C)
	}

	// At 25°C, retention should be substantial (>10 years with good Ea)
	t.Logf("Retention: 25°C=%.1f yr, 70°C=%.1f yr, 85°C=%.2f yr", ret25C, ret70C, ret85C)
}

func TestThermalPhysicsModel_PrAtTemperature(t *testing.T) {
	model := DefaultHZOThermalModel()

	// Reference Pr at 300K
	prRef := model.RefPr
	pr300 := model.PrAtTemperature(300)

	// Should be close to reference
	if pr300 < prRef*0.95 || pr300 > prRef*1.05 {
		t.Errorf("Pr at 300K should be ~%.3f, got %.3f", prRef, pr300)
	}

	// At Curie temperature, Pr should be 0
	prCurie := model.PrAtTemperature(model.CurieTempK)
	if prCurie != 0 {
		t.Errorf("Pr at Curie temp should be 0, got %f", prCurie)
	}

	// Above Curie temperature, Pr should be 0
	prAboveCurie := model.PrAtTemperature(model.CurieTempK + 10)
	if prAboveCurie != 0 {
		t.Errorf("Pr above Curie temp should be 0, got %f", prAboveCurie)
	}

	// Pr should decrease with temperature
	pr400 := model.PrAtTemperature(400)
	if pr400 >= pr300 {
		t.Error("Pr should decrease with increasing temperature")
	}
}

func TestThermalPhysicsModel_EcAtTemperature(t *testing.T) {
	model := DefaultHZOThermalModel()

	ec300 := model.EcAtTemperature(300)
	ec400 := model.EcAtTemperature(400)

	// Ec should decrease with temperature
	if ec400 >= ec300 {
		t.Error("Ec should decrease with increasing temperature")
	}

	// At Curie temperature, Ec should be 0
	ecCurie := model.EcAtTemperature(model.CurieTempK)
	if ecCurie != 0 {
		t.Errorf("Ec at Curie temp should be 0, got %f", ecCurie)
	}
}

func TestThermalPhysicsModel_EffectiveLevels(t *testing.T) {
	model := DefaultHZOThermalModel()
	nominalLevels := 30

	// At room temperature, should be close to nominal
	levels300 := model.EffectiveLevelsAtTemperature(300, nominalLevels)
	if levels300 < nominalLevels-5 || levels300 > nominalLevels {
		t.Errorf("Expected ~%d levels at 300K, got %d", nominalLevels, levels300)
	}

	// At high temperature, should have fewer levels
	levels400 := model.EffectiveLevelsAtTemperature(400, nominalLevels)
	if levels400 >= levels300 {
		t.Error("Should have fewer effective levels at higher temperature")
	}

	// Should never go below 2 (binary)
	levelsHot := model.EffectiveLevelsAtTemperature(700, nominalLevels)
	if levelsHot < 2 {
		t.Errorf("Minimum levels should be 2, got %d", levelsHot)
	}
}

func TestThermalPhysicsModel_AutomotiveReport(t *testing.T) {
	model := DefaultHZOThermalModel()
	report := model.GetAutomotiveReport(30)

	if report.Material != "HZO" {
		t.Errorf("Expected material HZO, got %s", report.Material)
	}

	// Industrial should have more levels than Grade 0 (higher temp)
	if report.IndustrialLevels <= report.Grade0LevelsAt150C {
		t.Error("Industrial grade should have more levels than Grade 0")
	}

	// Industrial retention should be longer than Grade 0
	if report.IndustrialRetention <= report.Grade0RetentionYrs {
		t.Error("Industrial retention should be longer than Grade 0")
	}
}

func TestTemperaturePresets(t *testing.T) {
	// Verify preset values are reasonable
	if TempCryogenic != 77.0 {
		t.Errorf("Cryogenic should be 77K (LN2), got %f", TempCryogenic)
	}
	if TempRoom != 300.0 {
		t.Errorf("Room temp should be 300K, got %f", TempRoom)
	}
	if TempIndustrial != 358.0 {
		t.Errorf("Industrial should be 358K (85°C), got %f", TempIndustrial)
	}
	if TempAutomotive != 400.0 {
		t.Errorf("Automotive should be 400K (127°C), got %f", TempAutomotive)
	}
}

func TestTemperatureGrades(t *testing.T) {
	// Grade 0 should be widest range
	if TempGrade0Max-TempGrade0Min <= TempGrade1Max-TempGrade1Min {
		t.Error("Grade 0 should have widest temperature range")
	}

	// All grades should have min at -40°C (233K) except Grade 3
	if TempGrade0Min != 233 || TempGrade1Min != 233 || TempGrade2Min != 233 {
		t.Error("Grades 0-2 should have min temp at -40°C (233K)")
	}

	// Grade 3 min is 0°C (273K)
	if TempGrade3Min != 273 {
		t.Errorf("Grade 3 min should be 273K (0°C), got %f", TempGrade3Min)
	}
}
