package mnisttrain

import "testing"

func TestQuantizeBounds(t *testing.T) {
	q := quantize(2.0, -1.0, 1.0, 5)
	if q > 1.0 {
		t.Fatalf("quantize should clamp to max, got %v", q)
	}
	q = quantize(-2.0, -1.0, 1.0, 5)
	if q < -1.0 {
		t.Fatalf("quantize should clamp to min, got %v", q)
	}
}

func TestArgmax(t *testing.T) {
	idx := argmax([]float64{0.1, 0.5, 0.3})
	if idx != 1 {
		t.Fatalf("argmax=%d want 1", idx)
	}
}
