package viewmodel

import "testing"

func TestPayloadStringParsesRequiredString(t *testing.T) {
	got, err := PayloadString(map[string]string{"mode": "write"}, "mode")
	if err != nil {
		t.Fatalf("PayloadString returned error: %v", err)
	}
	if got != "write" {
		t.Fatalf("PayloadString = %q, want write", got)
	}
}

func TestPayloadStringRejectsMissingString(t *testing.T) {
	if _, err := PayloadString(map[string]string{}, "mode"); err == nil {
		t.Fatal("expected missing key error")
	}
}

func TestPayloadStringInValidatesAllowedValues(t *testing.T) {
	if !PayloadStringIn("WRITE", "READ", "WRITE", "COMPUTE") {
		t.Fatal("expected WRITE to be accepted")
	}
	if PayloadStringIn("ERASE", "READ", "WRITE", "COMPUTE") {
		t.Fatal("expected ERASE to be rejected")
	}
}

func TestPayloadFloatParsesRequiredFloat(t *testing.T) {
	got, err := PayloadFloat(map[string]string{"gain": "1.25"}, "gain")
	if err != nil {
		t.Fatalf("PayloadFloat returned error: %v", err)
	}
	if got != 1.25 {
		t.Fatalf("PayloadFloat = %.3f, want 1.25", got)
	}
}

func TestPayloadFloatRejectsMissingAndInvalidFloat(t *testing.T) {
	if _, err := PayloadFloat(map[string]string{}, "gain"); err == nil {
		t.Fatal("expected missing key error")
	}
	if _, err := PayloadFloat(map[string]string{"gain": "large"}, "gain"); err == nil {
		t.Fatal("expected invalid float error")
	}
}

func TestOptionalPayloadFloatKeepsDefaultWhenMissing(t *testing.T) {
	got, err := OptionalPayloadFloat(map[string]string{}, "min", -1.5)
	if err != nil {
		t.Fatalf("OptionalPayloadFloat returned error: %v", err)
	}
	if got != -1.5 {
		t.Fatalf("OptionalPayloadFloat = %.3f, want default -1.5", got)
	}
	got, err = OptionalPayloadFloat(map[string]string{"min": "-2.5"}, "min", -1.5)
	if err != nil {
		t.Fatalf("OptionalPayloadFloat returned error for present key: %v", err)
	}
	if got != -2.5 {
		t.Fatalf("OptionalPayloadFloat present = %.3f, want -2.5", got)
	}
}

func TestPayloadIntParsesRequiredInteger(t *testing.T) {
	got, err := PayloadInt(map[string]string{"rows": "16"}, "rows")
	if err != nil {
		t.Fatalf("PayloadInt returned error: %v", err)
	}
	if got != 16 {
		t.Fatalf("PayloadInt = %d, want 16", got)
	}
}

func TestPayloadIntRejectsMissingAndInvalidInteger(t *testing.T) {
	if _, err := PayloadInt(map[string]string{}, "rows"); err == nil {
		t.Fatal("expected missing key error")
	}
	if _, err := PayloadInt(map[string]string{"rows": "wide"}, "rows"); err == nil {
		t.Fatal("expected invalid integer error")
	}
}

func TestOptionalPayloadIntKeepsDefaultWhenMissing(t *testing.T) {
	got, err := OptionalPayloadInt(map[string]string{}, "rows", 8)
	if err != nil {
		t.Fatalf("OptionalPayloadInt returned error: %v", err)
	}
	if got != 8 {
		t.Fatalf("OptionalPayloadInt = %d, want default 8", got)
	}

	got, err = OptionalPayloadInt(map[string]string{"rows": "12"}, "rows", 8)
	if err != nil {
		t.Fatalf("OptionalPayloadInt returned error for present key: %v", err)
	}
	if got != 12 {
		t.Fatalf("OptionalPayloadInt present = %d, want 12", got)
	}
}
