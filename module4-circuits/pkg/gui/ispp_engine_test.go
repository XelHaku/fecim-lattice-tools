package gui

import "testing"

func TestISPPEngine_DefaultAndSetGet(t *testing.T) {
	ds := newTestDeviceState(8, 8)

	if got := ds.GetISPPEngine(); got != ISPPEngineLevel {
		t.Fatalf("default ISPP engine: got %v, want %v", got, ISPPEngineLevel)
	}

	ds.SetISPPEngine(ISPPEngineLK)
	if got := ds.GetISPPEngine(); got != ISPPEngineLK {
		t.Fatalf("set/get ISPP engine: got %v, want %v", got, ISPPEngineLK)
	}

	ds.SetISPPEngine(ISPPEngineLevel)
	if got := ds.GetISPPEngine(); got != ISPPEngineLevel {
		t.Fatalf("set/get ISPP engine: got %v, want %v", got, ISPPEngineLevel)
	}
}

func TestISPPEngine_StringLabels(t *testing.T) {
	tests := []struct {
		engine ISPPEngine
		label  string
	}{
		{ISPPEngineLevel, "Fast (Level)"},
		{ISPPEngineLK, "L-K (Physics)"},
		{ISPPEngine(99), "Unknown"},
	}

	for _, tc := range tests {
		if got := tc.engine.String(); got != tc.label {
			t.Fatalf("engine %v label: got %q, want %q", tc.engine, got, tc.label)
		}
	}
}
