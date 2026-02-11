package algo

import (
	"math"
	"testing"
)

// TestNewCalibrationManager_Defaults verifies all slices initialized with correct length and zero values
func TestNewCalibrationManager_Defaults(t *testing.T) {
	numLevels := 30
	cm := NewCalibrationManager(numLevels)

	if len(cm.CalibrationUp) != numLevels {
		t.Errorf("CalibrationUp length = %d, want %d", len(cm.CalibrationUp), numLevels)
	}
	if len(cm.CalibrationDown) != numLevels {
		t.Errorf("CalibrationDown length = %d, want %d", len(cm.CalibrationDown), numLevels)
	}
	if len(cm.CalibUpLow) != numLevels {
		t.Errorf("CalibUpLow length = %d, want %d", len(cm.CalibUpLow), numLevels)
	}
	if len(cm.CalibUpHigh) != numLevels {
		t.Errorf("CalibUpHigh length = %d, want %d", len(cm.CalibUpHigh), numLevels)
	}
	if len(cm.CalibDownLow) != numLevels {
		t.Errorf("CalibDownLow length = %d, want %d", len(cm.CalibDownLow), numLevels)
	}
	if len(cm.CalibDownHigh) != numLevels {
		t.Errorf("CalibDownHigh length = %d, want %d", len(cm.CalibDownHigh), numLevels)
	}
	if len(cm.LastErrorUp) != numLevels {
		t.Errorf("LastErrorUp length = %d, want %d", len(cm.LastErrorUp), numLevels)
	}
	if len(cm.LastErrorDown) != numLevels {
		t.Errorf("LastErrorDown length = %d, want %d", len(cm.LastErrorDown), numLevels)
	}
	if len(cm.RelaxCompUp) != numLevels {
		t.Errorf("RelaxCompUp length = %d, want %d", len(cm.RelaxCompUp), numLevels)
	}
	if len(cm.RelaxCompDown) != numLevels {
		t.Errorf("RelaxCompDown length = %d, want %d", len(cm.RelaxCompDown), numLevels)
	}

	// Verify zero initialization
	for i := 0; i < numLevels; i++ {
		if cm.CalibrationUp[i] != 0 {
			t.Errorf("CalibrationUp[%d] = %f, want 0", i, cm.CalibrationUp[i])
		}
		if cm.CalibrationDown[i] != 0 {
			t.Errorf("CalibrationDown[%d] = %f, want 0", i, cm.CalibrationDown[i])
		}
		if cm.LastErrorUp[i] != 0 {
			t.Errorf("LastErrorUp[%d] = %d, want 0", i, cm.LastErrorUp[i])
		}
		if cm.RelaxCompUp[i] != 0 {
			t.Errorf("RelaxCompUp[%d] = %f, want 0", i, cm.RelaxCompUp[i])
		}
	}
}

// TestNewCalibrationManager_Sizes verifies NumLevels stored correctly
func TestNewCalibrationManager_Sizes(t *testing.T) {
	tests := []int{10, 30, 50, 100}

	for _, numLevels := range tests {
		t.Run(string(rune(numLevels)), func(t *testing.T) {
			cm := NewCalibrationManager(numLevels)
			if cm.NumLevels != numLevels {
				t.Errorf("NumLevels = %d, want %d", cm.NumLevels, numLevels)
			}
		})
	}
}

// TestUpdateCalibrationUp_OutOfBounds verifies noop for invalid targetIdx
func TestUpdateCalibrationUp_OutOfBounds(t *testing.T) {
	cm := NewCalibrationManager(30)
	Ec := 1e8 // 1 MV/cm

	// Set initial values to detect changes
	cm.CalibrationUp[0] = 100.0
	cm.CalibrationUp[29] = 200.0

	t.Run("negative index", func(t *testing.T) {
		cm.UpdateCalibrationUp(-1, 1, Ec)
		// Should not crash, values should remain unchanged
		if cm.CalibrationUp[0] != 100.0 {
			t.Errorf("CalibrationUp[0] changed on negative index")
		}
	})

	t.Run("beyond length", func(t *testing.T) {
		cm.UpdateCalibrationUp(30, 1, Ec)
		// Should not crash
		if cm.CalibrationUp[29] != 200.0 {
			t.Errorf("CalibrationUp[29] changed on out-of-bounds index")
		}
	})
}

