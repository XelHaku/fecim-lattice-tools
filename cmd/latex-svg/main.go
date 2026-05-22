// Command latex-svg converts a LaTeX equation file into an SVG image.
package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
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
	os.Exit(runLatexSVG(os.Args[1:], os.Stdout, os.Stderr))
}

func runLatexSVG(args []string, stdout, stderr io.Writer) int {
	var opts options
	flags := flag.NewFlagSet("latex-svg", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&opts.inputPath, "in", "", "Path to LaTeX equation file (required)")
	flags.StringVar(&opts.outputPath, "out", "", "Output SVG path (default: input basename + .svg)")
	flags.StringVar(&opts.preamblePath, "preamble", "", "Optional LaTeX preamble file to include when wrapping")
	flags.BoolVar(&opts.forceWrap, "force-wrap", false, "Wrap input in a minimal document even if it looks like a full LaTeX file")
	flags.BoolVar(&opts.inlineMath, "inline", false, "Wrap equation as inline math (\\(...\\)) instead of display math (\\[...\\])")
	flags.BoolVar(&opts.keepTemp, "keep", false, "Keep intermediate TeX/DVI files for debugging")
	flags.StringVar(&opts.latexBinary, "latex", "latex", "LaTeX binary to use (latex, xelatex, etc.)")
	flags.StringVar(&opts.dvisvgmBinary, "dvisvgm", "dvisvgm", "dvisvgm binary to use")
	flags.BoolVar(&opts.useFonts, "fonts", false, "Keep fonts in SVG (default converts glyphs to paths)")
	flags.StringVar(&opts.bboxMode, "bbox", "min", "dvisvgm bbox mode (min, preview, exact, etc.)")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if opts.inputPath == "" {
		fmt.Fprintln(stderr, "Error: -in is required")
		flags.Usage()
		return 2
	}

	inputData, err := os.ReadFile(opts.inputPath)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read input file: %v\n", err)
		return 1
	}
	if len(bytes.TrimSpace(inputData)) == 0 {
		fmt.Fprintln(stderr, "Error: input file is empty")
		flags.Usage()
		return 2
	}

	preamble, err := readOptionalPreamble(opts.preamblePath)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read preamble file: %v\n", err)
		return 1
	}

	texContent := prepareTex(string(inputData), preamble, opts)

	if opts.outputPath == "" {
		opts.outputPath = defaultOutputPath(opts.inputPath)
	}
	if absPath, err := filepath.Abs(opts.outputPath); err == nil {
		opts.outputPath = absPath
	}

	if err := os.MkdirAll(filepath.Dir(opts.outputPath), 0755); err != nil {
		fmt.Fprintf(stderr, "failed to create output directory: %v\n", err)
		return 1
	}

	tempDir, cleanup, err := createWorkDir(opts.keepTemp)
	if err != nil {
		fmt.Fprintf(stderr, "failed to create temp dir: %v\n", err)
		return 1
	}
	defer cleanup()

	texFile := filepath.Join(tempDir, "equation.tex")
	if err := os.WriteFile(texFile, []byte(texContent), 0644); err != nil {
		fmt.Fprintf(stderr, "failed to write temp tex: %v\n", err)
		return 1
	}

	if err := ensureBinary(opts.latexBinary); err != nil {
		fmt.Fprintf(stderr, "latex binary not found: %v\n", err)
		return 1
	}
	if err := ensureBinary(opts.dvisvgmBinary); err != nil {
		fmt.Fprintf(stderr, "dvisvgm binary not found: %v\n", err)
		return 1
	}

	dviFile, err := runLatex(opts, tempDir, texFile)
	if err != nil {
		fmt.Fprintf(stderr, "latex failed: %v\n", err)
		return 1
	}

	if err := runDvisvgm(opts, dviFile, opts.outputPath); err != nil {
		fmt.Fprintf(stderr, "dvisvgm failed: %v\n", err)
		return 1
	}
	if err := normalizeSVGViewBox(opts.outputPath); err != nil {
		fmt.Fprintf(stderr, "failed to normalize svg viewBox: %v\n", err)
		return 1
	}
	if err := inlineSVGUses(opts.outputPath); err != nil {
		fmt.Fprintf(stderr, "failed to inline svg uses: %v\n", err)
		return 1
	}
	if err := sanitizeSVGRoot(opts.outputPath); err != nil {
		fmt.Fprintf(stderr, "failed to sanitize svg root: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "SVG written to %s\n", opts.outputPath)
	if opts.keepTemp {
		fmt.Fprintf(stdout, "Temp files kept at %s\n", tempDir)
	}
	return 0
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
	b.WriteString("\\documentclass{article}\n")
	b.WriteString("\\usepackage{amsmath,amssymb}\n")
	b.WriteString("\\usepackage{bm}\n")
	b.WriteString("\\pagestyle{empty}\n")
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

var viewBoxRe = regexp.MustCompile(`viewBox=['"]([^'"]+)['"]`)
var svgRootRe = regexp.MustCompile(`(?s)<svg[^>]*>`)

type svgPathDef struct {
	d     string
	attrs []xml.Attr
}

func normalizeSVGViewBox(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)
	match := viewBoxRe.FindStringSubmatch(content)
	if len(match) < 2 {
		return fmt.Errorf("viewBox not found")
	}
	fields := strings.Fields(match[1])
	if len(fields) != 4 {
		return fmt.Errorf("unexpected viewBox format: %q", match[1])
	}
	minX, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return fmt.Errorf("invalid viewBox minX: %w", err)
	}
	minY, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return fmt.Errorf("invalid viewBox minY: %w", err)
	}
	width, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return fmt.Errorf("invalid viewBox width: %w", err)
	}
	height, err := strconv.ParseFloat(fields[3], 64)
	if err != nil {
		return fmt.Errorf("invalid viewBox height: %w", err)
	}

	if math.Abs(minX) < 1e-6 && math.Abs(minY) < 1e-6 {
		return nil
	}

	newViewBox := fmt.Sprintf("viewBox='0 0 %s %s'", formatFloat(width), formatFloat(height))
	content = viewBoxRe.ReplaceAllString(content, newViewBox)

	insertAfter := "</defs>"
	translate := fmt.Sprintf("<g transform='translate(%s %s)'>", formatFloat(-minX), formatFloat(-minY))
	if strings.Contains(content, insertAfter) {
		content = strings.Replace(content, insertAfter, insertAfter+"\n"+translate, 1)
		content = strings.Replace(content, "</svg>", "</g>\n</svg>", 1)
	} else {
		return fmt.Errorf("defs block not found for translation insert")
	}

	return os.WriteFile(path, []byte(content), 0644)
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func inlineSVGUses(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	dec := xml.NewDecoder(bytes.NewReader(data))
	var out bytes.Buffer
	enc := xml.NewEncoder(&out)

	defs := map[string]svgPathDef{}
	inDefs := false
	skipUseEnd := false

	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "defs" {
				inDefs = true
				if err := enc.EncodeToken(t); err != nil {
					return err
				}
				continue
			}
			if inDefs {
				if t.Name.Local == "path" {
					id := attrValue(t.Attr, "id")
					d := attrValue(t.Attr, "d")
					if id != "" && d != "" {
						defs[id] = svgPathDef{d: d, attrs: t.Attr}
					}
				}
				if err := enc.EncodeToken(t); err != nil {
					return err
				}
				continue
			}
			if t.Name.Local == "use" {
				href := attrValueNS(t.Attr, "http://www.w3.org/1999/xlink", "href")
				if href == "" {
					href = attrValue(t.Attr, "href")
				}
				if strings.HasPrefix(href, "#") {
					href = href[1:]
				}
				if def, ok := defs[href]; ok {
					x := attrValue(t.Attr, "x")
					y := attrValue(t.Attr, "y")
					transform := attrValue(t.Attr, "transform")
					translate := ""
					if x != "" || y != "" {
						if x == "" {
							x = "0"
						}
						if y == "" {
							y = "0"
						}
						translate = fmt.Sprintf("translate(%s %s)", x, y)
					}
					if translate != "" {
						if transform != "" {
							transform = translate + " " + transform
						} else {
							transform = translate
						}
					}

					attrs := []xml.Attr{{Name: xml.Name{Local: "d"}, Value: def.d}}
					for _, attr := range t.Attr {
						if attr.Name.Local == "x" || attr.Name.Local == "y" || attr.Name.Local == "href" {
							continue
						}
						if attr.Name.Space == "http://www.w3.org/1999/xlink" && attr.Name.Local == "href" {
							continue
						}
						if attr.Name.Local == "transform" {
							continue
						}
						attrs = append(attrs, attr)
					}
					if transform != "" {
						attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "transform"}, Value: transform})
					}
					if len(def.attrs) > 0 {
						attrs = appendUniqueAttrs(attrs, inheritPathAttrs(def.attrs))
					}
					start := xml.StartElement{Name: xml.Name{Local: "path"}, Attr: attrs}
					if err := enc.EncodeToken(start); err != nil {
						return err
					}
					if err := enc.EncodeToken(xml.EndElement{Name: start.Name}); err != nil {
						return err
					}
					skipUseEnd = true
					continue
				}
			}
			if err := enc.EncodeToken(t); err != nil {
				return err
			}
		case xml.EndElement:
			if skipUseEnd && t.Name.Local == "use" {
				skipUseEnd = false
				continue
			}
			if inDefs && t.Name.Local == "defs" {
				inDefs = false
			}
			if err := enc.EncodeToken(t); err != nil {
				return err
			}
		default:
			if err := enc.EncodeToken(tok); err != nil {
				return err
			}
		}
	}

	if err := enc.Flush(); err != nil {
		return err
	}
	return os.WriteFile(path, out.Bytes(), 0644)
}

