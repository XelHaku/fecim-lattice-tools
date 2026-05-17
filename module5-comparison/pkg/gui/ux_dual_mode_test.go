//go:build legacy_fyne

package gui

import "testing"

func TestM5UX01_DualModeToggleValues(t *testing.T) {
	if PresentationModeFromString("Engineer") != PresentationModeEngineer {
		t.Fatal("Engineer mode should parse")
	}
	if PresentationModeFromString("Briefing") != PresentationModeBriefing {
		t.Fatal("Briefing mode should parse")
	}
}
