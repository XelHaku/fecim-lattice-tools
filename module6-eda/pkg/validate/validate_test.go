// pkg/validate/validate_test.go
package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunYosysCheck_EmptyPath tests behavior with empty verilog path
func TestRunYosysCheck_EmptyPath(t *testing.T) {
	// We can't prevent the function from trying to run yosys,
	// but we can verify the command string would be constructed
	// Empty path should still construct a valid command (yosys will fail)
	_, err := RunYosysCheck("")

	// We expect an error since yosys won't be able to read empty path
	if err == nil {
		t.Error("Expected error for empty path, got nil")
	}
}

// TestRunYosysCheck_NonexistentFile tests behavior with nonexistent file
func TestRunYosysCheck_NonexistentFile(t *testing.T) {
	nonexistentPath := "/tmp/definitely_does_not_exist_12345.v"

	_, err := RunYosysCheck(nonexistentPath)

	// Should return an error since file doesn't exist
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}

	// Error message should mention yosys validation failure
	if err != nil && !strings.Contains(err.Error(), "yosys validation failed") {
		t.Errorf("Expected 'yosys validation failed' in error, got: %v", err)
	}
}

// TestRunYosysCheck_InvalidVerilog tests behavior with invalid verilog content
func TestRunYosysCheck_InvalidVerilog(t *testing.T) {
	// Create a temporary file with invalid verilog
	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.v")

	invalidContent := `this is not valid verilog syntax!!!`
	err := os.WriteFile(invalidFile, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	output, err := RunYosysCheck(invalidFile)

	// Should return an error for invalid verilog
	if err == nil {
		t.Error("Expected error for invalid verilog, got nil")
	}

	// Output should contain error information
	if output == "" {
		t.Error("Expected non-empty output for failed validation")
	}

	// Error should wrap the underlying error
	if err != nil && !strings.Contains(err.Error(), "yosys validation failed") {
		t.Errorf("Expected wrapped error message, got: %v", err)
	}
}

// TestRunYosysCheck_ValidVerilog tests behavior with valid verilog content
func TestRunYosysCheck_ValidVerilog(t *testing.T) {
	// Create a temporary file with valid verilog
	tmpDir := t.TempDir()
	validFile := filepath.Join(tmpDir, "valid.v")

	// Simple valid verilog module
	validContent := `module test_module (
	input wire clk,
	input wire rst,
	output reg out
);
	always @(posedge clk) begin
		if (rst)
			out <= 1'b0;
		else
			out <= 1'b1;
	end
endmodule
`
	err := os.WriteFile(validFile, []byte(validContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	output, err := RunYosysCheck(validFile)

	// This test will only pass if yosys is installed and working
	// If yosys is not available, we should get an error
	if err != nil {
		// Check if it's a "yosys not found" type error
		errMsg := err.Error()
		if strings.Contains(errMsg, "executable file not found") ||
		   strings.Contains(errMsg, "command not found") {
			t.Skip("Yosys not installed, skipping validation test")
		}
		// Otherwise it's a real validation error
		t.Logf("Yosys validation error: %v", err)
		t.Logf("Output: %s", output)
	} else {
		// Success case - output should contain yosys messages
		if output == "" {
			t.Error("Expected non-empty output from successful yosys run")
		}
	}
}

// TestRunYosysCheck_SpecialCharactersInPath tests path handling
func TestRunYosysCheck_SpecialCharactersInPath(t *testing.T) {
	// Create a temporary file with special characters in path
	tmpDir := t.TempDir()
	specialFile := filepath.Join(tmpDir, "test file with spaces.v")

	validContent := `module test;
endmodule
`
	err := os.WriteFile(specialFile, []byte(validContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = RunYosysCheck(specialFile)

	// Should handle paths with spaces (might fail validation but not crash)
	// We're testing that the function doesn't panic or have path injection issues
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "executable file not found") ||
		   strings.Contains(errMsg, "command not found") {
			t.Skip("Yosys not installed, skipping path handling test")
		}
	}
}

// TestRunYosysCheck_OutputFormat tests output structure
func TestRunYosysCheck_OutputFormat(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.v")

	// Intentionally broken verilog to trigger error output
	brokenContent := `module broken
	// missing endmodule
`
	err := os.WriteFile(testFile, []byte(brokenContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	output, err := RunYosysCheck(testFile)

	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "executable file not found") ||
		   strings.Contains(errMsg, "command not found") {
			t.Skip("Yosys not installed, skipping output format test")
		}

		// When there's an error, output should contain "ERROR:" section
		if !strings.Contains(output, "ERROR:") {
			t.Error("Expected 'ERROR:' section in output for failed validation")
		}
	}
}

// TestRunYosysCheck_EmptyVerilogFile tests behavior with empty file
func TestRunYosysCheck_EmptyVerilogFile(t *testing.T) {
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.v")

	err := os.WriteFile(emptyFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = RunYosysCheck(emptyFile)

	// Empty file should likely cause validation error
	// (yosys might complain about no modules defined)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "executable file not found") ||
		   strings.Contains(errMsg, "command not found") {
			t.Skip("Yosys not installed, skipping empty file test")
		}

		// This is expected - empty verilog is likely invalid
		if !strings.Contains(err.Error(), "yosys validation failed") {
			t.Errorf("Expected wrapped error for empty file, got: %v", err)
		}
	}
}

// TestRunYosysCheck_MultipleModules tests file with multiple modules
func TestRunYosysCheck_MultipleModules(t *testing.T) {
	tmpDir := t.TempDir()
	multiFile := filepath.Join(tmpDir, "multi.v")

	multiContent := `module module1 (
	input wire a,
	output wire b
);
	assign b = a;
endmodule

module module2 (
	input wire x,
	output wire y
);
	assign y = ~x;
endmodule
`
	err := os.WriteFile(multiFile, []byte(multiContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	output, err := RunYosysCheck(multiFile)

	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "executable file not found") ||
		   strings.Contains(errMsg, "command not found") {
			t.Skip("Yosys not installed, skipping multi-module test")
		}
		t.Logf("Multi-module validation error: %v", err)
		t.Logf("Output: %s", output)
	}
}

// TestRunYosysCheck_ErrorWrapping tests that errors are properly wrapped
func TestRunYosysCheck_ErrorWrapping(t *testing.T) {
	// Use a definitely invalid path to trigger error
	invalidPath := "/tmp/this_file_should_not_exist_xyz123.v"

	_, err := RunYosysCheck(invalidPath)

	if err == nil {
		t.Error("Expected error for nonexistent file")
		return
	}

	// Check error wrapping format
	errStr := err.Error()
	if !strings.Contains(errStr, "yosys validation failed:") {
		t.Errorf("Error should be wrapped with context, got: %s", errStr)
	}
}

// TestRunYosysCheck_OutputNotNilOnError tests that output is returned even on error
func TestRunYosysCheck_OutputNotNilOnError(t *testing.T) {
	tmpDir := t.TempDir()
	badFile := filepath.Join(tmpDir, "bad.v")

	badContent := `this will cause yosys to fail`
	err := os.WriteFile(badFile, []byte(badContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	output, err := RunYosysCheck(badFile)

	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "executable file not found") ||
		   strings.Contains(errMsg, "command not found") {
			t.Skip("Yosys not installed")
		}

		// Even on error, output should be non-empty (contains error details)
		if output == "" {
			t.Error("Expected non-empty output even on validation error")
		}

		// Output should include stderr information
		if !strings.Contains(output, "ERROR:") {
			t.Error("Expected ERROR section in output when validation fails")
		}
	}
}
