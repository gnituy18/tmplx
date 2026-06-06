package main

import "testing"

// Formatting edges: how non-string scalars render via fmt-style output. Notably
// a whole-number float (1.0) prints as "1" (Go's default float formatting).
func TestRenderEdges(t *testing.T) {
	ctx := newPage(t, "/render-edges")
	cases := []struct{ sel, want string }{
		{`[data-test="re-f"]`, "[1]"},     // float64(1.0) -> "1"
		{`[data-test="re-g"]`, "[0.5]"},   // float64(0.5) -> "0.5"
		{`[data-test="re-neg"]`, "[-42]"}, // negative int
		{`[data-test="re-emoji"]`, "[hello world]"},
		{`[data-test="re-zero"]`, "[0]"}, // zero value renders, not blank
	}
	for _, c := range cases {
		waitText(t, ctx, c.sel, c.want)
	}
}
