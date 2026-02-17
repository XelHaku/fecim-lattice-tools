package edahello

import "testing"

func TestCompileMapsWeightsToCells(t *testing.T) {
	w := [][]float64{{-1, 0}, {0.5, 1}}
	cells := Compile(w)
	if len(cells) != 4 {
		t.Fatalf("expected 4 cells, got %d", len(cells))
	}
	for _, c := range cells {
		if c.Level < 0 || c.Level > 29 {
			t.Fatalf("level out of range: %d", c.Level)
		}
		if c.Conductance < 10.0 || c.Conductance > 100.0 {
			t.Fatalf("conductance out of range: %f", c.Conductance)
		}
	}
}
