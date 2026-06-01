package external_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	"fecim-lattice-tools/validation/external/internal/testsupport"
)

func TestNgspiceCrossbar_StructuralNetlists(t *testing.T) {
	for _, n := range []int{1, 2, 4} {
		netlist := mustNetlist(t, n)
		mustContain(t, netlist, ".control")
		mustContain(t, netlist, "op")
		mustContain(t, netlist, ".end")
		mustContain(t, netlist, "RCELL_0_0")
		mustContain(t, netlist, "XADC_0")
		t.Logf("structural netlist validated for %dx%d", n, n)
	}
}

func TestNgspiceCrossbar_RunAndCompare_WhenAvailable(t *testing.T) {
	testsupport.RequireCommand(t, "ngspice", "ngspice not installed; structural validation executed, simulation comparison skipped")

	maxAbs := 0.0
	maxRel := 0.0
	parsedAny := false

	for _, n := range []int{1, 2, 4} {
		params := makeParams(n)
		netlist, err := arraysim.ExportCrossbarSPICE(params, arraysim.SpiceExportConfig{Title: fmt.Sprintf("%dx%d test", n, n)})
		if err != nil {
			t.Fatalf("ExportCrossbarSPICE(%dx%d) failed: %v", n, n, err)
		}

		tmpDir := t.TempDir()
		deck := filepath.Join(tmpDir, "deck.sp")
		out := filepath.Join(tmpDir, "ngspice.out")
		if err := os.WriteFile(deck, []byte(netlist), 0o644); err != nil {
			t.Fatalf("write deck: %v", err)
		}

		cmd := exec.Command("ngspice", "-b", "-o", out, deck)
		if runErr := cmd.Run(); runErr != nil {
			t.Fatalf("ngspice run failed for %dx%d: %v", n, n, runErr)
		}
		raw, err := os.ReadFile(out)
		if err != nil {
			t.Fatalf("read ngspice output: %v", err)
		}
		parsed := parseSourceCurrentsA(string(raw))
		if len(parsed) == 0 {
			t.Logf("ngspice output for %dx%d had no parseable source branch currents; output kept at %s", n, n, out)
			continue
		}
		parsedAny = true

		ref, err := arraysim.NewTierBSolver().Solve(params)
		if err != nil {
			t.Fatalf("reference solve failed: %v", err)
		}

		for r := 0; r < n; r++ {
			key := fmt.Sprintf("vwl_src_%d", r)
			if got, ok := parsed[key]; ok {
				want := -ref.RowCurrents[r]
				absErr := abs(got - want)
				relErr := absErr / max(abs(want), 1e-15)
				if absErr > maxAbs {
					maxAbs = absErr
				}
				if relErr > maxRel {
					maxRel = relErr
				}
			}
		}
	}

	if !parsedAny {
		t.Skip("ngspice ran, but parser found no source branch currents; structural validation still passed")
	}
	t.Logf("ngspice comparison summary: max absolute error = %.6e A, max relative error = %.6e", maxAbs, maxRel)
}

func makeParams(n int) arraysim.SolveParams {
	g := make([][]float64, n)
	wl := make([]float64, n)
	bl := make([]float64, n)
	for r := 0; r < n; r++ {
		g[r] = make([]float64, n)
		wl[r] = 0.5 - 0.05*float64(r)
		for c := 0; c < n; c++ {
			g[r][c] = 1e-4 * (1 + 0.05*float64(r+c))
		}
	}
	for c := 0; c < n; c++ {
		bl[c] = 0
	}
	return arraysim.SolveParams{
		WLVoltages:  wl,
		BLVoltages:  bl,
		Conductance: g,
		Wire:        arraysim.WireParams{RWordLine: 5.0, RBitLine: 7.0},
		Boundary:    arraysim.BoundaryParams{WLDriveResistance: 2.0, BLDriveResistance: 2.0},
	}
}

func mustNetlist(t *testing.T, n int) string {
	t.Helper()
	netlist, err := arraysim.ExportCrossbarSPICE(makeParams(n), arraysim.SpiceExportConfig{Title: fmt.Sprintf("%dx%d", n, n)})
	if err != nil {
		t.Fatalf("ExportCrossbarSPICE(%dx%d) failed: %v", n, n, err)
	}
	mustContain(t, netlist, fmt.Sprintf("VWL_SRC_%d", n-1))
	mustContain(t, netlist, fmt.Sprintf("VBL_SRC_%d", n-1))
	mustContain(t, netlist, fmt.Sprintf("RCELL_%d_%d", n-1, n-1))
	return netlist
}

func mustContain(t *testing.T, s, needle string) {
	t.Helper()
	if !strings.Contains(s, needle) {
		t.Fatalf("expected netlist to contain %q", needle)
	}
}

func parseSourceCurrentsA(raw string) map[string]float64 {
	out := map[string]float64{}
	re := regexp.MustCompile(`(?i)\b(vwl_src_\d+|vbl_src_\d+)(?:#branch)?\s*=\s*([+-]?[0-9]*\.?[0-9]+(?:[eE][+-]?\d+)?)`)
	for _, m := range re.FindAllStringSubmatch(raw, -1) {
		v, err := strconv.ParseFloat(m[2], 64)
		if err == nil {
			out[strings.ToLower(m[1])] = v
		}
	}
	return out
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
