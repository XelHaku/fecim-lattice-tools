package main

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestGetDemos(t *testing.T) {
	demos := GetDemos()

	// Should have 8 demos
	if len(demos) != 8 {
		t.Errorf("Expected 8 demos, got %d", len(demos))
	}

	// Verify demo numbers are 1-8
	for i, demo := range demos {
		expectedNum := i + 1
		if demo.Number != expectedNum {
			t.Errorf("Demo %d has Number %d, expected %d", i, demo.Number, expectedNum)
		}
	}

	// Demos 1-3 should be ready
	readyDemos := []int{1, 2, 3}
	for _, num := range readyDemos {
		if !demos[num-1].Ready {
			t.Errorf("Demo %d should be ready", num)
		}
	}

	// Demos 4-8 should not be ready
	notReadyDemos := []int{4, 5, 6, 7, 8}
	for _, num := range notReadyDemos {
		if demos[num-1].Ready {
			t.Errorf("Demo %d should not be ready", num)
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

func TestComingSoonTabNotReady(t *testing.T) {
	demos := GetDemos()

	// Count ready vs not ready
	readyCount := 0
	for _, d := range demos {
		if d.Ready {
			readyCount++
		}
	}

	// Should have exactly 3 ready demos
	if readyCount != 3 {
		t.Errorf("Expected 3 ready demos, got %d", readyCount)
	}
}
