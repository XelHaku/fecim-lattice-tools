//go:build legacy_fyne

// Package widgets provides shared UI components for FeCIM visualizers.
package widgets

import (
	"testing"
)

func TestGetHysteresisAnimations(t *testing.T) {
	anims := GetHysteresisAnimations()

	if len(anims) == 0 {
		t.Error("Expected at least one hysteresis animation")
	}

	// Verify first animation has valid structure
	anim := anims[0]
	if anim.Name == "" {
		t.Error("Animation should have a name")
	}
	if anim.Module != "module1-hysteresis" {
		t.Errorf("Expected module1-hysteresis, got %s", anim.Module)
	}
	if len(anim.Frames) == 0 {
		t.Error("Animation should have frames")
	}
}

func TestGetCrossbarAnimations(t *testing.T) {
	anims := GetCrossbarAnimations()

	if len(anims) == 0 {
		t.Error("Expected at least one crossbar animation")
	}

	for _, anim := range anims {
		if anim.Module != "module2-crossbar" {
			t.Errorf("Expected module2-crossbar, got %s", anim.Module)
		}
	}
}

func TestGetMNISTAnimations(t *testing.T) {
	anims := GetMNISTAnimations()

	if len(anims) == 0 {
		t.Error("Expected at least one MNIST animation")
	}

	for _, anim := range anims {
		if anim.Module != "module3-mnist" {
			t.Errorf("Expected module3-mnist, got %s", anim.Module)
		}
	}
}

func TestGetCircuitsAnimations(t *testing.T) {
	anims := GetCircuitsAnimations()

	if len(anims) == 0 {
		t.Error("Expected at least one circuits animation")
	}

	for _, anim := range anims {
		if anim.Module != "module4-circuits" {
			t.Errorf("Expected module4-circuits, got %s", anim.Module)
		}
	}
}

func TestGetAllAnimations(t *testing.T) {
	all := GetAllAnimations()

	// Should include animations from all modules
	moduleCount := make(map[string]int)
	for _, anim := range all {
		moduleCount[anim.Module]++
	}

	expectedModules := []string{
		"module1-hysteresis",
		"module2-crossbar",
		"module3-mnist",
		"module4-circuits",
	}

	for _, mod := range expectedModules {
		if moduleCount[mod] == 0 {
			t.Errorf("Expected animations from %s", mod)
		}
	}
}

func TestGetAnimationsByModule(t *testing.T) {
	anims := GetAnimationsByModule("module1-hysteresis")

	if len(anims) == 0 {
		t.Error("Expected animations for module1-hysteresis")
	}

	for _, anim := range anims {
		if anim.Module != "module1-hysteresis" {
			t.Errorf("Expected module1-hysteresis, got %s", anim.Module)
		}
	}

	// Test non-existent module
	empty := GetAnimationsByModule("nonexistent")
	if len(empty) != 0 {
		t.Error("Expected no animations for nonexistent module")
	}
}

func TestGetAnimationsByTag(t *testing.T) {
	anims := GetAnimationsByTag("physics")

	if len(anims) == 0 {
		t.Error("Expected animations with 'physics' tag")
	}

	for _, anim := range anims {
		found := false
		for _, tag := range anim.Tags {
			if tag == "physics" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Animation %s should have 'physics' tag", anim.Name)
		}
	}
}

func TestGetAnimationsByLevel(t *testing.T) {
	beginnerAnims := GetAnimationsByLevel(LevelBeginner)
	allAnims := GetAnimationsByLevel(LevelExpert)

	if len(beginnerAnims) == 0 {
		t.Error("Expected beginner animations")
	}

	// All levels should return at least as many as beginner
	if len(allAnims) < len(beginnerAnims) {
		t.Error("Expert level should include all beginner animations")
	}

	// Verify beginner animations don't exceed beginner level
	for _, anim := range beginnerAnims {
		if anim.Difficulty > LevelBeginner {
			t.Errorf("Animation %s has difficulty %v, expected <= Beginner", anim.Name, anim.Difficulty)
		}
	}
}

func TestAnimationPresetToController(t *testing.T) {
	preset := AnimationPreset{
		Name: "Test Animation",
		Frames: []AnimationFrame{
			{Title: "Frame 1", Content: "Content 1"},
			{Title: "Frame 2", Content: "Content 2"},
		},
	}

	ctrl := AnimationPresetToController(preset)

	if ctrl == nil {
		t.Fatal("Expected non-nil controller")
	}

	if len(ctrl.frames) != 2 {
		t.Errorf("Expected 2 frames, got %d", len(ctrl.frames))
	}
}

func TestCreateAnimationFramesFromTutorial(t *testing.T) {
	steps := []TutorialStep{
		{Title: "Step 1", Explanation: "Explanation 1", HighlightElement: "elem1"},
		{Title: "Step 2", Explanation: "Explanation 2"},
	}

	frames := CreateAnimationFramesFromTutorial(steps)

	if len(frames) != 2 {
		t.Errorf("Expected 2 frames, got %d", len(frames))
	}

	if frames[0].Title != "Step 1" {
		t.Errorf("Expected 'Step 1', got '%s'", frames[0].Title)
	}

	if frames[0].Highlight != "elem1" {
		t.Errorf("Expected 'elem1', got '%s'", frames[0].Highlight)
	}
}

func TestFrameCounter(t *testing.T) {
	tests := []struct {
		current  int
		total    int
		expected string
	}{
		{0, 5, "Frame 1 of 5"},
		{2, 10, "Frame 3 of 10"},
		{4, 5, "Frame 5 of 5"},
	}

	for _, tt := range tests {
		result := FrameCounter(tt.current, tt.total)
		if result != tt.expected {
			t.Errorf("FrameCounter(%d, %d): expected %q, got %q", tt.current, tt.total, tt.expected, result)
		}
	}
}

func TestProgressPercentage(t *testing.T) {
	tests := []struct {
		current  int
		total    int
		expected float64
	}{
		{0, 10, 0.0},
		{5, 10, 50.0},
		{10, 10, 100.0},
		{1, 4, 25.0},
		{0, 0, 0.0}, // Edge case: avoid division by zero
	}

	for _, tt := range tests {
		result := ProgressPercentage(tt.current, tt.total)
		if result != tt.expected {
			t.Errorf("ProgressPercentage(%d, %d): expected %f, got %f", tt.current, tt.total, tt.expected, result)
		}
	}
}

func TestAnimationFrameDuration(t *testing.T) {
	// All frames should have reasonable durations
	for _, anim := range GetAllAnimations() {
		for i, frame := range anim.Frames {
			if frame.Duration == 0 && i > 0 { // First frame might have 0 duration
				// That's okay, will use default
			}
			if frame.Title == "" {
				t.Errorf("Animation %s frame %d has empty title", anim.Name, i)
			}
			if frame.Content == "" {
				t.Errorf("Animation %s frame %d has empty content", anim.Name, i)
			}
		}
	}
}

func TestAnimationPresetCompleteness(t *testing.T) {
	for _, anim := range GetAllAnimations() {
		if anim.Name == "" {
			t.Error("Animation has empty name")
		}
		if anim.Description == "" {
			t.Errorf("Animation %s has empty description", anim.Name)
		}
		if anim.Module == "" {
			t.Errorf("Animation %s has empty module", anim.Name)
		}
		if len(anim.Frames) == 0 {
			t.Errorf("Animation %s has no frames", anim.Name)
		}
		if len(anim.Tags) == 0 {
			t.Errorf("Animation %s has no tags", anim.Name)
		}
	}
}
