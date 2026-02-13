package widgets

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestNewUncertaintyOverlay_RendersExpectedText(t *testing.T) {
	test.NewApp()
	label := NewUncertaintyOverlay(12.5, 0.3, 95)
	want := "12.5 ± 0.3 (95.0%)"
	if label.Text != want {
		t.Fatalf("unexpected overlay text: got %q want %q", label.Text, want)
	}
}
