// Package widgets provides shared UI components for FeCIM visualizers.
package widgets

import (
	"testing"
	"time"
)

func TestGetHysteresisTutorials(t *testing.T) {
	tutorials := GetHysteresisTutorials()

	if len(tutorials) == 0 {
		t.Error("Expected at least one hysteresis tutorial")
	}

	// Check first tutorial structure
	tut := tutorials[0]
	if tut.ID == "" {
		t.Error("Tutorial should have an ID")
	}
	if tut.Name == "" {
		t.Error("Tutorial should have a name")
	}
	if tut.Module != "module1-hysteresis" {
		t.Errorf("Expected module1-hysteresis, got %s", tut.Module)
	}
	if len(tut.Steps) == 0 {
		t.Error("Tutorial should have steps")
	}
	if len(tut.LearningGoals) == 0 {
		t.Error("Tutorial should have learning goals")
	}
}

func TestGetCrossbarTutorials(t *testing.T) {
	tutorials := GetCrossbarTutorials()

	if len(tutorials) == 0 {
		t.Error("Expected at least one crossbar tutorial")
	}

	for _, tut := range tutorials {
		if tut.Module != "module2-crossbar" {
			t.Errorf("Expected module2-crossbar, got %s", tut.Module)
		}
	}
}

func TestGetMNISTTutorials(t *testing.T) {
	tutorials := GetMNISTTutorials()

	if len(tutorials) == 0 {
		t.Error("Expected at least one MNIST tutorial")
	}

	for _, tut := range tutorials {
		if tut.Module != "module3-mnist" {
			t.Errorf("Expected module3-mnist, got %s", tut.Module)
		}
	}
}

func TestGetAllTutorials(t *testing.T) {
	all := GetAllTutorials()

	// Should include tutorials from multiple modules
	moduleCount := make(map[string]int)
	for _, tut := range all {
		moduleCount[tut.Module]++
	}

	if len(moduleCount) < 2 {
		t.Error("Expected tutorials from at least 2 modules")
	}
}

func TestGetTutorialByID(t *testing.T) {
	// Get a known tutorial
	tut := GetTutorialByID("hys-intro")

	if tut == nil {
		t.Fatal("Expected to find hys-intro tutorial")
	}

	if tut.Name != "Introduction to Ferroelectricity" {
		t.Errorf("Unexpected tutorial name: %s", tut.Name)
	}

	// Try non-existent ID
	notFound := GetTutorialByID("nonexistent-id")
	if notFound != nil {
		t.Error("Expected nil for nonexistent tutorial")
	}
}

func TestGetTutorialsByModule(t *testing.T) {
	tutorials := GetTutorialsByModule("module1-hysteresis")

	if len(tutorials) == 0 {
		t.Error("Expected tutorials for module1-hysteresis")
	}

	for _, tut := range tutorials {
		if tut.Module != "module1-hysteresis" {
			t.Errorf("Expected module1-hysteresis, got %s", tut.Module)
		}
	}

	// Test non-existent module
	empty := GetTutorialsByModule("nonexistent-module")
	if len(empty) != 0 {
		t.Error("Expected no tutorials for nonexistent module")
	}
}

func TestTutorialToController(t *testing.T) {
	tutorial := InteractiveTutorial{
		ID:   "test-tut",
		Name: "Test Tutorial",
		Steps: []TutorialStep{
			{Title: "Step 1", Duration: 50 * time.Millisecond},
			{Title: "Step 2", Duration: 50 * time.Millisecond},
		},
	}

	ctrl := TutorialToController(tutorial)

	if ctrl == nil {
		t.Fatal("Expected non-nil controller")
	}

	if ctrl.TotalSteps() != 2 {
		t.Errorf("Expected 2 steps, got %d", ctrl.TotalSteps())
	}
}

func TestTutorialPrerequisites(t *testing.T) {
	// Check that tutorials with prerequisites reference valid IDs
	allTutorials := GetAllTutorials()
	tutorialIDs := make(map[string]bool)

	for _, tut := range allTutorials {
		tutorialIDs[tut.ID] = true
	}

	for _, tut := range allTutorials {
		for _, prereq := range tut.Prerequisites {
			if !tutorialIDs[prereq] {
				t.Errorf("Tutorial %s has invalid prerequisite: %s", tut.ID, prereq)
			}
		}
	}
}

