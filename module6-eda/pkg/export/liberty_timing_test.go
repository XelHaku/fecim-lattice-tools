package export

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

func extractScalarValue(t *testing.T, lib, key string) float64 {
	t.Helper()
	re := regexp.MustCompile(key + `\(scalar\)\s*\{\s*values\("([0-9.]+)"\)\s*;\s*\}`)
	m := re.FindStringSubmatch(lib)
	if len(m) < 2 {
		t.Fatalf("failed to find %s scalar value", key)
	}
	v, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		t.Fatalf("failed to parse %s scalar value %q: %v", key, m[1], err)
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

func TestGenerateLiberty_TimingAndPins(t *testing.T) {
	cfg := config.DefaultCellConfig()
	lib := GenerateLiberty(cfg)

	rise := extractScalarValue(t, lib, "cell_rise")
	fall := extractScalarValue(t, lib, "cell_fall")
	if rise <= 0 {
		t.Fatalf("cell_rise must be > 0, got %v", rise)
	}
	if rise <= fall {
		t.Fatalf("expected write (cell_rise) slower than read (cell_fall): rise=%v fall=%v", rise, fall)
	}

	area := extractArea(t, lib)
	if area <= 0 {
		t.Fatalf("area must be > 0, got %v", area)
	}

	requiredPins := []string{"pin(WL)", "pin(BL)", "pin(VPWR)", "pin(VGND)"}
	for _, pin := range requiredPins {
		if !strings.Contains(lib, pin) {
			t.Fatalf("missing required pin %s", pin)
		}
	}
}
