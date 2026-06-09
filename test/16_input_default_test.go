package main

// Uncontrolled input with a STATIC default (value="seed", not bound). The default
// seeds the field once; after that the DOM owns the value. A user edit survives
// the component's re-render: morph is attribute-only and the value attribute is
// unchanged ("seed"), and the dirty-value flag keeps the typed-over value. No
// value binding => no marker => uncontrolled.

import (
	"strings"
	"testing"

	"github.com/chromedp/chromedp"
)

func TestUncontrolledStaticDefaultSurvivesSwap(t *testing.T) {
	ctx := newPage(t, "/definput")

	var v0 string
	if err := chromedp.Run(ctx, chromedp.Evaluate(
		`document.querySelector('[data-test="def-input"]').value`, &v0)); err != nil {
		t.Fatalf("read initial value: %v", err)
	}
	if v0 != "seed" {
		t.Fatalf("static default not seeded: got %q, want %q", v0, "seed")
	}

	// Editing fires tx-oninput="n++", which swaps the whole component.
	if err := chromedp.Run(ctx,
		chromedp.Focus(`[data-test="def-input"]`, chromedp.ByQuery),
		chromedp.SendKeys(`[data-test="def-input"]`, "X", chromedp.ByQuery),
	); err != nil {
		t.Fatalf("type: %v", err)
	}
	waitText(t, ctx, `[data-test="def-count"]`, "1")

	var v1 string
	if err := chromedp.Run(ctx, chromedp.Evaluate(
		`document.querySelector('[data-test="def-input"]').value`, &v1)); err != nil {
		t.Fatalf("read edited value: %v", err)
	}
	// The swap must not re-apply the static default "seed": the edit survives.
	if !strings.Contains(v1, "seed") || !strings.Contains(v1, "X") {
		t.Fatalf("edit clobbered by swap: got %q, want it to keep both \"seed\" and the typed \"X\"", v1)
	}
}
