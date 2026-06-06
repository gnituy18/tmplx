package main

import "testing"

// A component can be the target of a tx-for, and its props can reference the
// loop variables — each iteration produces a configured instance.
func TestLoopVarAsComponentProp(t *testing.T) {
	ctx := newPage(t, "/loopvar-prop")
	if n := countNodes(t, ctx, `[data-test="stat"]`); n != 3 {
		t.Fatalf("stat count: got %d, want 3", n)
	}
	for i, want := range []string{"x: 0", "y: 1", "z: 2"} {
		if got := nthText(t, ctx, `[data-test="stat"]`, i); got != want {
			t.Errorf("stat[%d]: got %q, want %q", i, got, want)
		}
	}
}
