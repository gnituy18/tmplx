package main

// Counter: state-mutating handlers, a bound func-prop, derived values, ident-ref props, and an if/else-if/else sign.

import (
	"fmt"
	"testing"
)

func TestCounterFunctionProp(t *testing.T) {
	ctx := newPage(t, "/counter")
	waitText(t, ctx, `[data-test="counter-h1"]`, "Counter: 0 (doubled: 0)")
	clickNth(t, ctx, `[data-test="btn"]`, 0) // +3
	waitText(t, ctx, `[data-test="counter-h1"]`, "Counter: 3 (doubled: 6)")
	clickNth(t, ctx, `[data-test="btn"]`, 1) // +5
	waitText(t, ctx, `[data-test="counter-h1"]`, "Counter: 8 (doubled: 16)")
}

func TestCounterAnonMinus(t *testing.T) {
	ctx := newPage(t, "/counter")
	clickNth(t, ctx, `[data-test="btn"]`, 0) // +3
	waitText(t, ctx, `[data-test="counter-h1"]`, "Counter: 3 (doubled: 6)")
	clickNth(t, ctx, `[data-test="btn"]`, 0) // +3 -> 6
	waitText(t, ctx, `[data-test="counter-h1"]`, "Counter: 6 (doubled: 12)")
	click(t, ctx, `[data-test="minus"]`)
	waitText(t, ctx, `[data-test="counter-h1"]`, "Counter: 5 (doubled: 10)")
}

func TestCounterReset(t *testing.T) {
	ctx := newPage(t, "/counter")
	clickNth(t, ctx, `[data-test="btn"]`, 0) // +3
	waitText(t, ctx, `[data-test="counter-h1"]`, "Counter: 3 (doubled: 6)")
	click(t, ctx, `[data-test="reset"]`)
	waitText(t, ctx, `[data-test="counter-h1"]`, "Counter: 0 (doubled: 0)")
}

func TestCounterIdentRefStats(t *testing.T) {
	ctx := newPage(t, "/counter")
	if got := nthText(t, ctx, `[data-test="stat"]`, 0); got != "Live: 0" {
		t.Errorf("Live: got %q", got)
	}
	if got := nthText(t, ctx, `[data-test="stat"]`, 1); got != "Doubled: 0" {
		t.Errorf("Doubled: got %q", got)
	}
	clickNth(t, ctx, `[data-test="btn"]`, 0) // +3
	waitText(t, ctx, `[data-test="counter-h1"]`, "Counter: 3 (doubled: 6)")
	if got := nthText(t, ctx, `[data-test="stat"]`, 0); got != "Live: 3" {
		t.Errorf("Live after +3: got %q", got)
	}
	if got := nthText(t, ctx, `[data-test="stat"]`, 1); got != "Doubled: 6" {
		t.Errorf("Doubled after +3: got %q", got)
	}
}

func TestCounterConditionalSign(t *testing.T) {
	ctx := newPage(t, "/counter")
	waitText(t, ctx, `[data-test="sign"]`, "zero")
	clickNth(t, ctx, `[data-test="btn"]`, 0)
	waitText(t, ctx, `[data-test="sign"]`, "positive")
	wantCounters := []int{2, 1, 0, -1, -2}
	for i, c := range wantCounters {
		click(t, ctx, `[data-test="minus"]`)
		waitText(t, ctx, `[data-test="counter-h1"]`, fmt.Sprintf("Counter: %d (doubled: %d)", c, c*2))
		_ = i
	}
	waitText(t, ctx, `[data-test="sign"]`, "negative")
}
