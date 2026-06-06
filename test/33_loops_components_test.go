package main

// tx-for over components with tx-key: keyed identity keeps per-row component state stable.

import "testing"

func TestLoopComponentsForKeys(t *testing.T) {
	ctx := newPage(t, "/loop-comps")
	if n := countNodes(t, ctx, `[data-test="cbtn"]`); n != 3 {
		t.Fatalf("counters: got %d, want 3", n)
	}
	clickNth(t, ctx, `[data-test="cbtn"]`, 1) // middle (green)
	clickNth(t, ctx, `[data-test="cbtn"]`, 1)
	waitNthText(t, ctx, `[data-test="cbtn"]`, 1, "click me (2)")
	waitNthText(t, ctx, `[data-test="cbtn"]`, 0, "click me (0)")
	waitNthText(t, ctx, `[data-test="cbtn"]`, 2, "click me (0)")
}
