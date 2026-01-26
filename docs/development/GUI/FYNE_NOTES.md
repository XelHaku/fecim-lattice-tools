# Fyne Development Notes

Best practices for Fyne GUI development in FeCIM Lattice Tools.

## Layout Best Practices

### Container Types & When to Use

| Container | Use When |
|-----------|----------|
| VBox/HBox | Stacking widgets vertically/horizontally |
| Border | App layouts with header/footer/sidebar |
| Grid | Uniform grid of same-sized elements |
| AdaptiveGrid | Responsive card layouts |
| Stack | Layering elements (overlays, backgrounds) |
| NewWithoutLayout | Only for pixel-perfect canvas drawings |

### Scroll Containers

- Use `container.NewVScroll()` for content that may exceed viewport
- Set MinSize on scroll container, not child content
- Types: `NewScroll` (both), `NewVScroll` (vertical), `NewHScroll` (horizontal)

### Responsive Grids

```go
// Adapts automatically to available width
container.NewAdaptiveGrid(2, card1, card2, card3, card4)
```

## Canvas-Based Drawings

### Sizing Rules
- Always call `container.Resize()` with explicit dimensions
- Account for labels/legends in size calculations
- Add 20-40px padding around diagram content

### Coordinate System
- Origin: Top-left at (0, 0)
- All positions relative to parent
- Baseline: 120 DPI (1 unit = 1px at 120 DPI)

### Text Positioning
```go
// Center text horizontally in a box
textWidth := float32(len(text)) * charWidth  // ~6-7px per char
textX := boxX + (boxWidth - textWidth) / 2
```

## Text Handling

### Wrapping
```go
label := widget.NewLabel("Long text...")
label.Wrapping = fyne.TextWrapWord  // Always set for long text
```

| Mode | Behavior |
|------|----------|
| TextWrapOff | Extends widget width |
| TextWrapWord | Breaks at word boundaries |
| TextWrapBreak | Breaks at any character |

### Truncation (v2.4+)
```go
label.Truncation = fyne.TextTruncateEllipsis  // Adds "..." when cut
```

## Threading (CRITICAL)

### fyne.Do() Rule

**USE fyne.Do() when:**
- Your code creates a goroutine with `go`
- That goroutine needs to update UI

**DON'T use fyne.Do() when:**
- In Fyne callbacks (OnTapped, OnChanged, etc.)
- In event receivers
- In main application setup

```go
// CORRECT - background goroutine
go func() {
    result := doExpensiveWork()
    fyne.Do(func() {
        label.SetText(result)
    })
}()

// WRONG - already on main thread
button.OnTapped = func() {
    fyne.Do(func() {  // Unnecessary!
        label.SetText("Clicked")
    })
}
```

## Performance Tips

1. **Batch updates** - Update multiple properties, then Refresh() once
2. **Viewport management** - Only render visible content for large lists
3. **Avoid nested scrolls** - Generally confusing UX
4. **Test with real data** - Profile with realistic dataset sizes

## Common Pitfalls

1. **Thread safety** - Always use fyne.Do() from goroutines
2. **Manual positioning** - Prefer layouts over NewWithoutLayout
3. **Excessive Refresh()** - Only call when data actually changes
4. **Hardcoded sizes** - Use MinSize() and relative sizing

## References

- [Fyne Documentation](https://docs.fyne.io/)
- [Layout Guide](https://docs.fyne.io/explore/layouts.html)
- [Container API](https://docs.fyne.io/explore/container.html)
