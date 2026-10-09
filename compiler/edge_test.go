package compiler

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// Non-ASCII runes inside { expr } must survive byte-exact (the scanner once
// truncated each rune to its low byte).
func TestUnicodeInExpression(t *testing.T) {
	code, diags := compilePage(`<html><head><title>x</title></head><body><p>{ "café 世界" }</p></body></html>`)
	if len(diags) > 0 {
		t.Fatalf("unicode in an expression should compile: %v", diags)
	}
	if !strings.Contains(string(code), `café 世界`) {
		t.Errorf("expected the string literal to survive byte-exact:\n%s", code)
	}
	typecheck(t, code)
}

// An unclosed { or quote used to be silently swallowed by the renderer; it
// must be a diagnostic.
func TestUnclosedBraceDiagnostic(t *testing.T) {
	_, diags := compilePage(`<html><head><title>x</title></head><body><p>before { oops</p></body></html>`)
	if len(diags) == 0 || !strings.Contains(strings.Join(diags, " "), "unclosed {") {
		t.Errorf("want an unclosed-brace diagnostic, got: %v", diags)
	}

	_, diags = compilePage(`<html><head><title>x</title></head><body><p title="{ oops">x</p></body></html>`)
	if len(diags) == 0 || !strings.Contains(strings.Join(diags, " "), "unclosed {") {
		t.Errorf("want an unclosed-brace diagnostic for the attribute, got: %v", diags)
	}
}

// tx-ignore must disable interpolation in attributes too (the docs always
// said so; the renderer used to interpolate them anyway).
func TestIgnoredAttrInterpolation(t *testing.T) {
	code, diags := compilePage(`<html><head><title>x</title></head><body><p tx-ignore title="{ x }">y</p></body></html>`)
	if len(diags) > 0 {
		t.Fatalf("tx-ignore attr should compile without probing { x }: %v", diags)
	}
	if !strings.Contains(string(code), `title=\"{ x }\"`) {
		t.Errorf("expected the attribute emitted literally:\n%s", code)
	}
	typecheck(t, code)
}

// Static attribute text must be HTML-escaped on output: a quote entering via
// an entity (&quot; -> ") must not break out of the generated attribute.
func TestStaticAttrEscaped(t *testing.T) {
	code, diags := compilePage(`<html><head><title>x</title></head><body><p title="a&quot;b">x</p></body></html>`)
	if len(diags) > 0 {
		t.Fatal(diags)
	}
	if !strings.Contains(string(code), `a&#34;b`) {
		t.Errorf("expected the quote re-escaped in the emitted attribute:\n%s", code)
	}
}

// <tx-foo /> parses as an unclosed tag and swallows its siblings; it must be
// rejected with a position instead.
func TestSelfClosingComponent(t *testing.T) {
	c := &Compiler{}
	c.NewComponent("box.html", []byte(`<p>box</p>`))
	c.NewPage("index.html", []byte("<html><head><title>x</title></head><body>\n<tx-box />\n<p>after</p></body></html>"))
	var found bool
	for _, d := range c.Diagnose() {
		if strings.Contains(d.Message, "self-closing") {
			found = true
			if d.Span.Start.Line != 2 {
				t.Errorf("want the diagnostic on line 2, got %+v", d.Span)
			}
		}
	}
	if !found {
		t.Error("want a self-closing-component diagnostic")
	}
}

// init's body is inlined into the constructor; parameters or results would be
// undefined there, so they must be rejected. Generic funcs become methods,
// which cannot have type parameters.
func TestScriptSignatureLimits(t *testing.T) {
	_, diags := compilePage(`<html><head><script type="text/tmplx">
var n int
func init(x int) { n = x }
</script></head><body>{ n }</body></html>`)
	if !strings.Contains(strings.Join(diags, " "), "init cannot have parameters") {
		t.Errorf("want the init-signature diagnostic, got: %v", diags)
	}

	_, diags = compilePage(`<html><head><script type="text/tmplx">
var n int
func g[T any]() {}
</script></head><body>{ n }</body></html>`)
	if !strings.Contains(strings.Join(diags, " "), "generic functions are not supported") {
		t.Errorf("want the generics diagnostic, got: %v", diags)
	}
}

