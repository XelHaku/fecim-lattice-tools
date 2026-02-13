package physics

import "testing"

func TestAgingEngineWakeupFatigueRetention(t *testing.T) {
	eng := NewAgingEngine(0.24)

	prCycle1 := eng.ApplyCycle(1, 0)
	prCycle500 := eng.ApplyCycle(500, 0)
	prCycle1000 := eng.ApplyCycle(1000, 0)
	prCycle100k := eng.ApplyCycle(100000, 0)
	prRetention := eng.ApplyCycle(1000, 3600*24*120) // 120 days hold

	if prCycle500 <= prCycle1 || prCycle1000 <= prCycle1 {
		t.Fatalf("wake-up failed: cycle1=%.6f cycle500=%.6f cycle1000=%.6f", prCycle1, prCycle500, prCycle1000)
	}
	if prCycle100k >= prCycle1000 {
		t.Fatalf("fatigue failed: cycle100k=%.6f should be < cycle1000=%.6f", prCycle100k, prCycle1000)
	}
	if prRetention >= prCycle1000 {
		t.Fatalf("retention failed: retained=%.6f should be < fresh@1000=%.6f", prRetention, prCycle1000)
	}
	if len(eng.CycleHistory) < 5 {
		t.Fatalf("expected cycle history entries, got %d", len(eng.CycleHistory))
	}
}
