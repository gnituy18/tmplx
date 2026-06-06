package main

import "testing"

// A component value-prop can be an arbitrary expression (not just an ident or
// literal), and it stays reactive: changing a dependency re-renders the child.
func TestExpressionValuedProp(t *testing.T) {
	ctx := newPage(t, "/expr-prop")
	waitText(t, ctx, `[data-test="stat"]`, "sum: 9") // 4*2+1
	click(t, ctx, `[data-test="ep-inc"]`)            // base -> 5
	waitText(t, ctx, `[data-test="stat"]`, "sum: 11") // 5*2+1
}
