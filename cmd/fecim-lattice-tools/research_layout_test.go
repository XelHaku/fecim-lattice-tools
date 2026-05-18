package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResearchLedgerLayoutExists(t *testing.T) {
	root := repoRoot()
	requiredDirs := []string{
		"research/papers",
		"research/sources",
		"research/parsed",
		"research/chunks",
		"research/extracted",
		"research/graphs",
		"research/manifests",
		"research/reports",
		"research/index",
		"citations/claims",
	}
	for _, dir := range requiredDirs {
		info, err := os.Stat(filepath.Join(root, dir))
		if err != nil {
			t.Fatalf("expected %s to exist: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected %s to be a directory", dir)
		}
	}
}

func TestResearchGitignoreKeepsLedgerAndIgnoresCaches(t *testing.T) {
	root := repoRoot()
	body, err := os.ReadFile(filepath.Join(root, "research/.gitignore"))
	if err != nil {
		t.Fatalf("read research/.gitignore: %v", err)
	}
	text := string(body)
	required := []string{
		"/index/pyserini/",
		"/index/lancedb/",
		"/index/models/",
		"/.cache/",
		"!/index/.gitkeep",
	}
	for _, phrase := range required {
		if !strings.Contains(text, phrase) {
			t.Fatalf("research/.gitignore must contain %q", phrase)
		}
	}
}
