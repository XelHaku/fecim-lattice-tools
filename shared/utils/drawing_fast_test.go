package utils

import (
	"image"
	"image/color"
	"testing"
)

func TestDrawRectFast(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	c := color.RGBA{255, 0, 0, 255}

	DrawRectFast(img, 10, 10, 20, 20, c)

	// Check center pixel
	pixel := img.RGBAAt(15, 15)
	if pixel != c {
		t.Errorf("center pixel should be %v, got %v", c, pixel)
	}

	// Check outside pixel
	pixel = img.RGBAAt(5, 5)
	if pixel.R != 0 || pixel.G != 0 || pixel.B != 0 {
		t.Error("outside pixel should be black")
	}
}

func TestDrawRectFast_ClipsCorrectly(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	c := color.RGBA{0, 255, 0, 255}

	// Draw rectangle that extends past image bounds
	DrawRectFast(img, 40, 40, 20, 20, c)

	// Check corner that should be drawn
	pixel := img.RGBAAt(45, 45)
	if pixel != c {
		t.Errorf("clipped pixel should be %v, got %v", c, pixel)
	}

	// Verify no panic for negative coords
	DrawRectFast(img, -10, -10, 20, 20, c)
	pixel = img.RGBAAt(5, 5)
	if pixel != c {
		t.Error("negative-start rect should still draw visible portion")
	}
}

func TestFillImageFast(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	c := color.RGBA{128, 64, 32, 255}

	FillImageFast(img, c)

	// Check multiple pixels
	pixels := []struct{ x, y int }{
		{0, 0},
		{99, 99},
		{50, 50},
		{0, 99},
		{99, 0},
	}

	for _, p := range pixels {
		pixel := img.RGBAAt(p.x, p.y)
		if pixel != c {
			t.Errorf("pixel at (%d,%d) should be %v, got %v", p.x, p.y, c, pixel)
		}
	}
}

func TestDrawHLineFast(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	c := color.RGBA{255, 255, 0, 255}

	DrawHLineFast(img, 10, 50, 80, c)

	// Check line pixels
	for x := 10; x < 90; x++ {
		pixel := img.RGBAAt(x, 50)
		if pixel != c {
			t.Errorf("line pixel at (%d,50) should be %v, got %v", x, c, pixel)
		}
	}

	// Check pixel before line
	pixel := img.RGBAAt(5, 50)
	if pixel.R != 0 || pixel.G != 0 || pixel.B != 0 {
		t.Error("pixel before line should be black")
	}
}

func TestDrawVLineFast(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	c := color.RGBA{0, 255, 255, 255}

	DrawVLineFast(img, 50, 10, 80, c)

	// Check line pixels
	for y := 10; y < 90; y++ {
		pixel := img.RGBAAt(50, y)
		if pixel != c {
			t.Errorf("line pixel at (50,%d) should be %v, got %v", y, c, pixel)
		}
	}
}

func TestDrawGridFast(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	c := color.RGBA{100, 100, 100, 255}

	DrawGridFast(img, 20, 20, 5, 5, c)

	// Check vertical lines at expected positions
	for col := 0; col <= 5; col++ {
		x := col * 20
		if x >= 100 {
			break
		}
		pixel := img.RGBAAt(x, 50)
		if pixel != c {
			t.Errorf("vertical line at x=%d should exist", x)
		}
	}

	// Check horizontal lines
	for row := 0; row <= 5; row++ {
		y := row * 20
		if y >= 100 {
			break
		}
		pixel := img.RGBAAt(50, y)
		if pixel != c {
			t.Errorf("horizontal line at y=%d should exist", y)
		}
	}
}

func TestBlendPixelFast(t *testing.T) {
	t.Run("fully opaque", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 10, 10))
		FillImageFast(img, color.RGBA{100, 100, 100, 255})

		BlendPixelFast(img, 5, 5, color.RGBA{200, 0, 0, 255})

		pixel := img.RGBAAt(5, 5)
		if pixel.R != 200 || pixel.G != 0 || pixel.B != 0 {
			t.Errorf("opaque blend should overwrite, got %v", pixel)
		}
	})

	t.Run("semi-transparent", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 10, 10))
		FillImageFast(img, color.RGBA{0, 0, 0, 255})

		BlendPixelFast(img, 5, 5, color.RGBA{200, 0, 0, 128})

		pixel := img.RGBAAt(5, 5)
		// Should be approximately half of 200
		if pixel.R < 90 || pixel.R > 110 {
			t.Errorf("semi-transparent blend should result in ~100 red, got %d", pixel.R)
		}
	})

	t.Run("fully transparent", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 10, 10))
		FillImageFast(img, color.RGBA{50, 50, 50, 255})

		BlendPixelFast(img, 5, 5, color.RGBA{200, 0, 0, 0})

		pixel := img.RGBAAt(5, 5)
		if pixel.R != 50 || pixel.G != 50 || pixel.B != 50 {
			t.Errorf("transparent blend should not change pixel, got %v", pixel)
		}
	})
}

