//go:build !cgo

package gogpuapp

import "github.com/gogpu/gg"

func newOffscreenContext(width, height int) *gg.Context {
	dc := gg.NewContext(width, height)
	dc.SetRasterizerMode(gg.RasterizerAnalytic)
	dc.SetTextMode(gg.TextModeBitmap)
	return dc
}
