package neural

import (
	"testing"
)

// TestSweepQuantizationLevels_Basic verifies the function runs with synthetic data.
func TestSweepQuantizationLevels_Basic(t *testing.T) {
	net := NewDualModeNetwork(4, 4, 2)

	images := make([][]float64, 10)
	labels := make([]int, 10)
	for i := range images {
		images[i] = []float64{1, 0, 0, 0}
		labels[i] = 0
	}

	levels := []int{2, 4, 8, 16, 30}
	results, err := SweepQuantizationLevels(net, images, labels, levels)
	if err != nil {
		t.Fatalf("SweepQuantizationLevels failed: %v", err)
	}
	if len(results) != len(levels) {
		t.Fatalf("expected %d results, got %d", len(levels), len(results))
	}
	for i, r := range results {
		if r.NumLevels != levels[i] {
			t.Errorf("result[%d].NumLevels=%d, want %d", i, r.NumLevels, levels[i])
		}
		if r.Total != len(images) {
			t.Errorf("result[%d].Total=%d, want %d", i, r.Total, len(images))
		}
		if r.Accuracy < 0 || r.Accuracy > 1 {
			t.Errorf("result[%d].Accuracy=%g out of range [0,1]", i, r.Accuracy)
		}
	}
}

func TestSweepADCBits_Basic(t *testing.T) {
	net := NewDualModeNetwork(4, 4, 2)
	images := make([][]float64, 10)
	labels := make([]int, 10)
	for i := range images {
		images[i] = []float64{1, 0, 0, 0}
		labels[i] = 0
	}

	bits := []int{3, 4, 6, 8}
	results, err := SweepADCBits(net, images, labels, bits)
	if err != nil {
		t.Fatalf("SweepADCBits failed: %v", err)
	}
	if len(results) != len(bits) {
		t.Fatalf("expected %d results, got %d", len(bits), len(results))
	}
	for i, r := range results {
		if r.ADCBits != bits[i] {
			t.Errorf("result[%d].ADCBits=%d, want %d", i, r.ADCBits, bits[i])
		}
		if r.Total != len(images) {
			t.Errorf("result[%d].Total=%d, want %d", i, r.Total, len(images))
		}
		if r.Accuracy < 0 || r.Accuracy > 1 {
			t.Errorf("result[%d].Accuracy=%g out of range [0,1]", i, r.Accuracy)
		}
	}
}

func TestSweepNoiseLevel_Basic(t *testing.T) {
	net := NewDualModeNetwork(4, 4, 2)
	images := make([][]float64, 10)
	labels := make([]int, 10)
	for i := range images {
		images[i] = []float64{1, 0, 0, 0}
		labels[i] = 0
	}

	noises := []float64{0.0, 0.05, 0.10, 0.20}
	results, err := SweepNoiseLevel(net, images, labels, noises)
	if err != nil {
		t.Fatalf("SweepNoiseLevel failed: %v", err)
	}
	if len(results) != len(noises) {
		t.Fatalf("expected %d results, got %d", len(noises), len(results))
	}
	for i, r := range results {
		if r.NoiseLevel != noises[i] {
			t.Errorf("result[%d].NoiseLevel=%g, want %g", i, r.NoiseLevel, noises[i])
		}
		if r.Total != len(images) {
			t.Errorf("result[%d].Total=%d, want %d", i, r.Total, len(images))
		}
		if r.Accuracy < 0 || r.Accuracy > 1 {
			t.Errorf("result[%d].Accuracy=%g out of range [0,1]", i, r.Accuracy)
		}
	}
}

func TestSweepQuantizationLevels_InvalidInputs(t *testing.T) {
	net := NewDualModeNetwork(4, 4, 2)
	images := [][]float64{{1, 0, 0, 0}}
	labels := []int{0}

	// Empty levels
	_, err := SweepQuantizationLevels(net, images, labels, []int{})
	if err == nil {
		t.Error("expected error for empty levels")
	}

	// Level < 2
	_, err = SweepQuantizationLevels(net, images, labels, []int{1})
	if err == nil {
		t.Error("expected error for level < 2")
	}

	// Mismatched images/labels
	_, err = SweepQuantizationLevels(net, images, []int{0, 1}, []int{2})
	if err == nil {
		t.Error("expected error for images/labels mismatch")
	}
}

func TestSweepADCBits_InvalidInputs(t *testing.T) {
	net := NewDualModeNetwork(4, 4, 2)
	images := [][]float64{{1, 0, 0, 0}}
	labels := []int{0}

	// Empty bitWidths
	_, err := SweepADCBits(net, images, labels, []int{})
	if err == nil {
		t.Error("expected error for empty bitWidths")
	}

	// Bits < 3
	_, err = SweepADCBits(net, images, labels, []int{2})
	if err == nil {
		t.Error("expected error for bits < 3")
	}

	// Bits > 16
	_, err = SweepADCBits(net, images, labels, []int{17})
	if err == nil {
		t.Error("expected error for bits > 16")
	}
}

func TestSweepNoiseLevel_InvalidInputs(t *testing.T) {
	net := NewDualModeNetwork(4, 4, 2)
	images := [][]float64{{1, 0, 0, 0}}
	labels := []int{0}

	// Empty noiseLevels
	_, err := SweepNoiseLevel(net, images, labels, []float64{})
	if err == nil {
		t.Error("expected error for empty noiseLevels")
	}

	// Noise < 0
	_, err = SweepNoiseLevel(net, images, labels, []float64{-0.01})
	if err == nil {
		t.Error("expected error for noise < 0")
	}

	// Noise > 0.20
	_, err = SweepNoiseLevel(net, images, labels, []float64{0.21})
	if err == nil {
		t.Error("expected error for noise > 0.20")
	}
}

func TestSweepQuantizationLevels_RestoresConfig(t *testing.T) {
	net := NewDualModeNetwork(4, 4, 2)
	net.SetNumLevels(30)
	images := [][]float64{{1, 0, 0, 0}}
	labels := []int{0}

	_, _ = SweepQuantizationLevels(net, images, labels, []int{2, 8})

	if got := net.GetNumLevels(); got != 30 {
		t.Errorf("NumLevels not restored: got %d, want 30", got)
	}
}

func TestSweepADCBits_RestoresConfig(t *testing.T) {
	net := NewDualModeNetwork(4, 4, 2)
	net.SetADCBits(8)
	images := [][]float64{{1, 0, 0, 0}}
	labels := []int{0}

	_, _ = SweepADCBits(net, images, labels, []int{3, 6})

	net.mu.RLock()
	got := net.Config.ADCBits
	net.mu.RUnlock()
	if got != 8 {
		t.Errorf("ADCBits not restored: got %d, want 8", got)
	}
}

func TestSweepNoiseLevel_RestoresConfig(t *testing.T) {
	net := NewDualModeNetwork(4, 4, 2)
	net.SetNoiseLevel(0.05)
	images := [][]float64{{1, 0, 0, 0}}
	labels := []int{0}

	_, _ = SweepNoiseLevel(net, images, labels, []float64{0.0, 0.10})

	net.mu.RLock()
	got := net.Config.NoiseLevel
	net.mu.RUnlock()
	if got != 0.05 {
		t.Errorf("NoiseLevel not restored: got %g, want 0.05", got)
	}
}
