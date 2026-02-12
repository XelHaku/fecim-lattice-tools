package main

import (
	"encoding/xml"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsFullDocument(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty", "", false},
		{"snippet", "\\rho_{eff} dP/dt = 0", false},
		{"documentclass", "\\documentclass{article}\n\\begin{document}\n", true},
		{"begin_document", "\\begin{document}\nX\\end{document}", true},
	}

	for _, c := range cases {
		if got := isFullDocument(c.input); got != c.want {
			t.Fatalf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}

func TestWrapLatex(t *testing.T) {
	body := "E = mc^2"
	out := wrapLatex(body, "", false)
	if !containsAll(out, []string{"\\documentclass", "\\begin{document}", "\\[", body, "\\]", "\\end{document}"}) {
		t.Fatalf("wrapLatex missing expected content")
	}

	inline := wrapLatex(body, "", true)
	if !containsAll(inline, []string{"\\(", body, "\\)"}) {
		t.Fatalf("wrapLatex inline missing expected content")
	}
}

func TestNormalizeSVGViewBox(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "equation.svg")
	input := `<?xml version='1.0' encoding='UTF-8'?>
<svg version='1.1' xmlns='http://www.w3.org/2000/svg' width='100pt' height='50pt' viewBox='10 20 100 50'>
<defs></defs>
<g id='page1'><rect x='10' y='20' width='20' height='10'/></g>
</svg>`
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatalf("write temp svg: %v", err)
	}

	if err := normalizeSVGViewBox(path); err != nil {
		t.Fatalf("normalizeSVGViewBox: %v", err)
	}

	out, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read normalized svg: %v", err)
	}
	output := string(out)
	if !strings.Contains(output, "viewBox='0 0 100 50'") {
		t.Fatalf("viewBox not normalized: %s", output)
	}
	if !strings.Contains(output, "<g transform='translate(-10 -20)'>") {
		t.Fatalf("missing translate wrapper: %s", output)
	}
	if !strings.Contains(output, "</g>\n</svg>") {
		t.Fatalf("missing wrapper close: %s", output)
	}
}

func TestInlineSVGUses(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "inline.svg")
	input := `<?xml version='1.0' encoding='UTF-8'?>
<svg xmlns='http://www.w3.org/2000/svg' xmlns:xlink='http://www.w3.org/1999/xlink' viewBox='0 0 10 10'>
<defs>
<path id='g1' d='M0 0 L1 0'/>
</defs>
<use x='2' y='3' xlink:href='#g1'/>
</svg>`
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatalf("write temp svg: %v", err)
	}

	if err := inlineSVGUses(path); err != nil {
		t.Fatalf("inlineSVGUses: %v", err)
	}

	out, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read inline svg: %v", err)
	}
	output := string(out)
	if strings.Contains(output, "<use") {
		t.Fatalf("use element still present: %s", output)
	}
	if !strings.Contains(output, "translate(2 3)") {
		t.Fatalf("missing translate transform: %s", output)
	}
	if !strings.Contains(output, "d=\"M0 0 L1 0\"") {
		t.Fatalf("missing inlined path: %s", output)
	}
}

func TestParseFlags(t *testing.T) {
	orig := flag.CommandLine
	defer func() { flag.CommandLine = orig }()

	flag.CommandLine = flag.NewFlagSet("latex-svg", flag.ContinueOnError)
	_ = flag.CommandLine.Parse([]string{})
	os.Args = []string{"latex-svg", "-in", "eq.tex", "-out", "out.svg", "-inline", "-keep", "-latex", "pdflatex", "-dvisvgm", "dvisvgm-custom", "-fonts", "-bbox", "exact", "-force-wrap"}

	opts := parseFlags()
	if opts.inputPath != "eq.tex" || opts.outputPath != "out.svg" || !opts.inlineMath || !opts.keepTemp || !opts.forceWrap {
		t.Fatalf("unexpected parsed options: %+v", opts)
	}
	if opts.latexBinary != "pdflatex" || opts.dvisvgmBinary != "dvisvgm-custom" || !opts.useFonts || opts.bboxMode != "exact" {
		t.Fatalf("unexpected binary/bbox flags: %+v", opts)
	}
}

func TestReadOptionalPreamble(t *testing.T) {
	if got, err := readOptionalPreamble(""); err != nil || got != "" {
		t.Fatalf("empty path: got %q err=%v", got, err)
	}

	dir := t.TempDir()
	p := filepath.Join(dir, "preamble.tex")
	if err := os.WriteFile(p, []byte("\\usepackage{siunitx}\n"), 0644); err != nil {
		t.Fatalf("write preamble: %v", err)
	}
	got, err := readOptionalPreamble(p)
	if err != nil || !strings.Contains(got, "siunitx") {
		t.Fatalf("read preamble: got %q err=%v", got, err)
	}
	if _, err := readOptionalPreamble(filepath.Join(dir, "missing.tex")); err == nil {
		t.Fatalf("expected error for missing preamble")
	}
}

func TestPrepareTex(t *testing.T) {
	full := "\\documentclass{article}\n\\begin{document}x\\end{document}"
	if got := prepareTex(full, "", options{}); got != full {
		t.Fatalf("full document should not be wrapped")
	}
	wrapped := prepareTex("x+y", "\\usepackage{physics}", options{forceWrap: true, inlineMath: true})
	if !containsAll(wrapped, []string{"\\usepackage{physics}", "\\(", "x+y", "\\)"}) {
		t.Fatalf("force wrapped output missing pieces: %s", wrapped)
	}
}

