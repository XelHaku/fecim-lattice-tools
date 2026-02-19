package compact

import (
	"fmt"
	"sort"
	"strings"
)

// GenerateVerilogAHeader returns a Verilog-A module header string with one
// `parameter real` declaration per entry in params.
//
// Parameters are emitted in sorted key order for deterministic output.
// The returned string is a syntactically valid Verilog-A module stub:
//
//	`include "disciplines.vams"
//	module modelName(/* ports omitted */);
//	  parameter real key = value;
//	  ...
//	endmodule
func GenerateVerilogAHeader(modelName string, params map[string]float64) string {
	keys := sortedKeys(params)

	var sb strings.Builder
	sb.WriteString("`include \"disciplines.vams\"\n")
	fmt.Fprintf(&sb, "module %s(/* ports omitted */);\n", modelName)
	for _, k := range keys {
		fmt.Fprintf(&sb, "  parameter real %s = %g;\n", k, params[k])
	}
	sb.WriteString("endmodule\n")
	return sb.String()
}

// GenerateSPICESubcircuit returns a SPICE .SUBCKT block string with one
// `.param` statement per entry in params.
//
// nodes lists the external node names (e.g. ["drain", "gate", "source", "body"]).
// Parameters are emitted in sorted key order for deterministic output.
//
// Example output:
//
//	.SUBCKT subcktName drain gate source body
//	.param key=value
//	...
//	.ENDS subcktName
func GenerateSPICESubcircuit(subcktName string, nodes []string, params map[string]float64) string {
	keys := sortedKeys(params)

	var sb strings.Builder
	fmt.Fprintf(&sb, ".SUBCKT %s %s\n", subcktName, strings.Join(nodes, " "))
	for _, k := range keys {
		fmt.Fprintf(&sb, ".param %s=%g\n", k, params[k])
	}
	fmt.Fprintf(&sb, ".ENDS %s\n", subcktName)
	return sb.String()
}

// sortedKeys returns the keys of m in ascending lexicographic order.
func sortedKeys(m map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
