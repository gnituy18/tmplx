package main

// Comment markers (<!--tx:id-->...<!--tx:id_e-->) are emitted only for instances
// that can actually be a swap target — i.e. their own swap root (tx_target == tx_id),
// which is the args-less case. A prop-bearing component can never be swapped
// standalone, so it carries no markers.

import (
	"testing"

	"github.com/chromedp/chromedp"
)

func TestMarkersOnlyForSwappable(t *testing.T) {
	count := func(path, sentinel string) int {
		ctx := newPage(t, path)
		waitCount(t, ctx, sentinel, 2) // both fixtures render two instances
		var n int
		js := `(() => {
			const w = document.createTreeWalker(document.documentElement, NodeFilter.SHOW_COMMENT)
			let n = 0
			while (w.nextNode()) if (w.currentNode.nodeValue.startsWith('tx:')) n++
			return n
		})()`
		if err := chromedp.Run(ctx, chromedp.Evaluate(js, &n)); err != nil {
			t.Fatalf("count markers on %s: %v", path, err)
		}
		return n
	}

	// Two args-less counters are each their own swap root: 2 x {start,end} = 4.
	if got := count("/counter-comp", `[data-test="cbtn"]`); got != 4 {
		t.Fatalf("counter-comp: want 4 tx: markers (2 swappable counters), got %d", got)
	}
	// Two prop-bearing badges can't be swap targets: no markers at all.
	if got := count("/badge", `[data-test="badge-text"]`); got != 0 {
		t.Fatalf("badge: want 0 tx: markers (prop-bearing, non-swappable), got %d", got)
	}
}
