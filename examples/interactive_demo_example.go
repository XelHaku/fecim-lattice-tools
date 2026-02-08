//go:build ignore

// interactive_demo_example.go - Example of integrating interactive demos
//
// This example demonstrates how to use the enhanced demo system in FeCIM
// visualizers. It shows:
//   - Setting up demo mode selection
//   - Running interactive tutorials with quizzes
//   - Playing educational animations
//   - Displaying progress and controls
//
// Run: go run examples/interactive_demo_example.go
//
// For module-specific demos, see each module's cmd/ directory.
package main

import (
	"fmt"
	"os"
	"time"

	"fecim-lattice-tools/shared/widgets"
)

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║     FeCIM Interactive Demo System Example                        ║")
	fmt.Println("║     Demonstrating Educational Features                           ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "tutorials":
			listTutorials()
		case "animations":
			listAnimations()
		case "run-tutorial":
			if len(os.Args) > 2 {
				runTutorial(os.Args[2])
			} else {
				fmt.Println("Usage: interactive_demo_example run-tutorial <tutorial-id>")
			}
		case "run-animation":
			if len(os.Args) > 2 {
				runAnimation(os.Args[2])
			} else {
				fmt.Println("Usage: interactive_demo_example run-animation <animation-name>")
			}
		default:
			printUsage()
		}
		return
	}

	// Default: show overview
	showDemoOverview()
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  interactive_demo_example              - Show demo system overview")
	fmt.Println("  interactive_demo_example tutorials    - List available tutorials")
	fmt.Println("  interactive_demo_example animations   - List available animations")
	fmt.Println("  interactive_demo_example run-tutorial <id>    - Run a tutorial")
	fmt.Println("  interactive_demo_example run-animation <name> - Run an animation")
}

func showDemoOverview() {
	// =========================================================================
	// 1. DEMO MODE OVERVIEW
	// =========================================================================
	fmt.Println("1. DEMO MODES AVAILABLE")
	fmt.Println("   └─ The FeCIM demo system supports multiple learning modes:")
	fmt.Println()

	modes := []struct {
		icon string
		name string
		desc string
	}{
		{"⚡", "Quick Demo", "30-second automated demonstration"},
		{"📚", "Interactive Tutorial", "Step-by-step guided learning with quizzes"},
		{"🎬", "Educational Animation", "Visual concept explanations"},
		{"🧪", "Sandbox Mode", "Free exploration with hints"},
	}

	for _, m := range modes {
		fmt.Printf("   %s %s\n", m.icon, m.name)
		fmt.Printf("      %s\n\n", m.desc)
	}

	// =========================================================================
	// 2. TUTORIAL STATS
	// =========================================================================
	fmt.Println("2. TUTORIAL COVERAGE")
	tutorials := widgets.GetAllTutorials()
	byModule := make(map[string]int)
	totalSteps := 0

	for _, t := range tutorials {
		byModule[t.Module]++
		totalSteps += len(t.Steps)
	}

	fmt.Printf("   Total tutorials: %d\n", len(tutorials))
	fmt.Printf("   Total steps:     %d\n", totalSteps)
	fmt.Printf("   By module:\n")
	for mod, count := range byModule {
		fmt.Printf("     • %s: %d tutorials\n", mod, count)
	}
	fmt.Println()

	// =========================================================================
	// 3. ANIMATION STATS
	// =========================================================================
	fmt.Println("3. ANIMATION COVERAGE")
	animations := widgets.GetAllAnimations()
	animByModule := make(map[string]int)
	totalFrames := 0

	for _, a := range animations {
		animByModule[a.Module]++
		totalFrames += len(a.Frames)
	}

	fmt.Printf("   Total animations: %d\n", len(animations))
	fmt.Printf("   Total frames:     %d\n", totalFrames)
	fmt.Printf("   By module:\n")
	for mod, count := range animByModule {
		fmt.Printf("     • %s: %d animations\n", mod, count)
	}
	fmt.Println()

	// =========================================================================
	// 4. DIFFICULTY DISTRIBUTION
	// =========================================================================
	fmt.Println("4. DIFFICULTY DISTRIBUTION")
	byLevel := make(map[widgets.TutorialLevel]int)
	for _, t := range tutorials {
		byLevel[t.Difficulty]++
	}

	levels := []widgets.TutorialLevel{
		widgets.LevelBeginner,
		widgets.LevelIntermediate,
		widgets.LevelAdvanced,
		widgets.LevelExpert,
	}

	for _, level := range levels {
		count := byLevel[level]
		bar := ""
		for i := 0; i < count; i++ {
			bar += "█"
		}
		for i := count; i < 10; i++ {
			bar += "░"
		}
		fmt.Printf("   %-15s %s (%d)\n", level.String()+":", bar, count)
	}
	fmt.Println()

	// =========================================================================
	// 5. INTEGRATION EXAMPLE
	// =========================================================================
	fmt.Println("5. INTEGRATION EXAMPLE (Code)")
	fmt.Println("   See shared/widgets/ for the complete implementation.")
	fmt.Println()
	fmt.Println(`   // Create a tutorial controller
   tutorial := widgets.GetTutorialByID("hys-intro")
   ctrl := widgets.TutorialToController(*tutorial)
   
   // Set callbacks for UI updates
   ctrl.SetOnStepStart(func(step, total int, ts widgets.TutorialStep) {
       updateUI(step, total, ts)
   })
   
   ctrl.SetOnComplete(func(stats widgets.TutorialStats) {
       showCompletionScreen(stats)
   })
   
   // Start the tutorial
   ctrl.Start()`)
	fmt.Println()

	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println("Run with arguments to see more: tutorials, animations, run-tutorial, run-animation")
}

