package main

// Child calls a parent-bound func-prop (notify) alongside mutating its own local state.

import "testing"

func TestCompoundLocalAndFuncProp(t *testing.T) {
	ctx := newPage(t, "/compound")
	waitText(t, ctx, `[data-test="compound-local"]`, "0")
	waitText(t, ctx, `[data-test="compound-page-hits"]`, "page hits: 0")

	click(t, ctx, `[data-test="compound-btn"]`)
	waitText(t, ctx, `[data-test="compound-local"]`, "1")
	waitText(t, ctx, `[data-test="compound-page-hits"]`, "page hits: 1")

	click(t, ctx, `[data-test="compound-btn"]`)
	waitText(t, ctx, `[data-test="compound-local"]`, "2")
	waitText(t, ctx, `[data-test="compound-page-hits"]`, "page hits: 2")
}
