package compact

import "testing"

func TestFeFET_ThresholdV(t *testing.T) {
	cap := NewFeCap(DefaultFeCapHZO())
	fet := NewFeFET(cap, 1e-2, 0.4, 50e-6)

	// At zero polarisation, Vt should equal Vt0
	vt := fet.ThresholdV(0)
	if vt != 0.4 {
		t.Errorf("ThresholdV(0) = %g, want 0.4", vt)
	}

	// Positive polarisation lowers Vt
	vtPos := fet.ThresholdV(0.1)
	if vtPos >= vt {
		t.Errorf("ThresholdV(+P) = %g should be < ThresholdV(0) = %g", vtPos, vt)
	}

	// Negative polarisation raises Vt
	vtNeg := fet.ThresholdV(-0.1)
	if vtNeg <= vt {
		t.Errorf("ThresholdV(-P) = %g should be > ThresholdV(0) = %g", vtNeg, vt)
	}
}

func TestFeFET_DrainCurrentA_Off(t *testing.T) {
	cap := NewFeCap(DefaultFeCapHZO())
	fet := NewFeFET(cap, 1e-2, 0.4, 50e-6)

	// vgs well below Vt → off
	id := fet.DrainCurrentA(0.0, 0.1)
	if id != 0 {
		t.Errorf("DrainCurrentA(vgs=0, vds=0.1) = %g, want 0 (off region)", id)
	}
}

func TestFeFET_DrainCurrentA_Saturation(t *testing.T) {
	cap := NewFeCap(DefaultFeCapHZO())
	fet := NewFeFET(cap, 1e-2, 0.4, 50e-6)

	// vgs=1.0, vds=2.0 → saturation
	id := fet.DrainCurrentA(1.0, 2.0)
	if id <= 0 {
		t.Errorf("DrainCurrentA(vgs=1.0, vds=2.0) = %g, want > 0 (saturation)", id)
	}
}

func TestFeFET_DrainCurrentA_Linear(t *testing.T) {
	cap := NewFeCap(DefaultFeCapHZO())
	fet := NewFeFET(cap, 1e-2, 0.4, 50e-6)

	// vgs=1.0, small vds=0.05 → linear region
	id := fet.DrainCurrentA(1.0, 0.05)
	if id <= 0 {
		t.Errorf("DrainCurrentA(vgs=1.0, vds=0.05) = %g, want > 0 (linear)", id)
	}
}

func TestFeFET_DrainCurrent_IncreasesWithVgs(t *testing.T) {
	cap := NewFeCap(DefaultFeCapHZO())
	fet := NewFeFET(cap, 1e-2, 0.4, 50e-6)

	id1 := fet.DrainCurrentA(0.8, 1.0)
	id2 := fet.DrainCurrentA(1.2, 1.0)
	if id2 <= id1 {
		t.Errorf("higher vgs should give more current: id(vgs=1.2)=%g <= id(vgs=0.8)=%g", id2, id1)
	}
}
