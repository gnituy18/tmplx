package main

import "testing"

// A component can render another component. Two instances of the outer
// component each get their own inner component with independent state.
func TestNestedComponentsRenderAndIsolate(t *testing.T) {
	ctx := newPage(t, "/nested-comp")
	// outer props rendered
	if got := nthText(t, ctx, `[data-test="panel-heading"]`, 0); got != "Alpha" {
		t.Errorf("panel 0 heading: got %q", got)
	}
	if got := nthText(t, ctx, `[data-test="panel-heading"]`, 1); got != "Beta" {
		t.Errorf("panel 1 heading: got %q", got)
	}
	// each panel has its own inner counter
	if n := countNodes(t, ctx, `[data-test="cbtn"]`); n != 2 {
		t.Fatalf("inner counters: got %d, want 2", n)
	}
	// state isolation: clicking the first inner counter leaves the second at 0
	clickNth(t, ctx, `[data-test="cbtn"]`, 0)
	waitNthText(t, ctx, `[data-test="cbtn"]`, 0, "click me (1)")
	if got := nthText(t, ctx, `[data-test="cbtn"]`, 1); got != "click me (0)" {
		t.Errorf("second inner counter leaked state: got %q", got)
	}
}
