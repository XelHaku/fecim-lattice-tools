
## Educational Tooltips (2026-01-25)

### Pattern Used
Implemented info icon (ⓘ) tooltips using Fyne's dialog.ShowInformation:
```go
btn := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
    dialog.ShowInformation("Title", "Content", window)
})
btn.Importance = widget.LowImportance
```

### Why This Pattern Works
1. **Click-to-show** (not hover) - better for both desktop and touch
2. **LowImportance** styling - subtle, doesn't dominate UI
3. **Consistent 2-3 sentence format** - educational but not overwhelming
4. **Specific values included** - e.g., "Ec ≈ 1.0-1.5 MV/cm for HfO₂-ZrO₂"

### Locations Added
- **Module 1 Hysteresis**: Ec, Pr, 30 Levels (material params section)
- **Module 2 Crossbar**: 30 Levels (footer info)
- **Module 4 Circuits**: DAC, TIA, ADC, Ec (operations tab data paths)

### Key Physics Covered
- Coercive Field (Ec) - switching threshold
- Remanent Polarization (Pr) - non-volatile retention
- 30 Levels - analog storage advantage (4.9 bits/cell)
- DAC/ADC/TIA - peripheral circuits for WRITE/READ

### User Experience Benefit
Transforms technical terms into accessible explanations without cluttering the UI. Users can learn on-demand by clicking ⓘ icons near unfamiliar terms.
