package main

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"fecim-lattice-tools/shared/logging"
)

// TestHeadlessLKRun_EmitsISPPPhases ensures the LK headless run actually emits
// ISPP rows/phases in the CSV log. This is a lightweight integration check that
// the ISPP write/read demo path is exercised under the LK engine.
func TestHeadlessLKRun_EmitsISPPPhases(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping LK ISPP phases test in -short")
	}
	t.Setenv("FECIM_MATERIAL", "literature_superlattice")

	before := time.Now()
	if err := runMode("hysteresis", "lk"); err != nil {
		t.Fatalf("runMode: %v", err)
	}

	// Find newest hysteresis CSV written after the run.
	paths, err := filepath.Glob(filepath.Join(logging.LogsDir(), "hysteresis-*.csv"))
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

	var logPath string
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
		// fall back to newest
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
	seenPhase := map[string]bool{}
	for {
		row, err := r.Read()
		if err != nil {
			break
		}
		if row[waveIdx] != "ISPP" {
			continue
		}
		isppRows++
		p := row[phaseIdx]
		if p != "" {
			seenPhase[p] = true
		}
	}

	if isppRows == 0 {
		t.Fatalf("expected ISPP rows in %s, got 0", logPath)
	}
	// We expect the WRD demo state machine to at least enter PREP or WRITE.
	if !seenPhase["PREP"] && !seenPhase["WRITE"] {
		t.Fatalf("expected to see PREP or WRITE phases in ISPP rows (got phases=%v) in %s", keys(seenPhase), logPath)
	}
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
