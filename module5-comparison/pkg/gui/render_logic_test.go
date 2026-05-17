//go:build legacy_fyne

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
	p.SetPresentationMode(PresentationModeBriefing)
	title, _ := p.GetContent()
	if !strings.Contains(title, "Scenario") {
		t.Fatalf("expected briefing title, got %q", title)
	}
	p.SetPresentationMode(PresentationModeEngineer)
	title, _ = p.GetContent()
	if !strings.Contains(title, "Technical") {
		t.Fatalf("expected engineer title, got %q", title)
	}

	p.SetPhase(AutoDemoPhaseEnergyRace)
	title, _ = p.GetContent()
	if !strings.Contains(title, "Energy") {
		t.Fatalf("expected energy phase title, got %q", title)
	}
	p.SetPhase(AutoDemoPhaseMarket)
	_, content := p.GetContent()
	if !strings.Contains(content, "721B") {
		t.Fatalf("expected market phase content, got %q", content)
	}
	p.SetPhase(AutoDemoPhaseCalculator)
	_, content = p.GetContent()
	if !strings.Contains(content, "Try the calculator") {
		t.Fatalf("expected calculator phase content, got %q", content)
	}
}

func TestOperationLogAndComparisonText(t *testing.T) {
	a := fyneTest.NewApp()
	defer a.Quit()

	log := newComparisonOperationLog()
	for i := 0; i < 12; i++ {
		log.Add("step")
	}
	entries := log.GetEntries()
	maxEntries := log.GetMaxEntries()
	if len(entries) != maxEntries {
		t.Fatalf("expected capped entries=%d, got %d", maxEntries, len(entries))
	}

	p := NewComparisonEducationalPanel()
	p.SetComparison(1000, 100)
	_, content := p.GetContent()
	if !strings.Contains(content, "1000×") || !strings.Contains(content, "100×") {
		t.Fatalf("expected ratio text injected, got %q", content)
	}
}
