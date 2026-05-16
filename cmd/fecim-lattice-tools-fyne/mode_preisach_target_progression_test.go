package main

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"testing"
	"time"

	"fecim-lattice-tools/shared/logging"
)

func TestHeadlessPreisachRun_WRDTargetProgressionMatchesSequence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Preisach target progression in -short")
	}
	t.Setenv("FECIM_MATERIAL", "literature_superlattice")
	t.Setenv("FECIM_ISPP_TARGET_LEVELS", "3,15,27")

	before := time.Now()
	if err := runMode("hysteresis", "preisach"); err != nil {
		t.Fatalf("runMode: %v", err)
	}

	logPath := newestHysteresisLogAfterWithTB(t, before)
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

	waveIdx := mustCol(t, idx, "waveform")
	targetIdx := mustCol(t, idx, "wrd_target_level")
	phaseIdx := mustCol(t, idx, "wrd_phase_name")

	progression := []int{}
	lastTarget := 0
	for {
		row, err := r.Read()
		if err != nil {
			break
		}
		if row[waveIdx] != "ISPP" {
			continue
		}
		target, err := strconv.Atoi(row[targetIdx])
		if err != nil || target <= 0 {
			continue
		}
		if target != lastTarget {
			phase := row[phaseIdx]
			if len(progression) > 0 && phase != "PREP" && phase != "WRITE" {
				t.Fatalf("target changed %d→%d outside PREP/WRITE (phase=%q) in %s", lastTarget, target, phase, logPath)
			}
			progression = append(progression, target)
			lastTarget = target
		}
	}

	expected := []int{3, 15, 27}
	if len(progression) < len(expected) {
		t.Fatalf("short target progression in %s: got=%v wantPrefix=%v", logPath, progression, expected)
	}
	for i := range expected {
		if progression[i] != expected[i] {
			t.Fatalf("target progression mismatch in %s at step %d: got=%v wantPrefix=%v", logPath, i+1, progression, expected)
		}
	}
}

func newestHysteresisLogAfterWithTB(t *testing.T, after time.Time) string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(logging.LogsDir(), "hysteresis-*.csv"))
	if err != nil {
		t.Fatalf("glob logs: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("no hysteresis CSV logs found")
	}
	sort.Slice(paths, func(i, j int) bool {
		iInfo, _ := os.Stat(paths[i])
		jInfo, _ := os.Stat(paths[j])
		if iInfo == nil || jInfo == nil {
			return paths[i] < paths[j]
		}
		return iInfo.ModTime().After(jInfo.ModTime())
	})
	for _, p := range paths {
		info, err := os.Stat(p)
		if err == nil && info.ModTime().After(after.Add(-100*time.Millisecond)) {
			return p
		}
	}
	return paths[0]
}

func mustCol(t *testing.T, idx map[string]int, name string) int {
	t.Helper()
	v, ok := idx[name]
	if !ok {
		t.Fatalf("missing %s column", name)
	}
	return v
}
