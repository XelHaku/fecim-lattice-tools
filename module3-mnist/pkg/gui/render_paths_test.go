package gui

import "testing"

func TestWeightComparison_RenderPaths(t *testing.T) {
	w := &WeightComparisonWidget{}
	w.fpWeights = [][]float64{{-1, 0}, {0.5, 1}}
	w.quantWeights = [][]float64{{-0.8, -0.1}, {0.4, 0.6}}
	w.wMin, w.wMax, w.maxError = -1, 1, 0.4

	for _, mode := range []int{0, 1, 2} {
		w.showMode = mode
		img := w.generateImage(120, 80)
		if img.Bounds().Dx() != 120 {
			t.Fatalf("unexpected image size in mode %d: %v", mode, img.Bounds())
		}
	}

	d := &DualWeightHeatmap{layerIndex: 0}
	d.fpW1 = [][]float64{{-1, 0}, {0.2, 0.8}}
	d.quantW1 = [][]float64{{-0.9, -0.1}, {0.3, 0.7}}
	img := d.generateImage(200, 80)
	if img.Bounds().Dx() != 200 {
		t.Fatalf("unexpected dual heatmap size: %v", img.Bounds())
	}

	d.layerIndex = 1
	d.fpW2 = [][]float64{{-0.5, 0.5}, {0.1, 0.9}}
	d.quantW2 = [][]float64{{-0.4, 0.4}, {0.2, 0.8}}
	_ = d.generateImage(200, 80)
}
