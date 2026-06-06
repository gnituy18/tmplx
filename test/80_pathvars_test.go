package main

import "testing"

// Path variables: a page file named `{seg}.html` becomes a route with a `{seg}`
// segment, and `//tx:path seg` binds that URL segment to a string var.
// pages/echo/{msg}.html -> GET /echo/{msg}.

// The segment value flows into the page as the bound var.
func TestPathVarBinds(t *testing.T) {
	ctx := newPage(t, "/echo/hello")
	waitText(t, ctx, `[data-test="echo-msg"]`, "msg: hello")
}

// A different segment value yields a different binding (same route, no rebuild
// of the test).
func TestPathVarDifferentValue(t *testing.T) {
	ctx := newPage(t, "/echo/world")
	waitText(t, ctx, `[data-test="echo-msg"]`, "msg: world")
}
