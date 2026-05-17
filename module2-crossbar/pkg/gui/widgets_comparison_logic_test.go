//go:build legacy_fyne

package gui

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2/widget"
)

func TestBeforeAfterToggle_AdditionalLogic(t *testing.T) {
	b := NewBeforeAfterToggle(2, 2)
	b.idealData = [][]float64{{0.0, 0.5}, {1.0, 0.5}}
	b.actualData = [][]float64{{0.1, 0.4}, {0.8, 0.7}}

	d := b.computeDifference()
	if (d[0][0] < 0.099 || d[0][0] > 0.101) || (d[1][1] < 0.199 || d[1][1] > 0.201) {
		t.Fatalf("unexpected absolute diff map: %#v", d)
	}

	signed := b.computeSignedDifference()
	if signed[0][0] <= 0 || signed[0][1] >= 0 {
		t.Fatalf("expected signed normalization, got %#v", signed)
	}

	b.statsLabel = widget.NewLabel("")
	b.updateStatsLabel()
	if !strings.Contains(b.statsLabel.Text, "RMSE") {
		t.Fatalf("expected stats text, got %q", b.statsLabel.Text)
	}

	b.legendLabel = widget.NewLabel("legend")
	b.mode = "diff"
	b.updateLegend()
	if !b.legendLabel.Visible() || !strings.Contains(b.legendLabel.Text, "Blue") {
		t.Fatalf("expected diff legend visible, got visible=%v text=%q", b.legendLabel.Visible(), b.legendLabel.Text)
	}
	b.mode = "split"
	b.updateLegend()
	if b.legendLabel.Visible() {
		t.Fatal("expected legend hidden for non-diff mode")
	}
}
