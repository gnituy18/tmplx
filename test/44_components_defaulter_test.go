package main

// Prop and func-prop DEFAULTS, used when the parent omits them.

import "testing"

func TestPropDefaultsAndFuncVarSyntax(t *testing.T) {
	ctx := newPage(t, "/defaulter")
	waitText(t, ctx, `[data-test="def-label"]`, "default-label")
	waitText(t, ctx, `[data-test="def-n"]`, "0")
	click(t, ctx, `[data-test="def-btn"]`)
	waitText(t, ctx, `[data-test="def-n"]`, "1")
	click(t, ctx, `[data-test="def-btn"]`)
	waitText(t, ctx, `[data-test="def-n"]`, "2")
}
