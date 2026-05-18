package main

import "testing"

func TestResearchCommandsForwardExpectedArguments(t *testing.T) {
	var calls [][]string
	previous := researchRunner
	researchRunner = func(args []string) error {
		calls = append(calls, append([]string(nil), args...))
		return nil
	}
	defer func() { researchRunner = previous }()

	commands := [][]string{
		{"research", "ingest"},
		{"research", "index"},
		{"research", "search", "HZO coercive field Preisach"},
	}
	for _, cmd := range commands {
		if err := dispatchSubcommand(cmd); err != nil {
			t.Fatalf("dispatch %v: %v", cmd, err)
		}
	}
	if len(calls) != 3 {
		t.Fatalf("calls=%d want 3", len(calls))
	}
	if calls[0][0] != "ingest" || calls[1][0] != "index" || calls[2][0] != "search" {
		t.Fatalf("unexpected forwarded calls: %#v", calls)
	}
}
