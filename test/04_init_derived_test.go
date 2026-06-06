package main

// init() runs once at first render and may read/write derived vars (page + component).

import "testing"

func TestPageInitReadsDerived(t *testing.T) {
	ctx := newPage(t, "/init-derived-page")
	waitText(t, ctx, `[data-test="a"]`, "3")
	waitText(t, ctx, `[data-test="b"]`, "4")
	waitText(t, ctx, `[data-test="c"]`, "10")
	waitText(t, ctx, `[data-test="d"]`, "11")
	click(t, ctx, `[data-test="plus"]`)
	waitText(t, ctx, `[data-test="a"]`, "4")
	waitText(t, ctx, `[data-test="b"]`, "5")
	waitText(t, ctx, `[data-test="c"]`, "10")
	waitText(t, ctx, `[data-test="d"]`, "11")
	click(t, ctx, `[data-test="plus"]`)
	waitText(t, ctx, `[data-test="a"]`, "5")
	waitText(t, ctx, `[data-test="b"]`, "6")
	waitText(t, ctx, `[data-test="c"]`, "10")
	waitText(t, ctx, `[data-test="d"]`, "11")
}

func TestCompInitReadsDerived(t *testing.T) {
	ctx := newPage(t, "/init-derived-comp")
	waitText(t, ctx, `[data-test="comp-a"]`, "3")
	waitText(t, ctx, `[data-test="comp-b"]`, "4")
	click(t, ctx, `[data-test="comp-plus"]`)
	waitText(t, ctx, `[data-test="comp-a"]`, "4")
	waitText(t, ctx, `[data-test="comp-b"]`, "5")
	click(t, ctx, `[data-test="comp-plus"]`)
	waitText(t, ctx, `[data-test="comp-a"]`, "5")
	waitText(t, ctx, `[data-test="comp-b"]`, "6")
}
