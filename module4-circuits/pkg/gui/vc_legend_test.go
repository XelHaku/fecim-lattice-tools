//go:build legacy_fyne

package gui

import (
	"reflect"
	"strings"
	"testing"
)

func TestVCLegend_ShowsUnitString(t *testing.T) {
	spec := (&CircuitsApp{}).currentVCLegendSpec()
	if !strings.Contains(spec.Title, "(V)") {
		t.Fatalf("legend title missing unit V: %q", spec.Title)
	}
}

func TestVCLegend_RangeSymmetricAroundZero(t *testing.T) {
	spec := (&CircuitsApp{}).currentVCLegendSpec()
	if spec.Max <= 0 {
		t.Fatalf("expected positive max, got %.6f", spec.Max)
	}
	if got := spec.Min + spec.Max; got > 1e-9 || got < -1e-9 {
		t.Fatalf("expected symmetric range around 0, min=%.6f max=%.6f", spec.Min, spec.Max)
	}
	if !reflect.DeepEqual(spec.TickText, []string{"-Vmax", "0", "+Vmax"}) {
		t.Fatalf("unexpected VC tick labels: %#v", spec.TickText)
	}
	if spec.SignText != "+ = BL>WL" {
		t.Fatalf("unexpected VC sign semantics label: %q", spec.SignText)
	}
}

func TestVCOverlayColor_NeutralAtZero(t *testing.T) {
	c := vcOverlayColor(0, 2.0)
	if !(c.R >= 220 && c.G >= 220 && c.B >= 220) {
		t.Fatalf("expected near-neutral color at 0V, got RGB=(%d,%d,%d)", c.R, c.G, c.B)
	}
}

func TestVCOverlayColor_WarmAndCoolAtExtremes(t *testing.T) {
	warm := vcOverlayColor(+2.0, 2.0)
	cool := vcOverlayColor(-2.0, 2.0)

	if !(warm.R > warm.B && warm.R > warm.G) {
		t.Fatalf("expected warm color at +max, got RGB=(%d,%d,%d)", warm.R, warm.G, warm.B)
	}
	if !(cool.B > cool.R && cool.B > cool.G) {
		t.Fatalf("expected cool color at -max, got RGB=(%d,%d,%d)", cool.R, cool.G, cool.B)
	}
}
