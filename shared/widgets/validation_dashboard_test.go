//go:build legacy_fyne

package widgets

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestNewValidationDashboard_NotNil(t *testing.T) {
	test.NewApp()
	vd := NewValidationDashboard()
	if vd == nil {
		t.Fatal("expected non-nil dashboard")
	}
	if vd.Container == nil {
		t.Fatal("expected non-nil container")
	}
}

func TestValidationDashboard_ClaimCount(t *testing.T) {
	test.NewApp()
	vd := NewValidationDashboard()
	if len(vd.Claims) == 0 {
		t.Fatal("expected at least one claim")
	}
	// CLAIMS-MATRIX (16 internal) + literature comparisons (3 [CHECK]) + literature benchmarks (4 [LIT])
	if len(vd.Claims) != 23 {
		t.Fatalf("expected 23 claims, got %d", len(vd.Claims))
	}
}

func TestValidationDashboard_StatusValues(t *testing.T) {
	test.NewApp()
	vd := NewValidationDashboard()

	passCount, untestedCount := 0, 0
	for _, c := range vd.Claims {
		switch c.Status {
		case ClaimPass:
			passCount++
		case ClaimFail:
			// none expected in current matrix
		case ClaimUntested:
			untestedCount++
		default:
			t.Fatalf("unexpected status %q for claim %q", c.Status, c.Name)
		}
		if c.Name == "" {
			t.Fatal("claim with empty name")
		}
		if c.Tolerance == "" {
			t.Fatal("claim with empty tolerance")
		}
		if c.TestName == "" {
			t.Fatal("claim with empty test name")
		}
	}
	if passCount == 0 {
		t.Fatal("expected at least one Pass claim")
	}
	if untestedCount == 0 {
		t.Fatal("expected at least one Untested claim")
	}
}

func TestValidationDashboard_ProvenanceCoverage(t *testing.T) {
	test.NewApp()
	vd := NewValidationDashboard()
	seen := make(map[ConfidenceLevel]bool)
	for _, c := range vd.Claims {
		seen[c.Provenance] = true
	}
	for _, level := range []ConfidenceLevel{Measured, Calibrated, Estimated} {
		if !seen[level] {
			t.Fatalf("expected at least one claim with provenance %q", level)
		}
	}
}

func TestStatusColor_AllValues(t *testing.T) {
	test.NewApp()
	for _, s := range []ClaimStatusValue{ClaimPass, ClaimFail, ClaimUntested} {
		c := statusColor(s)
		if c == nil {
			t.Fatalf("nil color for status %q", s)
		}
	}
}
