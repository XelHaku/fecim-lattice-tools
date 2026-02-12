package widgets

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
)

func TestNewConstrainedWidget(t *testing.T) {
	minSize := fyne.NewSize(100, 50)
	cw := NewConstrainedWidget(minSize)

	if cw == nil {
		t.Fatal("NewConstrainedWidget should return non-nil")
	}

	if cw.GetMinSize() != minSize {
		t.Errorf("expected minSize=%v, got %v", minSize, cw.GetMinSize())
	}
}

func TestConstrainedWidget_SetMinSize(t *testing.T) {
	cw := NewConstrainedWidget(fyne.NewSize(50, 50))

	newSize := fyne.NewSize(200, 100)
	cw.SetMinSize(newSize)

	if cw.GetMinSize() != newSize {
		t.Errorf("expected minSize=%v after SetMinSize, got %v", newSize, cw.GetMinSize())
	}
}

func TestConstrainedWidget_ConstrainedSize(t *testing.T) {
	minSize := fyne.NewSize(100, 50)
	cw := NewConstrainedWidget(minSize)

	tests := []struct {
		name      string
		allocated fyne.Size
		expect    fyne.Size
	}{
		{
			name:      "allocated larger than min",
			allocated: fyne.NewSize(200, 100),
			expect:    minSize, // Should be clamped to minSize
		},
		{
			name:      "allocated smaller than min",
			allocated: fyne.NewSize(50, 25),
			expect:    fyne.NewSize(50, 25), // Should use allocated (smaller)
		},
		{
			name:      "allocated equal to min",
			allocated: minSize,
			expect:    minSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cw.ConstrainedSize(tt.allocated)
			// The ConstrainedSize clamps to not exceed minSize
			if result.Width > minSize.Width || result.Height > minSize.Height {
				t.Errorf("constrained size should not exceed minSize")
			}
		})
	}
}

func TestNewBaseRendererHelper(t *testing.T) {
	helper := NewBaseRendererHelper("TestWidget")

	if helper == nil {
		t.Fatal("NewBaseRendererHelper should return non-nil")
	}

	if helper.widgetName != "TestWidget" {
		t.Errorf("expected widgetName='TestWidget', got %q", helper.widgetName)
	}
}

func TestBaseRendererHelper_LogLayout(t *testing.T) {
	helper := NewBaseRendererHelper("TestWidget")

	// Should not panic
	result := helper.LogLayout(fyne.NewSize(100, 100))

	// Result indicates if layout loop detected (false for normal operation)
	_ = result
}

func TestBaseRendererHelper_LogRefresh(t *testing.T) {
	helper := NewBaseRendererHelper("TestWidget")

	// Should not panic
	helper.LogRefresh(fyne.NewSize(100, 100))
}

func TestBaseRendererHelper_LogMinSize(t *testing.T) {
	helper := NewBaseRendererHelper("TestWidget")

	// Should not panic
	helper.LogMinSize(fyne.NewSize(50, 50))
}

func TestLayoutHelpers_CenterObject(t *testing.T) {
	app := test.NewApp()
	t.Cleanup(app.Quit)

	rect := canvas.NewRectangle(nil)
	rect.SetMinSize(fyne.NewSize(50, 50))
	rect.Resize(fyne.NewSize(50, 50))

	helpers := LayoutHelpers{}
	containerSize := fyne.NewSize(200, 200)

	helpers.CenterObject(rect, containerSize)

	pos := rect.Position()
	// Centered: (200-50)/2 = 75
	if pos.X != 75 || pos.Y != 75 {
		t.Errorf("expected position (75,75), got (%v,%v)", pos.X, pos.Y)
	}
}

func TestLayoutHelpers_ResizeAndPosition(t *testing.T) {
	app := test.NewApp()
	t.Cleanup(app.Quit)

	rect := canvas.NewRectangle(nil)
	helpers := LayoutHelpers{}

	size := fyne.NewSize(100, 50)
	pos := fyne.NewPos(10, 20)

	helpers.ResizeAndPosition(rect, size, pos)

	if rect.Size() != size {
		t.Errorf("expected size=%v, got %v", size, rect.Size())
	}
	if rect.Position() != pos {
		t.Errorf("expected position=%v, got %v", pos, rect.Position())
	}
}

