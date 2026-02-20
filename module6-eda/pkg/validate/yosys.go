// pkg/validate/yosys.go
package validate

import (
	"bytes"
	"fmt"
	"os/exec"
)

// RunYosysCheck executes yosys to validate the verilog file
// Returns the output log and any error encountered
func RunYosysCheck(verilogPath string) (string, error) {
	// Command: yosys -p "read_verilog <file>; hierarchy -check; check"
	cmdStr := fmt.Sprintf("read_verilog %s; hierarchy -check; check", verilogPath)

	cmd := exec.Command("yosys", "-p", cmdStr)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()

	output := outBuf.String()
	if err != nil {
		// Append stderr if there was an error
		output += "\nERROR:\n" + errBuf.String()
		return output, fmt.Errorf("yosys validation failed: %w", err)
	}

	return output, nil
}

// RunYosysCheckWithTop executes yosys validation with explicit cell blackbox and top module.
//
// Using -top avoids ambiguous module selection when multiple modules are loaded.
// cellVerilogPath is read with -lib so the cell is treated as a blackbox (no synthesis).
// topModule must match the top-level module declaration in verilogPath.
//
// Command:
//
//	yosys -p "read_verilog -lib <cell>; read_verilog <file>; hierarchy -check -top <top>; check; stat"
func RunYosysCheckWithTop(verilogPath, cellVerilogPath, topModule string) (string, error) {
	var cmdStr string
	if cellVerilogPath != "" {
		// Read cell as blackbox (-lib), then design, then elaborate with explicit top
		cmdStr = fmt.Sprintf(
			"read_verilog -lib %s; read_verilog %s; hierarchy -check -top %s; check; stat",
			cellVerilogPath, verilogPath, topModule)
	} else {
		cmdStr = fmt.Sprintf(
			"read_verilog %s; hierarchy -check -top %s; check; stat",
			verilogPath, topModule)
	}

	cmd := exec.Command("yosys", "-p", cmdStr)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()

	output := outBuf.String()
	if err != nil {
		output += "\nERROR:\n" + errBuf.String()
		return output, fmt.Errorf("yosys validation failed: %w", err)
	}

	return output, nil
}
