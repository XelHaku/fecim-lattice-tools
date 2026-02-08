package widgets

import (
	"testing"

	"fyne.io/fyne/v2"
)

func TestNewKeyStat(t *testing.T) {
	config := KeyStatConfig{
		Label:   "Accuracy",
		Value:   "95.5%",
		MinSize: fyne.NewSize(120, 60),
	}
	ks := NewKeyStat(config)

	if ks == nil {
		t.Fatal("NewKeyStat returned nil")
	}
	if ks.GetValue() != "95.5%" {
		t.Errorf("Expected value '95.5%%', got '%s'", ks.GetValue())
	}
	if ks.MinSize() != config.MinSize {
		t.Errorf("Expected MinSize %v, got %v", config.MinSize, ks.MinSize())
	}
}

func TestKeyStat_SetValue(t *testing.T) {
	ks := NewKeyStat(KeyStatConfig{Value: "0"})

	ks.SetValue("100")
	if ks.GetValue() != "100" {
		t.Errorf("Expected value '100', got '%s'", ks.GetValue())
	}

	ks.SetValue("N/A")
	if ks.GetValue() != "N/A" {
		t.Errorf("Expected value 'N/A', got '%s'", ks.GetValue())
	}
}

func TestKeyStat_SetLabel(t *testing.T) {
	ks := NewKeyStat(KeyStatConfig{Label: "Original"})

	ks.SetLabel("Updated")
	// Note: We can't directly check the label from the public interface
	// but we can verify the method doesn't panic
}

func TestKeyStat_SetLabelAndValue(t *testing.T) {
	ks := NewKeyStat(KeyStatConfig{Label: "A", Value: "1"})

	ks.SetLabelAndValue("B", "2")
	if ks.GetValue() != "2" {
		t.Errorf("Expected value '2', got '%s'", ks.GetValue())
	}
}

func TestKeyStat_DefaultConfig(t *testing.T) {
	ks := NewKeyStat(KeyStatConfig{})

	if ks.MinSize().Width <= 0 || ks.MinSize().Height <= 0 {
		t.Error("Default MinSize should be positive")
	}
	if ks.GetValue() == "" {
		t.Error("Default value should not be empty")
	}
}

func TestNewKeyStatGroup(t *testing.T) {
	group := NewKeyStatGroup()

	if group == nil {
		t.Fatal("NewKeyStatGroup returned nil")
	}

	stats := group.All()
	if len(stats) != 0 {
		t.Errorf("Expected 0 stats, got %d", len(stats))
	}
}

func TestKeyStatGroup_Add(t *testing.T) {
	group := NewKeyStatGroup()

	stat1 := group.Add("accuracy", KeyStatConfig{Label: "Accuracy", Value: "90%"})
	stat2 := group.Add("loss", KeyStatConfig{Label: "Loss", Value: "0.1"})

	if stat1 == nil || stat2 == nil {
		t.Fatal("Add should return the created stat")
	}

	stats := group.All()
	if len(stats) != 2 {
		t.Errorf("Expected 2 stats, got %d", len(stats))
	}
}

func TestKeyStatGroup_Get(t *testing.T) {
	group := NewKeyStatGroup()
	group.Add("test", KeyStatConfig{Label: "Test", Value: "123"})

	stat := group.Get("test")
	if stat == nil {
		t.Fatal("Get should return the stat")
	}
	if stat.GetValue() != "123" {
		t.Errorf("Expected value '123', got '%s'", stat.GetValue())
	}

	// Non-existent stat
	nilStat := group.Get("nonexistent")
	if nilStat != nil {
		t.Error("Get should return nil for non-existent stat")
	}
}

func TestKeyStatGroup_SetValue(t *testing.T) {
	group := NewKeyStatGroup()
	group.Add("counter", KeyStatConfig{Label: "Count", Value: "0"})

	group.SetValue("counter", "42")

	stat := group.Get("counter")
	if stat.GetValue() != "42" {
		t.Errorf("Expected value '42', got '%s'", stat.GetValue())
	}

	// SetValue on non-existent stat should not panic
	group.SetValue("nonexistent", "value")
}

func TestKeyStatGroup_AsContainer(t *testing.T) {
	group := NewKeyStatGroup()
	group.Add("a", KeyStatConfig{Label: "A"})
	group.Add("b", KeyStatConfig{Label: "B"})

	container := group.AsContainer()
	if container == nil {
		t.Fatal("AsContainer should return a container")
	}
}
