package main

import (
	"strings"
	"testing"
)

// A <slot> may contain fallback content, rendered only when the parent provides
// no fill for it. With a fill, the fill replaces the fallback.
func TestSlotFallbackAndOverride(t *testing.T) {
	ctx := newPage(t, "/slot-fallback")

	// box[0]: no fill -> fallback content shows
	if got := strings.TrimSpace(nthText(t, ctx, `[data-test="box"]`, 0)); got != "fallback content" {
		t.Errorf("box 0 (no fill): got %q, want fallback content", got)
	}

	// box[1]: filled -> shows the fill, not the fallback
	if n := countNodes(t, ctx, `[data-test="box-fill"]`); n != 1 {
		t.Errorf("expected one filled slot, got %d", n)
	}
	if got := strings.TrimSpace(nthText(t, ctx, `[data-test="box"]`, 1)); got != "custom" {
		t.Errorf("box 1 (filled): got %q, want custom", got)
	}
}
