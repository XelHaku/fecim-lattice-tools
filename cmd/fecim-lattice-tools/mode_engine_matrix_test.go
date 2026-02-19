package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestRunMode_HysteresisEngineSelectorAliases(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping engine alias matrix in -short")
	}
	t.Setenv("FECIM_MATERIAL", "fecim_hzo")
	t.Setenv("FECIM_HEADLESS_FAST", "1")
	t.Setenv("FECIM_ISPP_TARGET_LEVELS", "mid")
	t.Setenv("FECIM_ISPP_MAX_PULSES", "120")
	t.Setenv("FECIM_HEADLESS_ALLOW_TIMEOUT", "1")

	for _, engine := range []string{"preisach", "p", "lk", "landau", " l-k "} {
		if err := runMode("hysteresis", engine); err != nil {
			t.Fatalf("engine alias %q should work, got err=%v", engine, err)
		}
	}

	if err := runMode("hysteresis", "definitely-invalid"); err == nil {
		t.Fatalf("expected invalid engine to fail")
	}
}

func TestHeadlessHysteresis_VerificationMatrix_NoNaNOrCrash(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping matrix test in -short")
	}

	engines := []string{"preisach", "lk"}
	materials := []string{
		"fecim_hzo",
		"fecim_hzo_target",
		"default_hzo",
		"literature_superlattice",
		"cryogenic_hzo",
		"hzo_standard_32",
		"hzo_ftj_140",
		"hzo_custom_14",
		"alscn",
	}

	for _, engine := range engines {
		for _, material := range materials {
			engine := engine
			material := material
			name := engine + "/" + material
			t.Run(name, func(t *testing.T) {
				before := time.Now()
				t.Setenv("FECIM_MATERIAL", material)
				t.Setenv("FECIM_HEADLESS_FAST", "1")
				t.Setenv("FECIM_ISPP_TARGET_LEVELS", "lo,mid,hi")
				t.Setenv("FECIM_RANGE_FRAC", "1")
				t.Setenv("FECIM_HEADLESS_ALLOW_TIMEOUT", "1")
				if err := runMode("hysteresis", engine); err != nil {
					t.Fatalf("runMode failed for %s/%s: %v", engine, material, err)
				}

				logPath, err := newestHysteresisLogAfter(before)
				if err != nil {
					t.Fatalf("unable to find log after run (%s/%s): %v", engine, material, err)
				}
				b, err := os.ReadFile(logPath)
				if err != nil {
					t.Fatalf("read log %s: %v", logPath, err)
				}
				if hasNonFiniteCSVValue(b) {
					t.Fatalf("non-finite value found in log %s for %s/%s", logPath, engine, material)
				}
			})
		}
	}
}

func TestHeadlessLK_ISPPConvergenceMatrix_AllMaterials_LO_MID_HI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping convergence matrix test in -short")
	}

	materials := []string{
		"fecim_hzo",
		"fecim_hzo_target",
		"default_hzo",
		"literature_superlattice",
		"cryogenic_hzo",
		"hzo_standard_32",
		"hzo_ftj_140",
		"hzo_custom_14",
		"alscn",
	}

	for _, material := range materials {
		material := material
		t.Run(material, func(t *testing.T) {
			before := time.Now()
			t.Setenv("FECIM_MATERIAL", material)
			t.Setenv("FECIM_HEADLESS_FAST", "1")
			t.Setenv("FECIM_ISPP_TARGET_LEVELS", "lo,mid,hi")
			t.Setenv("FECIM_RANGE_FRAC", "1")
			// FOCUS-77: mid-range regressions should converge, not time out.
			t.Setenv("FECIM_HEADLESS_ALLOW_TIMEOUT", "0")

			if err := runMode("hysteresis", "lk"); err != nil {
				t.Fatalf("runMode lk failed for %s: %v", material, err)
			}

			logPath, err := newestHysteresisLogAfter(before)
			if err != nil {
				t.Fatalf("unable to find log after run (%s): %v", material, err)
			}

			finalSuccessWrites, finalTotalWrites, err := readFinalWriteCounters(logPath)
			if err != nil {
				t.Fatalf("failed to parse write counters in %s: %v", logPath, err)
			}

			const expectedTargets = 3
			if finalSuccessWrites != expectedTargets || finalTotalWrites != expectedTargets {
				t.Fatalf("ISPP convergence failure for %s: success=%d total=%d expected=%d (log=%s)",
					material, finalSuccessWrites, finalTotalWrites, expectedTargets, logPath)
			}
		})
	}
}

func newestHysteresisLogAfter(after time.Time) (string, error) {
	paths, err := filepath.Glob(filepath.Join("logs", "hysteresis-*.csv"))
	if err != nil {
		return "", err
	}
	var newestPath string
	var newestTime time.Time
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if info.ModTime().Before(after.Add(-2 * time.Second)) {
			continue
		}
		if newestPath == "" || info.ModTime().After(newestTime) {
			newestPath = p
			newestTime = info.ModTime()
		}
	}
	if newestPath == "" {
		return "", os.ErrNotExist
	}
	return newestPath, nil
}

func readFinalWriteCounters(logPath string) (successWrites int, totalWrites int, err error) {
	f, err := os.Open(logPath)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	header, err := r.Read()
	if err != nil {
		return 0, 0, err
	}
	idxSuccess := -1
	idxTotal := -1
	for i, h := range header {
		switch h {
		case "wrd_success_writes":
			idxSuccess = i
		case "wrd_total_writes":
			idxTotal = i
		}
	}
	if idxSuccess < 0 || idxTotal < 0 {
		return 0, 0, os.ErrInvalid
	}

	lastSuccess := "0"
	lastTotal := "0"
	for {
		row, readErr := r.Read()
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			// Some historical logs end with a partially-written final line.
			// Keep the last valid counters instead of failing the whole regression.
			if !errors.Is(readErr, csv.ErrFieldCount) {
				return 0, 0, readErr
			}
		}
		if idxSuccess < len(row) {
			lastSuccess = row[idxSuccess]
		}
		if idxTotal < len(row) {
			lastTotal = row[idxTotal]
		}
	}

	sw, err := strconv.Atoi(strings.TrimSpace(lastSuccess))
	if err != nil {
		return 0, 0, err
	}
	tw, err := strconv.Atoi(strings.TrimSpace(lastTotal))
	if err != nil {
		return 0, 0, err
	}
	return sw, tw, nil
}

// nonFiniteCSVRe matches standalone NaN, +Inf, -Inf as CSV field values
// (delimited by commas or line boundaries). This avoids false positives from
// material names like "nanolaminate" that contain "nan" as a substring.
var nonFiniteCSVRe = regexp.MustCompile(`(?i)(^|,)(NaN|\+Inf|-Inf)(,|$)`)

// hasNonFiniteCSVValue checks whether any CSV field in the raw bytes contains
// a non-finite floating-point value (NaN, +Inf, -Inf).
func hasNonFiniteCSVValue(data []byte) bool {
	for _, line := range bytes.Split(data, []byte("\n")) {
		if nonFiniteCSVRe.Match(line) {
			return true
		}
	}
	return false
}
