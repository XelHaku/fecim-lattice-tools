package export

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	"fecim-lattice-tools/module6-eda/pkg/config"
)

func extractFirstNLDMValue(t *testing.T, lib, key string) float64 {
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

func extractArea(t *testing.T, lib string) float64 {
	t.Helper()
	re := regexp.MustCompile(`area\s*:\s*([0-9.]+)\s*;`)
	m := re.FindStringSubmatch(lib)
	if len(m) < 2 {
		t.Fatal("failed to find area")
	}
	v, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		t.Fatalf("failed to parse area %q: %v", m[1], err)
	}
	return v
}

func TestGenerateLiberty_NLDMTablesAndPins(t *testing.T) {
	cfg := config.DefaultCellConfig()
	lib := GenerateLiberty(cfg)

	if !strings.Contains(lib, "lu_table_template(fecim_nldm_7x7)") {
		t.Fatal("expected NLDM template")
	}
	if !strings.Contains(lib, "rise_transition(fecim_nldm_7x7)") || !strings.Contains(lib, "fall_transition(fecim_nldm_7x7)") {
		t.Fatal("expected transition NLDM tables")
	}

	rise := extractFirstNLDMValue(t, lib, "cell_rise")
	fall := extractFirstNLDMValue(t, lib, "cell_fall")
	if rise <= fall {
		t.Fatalf("expected write slower than read: rise=%v fall=%v", rise, fall)
	}

	if area := extractArea(t, lib); area <= 0 {
		t.Fatalf("area must be > 0, got %v", area)
	}
}

func TestGenerateMultiCornerLiberty_Contains9CornersAndScalesTiming(t *testing.T) {
	lib := GenerateMultiCornerLiberty(config.DefaultCellConfig())
	corners := []string{"ff_n40c", "ff_25c", "ff_125c", "tt_n40c", "tt_25c", "tt_125c", "ss_n40c", "ss_25c", "ss_125c"}
	for _, c := range corners {
		if !strings.Contains(lib, "operating_conditions("+c+")") {
			t.Fatalf("missing corner %s", c)
		}
	}
	if strings.Count(lib, "operating_conditions(") != 9 {
		t.Fatalf("expected 9 operating_conditions blocks, got %d", strings.Count(lib, "operating_conditions("))
	}

	ff := extractFirstNLDMValue(t, lib, "cell_rise")
	ssIdx := strings.LastIndex(lib, "library(fecim_cells_ss_125c)")
	if ssIdx < 0 {
		t.Fatal("missing SS 125C library")
	}
	ss := extractFirstNLDMValue(t, lib[ssIdx:], "cell_rise")
	if !(ss > ff) {
		t.Fatalf("expected slow corner timing > fast corner timing, got ss=%.3f ff=%.3f", ss, ff)
	}
}

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

func TestGenerateLibertyWithModule4Energy_WritesInternalPowerGroups(t *testing.T) {
	cfg := config.DefaultCellConfig()
	energy := &Module4EnergyModel{DACEnergyJ: 14.4e-15, MVMEnergyJ: 25.0e-15, TIAEnergyJ: 6.3e-15}

	lib := GenerateLibertyWithModule4Energy(cfg, energy)
	if strings.Count(lib, "internal_power()") < 3 {
		t.Fatalf("expected at least 3 internal_power groups, got %d", strings.Count(lib, "internal_power()"))
	}
	if !strings.Contains(lib, "related_pin : \"WL\"") {
		t.Fatal("missing WL internal_power group (DAC)")
	}
	if strings.Count(lib, "related_pin : \"BL\"") < 2 {
		t.Fatal("missing BL internal_power groups (MVM/TIA)")
	}
}
