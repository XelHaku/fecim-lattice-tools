package main

import "testing"

func TestNormalizeEngine_TableDriven(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", "preisach"},
		{"preisach", "preisach"},
		{"P", "preisach"},
		{" lk ", "lk"},
		{"landau", "lk"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		if got := normalizeEngine(tt.in); got != tt.want {
			t.Fatalf("normalizeEngine(%q)=%q want %q", tt.in, got, tt.want)
		}
	}
}

func TestCmdSkip_TableDriven(t *testing.T) {
	tests := []struct {
		args []string
		want int
	}{
		{nil, 0},
		{[]string{}, 0},
		{[]string{"gui"}, 1},
		{[]string{"cli"}, 0},
		{[]string{"gui", "-x"}, 1},
	}

	for _, tt := range tests {
		if got := cmdSkip(tt.args); got != tt.want {
			t.Fatalf("cmdSkip(%v)=%d want %d", tt.args, got, tt.want)
		}
	}
}