func TestDefaultOutputPath(t *testing.T) {
	if got := defaultOutputPath("eq.tex"); got != "eq.svg" {
		t.Fatalf("got %q", got)
	}
	if got := defaultOutputPath("equation"); got != "equation.svg" {
		t.Fatalf("got %q", got)
	}
}

func TestCreateWorkDirAndEnsureBinary(t *testing.T) {
	dir, cleanup, err := createWorkDir(false)
	if err != nil {
		t.Fatalf("createWorkDir: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("temp dir missing: %v", err)
	}
	cleanup()
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("expected temp dir removed, stat err=%v", err)
	}

	dir2, cleanup2, err := createWorkDir(true)
	if err != nil {
		t.Fatalf("createWorkDir keep: %v", err)
	}
	cleanup2()
	if _, err := os.Stat(dir2); err != nil {
		t.Fatalf("expected kept dir: %v", err)
	}
	_ = os.RemoveAll(dir2)

	if err := ensureBinary("sh"); err != nil {
		t.Fatalf("expected shell binary available: %v", err)
	}
	if err := ensureBinary("definitely-not-a-real-binary-12345"); err == nil {
		t.Fatalf("expected missing binary error")
	}
}

func TestRunCommandsErrorPaths(t *testing.T) {
	opts := options{latexBinary: "definitely-not-a-real-binary-12345", dvisvgmBinary: "definitely-not-a-real-binary-67890"}
	if _, err := runLatex(opts, t.TempDir(), filepath.Join(t.TempDir(), "equation.tex")); err == nil {
		t.Fatalf("expected runLatex error when binary is missing")
	}
	if err := runDvisvgm(opts, filepath.Join(t.TempDir(), "equation.dvi"), filepath.Join(t.TempDir(), "out.svg")); err == nil {
		t.Fatalf("expected runDvisvgm error when binary is missing")
	}
}

func TestNormalizeSVGViewBoxErrors(t *testing.T) {
	dir := t.TempDir()
	noViewBox := filepath.Join(dir, "novb.svg")
	if err := os.WriteFile(noViewBox, []byte("<svg><defs></defs></svg>"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := normalizeSVGViewBox(noViewBox); err == nil {
		t.Fatalf("expected missing viewBox error")
	}

	badFormat := filepath.Join(dir, "bad.svg")
	if err := os.WriteFile(badFormat, []byte("<svg viewBox='1 2 3'><defs></defs></svg>"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := normalizeSVGViewBox(badFormat); err == nil {
		t.Fatalf("expected malformed viewBox error")
	}

	zeroShift := filepath.Join(dir, "zero.svg")
	if err := os.WriteFile(zeroShift, []byte("<svg viewBox='0 0 10 20'><defs></defs></svg>"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := normalizeSVGViewBox(zeroShift); err != nil {
		t.Fatalf("expected no-op on zero origin: %v", err)
	}
}

func TestSanitizeSVGRoot(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "root.svg")
	input := "<svg width='10pt' height='20pt' viewBox='0 0 10 20' class='x'><g/></svg>"
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}
	if err := sanitizeSVGRoot(path); err != nil {
		t.Fatalf("sanitizeSVGRoot: %v", err)
	}
	out, _ := os.ReadFile(path)
	s := string(out)
	if !containsAll(s, []string{"xmlns='http://www.w3.org/2000/svg'", "xmlns:xlink='http://www.w3.org/1999/xlink'", "version='1.1'"}) {
		t.Fatalf("sanitized root missing attrs: %s", s)
	}

	bad := filepath.Join(dir, "badroot.svg")
	if err := os.WriteFile(bad, []byte("<svg width='10pt'><g/></svg>"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := sanitizeSVGRoot(bad); err == nil {
		t.Fatalf("expected missing attributes error")
	}
}

func TestHelpers(t *testing.T) {
	attrs := []xml.Attr{{Name: xml.Name{Local: "id"}, Value: "p1"}, {Name: xml.Name{Local: "d"}, Value: "M0 0"}, {Name: xml.Name{Local: "fill"}, Value: "none"}}
	inherited := inheritPathAttrs(attrs)
	if len(inherited) != 1 || inherited[0].Name.Local != "fill" {
		t.Fatalf("unexpected inherited attrs: %+v", inherited)
	}
	base := []xml.Attr{{Name: xml.Name{Local: "fill"}, Value: "red"}}
	extra := []xml.Attr{{Name: xml.Name{Local: "fill"}, Value: "blue"}, {Name: xml.Name{Local: "stroke"}, Value: "black"}}
	merged := appendUniqueAttrs(base, extra)
	if len(merged) != 2 {
		t.Fatalf("expected unique merge, got %+v", merged)
	}
	if got := attrValueFromTag("<svg width='10' viewBox='0 0 1 1'>", "viewBox"); got != "0 0 1 1" {
		t.Fatalf("attrValueFromTag mismatch: %q", got)
	}
	if got := formatFloat(10.5); got != "10.5" {
		t.Fatalf("formatFloat mismatch: %q", got)
	}
}

func containsAll(s string, parts []string) bool {
	for _, p := range parts {
		if !strings.Contains(s, p) {
			return false
		}
	}
	return true
}
