package isppconv

import "testing"

func TestRecoverCollapsedBoundsWidensDirectionallyWhenMoreFieldNeeded(t *testing.T) {
	bounds := Bounds{Min: 1.20, Max: 1.00, MinSet: true, MaxSet: true}

	recovered := RecoverCollapsedBounds(bounds, RecoveryInput{
		NeedMore:         true,
		CurrentMagnitude: 1.10,
		MaxMagnitude:     2.50,
		MinimumWidth:     0.04,
	})

	if !recovered.Changed {
		t.Fatal("expected collapsed bounds recovery to report a change")
	}
	if recovered.Bounds.Min < 1.10 {
		t.Fatalf("recovered minimum = %.3f, want to preserve at least current safe field 1.10", recovered.Bounds.Min)
	}
	if recovered.Bounds.Max <= recovered.Bounds.Min {
		t.Fatalf("recovered bounds still collapsed: min=%.3f max=%.3f", recovered.Bounds.Min, recovered.Bounds.Max)
	}
	if got := recovered.Bounds.Max - recovered.Bounds.Min; got < 0.04 {
		t.Fatalf("recovered width = %.3f, want at least 0.04", got)
	}
	if recovered.ResetToFullRange {
		t.Fatal("directional recovery must not reset to full range")
	}
}

func TestRecoverCollapsedBoundsResetsToFullRangeWhenDirectionUnknown(t *testing.T) {
	bounds := Bounds{Min: 1.20, Max: 1.00, MinSet: true, MaxSet: true}

	recovered := RecoverCollapsedBounds(bounds, RecoveryInput{
		CurrentMagnitude: 1.10,
		MaxMagnitude:     2.50,
		MinimumWidth:     0.04,
	})

	if !recovered.Changed {
		t.Fatal("expected collapsed bounds recovery to report a change")
	}
	if !recovered.ResetToFullRange {
		t.Fatal("unknown-direction recovery should explicitly report full-range reset")
	}
	if recovered.Bounds.Min != 0 || recovered.Bounds.Max != 2.50 {
		t.Fatalf("recovered bounds = [%.3f, %.3f], want [0, 2.50]", recovered.Bounds.Min, recovered.Bounds.Max)
	}
	if recovered.Bounds.MinSet || recovered.Bounds.MaxSet {
		t.Fatalf("full-range recovery should clear bound evidence flags: %+v", recovered.Bounds)
	}
}
