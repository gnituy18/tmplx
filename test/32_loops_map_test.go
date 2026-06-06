package main

import "testing"

// Ranging over a map. Go randomizes map iteration order, so we assert the SET of
// rendered entries (all present), not their order. This documents that keyed map
// loops render correct CONTENT; visual ordering is intentionally not asserted.
func TestMapLoopRendersAllEntries(t *testing.T) {
	ctx := newPage(t, "/map-loop")
	waitCount(t, ctx, `[data-test="score"]`, 3)
	got := map[string]bool{}
	for i := 0; i < 3; i++ {
		got[nthText(t, ctx, `[data-test="score"]`, i)] = true
	}
	for _, want := range []string{"a=1;", "b=2;", "c=3;"} {
		if !got[want] {
			t.Errorf("map loop missing entry %q (got %v)", want, got)
		}
	}
}
