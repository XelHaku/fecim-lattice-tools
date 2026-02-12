package validation

import "testing"

func TestExtractLEFData_MissingFile(t *testing.T) {
	_, _, err := extractLEFData("/does/not/exist.lef")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExtractLibData_MissingFile(t *testing.T) {
	_, _, err := extractLibData("/does/not/exist.lib")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExtractVerilogData_MissingFile(t *testing.T) {
	_, _, err := extractVerilogData("/does/not/exist.v")
	if err == nil {
		t.Fatal("expected error")
	}
}
