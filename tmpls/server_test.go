package tmpls

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestServerPublishesPositionedDiagnostics drives the server over a scripted
// stdio session: initialize, then open a page with a type error, and asserts a
// publishDiagnostics with a precise range comes back.
func TestServerPublishesPositionedDiagnostics(t *testing.T) {
	dir := t.TempDir()
	must(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module tmplstest\n\ngo 1.25\n"), 0644))
	must(t, os.MkdirAll(filepath.Join(dir, "pages"), 0755))
	pagePath := filepath.Join(dir, "pages", "index.html")
	must(t, os.WriteFile(pagePath, []byte("<html><head></head><body>ok</body></html>"), 0644)) // valid on disk

	// the open (in-memory) version has a type error on line 2
	buggy := "<html><head><script type=\"text/tmplx\">\n  var n int = \"oops\"\n</script></head><body><span>{ n }</span></body></html>"
	uri := pathToURI(pagePath)

	var in bytes.Buffer
	writeMsg(&in, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize",
		"params": map[string]any{"rootUri": pathToURI(dir)}})
	writeMsg(&in, map[string]any{"jsonrpc": "2.0", "method": "textDocument/didOpen",
		"params": map[string]any{"textDocument": map[string]any{"uri": uri, "text": buggy}}})

	s := &server{docs: map[string]string{}, published: map[string]bool{}}
	var out bytes.Buffer
	s.run(&in, &out) // returns on EOF

	msgs := parseMsgs(out.Bytes())

	// initialize must have replied with capabilities
	var sawInit bool
	for _, m := range msgs {
		if res, ok := m["result"].(map[string]any); ok {
			if _, ok := res["capabilities"]; ok {
				sawInit = true
			}
		}
	}
	if !sawInit {
		t.Fatal("no initialize result with capabilities")
	}

	// find the publishDiagnostics for our page
	var diags []any
	for _, m := range msgs {
		if m["method"] != "textDocument/publishDiagnostics" {
			continue
		}
		p, _ := m["params"].(map[string]any)
		if p["uri"] == uri {
			diags, _ = p["diagnostics"].([]any)
		}
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostics for %s, got: %s", uri, out.String())
	}
	d0 := diags[0].(map[string]any)
	rng := d0["range"].(map[string]any)
	start := rng["start"].(map[string]any)
	if line := int(start["line"].(float64)); line != 1 { // 0-based; the var is on .html line 2
		t.Errorf("expected diagnostic on line 1 (0-based), got %d", line)
	}
	if msg, _ := d0["message"].(string); !strings.Contains(msg, "int") {
		t.Errorf("expected the type-error message, got %q", msg)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func writeMsg(buf *bytes.Buffer, v any) {
	b, _ := json.Marshal(v)
	fmt.Fprintf(buf, "Content-Length: %d\r\n\r\n", len(b))
	buf.Write(b)
}

func parseMsgs(b []byte) []map[string]any {
	var msgs []map[string]any
	r := bufio.NewReader(bytes.NewReader(b))
	for {
		length := 0
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				return msgs
			}
			line = strings.TrimRight(line, "\r\n")
			if line == "" {
				break
			}
			if v, ok := strings.CutPrefix(line, "Content-Length:"); ok {
				fmt.Sscanf(strings.TrimSpace(v), "%d", &length)
			}
		}
		body := make([]byte, length)
		if _, err := io.ReadFull(r, body); err != nil {
			return msgs
		}
		var m map[string]any
		json.Unmarshal(body, &m)
		msgs = append(msgs, m)
	}
}

