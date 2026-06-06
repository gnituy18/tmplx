package main

// Propful component: prop rewiring on parent re-render + rapid clicks (exercises page-swap morph).

import "testing"

func TestBadgePropRewiring(t *testing.T) {
	ctx := newPage(t, "/badge")
	waitNthText(t, ctx, `[data-test="badge-text"]`, 0, "A: 0")
	waitNthText(t, ctx, `[data-test="badge-text"]`, 1, "B: 0")
	clickNth(t, ctx, `[data-test="badge-tick"]`, 0)
	clickNth(t, ctx, `[data-test="badge-tick"]`, 0)
	waitNthText(t, ctx, `[data-test="badge-text"]`, 0, "A: 2")
	waitNthText(t, ctx, `[data-test="badge-text"]`, 1, "B: 0")
	clickNth(t, ctx, `[data-test="badge-tick"]`, 1)
	waitNthText(t, ctx, `[data-test="badge-text"]`, 1, "B: 1")
	waitNthText(t, ctx, `[data-test="badge-text"]`, 0, "A: 2")
}
