package edacli

import "testing"

func TestValidateDesignName(t *testing.T) {
	if err := validateDesignName("safe-name_01"); err != nil {
		t.Fatalf("expected safe name, got %v", err)
	}
	bad := []string{"", "../escape", "a/b", "..", "name with space"}
	for _, n := range bad {
		if err := validateDesignName(n); err == nil {
			t.Fatalf("expected invalid name error for %q", n)
		}
	}
}

func TestValidateArrayGeometry(t *testing.T) {
	if err := validateArrayGeometry(128, 128); err != nil {
		t.Fatalf("expected valid geometry, got %v", err)
	}
	if err := validateArrayGeometry(0, 10); err == nil {
		t.Fatal("expected error for zero rows")
	}
	if err := validateArrayGeometry(maxArrayDim+1, 1); err == nil {
		t.Fatal("expected error for oversize dimension")
	}
}
