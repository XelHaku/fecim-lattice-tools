package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunMode_HysteresisEngineSelectorAliases(t *testing.T) {
	t.Setenv("FECIM_MATERIAL", "fecim_hzo")
	t.Setenv("FECIM_HEADLESS_FAST", "1")
	t.Setenv("FECIM_ISPP_TARGET_LEVELS", "mid")

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
				content := strings.ToLower(string(b))
				if strings.Contains(content, "nan") || strings.Contains(content, "+inf") || strings.Contains(content, "-inf") {
					t.Fatalf("non-finite value found in log %s for %s/%s", logPath, engine, material)
				}
			})
		}
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
