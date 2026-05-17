//go:build legacy_fyne

package widgets

import (
	"strings"
	"testing"
)

func TestAboutScienceData_CoversCoreTopics(t *testing.T) {
	sections := aboutScienceData()
	if len(sections) < 5 {
		t.Fatalf("expected at least 5 science sections, got %d", len(sections))
	}

	joined := ""
	for _, s := range sections {
		joined += strings.ToLower(s.Title+" "+s.Summary+" "+s.Details) + "\n"
		if strings.TrimSpace(s.Details) == "" {
			t.Fatalf("section %q has empty details", s.Title)
		}
	}

	required := []string{"fecim", "hzo", "hysteresis", "crossbar", "neuromorphic"}
	for _, term := range required {
		if !strings.Contains(joined, term) {
			t.Fatalf("missing required term %q in about science content", term)
		}
	}
}
