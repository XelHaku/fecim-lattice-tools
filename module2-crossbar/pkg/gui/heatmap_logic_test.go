//go:build legacy_fyne

package gui

import (
	"testing"

	"fecim-lattice-tools/shared/mathutil"
)

func TestHeatmap_ColorAndImagePaths(t *testing.T) {
	h := NewCrossbarHeatmap(3, 3)
	h.data = [][]float64{{0, 0.5, 1}, {0.2, 0.4, 0.8}, {1, 0.1, 0.6}}
	h.minVal, h.maxVal = 0, 1
	h.showGridLines = true
	h.showSelection = true
	h.selectedRow, h.selectedCol = 1, 1

	if img := h.generateImage(80, 80); img.Bounds().Dx() != 80 {
		t.Fatalf("unexpected image width: %d", img.Bounds().Dx())
	}

	for _, cm := range []string{"viridis", "plasma", "coolwarm", "fecim", "diverging", "unknown"} {
		h.colormap = cm
		c := h.valueToColor(0.25)
		if c.A != 255 {
			t.Fatalf("colormap %s returned non-opaque color: %+v", cm, c)
		}
		_ = h.selectionColor()
	}

	if mathutil.Clamp(-1, 0, 1) != 0 || mathutil.Clamp(2, 0, 1) != 1 || mathutil.Clamp(0.5, 0, 1) != 0.5 {
		t.Fatal("clamp bounds failed")
	}

	if c := divergingColor(-1); c.B != 255 {
		t.Fatalf("expected blue extreme for -1, got %+v", c)
	}
	if c := divergingColor(1); c.R != 255 || c.G != 0 {
		t.Fatalf("expected red extreme for +1, got %+v", c)
	}
}

func TestHeatmap_DimensionsAndAltText(t *testing.T) {
	h := NewCrossbarHeatmap(2, 2)
	h.data = [][]float64{{0.1, 0.2}, {0.3, 0.4}}
	h.selectedRow, h.selectedCol, h.showSelection = 1, 0, true
	alt := h.TextAlternative()
	if alt == "" {
		t.Fatal("expected non-empty text alternative")
	}

	h.SetDimensions(20, 2)
	if !h.showGridLines {
		t.Fatal("expected grid lines auto-enabled for large dimensions")
	}
	if len(h.data) != 20 || len(h.data[0]) != 2 {
		t.Fatalf("unexpected resized data dimensions: %dx%d", len(h.data), len(h.data[0]))
	}
}