// A handler value that escapes the func wrapper used to be silently dropped;
// it must be a diagnostic.
func TestHandlerEscapesWrapper(t *testing.T) {
	_, diags := compilePage(`<html><head><script type="text/tmplx">
var count int
</script></head><body><button tx-onclick="count++ }
func evil() { count--">x</button>{ count }</body></html>`)
	if !strings.Contains(strings.Join(diags, " "), "handler must be a sequence of statements") {
		t.Errorf("want the handler-shape diagnostic, got: %v", diags)
	}
}

// Only the full event.target.value chain is rewritten for the generated code;
// any other use of `event` would be an undefined identifier in routes.go, so
// it must be a diagnostic.
func TestEventLeftover(t *testing.T) {
	_, diags := compilePage(`<html><head><script type="text/tmplx">
var x string
</script></head><body><input tx-oninput="t := event.target
x = t.value">{ x }</body></html>`)
	if !strings.Contains(strings.Join(diags, " "), "only event.target.value is available") {
		t.Errorf("want the event-use diagnostic, got: %v", diags)
	}

	// the supported chain still works
	code, diags := compilePage(`<html><head><script type="text/tmplx">
var x string
</script></head><body><input tx-oninput="x = event.target.value">{ x }</body></html>`)
	if len(diags) > 0 {
		t.Fatalf("event.target.value should compile: %v", diags)
	}
	typecheck(t, code)
}

// Non-whitespace text ends a conditional chain: tx-else after it is a chain
// error, while whitespace between branches keeps the chain intact.
func TestTextEndsConditionalChain(t *testing.T) {
	_, diags := compilePage(`<html><head><title>x</title><script type="text/tmplx">
var f bool
</script></head><body><p tx-if="f">a</p> hello <p tx-else>b</p></body></html>`)
	if !strings.Contains(strings.Join(diags, " "), "must follow tx-if") {
		t.Errorf("want a broken-chain diagnostic for text between branches, got: %v", diags)
	}

	code, diags := compilePage(`<html><head><title>x</title><script type="text/tmplx">
var f bool
</script></head><body><p tx-if="f">a</p>
	<p tx-else>b</p></body></html>`)
	if len(diags) > 0 {
		t.Fatalf("whitespace between branches must keep the chain: %v", diags)
	}
	typecheck(t, code)
}

// tx-debounce must be validated, stripped from the rendered attrs, and
// re-emitted as data-tx-debounce for the runtime.
func TestDebounce(t *testing.T) {
	code, diags := compilePage(`<html><head><title>x</title><script type="text/tmplx">
var n int
</script></head><body><input tx-oninput="n++" tx-debounce="150"><span>{ n }</span></body></html>`)
	if len(diags) > 0 {
		t.Fatalf("tx-debounce should compile: %v", diags)
	}
	if !strings.Contains(string(code), ` data-tx-debounce=\"150\"`) {
		t.Error("want data-tx-debounce emitted for the runtime")
	}
	if strings.Contains(string(code), ` tx-debounce=`) {
		t.Error("the raw tx-debounce attribute must be stripped from output")
	}
	typecheck(t, code)

	_, diags = compilePage(`<html><head><title>x</title><script type="text/tmplx">
var n int
</script></head><body><input tx-oninput="n++" tx-debounce="fast"><span>{ n }</span></body></html>`)
	if !strings.Contains(strings.Join(diags, " "), "positive integer") {
		t.Errorf("want the bad-value diagnostic, got: %v", diags)
	}

	_, diags = compilePage(`<html><head><title>x</title></head><body><p tx-debounce="100">x</p></body></html>`)
	if !strings.Contains(strings.Join(diags, " "), "no tx-on handler") {
		t.Errorf("want the no-handler diagnostic, got: %v", diags)
	}
}