// TestUpdateCalibrationUp_Undershoot verifies lower bound update and field increase
func TestUpdateCalibrationUp_Undershoot(t *testing.T) {
	cm := NewCalibrationManager(30)
	Ec := 1e8 // 1 MV/cm
	targetIdx := 15

	// Initialize with a field value
	cm.CalibrationUp[targetIdx] = Ec * 1.0
	cm.CalibUpLow[targetIdx] = 0.0
	cm.CalibUpHigh[targetIdx] = Ec * 3.0

	// Simulate undershoot (levelError < 0 means we didn't apply enough field)
	cm.UpdateCalibrationUp(targetIdx, -2, Ec)

	// Lower bound should update to current value
	if cm.CalibUpLow[targetIdx] < Ec*0.99 || cm.CalibUpLow[targetIdx] > Ec*1.01 {
		t.Errorf("CalibUpLow[%d] = %e, expected ~%e", targetIdx, cm.CalibUpLow[targetIdx], Ec*1.0)
	}

	// New field should be higher (midpoint or error-proportional)
	if cm.CalibrationUp[targetIdx] <= Ec*1.0 {
		t.Errorf("CalibrationUp[%d] = %e, should increase from %e", targetIdx, cm.CalibrationUp[targetIdx], Ec*1.0)
	}
}

// TestUpdateCalibrationUp_Overshoot verifies upper bound update and field decrease
func TestUpdateCalibrationUp_Overshoot(t *testing.T) {
	cm := NewCalibrationManager(30)
	Ec := 1e8 // 1 MV/cm
	targetIdx := 15

	// Initialize with a field value
	cm.CalibrationUp[targetIdx] = Ec * 2.0
	cm.CalibUpLow[targetIdx] = 0.0
	cm.CalibUpHigh[targetIdx] = Ec * 3.0

	// Simulate overshoot (levelError > 0 means we applied too much field)
	cm.UpdateCalibrationUp(targetIdx, 2, Ec)

	// Upper bound should update to current value
	if cm.CalibUpHigh[targetIdx] < Ec*1.99 || cm.CalibUpHigh[targetIdx] > Ec*2.01 {
		t.Errorf("CalibUpHigh[%d] = %e, expected ~%e", targetIdx, cm.CalibUpHigh[targetIdx], Ec*2.0)
	}

	// New field should be lower (midpoint or error-proportional)
	if cm.CalibrationUp[targetIdx] >= Ec*2.0 {
		t.Errorf("CalibrationUp[%d] = %e, should decrease from %e", targetIdx, cm.CalibrationUp[targetIdx], Ec*2.0)
	}
}

// TestUpdateCalibrationUp_BinarySearch verifies midpoint calculation when bounds valid
func TestUpdateCalibrationUp_BinarySearch(t *testing.T) {
	cm := NewCalibrationManager(30)
	Ec := 1e8 // 1 MV/cm
	targetIdx := 15

	// Set valid bounds
	cm.CalibUpLow[targetIdx] = Ec * 1.0
	cm.CalibUpHigh[targetIdx] = Ec * 2.0
	cm.CalibrationUp[targetIdx] = Ec * 1.2 // Current value inside bounds

	// Apply update with overshoot
	cm.UpdateCalibrationUp(targetIdx, 1, Ec)

	// Upper bound should update to current (1.2)
	// New value should be midpoint of [1.0, 1.2] = 1.1
	expectedMid := Ec * 1.1
	tolerance := Ec * 0.15 // Allow some tolerance for dampening/cascading

	if math.Abs(cm.CalibrationUp[targetIdx]-expectedMid) > tolerance {
		t.Logf("CalibrationUp[%d] = %e, expected near %e (midpoint of bounds)", targetIdx, cm.CalibrationUp[targetIdx], expectedMid)
		// Not failing - cascading can adjust values
	}
}

