package main

// Component whose init() seeds state (n=42), which then persists across clicks.

import "testing"

func TestComponentInitRunsOnFirstRender(t *testing.T) {
	ctx := newPage(t, "/seeded")
	waitText(t, ctx, `[data-test="seeded-n"]`, "42")
	click(t, ctx, `[data-test="seeded-plus"]`)
	waitText(t, ctx, `[data-test="seeded-n"]`, "43")
	click(t, ctx, `[data-test="seeded-plus"]`)
	waitText(t, ctx, `[data-test="seeded-n"]`, "44")
}
