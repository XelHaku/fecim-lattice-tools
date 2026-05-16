package main

import (
	"encoding/csv"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"fecim-lattice-tools/shared/logging"
)

func TestHeadlessLKRun_ISPP5Targets_NoNaNInf_AndEFieldUnitsConsistent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping ISPP 5-target test in -short")
	}
	t.Setenv("FECIM_MATERIAL", "literature_superlattice")
	t.Setenv("FECIM_ISPP_TARGETS", "5")
	t.Setenv("FECIM_ISPP_TARGET_SEED", "1")
	t.Setenv("FECIM_HEADLESS_FAST", "1")

	before := time.Now()
	if err := runMode("hysteresis", "lk"); err != nil {
		t.Fatalf("runMode: %v", err)
	}

	logPath := newestHysteresisCSVAfter(t, before)
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

	requireCol := func(name string) int {
		i, ok := idx[name]
		if !ok {
			t.Fatalf("missing %s column in %s", name, logPath)
		}
		return i
	}

	waveIdx := requireCol("waveform")
	targetIdx := requireCol("wrd_target_level")
	eFieldVmIdx := requireCol("e_field_v_m")
	eFieldMVIdx := requireCol("e_field_mv_cm")
	ctrlFieldVmIdx := requireCol("controller_current_field_v_m")
	ctrlFieldMVIdx := requireCol("controller_current_field_mv_cm")
	ecMVIdx := requireCol("ec_mv_cm")

	const mvPerCM = 1e8
	const relTol = 1e-9
	const absTol = 1e-12

	isppRows := 0
	seenTargets := map[int]bool{}

	for {
		row, err := r.Read()
		if err != nil {
			break
		}

		for i, cell := range row {
			if strings.EqualFold(cell, "nan") || strings.EqualFold(cell, "+inf") || strings.EqualFold(cell, "-inf") || strings.EqualFold(cell, "inf") {
				t.Fatalf("non-finite token %q in row (col=%s) file=%s", cell, header[i], logPath)
			}
		}

		if row[waveIdx] != "ISPP" {
			continue
		}
		isppRows++

		target, err := strconv.Atoi(row[targetIdx])
		if err == nil && target > 0 {
			seenTargets[target] = true
		}

		eVM := mustParseFloat(t, row[eFieldVmIdx], "e_field_v_m")
		eMV := mustParseFloat(t, row[eFieldMVIdx], "e_field_mv_cm")
		assertFloatClose(t, eMV, eVM/mvPerCM, relTol, absTol, "e_field_mv_cm vs e_field_v_m")

		ctrlVM := mustParseFloat(t, row[ctrlFieldVmIdx], "controller_current_field_v_m")
		ctrlMV := mustParseFloat(t, row[ctrlFieldMVIdx], "controller_current_field_mv_cm")
		assertFloatClose(t, ctrlMV, ctrlVM/mvPerCM, relTol, absTol, "controller_current_field_mv_cm vs controller_current_field_v_m")

		ecMV := mustParseFloat(t, row[ecMVIdx], "ec_mv_cm")
		if ecMV <= 0 {
			t.Fatalf("ec_mv_cm must be >0, got %g", ecMV)
		}
	}

	if isppRows == 0 {
		t.Fatalf("expected ISPP rows in %s, got 0", logPath)
	}
	if len(seenTargets) < 5 {
		t.Fatalf("expected >=5 distinct ISPP targets, got %d (%v) in %s", len(seenTargets), sortedIntKeys(seenTargets), logPath)
	}
}

func newestHysteresisCSVAfter(t *testing.T, before time.Time) string {
	t.Helper()
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

	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if info.ModTime().After(before.Add(-100 * time.Millisecond)) {
			return p
		}
	}
	return paths[0]
}

func mustParseFloat(t *testing.T, s string, label string) float64 {
	t.Helper()
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		t.Fatalf("parse %s=%q: %v", label, s, err)
	}
	if math.IsNaN(v) || math.IsInf(v, 0) {
		t.Fatalf("non-finite parsed %s=%v", label, v)
	}
	return v
}

func assertFloatClose(t *testing.T, got, want, relTol, absTol float64, label string) {
	t.Helper()
	delta := math.Abs(got - want)
	limit := math.Max(absTol, relTol*math.Max(math.Abs(got), math.Abs(want)))
	if delta > limit {
		t.Fatalf("%s mismatch: got=%g want=%g delta=%g limit=%g", label, got, want, delta, limit)
	}
}

func sortedIntKeys(m map[int]bool) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}
