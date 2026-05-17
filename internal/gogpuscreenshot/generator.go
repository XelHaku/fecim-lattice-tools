//go:build !cgo

package gogpuscreenshot

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"fecim-lattice-tools/internal/gogpuapp"
	"fecim-lattice-tools/shared/viewmodel"
)

type appFrameScreenshot struct {
	module   string
	id       viewmodel.ModuleID
	filename string
}

var screenshotFilenamesByModule = map[viewmodel.ModuleID]string{
	viewmodel.ModuleHysteresis: "hysteresis-p-e-loop.png",
	viewmodel.ModuleCrossbar:   "crossbar-heatmap-8x8.png",
	viewmodel.ModuleMNIST:      "mnist-accuracy-sweep.png",
	viewmodel.ModuleCircuits:   "circuits-ispp-convergence.png",
	viewmodel.ModuleComparison: "comparison-architecture-bars.png",
	viewmodel.ModuleEDA:        "eda-design-overview.png",
	viewmodel.ModuleDocs:       "docs-overview.png",
}

func Run(args []string) error {
	opts, err := ParseOptions(args)
	if err != nil {
		return err
	}
	return Generate(opts)
}

func Generate(opts Options) error {
	if opts.Width <= 0 || opts.Height <= 0 {
		return fmt.Errorf("screenshot dimensions must be positive, got %dx%d", opts.Width, opts.Height)
	}
	if opts.OutputDir == "" {
		opts.OutputDir = DefaultOptions().OutputDir
	}

	screenshots, err := appFrameScreenshots()
	if err != nil {
		return err
	}

	count := 0
	total := matchedScreenshotCount(opts, screenshots)
	for _, screenshot := range screenshots {
		if !opts.Matches(screenshot.module) {
			continue
		}
		count++
		log.Printf("[%d/%d] %s", count, total, screenshot.filename)
		if err := captureAppFrame(opts, screenshot); err != nil {
			return err
		}
	}

	if count == 0 {
		return fmt.Errorf("no screenshots matched -only %q", opts.Only)
	}
	log.Printf("done - %d screens generated in %s/", count, opts.OutputDir)
	return nil
}

func appFrameScreenshots() ([]appFrameScreenshot, error) {
	return buildAppFrameScreenshots(viewmodel.KnownDescriptors())
}

func buildAppFrameScreenshots(descriptors []viewmodel.ModuleDescriptor) ([]appFrameScreenshot, error) {
	screenshots := make([]appFrameScreenshot, 0, len(descriptors))
	for _, descriptor := range descriptors {
		filename, ok := screenshotFilenamesByModule[descriptor.ID]
		if !ok {
			return nil, fmt.Errorf("no screenshot filename mapped for module %q", descriptor.ID)
		}
		screenshots = append(screenshots, appFrameScreenshot{
			module:   string(descriptor.ID),
			id:       descriptor.ID,
			filename: filename,
		})
	}
	return screenshots, nil
}

func matchedScreenshotCount(opts Options, screenshots []appFrameScreenshot) int {
	count := 0
	for _, screenshot := range screenshots {
		if opts.Matches(screenshot.module) {
			count++
		}
	}
	return count
}

func captureAppFrame(opts Options, screenshot appFrameScreenshot) error {
	img, err := gogpuapp.CaptureFrameImage(screenshot.id, opts.Width, opts.Height)
	if err != nil {
		return fmt.Errorf("screenshot: render %s: %w", screenshot.module, err)
	}
	return savePNG(opts.OutputPath(screenshot.filename), img)
}

func savePNG(path string, img image.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("screenshot: mkdir %s: %w", filepath.Dir(path), err)
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("screenshot: create %s: %w", path, err)
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("screenshot: encode %s: %w", path, err)
	}
	return nil
}