func TestTutorialStepStructure(t *testing.T) {
	for _, tut := range GetAllTutorials() {
		for i, step := range tut.Steps {
			if step.Title == "" {
				t.Errorf("Tutorial %s step %d has empty title", tut.ID, i)
			}
			if step.Explanation == "" {
				t.Errorf("Tutorial %s step %d has empty explanation", tut.ID, i)
			}

			// Check quiz structure if present
			if step.QuizQuestion != nil {
				quiz := step.QuizQuestion
				if quiz.Question == "" {
					t.Errorf("Tutorial %s step %d quiz has empty question", tut.ID, i)
				}
				if len(quiz.Options) < 2 {
					t.Errorf("Tutorial %s step %d quiz should have at least 2 options", tut.ID, i)
				}
				if quiz.CorrectIdx < 0 || quiz.CorrectIdx >= len(quiz.Options) {
					t.Errorf("Tutorial %s step %d quiz has invalid correct index", tut.ID, i)
				}
			}
		}
	}
}

func TestTutorialDurations(t *testing.T) {
	for _, tut := range GetAllTutorials() {
		if tut.Duration <= 0 {
			t.Errorf("Tutorial %s should have positive duration", tut.ID)
		}

		// Sanity check: tutorials shouldn't be longer than 1 hour
		if tut.Duration > time.Hour {
			t.Errorf("Tutorial %s duration seems too long: %v", tut.ID, tut.Duration)
		}
	}
}

func TestTutorialDifficultyProgression(t *testing.T) {
	// For tutorials with prerequisites, difficulty should be >= prerequisite difficulty
	allTutorials := GetAllTutorials()
	tutorialsByID := make(map[string]InteractiveTutorial)

	for _, tut := range allTutorials {
		tutorialsByID[tut.ID] = tut
	}

	for _, tut := range allTutorials {
		for _, prereqID := range tut.Prerequisites {
			prereq, exists := tutorialsByID[prereqID]
			if !exists {
				continue // Already checked in TestTutorialPrerequisites
			}

			if tut.Difficulty < prereq.Difficulty {
				t.Errorf("Tutorial %s (level %v) should not require %s (level %v)",
					tut.ID, tut.Difficulty, prereqID, prereq.Difficulty)
			}
		}
	}
}

func TestTutorialLearningGoals(t *testing.T) {
	for _, tut := range GetAllTutorials() {
		if len(tut.LearningGoals) == 0 {
			t.Errorf("Tutorial %s should have learning goals", tut.ID)
		}

		for i, goal := range tut.LearningGoals {
			if goal == "" {
				t.Errorf("Tutorial %s learning goal %d is empty", tut.ID, i)
			}
		}
	}
}

func TestTutorialQuizCoverage(t *testing.T) {
	// Each tutorial should have at least one comprehension check
	for _, tut := range GetAllTutorials() {
		hasQuiz := false
		for _, step := range tut.Steps {
			if step.QuizQuestion != nil {
				hasQuiz = true
				break
			}
		}

		if !hasQuiz {
			t.Logf("Tutorial %s has no quiz questions (consider adding)", tut.ID)
		}
	}
}

func TestTutorialStepDurations(t *testing.T) {
	for _, tut := range GetAllTutorials() {
		for i, step := range tut.Steps {
			// Steps with user prompts should have 0 duration (wait for user)
			if step.UserPrompt != "" && step.Duration != 0 {
				t.Logf("Tutorial %s step %d has both UserPrompt and Duration - Duration may be ignored",
					tut.ID, i)
			}

			// Steps with quizzes should have 0 duration (wait for answer)
			if step.QuizQuestion != nil && step.Duration != 0 {
				t.Logf("Tutorial %s step %d has both Quiz and Duration - Duration may be ignored",
					tut.ID, i)
			}
		}
	}
}
