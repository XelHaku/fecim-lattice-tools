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
