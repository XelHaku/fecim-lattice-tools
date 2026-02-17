package display

import "testing"

func TestShouldShowOverlay(t *testing.T) {
	tests := []struct {
		name           string
		overlayEnabled bool
		isSelected     bool
		cellSize       int
		wantOverlay    bool
		wantSelected   bool
	}{
		{"disabled", false, true, 50, false, false},
		{"large_all_cells", true, false, 45, true, false},
		{"medium_selected", true, true, 40, true, true},
		{"medium_unselected", true, false, 40, false, false},
		{"small_hidden", true, true, 20, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldShowOverlay(tt.overlayEnabled, tt.isSelected, tt.cellSize); got != tt.wantOverlay {
				t.Fatalf("ShouldShowOverlay()=%v want %v", got, tt.wantOverlay)
			}
			if got := ShouldShowSelectedOnly(tt.overlayEnabled, tt.isSelected, tt.cellSize); got != tt.wantSelected {
				t.Fatalf("ShouldShowSelectedOnly()=%v want %v", got, tt.wantSelected)
			}
		})
	}
}
