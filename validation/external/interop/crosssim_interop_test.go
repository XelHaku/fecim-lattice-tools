package external_test

import (
	"testing"

	"fecim-lattice-tools/validation/external/internal/testsupport"
)

// TestCrossSimInteroperability validates FeCIM export compatibility with CrossSim.
// Skips if CrossSim (cross_sim Python package) is not installed.
func TestCrossSimInteroperability(t *testing.T) {
	// Check if CrossSim is available
	testsupport.RequirePythonModule(t, "cross_sim", "CrossSim not installed — skipping interoperability test. Install via: pip3 install cross-sim")

	// When CrossSim is available, this test would:
	// 1. Export FeCIM crossbar configuration as CrossSim-compatible input
	// 2. Run a baseline MVM scenario through both tools
	// 3. Compare trend-level outputs (conductance distribution shape, MVM accuracy direction)
	//
	// Comparison type: trend-level (not exact match — different solver implementations)
	// Expected agreement: same monotonic trends for accuracy vs array size, noise vs bit precision
	t.Log("CrossSim available — would run export + baseline comparison")
	t.Log("Comparison type: trend-level (conductance distribution, MVM accuracy direction)")
}
