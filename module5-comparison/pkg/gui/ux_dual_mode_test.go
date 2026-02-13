package gui

import "testing"

func TestM5UX01_DualModeToggleValues(t *testing.T) {
	if PresentationModeFromString("Engineer") != PresentationModeEngineer {
		t.Fatal("Engineer mode should parse")
	}
	if PresentationModeFromString("Investor") != PresentationModeInvestor {
		t.Fatal("Investor mode should parse")
	}
}
