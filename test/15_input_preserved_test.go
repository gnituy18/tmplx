package main

// An <input> inside a component keeps its value AND focus across the component's
// own re-render. The swap morphs the marker-delimited region in place (no longer
// a range-splice that destroys the node), and morph skips tx-on* inputs, so the
// value the user is typing is never clobbered.

import (
	"testing"

	"github.com/chromedp/chromedp"
)

func TestInputValueAndFocusSurviveComponentSwap(t *testing.T) {
	ctx := newPage(t, "/liveinput")
	waitText(t, ctx, `[data-test="li-count"]`, "0")

	// Each keystroke fires tx-oninput="n++", which swaps the whole component.
	if err := chromedp.Run(ctx,
		chromedp.Focus(`[data-test="li-input"]`, chromedp.ByQuery),
		chromedp.SendKeys(`[data-test="li-input"]`, "abc", chromedp.ByQuery),
	); err != nil {
		t.Fatalf("type: %v", err)
	}
	waitText(t, ctx, `[data-test="li-count"]`, "3") // three swaps landed

	var val string
	var focused bool
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelector('[data-test="li-input"]').value`, &val),
		chromedp.Evaluate(`document.activeElement === document.querySelector('[data-test="li-input"]')`, &focused),
	); err != nil {
		t.Fatalf("read input: %v", err)
	}
	if val != "abc" {
		t.Fatalf("input value clobbered by swap: got %q, want %q", val, "abc")
	}
	if !focused {
		t.Fatalf("input lost focus across swap")
	}
}
