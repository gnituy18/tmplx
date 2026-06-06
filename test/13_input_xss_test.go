package main

import "testing"

// Edge case (dynamic XSS): user-typed HTML round-trips through state and is
// echoed as escaped text, never injected as markup.
func TestInputEchoEscapesTypedHTML(t *testing.T) {
	ctx := newPage(t, "/input")
	setValue(t, ctx, `[data-test="text-input"]`, "<b>x</b>")
	// textContent decodes entities -> we see the literal characters back, with
	// the correct length (8), and no injected <b>.
	waitText(t, ctx, `[data-test="echo"]`, "You typed: <b>x</b> (8 chars)")
	if n := countNodes(t, ctx, `[data-test="echo"] b`); n != 0 {
		t.Errorf("typed HTML was injected as %d <b> element(s); must be escaped", n)
	}
}
