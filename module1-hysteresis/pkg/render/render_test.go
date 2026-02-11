// Package render provides tests for rendering utilities.
package render

import (
	"math"
	"testing"
)

// TestPolarizationToColor verifies color mapping for different polarization values.
func TestPolarizationToColor(t *testing.T) {
	cm := DefaultColorMap()

	tests := []struct {
		name     string
		normP    float64
		expected Color
	}{
		{
			name:     "positive max (+1)",
			normP:    1.0,
			expected: cm.Positive,
		},
		{
			name:     "negative max (-1)",
			normP:    -1.0,
			expected: cm.Negative,
		},
		{
			name:     "zero",
			normP:    0.0,
			expected: cm.Zero,
		},
		{
			name:  "positive mid (+0.5)",
			normP: 0.5,
			expected: Color{
				R: cm.Zero.R + 0.5*(cm.Positive.R-cm.Zero.R),
				G: cm.Zero.G + 0.5*(cm.Positive.G-cm.Zero.G),
				B: cm.Zero.B + 0.5*(cm.Positive.B-cm.Zero.B),
				A: 1.0,
			},
		},
		{
			name:  "negative mid (-0.5)",
			normP: -0.5,
			expected: Color{
				R: cm.Zero.R + 0.5*(cm.Negative.R-cm.Zero.R),
				G: cm.Zero.G + 0.5*(cm.Negative.G-cm.Zero.G),
				B: cm.Zero.B + 0.5*(cm.Negative.B-cm.Zero.B),
				A: 1.0,
			},
		},
		{
			name:     "above range (+2.0) clamped",
			normP:    2.0,
			expected: cm.Positive,
		},
		{
			name:     "below range (-2.0) clamped",
			normP:    -2.0,
			expected: cm.Negative,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cm.PolarizationToColor(tt.normP)

			// Compare with small epsilon for floating point
			epsilon := float32(1e-6)
			if math.Abs(float64(result.R-tt.expected.R)) > float64(epsilon) ||
				math.Abs(float64(result.G-tt.expected.G)) > float64(epsilon) ||
				math.Abs(float64(result.B-tt.expected.B)) > float64(epsilon) ||
				math.Abs(float64(result.A-tt.expected.A)) > float64(epsilon) {
				t.Errorf("PolarizationToColor(%v) = %+v, want %+v", tt.normP, result, tt.expected)
			}
		})
	}
}

// TestNormalizeToScreen verifies coordinate transformation.
func TestNormalizeToScreen(t *testing.T) {
	hp := NewHysteresisPlot(2.0, 50.0)

	tests := []struct {
		name     string
		E        float64
		P        float64
		expectedX float64
		expectedY float64
	}{
		{
			name:      "center (0,0)",
			E:         0.0,
			P:         0.0,
			expectedX: 0.5,
			expectedY: 0.5,
		},
		{
			name:      "max E, max P",
			E:         2.0,
			P:         50.0,
			expectedX: 1.0,
			expectedY: 1.0,
		},
		{
			name:      "min E, min P",
			E:         -2.0,
			P:         -50.0,
			expectedX: 0.0,
			expectedY: 0.0,
		},
		{
			name:      "mid positive",
			E:         1.0,
			P:         25.0,
			expectedX: 0.75,
			expectedY: 0.75,
		},
		{
			name:      "mid negative",
			E:         -1.0,
			P:         -25.0,
			expectedX: 0.25,
			expectedY: 0.25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y := hp.NormalizeToScreen(tt.E, tt.P)
			epsilon := 1e-10

			if math.Abs(x-tt.expectedX) > epsilon || math.Abs(y-tt.expectedY) > epsilon {
				t.Errorf("NormalizeToScreen(%v, %v) = (%v, %v), want (%v, %v)",
					tt.E, tt.P, x, y, tt.expectedX, tt.expectedY)
			}
		})
	}
}

