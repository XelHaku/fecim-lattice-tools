//go:build legacy_fyne

package gui

import (
	"regexp"
	"testing"
)

func TestFormatReadStatusLine_PrecisionAndUnits(t *testing.T) {
	got := formatReadStatusLine(2, 3, 11, -1.236, 0.04789, 127)

	patterns := []string{
		`I=-1\.24 µA`,
		`TIA=\+0\.05 V`,
		`ADC=127`,
	}
	for _, p := range patterns {
		if !regexp.MustCompile(p).MatchString(got) {
			t.Fatalf("status line %q does not match %q", got, p)
		}
	}
}
