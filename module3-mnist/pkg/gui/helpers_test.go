package gui

import (
	"image/color"
	"testing"

	"fecim-lattice-tools/shared/mathutil"
	"fecim-lattice-tools/shared/physics"
)

// TestFormatEnergy verifies energy formatting to human-readable units.
// Note: The actual formatting logic is now in shared/physics package,
// this test verifies the integration works correctly.
func TestFormatEnergy(t *testing.T) {
	tests := []struct {
		joules   float64
		expected string
	}{
		{1e-15, "1.00 fJ"},   // fJ range (shared/physics adds fJ support)
		{1e-13, "100.00 fJ"}, // fJ range (1e-13 J = 100 fJ)
		{5e-10, "500.00 pJ"}, // pJ range
		{1e-9, "1.00 nJ"},
		{5e-7, "500.00 nJ"}, // nJ range
		{1e-6, "1.00 µJ"},
		{5e-4, "500.00 µJ"}, // µJ range
		{1e-3, "1.00 mJ"},
		{0.5, "500.00 mJ"},
		{1.0, "1.00 J"},
		{5.0, "5.00 J"},
	}

	for _, tc := range tests {
		result := physics.FormatEnergy(tc.joules)
		if result != tc.expected {
			t.Errorf("physics.FormatEnergy(%e) = %q, expected %q", tc.joules, result, tc.expected)
		}
	}
}

// TestWeightToColor verifies color mapping for weight visualization.
func TestWeightToColor(t *testing.T) {
	tests := []struct {
		normalized float64
		desc       string
		checkFn    func(color.Color) bool
	}{
		{0.0, "minimum (blue)", func(c color.Color) bool {
			r, g, b, _ := c.RGBA()
			return r>>8 == 0 && g>>8 == 0 && b>>8 == 255
		}},
		{0.5, "middle (white)", func(c color.Color) bool {
			r, g, b, _ := c.RGBA()
			return r>>8 == 255 && g>>8 == 255 && b>>8 == 255
		}},
		{1.0, "maximum (red)", func(c color.Color) bool {
			r, g, b, _ := c.RGBA()
			return r>>8 == 255 && g>>8 == 0 && b>>8 == 0
		}},
	}

	for _, tc := range tests {
		result := weightToColor(tc.normalized)
		if !tc.checkFn(result) {
			r, g, b, _ := result.RGBA()
			t.Errorf("weightToColor(%.1f) %s: got RGB(%d,%d,%d)", tc.normalized, tc.desc, r>>8, g>>8, b>>8)
		}
	}
}

// TestWeightToColorGradient verifies smooth color transition.
func TestWeightToColorGradient(t *testing.T) {
	// Test that colors transition smoothly
	prevR, prevG, prevB := uint8(0), uint8(0), uint8(255)

	for i := 0; i <= 10; i++ {
		normalized := float64(i) / 10.0
		c := weightToColor(normalized)
		r, g, b, _ := c.RGBA()
		currR, currG, currB := uint8(r>>8), uint8(g>>8), uint8(b>>8)

		// Check for reasonable transitions (no sudden jumps > 100 except at boundaries)
		if i > 0 && i < 10 {
			diffR := int(currR) - int(prevR)
			diffG := int(currG) - int(prevG)
			diffB := int(currB) - int(prevB)

			if abs(diffR) > 100 || abs(diffG) > 100 || abs(diffB) > 100 {
				t.Errorf("Abrupt color change at normalized=%.1f: prev=(%d,%d,%d) curr=(%d,%d,%d)",
					normalized, prevR, prevG, prevB, currR, currG, currB)
			}
		}

		prevR, prevG, prevB = currR, currG, currB
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// TestClamp verifies value clamping.
func TestClamp(t *testing.T) {
	tests := []struct {
		v, min, max float64
		expected    float64
	}{
		{5, 0, 10, 5},    // In range
		{-1, 0, 10, 0},   // Below min
		{15, 0, 10, 10},  // Above max
		{0, 0, 10, 0},    // At min
		{10, 0, 10, 10},  // At max
		{0.5, 0, 1, 0.5}, // Float in range
		{-0.1, 0, 1, 0},  // Float below min
		{1.5, 0, 1, 1},   // Float above max
	}

	for _, tc := range tests {
		result := mathutil.Clamp(tc.v, tc.min, tc.max)
		if result != tc.expected {
			t.Errorf("mathutil.Clamp(%v, %v, %v) = %v, expected %v", tc.v, tc.min, tc.max, result, tc.expected)
		}
	}
}

// TestGenerateSyntheticData verifies synthetic data generation.
func TestGenerateSyntheticData(t *testing.T) {
	count := 100
	images, labels := generateSyntheticData(count)

	// Check counts
	if len(images) != count {
		t.Errorf("generateSyntheticData returned %d images, expected %d", len(images), count)
	}
	if len(labels) != count {
		t.Errorf("generateSyntheticData returned %d labels, expected %d", len(labels), count)
	}

	// Check image dimensions
	for i, img := range images {
		if len(img) != 784 {
			t.Errorf("Image %d has %d pixels, expected 784", i, len(img))
		}
	}

	// Check label range
	for i, label := range labels {
		if label < 0 || label > 9 {
			t.Errorf("Label %d is %d, expected 0-9", i, label)
		}
	}

	// Check pixel values are normalized
	for i, img := range images {
		for j, pixel := range img {
			if pixel < 0 || pixel > 1 {
				t.Errorf("Pixel [%d][%d] = %v, expected 0-1", i, j, pixel)
			}
		}
	}

	// Check that different labels produce different patterns
	// Each digit should have a distinct pattern
	labelSums := make(map[int]float64)
	for i, img := range images {
		sum := 0.0
		for _, pixel := range img {
			sum += pixel
		}
		labelSums[labels[i]] = sum
	}

	// With 100 images and 10 labels, we should have multiple labels represented
	if len(labelSums) < 5 {
		t.Errorf("Too few label types: got %d, expected at least 5", len(labelSums))
	}
}

// TestMinHelper verifies the min helper function.
func TestMinHelper(t *testing.T) {
	tests := []struct {
		a, b     int
		expected int
	}{
		{5, 10, 5},
		{10, 5, 5},
		{5, 5, 5},
		{-1, 1, -1},
		{0, 0, 0},
	}

	for _, tc := range tests {
		result := min(tc.a, tc.b)
		if result != tc.expected {
			t.Errorf("min(%d, %d) = %d, expected %d", tc.a, tc.b, result, tc.expected)
		}
	}
}
