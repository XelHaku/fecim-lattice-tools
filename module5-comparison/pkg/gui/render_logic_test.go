package gui

import (
	"image"
	"image/color"
	"strings"
	"testing"

	fyneTest "fyne.io/fyne/v2/test"
)

func TestWidgetImageGenerators(t *testing.T) {
	e := &EnergyBarChart{}
	e.cpuSpec = EnergySpec{EnergyFJ: 1000}
	e.gpuSpec = EnergySpec{EnergyFJ: 100}
	e.fecimSpec = EnergySpec{EnergyFJ: 1}
	img := e.generateImage(420, 220)
	if img.Bounds().Dx() != 420 || img.Bounds().Dy() != 220 {
		t.Fatalf("unexpected energy chart bounds: %v", img.Bounds())
	}

	a := &ArchitectureDiagram{}
	img2 := a.generateImage(420, 180)
	if img2.Bounds().Dx() != 420 {
		t.Fatalf("unexpected architecture diagram bounds: %v", img2.Bounds())
	}

	rgba := image.NewRGBA(image.Rect(0, 0, 20, 20))
	drawBox(rgba, 2, 2, 10, 10, color.RGBA{200, 100, 50, 255})
	drawBar(rgba, 1, 1, 5, 3, color.RGBA{20, 40, 60, 255})
}

func TestEducationalPanelModeAndPhaseContent(t *testing.T) {
	a := fyneTest.NewApp()
	defer a.Quit()

	p := NewComparisonEducationalPanel()
	p.SetPresentationMode(PresentationModeInvestor)
	if !strings.Contains(p.title, "Investment") {
		t.Fatalf("expected investor title, got %q", p.title)
	}
	p.SetPresentationMode(PresentationModeEngineer)
	if !strings.Contains(p.title, "Technical") {
		t.Fatalf("expected engineer title, got %q", p.title)
	}

	p.SetPhase(AutoDemoPhaseEnergyRace)
	if !strings.Contains(p.title, "Energy") {
		t.Fatalf("expected energy phase title, got %q", p.title)
	}
	p.SetPhase(AutoDemoPhaseMarket)
	if !strings.Contains(p.content, "721B") {
		t.Fatalf("expected market phase content, got %q", p.content)
	}
	p.SetPhase(AutoDemoPhaseCalculator)
	if !strings.Contains(p.content, "Try the calculator") {
		t.Fatalf("expected calculator phase content, got %q", p.content)
	}
}

func TestOperationLogAndComparisonText(t *testing.T) {
	a := fyneTest.NewApp()
	defer a.Quit()

	log := NewComparisonOperationLog()
	for i := 0; i < 12; i++ {
		log.Add("step")
	}
	if len(log.entries) != log.maxEntries {
		t.Fatalf("expected capped entries=%d, got %d", log.maxEntries, len(log.entries))
	}

	p := NewComparisonEducationalPanel()
	p.SetComparison(1000, 100)
	if !strings.Contains(p.content, "1000×") || !strings.Contains(p.content, "100×") {
		t.Fatalf("expected ratio text injected, got %q", p.content)
	}
}
