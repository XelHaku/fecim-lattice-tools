package main

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestGetDemos(t *testing.T) {
	demos := GetDemos()

	// Should have 6 consolidated demos
	if len(demos) != 6 {
		t.Errorf("Expected 6 demos, got %d", len(demos))
	}

	// Verify demo numbers are 1-6
	for i, demo := range demos {
		expectedNum := i + 1
		if demo.Number != expectedNum {
			t.Errorf("Demo %d has Number %d, expected %d", i, demo.Number, expectedNum)
		}
	}

	// All demos 1-6 should be ready
	for _, demo := range demos {
		if !demo.Ready {
			t.Errorf("Demo %d should be ready", demo.Number)
		}
	}
}

func TestDemoInfoFields(t *testing.T) {
	demos := GetDemos()

	for i, demo := range demos {
		// All demos should have a title
		if demo.Title == "" {
			t.Errorf("Demo %d has empty title", i+1)
		}

		// All demos should have a subtitle
		if demo.Subtitle == "" {
			t.Errorf("Demo %d has empty subtitle", i+1)
		}

		// All demos should have a description
		if demo.Description == "" {
			t.Errorf("Demo %d has empty description", i+1)
		}
	}
}

func TestNewDemoCard(t *testing.T) {
	info := DemoInfo{
		Number:      1,
		Title:       "Test Demo",
		Subtitle:    "Test Subtitle",
		Description: "Test description",
		Ready:       true,
	}

	tapped := false
	card := NewDemoCard(info, func() {
		tapped = true
	})

	// Should not be nil
	if card == nil {
		t.Fatal("NewDemoCard returned nil")
	}

	// MinSize should be set
	minSize := card.MinSize()
	if minSize.Width <= 0 || minSize.Height <= 0 {
		t.Errorf("MinSize is invalid: %v", minSize)
	}

	// Verify info is stored
	if card.info.Number != 1 {
		t.Errorf("Card info not stored correctly")
	}

	// Tapping a ready card should trigger callback
	card.Tapped(nil)
	if !tapped {
		t.Error("Tap callback not triggered for ready card")
	}
}

func TestDemoCardNotReadyNoCallback(t *testing.T) {
	info := DemoInfo{
		Number:      4,
		Title:       "Coming Soon",
		Subtitle:    "Test",
		Description: "Test",
		Ready:       false, // Not ready
	}

	tapped := false
	card := NewDemoCard(info, func() {
		tapped = true
	})

	// Tapping a not-ready card should NOT trigger callback
	card.Tapped(nil)
	if tapped {
		t.Error("Tap callback should not trigger for not-ready card")
	}
}

func TestDemoCardRenderer(t *testing.T) {
	info := DemoInfo{
		Number:      1,
		Title:       "Test",
		Subtitle:    "Sub",
		Description: "Description",
		Ready:       true,
	}

	card := NewDemoCard(info, nil)
	renderer := card.CreateRenderer()

	// Should not be nil
	if renderer == nil {
		t.Fatal("CreateRenderer returned nil")
	}

	// Should have MinSize
	minSize := renderer.MinSize()
	if minSize.Width <= 0 || minSize.Height <= 0 {
		t.Errorf("Renderer MinSize is invalid: %v", minSize)
	}

	// Refresh should not panic
	renderer.Refresh()

	// Objects should not be empty after refresh
	objects := renderer.Objects()
	if len(objects) == 0 {
		t.Error("Renderer has no objects after Refresh")
	}

	// Destroy should not panic
	renderer.Destroy()
}

func TestCreateLauncherContent(t *testing.T) {
	content := CreateLauncherContent(func(demoNum int) {
		_ = demoNum // Callback for tab switching
	})

	// Should not be nil
	if content == nil {
		t.Fatal("CreateLauncherContent returned nil")
	}

	// MinSize should be reasonable
	minSize := content.MinSize()
	if minSize.Width <= 0 || minSize.Height <= 0 {
		t.Errorf("Content MinSize is invalid: %v", minSize)
	}
}

func TestSplitWords(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"hello world", []string{"hello", "world"}},
		{"single", []string{"single"}},
		{"", []string{}},
		{"  multiple   spaces  ", []string{"multiple", "spaces"}},
		{"one two three", []string{"one", "two", "three"}},
	}

	for _, tt := range tests {
		result := splitWords(tt.input)

		if len(result) != len(tt.expected) {
			t.Errorf("splitWords(%q) = %v, want %v", tt.input, result, tt.expected)
			continue
		}

		for i, word := range result {
			if word != tt.expected[i] {
				t.Errorf("splitWords(%q)[%d] = %q, want %q", tt.input, i, word, tt.expected[i])
			}
		}
	}
}

func TestDemoCardWithTestApp(t *testing.T) {
	// Use Fyne test app for proper widget testing
	app := test.NewApp()
	defer app.Quit()

	info := DemoInfo{
		Number:      1,
		Title:       "Test Demo",
		Subtitle:    "Test Subtitle",
		Description: "A longer description that might wrap across multiple lines",
		Ready:       true,
	}

	card := NewDemoCard(info, nil)

	// Create in test window
	window := app.NewWindow("Test")
	window.SetContent(card)

	// Card should render without panic
	window.Show()
}

func TestAllDemosReady(t *testing.T) {
	demos := GetDemos()

	// Count ready demos
	readyCount := 0
	for _, d := range demos {
		if d.Ready {
			readyCount++
		}
	}

	// All 6 demos should be ready
	if readyCount != 6 {
		t.Errorf("Expected 6 ready demos, got %d", readyCount)
	}
}

func TestDrawPreviewThumbnail(t *testing.T) {
	// Use Fyne test app for proper widget testing
	app := test.NewApp()
	defer app.Quit()
	_ = app // Ensure test app is initialized

	accentColor := struct{ R, G, B, A uint8 }{0, 212, 255, 255}

	// Test preview thumbnail generation for all 6 demos
	for demoNum := 1; demoNum <= 6; demoNum++ {
		objects := drawPreviewThumbnail(demoNum, 10, 10, 100, 70, accentColor)

		// Each preview should generate multiple canvas objects
		if len(objects) < 3 {
			t.Errorf("Demo %d preview has too few objects: %d", demoNum, len(objects))
		}

		// First object should be background rectangle
		if objects[0] == nil {
			t.Errorf("Demo %d preview has nil background", demoNum)
		}
	}
}

func TestDrawPreviewThumbnailEdgeCases(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	_ = app

	accentColor := struct{ R, G, B, A uint8 }{0, 212, 255, 255}

	// Test with different sizes
	objects := drawPreviewThumbnail(1, 0, 0, 50, 50, accentColor)
	if len(objects) == 0 {
		t.Error("Preview should generate objects even with small size")
	}

	// Test with larger size
	objects = drawPreviewThumbnail(2, 100, 100, 200, 150, accentColor)
	if len(objects) == 0 {
		t.Error("Preview should generate objects with larger size")
	}
}
