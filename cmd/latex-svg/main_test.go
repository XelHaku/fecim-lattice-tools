package main

import (
	"strings"
	"testing"
)

func TestIsFullDocument(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty", "", false},
		{"snippet", "\\rho_{eff} dP/dt = 0", false},
		{"documentclass", "\\documentclass{article}\n\\begin{document}\n", true},
		{"begin_document", "\\begin{document}\nX\\end{document}", true},
	}

	for _, c := range cases {
		if got := isFullDocument(c.input); got != c.want {
			t.Fatalf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}

func TestWrapLatex(t *testing.T) {
	body := "E = mc^2"
	out := wrapLatex(body, "", false)
	if !containsAll(out, []string{"\\documentclass", "\\begin{document}", "\\[", body, "\\]", "\\end{document}"}) {
		t.Fatalf("wrapLatex missing expected content")
	}

	inline := wrapLatex(body, "", true)
	if !containsAll(inline, []string{"\\(", body, "\\)"}) {
		t.Fatalf("wrapLatex inline missing expected content")
	}
}

func containsAll(s string, parts []string) bool {
	for _, p := range parts {
		if !strings.Contains(s, p) {
			return false
		}
	}
	return true
}
