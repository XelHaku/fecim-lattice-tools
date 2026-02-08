package widgets

import (
	"errors"
	"testing"
	"time"

	"fecim-lattice-tools/shared/validation"
)

func TestToolStatusText(t *testing.T) {
	got := ToolStatusText(validation.StatusInstalled, "CrossSim", ToolStatusSymbolOnly)
	if got != "✓" {
		t.Fatalf("symbol-only: got %q, want %q", got, "✓")
	}

	got = ToolStatusText(validation.StatusNotInstalled, "CrossSim", ToolStatusSymbolWithName)
	if got != "✗ CrossSim" {
		t.Fatalf("with-name: got %q, want %q", got, "✗ CrossSim")
	}
}

func TestToolBusyText(t *testing.T) {
	got := ToolBusyText("BadCrossbar", ToolStatusSymbolOnly)
	if got != "..." {
		t.Fatalf("symbol-only: got %q, want %q", got, "...")
	}

	got = ToolBusyText("BadCrossbar", ToolStatusSymbolWithName)
	if got != "... BadCrossbar" {
		t.Fatalf("with-name: got %q, want %q", got, "... BadCrossbar")
	}
}

func TestFormatToolValidationResult(t *testing.T) {
	pass := &validation.ValidationResult{Tool: "CrossSim", Passed: true, Elapsed: 10 * time.Millisecond}
	if got := FormatToolValidationResult(pass, ToolMessageUnicode); got != "✓ PASS" {
		t.Fatalf("unicode pass: got %q", got)
	}
	if got := FormatToolValidationResult(pass, ToolMessageASCII); got != "V PASS" {
		t.Fatalf("ascii pass: got %q", got)
	}

	notInstalled := &validation.ValidationResult{Tool: "CrossSim", Passed: false, Error: errors.New("CrossSim not installed"), Elapsed: 10 * time.Millisecond}
	if got := FormatToolValidationResult(notInstalled, ToolMessageUnicode); got != "○ NOT INSTALLED" {
		t.Fatalf("unicode not-installed: got %q", got)
	}
	if got := FormatToolValidationResult(notInstalled, ToolMessageUnicodeSkip); got != "○ SKIP (not installed)" {
		t.Fatalf("unicode-skip not-installed: got %q", got)
	}
	if got := FormatToolValidationResult(notInstalled, ToolMessageASCII); got != "O NOT INSTALLED" {
		t.Fatalf("ascii not-installed: got %q", got)
	}

	fail := &validation.ValidationResult{Tool: "CrossSim", Passed: false, Error: errors.New("something else"), Elapsed: 10 * time.Millisecond}
	if got := FormatToolValidationResult(fail, ToolMessageUnicode); got != "✗ FAIL" {
		t.Fatalf("unicode fail: got %q", got)
	}
	if got := FormatToolValidationResult(fail, ToolMessageASCII); got != "X FAIL" {
		t.Fatalf("ascii fail: got %q", got)
	}

	if got := FormatToolValidationResult(nil, ToolMessageUnicode); got != "✗ FAIL" {
		t.Fatalf("nil result: got %q", got)
	}
}
