package main

import "testing"

// tx-if / tx-else on an element INSIDE a tx-for: the condition can reference the
// enclosing loop variable, choosing a branch per row. (tx-for + tx-if on the
// SAME element is a separate, unsupported case — see memory.)
func TestConditionalRowsInsideLoop(t *testing.T) {
	ctx := newPage(t, "/filter")
	// nums = 1..6; n > 3 is "big" (4,5,6), else "small" (1,2,3)
	if n := countNodes(t, ctx, `[data-test="big"]`); n != 3 {
		t.Errorf("big rows: got %d, want 3", n)
	}
	if n := countNodes(t, ctx, `[data-test="small"]`); n != 3 {
		t.Errorf("small rows: got %d, want 3", n)
	}
	if got := nthText(t, ctx, `[data-test="small"]`, 0); got != "1-small" {
		t.Errorf("first small: got %q, want 1-small", got)
	}
	if got := nthText(t, ctx, `[data-test="big"]`, 0); got != "4-big" {
		t.Errorf("first big: got %q, want 4-big", got)
	}
}