// State must round-trip through the #tx-saved JSON blob, so a type that can't
// (func, chan, complex, non-empty interface, bad map key) is rejected at
// compile time with the offending part named. Previously only an ANNOTATED
// func type was caught; the inferred form compiled green and broke at runtime.
func TestStateJSONRoundTrip(t *testing.T) {
	bad := []struct{ decl, use, want string }{
		{"var f = func() int { return 1 }", "{ f() }", "use //tx:prop"},
		{"var ch chan int", "{ len(ch) }", "not serializable"},
		{"var z complex128", "{ real(z) }", "not supported by encoding/json"},
		{"var e error", "{ e }", "interface error"},
		{"var m map[float64]int", "{ len(m) }", "map key"},
	}
	for _, c := range bad {
		_, diags := compilePage(`<html><head><script type="text/tmplx">
` + c.decl + `
</script></head><body>` + c.use + `</body></html>`)
		if !strings.Contains(strings.Join(diags, " "), c.want) {
			t.Errorf("%s: want a diagnostic containing %q, got: %v", c.decl, c.want, diags)
		}
	}

	// marshaler types (time.Time) round-trip by contract; []byte and
	// map[string]any are fine
	code, diags := compilePage(`<html><head><script type="text/tmplx">
import "time"
var t0 time.Time
var xs []byte
var m map[string]any
</script></head><body>{ t0.Year() } { len(xs) } { len(m) }</body></html>`)
	if len(diags) > 0 {
		t.Fatalf("marshaler/byte/any state should compile: %v", diags)
	}
	typecheck(t, code)
}

