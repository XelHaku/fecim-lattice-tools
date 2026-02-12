package widgets

import "testing"

func TestPhaseIndicator_CurrentPhaseUpdates(t *testing.T) {
	p := NewPhaseIndicator()
	phase, mode := p.CurrentPhase()
	if phase != -1 || mode != "" {
		t.Fatalf("initial phase=%d mode=%q, want -1 and empty", phase, mode)
	}

	p.SetPhase(PhaseProgram, "wrd")
	phase, mode = p.CurrentPhase()
	if phase != PhaseProgram || mode != "wrd" {
		t.Fatalf("after SetPhase got phase=%d mode=%q, want %d and wrd", phase, mode, PhaseProgram)
	}
}
