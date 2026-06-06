package main

// show-counter: a state-mutating handler plus a pure func called in a template expression ({ showCounter() }).

import "testing"

func TestShowCounterTemplateFuncCall(t *testing.T) {
	ctx := newPage(t, "/show-counter")
	waitText(t, ctx, `[data-test="sc-h1"]`, "Counter: 0 (doubled: 0)")
	waitText(t, ctx, `[data-test="sc-show"]`, "1")
	click(t, ctx, `[data-test="sc-minus"]`)
	waitText(t, ctx, `[data-test="sc-h1"]`, "Counter: -1 (doubled: -2)")
	waitText(t, ctx, `[data-test="sc-show"]`, "0")
	click(t, ctx, `[data-test="sc-reset"]`)
	waitText(t, ctx, `[data-test="sc-show"]`, "1")
}
