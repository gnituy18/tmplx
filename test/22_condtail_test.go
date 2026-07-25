package main

import "testing"

// Content after a false tx-if sibling must still render: a false branch
// skips only its own element, never the rest of the parent. (Regression:
// the open if-brace used to swallow every following sibling.)
func TestContentAfterFalseConditional(t *testing.T) {
	ctx := newPage(t, "/condtail")

	waitText(t, ctx, "#a", "hello")  // interpolation after a false tx-if
	waitText(t, ctx, "#b", "hello")  // before the conditional (control)
	waitText(t, ctx, "#c", "static") // static text after a false tx-if
	waitText(t, ctx, "#d", "xhello") // after a true tx-if

	// and the swap path agrees with the GET path
	click(t, ctx, "#toggle")
	waitText(t, ctx, "#a", "xhello")
	click(t, ctx, "#toggle")
	waitText(t, ctx, "#a", "hello")
}
