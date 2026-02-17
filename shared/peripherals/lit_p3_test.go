package peripherals

// Tests for LIT-P3-03 (charge pump staging), LIT-P3-04 (thermometer DAC),
// and LIT-P3-05 (DAC glitch energy model).

import (
	"math"
	"testing"
)

// --- LIT-P3-03: Charge pump staging ---

func TestStagesRequired(t *testing.T) {
	// 1 V supply, 0.3 V diode drop:
	// N = ceil((Vout - Vin) / (Vin - Vdrop)) = ceil((Vout-1)/0.7)
	cases := []struct {
		targetV, inputV, diodeDrop float64
		wantMin                    int // minimum stages
	}{
		{1.5, 1.0, 0.3, 1}, // ceil(0.5/0.7)=1
		{3.0, 1.0, 0.3, 3}, // ceil(2.0/0.7)=3
		{5.0, 1.0, 0.3, 6}, // ceil(4.0/0.7)=6 (conservative)
		{1.0, 1.0, 0.3, 1}, // at supply → still at least 1
	}
	for _, tc := range cases {
		n := StagesRequired(tc.targetV, tc.inputV, tc.diodeDrop)
		if n < tc.wantMin {
			t.Errorf("StagesRequired(%.1f,%.1f,%.1f)=%d want >=%d",
				tc.targetV, tc.inputV, tc.diodeDrop, n, tc.wantMin)
		}
		// Verify the stages actually achieve the target (ideal, no losses).
		idealOut := float64(n+1)*tc.inputV - float64(n)*tc.diodeDrop
		if idealOut < math.Abs(tc.targetV)-1e-9 {
			t.Errorf("StagesRequired: %d stages ideal=%.2f V < target=%.2f V",
				n, idealOut, tc.targetV)
		}
	}
}

func TestFeCAPChargePump(t *testing.T) {
	cp := FeCAPChargePump(1.0)
	if cp.Stages != 2 {
		t.Errorf("FeCAPChargePump stages=%d want 2", cp.Stages)
	}
	if cp.OutputVoltage != 1.5 {
		t.Errorf("FeCAPChargePump output=%.1f want 1.5", cp.OutputVoltage)
	}
	// Actual output must be ≥ target.
	actual := cp.ActualOutputVoltage()
	if actual < cp.OutputVoltage-1e-9 {
		t.Errorf("FeCAPChargePump actual=%.3f < target=%.3f", actual, cp.OutputVoltage)
	}
	// Energy per cycle: E = N × Cfb × Vin²
	e := cp.EnergyPerCycle()
	expected := float64(cp.Stages) * cp.FlyCapacitance * cp.InputVoltage * cp.InputVoltage
	if math.Abs(e-expected) > 1e-30 {
		t.Errorf("EnergyPerCycle=%e want %e", e, expected)
	}
}

func TestFeFETChargePump(t *testing.T) {
	targets := []float64{3.0, 5.0}
	for _, tV := range targets {
		cp := FeFETChargePump(1.0, tV)
		actual := cp.ActualOutputVoltage()
		if actual < tV-0.5 { // allow 0.5V margin (efficiency losses)
			t.Errorf("FeFETChargePump(%.0fV) actual=%.3f V < target %.1f V - 0.5",
				tV, actual, tV)
		}
		if cp.Stages < 1 {
			t.Errorf("FeFETChargePump(%.0fV) stages=%d < 1", tV, cp.Stages)
		}
	}
}

func TestChargePumpEnergyPerCycle(t *testing.T) {
	cp := DefaultChargePump()
	e := cp.EnergyPerCycle()
	// E = 2 × 100pF × 1V² = 200 fJ
	expected := 2 * 100e-12 * 1.0 * 1.0
	if math.Abs(e-expected)/expected > 1e-9 {
		t.Errorf("EnergyPerCycle=%e want %e", e, expected)
	}
}

// --- LIT-P3-04: Thermometer DAC ---

