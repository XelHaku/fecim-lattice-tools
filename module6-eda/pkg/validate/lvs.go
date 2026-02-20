package validate

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

// ValidateLVSConsistency checks LEF pins == Verilog ports == SPICE subckt ports.
func ValidateLVSConsistency(lefPath, verilogPath, spicePath string) error {
	lefRaw, err := os.ReadFile(lefPath)
	if err != nil {
		return fmt.Errorf("read lef: %w", err)
	}
	verilogRaw, err := os.ReadFile(verilogPath)
	if err != nil {
		return fmt.Errorf("read verilog: %w", err)
	}
	spiceRaw, err := os.ReadFile(spicePath)
	if err != nil {
		return fmt.Errorf("read spice: %w", err)
	}

	lefPins, lefCell := parseLEFPinsAndMacro(string(lefRaw))
	vPins, vCell := parseVerilogPortsAndModule(string(verilogRaw))
	sPins, sCell := parseSpicePortsAndSubckt(string(spiceRaw))

	if lefCell == "" || vCell == "" || sCell == "" {
		return fmt.Errorf("missing module names in one or more files: LEF=%q Verilog=%q SPICE=%q", lefCell, vCell, sCell)
	}
	if lefCell != vCell || lefCell != sCell {
		return fmt.Errorf("cell/module mismatch: LEF=%s Verilog=%s SPICE=%s", lefCell, vCell, sCell)
	}
	if !sameSet(lefPins, vPins) {
		return fmt.Errorf("pin mismatch LEF vs Verilog: LEF=%v Verilog=%v", lefPins, vPins)
	}
	if !sameSet(lefPins, sPins) {
		return fmt.Errorf("pin mismatch LEF vs SPICE: LEF=%v SPICE=%v", lefPins, sPins)
	}

	return nil
}

func parseLEFPinsAndMacro(content string) ([]string, string) {
	macroRe := regexp.MustCompile(`(?m)^\s*MACRO\s+(\w+)`)
	pinRe := regexp.MustCompile(`(?m)^\s*PIN\s+(\w+)`)

	macro := ""
	if m := macroRe.FindStringSubmatch(content); m != nil {
		macro = m[1]
	}
	pins := make([]string, 0)
	for _, m := range pinRe.FindAllStringSubmatch(content, -1) {
		pins = append(pins, m[1])
	}
	return uniqueSorted(pins), macro
}

func parseVerilogPortsAndModule(content string) ([]string, string) {
	moduleRe := regexp.MustCompile(`(?m)^\s*module\s+(\w+)`)
	// Match both ANSI-style (input wire WL,  // comment) and non-ANSI (input WL;) declarations.
	// Captures the identifier immediately after the optional wire/reg keyword.
	// Non-ANSI style groups all ports into one multi-line match via [^;]+; which causes
	// comment words like "Ground" to be mistaken for the last port name (VGND).
	declRe := regexp.MustCompile(`(?m)^\s*(?:input|output|inout)\s+(?:wire\s+|reg\s+)?(\w+)`)

	module := ""
	if m := moduleRe.FindStringSubmatch(content); m != nil {
		module = m[1]
	}
	var ports []string
	for _, m := range declRe.FindAllStringSubmatch(content, -1) {
		ports = append(ports, m[1])
	}
	return uniqueSorted(ports), module
}

func parseSpicePortsAndSubckt(content string) ([]string, string) {
	subcktRe := regexp.MustCompile(`(?im)^\s*\.subckt\s+(\w+)\s+(.+)$`)
	m := subcktRe.FindStringSubmatch(content)
	if m == nil {
		return nil, ""
	}
	name := m[1]
	ports := strings.Fields(strings.TrimSpace(m[2]))
	return uniqueSorted(ports), name
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func uniqueSorted(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
