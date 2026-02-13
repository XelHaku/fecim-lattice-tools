package export

import (
	"regexp"
	"sort"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	"fecim-lattice-tools/module6-eda/pkg/config"
)

func TestRoundtripCrossFormatValidation(t *testing.T) {
	t.Parallel()

	cellCfg := config.DefaultCellConfig()
	arrCfg := config.DefaultArrayConfig()

	// 1) Generate all requested EDA outputs for default configurations
	lefPassive := GenerateLEF(cellCfg)
	libPassive := GenerateLiberty(cellCfg)

	cellCfg1 := cellCfg
	cellCfg1.CellType = "1t1r"
	lef1T1R := GenerateLEF(cellCfg1)
	lib1T1R := GenerateLiberty(cellCfg1)

	cellCfg2 := cellCfg
	cellCfg2.CellType = "2t1r"
	lef2T1R := GenerateLEF(cellCfg2)
	lib2T1R := GenerateLiberty(cellCfg2)

	cellVerilog := GenerateCellVerilog(cellCfg)
	arrayVerilog := GenerateArrayVerilog(arrCfg)

	spicePassive := generateSPICEForArch(t, compiler.ArchPassive)
	spice1T1R := generateSPICEForArch(t, compiler.Arch1T1R)
	spice2T1R := generateSPICEForArch(t, compiler.Arch2T1R)

	// 2) Per-format content validation
	assertLEFCore(t, lefPassive)
	assertLEFCore(t, lef1T1R)
	assertLEFCore(t, lef2T1R)

	assertLibertyCore(t, libPassive)
	assertLibertyCore(t, lib1T1R)
	assertLibertyCore(t, lib2T1R)

	assertVerilogCore(t, cellVerilog)
	assertVerilogCore(t, arrayVerilog)

	assertSPICECore(t, spicePassive)
	assertSPICECore(t, spice1T1R)
	assertSPICECore(t, spice2T1R)

	// 3) Cross-format LVS-like pin consistency checks
	assertCrossFormatPins(t, "passive", lefPassive, cellVerilog, spicePassive, []string{"WL", "BL", "VPWR", "VGND"})
	assertCrossFormatPins(t, "1t1r", lef1T1R, GenerateCellVerilog(cellCfg1), spice1T1R, []string{"WL", "BL", "SL", "VPWR", "VGND"})
	assertCrossFormatPins(t, "2t1r", lef2T1R, GenerateCellVerilog(cellCfg2), spice2T1R, []string{"WL", "BL", "SL", "CSL", "VPWR", "VGND"})

	// 4) Explicit architecture-specific pin requirements
	assertContainsPin(t, extractLEFPins(lef1T1R), "SL", "LEF 1T1R")
	assertContainsPin(t, extractVerilogPorts(GenerateCellVerilog(cellCfg1)), "SL", "Verilog 1T1R")
	assertContainsPin(t, extractSPICEPins(spice1T1R), "SL", "SPICE 1T1R")

	assertContainsPin(t, extractLEFPins(lef2T1R), "CSL", "LEF 2T1R")
	assertContainsPin(t, extractVerilogPorts(GenerateCellVerilog(cellCfg2)), "CSL", "Verilog 2T1R")
	assertContainsPin(t, extractSPICEPins(spice2T1R), "CSL", "SPICE 2T1R")
}

func generateSPICEForArch(t *testing.T, arch string) string {
	t.Helper()
	cfg := compiler.NewStorageConfig(4, 4)
	switch arch {
	case compiler.Arch1T1R:
		cfg.With1T1R()
	case compiler.Arch2T1R:
		cfg.With2T1R()
	}
	design, err := compiler.GenerateDesign(cfg)
	if err != nil {
		t.Fatalf("GenerateDesign(%s) failed: %v", arch, err)
	}
	return GenerateSPICE(design, 1.8)
}

func assertLEFCore(t *testing.T, lef string) {
	t.Helper()
	for _, want := range []string{"MACRO", "SIZE", "PIN WL", "PIN BL", "PIN VPWR", "PIN VGND"} {
		if !strings.Contains(lef, want) {
			t.Fatalf("LEF missing %q", want)
		}
	}
}

func assertLibertyCore(t *testing.T, lib string) {
	t.Helper()
	for _, want := range []string{"cell(", "pin(WL)", "pin(BL)", "timing()"} {
		if !strings.Contains(lib, want) {
			t.Fatalf("Liberty missing %q", want)
		}
	}
}

func assertVerilogCore(t *testing.T, v string) {
	t.Helper()
	for _, want := range []string{"module", "input", "WL", "BL"} {
		if !strings.Contains(v, want) {
			t.Fatalf("Verilog missing %q", want)
		}
	}
	if !strings.Contains(v, "output") && !strings.Contains(v, "inout") {
		t.Fatalf("Verilog missing BL direction (expected output or inout)")
	}
}

