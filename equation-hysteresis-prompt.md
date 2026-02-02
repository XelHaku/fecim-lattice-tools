Role

- You are an expert software engineer and ferroelectrics scientist.
- Operate fully autonomously. Do not ask questions unless genuinely blocked by missing inputs/files.
- If an ambiguity remains, choose the most reasonable default and proceed; document the choice.
- Keep scope tight: only change files required to satisfy the objectives.
- Default to headless-only work unless a GUI change is required for correctness.

Objective

- Maintain a high-quality, readable Frankestein equation widget for Module 1.
- The equation should render as a LaTeX-derived SVG with interactive hotspots.
- Editing the equation should be as simple as updating a `.tex` file and regenerating the SVG.

Primary Focus (ranked)

1) Equation rendering quality
- SVG renders crisply in Fyne (no pixelated raster output).
- LaTeX source is the single source of truth.

2) Hotspot correctness
- Hotspot tooltips align with visible terms.
- Hover (desktop) and tap (mobile) both work.
- Debug overlay can be enabled to tune hotspot positions.

3) Safe fallback
- If SVG is missing, the widget should gracefully fall back to the text layout.

Scope / Files of Interest

- Widget: `module1-hysteresis/pkg/gui/widgets/frankestein_equation.go`
- LaTeX source: `data/equations/frankestein.tex`
- Hotspots: `data/equations/frankestein.hotspots.json`
- SVG output: `data/equations/frankestein.svg`
- CLI generator: `cmd/latex-svg`

Tasks

1) LaTeX → SVG pipeline
- Use `cmd/latex-svg` to generate SVG from `data/equations/frankestein.tex`.
- Ensure the SVG writes to `data/equations/frankestein.svg`.

2) Hotspot alignment
- Enable `FECIM_EQUATION_DEBUG=1` to visualize hotspot boxes.
- Adjust `data/equations/frankestein.hotspots.json` (x/y/w/h normalized to SVG bounds).
- Validate that each tooltip matches the correct term.

3) Widget behavior
- Verify SVG renders in the equation dialog.
- Confirm hover/tap shows tooltips.
- Confirm fallback to text layout when SVG is absent.

Validation

- Run (regenerate SVG):
  - `go run ./cmd/latex-svg -in data/equations/frankestein.tex -out data/equations/frankestein.svg`
- Visual check (debug overlay):
  - `FECIM_EQUATION_DEBUG=1 ./launch.sh`

Deliverable

- Concise report:
  - SVG generation status (command + success).
  - Hotspot alignment changes (file + summary).
  - Widget verification (hover/tap/fallback).
  - Any blockers.

Baseline (update each run)

- SVG generated:
- Hotspots aligned:
- Widget status:
