package arraysim

import (
	"fmt"
	"testing"
)

func TestReadMargin_SweepArraySizesAndLevels(t *testing.T) {
	t.Logf("available coupling modes: %s, %s, %s", CouplingIdeal.String(), CouplingTierA.String(), CouplingTierB.String())

	sizes := []int{8, 16, 32, 64, 128}
	levelsSweep := []int{2, 4, 8, 16}
	mode := CouplingTierA

	results := make(map[int]map[int]ReadMarginResult)
	for _, size := range sizes {
		results[size] = make(map[int]ReadMarginResult)
		for _, levels := range levelsSweep {
			res := ReadMarginAnalysis(ArrayConfig{Rows: size, Cols: size, CouplingMode: mode}, levels)
			results[size][levels] = res
		}
	}

	t.Logf("Read margin sweep (%s)", mode.String())
	t.Logf("array_size\\levels |     2 |     4 |     8 |    16  (mV)")
	for _, size := range sizes {
		line := fmt.Sprintf("%15d |", size)
		for _, levels := range levelsSweep {
			line += fmt.Sprintf(" %6.2f |", 1e3*results[size][levels].MinMarginV)
		}
		t.Log(line)
	}

	for _, levels := range levelsSweep {
		prev := results[sizes[0]][levels].MinMarginV
		for i := 1; i < len(sizes); i++ {
			curr := results[sizes[i]][levels].MinMarginV
			if curr > prev+1e-9 {
				t.Fatalf("expected margin to not increase with array size at levels=%d: size=%d margin=%.6f mV > previous %.6f mV",
					levels, sizes[i], curr*1e3, prev*1e3)
			}
			prev = curr
		}
	}

	for _, size := range sizes {
		prev := results[size][levelsSweep[0]].MinMarginV
		for i := 1; i < len(levelsSweep); i++ {
			curr := results[size][levelsSweep[i]].MinMarginV
			if curr > prev+1e-9 {
				t.Fatalf("expected margin to not increase with level count at size=%d levels=%d: %.6f mV > previous %.6f mV",
					size, levelsSweep[i], curr*1e3, prev*1e3)
			}
			prev = curr
		}
	}

	worst := results[128][16]
	if worst.Reliable {
		t.Fatalf("expected at least one unreliable point; 128x128/16-levels remained reliable (min margin %.3f mV)", worst.MinMarginV*1e3)
	}
}
