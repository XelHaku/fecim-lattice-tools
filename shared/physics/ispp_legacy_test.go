package physics

import "testing"

func TestHysteresisDirection_String(t *testing.T) {
	tests := []struct {
		direction HysteresisDirection
		want      string
	}{
		{DirectionUnknown, "Unknown"},
		{DirectionAscending, "Ascending"},
		{DirectionDescending, "Descending"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.direction.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDirection(t *testing.T) {
	tests := []struct {
		name         string
		currentLevel int
		targetLevel  int
		want         HysteresisDirection
	}{
		{"Ascending", 5, 10, DirectionAscending},
		{"Descending", 20, 15, DirectionDescending},
		{"Equal", 10, 10, DirectionUnknown},
		{"FromZero", 0, 5, DirectionAscending},
		{"ToZero", 5, 0, DirectionDescending},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDirection(tt.currentLevel, tt.targetLevel)
			if got != tt.want {
				t.Errorf("GetDirection(%d, %d) = %v, want %v",
					tt.currentLevel, tt.targetLevel, got, tt.want)
			}
		})
	}
}

func TestDefaultISPPConfig(t *testing.T) {
	config := DefaultISPPConfig()

	tests := []struct {
		name  string
		got   float64
		want  float64
		isInt bool
		gotI  int
		wantI int
	}{
		{"StartRatio", config.StartRatio, 0.7, false, 0, 0},
		{"StepPercent", config.StepPercent, 0.01, false, 0, 0},
		{"SafetyCap", config.SafetyCap, 2.2, false, 0, 0},
		{"MaxPulses", 0, 0, true, config.MaxPulses, 40},
		{"Tolerance", 0, 0, true, config.Tolerance, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isInt {
				if tt.gotI != tt.wantI {
					t.Errorf("%s = %d, want %d", tt.name, tt.gotI, tt.wantI)
				}
			} else {
				if tt.got != tt.want {
					t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
				}
			}
		})
	}
}