func assertSPICECore(t *testing.T, spice string) {
	t.Helper()
	if !strings.Contains(spice, "VDD") {
		t.Fatal("SPICE missing VDD")
	}
	if !regexp.MustCompile(`(?m)^(?:R_|X_)\d*_?\d*_?\d*\s+`).MatchString(spice) {
		t.Fatal("SPICE missing at least one cell element")
	}
	if !strings.Contains(strings.ToLower(spice), ".end") {
		t.Fatal("SPICE missing .end")
	}
}

func assertCrossFormatPins(t *testing.T, label, lef, verilog, spice string, expected []string) {
	t.Helper()
	lefPins := extractLEFPins(lef)
	verilogPins := extractVerilogPorts(verilog)
	spicePins := extractSPICEPins(spice)
	expectedSet := toSet(expected)

	assertSetEqual(t, label+" LEF", lefPins, expectedSet)
	assertSetEqual(t, label+" Verilog", verilogPins, expectedSet)
	assertSetEqual(t, label+" SPICE", spicePins, expectedSet)
}

func extractLEFPins(lef string) map[string]struct{} {
	re := regexp.MustCompile(`(?m)^\s*PIN\s+([A-Za-z_][A-Za-z0-9_]*)`)
	m := make(map[string]struct{})
	for _, g := range re.FindAllStringSubmatch(lef, -1) {
		m[strings.ToUpper(g[1])] = struct{}{}
	}
	return m
}

func extractVerilogPorts(v string) map[string]struct{} {
	re := regexp.MustCompile(`(?m)^\s*(?:input|output|inout)\s+(?:wire\s+)?(?:\[[^\]]+\]\s+)?([A-Za-z_][A-Za-z0-9_]*)`)
	m := make(map[string]struct{})
	for _, g := range re.FindAllStringSubmatch(v, -1) {
		m[strings.ToUpper(g[1])] = struct{}{}
	}
	return m
}

func extractSPICEPins(spice string) map[string]struct{} {
	reSubckt := regexp.MustCompile(`(?im)^\.subckt\s+(\S+)\s+(.+)$`)
	all := reSubckt.FindAllStringSubmatch(spice, -1)
	for _, g := range all {
		if len(g) == 3 && strings.EqualFold(g[1], "fecim_cell") {
			fields := strings.Fields(g[2])
			m := make(map[string]struct{}, len(fields))
			for _, f := range fields {
				m[strings.ToUpper(f)] = struct{}{}
			}
			return m
		}
	}
	for _, g := range all {
		if len(g) != 3 || !regexp.MustCompile(`(?i)\bWL\b`).MatchString(g[2]) {
			continue
		}
		fields := strings.Fields(g[2])
		m := make(map[string]struct{}, len(fields))
		for _, f := range fields {
			m[strings.ToUpper(f)] = struct{}{}
		}
		return m
	}

	// Fallback inference if subckt header is absent.
	m := make(map[string]struct{})
	if regexp.MustCompile(`(?im)\bwl\d+\b`).MatchString(spice) {
		m["WL"] = struct{}{}
	}
	if regexp.MustCompile(`(?im)\bbl\d+\b`).MatchString(spice) {
		m["BL"] = struct{}{}
	}
	if regexp.MustCompile(`(?im)\bvdd\b`).MatchString(spice) {
		m["VPWR"] = struct{}{}
	}
	if strings.Contains(spice, " 0 ") || strings.Contains(spice, "(0") {
		m["VGND"] = struct{}{}
	}
	if regexp.MustCompile(`(?im)\bsl\d+\b`).MatchString(spice) {
		m["SL"] = struct{}{}
	}
	if regexp.MustCompile(`(?im)\bcsl\d+\b`).MatchString(spice) {
		m["CSL"] = struct{}{}
	}
	return m
}

func toSet(in []string) map[string]struct{} {
	m := make(map[string]struct{}, len(in))
	for _, s := range in {
		m[strings.ToUpper(s)] = struct{}{}
	}
	return m
}

func assertSetEqual(t *testing.T, label string, got, want map[string]struct{}) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s pin count mismatch: got=%v want=%v", label, sortedKeys(got), sortedKeys(want))
	}
	for k := range want {
		if _, ok := got[k]; !ok {
			t.Fatalf("%s missing pin %s (got=%v)", label, k, sortedKeys(got))
		}
	}
}

func assertContainsPin(t *testing.T, set map[string]struct{}, pin, label string) {
	t.Helper()
	if _, ok := set[strings.ToUpper(pin)]; !ok {
		t.Fatalf("%s missing required pin %s (got=%v)", label, pin, sortedKeys(set))
	}
}

func sortedKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
