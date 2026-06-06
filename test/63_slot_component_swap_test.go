package main

// A sealed component placed inside a slot still swaps on its own events. Its
// instance id carries the slot '@' separator (e.g. tx-box-1@tx-counter-1), so
// tx_dispatch must parse the component name across '@' (not just ':'). This is
// the exact shape of every <tx-example-wrapper> demo on tmplx.org.

import "testing"

func TestSlotWrappedComponentSwaps(t *testing.T) {
	ctx := newPage(t, "/slot-comp-swap")
	waitText(t, ctx, `[data-test="cbtn"]`, "click me (0)")
	click(t, ctx, `[data-test="cbtn"]`)
	waitText(t, ctx, `[data-test="cbtn"]`, "click me (1)")
	click(t, ctx, `[data-test="cbtn"]`)
	waitText(t, ctx, `[data-test="cbtn"]`, "click me (2)")
}