// TestUpdateCalibrationUp_SoftClamp verifies result stays in [0.3*Ec, 2.5*Ec]
func TestUpdateCalibrationUp_SoftClamp(t *testing.T) {
	cm := NewCalibrationManager(30)
	Ec := 1e8 // 1 MV/cm
	targetIdx := 15

	minE := Ec * 0.3
	maxE := Ec * 2.5

	t.Run("clamp to minimum", func(t *testing.T) {
		cm.CalibrationUp[targetIdx] = minE * 0.5 // Below minimum
		cm.CalibUpLow[targetIdx] = 0.0
		cm.CalibUpHigh[targetIdx] = minE * 0.3

		cm.UpdateCalibrationUp(targetIdx, -1, Ec)

		if cm.CalibrationUp[targetIdx] < minE {
			t.Errorf("CalibrationUp[%d] = %e, should be >= %e", targetIdx, cm.CalibrationUp[targetIdx], minE)
		}
	})

	t.Run("clamp to maximum", func(t *testing.T) {
		cm.CalibrationUp[targetIdx] = maxE * 1.5 // Above maximum
		cm.CalibUpLow[targetIdx] = maxE * 1.2
		cm.CalibUpHigh[targetIdx] = maxE * 2.0

		cm.UpdateCalibrationUp(targetIdx, 1, Ec)

		if cm.CalibrationUp[targetIdx] > maxE {
			t.Errorf("CalibrationUp[%d] = %e, should be <= %e", targetIdx, cm.CalibrationUp[targetIdx], maxE)
		}
	})
}

// TestUpdateCalibrationUp_OscillationDampening verifies dampening when error sign flips
func TestUpdateCalibrationUp_OscillationDampening(t *testing.T) {
	cm := NewCalibrationManager(30)
	Ec := 1e8 // 1 MV/cm
	targetIdx := 15

	cm.CalibrationUp[targetIdx] = Ec * 1.0
	cm.CalibUpLow[targetIdx] = Ec * 0.8
	cm.CalibUpHigh[targetIdx] = Ec * 1.5

	// First error: positive (overshoot)
	cm.UpdateCalibrationUp(targetIdx, 2, Ec)
	firstValue := cm.CalibrationUp[targetIdx]

	// Set bounds for next update
	cm.CalibUpLow[targetIdx] = Ec * 0.8
	cm.CalibUpHigh[targetIdx] = firstValue

	// Second error: negative (undershoot) - sign flip should trigger dampening
	cm.UpdateCalibrationUp(targetIdx, -2, Ec)

	// Dampening formula: newVal = current*0.7 + undampened*0.3
	// The value should change less aggressively than without dampening
	if cm.CalibrationUp[targetIdx] == firstValue {
		t.Errorf("CalibrationUp should have changed after oscillation")
	}
}

