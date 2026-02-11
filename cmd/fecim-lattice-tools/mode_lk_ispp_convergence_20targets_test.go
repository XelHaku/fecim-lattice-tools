package main

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

// Regression acceptance: LK ISPP converges 20 targets without timeout.
//
// This runs the headless LK hysteresis mode with a deterministic 20-target
// list (mix of extremes + mid-levels) and fails on any timeout/non-finite.
func TestHeadlessLKRun_ISPPConverges20Targets_NoTimeout(t *testing.T) {
	// Pick a stable, deterministic preset.
	t.Setenv("FECIM_MATERIAL", "fecim_hzo")

	// Drive 20 targets deterministically.
	t.Setenv("FECIM_ISPP_TARGETS", "20")
	t.Setenv("FECIM_ISPP_TARGET_SEED", "1")

	// Keep CI runtime reasonable while preserving LK dynamics.
	t.Setenv("FECIM_HEADLESS_FAST", "1")

	before := time.Now()
	if err := runMode("hysteresis", "lk"); err != nil {
		t.Fatalf("runMode: %v", err)
	}

	// Find newest hysteresis CSV written after the run.
	paths, err := filepath.Glob(filepath.Join("logs", "hysteresis-*.csv"))
	if err != nil {
		t.Fatalf("glob logs: %v", err)
	}
	if len(paths) == 0 {
		t.Fatalf("no hysteresis CSV logs found")
	}
	sort.Slice(paths, func(i, j int) bool {
		iInfo, _ := os.Stat(paths[i])
		jInfo, _ := os.Stat(paths[j])
		if iInfo == nil || jInfo == nil {
			return paths[i] < paths[j]
		}
		return iInfo.ModTime().After(jInfo.ModTime())
	})

	logPath := ""
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if info.ModTime().After(before.Add(-2 * time.Second)) {
			logPath = p
			break
		}
	}
	if logPath == "" {
		logPath = paths[0]
	}

	f, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("open %s: %v", logPath, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	header, err := r.Read()
	if err != nil {
		t.Fatalf("read header: %v", err)
	}
	idx := map[string]int{}
	for i, h := range header {
		idx[h] = i
	}

	waveIdx, ok := idx["waveform"]
	if !ok {
		t.Fatalf("missing waveform column")
	}
	phaseIdx, ok := idx["wrd_phase_name"]
	if !ok {
		t.Fatalf("missing wrd_phase_name column")
	}

	isppRows := 0
	nonEmptyPhase := 0
	for {
		row, err := r.Read()
		if err != nil {
			break
		}
		if row[waveIdx] != "ISPP" {
			continue
		}
		isppRows++
		if row[phaseIdx] != "" {
			nonEmptyPhase++
		}
	}

	if isppRows == 0 {
		t.Fatalf("expected ISPP rows in %s, got 0", logPath)
	}
	if nonEmptyPhase == 0 {
		t.Fatalf("expected non-empty wrd_phase_name for ISPP rows in %s", logPath)
	}
}
