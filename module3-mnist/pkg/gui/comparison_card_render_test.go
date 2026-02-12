package gui

import "testing"

func TestComparisonCard_GenerateImagePaths(t *testing.T) {
	cc := &ComparisonCard{}
	img := cc.generateImage(300, 220)
	if img.Bounds().Dx() != 300 {
		t.Fatalf("unexpected idle image size: %v", img.Bounds())
	}

	cc.result = &ComparisonResult{
		FPPrediction:    1,
		FPConfidence:    0.6,
		FPProbabilities: []float64{0.1, 0.6, 0.3, 0, 0, 0, 0, 0, 0, 0},
		CIMPrediction:    2,
		CIMConfidence:    0.7,
		CIMProbabilities: []float64{0.05, 0.2, 0.7, 0.05, 0, 0, 0, 0, 0, 0},
		Match:            false,
	}
	img = cc.generateImage(560, 440)
	if img.Bounds().Dx() != 560 || img.Bounds().Dy() != 440 {
		t.Fatalf("unexpected result image size: %v", img.Bounds())
	}
}
