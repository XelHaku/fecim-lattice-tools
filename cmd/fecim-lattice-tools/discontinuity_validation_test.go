package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
	"testing"
	"time"
)

type discontinuity struct {
	Row        int
	Phase      string
	Target     int
	E          float64
	P          float64
	DeltaP     float64
	Ratio      float64
	Class      string
	Reason     string
	NearEcFrac float64
}

type continuitySummary struct {
	Engine       string
	LogPath      string
	TotalPoints  int
	RangeP       float64
	Flagged      []discontinuity
	Physical     int
	Spurious     int
	HeaderByName map[string]int
}

func TestHeadlessISPPContinuityValidation_PreisachVsLK(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping continuity validation in -short")
	}

	preisach := runAndAnalyzeDiscontinuities(t, "preisach")
	lk := runAndAnalyzeDiscontinuities(t, "lk")

	t.Logf("PREISACH summary: points=%d p_range=%0.9e flagged=%d physical=%d spurious=%d log=%s",
		preisach.TotalPoints, preisach.RangeP, len(preisach.Flagged), preisach.Physical, preisach.Spurious, preisach.LogPath)
	t.Logf("LK summary: points=%d p_range=%0.9e flagged=%d physical=%d spurious=%d log=%s",
		lk.TotalPoints, lk.RangeP, len(lk.Flagged), lk.Physical, lk.Spurious, lk.LogPath)

	if preisach.Spurious > 0 {
		t.Fatalf("Preisach spurious discontinuities detected: %d", preisach.Spurious)
	}

	for i, d := range preisach.Flagged {
		if d.Class != "PHYSICAL" {
			t.Fatalf("Preisach discontinuity #%d classified non-physical: row=%d phase=%s target=%d E=%0.9e P=%0.9e deltaP=%0.9e ratio=%0.6f nearEcFrac=%0.6f class=%s reason=%s",
				i+1, d.Row, d.Phase, d.Target, d.E, d.P, d.DeltaP, d.Ratio, d.NearEcFrac, d.Class, d.Reason)
		}
	}

	if len(preisach.Flagged) == 0 {
		t.Log("PREISACH discontinuity list is empty under current headless ISPP settings")
	}
	if len(lk.Flagged) == 0 {
		t.Log("LK discontinuity list is empty under current headless ISPP settings")
	}
}