// TestScreenToNDC verifies screen coordinate to NDC transformation.
func TestScreenToNDC(t *testing.T) {
	width, height := 1280, 720

	tests := []struct {
		name      string
		x, y      float64
		expectedX float64
		expectedY float64
	}{
		{
			name:      "center",
			x:         640,
			y:         360,
			expectedX: 0.0,
			expectedY: 0.0,
		},
		{
			name:      "top-left",
			x:         0,
			y:         0,
			expectedX: -1.0,
			expectedY: 1.0,
		},
		{
			name:      "bottom-right",
			x:         1280,
			y:         720,
			expectedX: 1.0,
			expectedY: -1.0,
		},
		{
			name:      "top-right",
			x:         1280,
			y:         0,
			expectedX: 1.0,
			expectedY: 1.0,
		},
		{
			name:      "bottom-left",
			x:         0,
			y:         720,
			expectedX: -1.0,
			expectedY: -1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ndcX, ndcY := ScreenToNDC(tt.x, tt.y, width, height)
			epsilon := 1e-10

			if math.Abs(ndcX-tt.expectedX) > epsilon || math.Abs(ndcY-tt.expectedY) > epsilon {
				t.Errorf("ScreenToNDC(%v, %v) = (%v, %v), want (%v, %v)",
					tt.x, tt.y, ndcX, ndcY, tt.expectedX, tt.expectedY)
			}
		})
	}
}

// TestScreenToNDCRoundtrip verifies coordinate transformation roundtrips.
func TestScreenToNDCRoundtrip(t *testing.T) {
	width, height := 1280, 720

	tests := []struct {
		name string
		x, y float64
	}{
		{"center", 640, 360},
		{"top-left", 0, 0},
		{"bottom-right", 1280, 720},
		{"arbitrary", 512, 288},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ndcX, ndcY := ScreenToNDC(tt.x, tt.y, width, height)
			x, y := NDCToScreen(ndcX, ndcY, width, height)

			epsilon := 1e-10
			if math.Abs(x-tt.x) > epsilon || math.Abs(y-tt.y) > epsilon {
				t.Errorf("Roundtrip failed: (%v, %v) -> NDC -> (%v, %v)", tt.x, tt.y, x, y)
			}
		})
	}
}

