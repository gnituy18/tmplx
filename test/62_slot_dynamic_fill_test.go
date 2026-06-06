package main

import "testing"

// A slot fill is authored in the parent, so it interpolates PARENT state and
// re-renders when that state changes (the fill is part of the parent's subtree).
func TestSlotFillUsesParentStateAndUpdates(t *testing.T) {
	ctx := newPage(t, "/slot-state")
	waitText(t, ctx, `[data-test="dynfill"]`, "filled: hi")
	click(t, ctx, `[data-test="change"]`)
	waitText(t, ctx, `[data-test="dynfill"]`, "filled: bye")
}
