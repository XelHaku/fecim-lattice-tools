//go:build legacy_fyne

package gui

import (
	"math"
	"testing"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

func TestLevelToConductance_GeometryScaling(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(4, 4)
	mat := sharedphysics.FeCIMMaterial()
	ds.SetMaterial(mat)

	refGeom := sharedphysics.GeometryFromMaterial(mat)

	// Increase A/t by changing area and thickness.
	geom := arraysim.DefaultCellGeometry()
	geom.Film = refGeom
	geom.Film.Area = refGeom.Area * 2
	geom.Film.Thickness = refGeom.Thickness / 2
	ds.SetCellGeometry(geom)

	actualGeom := ds.GetCellGeometry()
	scale := actualGeom.Film.ConductanceScale(refGeom)
	if scale <= 1.0 {
		t.Fatalf("expected conductance scale >1 after geometry update, got scale=%g (A=%g t=%g, refA=%g reft=%g)", scale, actualGeom.Film.Area, actualGeom.Film.Thickness, refGeom.Area, refGeom.Thickness)
	}

	levels := mat.GetNumLevels()
	level := levels / 2

	baseG := mat.DiscreteLevel(level, levels)
	want := baseG * scale
	got := ds.levelToConductance(level, levels)

	if math.Abs(got-want) > want*1e-12 {
		t.Fatalf("geometry-scaled conductance mismatch: got=%g S want=%g S scale=%g", got, want, scale)
	}

	gmin, gmax := ds.conductanceBounds()
	if gmin <= mat.Gmin || gmax <= mat.Gmax {
		t.Fatalf("expected scaled conductance bounds to increase with area/thickness scaling: gmin=%g gmax=%g", gmin, gmax)
	}
}
