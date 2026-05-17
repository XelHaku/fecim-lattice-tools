//go:build legacy_fyne

package gui

import (
	"math"
	"testing"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

// TestModule1MaterialFeedsModule4VoltageRanges verifies Module 1 material outputs
// (via ferroelectric re-exports) directly drive Module 4 coercive-voltage based ranges.
func TestModule1MaterialFeedsModule4VoltageRanges(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(4, 4)

	module1Mat := sharedphysics.CryogenicHZO()
	ds.SetMaterial(module1Mat)

	writeRange := ds.GetWriteRange()
	readRange := ds.GetReadRange()

	wantVc := module1Mat.CoerciveVoltage()
	wantWriteMax := ds.calibParams.FieldMaxRatio * wantVc
	if wantWriteMax > MaxPracticalVoltage {
		wantWriteMax = MaxPracticalVoltage
	}
	if math.Abs(writeRange.Max-wantWriteMax) > wantVc*1e-12 {
		t.Fatalf("writeRange.Max mismatch from module1 material Vc: got=%g V want=%g V", writeRange.Max, wantWriteMax)
	}
	if math.Abs(writeRange.Min+wantWriteMax) > wantVc*1e-12 {
		t.Fatalf("writeRange.Min mismatch (expected bipolar symmetric): got=%g V want=%g V", writeRange.Min, -wantWriteMax)
	}

	if writeRange.NumLevels != module1Mat.GetNumLevels() {
		t.Fatalf("numLevels mismatch: module4=%d module1=%d", writeRange.NumLevels, module1Mat.GetNumLevels())
	}

	// Read max follows calibration ratio * Vc (with module4 practical clamping).
	wantReadMax := ds.calibParams.FieldMinRatio * wantVc
	if wantReadMax > 1.0 {
		wantReadMax = 1.0
	}
	if wantReadMax < 0.1 {
		wantReadMax = 0.1
	}
	if math.Abs(readRange.Max-wantReadMax) > 1e-12 {
		t.Fatalf("readRange.Max mismatch: got=%g V want=%g V", readRange.Max, wantReadMax)
	}
}

// TestModule1ToModule4ConductanceConsistency verifies that for reference geometry,
// Module 4 uses exactly the same conductance mapping as Module 1/shared physics.
func TestModule1ToModule4ConductanceConsistency(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(2, 2)

	module1Mat := sharedphysics.FeCIMMaterial()
	ds.SetMaterial(module1Mat)

	levels := module1Mat.GetNumLevels()
	for _, level := range []int{0, levels / 2, levels - 1} {
		want := module1Mat.DiscreteLevel(level, levels)
		got := ds.levelToConductance(level, levels)
		if math.Abs(got-want) > want*1e-12 {
			t.Fatalf("level %d conductance mismatch: got=%g S want=%g S", level, got, want)
		}
	}
}

// TestIdealComputeHonorsGeometryScaling ensures the ideal compute path uses the same
// geometry-scaled conductance dependency as the coupling path.
func TestIdealComputeHonorsGeometryScaling(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(1, 1)

	mat := sharedphysics.FeCIMMaterial()
	ds.SetMaterial(mat)
	ds.SetCouplingMode(arraysim.CouplingIdeal)
	ds.SetWLSingle(0)
	ds.SetDACVoltage(0, 1.0)

	refGeom := sharedphysics.GeometryFromMaterial(mat)
	geom := arraysim.DefaultCellGeometry()
	geom.Film = refGeom
	geom.Film.Area = refGeom.Area * 2
	geom.Film.Thickness = refGeom.Thickness / 2 // net scale = 4x
	ds.SetCellGeometry(geom)

	levels := mat.GetNumLevels()
	level := levels / 2
	weights := [][]int{{level}}
	ds.Compute(weights, levels)

	scale := geom.Film.ConductanceScale(refGeom)
	wantConductanceS := mat.DiscreteLevel(level, levels) * scale
	wantCurrentUA := wantConductanceS * 1e6 * 1.0 // I(µA)=G(µS)*V with V=1.0
	gotCurrentUA := ds.GetRowCurrent(0)

	tol := math.Max(wantCurrentUA*1e-9, 1e-12)
	if math.Abs(gotCurrentUA-wantCurrentUA) > tol {
		t.Fatalf("geometry scaling not reflected in ideal compute current: got=%g µA want=%g µA scale=%g", gotCurrentUA, wantCurrentUA, scale)
	}
}
