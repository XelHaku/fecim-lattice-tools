package crossbar

import (
	"math"
	"testing"
)

// Tests for LIT-P2-07: FeFET subthreshold non-linear I-V model.

func TestFeFETIVParams_Defaults(t *testing.T) {
	iv := DefaultFeFETIVParams()
	if iv.SubthSlope != 1.2 {
		t.Errorf("SubthSlope=%v want 1.2", iv.SubthSlope)
	}
	if iv.TempK != 300 {
		t.Errorf("TempK=%v want 300", iv.TempK)
	}
	// V_T = kT/q ≈ 25.85 mV at 300 K
	vT := iv.ThermalVoltage()
	if math.Abs(vT-0.02585) > 0.0001 {
		t.Errorf("ThermalVoltage=%e want ~25.85 mV", vT)
	}
	// V_sat = 1.2 × 25.85 mV ≈ 31.02 mV
	vSat := iv.VSat()
	if math.Abs(vSat-0.03102) > 0.0005 {
		t.Errorf("VSat=%e want ~31 mV", vSat)
	}
}

func TestFeFETIVParams_Current_ZeroVoltage(t *testing.T) {
	iv := DefaultFeFETIVParams()
	if iv.Current(1.0, 0) != 0 {
		t.Errorf("Current(G=1, V=0) should be 0")
	}
	if iv.Current(1.0, 1e-16) != 0 {
		t.Errorf("Current(G=1, V=1e-16) below threshold should be 0")
	}
}

func TestFeFETIVParams_Current_LinearRegime(t *testing.T) {
	iv := DefaultFeFETIVParams()
	vSat := iv.VSat()
	// V_DS << V_sat: I ≈ G × V_DS (Ohmic)
	// Use V = V_sat/100 for good linear approximation
	vSmall := vSat / 100.0
	iNonlinear := iv.Current(1.0, vSmall)
	iOhmic := vSmall // G=1
	// Taylor: (1-exp(-x)) ≈ x - x²/2 for small x
	// Relative error ≈ x/2 = (V/V_sat)/2 ≈ 0.005 (0.5%)
	relErr := math.Abs(iNonlinear-iOhmic) / iOhmic
	if relErr > 0.01 {
		t.Errorf("Linear regime relative error %.2f%% > 1%% at V=Vsat/100", relErr*100)
	}
}

func TestFeFETIVParams_Current_Saturation(t *testing.T) {
	iv := DefaultFeFETIVParams()
	vSat := iv.VSat()
	// V_DS >> V_sat: I → G × V_sat
	vLarge := vSat * 100.0
	iNonlinear := iv.Current(1.0, vLarge)
	iSat := vSat // G=1, saturated current limit
	relErr := math.Abs(iNonlinear-iSat) / iSat
	if relErr > 0.001 { // within 0.1%
		t.Errorf("Saturation relative error %.4f%% > 0.1%% at V=100×Vsat", relErr*100)
	}
}

func TestFeFETIVParams_Current_Antisymmetric(t *testing.T) {
	iv := DefaultFeFETIVParams()
	// Current must be antisymmetric: I(-V) = -I(V)
	for _, v := range []float64{0.01, 0.05, 0.5, 2.0} {
		iPos := iv.Current(1.0, v)
		iNeg := iv.Current(1.0, -v)
		if math.Abs(iPos+iNeg) > 1e-20 {
			t.Errorf("Antisymmetry violated at V=%v: I(V)=%e I(-V)=%e", v, iPos, iNeg)
		}
	}
}

func TestFeFETIVParams_LinearityError(t *testing.T) {
	iv := DefaultFeFETIVParams()
	// Error at V=0 should be 0.
	if iv.LinearityError(0) != 0 {
		t.Errorf("LinearityError(0) != 0")
	}
	// Error increases with V (Ohmic model overestimates more at higher V).
	e1 := iv.LinearityError(0.01)
	e2 := iv.LinearityError(0.1)
	e3 := iv.LinearityError(1.0)
	if !(e1 < e2 && e2 < e3) {
		t.Errorf("LinearityError should increase with |V|: e(0.01)=%e e(0.1)=%e e(1.0)=%e", e1, e2, e3)
	}
	// All errors should be positive (linear model overestimates).
	if e1 <= 0 || e2 <= 0 || e3 <= 0 {
		t.Errorf("LinearityError should be positive (Ohmic overestimates): %e %e %e", e1, e2, e3)
	}
}

func TestMVMNonLinear_FeCAP_Error(t *testing.T) {
	cfg := DefaultFeCAPConfig(4, 4)
	a, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}
	_, err = a.MVMNonLinear([]float64{0.5, 0.5, 0.5, 0.5}, nil)
	if err == nil {
		t.Error("MVMNonLinear should return error for FeCAP arrays")
	}
}

func TestMVMNonLinear_InputLengthMismatch(t *testing.T) {
	a, err := NewArray(&Config{Rows: 4, Cols: 4, ADCBits: 4, DACBits: 4})
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}
	_, err = a.MVMNonLinear([]float64{0.5, 0.5}, nil) // 2 != 4
	if err == nil {
		t.Error("MVMNonLinear should return error when input length != cols")
	}
}

func TestMVMNonLinear_OutputInRange(t *testing.T) {
	// Output must always be in [0, 1] regardless of input voltage.
	a, err := NewArray(&Config{Rows: 4, Cols: 4, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			if err := a.ProgramWeight(r, c, 1.0); err != nil {
				t.Fatalf("ProgramWeight: %v", err)
			}
		}
	}
	for _, v := range []float64{0.001, 0.01, 0.1, 0.5, 1.0} {
		input := []float64{v, v, v, v}
		out, err := a.MVMNonLinear(input, nil)
		if err != nil {
			t.Fatalf("MVMNonLinear at V=%v: %v", v, err)
		}
		for r, o := range out {
			if o < 0 || o > 1+1e-9 {
				t.Errorf("row %d at V=%v: output %e out of [0,1]", r, v, o)
			}
		}
	}
}

func TestMVMNonLinear_PhysicsSaturation(t *testing.T) {
	// In the saturated regime (V >> V_sat), doubling V should barely change output.
	// In the linear regime (V << V_sat), doubling V should approximately double output.
	// Test at physics level: iv.Current(G, 2V)/iv.Current(G, V).
	iv := DefaultFeFETIVParams()
	vSat := iv.VSat()

	// Linear regime: V = V_sat/20 → doubling gives ~2× (within 5%)
	vLin := vSat / 20.0
	ratioLin := iv.Current(1.0, 2*vLin) / iv.Current(1.0, vLin)
	if math.Abs(ratioLin-2.0) > 0.05*2.0 {
		t.Errorf("Linear regime: Current(2V)/Current(V)=%e want ~2.0 at V=Vsat/20", ratioLin)
	}

	// Saturated regime: V = 10×V_sat → doubling gives <1.01× (barely changes)
	vSat10 := vSat * 10.0
	ratioSat := iv.Current(1.0, 2*vSat10) / iv.Current(1.0, vSat10)
	if ratioSat > 1.01 {
		t.Errorf("Saturated regime: Current(2V)/Current(V)=%e want <1.01 at V=10×Vsat", ratioSat)
	}
}
