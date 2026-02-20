package tabs

import (
	"strconv"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	"fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/module6-eda/pkg/export"
)

func TestCLIAndGUI_DEFEquivalenceByTopology(t *testing.T) {
	tests := []struct {
		name       string
		arch       string
		rows, cols int
		wantSL     bool
		wantCSL    bool
	}{
		{"passive", "passive", 4, 3, false, false},
		{"1t1r", "1t1r", 4, 3, true, false},
		{"2t1r", "2t1r", 4, 3, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guiCfg := config.ArrayConfig{Rows: tt.rows, Cols: tt.cols, Architecture: tt.arch, CellWidth: 0.46, CellHeight: 2.72}
			switch tt.arch {
			case "1t1r":
				guiCfg.CellWidth = 0.92
				guiCfg.CellHeight = 4.07
			case "2t1r":
				guiCfg.CellWidth = 1.38
				guiCfg.CellHeight = 4.07
			}
			guiDEF := generateBuilderDEF(guiCfg)

			cliCfg := compiler.NewStorageConfig(tt.rows, tt.cols)
			cliCfg.Architecture = tt.arch
			cliCfg.CellPitch = guiCfg.CellWidth
			cliCfg.RowHeight = guiCfg.CellHeight
			design := compiler.GenerateBlank(cliCfg)
			cliDEF := export.GenerateDEFWithDefaults(design)

			assertTopologyEquivalent(t, guiDEF, cliDEF, tt.rows, tt.cols, tt.wantSL, tt.wantCSL)
		})
	}
}

func assertTopologyEquivalent(t *testing.T, guiDEF, cliDEF string, rows, cols int, wantSL, wantCSL bool) {
	t.Helper()

	wantComponents := rows * cols
	if !strings.Contains(guiDEF, "COMPONENTS "+strconv.Itoa(wantComponents)) {
		t.Fatalf("gui DEF missing component count %d", wantComponents)
	}
	if !strings.Contains(cliDEF, "COMPONENTS "+strconv.Itoa(wantComponents)) {
		t.Fatalf("cli DEF missing component count %d", wantComponents)
	}

	// Count pin declarations only (exclude NETS section references).
	guiPins := section(guiDEF, "PINS", "END PINS")
	cliPins := section(cliDEF, "PINS", "END PINS")

	if strings.Count(guiPins, "- WL[") != rows || strings.Count(cliPins, "- WL[") != rows {
		t.Fatalf("WL pin count mismatch: gui=%d cli=%d rows=%d", strings.Count(guiPins, "- WL["), strings.Count(cliPins, "- WL["), rows)
	}
	if strings.Count(guiPins, "- BL[") != cols || strings.Count(cliPins, "- BL[") != cols {
		t.Fatalf("BL pin count mismatch: gui=%d cli=%d cols=%d", strings.Count(guiPins, "- BL["), strings.Count(cliPins, "- BL["), cols)
	}

	guiHasSL := strings.Contains(guiPins, "- SL[")
	cliHasSL := strings.Contains(cliPins, "- SL[")
	if guiHasSL != wantSL || cliHasSL != wantSL {
		t.Fatalf("SL parity mismatch (want %v): gui=%v cli=%v", wantSL, guiHasSL, cliHasSL)
	}

	guiHasCSL := strings.Contains(guiPins, "- CSL[")
	cliHasCSL := strings.Contains(cliPins, "- CSL[")
	if guiHasCSL != wantCSL || cliHasCSL != wantCSL {
		t.Fatalf("CSL parity mismatch (want %v): gui=%v cli=%v", wantCSL, guiHasCSL, cliHasCSL)
	}
}

func section(s, startToken, endToken string) string {
	start := strings.Index(s, startToken)
	if start < 0 {
		return ""
	}
	end := strings.Index(s[start:], endToken)
	if end < 0 {
		return s[start:]
	}
	return s[start : start+end]
}
