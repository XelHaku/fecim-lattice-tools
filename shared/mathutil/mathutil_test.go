package mathutil

import "testing"

func TestClamp(t *testing.T) {
	tests := []struct {
		v, min, max, want float64
	}{
		{5, 0, 10, 5},
		{-1, 0, 10, 0},
		{15, 0, 10, 10},
		{0, 0, 10, 0},
		{10, 0, 10, 10},
	}
	for _, tt := range tests {
		got := Clamp(tt.v, tt.min, tt.max)
		if got != tt.want {
			t.Errorf("Clamp(%v, %v, %v) = %v, want %v", tt.v, tt.min, tt.max, got, tt.want)
		}
	}
}

func TestClamp01(t *testing.T) {
	tests := []struct {
		v, want float64
	}{
		{0.5, 0.5},
		{-0.1, 0},
		{1.5, 1},
		{0, 0},
		{1, 1},
	}
	for _, tt := range tests {
		got := Clamp01(tt.v)
		if got != tt.want {
			t.Errorf("Clamp01(%v) = %v, want %v", tt.v, got, tt.want)
		}
	}
}

func TestClampInt(t *testing.T) {
	tests := []struct {
		v, min, max, want int
	}{
		{5, 0, 10, 5},
		{-1, 0, 10, 0},
		{15, 0, 10, 10},
	}
	for _, tt := range tests {
		got := ClampInt(tt.v, tt.min, tt.max)
		if got != tt.want {
			t.Errorf("ClampInt(%v, %v, %v) = %v, want %v", tt.v, tt.min, tt.max, got, tt.want)
		}
	}
}

func TestClampPositive(t *testing.T) {
	tests := []struct {
		v, want float64
	}{
		{5, 5},
		{-1, 0},
		{0, 0},
	}
	for _, tt := range tests {
		got := ClampPositive(tt.v)
		if got != tt.want {
			t.Errorf("ClampPositive(%v) = %v, want %v", tt.v, got, tt.want)
		}
	}
}

func TestLerpByIndex(t *testing.T) {
	tests := []struct {
		name     string
		index    int
		count    int
		min, max float64
		want     float64
	}{
		{name: "first", index: 0, count: 4, min: 0.2, max: 0.5, want: 0.2},
		{name: "middle", index: 1, count: 4, min: 0.2, max: 0.5, want: 0.3},
		{name: "last", index: 3, count: 4, min: 0.2, max: 0.5, want: 0.5},
		{name: "single", index: 0, count: 1, min: 0.2, max: 0.5, want: 0.2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LerpByIndex(tt.index, tt.count, tt.min, tt.max)
			if got != tt.want {
				t.Fatalf("LerpByIndex = %.12f, want %.12f", got, tt.want)
			}
		})
	}
}

func TestDot(t *testing.T) {
	got, err := Dot([]float64{1, 2, 3}, []float64{0.5, 1, 2})
	if err != nil {
		t.Fatalf("Dot returned error: %v", err)
	}
	if got != 8.5 {
		t.Fatalf("Dot = %.3f, want 8.5", got)
	}

	if _, err := Dot([]float64{1}, []float64{1, 2}); err == nil {
		t.Fatal("expected length mismatch error")
	}
}

func TestMatrixVectorMultiply(t *testing.T) {
	got, err := MatrixVectorMultiply([][]float64{{1, 2, 3}, {4, 5, 6}}, []float64{0.5, 1, 2})
	if err != nil {
		t.Fatalf("MatrixVectorMultiply returned error: %v", err)
	}
	want := []float64{8.5, 19}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("output[%d] = %.3f, want %.3f", i, got[i], want[i])
		}
	}

	if _, err := MatrixVectorMultiply([][]float64{{1}, {1, 2}}, []float64{1}); err == nil {
		t.Fatal("expected ragged matrix error")
	}
}

func TestAbsInt(t *testing.T) {
	tests := []struct {
		x, want int
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
		{-1, 1},
		{1, 1},
	}
	for _, tt := range tests {
		got := AbsInt(tt.x)
		if got != tt.want {
			t.Errorf("AbsInt(%v) = %v, want %v", tt.x, got, tt.want)
		}
	}
}
