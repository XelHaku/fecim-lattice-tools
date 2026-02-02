// Command latex-svg converts a LaTeX equation file into an SVG image.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type options struct {
	inputPath     string
	outputPath    string
	preamblePath  string
	forceWrap     bool
	inlineMath    bool
	keepTemp      bool
	latexBinary   string
	dvisvgmBinary string
	useFonts      bool
	bboxMode      string
}

func main() {
	opts := parseFlags()
	if opts.inputPath == "" {
		printUsageAndExit("-in is required")
	}

	inputData, err := os.ReadFile(opts.inputPath)
	if err != nil {
		exitError("failed to read input file", err)
	}
	if len(bytes.TrimSpace(inputData)) == 0 {
		printUsageAndExit("input file is empty")
	}

	preamble, err := readOptionalPreamble(opts.preamblePath)
	if err != nil {
		exitError("failed to read preamble file", err)
	}

	texContent := prepareTex(string(inputData), preamble, opts)

	if opts.outputPath == "" {
		opts.outputPath = defaultOutputPath(opts.inputPath)
	}

	if err := os.MkdirAll(filepath.Dir(opts.outputPath), 0755); err != nil {
		exitError("failed to create output directory", err)
	}

	tempDir, cleanup, err := createWorkDir(opts.keepTemp)
	if err != nil {
		exitError("failed to create temp dir", err)
	}
	defer cleanup()

	texFile := filepath.Join(tempDir, "equation.tex")
	if err := os.WriteFile(texFile, []byte(texContent), 0644); err != nil {
		exitError("failed to write temp tex", err)
	}

	if err := ensureBinary(opts.latexBinary); err != nil {
		exitError("latex binary not found", err)
	}
	if err := ensureBinary(opts.dvisvgmBinary); err != nil {
		exitError("dvisvgm binary not found", err)
	}

	dviFile, err := runLatex(opts, tempDir, texFile)
	if err != nil {
		exitError("latex failed", err)
	}

	if err := runDvisvgm(opts, dviFile, opts.outputPath); err != nil {
		exitError("dvisvgm failed", err)
	}

	fmt.Printf("SVG written to %s\n", opts.outputPath)
	if opts.keepTemp {
		fmt.Printf("Temp files kept at %s\n", tempDir)
	}
}

func parseFlags() options {
	var opts options
	flag.StringVar(&opts.inputPath, "in", "", "Path to LaTeX equation file (required)")
	flag.StringVar(&opts.outputPath, "out", "", "Output SVG path (default: input basename + .svg)")
	flag.StringVar(&opts.preamblePath, "preamble", "", "Optional LaTeX preamble file to include when wrapping")
	flag.BoolVar(&opts.forceWrap, "force-wrap", false, "Wrap input in a minimal document even if it looks like a full LaTeX file")
	flag.BoolVar(&opts.inlineMath, "inline", false, "Wrap equation as inline math (\\(...\\)) instead of display math (\\[...\\])")
	flag.BoolVar(&opts.keepTemp, "keep", false, "Keep intermediate TeX/DVI files for debugging")
	flag.StringVar(&opts.latexBinary, "latex", "latex", "LaTeX binary to use (latex, xelatex, etc.)")
	flag.StringVar(&opts.dvisvgmBinary, "dvisvgm", "dvisvgm", "dvisvgm binary to use")
	flag.BoolVar(&opts.useFonts, "fonts", false, "Keep fonts in SVG (default converts glyphs to paths)")
	flag.StringVar(&opts.bboxMode, "bbox", "min", "dvisvgm bbox mode (min, preview, exact, etc.)")
	flag.Parse()
	return opts
}

func printUsageAndExit(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, "Error: %s\n\n", msg)
	}
	flag.Usage()
	os.Exit(2)
}

func exitError(prefix string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", prefix, err)
	os.Exit(1)
}

func readOptionalPreamble(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func prepareTex(raw, preamble string, opts options) string {
	if !opts.forceWrap && isFullDocument(raw) {
		return raw
	}
	return wrapLatex(raw, preamble, opts.inlineMath)
}

func isFullDocument(raw string) bool {
	return strings.Contains(raw, "\\documentclass") || strings.Contains(raw, "\\begin{document}")
}

func wrapLatex(body, extraPreamble string, inline bool) string {
	var b strings.Builder
	b.WriteString("\\documentclass[preview]{standalone}\n")
	b.WriteString("\\usepackage{amsmath,amssymb}\n")
	b.WriteString("\\usepackage{bm}\n")
	if extraPreamble != "" {
		b.WriteString(strings.TrimSpace(extraPreamble))
		b.WriteString("\n")
	}
	b.WriteString("\\begin{document}\n")
	if inline {
		b.WriteString("\\(\n")
	} else {
		b.WriteString("\\[\n")
	}
	b.WriteString(strings.TrimSpace(body))
	b.WriteString("\n")
	if inline {
		b.WriteString("\\)\n")
	} else {
		b.WriteString("\\]\n")
	}
	b.WriteString("\\end{document}\n")
	return b.String()
}

func defaultOutputPath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	if ext == "" {
		return inputPath + ".svg"
	}
	return strings.TrimSuffix(inputPath, ext) + ".svg"
}

func createWorkDir(keep bool) (string, func(), error) {
	workDir, err := os.MkdirTemp("", "latex-svg-")
	if err != nil {
		return "", func() {}, err
	}
	cleanup := func() {
		if keep {
			return
		}
		_ = os.RemoveAll(workDir)
	}
	return workDir, cleanup, nil
}

func ensureBinary(name string) error {
	_, err := exec.LookPath(name)
	return err
}

func runLatex(opts options, workDir, texFile string) (string, error) {
	args := []string{
		"-interaction=nonstopmode",
		"-halt-on-error",
		"-output-directory", workDir,
		texFile,
	}
	cmd := exec.Command(opts.latexBinary, args...)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v\n%s", err, string(output))
	}
	base := strings.TrimSuffix(filepath.Base(texFile), filepath.Ext(texFile))
	dviFile := filepath.Join(workDir, base+".dvi")
	if _, err := os.Stat(dviFile); err != nil {
		return "", fmt.Errorf("missing DVI output: %v", err)
	}
	return dviFile, nil
}

func runDvisvgm(opts options, dviFile, outputPath string) error {
	args := []string{}
	if !opts.useFonts {
		args = append(args, "--no-fonts")
	}
	if opts.bboxMode != "" {
		args = append(args, "--bbox="+opts.bboxMode)
	}
	args = append(args, "-o", outputPath, dviFile)
	cmd := exec.Command(opts.dvisvgmBinary, args...)
	cmd.Dir = filepath.Dir(dviFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v\n%s", err, string(output))
	}
	return nil
}
