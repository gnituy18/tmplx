package main

import "testing"

// tx-if / tx-else-if / tx-else. Conditions are re-evaluated every render, so a
// state change can mount or unmount the guarded node (and any components in it).

// tx-if with no else: the node is in the DOM only while the condition is true.
func TestConditionalNoElse(t *testing.T) {
	ctx := newPage(t, "/conditionals")
	waitText(t, ctx, `[data-test="cond-show"]`, "visible")
	click(t, ctx, `[data-test="cond-toggle"]`)
	waitCount(t, ctx, `[data-test="cond-show"]`, 0)
	click(t, ctx, `[data-test="cond-toggle"]`)
	waitText(t, ctx, `[data-test="cond-show"]`, "visible")
}

// if/else-if/else chain: exactly one branch is present, and it switches as the
// driving value crosses the thresholds (0 -> none, 1 -> small, 6 -> big).
func TestConditionalChain(t *testing.T) {
	ctx := newPage(t, "/conditionals")
	waitText(t, ctx, `[data-test="bucket"]`, "none")
	click(t, ctx, `[data-test="cond-inc"]`)
	waitText(t, ctx, `[data-test="bucket"]`, "small")
	clickAll(t, ctx, `[data-test="cond-inc"]`, 5)
	waitText(t, ctx, `[data-test="bucket"]`, "big")
	// only ever one branch in the DOM at a time
	if n := countNodes(t, ctx, `[data-test="bucket"]`); n != 1 {
		t.Errorf("expected exactly one bucket branch, got %d", n)
	}
}

// A component inside a tx-if unmounts when the branch goes false and mounts
// again when true. On re-mount it starts FRESH: unmount purges the instance's
// entry from the client state blob (the swap response replaces the whole rebuilt
// subtree's state), so a hidden component's old state does not come back.
func TestConditionalComponentStateFreshOnRemount(t *testing.T) {
	ctx := newPage(t, "/conditionals")
	waitText(t, ctx, `[data-test="cbtn"]`, "click me (0)")
	click(t, ctx, `[data-test="cbtn"]`) // bump the inner counter to 1
	waitText(t, ctx, `[data-test="cbtn"]`, "click me (1)")

	click(t, ctx, `[data-test="cond-toggle"]`) // hide -> counter unmounts, state purged
	waitCount(t, ctx, `[data-test="cbtn"]`, 0)

	click(t, ctx, `[data-test="cond-toggle"]`) // show -> remounts fresh, not restored
	waitText(t, ctx, `[data-test="cbtn"]`, "click me (0)")
}
