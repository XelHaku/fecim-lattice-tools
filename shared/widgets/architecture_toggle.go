// Package widgets provides reusable UI components.
package widgets

import "fyne.io/fyne/v2/widget"

// ArchitectureToggleStyle controls how the toggle buttons are labeled.
type ArchitectureToggleStyle int

const (
	// ArchitectureToggleStylePlain keeps labels unchanged.
	ArchitectureToggleStylePlain ArchitectureToggleStyle = iota
	// ArchitectureToggleStyleBullet prefixes the selected label with a bullet.
	ArchitectureToggleStyleBullet
)

// ArchitectureToggleOptions configures the architecture toggle buttons.
type ArchitectureToggleOptions struct {
	Initial      string
	Style        ArchitectureToggleStyle
	LabelPassive string
	Label1T1R    string
	Label2T1R    string
	OnChanged    func(architecture string)
}

// ArchitectureToggle provides a shared PASSIVE/1T1R/2T1R button group.
type ArchitectureToggle struct {
	PassiveButton *widget.Button
	OneT1RButton  *widget.Button
	TwoT1RButton  *widget.Button
	Update        func(architecture string)
}

// NewArchitectureToggle creates a new architecture toggle button group.
func NewArchitectureToggle(opts ArchitectureToggleOptions) *ArchitectureToggle {
	initial := opts.Initial
	if initial == "" {
		initial = Architecture0T1R
	}

	labelPassive := opts.LabelPassive
	if labelPassive == "" {
		labelPassive = "PASSIVE"
	}
	label1T1R := opts.Label1T1R
	if label1T1R == "" {
		label1T1R = "1T1R"
	}
	label2T1R := opts.Label2T1R
	if label2T1R == "" {
		label2T1R = "2T1R"
	}

	current := initial

	setLabel := func(btn *widget.Button, label string, selected bool) {
		if opts.Style == ArchitectureToggleStyleBullet && selected {
			btn.SetText("● " + label)
			return
		}
		btn.SetText(label)
	}

	passiveBtn := widget.NewButton(labelPassive, nil)
	oneBtn := widget.NewButton(label1T1R, nil)
	twoBtn := widget.NewButton(label2T1R, nil)

	apply := func(arch string) {
		current = arch
		passiveBtn.Importance = widget.LowImportance
		oneBtn.Importance = widget.LowImportance
		twoBtn.Importance = widget.LowImportance

		switch arch {
		case Architecture1T1R:
			oneBtn.Importance = widget.HighImportance
		case Architecture2T1R:
			twoBtn.Importance = widget.HighImportance
		default:
			passiveBtn.Importance = widget.HighImportance
		}

		setLabel(passiveBtn, labelPassive, arch == Architecture0T1R)
		setLabel(oneBtn, label1T1R, arch == Architecture1T1R)
		setLabel(twoBtn, label2T1R, arch == Architecture2T1R)

		passiveBtn.Refresh()
		oneBtn.Refresh()
		twoBtn.Refresh()
	}

	passiveBtn.OnTapped = func() {
		if current == Architecture0T1R {
			return
		}
		apply(Architecture0T1R)
		if opts.OnChanged != nil {
			opts.OnChanged(Architecture0T1R)
		}
	}

	oneBtn.OnTapped = func() {
		if current == Architecture1T1R {
			return
		}
		apply(Architecture1T1R)
		if opts.OnChanged != nil {
			opts.OnChanged(Architecture1T1R)
		}
	}

	twoBtn.OnTapped = func() {
		if current == Architecture2T1R {
			return
		}
		apply(Architecture2T1R)
		if opts.OnChanged != nil {
			opts.OnChanged(Architecture2T1R)
		}
	}

	apply(initial)

	return &ArchitectureToggle{
		PassiveButton: passiveBtn,
		OneT1RButton:  oneBtn,
		TwoT1RButton:  twoBtn,
		Update:        apply,
	}
}