func TestDrawGradientRectFast(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	topColor := color.RGBA{255, 0, 0, 255}
	bottomColor := color.RGBA{0, 0, 255, 255}

	DrawGradientRectFast(img, 10, 10, 30, 50, topColor, bottomColor)

	// Check top row
	pixel := img.RGBAAt(20, 10)
	if pixel.R < 200 || pixel.B > 50 {
		t.Errorf("top should be mostly red, got R=%d B=%d", pixel.R, pixel.B)
	}

	// Check bottom row
	pixel = img.RGBAAt(20, 59)
	if pixel.B < 200 || pixel.R > 50 {
		t.Errorf("bottom should be mostly blue, got R=%d B=%d", pixel.R, pixel.B)
	}
}

// Benchmarks comparing Fast vs Original implementations

func BenchmarkDrawRectFast(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 500, 500))
	c := color.RGBA{255, 0, 0, 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DrawRectFast(img, 100, 100, 300, 300, c)
	}
}

func BenchmarkFillImageFast(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 500, 500))
	c := color.RGBA{128, 128, 128, 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FillImageFast(img, c)
	}
}

func BenchmarkFillImageOriginal(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 500, 500))
	c := color.RGBA{128, 128, 128, 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for y := 0; y < 500; y++ {
			for x := 0; x < 500; x++ {
				img.Set(x, y, c)
			}
		}
	}
}

func BenchmarkDrawGridFast(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 350, 350))
	c := color.RGBA{70, 75, 90, 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DrawGridFast(img, 12, 12, 28, 28, c)
	}
}

func BenchmarkDrawGradientRectFast(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 500, 500))
	top := color.RGBA{255, 0, 0, 255}
	bottom := color.RGBA{0, 0, 255, 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DrawGradientRectFast(img, 100, 100, 300, 300, top, bottom)
	}
}

func BenchmarkDrawGradientRectOriginal(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 500, 500))
	top := color.RGBA{255, 0, 0, 255}
	bottom := color.RGBA{0, 0, 255, 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DrawGradientRect(img, 100, 100, 300, 300, top, bottom)
	}
}

// BenchmarkDigitCanvasRender simulates the hot path in DigitCanvas.generateImage
func BenchmarkDigitCanvasRender(b *testing.B) {
	// Simulate a 350x350 canvas with 28x28 grid
	img := image.NewRGBA(image.Rect(0, 0, 350, 350))
	bgColor := color.RGBA{20, 20, 30, 255}
	gridColor := color.RGBA{70, 75, 90, 255}

	// Simulate some drawn pixels
	var pixels [28][28]float64
	for i := 10; i < 20; i++ {
		for j := 10; j < 20; j++ {
			pixels[i][j] = 0.5 + 0.5*float64(i*j%10)/10
		}
	}

	cellW := 350.0 / 28.0
	cellH := 350.0 / 28.0

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// Fill background
		FillImageFast(img, bgColor)

		// Draw pixels
		for py := 0; py < 28; py++ {
			for px := 0; px < 28; px++ {
				value := pixels[py][px]
				if value > 0 {
					intensity := uint8(value * 255)
					c := color.RGBA{
						R: uint8(float64(intensity) * 0.7),
						G: intensity,
						B: intensity,
						A: 255,
					}

					x0 := int(float64(px) * cellW)
					y0 := int(float64(py) * cellH)
					w := int(float64(px+1)*cellW) - x0
					h := int(float64(py+1)*cellH) - y0

					DrawRectFast(img, x0, y0, w, h, c)
				}
			}
		}

		// Draw grid
		DrawGridFast(img, int(cellW), int(cellH), 28, 28, gridColor)
	}
}
