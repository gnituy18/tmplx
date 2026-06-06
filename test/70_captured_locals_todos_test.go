package main

// Captured loop locals (removeTodo(i)); plus tx-action form add (known gap, still failing).

import "testing"

func TestTodosInitial(t *testing.T) {
	ctx := newPage(t, "/todos")
	waitText(t, ctx, `[data-test="count"]`, "total: 2")
	if n := countNodes(t, ctx, `[data-test="todo-text"]`); n != 2 {
		t.Fatalf("initial todos: got %d, want 2", n)
	}
}

func TestTodosRemoveCapturedLocal(t *testing.T) {
	ctx := newPage(t, "/todos")
	// remove first todo ("first")
	click(t, ctx, `[data-test="remove"]`)
	waitText(t, ctx, `[data-test="count"]`, "total: 1")
	if got := nthText(t, ctx, `[data-test="todo-text"]`, 0); got != "second" {
		t.Errorf("remaining todo: got %q, want 'second'", got)
	}
}

func TestTodosAddViaTxAction(t *testing.T) {
	ctx := newPage(t, "/todos")
	setValue(t, ctx, `[data-test="new-todo"]`, "third")
	click(t, ctx, `[data-test="add"]`)
	waitText(t, ctx, `[data-test="count"]`, "total: 3")
	waitText(t, ctx, `[data-test="last-added"]`, "Last added: third")
	if got := nthText(t, ctx, `[data-test="todo-text"]`, 2); got != "third" {
		t.Errorf("new todo: got %q, want 'third'", got)
	}
}
