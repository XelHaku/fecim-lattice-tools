// Package equations embeds equation SVG and hotspot assets for zero-cost access.
package equations

import _ "embed"

//go:embed frankestein.svg
var LkEquationSVG []byte

//go:embed preisach.svg
var PreisachEquationSVG []byte

//go:embed frankestein.hotspots.json
var LkHotspotsJSON []byte
