package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/chromedp/chromedp"
)

// fireInput dispatches n input events on selector back to back, within one
// JS task, so they always land inside a debounce window.
func fireInput(t *testing.T, ctx context.Context, selector string, n int) {
	t.Helper()
	js := fmt.Sprintf(`(() => {
		const el = document.querySelector(%q)
		if (!el) return false
		for (let i = 0; i < %d; i++) {
			el.value = "v" + i
			el.dispatchEvent(new Event("input"))
		}
		return true
	})()`, selector, n)
	var ok bool
	if err := chromedp.Run(ctx, chromedp.Evaluate(js, &ok)); err != nil || !ok {
		t.Fatalf("fireInput %s: ok=%v err=%v", selector, ok, err)
	}
}

// Without tx-debounce a burst sends one request per event; with it, only the
// last event of the burst is sent.
func TestDebounceCoalesces(t *testing.T) {
	ctx := newPage(t, "/debounce")

	fireInput(t, ctx, "#fast", 5)
	waitText(t, ctx, "#fastN", "5")

	fireInput(t, ctx, "#slow", 5)
	waitText(t, ctx, "#slowN", "1")
}

// Another event must flush pending debounces first, so handler order follows
// event order: slow++ from the input burst lands before the click's slow*10.
// (Without the flush, the click jumps the queue and the result is 1, not 10.)
func TestDebounceFlushOrder(t *testing.T) {
	ctx := newPage(t, "/debounce")

	fireInput(t, ctx, "#slow", 3)
	click(t, ctx, "#x10")
	waitText(t, ctx, "#slowN", "10")
}