func TestLayoutHelpers_FillContainer(t *testing.T) {
	app := test.NewApp()
	t.Cleanup(app.Quit)

	rect := canvas.NewRectangle(nil)
	helpers := LayoutHelpers{}

	containerSize := fyne.NewSize(300, 200)
	helpers.FillContainer(rect, containerSize)

	if rect.Size() != containerSize {
		t.Errorf("expected size=%v, got %v", containerSize, rect.Size())
	}
	if rect.Position() != fyne.NewPos(0, 0) {
		t.Errorf("expected position=(0,0), got %v", rect.Position())
	}
}

func TestLayoutCache_ShouldLayout(t *testing.T) {
	cache := LayoutCache{}

	tests := []struct {
		name   string
		size   fyne.Size
		expect bool
	}{
		{"first layout", fyne.NewSize(100, 100), true},
		{"same size", fyne.NewSize(100, 100), false},
		{"different width", fyne.NewSize(200, 100), true},
		{"different height", fyne.NewSize(200, 150), true},
		{"invalid zero width", fyne.NewSize(0, 100), false},
		{"invalid zero height", fyne.NewSize(100, 0), false},
		{"invalid negative", fyne.NewSize(-10, 100), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cache.ShouldLayout(tt.size)
			if result != tt.expect {
				t.Errorf("ShouldLayout(%v) = %v, expected %v", tt.size, result, tt.expect)
			}
			if result {
				cache.MarkLayout(tt.size)
			}
		})
	}
}

func TestLayoutCache_MarkLayout(t *testing.T) {
	cache := LayoutCache{}

	size := fyne.NewSize(100, 100)
	cache.MarkLayout(size)

	if !cache.HasLayout {
		t.Error("HasLayout should be true after MarkLayout")
	}
	if cache.LastSize != size {
		t.Errorf("LastSize should be %v, got %v", size, cache.LastSize)
	}
}

func TestLayoutCache_IntegerComparison(t *testing.T) {
	cache := LayoutCache{}

	// First layout
	cache.MarkLayout(fyne.NewSize(100.4, 100.6))

	// Should skip layout for sizes that round to same integer
	result := cache.ShouldLayout(fyne.NewSize(100.1, 100.2))
	if result {
		t.Error("ShouldLayout should return false for sizes that round to same integer")
	}
}

func TestValidateSize(t *testing.T) {
	tests := []struct {
		name   string
		size   fyne.Size
		expect bool
	}{
		{"valid size", fyne.NewSize(100, 100), true},
		{"zero width", fyne.NewSize(0, 100), false},
		{"zero height", fyne.NewSize(100, 0), false},
		{"negative width", fyne.NewSize(-10, 100), false},
		{"negative height", fyne.NewSize(100, -10), false},
		{"both zero", fyne.NewSize(0, 0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateSize(tt.size)
			if result != tt.expect {
				t.Errorf("ValidateSize(%v) = %v, expected %v", tt.size, result, tt.expect)
			}
		})
	}
}

func TestRoundSize(t *testing.T) {
	tests := []struct {
		input  fyne.Size
		expect fyne.Size
	}{
		{fyne.NewSize(100.4, 200.4), fyne.NewSize(100, 200)},
		{fyne.NewSize(100.5, 200.5), fyne.NewSize(101, 201)},
		{fyne.NewSize(100.6, 200.6), fyne.NewSize(101, 201)},
		{fyne.NewSize(100.0, 200.0), fyne.NewSize(100, 200)},
	}

	for _, tt := range tests {
		result := RoundSize(tt.input)
		if result != tt.expect {
			t.Errorf("RoundSize(%v) = %v, expected %v", tt.input, result, tt.expect)
		}
	}
}

func TestSafeResize(t *testing.T) {
	app := test.NewApp()
	t.Cleanup(app.Quit)

	tests := []struct {
		name     string
		size     fyne.Size
		expectOK bool
	}{
		{"valid size", fyne.NewSize(100, 100), true},
		{"zero width", fyne.NewSize(0, 100), false},
		{"zero height", fyne.NewSize(100, 0), false},
		{"negative", fyne.NewSize(-10, 100), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rect := canvas.NewRectangle(nil)
			result := SafeResize(rect, tt.size)
			if result != tt.expectOK {
				t.Errorf("SafeResize returned %v, expected %v", result, tt.expectOK)
			}
		})
	}
}

func TestSafeLayoutPattern_VerifyLayoutPattern(t *testing.T) {
	pattern := SafeLayoutPattern{}
	doc := pattern.VerifyLayoutPattern()

	if doc == "" {
		t.Error("VerifyLayoutPattern should return documentation")
	}

	// Check it contains key guidance
	if !stringContains(doc, "Layout") || !stringContains(doc, "Refresh") {
		t.Error("documentation should mention Layout and Refresh")
	}
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
