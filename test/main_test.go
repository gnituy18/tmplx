package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

var (
	serverAddr    string
	browserCtx    context.Context
	browserCancel context.CancelFunc
)

func TestMain(m *testing.M) {
	if err := exec.Command(os.ExpandEnv("$HOME/.local/bin/tmplx-dev")).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tmplx-dev failed: %v\n", err)
		os.Exit(1)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "listen: %v\n", err)
		os.Exit(1)
	}
	serverAddr = listener.Addr().String()

	mux := http.NewServeMux()
	for _, route := range Routes() {
		mux.Handle(route.Pattern, route.Handler)
	}
	server := &http.Server{Handler: mux}
	go server.Serve(listener)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), chromedp.DefaultExecAllocatorOptions[:]...)
	defer cancelAlloc()

	browserCtx, browserCancel = chromedp.NewContext(allocCtx)
	defer browserCancel()

	if err := chromedp.Run(browserCtx); err != nil {
		fmt.Fprintf(os.Stderr, "chromedp: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	_ = server.Shutdown(context.Background())
	os.Exit(code)
}

func newPage(t *testing.T, path string) context.Context {
	t.Helper()
	ctx, cancel := chromedp.NewContext(browserCtx)
	t.Cleanup(cancel)
	if err := chromedp.Run(ctx, chromedp.Navigate(urlFor(path))); err != nil {
		t.Fatalf("navigate %s: %v", path, err)
	}
	return ctx
}

func urlFor(path string) string {
	return fmt.Sprintf("http://%s%s", serverAddr, path)
}

func waitText(t *testing.T, ctx context.Context, selector, want string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	var got string
	js := fmt.Sprintf(`(() => { const el = document.querySelector(%q); return el ? el.textContent : ""; })()`, selector)
	for time.Now().Before(deadline) {
		_ = chromedp.Run(ctx, chromedp.Evaluate(js, &got))
		if got == want {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("waitText %s: got %q, want %q", selector, got, want)
}

func textOf(t *testing.T, ctx context.Context, selector string) string {
	t.Helper()
	var s string
	js := fmt.Sprintf(`(() => { const el = document.querySelector(%q); return el ? el.textContent : ""; })()`, selector)
	if err := chromedp.Run(ctx, chromedp.Evaluate(js, &s)); err != nil {
		t.Fatalf("text %s: %v", selector, err)
	}
	return s
}

// attrOf returns an element's attribute value (empty string if missing).
func attrOf(t *testing.T, ctx context.Context, selector, attr string) string {
	t.Helper()
	var s string
	js := fmt.Sprintf(`(() => { const el = document.querySelector(%q); return el ? (el.getAttribute(%q) || "") : ""; })()`, selector, attr)
	if err := chromedp.Run(ctx, chromedp.Evaluate(js, &s)); err != nil {
		t.Fatalf("attr %s.%s: %v", selector, attr, err)
	}
	return s
}

func click(t *testing.T, ctx context.Context, selector string) {
	t.Helper()
	js := fmt.Sprintf(`(() => { const el = document.querySelector(%q); if (!el) return false; el.click(); return true; })()`, selector)
	var ok bool
	if err := chromedp.Run(ctx, chromedp.Evaluate(js, &ok)); err != nil {
		t.Fatalf("click %s: %v", selector, err)
	}
	if !ok {
		t.Fatalf("click %s: no element matched", selector)
	}
}

func clickAll(t *testing.T, ctx context.Context, selector string, n int) {
	t.Helper()
	for range n {
		click(t, ctx, selector)
	}
}

func clickNth(t *testing.T, ctx context.Context, selector string, idx int) {
	t.Helper()
	js := fmt.Sprintf(`(() => { const el = document.querySelectorAll(%q)[%d]; if (!el) return false; el.click(); return true; })()`, selector, idx)
	var ok bool
	if err := chromedp.Run(ctx, chromedp.Evaluate(js, &ok)); err != nil {
		t.Fatalf("clickNth %s[%d]: %v", selector, idx, err)
	}
	if !ok {
		t.Fatalf("clickNth %s[%d]: no element", selector, idx)
	}
}

func waitNthText(t *testing.T, ctx context.Context, selector string, idx int, want string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	var got string
	for time.Now().Before(deadline) {
		js := fmt.Sprintf(`(() => { const el = document.querySelectorAll(%q)[%d]; return el ? el.textContent : ""; })()`, selector, idx)
		_ = chromedp.Run(ctx, chromedp.Evaluate(js, &got))
		if got == want {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("waitNthText %s[%d]: got %q, want %q", selector, idx, got, want)
}

func nthText(t *testing.T, ctx context.Context, selector string, idx int) string {
	t.Helper()
	var nodes []string
	js := fmt.Sprintf(`Array.from(document.querySelectorAll(%q)).map(e => e.textContent)`, selector)
	if err := chromedp.Run(ctx, chromedp.Evaluate(js, &nodes)); err != nil {
		t.Fatalf("eval: %v", err)
	}
	if idx >= len(nodes) {
		t.Fatalf("nthText %s[%d]: only %d nodes", selector, idx, len(nodes))
	}
	return nodes[idx]
}

// waitCount waits until exactly `want` nodes match selector (handy for
// asserting mount/unmount after an async swap).
func waitCount(t *testing.T, ctx context.Context, selector string, want int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	var got int
	js := fmt.Sprintf(`document.querySelectorAll(%q).length`, selector)
	for time.Now().Before(deadline) {
		_ = chromedp.Run(ctx, chromedp.Evaluate(js, &got))
		if got == want {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("waitCount %s: got %d, want %d", selector, got, want)
}

func countNodes(t *testing.T, ctx context.Context, selector string) int {
	t.Helper()
	var n int
	js := fmt.Sprintf(`document.querySelectorAll(%q).length`, selector)
	if err := chromedp.Run(ctx, chromedp.Evaluate(js, &n)); err != nil {
		t.Fatalf("count %s: %v", selector, err)
	}
	return n
}

func setValue(t *testing.T, ctx context.Context, selector, value string) {
	t.Helper()
	if err := chromedp.Run(ctx, chromedp.SendKeys(selector, value, chromedp.ByQuery)); err != nil {
		t.Fatalf("setValue %s: %v", selector, err)
	}
}
