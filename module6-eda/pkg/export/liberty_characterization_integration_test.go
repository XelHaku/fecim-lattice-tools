package export

import (
	"strings"
	"testing"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	"fecim-lattice-tools/module6-eda/pkg/config"
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

	lib := GenerateLibertyFromCharacterization(config.DefaultCellConfig(), &char)
	if !strings.Contains(lib, "cell_rise(fecim_nldm_7x7)") {
		t.Fatal("expected liberty NLDM timing block")
	}
	if got := extractFirstNLDMValue(t, lib, "cell_rise"); got <= 0 {
		t.Fatalf("expected characterized positive timing, got %.3f", got)
	}
}
