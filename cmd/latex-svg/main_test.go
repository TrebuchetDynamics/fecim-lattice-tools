package main

import (
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

func containsAll(s string, parts []string) bool {
	for _, p := range parts {
		if !strings.Contains(s, p) {
			return false
		}
	}
	return true
}
