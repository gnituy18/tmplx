package main

// input event: event.target.value flows into state and is echoed back with len().

import "testing"

func TestInputEventTargetValue(t *testing.T) {
	ctx := newPage(t, "/input")
	waitText(t, ctx, `[data-test="echo"]`, "You typed:  (0 chars)")
	setValue(t, ctx, `[data-test="text-input"]`, "hi")
	waitText(t, ctx, `[data-test="echo"]`, "You typed: hi (2 chars)")
}
