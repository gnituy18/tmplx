package main

// Two sealed component instances on one page, each with fully independent state.

import "testing"

func TestCounterComponentsIndependent(t *testing.T) {
	ctx := newPage(t, "/counter-comp")
	waitNthText(t, ctx, `[data-test="cbtn"]`, 0, "click me (0)")
	waitNthText(t, ctx, `[data-test="cbtn"]`, 1, "click me (0)")
	clickNth(t, ctx, `[data-test="cbtn"]`, 0)
	clickNth(t, ctx, `[data-test="cbtn"]`, 0)
	waitNthText(t, ctx, `[data-test="cbtn"]`, 0, "click me (2)")
	waitNthText(t, ctx, `[data-test="cbtn"]`, 1, "click me (0)")
	clickNth(t, ctx, `[data-test="cbtn"]`, 1)
	waitNthText(t, ctx, `[data-test="cbtn"]`, 1, "click me (1)")
	waitNthText(t, ctx, `[data-test="cbtn"]`, 0, "click me (2)")
}
