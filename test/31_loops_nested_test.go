package main

import "testing"

// A tx-for nested inside another tx-for: the inner loop runs once per outer
// iteration, producing the cartesian product. The inner element references both
// the outer and inner loop variables as ADJACENT interpolations ({ r }{ c }).
func TestNestedLoopMatrix(t *testing.T) {
	ctx := newPage(t, "/nested-loop")
	if n := countNodes(t, ctx, `[data-test="row"]`); n != 2 {
		t.Fatalf("row count: got %d, want 2", n)
	}
	if n := countNodes(t, ctx, `[data-test="cell"]`); n != 6 {
		t.Fatalf("cell count: got %d, want 6 (2 rows x 3 cols)", n)
	}
	want := []string{"x1", "x2", "x3", "y1", "y2", "y3"}
	for i, w := range want {
		if got := nthText(t, ctx, `[data-test="cell"]`, i); got != w {
			t.Errorf("cell[%d]: got %q, want %q", i, got, w)
		}
	}
}