func TestISPPResult_String(t *testing.T) {
	tests := []struct {
		result ISPPResult
		want   string
	}{
		{ISPPContinue, "Continue"},
		{ISPPSuccess, "Success"},
		{ISPPOvershoot, "Overshoot"},
		{ISPPMaxPulses, "MaxPulses"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.result.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewISPPCalculator(t *testing.T) {
	ec := 1.0
	numLevels := 30

	calc := NewISPPCalculator(ec, numLevels)

	if calc.Ec != ec {
		t.Errorf("Ec = %v, want %v", calc.Ec, ec)
	}
	if calc.NumLevels != numLevels {
		t.Errorf("NumLevels = %v, want %v", calc.NumLevels, numLevels)
	}

	// Verify default config
	defaultCfg := DefaultISPPConfig()
	if calc.Config.StartRatio != defaultCfg.StartRatio {
		t.Errorf("Config.StartRatio = %v, want %v", calc.Config.StartRatio, defaultCfg.StartRatio)
	}
	if calc.Config.StepPercent != defaultCfg.StepPercent {
		t.Errorf("Config.StepPercent = %v, want %v", calc.Config.StepPercent, defaultCfg.StepPercent)
	}
	if calc.Config.MaxPulses != defaultCfg.MaxPulses {
		t.Errorf("Config.MaxPulses = %v, want %v", calc.Config.MaxPulses, defaultCfg.MaxPulses)
	}
	if calc.Config.SafetyCap != defaultCfg.SafetyCap {
		t.Errorf("Config.SafetyCap = %v, want %v", calc.Config.SafetyCap, defaultCfg.SafetyCap)
	}
	if calc.Config.Tolerance != defaultCfg.Tolerance {
		t.Errorf("Config.Tolerance = %v, want %v", calc.Config.Tolerance, defaultCfg.Tolerance)
	}
}

func TestNewISPPCalculatorWithConfig(t *testing.T) {
	ec := 1.0
	numLevels := 30
	customCfg := ISPPConfig{
		StartRatio:  0.8,
		StepPercent: 0.02,
		MaxPulses:   50,
		SafetyCap:   2.5,
		Tolerance:   1,
	}

	calc := NewISPPCalculatorWithConfig(ec, numLevels, customCfg)

	if calc.Ec != ec {
		t.Errorf("Ec = %v, want %v", calc.Ec, ec)
	}
	if calc.NumLevels != numLevels {
		t.Errorf("NumLevels = %v, want %v", calc.NumLevels, numLevels)
	}
	if calc.Config.StartRatio != customCfg.StartRatio {
		t.Errorf("Config.StartRatio = %v, want %v", calc.Config.StartRatio, customCfg.StartRatio)
	}
	if calc.Config.StepPercent != customCfg.StepPercent {
		t.Errorf("Config.StepPercent = %v, want %v", calc.Config.StepPercent, customCfg.StepPercent)
	}
	if calc.Config.MaxPulses != customCfg.MaxPulses {
		t.Errorf("Config.MaxPulses = %v, want %v", calc.Config.MaxPulses, customCfg.MaxPulses)
	}
	if calc.Config.SafetyCap != customCfg.SafetyCap {
		t.Errorf("Config.SafetyCap = %v, want %v", calc.Config.SafetyCap, customCfg.SafetyCap)
	}
	if calc.Config.Tolerance != customCfg.Tolerance {
		t.Errorf("Config.Tolerance = %v, want %v", calc.Config.Tolerance, customCfg.Tolerance)
	}
}

func TestCalculateStartVoltage(t *testing.T) {
	calc := NewISPPCalculator(1.0, 30)

	tests := []struct {
		name               string
		calibratedVoltage  float64
		wantStartVoltage   float64
	}{
		{"Standard", 2.0, 1.4},   // 2.0 * 0.7
		{"Zero", 0.0, 0.0},
		{"Negative", -2.0, -1.4}, // -2.0 * 0.7
		{"LargeValue", 10.0, 7.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.CalculateStartVoltage(tt.calibratedVoltage)
			if got != tt.wantStartVoltage {
				t.Errorf("CalculateStartVoltage(%v) = %v, want %v",
					tt.calibratedVoltage, got, tt.wantStartVoltage)
			}
		})
	}
}

func TestCalculateVoltageStep(t *testing.T) {
	tests := []struct {
		name string
		ec   float64
		want float64
	}{
		{"StandardEc", 1.0, 0.01},  // 1.0 * 0.01
		{"LargeEc", 2.0, 0.02},     // 2.0 * 0.01
		{"SmallEc", 0.5, 0.005},    // 0.5 * 0.01
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewISPPCalculator(tt.ec, 30)
			got := calc.CalculateVoltageStep()
			if got != tt.want {
				t.Errorf("CalculateVoltageStep() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateNextVoltage(t *testing.T) {
	calc := NewISPPCalculator(1.0, 30)

	tests := []struct {
		name           string
		currentVoltage float64
		direction      HysteresisDirection
		want           float64
	}{
		{"AscendingPositive", 1.0, DirectionAscending, 1.01},
		{"AscendingZero", 0.0, DirectionAscending, 0.01},
		{"DescendingPositive", 1.0, DirectionDescending, 0.99},
		{"DescendingNegative", -1.0, DirectionDescending, -1.01},
		{"UnknownDefaultsAscending", 1.0, DirectionUnknown, 1.01}, // Unknown defaults to ascending behavior
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.CalculateNextVoltage(tt.currentVoltage, tt.direction)
			if got != tt.want {
				t.Errorf("CalculateNextVoltage(%v, %v) = %v, want %v",
					tt.currentVoltage, tt.direction, got, tt.want)
			}
		})
	}
}

func TestClampVoltage(t *testing.T) {
	calc := NewISPPCalculator(1.0, 30)

	tests := []struct {
		name      string
		voltage   float64
		direction HysteresisDirection
		want      float64
	}{
		// Ascending: clamp to [0, limit]
		{"AscendingWithinRange", 1.5, DirectionAscending, 1.5},
		{"AscendingAtZero", 0.0, DirectionAscending, 0.0},
		{"AscendingAtLimit", 2.2, DirectionAscending, 2.2},
		{"AscendingAboveLimit", 3.0, DirectionAscending, 2.2},
		{"AscendingBelowZero", -1.0, DirectionAscending, 0.0},

		// Descending: clamp to [-limit, 0]
		{"DescendingWithinRange", -1.5, DirectionDescending, -1.5},
		{"DescendingAtZero", 0.0, DirectionDescending, 0.0},
		{"DescendingAtLimit", -2.2, DirectionDescending, -2.2},
		{"DescendingBelowLimit", -3.0, DirectionDescending, -2.2},
		{"DescendingAboveZero", 1.0, DirectionDescending, 0.0},

		// Unknown: treated as Descending (else block clamps to [-limit, 0])
		{"UnknownPositive", 5.0, DirectionUnknown, 0.0},
		{"UnknownNegative", -5.0, DirectionUnknown, -2.2},
		{"UnknownWithinRange", -1.5, DirectionUnknown, -1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.ClampVoltage(tt.voltage, tt.direction)
			if got != tt.want {
				t.Errorf("ClampVoltage(%v, %v) = %v, want %v",
					tt.voltage, tt.direction, got, tt.want)
			}
		})
	}
}

func TestCheckResult(t *testing.T) {
	tests := []struct {
		name         string
		currentLevel int
		targetLevel  int
		direction    HysteresisDirection
		pulseCount   int
		tolerance    int
		want         ISPPResult
	}{
		// Exact target
		{"AtTarget", 10, 10, DirectionAscending, 5, 0, ISPPSuccess},

		// Within tolerance
		{"WithinTolerancePos", 11, 10, DirectionAscending, 5, 1, ISPPSuccess},
		{"WithinToleranceNeg", 9, 10, DirectionAscending, 5, 1, ISPPSuccess},
		{"OutsideTolerance", 12, 10, DirectionAscending, 5, 1, ISPPOvershoot},

		// Max pulses reached
		{"MaxPulsesReached", 5, 10, DirectionAscending, 40, 0, ISPPMaxPulses},

		// Overshoot
		{"AscendingOvershoot", 15, 10, DirectionAscending, 5, 0, ISPPOvershoot},
		{"DescendingOvershoot", 5, 10, DirectionDescending, 5, 0, ISPPOvershoot},

		// Continue
		{"AscendingContinue", 8, 10, DirectionAscending, 5, 0, ISPPContinue},
		{"DescendingContinue", 12, 10, DirectionDescending, 5, 0, ISPPContinue},

		// Unknown direction (else block checks descending overshoot: diff < 0)
		{"UnknownAtTarget", 10, 10, DirectionUnknown, 5, 0, ISPPSuccess},
		{"UnknownBelowTarget", 9, 10, DirectionUnknown, 5, 0, ISPPOvershoot}, // diff=-1 < 0
		{"UnknownAboveTarget", 12, 10, DirectionUnknown, 5, 0, ISPPContinue}, // diff=2 > 0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultISPPConfig()
			config.Tolerance = tt.tolerance
			calc := NewISPPCalculatorWithConfig(1.0, 30, config)

			got := calc.CheckResult(tt.currentLevel, tt.targetLevel, tt.direction, tt.pulseCount)
			if got != tt.want {
				t.Errorf("CheckResult(%d, %d, %v, %d) with tolerance=%d = %v, want %v",
					tt.currentLevel, tt.targetLevel, tt.direction, tt.pulseCount, tt.tolerance, got, tt.want)
			}
		})
	}
}

