package main

import "testing"

// A func-prop is usually a callback, but it can also return a value and be
// called inside an interpolation. Here the parent binds `triple` and the child
// renders `{ compute(5) }` -> 15.
func TestFuncPropReturnValue(t *testing.T) {
	ctx := newPage(t, "/funcprop-return")
	waitText(t, ctx, `[data-test="calc"]`, "result: 15")
}
