package main

// A slot fill must survive the filled component's own handler firing. The
// component is args-less, so it used to be its own swap target — but the
// standalone dispatch passes nil for every slot, so the fill vanished on the
// first click (regression). A fill-bearing instance now bubbles its swap to
// the parent, which re-renders the fill.

import (
	"testing"

	"github.com/chromedp/chromedp"
)

func TestSlotFillSurvivesOwnSwap(t *testing.T) {
	ctx := newPage(t, "/slot-fill-swap")
	waitText(t, ctx, `[data-test="sfill"]`, "PRECIOUS")

	click(t, ctx, `[data-test="sp-toggle"]`)
	waitText(t, ctx, `[data-test="sp-open"]`, "OPEN")   // the swap landed
	waitText(t, ctx, `[data-test="sfill"]`, "PRECIOUS") // and the fill survived

	// not a swap root => no comment markers for the panel
	var n int
	js := `(() => {
		const w = document.createTreeWalker(document.documentElement, NodeFilter.SHOW_COMMENT)
		let n = 0
		while (w.nextNode()) if (w.currentNode.nodeValue.startsWith('tx:')) n++
		return n
	})()`
	if err := chromedp.Run(ctx, chromedp.Evaluate(js, &n)); err != nil {
		t.Fatalf("count markers: %v", err)
	}
	if n != 0 {
		t.Errorf("fill-bearing panel must not carry swap markers, found %d", n)
	}
}
