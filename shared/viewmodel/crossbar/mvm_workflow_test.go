package crossbar

import "testing"

func TestMVMWorkflowComputesOutputVector(t *testing.T) {
	state := CrossbarState{
		Rows: 2,
		Cols: 3,
		Conductances: [][]float64{
			{1, 2, 3},
			{4, 5, 6},
		},
		InputVector:  []float64{0.5, 1.0, 2.0},
		OutputVector: make([]float64, 2),
	}

	updated := newMVMWorkflow(state).compute()

	want := []float64{8.5, 19}
	for i := range want {
		if updated.OutputVector[i] != want[i] {
			t.Fatalf("output[%d] = %.3f, want %.3f", i, updated.OutputVector[i], want[i])
		}
	}
}
