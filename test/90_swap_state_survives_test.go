package main

// Page state survives a component swap (only the sealed subtree rebuilds, page state is preserved).

import "testing"

func TestStateSurvivesPageRerender(t *testing.T) {
	ctx := newPage(t, "/state-survives")
	waitText(t, ctx, `[data-test="ss-page-n"]`, "page: 0")
	waitText(t, ctx, `[data-test="cbtn"]`, "click me (0)")

	clickAll(t, ctx, `[data-test="cbtn"]`, 3)
	waitText(t, ctx, `[data-test="cbtn"]`, "click me (3)")

	click(t, ctx, `[data-test="ss-page-plus"]`)
	waitText(t, ctx, `[data-test="ss-page-n"]`, "page: 1")
	waitText(t, ctx, `[data-test="cbtn"]`, "click me (3)")
}
