//go:build legacy_fyne

package gui

import (
	"testing"
	"time"
)

func TestFormattingHelpers(t *testing.T) {
	if got := formatHeroMoney(2.3e9); got != "$2.3B" {
		t.Fatalf("unexpected billions format: %q", got)
	}
	if got := formatHeroMoney(2.3e6); got != "$2M" {
		t.Fatalf("unexpected millions format: %q", got)
	}
	if got := formatHeroMoney(2.3e3); got != "$2K" {
		t.Fatalf("unexpected thousands format: %q", got)
	}

	if got := formatNumberWithSuffix(1.2e12); got != "1.2T" {
		t.Fatalf("unexpected suffix format: %q", got)
	}
	if got := formatCost(0.00001); got == "$0.00" {
		t.Fatalf("expected scientific notation path, got %q", got)
	}
	if got := formatPower(0.000002); got != "2.0 µW" {
		t.Fatalf("unexpected power format: %q", got)
	}
}

func TestComparisonModesAndPhases(t *testing.T) {
	if ComparisonModeIdle.String() != "IDLE" || ComparisonMode(99).String() != "UNKNOWN" {
		t.Fatal("comparison mode String mapping failed")
	}
	if PresentationModeFromString("Engineer") != PresentationModeEngineer {
		t.Fatal("presentation mode parse failed")
	}
	if PresentationModeFromString("???") != PresentationModeManual {
		t.Fatal("unknown presentation mode should default to manual")
	}

	if AutoDemoPhaseCalculator.PhaseDuration() != 15*time.Second {
		t.Fatal("calculator phase should be 15s")
	}
	if AutoDemoPhase(99).PhaseDuration() != 10*time.Second {
		t.Fatal("unknown phase should fallback to 10s")
	}
	if AutoDemoPhaseCompetitive.String() != "Competitive Matrix" {
		t.Fatal("phase string mapping failed")
	}
}

func TestWorkloadAndSavingsCalculations(t *testing.T) {
	if got := annualSavingsForScale(10, 2); got != 960000 {
		t.Fatalf("unexpected annual savings: %.0f", got)
	}

	app := &ComparisonApp{currentWorkload: "MNIST"}
	if got := app.getWorkloadMACs(); got != 784*128+128*10 {
		t.Fatalf("unexpected MNIST MACs: %d", got)
	}
	app.currentWorkload = "Unknown"
	if got := app.getWorkloadMACs(); got != 101632 {
		t.Fatalf("unknown workload should use default 101632, got %d", got)
	}

	energy, power, cost := energyPowerCost(1000, 100, 10)
	if energy <= 0 || power <= 0 || cost <= 0 {
		t.Fatalf("expected positive outputs, got energy=%g power=%g cost=%g", energy, power, cost)
	}
}