func TestThermometerDAC(t *testing.T) {
	d := ThermometerDAC(4)
	if d.Encoding != DACEncodingThermometer {
		t.Errorf("ThermometerDAC encoding=%v want DACEncodingThermometer", d.Encoding)
	}
	if d.Bits != 4 {
		t.Errorf("ThermometerDAC bits=%d want 4", d.Bits)
	}
	// Output voltages must match binary DAC (same output, different internal code).
	bin := DefaultDAC()
	for level := 0; level < d.Levels(); level++ {
		vTherm := d.Convert(level)
		vBin := bin.Convert(level)
		if math.Abs(vTherm-vBin) > 1e-12 {
			t.Errorf("ThermometerDAC.Convert(%d)=%e != BinaryDAC.Convert(%d)=%e",
				level, vTherm, level, vBin)
		}
	}
}

// --- LIT-P3-05: DAC glitch energy ---

func TestGlitchTransitions_Binary(t *testing.T) {
	d := DefaultDAC() // binary encoding

	// 7→8 in 4-bit: 0111→1000 — all 4 bits flip.
	trans := d.GlitchTransitions(7, 8)
	if trans != 4 {
		t.Errorf("binary GlitchTransitions(7,8)=%d want 4", trans)
	}

	// 0→1: only 1 bit flips.
	if d.GlitchTransitions(0, 1) != 1 {
		t.Errorf("binary GlitchTransitions(0,1)=%d want 1", d.GlitchTransitions(0, 1))
	}

	// Same code: no transitions.
	if d.GlitchTransitions(5, 5) != 0 {
		t.Errorf("binary GlitchTransitions(5,5)=%d want 0", d.GlitchTransitions(5, 5))
	}

	// Symmetric: from→to == to→from.
	if d.GlitchTransitions(3, 12) != d.GlitchTransitions(12, 3) {
		t.Error("GlitchTransitions not symmetric")
	}
}

func TestGlitchTransitions_Thermometer(t *testing.T) {
	d := ThermometerDAC(4)

	// 7→8: 1 step, 1 cell changes (no glitch in thermometer).
	trans := d.GlitchTransitions(7, 8)
	if trans != 1 {
		t.Errorf("thermometer GlitchTransitions(7,8)=%d want 1", trans)
	}

	// 0→15: 15 cells change (but one at a time — zero simultaneous glitch).
	if d.GlitchTransitions(0, 15) != 15 {
		t.Errorf("thermometer GlitchTransitions(0,15)=%d want 15", d.GlitchTransitions(0, 15))
	}
}

func TestGlitchEnergy(t *testing.T) {
	bin := DefaultDAC()
	therm := ThermometerDAC(4)

	// Binary worst-case: 7→8 (4 bits flip).
	eBin := bin.GlitchEnergy(7, 8)
	if eBin <= 0 {
		t.Errorf("binary GlitchEnergy(7,8)=%e want > 0", eBin)
	}

	// Thermometer: always 0 (no simultaneous multi-bit switching).
	eTherm := therm.GlitchEnergy(7, 8)
	if eTherm != 0 {
		t.Errorf("thermometer GlitchEnergy(7,8)=%e want 0", eTherm)
	}

	// Thermometer glitch = 0 even for large jumps.
	eThermBig := therm.GlitchEnergy(0, 15)
	if eThermBig != 0 {
		t.Errorf("thermometer GlitchEnergy(0,15)=%e want 0", eThermBig)
	}

	// Binary glitch scales with transitions (0→1 < 7→8).
	e01 := bin.GlitchEnergy(0, 1)
	if e01 >= eBin {
		t.Errorf("expected binary GlitchEnergy(0,1)=%e < GlitchEnergy(7,8)=%e", e01, eBin)
	}

	// Verify units: C_unit=0.2fF, V_lsb=(3V/15)=0.2V, 4 transitions
	// E = 4 × 0.2e-15 × 0.2² = 4 × 0.2e-15 × 0.04 = 3.2e-17 J
	cUnit := 0.2e-15
	vLSB := (bin.VrefHigh - bin.VrefLow) / float64(bin.Levels()-1)
	wantBin := 4.0 * cUnit * vLSB * vLSB
	if math.Abs(eBin-wantBin)/wantBin > 1e-9 {
		t.Errorf("binary GlitchEnergy(7,8)=%e want %e", eBin, wantBin)
	}
}
