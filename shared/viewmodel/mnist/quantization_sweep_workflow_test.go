package mnist

import "testing"

func TestQuantizationSweepWorkflowComputesAccuracyCurve(t *testing.T) {
	updated := newQuantizationSweepWorkflow(MNISTState{}).compute()

	wantLevels := []int{2, 4, 8, 16, 32, 64, 128}
	wantAccuracy := []float64{0.55, 0.65, 0.74, 0.79, 0.82, 0.84, 0.85}
	if len(updated.SweepLevels) != len(wantLevels) || len(updated.SweepAccuracy) != len(wantAccuracy) {
		t.Fatalf("sweep lengths = %d/%d", len(updated.SweepLevels), len(updated.SweepAccuracy))
	}
	for i := range wantLevels {
		if updated.SweepLevels[i] != wantLevels[i] {
			t.Fatalf("level[%d] = %d, want %d", i, updated.SweepLevels[i], wantLevels[i])
		}
		if updated.SweepAccuracy[i] != wantAccuracy[i] {
			t.Fatalf("accuracy[%d] = %.2f, want %.2f", i, updated.SweepAccuracy[i], wantAccuracy[i])
		}
	}
}
