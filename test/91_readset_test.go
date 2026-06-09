package main

// Read-set UP: a component swap POSTs only the state at or under its target id,
// never the page or a sibling subtree. Sealed components can't read above
// themselves, so the server never needs more than the target's own subtree.

import (
	"testing"

	"github.com/chromedp/chromedp"
)

func TestReadSetSendsOnlyTargetSubtree(t *testing.T) {
	ctx := newPage(t, "/counter-comp")
	waitNthText(t, ctx, `[data-test="cbtn"]`, 0, "click me (0)")

	// The full state-key universe (page + both counters) before any swap.
	var universe []string
	if err := chromedp.Run(ctx, chromedp.Evaluate(
		`Object.keys(JSON.parse(document.getElementById('tx-saved').textContent))`, &universe)); err != nil {
		t.Fatalf("read initial state: %v", err)
	}

	var ready bool
	if err := chromedp.Run(ctx, chromedp.Evaluate(`(() => {
		window.__body = null
		const orig = window.fetch
		window.fetch = (u, o) => { window.__body = (o && o.body) || ""; return orig(u, o) }
		return true
	})()`, &ready)); err != nil {
		t.Fatalf("install fetch hook: %v", err)
	}

	clickNth(t, ctx, `[data-test="cbtn"]`, 0)
	waitNthText(t, ctx, `[data-test="cbtn"]`, 0, "click me (1)")

	var res struct {
		Target string   `json:"target"`
		Keys   []string `json:"keys"`
	}
	if err := chromedp.Run(ctx, chromedp.Evaluate(`(() => {
		const p = new URLSearchParams(window.__body || "")
		return { target: p.get("target"), keys: [...p.keys()] }
	})()`, &res)); err != nil {
		t.Fatalf("read body: %v", err)
	}

	// Of the real state keys, the swap must carry only the target's own.
	inUniverse := map[string]bool{}
	for _, k := range universe {
		inUniverse[k] = true
	}
	var sentState []string
	for _, k := range res.Keys {
		if inUniverse[k] {
			sentState = append(sentState, k)
		}
	}
	if len(sentState) != 1 || sentState[0] != res.Target {
		t.Fatalf("read-set: clicking one counter should POST only its own state %q; sent state keys %v (universe %v)",
			res.Target, sentState, universe)
	}
}
