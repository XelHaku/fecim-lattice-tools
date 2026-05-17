//go:build legacy_fyne

package ispp

import "testing"

func TestClampTargetLevel(t *testing.T) {
	tests := []struct {
		target int
		levels int
		want   int
	}{
		{target: -1, levels: 32, want: 0},
		{target: 0, levels: 32, want: 0},
		{target: 31, levels: 32, want: 31},
		{target: 32, levels: 32, want: 31},
		{target: 5, levels: 0, want: 0},
	}
	for _, tt := range tests {
		if got := ClampTargetLevel(tt.target, tt.levels); got != tt.want {
			t.Fatalf("ClampTargetLevel(%d,%d)=%d want %d", tt.target, tt.levels, got, tt.want)
		}
	}
}
