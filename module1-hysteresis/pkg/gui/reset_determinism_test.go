package gui

import "testing"

func TestResetAllStateLocked_IsDeterministic(t *testing.T) {
	a := &App{
		physicsEngine: PhysicsPreisach,
		numLevels:     30,
		maxHistory:    200,
	}

	a.mu.Lock()
	a.polarization = 0.5
	a.normalizedP = 0.5
	a.electricField = 1e6
	a.simTime = 12.34
	for i := 0; i < 100; i++ {
		a.appendHistoryLocked(float64(i), float64(i)/100.0)
	}

	a.resetAllStateLocked()

	if a.polarization != 0 {
		t.Fatalf("polarization not reset: got %g, want 0", a.polarization)
	}
	if a.electricField != 0 {
		t.Fatalf("electricField not reset: got %g, want 0", a.electricField)
	}
	if got := a.historyLengthLocked(); got != 0 {
		t.Fatalf("history not reset: got %d samples, want 0", got)
	}
	if a.simTime != 0 {
		t.Fatalf("simTime not reset: got %g, want 0", a.simTime)
	}
	a.mu.Unlock()
}
