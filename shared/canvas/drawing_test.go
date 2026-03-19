package utils

import (
	"image"
	"image/color"
	"testing"

	"fecim-lattice-tools/shared/mathutil"
)

func TestDrawRect(t *testing.T) {
	tests := []struct {
		name   string
		imgW   int
		imgH   int
		x, y   int
		rectW  int
		rectH  int
		color  color.Color
		checkX int
		checkY int
		expect bool // true if pixel should be colored
	}{
		{
			name: "basic rectangle",
			imgW: 100, imgH: 100,
			x: 10, y: 10, rectW: 20, rectH: 20,
			color:  color.RGBA{255, 0, 0, 255},
			checkX: 15, checkY: 15,
			expect: true,
		},
		{
			name: "pixel outside rectangle",
			imgW: 100, imgH: 100,
			x: 10, y: 10, rectW: 20, rectH: 20,
			color:  color.RGBA{255, 0, 0, 255},
			checkX: 5, checkY: 5,
			expect: false,
		},
		{
			name: "rectangle at origin",
			imgW: 50, imgH: 50,
			x: 0, y: 0, rectW: 10, rectH: 10,
			color:  color.RGBA{0, 255, 0, 255},
			checkX: 5, checkY: 5,
			expect: true,
		},
		{
			name: "negative coordinates clipped",
			imgW: 100, imgH: 100,
			x: -5, y: -5, rectW: 20, rectH: 20,
			color:  color.RGBA{0, 0, 255, 255},
			checkX: 5, checkY: 5,
			expect: true,
		},
		{
			name: "rectangle extends beyond bounds",
			imgW: 50, imgH: 50,
			x: 40, y: 40, rectW: 20, rectH: 20,
			color:  color.RGBA{255, 255, 0, 255},
			checkX: 45, checkY: 45,
			expect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, tt.imgW, tt.imgH))

			DrawRect(img, tt.x, tt.y, tt.rectW, tt.rectH, tt.color)

			pixel := img.At(tt.checkX, tt.checkY)
			r, g, b, a := pixel.RGBA()
			isColored := r != 0 || g != 0 || b != 0 || a != 0

			if isColored != tt.expect {
				t.Errorf("pixel at (%d,%d): expected colored=%v, got colored=%v",
					tt.checkX, tt.checkY, tt.expect, isColored)
			}
		})
	}
}

func TestDrawRectBorder(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	borderColor := color.RGBA{255, 0, 0, 255}

	DrawRectBorder(img, 20, 20, 40, 30, borderColor)

	// Check top edge
	pixel := img.At(30, 20)
	if pixel != borderColor {
		t.Error("top edge not drawn correctly")
	}

	// Check bottom edge
	pixel = img.At(30, 49) // y + rectH - 1 = 20 + 30 - 1 = 49
	if pixel != borderColor {
		t.Error("bottom edge not drawn correctly")
	}

	// Check left edge
	pixel = img.At(20, 30)
	if pixel != borderColor {
		t.Error("left edge not drawn correctly")
	}

	// Check right edge
	pixel = img.At(59, 30) // x + rectW - 1 = 20 + 40 - 1 = 59
	if pixel != borderColor {
		t.Error("right edge not drawn correctly")
	}

	// Check center is NOT filled
	pixel = img.At(40, 35)
	r, g, b, a := pixel.RGBA()
	if r != 0 || g != 0 || b != 0 || a != 0 {
		t.Error("center should not be filled for border-only rectangle")
	}
}

func TestDrawRoundedRect(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	fillColor := color.RGBA{0, 255, 0, 255}

	DrawRoundedRect(img, 10, 10, 50, 40, 5, fillColor)

	// Check center is filled
	pixel := img.At(35, 30)
	if pixel != fillColor {
		t.Error("center should be filled")
	}

	// Check corner pixel outside the rounded area
	// Top-left corner at (10, 10) with radius 5
	// Pixel at (10, 10) should be outside the rounded corner
	pixel = img.At(10, 10)
	r, g, b, a := pixel.RGBA()
	if r != 0 || g != 0 || b != 0 || a != 0 {
		t.Error("corner pixel should be clipped due to rounding")
	}

	// Check pixel just inside the rounded corner area
	pixel = img.At(16, 16)
	if pixel != fillColor {
		t.Error("pixel inside rounded area should be filled")
	}
}

func TestDrawGradientRect(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	topColor := color.RGBA{255, 0, 0, 255}    // Red
	bottomColor := color.RGBA{0, 0, 255, 255} // Blue

	DrawGradientRect(img, 10, 10, 30, 50, topColor, bottomColor)

	// Check top row is red
	pixel := img.RGBAAt(20, 10)
	if pixel.R < 200 || pixel.B > 50 {
		t.Errorf("top should be mostly red, got R=%d B=%d", pixel.R, pixel.B)
	}

	// Check bottom row is blue
	pixel = img.RGBAAt(20, 59) // y + rectH - 1 = 10 + 50 - 1 = 59
	if pixel.B < 200 || pixel.R > 50 {
		t.Errorf("bottom should be mostly blue, got R=%d B=%d", pixel.R, pixel.B)
	}

	// Check middle has both colors
	pixel = img.RGBAAt(20, 35)
	if pixel.R < 50 || pixel.B < 50 {
		t.Errorf("middle should have both colors, got R=%d B=%d", pixel.R, pixel.B)
	}
}