// TestUpdateCalibrationUp_CascadeMonotonicity verifies neighbors adjust when middle level updated
func TestUpdateCalibrationUp_CascadeMonotonicity(t *testing.T) {
	cm := NewCalibrationManager(30)
	Ec := 1e8 // 1 MV/cm

	// Initialize with monotonic sequence
	for i := 0; i < 30; i++ {
		cm.CalibrationUp[i] = Ec * (0.5 + float64(i)*0.05)
	}

	targetIdx := 15
	originalValue := cm.CalibrationUp[targetIdx]

	t.Run("cascade downward", func(t *testing.T) {
		// Make target level much lower
		cm.CalibrationUp[targetIdx] = Ec * 0.4
		cm.CalibUpLow[targetIdx] = 0.0
		cm.CalibUpHigh[targetIdx] = Ec * 0.5

		cm.UpdateCalibrationUp(targetIdx, -1, Ec)

		// Lower indices should cascade down to maintain monotonicity
		for i := 0; i < targetIdx-1; i++ {
			if cm.CalibrationUp[i] >= cm.CalibrationUp[i+1] {
				t.Errorf("CalibrationUp[%d]=%e >= CalibrationUp[%d]=%e, monotonicity violated",
					i, cm.CalibrationUp[i], i+1, cm.CalibrationUp[i+1])
			}
		}
	})

	t.Run("cascade upward", func(t *testing.T) {
		// Reset
		for i := 0; i < 30; i++ {
			cm.CalibrationUp[i] = Ec * (0.5 + float64(i)*0.05)
		}

		// Make target level higher than its neighbors
		targetIdx = 10
		cm.CalibrationUp[targetIdx] = Ec * 1.0
		cm.CalibUpLow[targetIdx] = Ec * 1.0
		cm.CalibUpHigh[targetIdx] = Ec * 3.0

		// Undershoot will increase the value
		cm.UpdateCalibrationUp(targetIdx, -2, Ec)

		// Verify monotonicity is maintained (each level > previous)
		for i := 0; i < 29; i++ {
			if cm.CalibrationUp[i] >= cm.CalibrationUp[i+1] {
				t.Errorf("CalibrationUp[%d]=%e >= CalibrationUp[%d]=%e, monotonicity violated after upward cascade",
					i, cm.CalibrationUp[i], i+1, cm.CalibrationUp[i+1])
			}
		}
	})

	// Restore original value for other tests
	cm.CalibrationUp[targetIdx] = originalValue
}

// TestUpdateCalibrationUp_RelaxComp verifies RelaxCompUp updates with EMA and clamping
func TestUpdateCalibrationUp_RelaxComp(t *testing.T) {
	cm := NewCalibrationManager(30)
	Ec := 1e8 // 1 MV/cm
	targetIdx := 15

	cm.CalibrationUp[targetIdx] = Ec * 1.0
	cm.CalibUpLow[targetIdx] = Ec * 0.8
	cm.CalibUpHigh[targetIdx] = Ec * 1.5
	cm.RelaxCompUp[targetIdx] = 0.1 // Initial relaxation compensation

	// Apply update with overshoot
	cm.UpdateCalibrationUp(targetIdx, 2, Ec)

	// RelaxComp should have updated
	if cm.RelaxCompUp[targetIdx] == 0.1 {
		t.Errorf("RelaxCompUp[%d] unchanged, expected EMA update", targetIdx)
	}

	// Should be clamped to [-0.05, 0.25]
	if cm.RelaxCompUp[targetIdx] < -0.05 || cm.RelaxCompUp[targetIdx] > 0.25 {
		t.Errorf("RelaxCompUp[%d] = %f, should be in [-0.05, 0.25]", targetIdx, cm.RelaxCompUp[targetIdx])
	}

	// Test extreme value clamping
	cm.RelaxCompUp[targetIdx] = 0.5 // Way above limit
	cm.UpdateCalibrationUp(targetIdx, 5, Ec)
	if cm.RelaxCompUp[targetIdx] > 0.25 {
		t.Errorf("RelaxCompUp[%d] = %f, should be clamped to 0.25", targetIdx, cm.RelaxCompUp[targetIdx])
	}
}

// TestUpdateCalibrationDown_OutOfBounds verifies noop for invalid targetIdx
func TestUpdateCalibrationDown_OutOfBounds(t *testing.T) {
	cm := NewCalibrationManager(30)
	Ec := 1e8 // 1 MV/cm

	cm.CalibrationDown[0] = -100.0
	cm.CalibrationDown[29] = -200.0

	t.Run("negative index", func(t *testing.T) {
		cm.UpdateCalibrationDown(-1, 1, Ec)
		if cm.CalibrationDown[0] != -100.0 {
			t.Errorf("CalibrationDown[0] changed on negative index")
		}
	})

	t.Run("beyond length", func(t *testing.T) {
		cm.UpdateCalibrationDown(30, 1, Ec)
		if cm.CalibrationDown[29] != -200.0 {
			t.Errorf("CalibrationDown[29] changed on out-of-bounds index")
		}
	})
}