func attrValue(attrs []xml.Attr, name string) string {
	for _, attr := range attrs {
		if attr.Name.Local == name {
			return attr.Value
		}
	}
	return ""
}

func attrValueNS(attrs []xml.Attr, space, name string) string {
	for _, attr := range attrs {
		if attr.Name.Space == space && attr.Name.Local == name {
			return attr.Value
		}
	}
	return ""
}

func inheritPathAttrs(attrs []xml.Attr) []xml.Attr {
	var inherited []xml.Attr
	for _, attr := range attrs {
		if attr.Name.Local == "id" || attr.Name.Local == "d" {
			continue
		}
		inherited = append(inherited, attr)
	}
	return inherited
}

func appendUniqueAttrs(base []xml.Attr, extra []xml.Attr) []xml.Attr {
	seen := map[string]struct{}{}
	for _, attr := range base {
		key := attr.Name.Space + ":" + attr.Name.Local
		seen[key] = struct{}{}
	}
	for _, attr := range extra {
		key := attr.Name.Space + ":" + attr.Name.Local
		if _, ok := seen[key]; ok {
			continue
		}
		base = append(base, attr)
		seen[key] = struct{}{}
	}
	return base
}

func sanitizeSVGRoot(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)
	match := svgRootRe.FindString(content)
	if match == "" {
		return fmt.Errorf("svg root not found")
	}

	version := attrValueFromTag(match, "version")
	width := attrValueFromTag(match, "width")
	height := attrValueFromTag(match, "height")
	viewBox := attrValueFromTag(match, "viewBox")

	if width == "" || height == "" || viewBox == "" {
		return fmt.Errorf("svg root missing width/height/viewBox")
	}

	if version == "" {
		version = "1.1"
	}

	root := fmt.Sprintf("<svg version='%s' xmlns='http://www.w3.org/2000/svg' xmlns:xlink='http://www.w3.org/1999/xlink' width='%s' height='%s' viewBox='%s'>", version, width, height, viewBox)
	content = strings.Replace(content, match, root, 1)
	return os.WriteFile(path, []byte(content), 0644)
}

func attrValueFromTag(tag, name string) string {
	re := regexp.MustCompile(name + `=['"]([^'"]+)['"]`)
	m := re.FindStringSubmatch(tag)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}
