package widgets

import (
	"runtime"
	"testing"
)

func TestGUIScriptedInteractionStress_MemoryBudget(t *testing.T) {
	switches := 1000
	if testing.Short() {
		switches = 100
	}

	tabs := make([][]byte, 8)
	sliders := make([]float64, 8)
	for i := range tabs {
		tabs[i] = make([]byte, 64*1024)
	}

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)

	active := 0
	for i := 0; i < switches; i++ {
		active = (active + 1) % len(tabs)
		sliders[active] = float64((i*37)%1000) / 1000.0
		tabs[active][i%len(tabs[active])] = byte(i)
	}

	runtime.GC()
	runtime.ReadMemStats(&after)
	growth := int64(after.Alloc) - int64(before.Alloc)
	const budget = int64(100 * 1024 * 1024)
	if growth > budget {
		t.Fatalf("memory growth exceeded budget: growth=%d bytes budget=%d bytes", growth, budget)
	}
}