func TestGetResetVoltage(t *testing.T) {
	calc := NewISPPCalculator(1.0, 30)
	limit := -calc.Ec * calc.Config.SafetyCap // -2.2

	tests := []struct {
		name      string
		direction HysteresisDirection
		want      float64
	}{
		{"Ascending", DirectionAscending, limit},
		{"Descending", DirectionDescending, -limit},
		{"Unknown", DirectionUnknown, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.GetResetVoltage(tt.direction)
			if got != tt.want {
				t.Errorf("GetResetVoltage(%v) = %v, want %v", tt.direction, got, tt.want)
			}
		})
	}
}

func TestGetSaturationVoltage(t *testing.T) {
	calc := NewISPPCalculator(1.0, 30)
	limit := calc.Ec * calc.Config.SafetyCap // 2.2

	tests := []struct {
		name      string
		direction HysteresisDirection
		want      float64
	}{
		{"Ascending", DirectionAscending, limit},
		{"Descending", DirectionDescending, -limit},
		{"Unknown", DirectionUnknown, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.GetSaturationVoltage(tt.direction)
			if got != tt.want {
				t.Errorf("GetSaturationVoltage(%v) = %v, want %v", tt.direction, got, tt.want)
			}
		})
	}
}

func TestEstimatePulsesNeeded(t *testing.T) {
	calc := NewISPPCalculator(1.0, 30)

	tests := []struct {
		name         string
		currentLevel int
		targetLevel  int
		voltage      float64
		want         int
	}{
		{"SameLevel", 10, 10, 1.0, 1},
		{"SmallDiff", 10, 11, 1.0, 2},
		{"LargeDiff", 5, 15, 1.0, 11},
		{"NegativeDiff", 15, 10, 1.0, 6},
		{"FromZero", 0, 10, 1.0, 11},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.EstimatePulsesNeeded(tt.currentLevel, tt.targetLevel, tt.voltage)
			if got != tt.want {
				t.Errorf("EstimatePulsesNeeded(%d, %d, %v) = %v, want %v",
					tt.currentLevel, tt.targetLevel, tt.voltage, got, tt.want)
			}
		})
	}
}

