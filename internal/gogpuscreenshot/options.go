package gogpuscreenshot

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"
)

const (
	defaultWidth  = 1400
	defaultHeight = 900
)

type Options struct {
	OutputDir string
	Only      string
	Tag       string
	Width     int
	Height    int
}

func DefaultOptions() Options {
	return Options{
		OutputDir: "screenshots",
		Width:     defaultWidth,
		Height:    defaultHeight,
	}
}

func ParseOptions(args []string) (Options, error) {
	opts := DefaultOptions()
	fs := flag.NewFlagSet("fecim-screenshotter", flag.ContinueOnError)
	fs.StringVar(&opts.OutputDir, "out", opts.OutputDir, "output directory for screenshots")
	fs.StringVar(&opts.Only, "only", opts.Only, "capture only a single module")
	fs.StringVar(&opts.Tag, "tag", opts.Tag, "filename tag suffix")
	fs.IntVar(&opts.Width, "w", opts.Width, "output image width")
	fs.IntVar(&opts.Height, "h", opts.Height, "output image height")
	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}
	opts.OutputDir = filepath.Clean(strings.TrimSpace(opts.OutputDir))
	opts.Only = normalizeModuleName(opts.Only)
	opts.Tag = strings.TrimSpace(opts.Tag)
	if opts.OutputDir == "." || opts.OutputDir == "" {
		opts.OutputDir = "screenshots"
	}
	if opts.Width <= 0 || opts.Height <= 0 {
		return Options{}, fmt.Errorf("screenshot dimensions must be positive, got %dx%d", opts.Width, opts.Height)
	}
	return opts, nil
}

func (o Options) OutputPath(filename string) string {
	name := filename
	if tag := strings.TrimSpace(o.Tag); tag != "" {
		ext := filepath.Ext(filename)
		stem := strings.TrimSuffix(filename, ext)
		name = stem + "_" + tag + ext
	}
	return filepath.Join(o.OutputDir, name)
}

func (o Options) Matches(module string) bool {
	only := normalizeModuleName(o.Only)
	return only == "" || only == normalizeModuleName(module)
}

func normalizeModuleName(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "all":
		return ""
	case "1", "m1", "module1":
		return "hysteresis"
	case "2", "m2", "module2":
		return "crossbar"
	case "3", "m3", "module3":
		return "mnist"
	case "4", "m4", "module4":
		return "circuits"
	case "5", "m5", "module5":
		return "comparison"
	case "6", "m6", "module6":
		return "eda"
	case "7", "m7", "module7", "documentation":
		return "docs"
	default:
		return strings.ToLower(strings.TrimSpace(raw))
	}
}
