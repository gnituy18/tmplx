package main

import "testing"

// tx-for renders one copy of the element per iteration. tx-key gives each row a
// stable identity. Conditions/loops must be pure (re-run in both passes).

// A loop exposing index and value; rows render in order with both bound.
func TestLoopIndexAndValue(t *testing.T) {
	ctx := newPage(t, "/loops")
	if n := countNodes(t, ctx, `[data-test="item"]`); n != 3 {
		t.Fatalf("item count: got %d, want 3", n)
	}
	for i, want := range []string{"0:a", "1:b", "2:c"} {
		if got := nthText(t, ctx, `[data-test="item"]`, i); got != want {
			t.Errorf("item[%d]: got %q, want %q", i, got, want)
		}
	}
}

// An empty slice produces zero rows (no leftover/placeholder node).
func TestLoopEmpty(t *testing.T) {
	ctx := newPage(t, "/loops")
	if n := countNodes(t, ctx, `[data-test="empty-item"]`); n != 0 {
		t.Errorf("empty loop rendered %d rows, want 0", n)
	}
}

// Ranging over an integer (Go 1.22) yields 0..n-1.
func TestLoopRangeOverInt(t *testing.T) {
	ctx := newPage(t, "/loops")
	if n := countNodes(t, ctx, `[data-test="num"]`); n != 4 {
		t.Fatalf("num count: got %d, want 4", n)
	}
	for i, want := range []string{"0", "1", "2", "3"} {
		if got := nthText(t, ctx, `[data-test="num"]`, i); got != want {
			t.Errorf("num[%d]: got %q, want %q", i, got, want)
		}
	}
}