// TestQuantizeTo30Levels verifies discretization logic.
func TestQuantizeTo30Levels(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected float64
	}{
		{
			name:     "zero",
			value:    0.0,
			expected: 0.0,
		},
		{
			name:     "one",
			value:    1.0,
			expected: 1.0,
		},
		{
			name:     "mid (0.5)",
			value:    0.5,
			expected: 15.0 / 29.0, // Round to level 15
		},
		{
			name:     "quarter (0.25)",
			value:    0.25,
			expected: 7.0 / 29.0, // Round to level 7
		},
		{
			name:     "three-quarters (0.75)",
			value:    0.75,
			expected: 22.0 / 29.0, // Round to level 22
		},
		{
			name:     "below range (-0.5) clamped to 0",
			value:    -0.5,
			expected: 0.0,
		},
		{
			name:     "above range (1.5) clamped to 1",
			value:    1.5,
			expected: 1.0,
		},
		{
			name:     "near zero (0.01)",
			value:    0.01,
			expected: 0.0, // Round to level 0
		},
		{
			name:     "near one (0.99)",
			value:    0.99,
			expected: 1.0, // Round to level 29
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := QuantizeTo30Levels(tt.value)
			epsilon := 1e-10

			if math.Abs(result-tt.expected) > epsilon {
				t.Errorf("QuantizeTo30Levels(%v) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

// TestGetLevel verifies discrete level calculation.
func TestGetLevel(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected int
	}{
		{"zero", 0.0, 0},
		{"one", 1.0, 29},
		{"mid (0.5)", 0.5, 15},
		{"quarter (0.25)", 0.25, 7},
		{"three-quarters (0.75)", 0.75, 22},
		{"below range (-0.5) clamped", -0.5, 0},
		{"above range (1.5) clamped", 1.5, 29},
		{"near zero (0.01)", 0.01, 0},
		{"near one (0.99)", 0.99, 29},
		{"exact level boundary (7/29)", 7.0 / 29.0, 7},
		{"exact level boundary (15/29)", 15.0 / 29.0, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLevel(tt.value)

			if result != tt.expected {
				t.Errorf("GetLevel(%v) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

// TestAddPoint verifies point accumulation.
func TestAddPoint(t *testing.T) {
	hp := NewHysteresisPlot(2.0, 50.0)

	// Initially empty
	if len(hp.Points) != 0 {
		t.Errorf("New plot should have 0 points, got %d", len(hp.Points))
	}

	// Add a point
	hp.AddPoint(1.0, 25.0)
	if len(hp.Points) != 1 {
		t.Errorf("After adding 1 point, should have 1 point, got %d", len(hp.Points))
	}
	if hp.CurrentE != 1.0 || hp.CurrentP != 25.0 {
		t.Errorf("Current position should be (1.0, 25.0), got (%v, %v)", hp.CurrentE, hp.CurrentP)
	}

	// Add more points
	hp.AddPoint(1.5, 40.0)
	hp.AddPoint(2.0, 50.0)
	if len(hp.Points) != 3 {
		t.Errorf("After adding 3 points, should have 3 points, got %d", len(hp.Points))
	}

	// Verify last point
	lastPoint := hp.Points[len(hp.Points)-1]
	if lastPoint.X != 2.0 || lastPoint.Y != 50.0 {
		t.Errorf("Last point should be (2.0, 50.0), got (%v, %v)", lastPoint.X, lastPoint.Y)
	}
}

// TestAddPointBufferLimit verifies circular buffer behavior.
func TestAddPointBufferLimit(t *testing.T) {
	hp := NewHysteresisPlot(2.0, 50.0)

	// Add points up to the limit (10000)
	for i := 0; i < 10001; i++ {
		hp.AddPoint(float64(i), float64(i))
	}

	// Should have trimmed to 9000 (10000 - 1000)
	if len(hp.Points) != 9001 {
		t.Errorf("After exceeding buffer limit, should have 9001 points, got %d", len(hp.Points))
	}

	// First point should be 1000 (trimmed first 1000)
	if hp.Points[0].X != 1000.0 {
		t.Errorf("First point after trim should be 1000.0, got %v", hp.Points[0].X)
	}
}

// TestGenerateGridLines verifies grid generation.
func TestGenerateGridLines(t *testing.T) {
	tests := []struct {
		name      string
		xMin      float64
		xMax      float64
		yMin      float64
		yMax      float64
		divisions int
		expected  int
	}{
		{
			name:      "10 divisions",
			xMin:      -2.0,
			xMax:      2.0,
			yMin:      -50.0,
			yMax:      50.0,
			divisions: 10,
			expected:  (10 + 1) * 2 * 2, // (divisions+1) * 2 endpoints * 2 directions
		},
		{
			name:      "5 divisions",
			xMin:      0.0,
			xMax:      1.0,
			yMin:      0.0,
			yMax:      1.0,
			divisions: 5,
			expected:  (5 + 1) * 2 * 2,
		},
		{
			name:      "1 division",
			xMin:      0.0,
			xMax:      1.0,
			yMin:      0.0,
			yMax:      1.0,
			divisions: 1,
			expected:  (1 + 1) * 2 * 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			points := GenerateGridLines(tt.xMin, tt.xMax, tt.yMin, tt.yMax, tt.divisions)

			if len(points) != tt.expected {
				t.Errorf("GenerateGridLines returned %d points, want %d", len(points), tt.expected)
			}

			// Verify that points alternate between start and end
			for i := 0; i < len(points); i += 2 {
				if i+1 >= len(points) {
					break
				}
				// Each pair should form a line segment
				p1 := points[i]
				p2 := points[i+1]

				// Either X or Y should match (horizontal or vertical line)
				if p1.X != p2.X && p1.Y != p2.Y {
					t.Errorf("Point pair %d should form a line: (%v, %v) to (%v, %v)",
						i/2, p1.X, p1.Y, p2.X, p2.Y)
				}
			}
		})
	}
}

// TestLerpColor verifies color interpolation.
func TestLerpColor(t *testing.T) {
	red := Color{1.0, 0.0, 0.0, 1.0}
	blue := Color{0.0, 0.0, 1.0, 1.0}

	tests := []struct {
		name     string
		a, b     Color
		t        float32
		expected Color
	}{
		{
			name:     "t=0 returns a",
			a:        red,
			b:        blue,
			t:        0.0,
			expected: red,
		},
		{
			name:     "t=1 returns b",
			a:        red,
			b:        blue,
			t:        1.0,
			expected: blue,
		},
		{
			name:     "t=0.5 returns midpoint",
			a:        red,
			b:        blue,
			t:        0.5,
			expected: Color{0.5, 0.0, 0.5, 1.0},
		},
		{
			name:     "t=0.25",
			a:        red,
			b:        blue,
			t:        0.25,
			expected: Color{0.75, 0.0, 0.25, 1.0},
		},
		{
			name:     "t=0.75",
			a:        red,
			b:        blue,
			t:        0.75,
			expected: Color{0.25, 0.0, 0.75, 1.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LerpColor(tt.a, tt.b, tt.t)
			epsilon := float32(1e-6)

			if math.Abs(float64(result.R-tt.expected.R)) > float64(epsilon) ||
				math.Abs(float64(result.G-tt.expected.G)) > float64(epsilon) ||
				math.Abs(float64(result.B-tt.expected.B)) > float64(epsilon) ||
				math.Abs(float64(result.A-tt.expected.A)) > float64(epsilon) {
				t.Errorf("LerpColor(%+v, %+v, %v) = %+v, want %+v",
					tt.a, tt.b, tt.t, result, tt.expected)
			}
		})
	}
}

// TestSmoothStep verifies smoothstep function behavior.
func TestSmoothStep(t *testing.T) {
	tests := []struct {
		name     string
		edge0    float64
		edge1    float64
		x        float64
		expected float64
	}{
		{
			name:     "x before edge0",
			edge0:    0.0,
			edge1:    1.0,
			x:        -0.5,
			expected: 0.0,
		},
		{
			name:     "x after edge1",
			edge0:    0.0,
			edge1:    1.0,
			x:        1.5,
			expected: 1.0,
		},
		{
			name:     "x at edge0",
			edge0:    0.0,
			edge1:    1.0,
			x:        0.0,
			expected: 0.0,
		},
		{
			name:     "x at edge1",
			edge0:    0.0,
			edge1:    1.0,
			x:        1.0,
			expected: 1.0,
		},
		{
			name:     "x at midpoint",
			edge0:    0.0,
			edge1:    1.0,
			x:        0.5,
			expected: 0.5, // t=0.5 -> 0.5*0.5*(3-2*0.5) = 0.5
		},
		{
			name:     "x at 0.25",
			edge0:    0.0,
			edge1:    1.0,
			x:        0.25,
			expected: 0.15625, // t=0.25 -> 0.25*0.25*(3-0.5)
		},
		{
			name:     "negative range",
			edge0:    -1.0,
			edge1:    0.0,
			x:        -0.5,
			expected: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SmoothStep(tt.edge0, tt.edge1, tt.x)
			epsilon := 1e-6

			if math.Abs(result-tt.expected) > epsilon {
				t.Errorf("SmoothStep(%v, %v, %v) = %v, want %v",
					tt.edge0, tt.edge1, tt.x, result, tt.expected)
			}
		})
	}
}

// TestSmoothStepMonotonic verifies that smoothstep is monotonically increasing.
func TestSmoothStepMonotonic(t *testing.T) {
	prev := 0.0
	for x := 0.0; x <= 1.0; x += 0.01 {
		result := SmoothStep(0.0, 1.0, x)
		if result < prev {
			t.Errorf("SmoothStep not monotonic at x=%v: %v < %v", x, result, prev)
		}
		prev = result
	}
}

// TestCalculateTickPositions verifies tick mark generation.
func TestCalculateTickPositions(t *testing.T) {
	tests := []struct {
		name     string
		min      float64
		max      float64
		numTicks int
		expected []float64
	}{
		{
			name:     "0 to 1, 10 ticks",
			min:      0.0,
			max:      1.0,
			numTicks: 10,
			expected: []float64{0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
		},
		{
			name:     "-1 to 1, 4 ticks",
			min:      -1.0,
			max:      1.0,
			numTicks: 4,
			expected: []float64{-1.0, -0.5, 0.0, 0.5, 1.0},
		},
		{
			name:     "0 to 100, 5 ticks",
			min:      0.0,
			max:      100.0,
			numTicks: 5,
			expected: []float64{0.0, 20.0, 40.0, 60.0, 80.0, 100.0},
		},
		{
			name:     "single tick (0 divisions)",
			min:      0.0,
			max:      10.0,
			numTicks: 0,
			expected: []float64{0.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateTickPositions(tt.min, tt.max, tt.numTicks)

			if len(result) != len(tt.expected) {
				t.Errorf("CalculateTickPositions returned %d ticks, want %d", len(result), len(tt.expected))
				return
			}

			epsilon := 1e-9
			for i := range result {
				if math.Abs(result[i]-tt.expected[i]) > epsilon {
					t.Errorf("Tick %d: got %v, want %v", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// TestFormatTickLabel verifies tick label formatting.
func TestFormatTickLabel(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected string
	}{
		{"zero", 0.0, "0"},
		{"small positive", 1.23, "1.23"},
		{"small negative", -4.56, "-4.56"},
		{"large positive", 1234.0, "1.2e+03"},
		{"large negative", -5678.0, "-5.7e+03"},
		{"very small positive", 0.001, "1.0e-03"},
		{"very small negative", -0.001, "-1.0e-03"},
		{"near zero", 1e-11, "0"},
		{"exact integer", 42.0, "42.00"},
		{"two decimals", 3.14159, "3.14"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTickLabel(tt.value)

			if result != tt.expected {
				t.Errorf("FormatTickLabel(%v) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

// TestGeneratePlotVertices verifies plot vertex generation.
func TestGeneratePlotVertices(t *testing.T) {
	hp := NewHysteresisPlot(2.0, 50.0)
	hp.AddPoint(0.0, 0.0)
	hp.AddPoint(1.0, 25.0)
	hp.AddPoint(2.0, 50.0)

	cfg := DefaultPlotConfig()

	vertices := GeneratePlotVertices(hp, cfg)

	// Should have non-zero vertices (grid + axes + curve + marker)
	if len(vertices) == 0 {
		t.Error("GeneratePlotVertices should return non-zero vertices")
	}

	// With grid enabled, should have grid vertices
	if !cfg.ShowGrid {
		t.Error("Default config should have grid enabled")
	}

	// Test without grid
	cfg.ShowGrid = false
	verticesNoGrid := GeneratePlotVertices(hp, cfg)

	// Should have fewer vertices without grid
	if len(verticesNoGrid) >= len(vertices) {
		t.Errorf("Vertices without grid (%d) should be fewer than with grid (%d)",
			len(verticesNoGrid), len(vertices))
	}
}

// TestGeneratePlotVerticesEdgeCases verifies edge case handling.
func TestGeneratePlotVerticesEdgeCases(t *testing.T) {
	t.Run("empty plot", func(t *testing.T) {
		hp := NewHysteresisPlot(2.0, 50.0)
		cfg := DefaultPlotConfig()

		vertices := GeneratePlotVertices(hp, cfg)

		// Should still generate grid and axes, just no curve
		if len(vertices) == 0 {
			t.Error("Should generate vertices even with empty plot")
		}
	})

	t.Run("single point", func(t *testing.T) {
		hp := NewHysteresisPlot(2.0, 50.0)
		hp.AddPoint(0.0, 0.0)
		cfg := DefaultPlotConfig()

		vertices := GeneratePlotVertices(hp, cfg)

		// Should generate grid, axes, and marker (no curve line)
		if len(vertices) == 0 {
			t.Error("Should generate vertices with single point")
		}
	})

	t.Run("zero range", func(t *testing.T) {
		hp := NewHysteresisPlot(0.0, 0.0)
		hp.AddPoint(0.0, 0.0)
		cfg := DefaultPlotConfig()

		// Should not panic with zero range
		vertices := GeneratePlotVertices(hp, cfg)

		if len(vertices) == 0 {
			t.Error("Should handle zero range without panicking")
		}
	})
}

// TestLevelIndicatorSetFromPolarization verifies level setting from polarization.
func TestLevelIndicatorSetFromPolarization(t *testing.T) {
	li := NewLevelIndicator()

	tests := []struct {
		name     string
		normP    float64
		expected int
	}{
		{"max positive (+1)", 1.0, 29},
		{"max negative (-1)", -1.0, 0},
		{"zero", 0.0, 15},
		{"mid positive", 0.5, 22},
		{"mid negative", -0.5, 7},
		{"above range (+2) clamped", 2.0, 29},
		{"below range (-2) clamped", -2.0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			li.SetFromPolarization(tt.normP)

			if li.CurrentLevel != tt.expected {
				t.Errorf("SetFromPolarization(%v) set level to %d, want %d",
					tt.normP, li.CurrentLevel, tt.expected)
			}
		})
	}
}

// TestDrawAxes verifies axis point generation.
func TestDrawAxes(t *testing.T) {
	points := DrawAxes(-2.0, 2.0, -50.0, 50.0)

	// Should return 4 points (2 for X-axis, 2 for Y-axis)
	if len(points) != 4 {
		t.Errorf("DrawAxes should return 4 points, got %d", len(points))
	}

	// X-axis points
	if points[0].X != -2.0 || points[0].Y != 0.0 {
		t.Errorf("X-axis start point incorrect: got (%v, %v), want (-2.0, 0.0)",
			points[0].X, points[0].Y)
	}
	if points[1].X != 2.0 || points[1].Y != 0.0 {
		t.Errorf("X-axis end point incorrect: got (%v, %v), want (2.0, 0.0)",
			points[1].X, points[1].Y)
	}

	// Y-axis points
	if points[2].X != 0.0 || points[2].Y != -50.0 {
		t.Errorf("Y-axis start point incorrect: got (%v, %v), want (0.0, -50.0)",
			points[2].X, points[2].Y)
	}
	if points[3].X != 0.0 || points[3].Y != 50.0 {
		t.Errorf("Y-axis end point incorrect: got (%v, %v), want (0.0, 50.0)",
			points[3].X, points[3].Y)
	}
}

// TestClearPlot verifies plot clearing functionality.
func TestClearPlot(t *testing.T) {
	hp := NewHysteresisPlot(2.0, 50.0)

	// Add some points
	hp.AddPoint(1.0, 25.0)
	hp.AddPoint(2.0, 50.0)

	if len(hp.Points) == 0 {
		t.Fatal("Points should be non-zero before clear")
	}

	// Clear
	hp.Clear()

	if len(hp.Points) != 0 {
		t.Errorf("After Clear(), should have 0 points, got %d", len(hp.Points))
	}

	if hp.CurrentE != 0.0 || hp.CurrentP != 0.0 {
		t.Errorf("After Clear(), current position should be (0, 0), got (%v, %v)",
			hp.CurrentE, hp.CurrentP)
	}
}
