package system

import "testing"

func TestDACLatencyNS(t *testing.T) {
	l := NewLatencyModel(64, 64, Node65nm)
	v := l.DACLatencyNS()
	if v <= 0 {
		t.Errorf("DACLatencyNS() = %g, want > 0", v)
	}
}

func TestCrossbarSettlingNS(t *testing.T) {
	l := NewLatencyModel(64, 64, Node65nm)
	v := l.CrossbarSettlingNS()
	if v <= 0 {
		t.Errorf("CrossbarSettlingNS() = %g, want > 0", v)
	}
}

func TestADCLatencyNS(t *testing.T) {
	l := NewLatencyModel(64, 64, Node65nm)
	v := l.ADCLatencyNS(4)
	if v <= 0 {
		t.Errorf("ADCLatencyNS(4) = %g, want > 0", v)
	}
	// More bits → longer latency
	v8 := l.ADCLatencyNS(8)
	if v8 <= v {
		t.Errorf("ADCLatencyNS(8)=%g should be > ADCLatencyNS(4)=%g", v8, v)
	}
}

func TestTotalPipelineNS(t *testing.T) {
	l := NewLatencyModel(64, 64, Node65nm)
	total := l.TotalPipelineNS(4)
	expected := l.DACLatencyNS() + l.CrossbarSettlingNS() + l.ADCLatencyNS(4)
	if total != expected {
		t.Errorf("TotalPipelineNS(4) = %g, want %g", total, expected)
	}
	if total <= 0 {
		t.Errorf("TotalPipelineNS(4) = %g, want > 0", total)
	}
}

func TestLatency_FasterNode(t *testing.T) {
	l65 := NewLatencyModel(64, 64, Node65nm)
	l14 := NewLatencyModel(64, 64, Node14nm)
	if l14.TotalPipelineNS(4) >= l65.TotalPipelineNS(4) {
		t.Errorf("14nm total latency %g should be < 65nm %g",
			l14.TotalPipelineNS(4), l65.TotalPipelineNS(4))
	}
}