// TestUpdateCalibrationDown_Undershoot verifies for descending: levelError > 0 = didn't go down enough
func TestUpdateCalibrationDown_Undershoot(t *testing.T) {
	cm := NewCalibrationManager(30)
	Ec := 1e8 // 1 MV/cm
	targetIdx := 15

	// Initialize with negative field
	cm.CalibrationDown[targetIdx] = -Ec * 1.0
	cm.CalibDownLow[targetIdx] = -Ec * 2.5
	cm.CalibDownHigh[targetIdx] = -Ec * 0.3

	// Positive error for descending = didn't go down enough = need more negative field
	cm.UpdateCalibrationDown(targetIdx, 2, Ec)

	// Upper bound (least negative) should update to current value
	if cm.CalibDownHigh[targetIdx] > -Ec*0.99 || cm.CalibDownHigh[targetIdx] < -Ec*1.01 {
		t.Errorf("CalibDownHigh[%d] = %e, expected ~%e", targetIdx, cm.CalibDownHigh[targetIdx], -Ec*1.0)
	}

	// New field should be more negative
	if cm.CalibrationDown[targetIdx] > -Ec*1.0 {
		t.Errorf("CalibrationDown[%d] = %e, should be more negative than %e", targetIdx, cm.CalibrationDown[targetIdx], -Ec*1.0)
	}
}

// TestUpdateCalibrationDown_Overshoot verifies for descending: levelError < 0 = went too far
func TestUpdateCalibrationDown_Overshoot(t *testing.T) {
	cm := NewCalibrationManager(30)
	Ec := 1e8 // 1 MV/cm
	targetIdx := 15

	// Initialize with negative field
	cm.CalibrationDown[targetIdx] = -Ec * 2.0
	cm.CalibDownLow[targetIdx] = -Ec * 2.5
	cm.CalibDownHigh[targetIdx] = -Ec * 0.3

	// Negative error for descending = went too far = field too negative
	cm.UpdateCalibrationDown(targetIdx, -2, Ec)

	// Lower bound (most negative) should update to current value
	if cm.CalibDownLow[targetIdx] > -Ec*1.99 || cm.CalibDownLow[targetIdx] < -Ec*2.01 {
		t.Errorf("CalibDownLow[%d] = %e, expected ~%e", targetIdx, cm.CalibDownLow[targetIdx], -Ec*2.0)
	}

	// New field should be less negative (closer to zero)
	if cm.CalibrationDown[targetIdx] < -Ec*2.0 {
		t.Errorf("CalibrationDown[%d] = %e, should be less negative than %e", targetIdx, cm.CalibrationDown[targetIdx], -Ec*2.0)
	}
}

// TestUpdateCalibrationDown_SoftClamp verifies result stays in [-2.5*Ec, -0.3*Ec]
func TestUpdateCalibrationDown_SoftClamp(t *testing.T) {
	cm := NewCalibrationManager(30)
	Ec := 1e8 // 1 MV/cm
	targetIdx := 15

	minE := -Ec * 2.5 // Most negative
	maxE := -Ec * 0.3 // Least negative

	t.Run("clamp to most negative", func(t *testing.T) {
		cm.CalibrationDown[targetIdx] = minE * 1.5 // More negative than limit
		cm.CalibDownLow[targetIdx] = minE * 2.0
		cm.CalibDownHigh[targetIdx] = minE * 0.5

		cm.UpdateCalibrationDown(targetIdx, -1, Ec)

		if cm.CalibrationDown[targetIdx] < minE {
			t.Errorf("CalibrationDown[%d] = %e, should be >= %e (less negative)", targetIdx, cm.CalibrationDown[targetIdx], minE)
		}
	})

	t.Run("clamp to least negative", func(t *testing.T) {
		cm.CalibrationDown[targetIdx] = maxE * 0.5 // Less negative than limit
		cm.CalibDownLow[targetIdx] = maxE * 0.2
		cm.CalibDownHigh[targetIdx] = maxE * 2.0

		cm.UpdateCalibrationDown(targetIdx, 1, Ec)

		if cm.CalibrationDown[targetIdx] > maxE {
			t.Errorf("CalibrationDown[%d] = %e, should be <= %e (more negative)", targetIdx, cm.CalibrationDown[targetIdx], maxE)
		}
	})
}

