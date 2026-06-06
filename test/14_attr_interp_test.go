package main

import "testing"

// Interpolation inside an attribute value: { expr } in an attribute is rendered
// (and HTML-escaped, so it can't break out of the attribute).
func TestAttrInterpolation(t *testing.T) {
	ctx := newPage(t, "/attr-interp")
	if href := attrOf(t, ctx, `[data-test="link"]`, "href"); href != "/echo/alice" {
		t.Errorf("href: got %q, want /echo/alice", href)
	}
	if cls := attrOf(t, ctx, `[data-test="dyn-class"]`, "class"); cls != "box active" {
		t.Errorf("class: got %q, want 'box active'", cls)
	}
}
