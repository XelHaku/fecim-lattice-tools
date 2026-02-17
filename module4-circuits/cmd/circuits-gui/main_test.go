package circuitsgui

import "testing"

func TestIsVerbosityToken(t *testing.T) {
	valid := []string{"0", "off", "1", "info", "2", "debug", "3", "trace", "all"}
	for _, v := range valid {
		if !isVerbosityToken(v) {
			t.Fatalf("expected token %q to be valid", v)
		}
	}
	invalid := []string{"", "verbose", "5", "banana"}
	for _, v := range invalid {
		if isVerbosityToken(v) {
			t.Fatalf("expected token %q to be invalid", v)
		}
	}
}
