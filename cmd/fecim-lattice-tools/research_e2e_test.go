package main

import (
	"reflect"
	"testing"
)

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
	want := [][]string{
		{"ingest"},
		{"index"},
		{"search", "HZO coercive field Preisach"},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("forwarded calls = %#v, want %#v", calls, want)
	}
}
