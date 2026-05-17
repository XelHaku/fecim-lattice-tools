//go:build legacy_fyne

package gui

import (
	"strings"
	"testing"
)

// TestMarketDataNoMythic verifies no "Mythic" string appears in competitors
func TestMarketDataNoMythic(t *testing.T) {
	for _, comp := range competitors {
		if strings.Contains(comp.Name, "Mythic") {
			t.Errorf("Found 'Mythic' in competitor name: %s", comp.Name)
		}
	}
}

// TestMarketDataHasIBM verifies IBM Analog AI is in competitors
func TestMarketDataHasIBM(t *testing.T) {
	foundIBM := false
	for _, comp := range competitors {
		if strings.Contains(comp.Name, "IBM") && strings.Contains(comp.Name, "Analog") {
			foundIBM = true
			break
		}
	}
	if !foundIBM {
		t.Error("IBM Analog AI not found in competitors list")
	}
}

// TestMarketSegmentYears verifies MarketSegment has Y2024, Y2026, Y2030 fields (not Y2025)
func TestMarketSegmentYears(t *testing.T) {
	// Test that we can create a MarketSegment with the expected fields
	seg := MarketSegment{
		Name:  "Test",
		Y2024: 100.0,
		Y2026: 150.0,
		Y2030: 200.0,
	}

	// Verify the fields exist and can be read
	if seg.Y2024 != 100.0 {
		t.Errorf("Y2024 field not working correctly, got %f, want 100.0", seg.Y2024)
	}
	if seg.Y2026 != 150.0 {
		t.Errorf("Y2026 field not working correctly, got %f, want 150.0", seg.Y2026)
	}
	if seg.Y2030 != 200.0 {
		t.Errorf("Y2030 field not working correctly, got %f, want 200.0", seg.Y2030)
	}

	// Verify actual market data uses these fields
	for _, mkt := range marketData {
		if mkt.Y2024 == 0 && mkt.Y2026 == 0 && mkt.Y2030 == 0 {
			t.Errorf("Market segment %s has zero values for all years", mkt.Name)
		}
		if mkt.Y2024 > mkt.Y2030 {
			t.Errorf("Market segment %s has Y2024 > Y2030 (expecting growth)", mkt.Name)
		}
	}
}

// TestCompetitorFeCIMHighlighted verifies FeCIM competitor has Highlight=true
func TestCompetitorFeCIMHighlighted(t *testing.T) {
	foundFeCIM := false
	for _, comp := range competitors {
		if comp.Name == "FeCIM" {
			foundFeCIM = true
			if !comp.Highlight {
				t.Error("FeCIM competitor should have Highlight=true")
			}
			break
		}
	}
	if !foundFeCIM {
		t.Error("FeCIM not found in competitors list")
	}
}

func TestCompetitorScoresAreConfidenceBounded(t *testing.T) {
	for _, comp := range competitors {
		scores := []float64{comp.Energy, comp.Speed, comp.Endurance, comp.CMOS, comp.Scalable}
		for _, score := range scores {
			if score < 0 || score > 1 {
				t.Fatalf("score out of bounds for %s: %.3f", comp.Name, score)
			}
		}
	}
}
