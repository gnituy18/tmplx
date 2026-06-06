package main

// Default slot + a named slot with multiple fills.

import "testing"

func TestSlotsDefaultAndNamed(t *testing.T) {
	ctx := newPage(t, "/slots")
	if got := textOf(t, ctx, `[data-test="card-title"]`); got != "Recipe" {
		t.Errorf("title: got %q, want 'Recipe'", got)
	}
	if got := textOf(t, ctx, `[data-test="default-fill"]`); got != "Mix and bake." {
		t.Errorf("default slot: got %q", got)
	}
	if got := textOf(t, ctx, `[data-test="footer-fill-1"]`); got != "Approved" {
		t.Errorf("footer slot 1: got %q", got)
	}
	if got := textOf(t, ctx, `[data-test="footer-fill-2"]`); got != "By chef" {
		t.Errorf("footer slot 2 (would be missing pre-fix): got %q", got)
	}
}
