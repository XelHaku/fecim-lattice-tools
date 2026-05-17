//go:build legacy_fyne

package gui

import (
	"math"
	"testing"
)

func TestPreprocessDigitCenters(t *testing.T) {
	pixels := make([]float64, preprocessSize*preprocessSize)
	for y := 2; y < preprocessSize-2; y++ {
		pixels[y*preprocessSize+2] = 1.0
	}

	out := PreprocessDigit(pixels)
	if len(out) != preprocessSize*preprocessSize {
		t.Fatalf("expected %d pixels, got %d", preprocessSize*preprocessSize, len(out))
	}

	comX, comY := centerOfMass(out)
	if math.Abs(comX-float64(preprocessSize/2)) > 1.5 || math.Abs(comY-float64(preprocessSize/2)) > 1.5 {
		t.Errorf("center of mass not centered: com=(%.2f, %.2f)", comX, comY)
	}
}

func TestPreprocessDigitBlank(t *testing.T) {
	pixels := make([]float64, preprocessSize*preprocessSize)
	out := PreprocessDigit(pixels)
	maxVal := 0.0
	for _, v := range out {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal != 0 {
		t.Fatalf("expected blank output, got max value %.3f", maxVal)
	}
}

func centerOfMass(pixels []float64) (float64, float64) {
	sum, sumX, sumY := 0.0, 0.0, 0.0
	for y := 0; y < preprocessSize; y++ {
		for x := 0; x < preprocessSize; x++ {
			v := pixels[y*preprocessSize+x]
			sum += v
			sumX += v * float64(x)
			sumY += v * float64(y)
		}
	}
	if sum == 0 {
		return float64(preprocessSize) / 2, float64(preprocessSize) / 2
	}
	return sumX / sum, sumY / sum
}
