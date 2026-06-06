package main

import "testing"

// Derived variables (a var whose initializer references other vars) are
// recomputed every render. A chain a -> b -> c must cascade: changing a updates
// b, which updates c.
func TestDerivedChainCascades(t *testing.T) {
	ctx := newPage(t, "/derived-chain")
	waitText(t, ctx, `[data-test="dc-a"]`, "a=1")
	waitText(t, ctx, `[data-test="dc-b"]`, "b=2")  // a + 1
	waitText(t, ctx, `[data-test="dc-c"]`, "c=20") // b * 10

	click(t, ctx, `[data-test="dc-inc"]`) // a -> 2
	waitText(t, ctx, `[data-test="dc-a"]`, "a=2")
	waitText(t, ctx, `[data-test="dc-b"]`, "b=3")  // recomputed
	waitText(t, ctx, `[data-test="dc-c"]`, "c=30") // recomputed transitively
}
