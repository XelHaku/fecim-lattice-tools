package main

import "testing"

func TestHeadlessEngineRegistryExposesHysteresisPort(t *testing.T) {
	entries := HeadlessEngineRegistry()
	if len(entries) == 0 {
		t.Fatal("HeadlessEngineRegistry returned no entries")
	}

	var found bool
	for _, entry := range entries {
		if entry.Mode != "hysteresis" {
			continue
		}
		found = true
		if entry.Port == nil {
			t.Fatal("hysteresis headless engine port is nil")
		}
	}
	if !found {
		t.Fatal("HeadlessEngineRegistry missing hysteresis entry")
	}
}

func TestRunModeUnknownModeListsRegistryModes(t *testing.T) {
	err := runMode("missing", "")
	if err == nil {
		t.Fatal("runMode missing returned nil error")
	}
	want := `unknown mode "missing" (expected: hysteresis)`
	if err.Error() != want {
		t.Fatalf("runMode missing error = %q, want %q", err.Error(), want)
	}
}
