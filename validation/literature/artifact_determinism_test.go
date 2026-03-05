package literature

import (
	"testing"

	sharedphysics "fecim-lattice-tools/shared/physics"
)

func TestFORCMetrics_GeneratedAtDeterministic(t *testing.T) {
	m := computeFORCMetrics(t, "fecim_hzo", sharedphysics.FeCIMMaterial())
	if m.Generated != "1970-01-01T00:00:00Z" {
		t.Fatalf("expected deterministic generated_at, got %q", m.Generated)
	}
}

func TestSwitchingKineticsMetrics_GeneratedAtDeterministic(t *testing.T) {
	m := computeSwitchingKineticsMetrics(t, "fecim_hzo", sharedphysics.FeCIMMaterial())
	if m.Generated != "1970-01-01T00:00:00Z" {
		t.Fatalf("expected deterministic generated_at, got %q", m.Generated)
	}
}
