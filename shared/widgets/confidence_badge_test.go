package widgets

import (
	"testing"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestNewConfidenceBadge_Levels(t *testing.T) {
	test.NewApp()
	cases := []ConfidenceLevel{Measured, Calibrated, Estimated, Placeholder}
	for _, level := range cases {
		badge := NewConfidenceBadge(level)
		if badge == nil {
			t.Fatalf("badge nil for %s", level)
		}
		if len(badge.Objects) != 2 {
			t.Fatalf("expected 2 objects for %s, got %d", level, len(badge.Objects))
		}
		label, ok := badge.Objects[1].(*widget.Label)
		if !ok {
			t.Fatalf("expected label object for %s", level)
		}
		if label.Text != string(level) {
			t.Fatalf("expected label %q got %q", level, label.Text)
		}
	}
}

func TestNewConfidenceBadge_DefaultsPlaceholder(t *testing.T) {
	test.NewApp()
	badge := NewConfidenceBadge("")
	label := badge.Objects[1].(*widget.Label)
	if label.Text != string(Placeholder) {
		t.Fatalf("expected %q got %q", Placeholder, label.Text)
	}
	_ = container.NewVBox(badge) // ensure usable in containers without panic
}
