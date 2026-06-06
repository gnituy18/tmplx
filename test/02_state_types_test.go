package main

import "testing"

// State variables can be types other than int/string. These check that bool,
// float64, and slice state render and drive behavior correctly.

// bool and float64 render via fmt-style formatting; slice via len().
func TestStateTypesRender(t *testing.T) {
	ctx := newPage(t, "/state-types")
	waitText(t, ctx, `[data-test="flag"]`, "flag: true")
	waitText(t, ctx, `[data-test="ratio"]`, "ratio: 1.5")
	waitText(t, ctx, `[data-test="nums-len"]`, "count: 3")
}

// A bool can be used directly as a tx-if condition, and a handler can flip it
// with negation; the conditionally-rendered node appears/disappears.
func TestStateBoolToggleCondition(t *testing.T) {
	ctx := newPage(t, "/state-types")
	waitText(t, ctx, `[data-test="when-on"]`, "flag is on")

	click(t, ctx, `[data-test="toggle"]`)
	waitText(t, ctx, `[data-test="flag"]`, "flag: false")
	if n := countNodes(t, ctx, `[data-test="when-on"]`); n != 0 {
		t.Errorf("after toggle off, conditional node still present (%d)", n)
	}

	click(t, ctx, `[data-test="toggle"]`)
	waitText(t, ctx, `[data-test="when-on"]`, "flag is on")
}
