package system

import "testing"

func TestLeakagePowerUW(t *testing.T) {
	p := NewPowerModel(64, 64, 1.0, 1e-12)
	v := p.LeakagePowerUW()
	if v <= 0 {
		t.Errorf("LeakagePowerUW() = %g, want > 0", v)
	}
}

func TestSwitchingPowerUW(t *testing.T) {
	p := NewPowerModel(64, 64, 1.0, 1e-12)
	v := p.SwitchingPowerUW(100)
	if v <= 0 {
		t.Errorf("SwitchingPowerUW(100) = %g, want > 0", v)
	}
	// Higher frequency → more switching power
	v200 := p.SwitchingPowerUW(200)
	if v200 <= v {
		t.Errorf("SwitchingPowerUW(200)=%g should be > SwitchingPowerUW(100)=%g", v200, v)
	}
}

func TestTotalPowerUW(t *testing.T) {
	p := NewPowerModel(64, 64, 1.0, 1e-12)
	total := p.TotalPowerUW(100)
	expected := p.LeakagePowerUW() + p.SwitchingPowerUW(100)
	if total != expected {
		t.Errorf("TotalPowerUW(100) = %g, want %g", total, expected)
	}
	if total <= 0 {
		t.Errorf("TotalPowerUW(100) = %g, want > 0", total)
	}
}
