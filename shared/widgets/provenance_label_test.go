//go:build legacy_fyne

package widgets

import (
	"testing"

	"fyne.io/fyne/v2/test"

	"fecim-lattice-tools/shared/physics"
)

func TestNewProvenanceLabel_AllLevels(t *testing.T) {
	test.NewApp()

	cases := []struct {
		prov    physics.Provenance
		display string
	}{
		{physics.ProvenanceMeasured, "Measured"},
		{physics.ProvenanceCalibrated, "Calibrated"},
		{physics.ProvenanceEstimated, "Estimated"},
		{physics.ProvenancePlaceholder, "Placeholder"},
	}

	for _, tc := range cases {
		pl := NewProvenanceLabel("24.5 uC/cm2", tc.prov, "Some ref")
		if pl == nil {
			t.Fatalf("nil for provenance %s", tc.prov)
		}
		if pl.FormattedValue != "24.5 uC/cm2" {
			t.Fatalf("expected value %q, got %q", "24.5 uC/cm2", pl.FormattedValue)
		}
		if pl.Provenance != tc.prov {
			t.Fatalf("expected provenance %q, got %q", tc.prov, pl.Provenance)
		}
		if provenanceDisplayName(tc.prov) != tc.display {
			t.Fatalf("expected display %q, got %q", tc.display, provenanceDisplayName(tc.prov))
		}
		// Renderer must not panic.
		r := pl.CreateRenderer()
		if r == nil {
			t.Fatalf("nil renderer for %s", tc.prov)
		}
	}
}

func TestNewProvenanceLabel_DefaultsPlaceholder(t *testing.T) {
	test.NewApp()

	pl := NewProvenanceLabel("0.0", "", "")
	if pl.Provenance != physics.ProvenancePlaceholder {
		t.Fatalf("expected placeholder, got %q", pl.Provenance)
	}
}

func TestNewProvenanceLabel_NoSourceRef(t *testing.T) {
	test.NewApp()

	pl := NewProvenanceLabel("10 MV/cm", physics.ProvenanceMeasured, "")
	if pl.SourceRef != "" {
		t.Fatalf("expected empty source ref, got %q", pl.SourceRef)
	}
	// Renderer should still work without source.
	r := pl.CreateRenderer()
	if r == nil {
		t.Fatal("nil renderer without source ref")
	}
}

func TestProvenanceColor_Distinct(t *testing.T) {
	colors := []physics.Provenance{
		physics.ProvenanceMeasured,
		physics.ProvenanceCalibrated,
		physics.ProvenanceEstimated,
		physics.ProvenancePlaceholder,
	}
	seen := make(map[[4]uint8]physics.Provenance)
	for _, p := range colors {
		c := provenanceColor(p)
		r, g, b, a := c.RGBA()
		key := [4]uint8{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
		if prev, dup := seen[key]; dup {
			t.Fatalf("duplicate colour for %s and %s", prev, p)
		}
		seen[key] = p
	}
}
