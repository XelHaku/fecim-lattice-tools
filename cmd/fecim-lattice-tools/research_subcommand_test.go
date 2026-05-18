package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDispatchResearchSubcommandUsesResearchRunner(t *testing.T) {
	var got []string
	previous := researchRunner
	researchRunner = func(args []string) error {
		got = append([]string(nil), args...)
		return nil
	}
	defer func() { researchRunner = previous }()

	if err := dispatchSubcommand([]string{"research", "search", "HZO coercive field"}); err != nil {
		t.Fatalf("dispatch research: %v", err)
	}
	want := []string{"search", "HZO coercive field"}
	if len(got) != len(want) {
		t.Fatalf("research runner args len=%d want=%d args=%v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg %d=%q want %q", i, got[i], want[i])
		}
	}
}

func TestDispatchResearchSubcommandPropagatesRunnerError(t *testing.T) {
	previous := researchRunner
	researchRunner = func(args []string) error {
		return errors.New("research tool failed")
	}
	defer func() { researchRunner = previous }()

	err := dispatchSubcommand([]string{"research", "ingest"})
	if err == nil || !strings.Contains(err.Error(), "research tool failed") {
		t.Fatalf("expected runner error, got %v", err)
	}
}

func TestRootUsageListsResearchSubcommand(t *testing.T) {
	var buf bytes.Buffer
	printRootUsage(&buf)
	text := buf.String()
	if !strings.Contains(text, "research") {
		t.Fatalf("root usage must mention research subcommand:\n%s", text)
	}
	if !strings.Contains(text, "research ingest") {
		t.Fatalf("root usage must include research example:\n%s", text)
	}
}

func TestRunResearchToolFindsRepoRootWhenCalledOutsideRepo(t *testing.T) {
	root, err := filepath.Abs(repoRoot())
	if err != nil {
		t.Fatalf("abs repo root: %v", err)
	}
	fakePython := filepath.Join(t.TempDir(), "fake-python")
	cwdPath := filepath.Join(t.TempDir(), "cwd.txt")
	script := "#!/bin/sh\n" +
		"test -f \"$1\" || exit 17\n" +
		"pwd > \"$FECIM_FAKE_PYTHON_CWD\"\n"
	if err := os.WriteFile(fakePython, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake python: %v", err)
	}
	t.Setenv("FECIM_RESEARCH_PYTHON", fakePython)
	t.Setenv("FECIM_FAKE_PYTHON_CWD", cwdPath)

	previousCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	outside := t.TempDir()
	if err := os.Chdir(outside); err != nil {
		t.Fatalf("chdir outside repo: %v", err)
	}
	defer func() {
		if err := os.Chdir(previousCwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()

	if err := runResearchTool([]string{"--repo-root", root, "--help"}); err != nil {
		t.Fatalf("run research tool outside repo: %v", err)
	}
	cwd, err := os.ReadFile(cwdPath)
	if err != nil {
		t.Fatalf("read fake python cwd: %v", err)
	}
	if strings.TrimSpace(string(cwd)) != root {
		t.Fatalf("research tool cwd = %q, want %q", strings.TrimSpace(string(cwd)), root)
	}
}
