//go:build legacy_fyne

package gui

import (
	"os"
	"strings"
	"testing"
)

func TestResolveCSVColumnsFromEnv_DefaultFallback(t *testing.T) {
	t.Setenv("FECIM_EXPORT_COLUMNS", "")
	cols := resolveCSVColumnsFromEnv()
	if got, want := strings.Join(cols, ","), "e_field_mv_cm,polarization_uc_cm2"; got != want {
		t.Fatalf("default columns = %q, want %q", got, want)
	}
}

func TestResolveCSVColumnsFromEnv_FilterInvalid(t *testing.T) {
	t.Setenv("FECIM_EXPORT_COLUMNS", "index,foo,e_field_v_m,polarization_c_m2,index")
	cols := resolveCSVColumnsFromEnv()
	if got, want := strings.Join(cols, ","), "index,e_field_v_m,polarization_c_m2"; got != want {
		t.Fatalf("columns = %q, want %q", got, want)
	}
}

func TestBuildCSVContent_CustomColumns(t *testing.T) {
	_ = os.Setenv("FECIM_EXPORT_COLUMNS", "")
	content, err := buildCSVContent([]float64{1.0}, []float64{2.0}, []string{"index", "e_field_v_m", "polarization_c_m2"})
	if err != nil {
		t.Fatalf("buildCSVContent error: %v", err)
	}
	out := string(content)
	if !strings.Contains(out, "Index,E_field_V_m,Polarization_C_m2") {
		t.Fatalf("missing expected header in %q", out)
	}
	if !strings.Contains(out, "0,1.000000e+00,2.000000e+00") {
		t.Fatalf("missing expected data row in %q", out)
	}
}
