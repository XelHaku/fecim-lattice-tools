package overlay

import (
	"image/color"
	"testing"
)

func TestDimmedCellColor(t *testing.T) {
	base := color.RGBA{R: 200, G: 100, B: 40, A: 7}
	got := DimmedCellColor(base)
	if got.A != 255 {
		t.Fatalf("alpha=%d want 255", got.A)
	}
	if got.R != 150 || got.G != 75 || got.B != 30 {
		t.Fatalf("rgb=(%d,%d,%d) want (150,75,30)", got.R, got.G, got.B)
	}
}
