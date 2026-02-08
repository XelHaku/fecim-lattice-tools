package help

import (
	"testing"
)

func TestQuickTipsExist(t *testing.T) {
	if len(quickTips) == 0 {
		t.Error("Expected quick tips to be defined")
	}
}

func TestQuickTipsHaveContent(t *testing.T) {
	for i, tip := range quickTips {
		if tip.Title == "" {
			t.Errorf("Tip %d has empty title", i)
		}
		if tip.Content == "" {
			t.Errorf("Tip %d has empty content", i)
		}
		if tip.Icon == nil {
			t.Errorf("Tip %d has nil icon", i)
		}
	}
}

func TestTipManagerGetRandomTip(t *testing.T) {
	// Create a manager without preferences/window (for testing)
	tm := &TipManager{
		tips:        quickTips,
		showOnStart: true,
	}
	
	// Get a random tip
	tip := tm.GetRandomTip()
	if tip.Title == "" {
		t.Error("GetRandomTip returned empty tip")
	}
}

func TestTipManagerGetTipForModule(t *testing.T) {
	tm := &TipManager{
		tips:        quickTips,
		showOnStart: true,
	}
	
	// Get a tip for hysteresis module
	tip := tm.GetTipForModule("hysteresis")
	if tip != nil {
		if tip.Module != "hysteresis" {
			t.Errorf("Expected hysteresis tip, got module '%s'", tip.Module)
		}
	}
	
	// Get a tip for a module that doesn't have specific tips
	tip = tm.GetTipForModule("nonexistent")
	if tip != nil {
		t.Error("Expected nil for nonexistent module")
	}
}

func TestModuleSpecificTipsExist(t *testing.T) {
	modules := map[string]bool{
		"hysteresis": false,
		"crossbar":   false,
		"mnist":      false,
		"comparison": false,
	}
	
	for _, tip := range quickTips {
		if tip.Module != "" {
			modules[tip.Module] = true
		}
	}
	
	for mod, found := range modules {
		if !found {
			t.Errorf("Expected at least one tip for module '%s'", mod)
		}
	}
}
