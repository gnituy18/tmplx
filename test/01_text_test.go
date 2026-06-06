package main

import "testing"

// Text interpolation { expr } — the most basic feature. Each { ... } is an
// arbitrary Go expression whose value is rendered and HTML-escaped.

// Several interpolations in one text node, including a bare string literal,
// all concatenate with the surrounding static text.
func TestTextMultipleInterpolations(t *testing.T) {
	ctx := newPage(t, "/text")
	waitText(t, ctx, `[data-test="greet"]`, "Hello, world! You have 3 items.")
}

// Interpolations can be any Go expression: arithmetic, builtins (len), etc.
func TestTextExpressions(t *testing.T) {
	ctx := newPage(t, "/text")
	waitText(t, ctx, `[data-test="expr"]`, "2+2=4 len=5 up=30")
}

// Two interpolations with no text between them must each render separately
// ({ name }{ count } -> "world3"), not be concatenated into one expression.
func TestTextAdjacentInterpolations(t *testing.T) {
	ctx := newPage(t, "/text")
	waitText(t, ctx, `[data-test="adjacent"]`, "world3")
}

// Edge case (XSS safety): a string containing HTML must be escaped, not
// injected — the markup shows up as literal text and creates no real elements.
func TestTextInterpolationIsEscaped(t *testing.T) {
	ctx := newPage(t, "/text")
	// textContent decodes entities, so we see the raw characters back...
	waitText(t, ctx, `[data-test="escape"]`, "<b>bold</b> & 'q'")
	// ...but crucially no <b> element was actually created.
	if n := countNodes(t, ctx, `[data-test="escape"] b`); n != 0 {
		t.Errorf("interpolation injected %d <b> element(s); must be escaped", n)
	}
}

// tx-ignore is the escape hatch for literal braces: the element's text is
// rendered verbatim, with no { } interpolation, and the probe skips it (so an
// "undefined" name inside the braces is harmless).
func TestTextTxIgnoreRendersBracesLiterally(t *testing.T) {
	ctx := newPage(t, "/text")
	waitText(t, ctx, `[data-test="ignored"]`, "literal { name } and { undefinedVar } stay")
	// the tx-ignore attribute itself must not leak into the rendered HTML
	if v := attrOf(t, ctx, `[data-test="ignored"]`, "tx-ignore"); v != "" {
		t.Errorf("tx-ignore attribute leaked into output: %q", v)
	}
}