// A page and a component sharing a rel path must each get their own
// diagnostics (the uri maps are keyed by the compiler's components// pages/
// identity, so neither overwrites the other).
func TestServerSameNamePageAndComponent(t *testing.T) {
	dir := t.TempDir()
	must(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module collide\n\ngo 1.25\n"), 0644))
	must(t, os.MkdirAll(filepath.Join(dir, "pages"), 0755))
	must(t, os.MkdirAll(filepath.Join(dir, "components"), 0755))
	compPath := filepath.Join(dir, "components", "foo.html")
	pagePath := filepath.Join(dir, "pages", "foo.html")
	pageText := "<html><head><script type=\"text/tmplx\">\nvar b int\n</script></head><body>y</body></html>"
	must(t, os.WriteFile(compPath, []byte("<script type=\"text/tmplx\">\nvar a int\n</script><p>x</p>"), 0644))
	must(t, os.WriteFile(pagePath, []byte(pageText), 0644))

	var in bytes.Buffer
	writeMsg(&in, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize",
		"params": map[string]any{"rootUri": pathToURI(dir)}})
	writeMsg(&in, map[string]any{"jsonrpc": "2.0", "method": "textDocument/didOpen",
		"params": map[string]any{"textDocument": map[string]any{"uri": pathToURI(pagePath), "text": pageText}}})

	s := &server{docs: map[string]string{}, published: map[string]bool{}}
	var out bytes.Buffer
	s.run(&in, &out)

	msgByURI := map[string]string{}
	for _, m := range parseMsgs(out.Bytes()) {
		if m["method"] != "textDocument/publishDiagnostics" {
			continue
		}
		p, _ := m["params"].(map[string]any)
		if diags, _ := p["diagnostics"].([]any); len(diags) > 0 {
			d0 := diags[0].(map[string]any)
			msgByURI[p["uri"].(string)], _ = d0["message"].(string)
		}
	}
	if !strings.Contains(msgByURI[pathToURI(compPath)], "a declared") {
		t.Errorf("component diagnostics misrouted: %v", msgByURI)
	}
	if !strings.Contains(msgByURI[pathToURI(pagePath)], "b declared") {
		t.Errorf("page diagnostics misrouted: %v", msgByURI)
	}
}

func TestServerFormatsScriptBlock(t *testing.T) {
	dir := t.TempDir()
	must(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module fmttest\n\ngo 1.25\n"), 0644))
	must(t, os.MkdirAll(filepath.Join(dir, "pages"), 0755))
	page := filepath.Join(dir, "pages", "index.html")
	must(t, os.WriteFile(page, []byte("x"), 0644))
	uri := pathToURI(page)
	messy := "<html><head><script type=\"text/tmplx\">\nvar   x    int\n</script></head><body>{ x }</body></html>"

	var in bytes.Buffer
	writeMsg(&in, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{"rootUri": pathToURI(dir)}})
	writeMsg(&in, map[string]any{"jsonrpc": "2.0", "method": "textDocument/didOpen", "params": map[string]any{"textDocument": map[string]any{"uri": uri, "text": messy}}})
	writeMsg(&in, map[string]any{"jsonrpc": "2.0", "id": 2, "method": "textDocument/formatting", "params": map[string]any{"textDocument": map[string]any{"uri": uri}}})

	s := &server{docs: map[string]string{}, published: map[string]bool{}}
	var out bytes.Buffer
	s.run(&in, &out)

	var edited string
	for _, m := range parseMsgs(out.Bytes()) {
		if edits, ok := m["result"].([]any); ok && len(edits) > 0 {
			if e, ok := edits[0].(map[string]any); ok {
				if nt, ok := e["newText"].(string); ok {
					edited = nt
				}
			}
		}
	}
	if !strings.Contains(edited, "var x int") {
		t.Errorf("expected gofmt'd script in the format edit, got: %q", edited)
	}
}

