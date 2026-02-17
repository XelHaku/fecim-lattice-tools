package peripherals

// Tests for the ChargeAmplifier (LIT-P2-03).

import (
	"math"
	"testing"
)

func TestChargeAmplifier_Defaults(t *testing.T) {
	ca := DefaultChargeAmplifier()
	if ca.Cfb <= 0 {
		t.Errorf("Cfb=%e <= 0", ca.Cfb)
	}
	if ca.Bandwidth <= 0 {
		t.Errorf("Bandwidth=%e <= 0", ca.Bandwidth)
	}
	if ca.MaxOutputVoltage <= 0 {
		t.Errorf("MaxOutputVoltage=%e <= 0", ca.MaxOutputVoltage)
	}
}

func TestChargeAmplifier_Sense(t *testing.T) {
	ca := DefaultChargeAmplifier()

	// V_out = Q / Cfb
	q := 64e-15 // 64 fC
	vOut := ca.Sense(q)
	want := q / ca.Cfb
	if math.Abs(vOut-want) > 1e-12 {
		t.Errorf("Sense(%e): got %e want %e", q, vOut, want)
	}

	// Clipping at MaxOutputVoltage.
	vClipped := ca.Sense(1000e-15) // huge charge
	if vClipped != ca.MaxOutputVoltage {
		t.Errorf("clipped output %e != MaxOutputVoltage %e", vClipped, ca.MaxOutputVoltage)
	}
	vNeg := ca.Sense(-1000e-15)
	if vNeg != -ca.MaxOutputVoltage {
		t.Errorf("negative clipped output %e != -%e", vNeg, ca.MaxOutputVoltage)
	}

	// Zero charge → zero output.
	if ca.Sense(0) != 0 {
		t.Errorf("Sense(0) = %e != 0", ca.Sense(0))
	}
}

func TestChargeAmplifier_SNR(t *testing.T) {
	ca := DefaultChargeAmplifier()

	// Larger charge → higher SNR.
	snr1 := ca.SNR(10e-15)
	snr2 := ca.SNR(100e-15)
	if snr2 <= snr1 {
		t.Errorf("SNR should increase with charge: SNR(10fC)=%e SNR(100fC)=%e", snr1, snr2)
	}
}

func TestChargeAmplifier_SettlingTime(t *testing.T) {
	ca := DefaultChargeAmplifier()
	ts := ca.SettlingTime()
	// For 500 MHz BW: ts = 7/(2π*500e6) ≈ 2.2 ns
	if ts <= 0 || ts > 100e-9 {
		t.Errorf("SettlingTime=%e out of range (0, 100 ns)", ts)
	}
}

func TestChargeAmplifier_MinDetectableCharge(t *testing.T) {
	ca := DefaultChargeAmplifier()
	qMin := ca.MinDetectableCharge()
	if qMin <= 0 {
		t.Errorf("MinDetectableCharge=%e <= 0", qMin)
	}
	// At SNR=1 the signal equals the noise floor.
	if math.Abs(ca.SNR(qMin)-1) > 0.01 {
		t.Errorf("SNR at MinDetectable=%e != 1.0", ca.SNR(qMin))
	}
}
