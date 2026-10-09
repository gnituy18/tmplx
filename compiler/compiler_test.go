package compiler

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// typecheck parses and type-checks generated routes.go bytes against the
// standard library, so import and codegen mistakes fail here instead of in
// the user's go build.
func typecheck(t *testing.T, code []byte) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "routes.go", code, 0)
	if err != nil {
		t.Fatalf("generated code does not parse: %v", err)
	}
	if _, err := (&types.Config{Importer: importer.Default()}).Check("main", fset, []*ast.File{f}, nil); err != nil {
		t.Errorf("generated code does not typecheck: %v", err)
	}
}

// TestCompileProject is the bridge to the E2E suite: the same fixture project
// the browser tests exercise (test/) must compile and typecheck at the bytes
// level, with no browser in the loop.
func TestCompileProject(t *testing.T) {
	c := &Compiler{}
	for _, dir := range []string{"../test/components", "../test/pages"} {
		page := strings.HasSuffix(dir, "pages")
		err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() || filepath.Ext(path) != ".html" {
				return err
			}
			rel, _ := filepath.Rel(dir, path)
			b, rerr := os.ReadFile(path)
			if rerr != nil {
				return rerr
			}
			if page {
				c.NewPage(filepath.ToSlash(rel), b)
			} else {
				c.NewComponent(filepath.ToSlash(rel), b)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	code, err := c.Compile()
	if err != nil {
		t.Fatalf("the E2E project must compile: %v", err)
	}
	typecheck(t, code)
}

// compilePage compiles one in-memory page the way the playground does (stdlib
// imports only) and returns the generated Go plus any diagnostics.
func compilePage(src string) (code []byte, diagnostics []string) {
	c := &Compiler{}
	c.NewPage("index.html", []byte(src))
	code, err := c.Compile()
	return code, Messages(err)
}

func TestCompileCounter(t *testing.T) {
	src := `<!DOCTYPE html>
<html><head><title>x</title></head>
<body>
<script type="text/tmplx">
  var counter int
</script>
<button tx-onclick="counter++">+</button>
<span>{ counter }</span>
</body></html>`
	code, diags := compilePage(src)
	if len(diags) > 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	for _, want := range []string{"package main", "V_counter int", "func Routes() []TxRoute", "tx_eh1", "html.EscapeString"} {
		if !strings.Contains(string(code), want) {
			t.Errorf("generated code missing %q\n---\n%s", want, code)
		}
	}
}

func TestCompileUndefinedVar(t *testing.T) {
	code, diags := compilePage(`<html><head></head><body><span>{ missing }</span></body></html>`)
	if len(diags) == 0 {
		t.Fatal("expected a diagnostic for the undefined variable")
	}
	if len(code) != 0 {
		t.Errorf("expected empty code on error, got %d bytes", len(code))
	}
	// M3: the template-interpolation error is positioned at the expression.
	if !regexp.MustCompile(`index\.html:1:\d+:.*undefined`).MatchString(strings.Join(diags, "\n")) {
		t.Errorf("expected a positioned 'undefined' diagnostic, got: %v", diags)
	}
}

func TestCompileSyntaxError(t *testing.T) {
	_, diags := compilePage(`<html><head><script type="text/tmplx">
  var x int =
</script></head><body></body></html>`)
	if len(diags) == 0 {
		t.Fatal("expected a diagnostic for the Go syntax error")
	}
	// M1: the diagnostic carries a precise .html position (index.html:line:col).
	if !regexp.MustCompile(`index\.html:\d+:\d+:`).MatchString(diags[0]) {
		t.Errorf("expected a positioned diagnostic, got: %q", diags[0])
	}
}

// M2: a type error in a script var (undefined name) reports a precise position.
func TestCompileScriptTypeErrorPositioned(t *testing.T) {
	src := `<html><head><script type="text/tmplx">
  var n int = "oops"
</script></head><body><span>{ n }</span></body></html>`
	_, diags := compilePage(src)
	if len(diags) == 0 {
		t.Fatal("expected a diagnostic")
	}
	if !regexp.MustCompile(`index\.html:\d+:\d+:`).MatchString(strings.Join(diags, "\n")) {
		t.Errorf("expected a positioned type-error diagnostic, got: %v", diags)
	}
}

func TestCompileStdlibImport(t *testing.T) {
	src := `<html><head><script type="text/tmplx">
  import "strings"
  var greeting string = strings.ToUpper("hi")
</script></head><body><p>{ greeting }</p></body></html>`
	code, diags := compilePage(src)
	if len(diags) > 0 {
		t.Fatalf("stdlib import should compile: %v", diags)
	}
	if !strings.Contains(string(code), `"strings"`) {
		t.Errorf("expected strings import in output:\n%s", code)
	}
}

// A non-stdlib import the server can't resolve must surface as an ordinary
// type-check error that names the import, not silently compile.
func TestCompileUnresolvableImport(t *testing.T) {
	src := `<html><head><script type="text/tmplx">
  import "github.com/nope/nope"
  var x = nope.Thing
</script></head><body><p>{ x }</p></body></html>`
	code, diags := compilePage(src)
	if len(code) != 0 {
		t.Error("expected no code when an import cannot be resolved")
	}
	if len(diags) == 0 || !strings.Contains(strings.Join(diags, " "), "github.com/nope/nope") {
		t.Errorf("expected a diagnostic naming the unresolvable import, got: %v", diags)
	}
}

func TestCompileConditionalAndLoop(t *testing.T) {
	src := `<html><head><script type="text/tmplx">
  var items = []string{"a", "b"}
  var show bool = true
</script></head>
<body>
  <ul><li tx-for="i, x := range items" tx-key="i">{ x }</li></ul>
  <p tx-if="show">yes</p>
</body></html>`
	code, diags := compilePage(src)
	if len(diags) > 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !strings.Contains(string(code), "for ") || !strings.Contains(string(code), "if tx_comp.V_show") {
		t.Errorf("expected loop + conditional codegen:\n%s", code)
	}
}

// Two compiles in a row must not leak state between them. With no package-level
// state, distinct Configs are independent by construction.
func TestCompileIsolation(t *testing.T) {
	_, d1 := compilePage(`<html><head><script type="text/tmplx">
  var a int
</script></head><body>{ a }</body></html>`)
	if len(d1) > 0 {
		t.Fatalf("first compile failed: %v", d1)
	}
	code, d2 := compilePage(`<html><head><script type="text/tmplx">
  var b int
</script></head><body>{ b }</body></html>`)
	if len(d2) > 0 {
		t.Fatalf("second compile failed: %v", d2)
	}
	if strings.Contains(string(code), "V_a") {
		t.Error("state from the first compile leaked into the second")
	}
}

// The emitter owns the import block (no goimports): user imports are deduped
// across files, the generated set is added only when not already bound, and
// the result must typecheck.
func TestGeneratedImports(t *testing.T) {
	c := &Compiler{}
	c.NewPage("a.html", []byte(`<html><head><script type="text/tmplx">
import "strings"
var x = strings.ToUpper("a")
</script></head><body>{ x }</body></html>`))
	c.NewPage("b.html", []byte(`<html><head><script type="text/tmplx">
import "strings"
import "fmt"
var y = fmt.Sprint(strings.ToLower("B"))
</script></head><body>{ y }</body></html>`))
	code, err := c.Compile()
	if err != nil {
		t.Fatal(err)
	}
	for _, im := range []string{"\t\"strings\"\n", "\t\"fmt\"\n"} {
		if n := strings.Count(string(code), im); n != 1 {
			t.Errorf("want exactly one %s import, got %d", strings.TrimSpace(im), n)
		}
	}
	fset := token.NewFileSet()
	f, perr := parser.ParseFile(fset, "routes.go", code, 0)
	if perr != nil {
		t.Fatal(perr)
	}
	conf := types.Config{Importer: importer.Default()}
	if _, terr := conf.Check("main", fset, []*ast.File{f}, nil); terr != nil {
		t.Errorf("generated code does not typecheck: %v", terr)
	}
}

// A slot name that isn't a valid Go identifier must be goIdent-encoded the
// same way at the fill function's definition and its call.
func TestHyphenatedSlotName(t *testing.T) {
	c := &Compiler{}
	c.NewComponent("card.html", []byte(`<div><slot name="side-bar"></slot></div>`))
	c.NewPage("index.html", []byte(`<html><head><title>x</title></head><body><tx-card><p slot="side-bar">hi</p></tx-card></body></html>`))
	code, err := c.Compile()
	if err != nil {
		t.Fatalf("slot name side-bar should compile: %v", err)
	}
	if !strings.Contains(string(code), "tx_render_fill_tx_HY_card_1_side_HY_bar(tx_w *bytes.Buffer") {
		t.Error("fill definition should use the goIdent-encoded slot name")
	}
}

// posOf returns the 1-based line/col of the first occurrence of sub in src.
func posOf(src, sub string) (int, int) {
	i := strings.Index(src, sub)
	line := strings.Count(src[:i], "\n") + 1
	col := i - strings.LastIndex(src[:i], "\n")
	return line, col
}

// The semantic foundation: go-to-definition + hover resolve through go/types
// (same-file, cross-package, cross-component) rather than name matching.
func TestSemanticQueries(t *testing.T) {
	c := &Compiler{}
	c.NewComponent("box.html", []byte("<script type=\"text/tmplx\">\n//tx:prop\nvar label string\n</script><p>{ label }</p>"))
	page := `<html><head><script type="text/tmplx">
import "fmt"
var count int
func show() { fmt.Println(count) }
</script></head>
<body>
  <button tx-onclick="show()">{ count }</button>
  <tx-box label="x"></tx-box>
</body></html>`
	c.NewPage("index.html", []byte(page))

	// count in the { count } text interpolation -> its declaration (line 3)
	l, col := posOf(page, "count }")
	if d, ok := c.Definition("pages/index.html", l, col); !ok || d.Path != "pages/index.html" || d.Span.Start.Line != 3 {
		t.Errorf("text-interp count: want pages/index.html:3, got %+v ok=%v", d, ok)
	}
	// fmt.Println inside the script -> external fmt.Println
	l, col = posOf(page, "Println")
	if d, ok := c.Definition("pages/index.html", l, col); !ok || d.Pkg != "fmt" || d.Name != "Println" {
		t.Errorf("fmt.Println: want external fmt.Println, got %+v ok=%v", d, ok)
	}
	// show() handler call -> the func decl (line 4)
	l, col = posOf(page, "show()\"")
	if d, ok := c.Definition("pages/index.html", l, col); !ok || d.Span.Start.Line != 4 {
		t.Errorf("handler show: want line 4, got %+v ok=%v", d, ok)
	}
	// label prop attr on <tx-box> -> box.html prop decl
	l, col = posOf(page, "label=")
	if d, ok := c.Definition("pages/index.html", l, col); !ok || d.Path != "components/box.html" || d.Span.Start.Line != 3 {
		t.Errorf("prop attr label: want components/box.html:3, got %+v ok=%v", d, ok)
	}
	// hover on count reports its type
	l, col = posOf(page, "count }")
	if h, _, ok := c.Hover("pages/index.html", l, col); !ok || !strings.Contains(h, "var count int") {
		t.Errorf("hover count: want 'var count int', got %q ok=%v", h, ok)
	}
	// references to count: declaration + script use + template use = 3
	l, col = posOf(page, "count int")
	if refs := c.References("pages/index.html", l, col); len(refs) != 3 {
		t.Errorf("references to count: want 3, got %d (%+v)", len(refs), refs)
	}
	// document symbols include count (state) and show (func)
	names := map[string]bool{}
	for _, s := range c.Symbols("pages/index.html") {
		names[s.Name] = true
	}
	if !names["count"] || !names["show"] {
		t.Errorf("symbols: want count + show, got %v", names)
	}
}

// A page and a component may share a rel path; diagnostics must say which is
// which (components/foo.html vs pages/foo.html).
func TestSameNamePageAndComponent(t *testing.T) {
	c := &Compiler{}
	c.NewComponent("foo.html", []byte(`<script type="text/tmplx">var a int</script><p>x</p>`))
	c.NewPage("foo.html", []byte(`<html><head><script type="text/tmplx">var b int</script></head><body>y</body></html>`))

	byPath := map[string]string{}
	for _, d := range c.Diagnose() {
		byPath[d.Path] = d.Message
	}
	if !strings.Contains(byPath["components/foo.html"], "a declared") {
		t.Errorf("want the component's error under components/foo.html, got %v", byPath)
	}
	if !strings.Contains(byPath["pages/foo.html"], "b declared") {
		t.Errorf("want the page's error under pages/foo.html, got %v", byPath)
	}
}

// The CLI passes the user's actual dir names (-components-dir / -pages-dir),
// so diagnostics must name files the way the user knows them.
func TestCustomDirNames(t *testing.T) {
	c := &Compiler{ComponentsDir: "widgets", PagesDir: "web/views"}
	c.NewComponent("foo.html", []byte(`<script type="text/tmplx">var a int</script><p>x</p>`))
	c.NewPage("foo.html", []byte(`<html><head><script type="text/tmplx">var b int</script></head><body>y</body></html>`))

	paths := map[string]bool{}
	for _, d := range c.Diagnose() {
		paths[d.Path] = true
	}
	if !paths["widgets/foo.html"] || !paths["web/views/foo.html"] {
		t.Errorf("want widgets/foo.html and web/views/foo.html, got %v", paths)
	}
}
