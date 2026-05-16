//go:build !ci
// +build !ci

package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"fyne.io/fyne/v2"
)

func captureCanvasWithX11Fallback(window fyne.Window, windowTitle string) image.Image {
	if window == nil || window.Canvas() == nil {
		return nil
	}
	img := window.Canvas().Capture()
	if !isAllBlackOrTransparent(img) {
		return img
	}
	for i := 0; i < 2; i++ {
		fallback, err := captureWindowWithImport(windowTitle)
		if err == nil && !isAllBlackOrTransparent(fallback) {
			return fallback
		}
		time.Sleep(150 * time.Millisecond)
	}
	return img
}

func isAllBlackOrTransparent(img image.Image) bool {
	if img == nil {
		return true
	}
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bb, a := img.At(x, y).RGBA()
			if r != 0 || g != 0 || bb != 0 || a != 0 {
				return false
			}
		}
	}
	return true
}

func captureWindowWithImport(windowTitle string) (image.Image, error) {
	if strings.TrimSpace(windowTitle) == "" {
		return nil, fmt.Errorf("window title is empty")
	}
	if strings.TrimSpace(os.Getenv("DISPLAY")) == "" {
		return nil, fmt.Errorf("DISPLAY is not set")
	}

	importBin, err := exec.LookPath("import")
	if err != nil {
		return nil, fmt.Errorf("imagemagick import not found: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "fecim-x11-fallback-*.png")
	if err != nil {
		return nil, fmt.Errorf("create temp capture file: %w", err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	defer os.Remove(tmpPath)

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, importBin, "-window", windowTitle, tmpPath)
	cmd.Stdout = io.Discard
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("import capture failed: %s", msg)
	}

	f, err := os.Open(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("open temp capture file: %w", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode import capture: %w", err)
	}
	return img, nil
}