// An imported struct with an unexported field would silently lose that field
// on every event (encoding/json drops it); reject it at compile time, naming
// the field. Hermetic via a local replace module.
func TestStateUnexportedFieldRejected(t *testing.T) {
	dir := t.TempDir()
	write := func(rel, content string) {
		t.Helper()
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module leaky\n\ngo 1.25\n\nrequire example.com/dep v0.0.0\n\nreplace example.com/dep => ./dep\n")
	write("dep/go.mod", "module example.com/dep\n\ngo 1.25\n")
	write("dep/dep.go", "package dep\n\ntype User struct {\n\tName   string\n\tsecret int\n}\n")

	c := &Compiler{Importer: PackageImporter(dir)}
	c.NewPage("index.html", []byte(`<html><head><script type="text/tmplx">
import "example.com/dep"
var u dep.User
</script></head><body>{ u.Name }</body></html>`))
	_, err := c.Compile()
	if err == nil || !strings.Contains(err.Error(), "secret") {
		t.Fatalf("want a diagnostic naming the unexported field, got: %v", err)
	}
}

// A script-only import (a go.mod dependency no .go file imports yet) is
// invisible to the eager packages.Load closure, so PackageImporter must
// resolve it on demand through the module. Hermetic via a local replace.
func TestScriptOnlyImportResolvesThroughModule(t *testing.T) {
	dir := t.TempDir()
	write := func(rel, content string) {
		t.Helper()
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module scriptonly\n\ngo 1.25\n\nrequire example.com/dep v0.0.0\n\nreplace example.com/dep => ./dep\n")
	write("dep/go.mod", "module example.com/dep\n\ngo 1.25\n")
	write("dep/dep.go", "package dep\n\nconst X = 7\n")

	c := &Compiler{Importer: PackageImporter(dir)}
	c.NewPage("index.html", []byte(`<html><head><script type="text/tmplx">
import "example.com/dep"
var x = dep.X
</script></head><body>{ x }</body></html>`))
	code, err := c.Compile()
	if err != nil {
		t.Fatalf("script-only import should resolve through go.mod: %v", err)
	}
	if !strings.Contains(string(code), `"example.com/dep"`) {
		t.Errorf("generated code should carry the import:\n%s", code)
	}
}

// A component that mounts itself through unconditional children would recurse
// forever at runtime; it must be a compile error. A tx-if anywhere on the
// chain breaks the cycle.
func TestComponentCycle(t *testing.T) {
	c := &Compiler{}
	c.NewComponent("a.html", []byte(`<div><tx-b></tx-b></div>`))
	c.NewComponent("b.html", []byte(`<div><tx-a></tx-a></div>`))
	c.NewPage("index.html", []byte(`<html><head><title>x</title></head><body><tx-a></tx-a></body></html>`))
	var found bool
	for _, d := range c.Diagnose() {
		if strings.Contains(d.Message, "component cycle") {
			found = true
		}
	}
	if !found {
		t.Error("want a component-cycle diagnostic for tx-a <-> tx-b")
	}

	guarded := &Compiler{}
	guarded.NewComponent("a.html", []byte(`<script type="text/tmplx">
//tx:prop
var depth int = 0
</script><div tx-if="depth < 3"><tx-a depth="depth + 1"></tx-a></div>`))
	guarded.NewPage("index.html", []byte(`<html><head><title>x</title></head><body><tx-a></tx-a></body></html>`))
	if diags := guarded.Diagnose(); len(diags) > 0 {
		t.Errorf("a tx-if-guarded recursion should compile, got: %v", diags)
	}
}

// Every duplicate tag gets its own positioned error, and the first script is
// still analyzed (its own errors surface in the same run, not after cleanup).
func TestEveryDuplicateTagReported(t *testing.T) {
	c := &Compiler{}
	c.NewComponent("multi.html", []byte(`<script type="text/tmplx">
type T int
</script>
<p>x</p>
<script type="text/tmplx">
</script>
<script type="text/tmplx">
</script>
<style></style>
<style></style>
`))
	ds := c.Diagnose()
	var got []string
	for _, d := range ds {
		got = append(got, d.Error())
	}
	want := []string{
		`components/multi.html:2:1: type declarations are not supported in the script block (declare types in a separate Go package and import them): type T int`,
		`components/multi.html:5:1: multiple <script type="text/tmplx"> elements (only one allowed)`,
		`components/multi.html:7:1: multiple <script type="text/tmplx"> elements (only one allowed)`,
		`components/multi.html:10:1: multiple <style> elements (only one allowed)`,
	}
	if len(got) != len(want) {
		t.Fatalf("got %d diagnostics, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("diagnostic %d:\n got  %s\n want %s", i, got[i], want[i])
		}
	}

	// a duplicate's span covers the whole element (the fix is deleting it)
	wantEnds := []Pos{{Line: 6, Col: 10}, {Line: 8, Col: 10}, {Line: 10, Col: 16}}
	var gotEnds []Pos
	for _, d := range ds {
		if strings.Contains(d.Message, "multiple") {
			gotEnds = append(gotEnds, d.Span.End)
		}
	}
	if !slices.Equal(gotEnds, wantEnds) {
		t.Errorf("duplicate span ends: got %v, want %v", gotEnds, wantEnds)
	}
}

// The tx- namespace is policed on the token stream: unknown names, stray end
// tags, and unclosed tags each get their own positioned error (html.Parse
// repairs all three silently, and an unclosed tag's swallowed siblings can
// vanish as fills for slots the component never declared).
func TestReservedTagMistakes(t *testing.T) {
	c := &Compiler{}
	c.NewComponent("card.html", []byte("<script type=\"text/tmplx\">\n</script>\n<p><slot></slot></p>\n"))
	c.NewPage("index.html", []byte(`<!DOCTYPE html>
<html><head><title>x</title><script type="text/tmplx">
</script></head>
<body>
<tx-crad></tx-crad>
</tx-card>
<tx-card>
</body>
</html>`))
	var got []string
	for _, d := range c.Diagnose() {
		got = append(got, d.Error())
	}
	want := []string{
		"pages/index.html:5:1: unknown component <tx-crad> (no component file defines it)",
		"pages/index.html:6:1: </tx-card> has no matching <tx-card>",
		"pages/index.html:7:1: <tx-card> is never closed, so it swallows everything after it; add </tx-card>",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d diagnostics, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("diagnostic %d:\n got  %s\n want %s", i, got[i], want[i])
		}
	}
}

// A comment mentioning a tag ("<p") must not shadow the next real tag's
// position: annotatePositions consumes comments without recording them.
func TestCommentDoesNotShadowPositions(t *testing.T) {
	_, diags := compilePage(`<html><head><title>x</title><script type="text/tmplx">
var on = true
</script></head><body><span>{ on }</span>
<!-- try <p tx-if> --><p tx-if="on ==">b</p>
</body></html>`)
	want := "pages/index.html:4:23: tx-if on <p>"
	if !strings.Contains(strings.Join(diags, "\n"), want) {
		t.Errorf("want the tx-if error anchored at the real <p> (%s), got: %v", want, diags)
	}
}

// A text node that duplicates an attribute value must not anchor at the
// attribute's copy: annotatePositions consumes the whole open tag.
func TestAttrDoesNotShadowTextPosition(t *testing.T) {
	_, diags := compilePage(`<html><head><title>x</title><script type="text/tmplx">
var on = true
</script></head><body><span>{ on }</span>
<p title="{ bad }">{ bad }</p>
</body></html>`)
	want := "pages/index.html:4:22: text in <p>"
	if !strings.Contains(strings.Join(diags, "\n"), want) {
		t.Errorf("want the text error anchored at the real text (%s), got: %v", want, diags)
	}

	// a '>' inside a quoted value must not end the open-tag scan early
	_, diags = compilePage(`<html><head><title>x</title><script type="text/tmplx">
var on = true
</script></head><body><span>{ on }</span>
<p title="a > b { bad }">{ bad }</p>
</body></html>`)
	want = "pages/index.html:4:28: text in <p>"
	if !strings.Contains(strings.Join(diags, "\n"), want) {
		t.Errorf("want the text error anchored after the whole open tag (%s), got: %v", want, diags)
	}
}

// A tx-* attribute value must be quoted: HTML truncates an unquoted value at
// the first space or '>', and the truncation can still be valid Go (tx-if=a>b
// silently becomes tx-if="a"), so the spelling is rejected outright. Plain
// HTML attributes and bare tx flags stay legal.
func TestUnquotedTxAttrValue(t *testing.T) {
	_, diags := compilePage(`<html><head><title>x</title><script type="text/tmplx">
var a = true
</script></head><body>
<p tx-if=a>b</p>
<p class=big tx-if="a" tx-else>quoted tx and plain unquoted are fine</p>
</body></html>`)
	if len(diags) != 1 || !strings.Contains(diags[0], `pages/index.html:4:4: tx-if value must be quoted`) {
		t.Errorf("want exactly the positioned unquoted-tx-if diagnostic, got: %v", diags)
	}
}

// A component's script is split out of the template, but its bytes must still
// be consumed by the position sweep: a Go string containing "<p " would
// otherwise shadow the next real <p>'s position.
func TestScriptTextDoesNotShadowPositions(t *testing.T) {
	c := &Compiler{}
	c.NewComponent("shadow.html", []byte(`<script type="text/tmplx">
  var s = "<p "
  var x = 1
</script>
<p tx-if="x ==">bad</p>
`))
	var got []string
	for _, d := range c.Diagnose() {
		got = append(got, d.Error())
	}
	want := "components/shadow.html:5:1: tx-if on <p>"
	if !strings.Contains(strings.Join(got, "\n"), want) {
		t.Errorf("want the tx-if error anchored at the real <p> (%s), got: %v", want, got)
	}
}

// A component's tmplx script and style must be top-level: a nested tmplx
// script is never parsed and would render its Go source into the page.
func TestNestedScriptAndStyleRejected(t *testing.T) {
	c := &Compiler{}
	c.NewComponent("nested.html", []byte(`<div>
  <script type="text/tmplx">
    var secret = "x"
  </script>
  <style>p {}</style>
</div>
`))
	var got []string
	for _, d := range c.Diagnose() {
		got = append(got, d.Error())
	}
	want := []string{
		`components/nested.html:2:3: <script type="text/tmplx"> must be a top-level element in a component, not inside <div>`,
		`components/nested.html:5:3: <style> must be a top-level element in a component, not inside <div>`,
	}
	if len(got) != len(want) {
		t.Fatalf("got %d diagnostics, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("diagnostic %d:\n got  %s\n want %s", i, got[i], want[i])
		}
	}
}