func TestLevelError(t *testing.T) {
	tests := []struct {
		name    string
		current int
		target  int
		want    int
	}{
		{"Positive", 15, 10, 5},
		{"Negative", 10, 15, -5},
		{"Zero", 10, 10, 0},
		{"LargePositive", 30, 5, 25},
		{"LargeNegative", 5, 30, -25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LevelError(tt.current, tt.target)
			if got != tt.want {
				t.Errorf("LevelError(%d, %d) = %v, want %v", tt.current, tt.target, got, tt.want)
			}
		})
	}
}

func TestIsOvershoot(t *testing.T) {
	calc := NewISPPCalculator(1.0, 30)

	tests := []struct {
		name      string
		current   int
		target    int
		direction HysteresisDirection
		want      bool
	}{
		// Ascending: overshoot when current > target
		{"AscendingOvershoot", 15, 10, DirectionAscending, true},
		{"AscendingNotOvershoot", 8, 10, DirectionAscending, false},
		{"AscendingAtTarget", 10, 10, DirectionAscending, false},

		// Descending: overshoot when current < target
		{"DescendingOvershoot", 5, 10, DirectionDescending, true},
		{"DescendingNotOvershoot", 12, 10, DirectionDescending, false},
		{"DescendingAtTarget", 10, 10, DirectionDescending, false},

		// Unknown: treated as Descending (checks diff < 0)
		{"UnknownAbove", 15, 10, DirectionUnknown, false},  // diff=5 > 0
		{"UnknownBelow", 5, 10, DirectionUnknown, true},    // diff=-5 < 0 (overshoot)
		{"UnknownAtTarget", 10, 10, DirectionUnknown, false}, // diff=0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.IsOvershoot(tt.current, tt.target, tt.direction)
			if got != tt.want {
				t.Errorf("IsOvershoot(%d, %d, %v) = %v, want %v",
					tt.current, tt.target, tt.direction, got, tt.want)
			}
		})
	}
}
