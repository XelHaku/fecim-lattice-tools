// Package integration provides cross-module integration tests.
// This test was moved from module6-eda/pkg/export/ because it imports
// module4-circuits/pkg/arraysim, violating the rule that modules should
// only import from shared/.
package integration

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	"fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/module6-eda/pkg/export"
)

func TestEndToEnd_M4TransientCharacterizationFeedsM6Liberty(t *testing.T) {
	arrCfg := arraysim.ArrayConfig{}
	res := arraysim.TransientResult{
		TimeNs:       []float64{1, 2, 3, 4},
		Polarization: []float64{0.1, 0.2, 0.95, 1.0},
		Current:      []float64{1e-6, 1.3e-6, 1.30001e-6, 1.30001e-6},
		Energy_fJ:    15,
	}
	char := arraysim.CharacterizeTransientResult(arrCfg, res)

	lib := export.GenerateLibertyFromCharacterization(config.DefaultCellConfig(), &char)
	if !strings.Contains(lib, "cell_rise(fecim_nldm_7x7)") {
		t.Fatal("expected liberty NLDM timing block")
	}
	if got := extractFirstNLDMValueInteg(t, lib, "cell_rise"); got <= 0 {
		t.Fatalf("expected characterized positive timing, got %.3f", got)
	}
}

// extractFirstNLDMValueInteg is a local copy of the test helper from
// module6-eda/pkg/export (not exported, so it must be duplicated here).
func extractFirstNLDMValueInteg(t *testing.T, lib, key string) float64 {
	t.Helper()
	re := regexp.MustCompile(key + `\(fecim_nldm_7x7\)\s*\{\s*values\(\\\n\s*"([0-9.]+),`)
	m := re.FindStringSubmatch(lib)
	if len(m) < 2 {
		t.Fatalf("failed to find %s nldm first value", key)
	}
	v, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		t.Fatalf("failed to parse %s value %q: %v", key, m[1], err)
	}
	return v
}