func runAndAnalyzeDiscontinuities(t *testing.T, engine string) continuitySummary {
	t.Helper()
	t.Setenv("FECIM_MATERIAL", "literature_superlattice")
	t.Setenv("FECIM_ISPP_TARGET_LEVELS", "3,15,27")
	t.Setenv("FECIM_HEADLESS_FAST", "1")

	before := time.Now()
	if err := runMode("hysteresis", engine); err != nil {
		t.Fatalf("runMode(%s): %v", engine, err)
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

	require := func(name string) int {
		i, ok := idx[name]
		if !ok {
			t.Fatalf("missing %s in %s", name, logPath)
		}
		return i
	}

	waveIdx := require("waveform")
	phaseIdx := require("wrd_phase_name")
	targetIdx := require("wrd_target_level")
	eIdx := require("e_field_v_m")
	pIdx := require("polarization_c_m2")
	ecIdx := require("ec_mv_cm")

	type point struct {
		row    int
		phase  string
		target int
		e      float64
		p      float64
		ecVpm  float64
	}
	points := make([]point, 0, 1024)
	rowNum := 1
	pMin := math.Inf(1)
	pMax := math.Inf(-1)

	for {
		row, err := r.Read()
		if err != nil {
			break
		}
		rowNum++
		if row[waveIdx] != "ISPP" {
			continue
		}
		e, err := strconv.ParseFloat(row[eIdx], 64)
		if err != nil {
			t.Fatalf("parse e_field_v_m row=%d: %v", rowNum, err)
		}
		p, err := strconv.ParseFloat(row[pIdx], 64)
		if err != nil {
			t.Fatalf("parse polarization_c_m2 row=%d: %v", rowNum, err)
		}
		ecMVcm, err := strconv.ParseFloat(row[ecIdx], 64)
		if err != nil {
			t.Fatalf("parse ec_mv_cm row=%d: %v", rowNum, err)
		}
		target, _ := strconv.Atoi(row[targetIdx])
		ecVpm := ecMVcm * 1e8
		points = append(points, point{row: rowNum, phase: row[phaseIdx], target: target, e: e, p: p, ecVpm: ecVpm})
		if p < pMin {
			pMin = p
		}
		if p > pMax {
			pMax = p
		}
	}

	if len(points) < 2 {
		t.Fatalf("insufficient ISPP points for %s in %s: %d", engine, logPath, len(points))
	}

	pRange := pMax - pMin
	if pRange <= 0 {
		t.Fatalf("invalid p_range for %s in %s: %0.9e", engine, logPath, pRange)
	}

	summary := continuitySummary{Engine: engine, LogPath: logPath, TotalPoints: len(points), RangeP: pRange, HeaderByName: idx}
	for i := 1; i < len(points); i++ {
		prev := points[i-1]
		cur := points[i]
		if cur.phase != prev.phase {
			continue
		}
		deltaP := math.Abs(cur.p - prev.p)
		ratio := deltaP / pRange
		if ratio <= 0.10 {
			continue
		}

		class, reason, nearEcFrac := classifyDiscontinuity(prev.e, cur.e, cur.ecVpm, cur.phase)
		d := discontinuity{
			Row:        cur.row,
			Phase:      cur.phase,
			Target:     cur.target,
			E:          cur.e,
			P:          cur.p,
			DeltaP:     deltaP,
			Ratio:      ratio,
			Class:      class,
			Reason:     reason,
			NearEcFrac: nearEcFrac,
		}
		summary.Flagged = append(summary.Flagged, d)
		if class == "PHYSICAL" {
			summary.Physical++
		} else {
			summary.Spurious++
		}
		t.Logf("%s DISCONTINUITY row=%d phase=%s target=%d E=%0.9e P=%0.9e deltaP=%0.9e ratio=%0.6f class=%s nearEcFrac=%0.6f reason=%s",
			engine, d.Row, d.Phase, d.Target, d.E, d.P, d.DeltaP, d.Ratio, d.Class, d.NearEcFrac, d.Reason)
	}

	return summary
}

func classifyDiscontinuity(prevE, curE, ecVpm float64, phase string) (class, reason string, nearEcFrac float64) {
	if ecVpm <= 0 {
		return "SPURIOUS", "invalid Ec", 1.0
	}
	nearEcFrac = math.Abs(math.Abs(curE)-ecVpm) / ecVpm

	// In PROG_VERIFY the controller intentionally reverses polarity during
	// overshoot recovery. Large |ΔP| on a same-phase polarity flip is expected.
	if prevE != 0 && curE != 0 && (prevE > 0) != (curE > 0) {
		return "PHYSICAL", "same-phase polarity reversal during verify/reset", nearEcFrac
	}
	if nearEcFrac <= 0.25 {
		return "PHYSICAL", fmt.Sprintf("|E| close to Ec (| |E|-Ec |/Ec=%0.6f)", nearEcFrac), nearEcFrac
	}
	if phase == "WRITE" && math.Abs(curE) >= 0.85*ecVpm {
		return "PHYSICAL", "high-field WRITE switching region", nearEcFrac
	}
	if phase == "PREP" && math.Abs(curE) >= 1.5*ecVpm {
		return "PHYSICAL", "high-field preparation reset", nearEcFrac
	}
	if phase == "PROG_VERIFY" && math.Abs(curE) >= 0.45*ecVpm {
		return "PHYSICAL", "minor-loop wipe-out threshold crossing in verify", nearEcFrac
	}
	return "SPURIOUS", fmt.Sprintf("large jump away from switching threshold (| |E|-Ec |/Ec=%0.6f)", nearEcFrac), nearEcFrac
}