func TestDrawGlowCircle(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	centerColor := color.RGBA{255, 255, 255, 255}
	glowColor := color.RGBA{255, 255, 0, 128}

	DrawGlowCircle(img, 50, 50, 10, centerColor, glowColor)

	// Check center is white
	pixel := img.RGBAAt(50, 50)
	if pixel != centerColor {
		t.Errorf("center should be white, got %v", pixel)
	}

	// Check inside radius is white
	pixel = img.RGBAAt(55, 50)
	if pixel != centerColor {
		t.Errorf("inside radius should be white, got %v", pixel)
	}

	// Check outside center but inside glow has some color
	pixel = img.RGBAAt(62, 50) // Just outside radius but inside glow
	if pixel.R == 0 && pixel.G == 0 && pixel.B == 0 {
		t.Error("glow region should have some color")
	}
}

func TestLevelToColor(t *testing.T) {
	tests := []struct {
		name     string
		level    int
		maxLevel int
		expectR  uint8 // approximate
		expectB  uint8 // approximate
	}{
		{
			name:     "single level returns white",
			level:    0,
			maxLevel: 1,
			expectR:  255,
			expectB:  255,
		},
		{
			name:     "level 0 is blue",
			level:    0,
			maxLevel: 30,
			expectR:  0,
			expectB:  255,
		},
		{
			name:     "max level is red",
			level:    29,
			maxLevel: 30,
			expectR:  255,
			expectB:  0,
		},
		{
			name:     "mid level is white",
			level:    15,
			maxLevel: 31, // odd so mid is exactly at center
			expectR:  255,
			expectB:  255,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := LevelToColor(tt.level, tt.maxLevel)

			// Allow some tolerance for interpolation
			tolerance := uint8(30)

			if mathutil.AbsInt(int(c.R)-int(tt.expectR)) > int(tolerance) {
				t.Errorf("expected R~%d, got R=%d", tt.expectR, c.R)
			}
			if mathutil.AbsInt(int(c.B)-int(tt.expectB)) > int(tolerance) {
				t.Errorf("expected B~%d, got B=%d", tt.expectB, c.B)
			}
			if c.A != 255 {
				t.Errorf("expected A=255, got A=%d", c.A)
			}
		})
	}
}

func TestDrawThickLine(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	lineColor := color.RGBA{255, 0, 0, 255}

	// Draw horizontal line
	DrawThickLine(img, 10, 50, 90, 50, 3, lineColor)

	// Check line center
	pixel := img.At(50, 50)
	if pixel != lineColor {
		t.Error("line center should be colored")
	}

	// Check line start
	pixel = img.At(10, 50)
	if pixel != lineColor {
		t.Error("line start should be colored")
	}

	// Check line end
	pixel = img.At(90, 50)
	if pixel != lineColor {
		t.Error("line end should be colored")
	}

	// Check thickness - above line center
	pixel = img.At(50, 49)
	if pixel != lineColor {
		t.Error("line should have thickness above center")
	}

	// Check outside thickness
	pixel = img.At(50, 45)
	r, g, b, a := pixel.RGBA()
	if r != 0 || g != 0 || b != 0 || a != 0 {
		t.Error("outside line thickness should not be colored")
	}
}

func TestDrawThickLine_Diagonal(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	lineColor := color.RGBA{0, 255, 0, 255}

	// Draw diagonal line
	DrawThickLine(img, 10, 10, 90, 90, 2, lineColor)

	// Check along the diagonal
	pixel := img.At(50, 50)
	if pixel != lineColor {
		t.Error("diagonal line center should be colored")
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		input  int
		expect int
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
		{-100, 100},
	}

	for _, tt := range tests {
		result := mathutil.AbsInt(tt.input)
		if result != tt.expect {
			t.Errorf("mathutil.AbsInt(%d) = %d, expected %d", tt.input, result, tt.expect)
		}
	}
}

func BenchmarkDrawRect(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 500, 500))
	c := color.RGBA{255, 0, 0, 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DrawRect(img, 100, 100, 300, 300, c)
	}
}

func BenchmarkDrawRoundedRect(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 500, 500))
	c := color.RGBA{255, 0, 0, 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DrawRoundedRect(img, 100, 100, 300, 300, 10, c)
	}
}

func BenchmarkLevelToColor(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = LevelToColor(i%30, 30)
	}
}