func listTutorials() {
	tutorials := widgets.GetAllTutorials()

	fmt.Println("AVAILABLE INTERACTIVE TUTORIALS")
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println()

	currentModule := ""
	for _, t := range tutorials {
		if t.Module != currentModule {
			if currentModule != "" {
				fmt.Println()
			}
			currentModule = t.Module
			fmt.Printf("📂 %s\n", currentModule)
			fmt.Println("────────────────────────────────────────────────────────────────")
		}

		// Duration in minutes
		mins := int(t.Duration.Minutes())

		fmt.Printf("   ID: %-15s  %s\n", t.ID, t.Name)
		fmt.Printf("   └─ %s\n", t.Description)
		fmt.Printf("      Level: %-12s  Duration: ~%dmin  Steps: %d\n",
			t.Difficulty, mins, len(t.Steps))

		if len(t.Prerequisites) > 0 {
			fmt.Printf("      Prerequisites: %v\n", t.Prerequisites)
		}

		fmt.Printf("      Learning Goals:\n")
		for _, goal := range t.LearningGoals {
			fmt.Printf("        • %s\n", goal)
		}
		fmt.Println()
	}
}

func listAnimations() {
	animations := widgets.GetAllAnimations()

	fmt.Println("AVAILABLE EDUCATIONAL ANIMATIONS")
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println()

	currentModule := ""
	for _, a := range animations {
		if a.Module != currentModule {
			if currentModule != "" {
				fmt.Println()
			}
			currentModule = a.Module
			fmt.Printf("📂 %s\n", currentModule)
			fmt.Println("────────────────────────────────────────────────────────────────")
		}

		// Calculate total duration
		totalDuration := time.Duration(0)
		for _, f := range a.Frames {
			totalDuration += f.Duration
		}
		secs := int(totalDuration.Seconds())

		fmt.Printf("   Name: %s\n", a.Name)
		fmt.Printf("   └─ %s\n", a.Description)
		fmt.Printf("      Level: %-12s  Duration: ~%ds  Frames: %d\n",
			a.Difficulty, secs, len(a.Frames))
		fmt.Printf("      Tags: %v\n", a.Tags)
		fmt.Println()
	}
}

