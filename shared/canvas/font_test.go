package utils

import (
	"image"
	"image/color"
	"testing"
)

func TestDrawSimpleText(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 50))
	textColor := color.RGBA{255, 255, 255, 255}

	DrawSimpleText(img, "123", 10, 10, textColor)

	// Check that some pixels are drawn (font should render something)
	hasPixels := false
	for y := 10; y < 20; y++ {
		for x := 10; x < 50; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r > 0 || g > 0 || b > 0 {
				hasPixels = true
				break
			}
		}
		if hasPixels {
			break
		}
	}

	if !hasPixels {
		t.Error("DrawSimpleText should render pixels for valid characters")
	}
}

func TestDrawSimpleChar(t *testing.T) {
	tests := []struct {
		char   rune
		exists bool
		name   string
	}{
		{'0', true, "digit 0"},
		{'9', true, "digit 9"},
		{'A', true, "uppercase A"},
		{'Z', true, "uppercase Z"},
		{'a', true, "lowercase a"},
		{'z', true, "lowercase z"},
		{'.', true, "period"},
		{'-', true, "dash"},
		{' ', true, "space"},
		{'@', false, "at sign (not in font)"},
		{'#', false, "hash (not in font)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, 20, 20))
			textColor := color.RGBA{255, 255, 255, 255}

			DrawSimpleChar(img, tt.char, 2, 2, textColor)

			// Count colored pixels
			coloredPixels := 0
			for y := 0; y < 20; y++ {
				for x := 0; x < 20; x++ {
					r, g, b, _ := img.At(x, y).RGBA()
					if r > 0 || g > 0 || b > 0 {
						coloredPixels++
					}
				}
			}

			if tt.exists && tt.char != ' ' {
				if coloredPixels == 0 {
					t.Errorf("character '%c' should render pixels", tt.char)
				}
			} else if !tt.exists {
				if coloredPixels > 0 {
					t.Errorf("character '%c' should not exist in font, but rendered %d pixels", tt.char, coloredPixels)
				}
			}
		})
	}
}

func TestDrawScaledText(t *testing.T) {
	tests := []struct {
		name  string
		scale int
	}{
		{"scale 1", 1},
		{"scale 2", 2},
		{"scale 3", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, 200, 100))
			textColor := color.RGBA{255, 255, 255, 255}

			DrawScaledText(img, "A", 10, 10, tt.scale, textColor)

			// Count colored pixels
			coloredPixels := 0
			for y := 0; y < 100; y++ {
				for x := 0; x < 200; x++ {
					r, g, b, _ := img.At(x, y).RGBA()
					if r > 0 || g > 0 || b > 0 {
						coloredPixels++
					}
				}
			}

			if coloredPixels == 0 {
				t.Error("scaled text should render pixels")
			}

			// Scaled text should have more pixels
			// Each pixel in the pattern becomes scale*scale pixels
			// So scale 2 should have roughly 4x pixels of scale 1
		})
	}
}

func TestDrawScaledChar_Bounds(t *testing.T) {
	// Test that drawing outside bounds doesn't panic
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	textColor := color.RGBA{255, 255, 255, 255}

	// These should not panic
	DrawScaledChar(img, 'A', -10, -10, 1, textColor)
	DrawScaledChar(img, 'A', 100, 100, 1, textColor)
	DrawScaledChar(img, 'A', -5, 5, 2, textColor)
}

func TestFontPatterns(t *testing.T) {
	// Verify font patterns have correct dimensions
	for char, pattern := range fontPatterns {
		if len(pattern) != 7 {
			t.Errorf("character '%c' pattern should have 7 rows, has %d", char, len(pattern))
		}
		for i, row := range pattern {
			if len(row) != 5 {
				t.Errorf("character '%c' row %d should have 5 columns, has %d", char, i, len(row))
			}
			// Verify only '0' and '1' characters
			for _, c := range row {
				if c != '0' && c != '1' {
					t.Errorf("character '%c' row %d has invalid character '%c'", char, i, c)
				}
			}
		}
	}
}

func TestDrawTextMultipleCharacters(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 20))
	textColor := color.RGBA{255, 255, 255, 255}

	DrawSimpleText(img, "ABC", 5, 5, textColor)

	// Check that characters are spaced apart (7 pixels per character at scale 1)
	// Character A starts at x=5
	// Character B starts at x=12
	// Character C starts at x=19

	// Verify we have pixels in all three character regions
	regions := []struct {
		startX int
		endX   int
	}{
		{5, 10},  // A
		{12, 17}, // B
		{19, 24}, // C
	}

	for i, region := range regions {
		hasPixels := false
		for y := 5; y < 15; y++ {
			for x := region.startX; x < region.endX; x++ {
				r, g, b, _ := img.At(x, y).RGBA()
				if r > 0 || g > 0 || b > 0 {
					hasPixels = true
					break
				}
			}
			if hasPixels {
				break
			}
		}
		if !hasPixels {
			t.Errorf("character region %d should have pixels", i)
		}
	}
}

func TestFontPatternConsistency(t *testing.T) {
	// Test that common characters are present
	expectedChars := []rune{
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
		'N', 'O', 'P', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
		'.', '-', ':', ' ',
	}

	for _, ch := range expectedChars {
		if _, ok := fontPatterns[ch]; !ok {
			t.Errorf("expected character '%c' to be in font patterns", ch)
		}
	}
}

func BenchmarkDrawSimpleText(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 50))
	textColor := color.RGBA{255, 255, 255, 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DrawSimpleText(img, "FeCIM 30", 10, 10, textColor)
	}
}

func BenchmarkDrawScaledText(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 400, 100))
	textColor := color.RGBA{255, 255, 255, 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DrawScaledText(img, "FeCIM 30", 10, 10, 3, textColor)
	}
}
