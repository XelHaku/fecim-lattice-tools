// module6-eda/cmd/hello/main.go
package main

import (
	"fmt"
	"math"
)

// === CORE DATA STRUCTURES ===

type CellAssignment struct {
	Row         int
	Col         int
	Weight      float64 // Original weight
	Level       int     // 0 to 29 (30 levels)
	Conductance float64 // μS
}

// === CORE COMPILER FUNCTION ===

func Compile(weights [][]float64) []CellAssignment {
	// Find weight range
	wMin, wMax := weights[0][0], weights[0][0]
	for _, row := range weights {
		for _, w := range row {
			if w < wMin {
				wMin = w
			}
			if w > wMax {
				wMax = w
			}
		}
	}
	wAbsMax := math.Max(math.Abs(wMin), math.Abs(wMax))

	// Compile each weight to a cell
	var cells []CellAssignment
	levels := 30
	gMin, gMax := 10.0, 100.0 // μS

	for i, row := range weights {
		for j, w := range row {
			// Quantize: [-wAbsMax, +wAbsMax] → [0, 29]
			normalized := (w + wAbsMax) / (2 * wAbsMax)
			level := int(math.Round(normalized * float64(levels-1)))
			if level < 0 {
				level = 0
			}
			if level >= levels {
				level = levels - 1
			}

			// Map to conductance
			conductance := gMin + float64(level)/float64(levels-1)*(gMax-gMin)

			cells = append(cells, CellAssignment{
				Row:         i,
				Col:         j,
				Weight:      w,
				Level:       level,
				Conductance: conductance,
			})
		}
	}

	return cells
}

// === MAIN: PROVE IT WORKS ===

func main() {
	fmt.Println("=== FeCIM EDA: Hello World ===")
	fmt.Println()

	// Sample 4x4 weight matrix
	weights := [][]float64{
		{0.5, -0.3, 0.8, -0.1},
		{-0.7, 0.2, -0.9, 0.4},
		{0.1, -0.5, 0.6, -0.8},
		{-0.2, 0.9, -0.4, 0.7},
	}

	fmt.Println("INPUT: 4x4 Weight Matrix")
	fmt.Println("------------------------")
	for _, row := range weights {
		fmt.Printf("  %v\n", row)
	}
	fmt.Println()

	// Compile
	cells := Compile(weights)

	fmt.Println("OUTPUT: FeCIM Cell Assignments")
	fmt.Println("------------------------------")
	fmt.Printf("%-4s %-4s %-8s %-6s %-12s\n", "Row", "Col", "Weight", "Level", "Conductance")
	fmt.Println("---- ---- -------- ------ ------------")

	for _, c := range cells {
		fmt.Printf("%-4d %-4d %+7.3f  %-6d %.2f μS\n",
			c.Row, c.Col, c.Weight, c.Level, c.Conductance)
	}

	fmt.Println()
	fmt.Println("=== COMPILATION SUCCESSFUL ===")
	fmt.Printf("Total cells: %d\n", len(cells))
	fmt.Printf("Levels used: 30 (demo baseline; conference claim)\n")
	fmt.Printf("Conductance range: 10.0 - 100.0 μS\n")
}