func runTutorial(id string) {
	tutorial := widgets.GetTutorialByID(id)
	if tutorial == nil {
		fmt.Printf("Tutorial '%s' not found.\n", id)
		fmt.Println("Run 'tutorials' to see available tutorial IDs.")
		return
	}

	fmt.Printf("╔══════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║ TUTORIAL: %-55s ║\n", tutorial.Name)
	fmt.Printf("╚══════════════════════════════════════════════════════════════════╝\n\n")

	fmt.Printf("📖 %s\n\n", tutorial.Description)
	fmt.Printf("📊 Level: %s | Duration: ~%d min | Steps: %d\n\n",
		tutorial.Difficulty, int(tutorial.Duration.Minutes()), len(tutorial.Steps))

	fmt.Println("🎯 Learning Goals:")
	for _, goal := range tutorial.LearningGoals {
		fmt.Printf("   • %s\n", goal)
	}
	fmt.Println()

	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("Starting tutorial simulation (text-only mode)...")
	fmt.Println()

	// Create controller and run steps
	ctrl := widgets.TutorialToController(*tutorial)

	ctrl.SetOnStepStart(func(step int, total int, ts widgets.TutorialStep) {
		fmt.Printf("┌─ Step %d/%d: %s\n", step+1, total, ts.Title)
		fmt.Printf("│\n")
		// Word wrap the explanation
		words := splitWords(ts.Explanation)
		line := "│  "
		for _, word := range words {
			if len(line)+len(word) > 70 {
				fmt.Println(line)
				line = "│  " + word
			} else {
				if line == "│  " {
					line += word
				} else {
					line += " " + word
				}
			}
		}
		if line != "│  " {
			fmt.Println(line)
		}

		if ts.QuizQuestion != nil {
			fmt.Println("│")
			fmt.Printf("│  ❓ QUIZ: %s\n", ts.QuizQuestion.Question)
			for i, opt := range ts.QuizQuestion.Options {
				marker := "○"
				if i == ts.QuizQuestion.CorrectIdx {
					marker = "●"
				}
				fmt.Printf("│     %s %s\n", marker, opt)
			}
			fmt.Printf("│  → Answer: %s\n", ts.QuizQuestion.Options[ts.QuizQuestion.CorrectIdx])
		}

		if ts.UserPrompt != "" {
			fmt.Println("│")
			fmt.Printf("│  💡 %s\n", ts.UserPrompt)
		}

		fmt.Println("└────────────────────────────────────────────────────────────────")
		fmt.Println()
	})

	ctrl.SetOnComplete(func(stats widgets.TutorialStats) {
		fmt.Println()
		fmt.Println("════════════════════════════════════════════════════════════════")
		fmt.Println("🎉 TUTORIAL COMPLETE!")
		fmt.Printf("   Steps completed: %d/%d\n", stats.CompletedSteps, stats.TotalSteps)
		fmt.Printf("   Quizzes: %d/%d correct\n", stats.QuizCorrect, stats.QuizTotal)
		fmt.Printf("   Time: %s\n", stats.TotalTime.Round(time.Second))
		fmt.Println("════════════════════════════════════════════════════════════════")
	})

	ctrl.Start()

	// Wait for completion (use shorter durations for text demo)
	time.Sleep(time.Duration(len(tutorial.Steps)*100) * time.Millisecond + 500*time.Millisecond)
}

func runAnimation(name string) {
	var animation *widgets.AnimationPreset
	for _, a := range widgets.GetAllAnimations() {
		if a.Name == name {
			animation = &a
			break
		}
	}

	if animation == nil {
		fmt.Printf("Animation '%s' not found.\n", name)
		fmt.Println("Run 'animations' to see available animation names.")
		return
	}

	fmt.Printf("╔══════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║ ANIMATION: %-54s ║\n", animation.Name)
	fmt.Printf("╚══════════════════════════════════════════════════════════════════╝\n\n")

	fmt.Printf("🎬 %s\n\n", animation.Description)
	fmt.Printf("📊 Level: %s | Frames: %d\n\n", animation.Difficulty, len(animation.Frames))

	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("Playing animation (text-only mode)...")
	fmt.Println()

	ctrl := widgets.AnimationPresetToController(*animation)

	ctrl.SetOnFrame(func(frame int, af widgets.AnimationFrame) {
		progress := widgets.ProgressPercentage(frame, len(animation.Frames))
		counter := widgets.FrameCounter(frame, len(animation.Frames))

		fmt.Printf("┌─ %s ────────────────────────────────────── %.0f%%\n", counter, progress)
		fmt.Printf("│ 📌 %s\n", af.Title)
		fmt.Printf("│\n")

		// Word wrap content
		words := splitWords(af.Content)
		line := "│  "
		for _, word := range words {
			if len(line)+len(word) > 70 {
				fmt.Println(line)
				line = "│  " + word
			} else {
				if line == "│  " {
					line += word
				} else {
					line += " " + word
				}
			}
		}
		if line != "│  " {
			fmt.Println(line)
		}

		if af.Highlight != "" {
			fmt.Printf("│\n│  [Highlight: %s]\n", af.Highlight)
		}

		fmt.Println("└────────────────────────────────────────────────────────────────")
		fmt.Println()
	})

	ctrl.Start()

	// Wait for all frames
	totalDuration := time.Duration(0)
	for _, f := range animation.Frames {
		totalDuration += f.Duration
	}
	time.Sleep(totalDuration + 500*time.Millisecond)

	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("🎬 Animation complete!")
	fmt.Println("════════════════════════════════════════════════════════════════")
}

func splitWords(s string) []string {
	var words []string
	word := ""
	for _, c := range s {
		if c == ' ' || c == '\n' {
			if word != "" {
				words = append(words, word)
				word = ""
			}
			if c == '\n' {
				words = append(words, "\n")
			}
		} else {
			word += string(c)
		}
	}
	if word != "" {
		words = append(words, word)
	}
	return words
}