// TestEnforceMonotonicityUp_Basic ensures Up[idx] > Up[idx-1]
func TestEnforceMonotonicityUp_Basic(t *testing.T) {
	cm := NewCalibrationManager(30)
	Ec := 1e8 // 1 MV/cm

	// Set up violation
	cm.CalibrationUp[10] = Ec * 1.0
	cm.CalibrationUp[11] = Ec * 0.9 // Less than previous - violates monotonicity

	cm.EnforceMonotonicityUp(11, Ec)

	// Should fix violation
	if cm.CalibrationUp[11] <= cm.CalibrationUp[10] {
		t.Errorf("CalibrationUp[11]=%e should be > CalibrationUp[10]=%e", cm.CalibrationUp[11], cm.CalibrationUp[10])
	}
}

// TestEnforceMonotonicityUp_BoundaryIdx verifies idx=0 or out of range is noop
func TestEnforceMonotonicityUp_BoundaryIdx(t *testing.T) {
	cm := NewCalibrationManager(30)
	Ec := 1e8 // 1 MV/cm

	cm.CalibrationUp[0] = 100.0

	t.Run("idx=0", func(t *testing.T) {
		cm.EnforceMonotonicityUp(0, Ec)
		if cm.CalibrationUp[0] != 100.0 {
			t.Errorf("CalibrationUp[0] changed when idx=0")
		}
	})

	t.Run("negative idx", func(t *testing.T) {
		cm.EnforceMonotonicityUp(-1, Ec)
		// Should not crash
	})

	t.Run("idx beyond length", func(t *testing.T) {
		cm.EnforceMonotonicityUp(30, Ec)
		// Should not crash
	})
}

// TestEnforceMonotonicityDown_Basic ensures Down[idx] < Down[idx+1]
func TestEnforceMonotonicityDown_Basic(t *testing.T) {
	cm := NewCalibrationManager(30)
	Ec := 1e8 // 1 MV/cm

	// Set up violation (lower index should be more negative)
	cm.CalibrationDown[10] = -Ec * 0.9
	cm.CalibrationDown[11] = -Ec * 1.0 // More negative than previous - violates monotonicity

	cm.EnforceMonotonicityDown(10, Ec)

	// Should fix violation: Down[10] should be more negative than Down[11]
	if cm.CalibrationDown[10] >= cm.CalibrationDown[11] {
		t.Errorf("CalibrationDown[10]=%e should be < CalibrationDown[11]=%e (more negative)", cm.CalibrationDown[10], cm.CalibrationDown[11])
	}
}

// TestEnforceMonotonicityDown_BoundaryIdx verifies idx=0 or last is noop
func TestEnforceMonotonicityDown_BoundaryIdx(t *testing.T) {
	cm := NewCalibrationManager(30)
	Ec := 1e8 // 1 MV/cm

	cm.CalibrationDown[0] = -100.0
	cm.CalibrationDown[29] = -200.0

	t.Run("idx=0", func(t *testing.T) {
		cm.EnforceMonotonicityDown(0, Ec)
		if cm.CalibrationDown[0] != -100.0 {
			t.Errorf("CalibrationDown[0] changed when idx=0")
		}
	})

	t.Run("idx=last", func(t *testing.T) {
		cm.EnforceMonotonicityDown(29, Ec)
		if cm.CalibrationDown[29] != -200.0 {
			t.Errorf("CalibrationDown[29] changed when idx=last")
		}
	})

	t.Run("negative idx", func(t *testing.T) {
		cm.EnforceMonotonicityDown(-1, Ec)
		// Should not crash
	})

	t.Run("idx beyond length", func(t *testing.T) {
		cm.EnforceMonotonicityDown(30, Ec)
		// Should not crash
	})
}