func TestServerGoToDefinition(t *testing.T) {
	dir := t.TempDir()
	must(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module deftest\n\ngo 1.25\n"), 0644))
	must(t, os.MkdirAll(filepath.Join(dir, "pages"), 0755))
	page := filepath.Join(dir, "pages", "index.html")
	// `counter` is declared on line 2 (0-based line 1); the usage is in the body.
	content := "<html><head><script type=\"text/tmplx\">\nvar counter int\n</script></head>\n<body><span>{ counter }</span></body></html>"
	must(t, os.WriteFile(page, []byte(content), 0644))
	uri := pathToURI(page)
	// the body line is line index 4: <body><span>{ counter }</span>...; put the cursor on "counter"
	col := strings.Index("<body><span>{ counter }</span></body></html>", "counter")

	var in bytes.Buffer
	writeMsg(&in, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{"rootUri": pathToURI(dir)}})
	writeMsg(&in, map[string]any{"jsonrpc": "2.0", "method": "textDocument/didOpen", "params": map[string]any{"textDocument": map[string]any{"uri": uri, "text": content}}})
	writeMsg(&in, map[string]any{"jsonrpc": "2.0", "id": 2, "method": "textDocument/definition", "params": map[string]any{"textDocument": map[string]any{"uri": uri}, "position": map[string]any{"line": 3, "character": col}}})

	s := &server{docs: map[string]string{}, published: map[string]bool{}}
	var out bytes.Buffer
	s.run(&in, &out)

	var gotLine float64 = -1
	for _, m := range parseMsgs(out.Bytes()) {
		if r, ok := m["result"].(map[string]any); ok {
			if rng, ok := r["range"].(map[string]any); ok {
				if st, ok := rng["start"].(map[string]any); ok {
					gotLine = st["line"].(float64)
				}
			}
		}
	}
	if int(gotLine) != 1 {
		t.Errorf("expected definition to jump to line 1 (var counter), got line %v", gotLine)
	}
}

func TestServerExternalDefAndHover(t *testing.T) {
	dir := t.TempDir()
	must(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module extdef\n\ngo 1.25\n"), 0644))
	must(t, os.MkdirAll(filepath.Join(dir, "pages"), 0755))
	page := filepath.Join(dir, "pages", "index.html")
	content := "<html><head><script type=\"text/tmplx\">\nimport \"fmt\"\nvar n int\nfunc logit() { fmt.Println(n) }\n</script></head>\n<body><button tx-onclick=\"logit()\">{ n }</button></body></html>"
	must(t, os.WriteFile(page, []byte(content), 0644))
	uri := pathToURI(page)
	lines := strings.Split(content, "\n")

	pos := func(i int, sub string) map[string]any { // 0-based LSP line + char
		return map[string]any{"line": i, "character": strings.Index(lines[i], sub)}
	}
	var in bytes.Buffer
	writeMsg(&in, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{"rootUri": pathToURI(dir)}})
	writeMsg(&in, map[string]any{"jsonrpc": "2.0", "method": "textDocument/didOpen", "params": map[string]any{"textDocument": map[string]any{"uri": uri, "text": content}}})
	// id 2: definition on Println (line 3) ; id 3: hover on n in { n } (line 5)
	writeMsg(&in, map[string]any{"jsonrpc": "2.0", "id": 2, "method": "textDocument/definition", "params": map[string]any{"textDocument": map[string]any{"uri": uri}, "position": pos(3, "Println")}})
	writeMsg(&in, map[string]any{"jsonrpc": "2.0", "id": 3, "method": "textDocument/hover", "params": map[string]any{"textDocument": map[string]any{"uri": uri}, "position": pos(5, "n }")}})

	s := &server{docs: map[string]string{}, published: map[string]bool{}}
	var out bytes.Buffer
	s.run(&in, &out)

	var defURI, hover string
	for _, m := range parseMsgs(out.Bytes()) {
		id, _ := m["id"].(float64)
		r, ok := m["result"].(map[string]any)
		if !ok {
			continue
		}
		if id == 2 {
			defURI, _ = r["uri"].(string)
		}
		if id == 3 {
			if cont, ok := r["contents"].(map[string]any); ok {
				hover, _ = cont["value"].(string)
			}
		}
	}
	if !strings.Contains(defURI, "/fmt/") || !strings.HasSuffix(defURI, ".go") {
		t.Errorf("fmt.Println definition should resolve into fmt's source, got uri %q", defURI)
	}
	if !strings.Contains(hover, "var n int") {
		t.Errorf("hover on n should report its type, got %q", hover)
	}
}
