// Package tmpls implements the tmplx language server (like gopls for Go): an LSP
// server over stdio that reuses the compiler package to publish diagnostics for
// .html pages and components. It is whole-project: a page's error can reference
// a component, so every change re-analyzes the project and republishes.
package tmpls

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go/types"
	"io"
	"io/fs"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gnituy18/tmplx/compiler"
	"golang.org/x/tools/go/packages"
)

// Run starts the language server, reading LSP messages from in and writing
// responses to out until EOF.
func Run(in io.Reader, out io.Writer) error {
	s := &server{docs: map[string]string{}, published: map[string]bool{}}
	return s.run(in, out)
}

type server struct {
	out   *bufio.Writer
	outMu sync.Mutex

	moduleDir, compDir, pageDir string
	imp                         types.Importer // cached user-package importer
	ready                       bool           // module + importer resolved

	docs      map[string]string // uri -> in-memory text for open documents
	published map[string]bool   // uris we've sent diagnostics for (to clear them)
}

// ---------------------------------------------------------------- JSON-RPC ---

type request struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

func (s *server) run(in io.Reader, out io.Writer) error {
	s.out = bufio.NewWriter(out)
	r := bufio.NewReader(in)
	for {
		req, err := readMessage(r)
		if err != nil {
			return err
		}
		s.handle(req)
	}
}

func readMessage(r *bufio.Reader) (*request, error) {
	length := 0
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
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
		return nil, err
	}
	var req request
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func (s *server) reply(id json.RawMessage, result any) {
	s.send(map[string]any{"jsonrpc": "2.0", "id": json.RawMessage(id), "result": result})
}

func (s *server) notify(method string, params any) {
	s.send(map[string]any{"jsonrpc": "2.0", "method": method, "params": params})
}

func (s *server) send(v any) {
	b, _ := json.Marshal(v)
	s.outMu.Lock()
	defer s.outMu.Unlock()
	fmt.Fprintf(s.out, "Content-Length: %d\r\n\r\n", len(b))
	s.out.Write(b)
	s.out.Flush()
}

// ----------------------------------------------------------------- handlers ---

func (s *server) handle(req *request) {
	switch req.Method {
	case "initialize":
		var p struct {
			RootURI string `json:"rootUri"`
		}
		json.Unmarshal(req.Params, &p)
		s.setup(uriToPath(p.RootURI))
		s.reply(req.ID, map[string]any{
			"capabilities": map[string]any{
				"textDocumentSync":           1, // full document sync
				"documentFormattingProvider": true,
				"definitionProvider":         true,
				"hoverProvider":              true,
				"documentSymbolProvider":     true,
				"referencesProvider":         true,
			},
			"serverInfo": map[string]any{"name": "tmpls"},
		})

	case "textDocument/didOpen":
		var p struct {
			TextDocument struct{ URI, Text string } `json:"textDocument"`
		}
		json.Unmarshal(req.Params, &p)
		s.docs[p.TextDocument.URI] = p.TextDocument.Text
		s.analyzeAndPublish()

	case "textDocument/didChange":
		var p struct {
			TextDocument   struct{ URI string } `json:"textDocument"`
			ContentChanges []struct {
				Text string `json:"text"`
			} `json:"contentChanges"`
		}
		json.Unmarshal(req.Params, &p)
		if len(p.ContentChanges) > 0 {
			s.docs[p.TextDocument.URI] = p.ContentChanges[len(p.ContentChanges)-1].Text
		}
		s.analyzeAndPublish()

	case "textDocument/didClose":
		var p struct {
			TextDocument struct{ URI string } `json:"textDocument"`
		}
		json.Unmarshal(req.Params, &p)
		delete(s.docs, p.TextDocument.URI)
		s.analyzeAndPublish()

	case "textDocument/formatting":
		var p struct {
			TextDocument struct{ URI string } `json:"textDocument"`
		}
		json.Unmarshal(req.Params, &p)
		text := s.docs[p.TextDocument.URI]
		formatted, err := compiler.Format([]byte(text))
		if err != nil || string(formatted) == text {
			s.reply(req.ID, nil) // don't edit on a Go syntax error or a no-op
			return
		}
		lines := strings.Split(text, "\n")
		end := lspPosition{Line: len(lines) - 1, Character: utf16Len(lines[len(lines)-1])}
		s.reply(req.ID, []map[string]any{{
			"range":   lspRange{Start: lspPosition{}, End: end},
			"newText": string(formatted),
		}})

	case "textDocument/definition":
		var p struct {
			TextDocument struct{ URI string } `json:"textDocument"`
			Position     lspPosition          `json:"position"`
		}
		json.Unmarshal(req.Params, &p)
		c, text, relToURI, uriToRel := s.build()
		rel := uriToRel[p.TextDocument.URI]
		if rel == "" {
			s.reply(req.ID, nil)
			return
		}
		line, col := lspToByte(text[p.TextDocument.URI], p.Position.Line, p.Position.Character)
		def, ok := c.Definition(rel, line, col)
		if !ok {
			s.reply(req.ID, nil)
			return
		}
		if def.Pkg != "" { // external Go symbol: load its package source and locate it
			if uri, rng, ok := s.resolveExternal(def.Pkg, def.Name); ok {
				s.reply(req.ID, map[string]any{"uri": uri, "range": rng})
				return
			}
			s.reply(req.ID, nil)
			return
		}
		defURI := relToURI[def.Path]
		s.reply(req.ID, map[string]any{"uri": defURI, "range": toLSPRange(text[defURI], def.Span)})

	case "textDocument/hover":
		var p struct {
			TextDocument struct{ URI string } `json:"textDocument"`
			Position     lspPosition          `json:"position"`
		}
		json.Unmarshal(req.Params, &p)
		c, text, _, uriToRel := s.build()
		rel := uriToRel[p.TextDocument.URI]
		if rel == "" {
			s.reply(req.ID, nil)
			return
		}
		line, col := lspToByte(text[p.TextDocument.URI], p.Position.Line, p.Position.Character)
		desc, span, ok := c.Hover(rel, line, col)
		if !ok {
			s.reply(req.ID, nil)
			return
		}
		s.reply(req.ID, map[string]any{
			"contents": map[string]any{"kind": "markdown", "value": "```go\n" + desc + "\n```"},
			"range":    toLSPRange(text[p.TextDocument.URI], span),
		})

	case "textDocument/documentSymbol":
		var p struct {
			TextDocument struct{ URI string } `json:"textDocument"`
		}
		json.Unmarshal(req.Params, &p)
		c, text, _, uriToRel := s.build()
		syms := []map[string]any{}
		for _, sym := range c.Symbols(uriToRel[p.TextDocument.URI]) {
			rng := toLSPRange(text[p.TextDocument.URI], sym.Span)
			syms = append(syms, map[string]any{"name": sym.Name, "kind": sym.Kind, "range": rng, "selectionRange": rng})
		}
		s.reply(req.ID, syms)

	case "textDocument/references":
		var p struct {
			TextDocument struct{ URI string } `json:"textDocument"`
			Position     lspPosition          `json:"position"`
		}
		json.Unmarshal(req.Params, &p)
		c, text, _, uriToRel := s.build()
		line, col := lspToByte(text[p.TextDocument.URI], p.Position.Line, p.Position.Character)
		locs := []map[string]any{}
		for _, span := range c.References(uriToRel[p.TextDocument.URI], line, col) {
			locs = append(locs, map[string]any{"uri": p.TextDocument.URI, "range": toLSPRange(text[p.TextDocument.URI], span)})
		}
		s.reply(req.ID, locs)

	case "shutdown":
		s.reply(req.ID, nil)

	case "exit":
		os.Exit(0)
	}
}

// ------------------------------------------------------------------ project ---

// setup resolves the module root, the components/pages dirs, and the
// user-package importer once. packages.Load shells out to `go list`, so it is
// cached and not redone per keystroke.
func (s *server) setup(root string) {
	if root == "" || s.ready {
		return
	}
	dir := root
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			log.Printf("tmpls: no go.mod under %s", root)
			return
		}
		dir = parent
	}
	s.moduleDir = dir
	s.compDir = filepath.Join(dir, "components")
	s.pageDir = filepath.Join(dir, "pages")

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedImports | packages.NeedDeps,
		Dir:  dir,
	}, "./...")
	if err == nil {
		loaded := map[string]*types.Package{}
		packages.Visit(pkgs, nil, func(pkg *packages.Package) {
			if pkg.Types != nil {
				loaded[pkg.PkgPath] = pkg.Types
			}
		})
		s.imp = compiler.PackageImporter(loaded)
	} else {
		log.Printf("tmpls: load packages: %v", err)
	}
	s.ready = true
}

// analyzeAndPublish rebuilds a Compiler from the project (open documents
// overriding disk), runs Diagnose, and publishes diagnostics per file.
// build assembles a Compiler from the project's component + page dirs, with open
// editor documents overriding what's on disk, and returns the path/text maps the
// LSP needs to translate between files, uris, and diagnostics.
func (s *server) build() (c *compiler.Compiler, text, relToURI, uriToRel map[string]string) {
	c = &compiler.Compiler{Importer: s.imp}
	text = map[string]string{}     // uri -> content
	relToURI = map[string]string{} // rel path (= a Component's Path) -> uri
	uriToRel = map[string]string{} // uri -> rel path

	add := func(dir string, page bool) {
		filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() || filepath.Ext(path) != ".html" {
				return nil
			}
			rel, _ := filepath.Rel(dir, path)
			rel = filepath.ToSlash(rel)
			uri := pathToURI(path)
			content, ok := s.docs[uri]
			if !ok {
				b, rerr := os.ReadFile(path)
				if rerr != nil {
					return nil
				}
				content = string(b)
			}
			relToURI[rel] = uri
			uriToRel[uri] = rel
			text[uri] = content
			if page {
				c.NewPage(rel, []byte(content))
			} else {
				c.NewComponent(rel, []byte(content))
			}
			return nil
		})
	}
	add(s.compDir, false)
	add(s.pageDir, true)
	return c, text, relToURI, uriToRel
}

func (s *server) analyzeAndPublish() {
	if !s.ready {
		return
	}
	c, text, relToURI, _ := s.build()
	uris := map[string]bool{} // every project file (to clear stale diagnostics)
	for _, uri := range relToURI {
		uris[uri] = true
	}

	byURI := map[string][]lspDiagnostic{}
	for _, d := range c.Diagnose() {
		uri := relToURI[d.Path]
		if uri == "" {
			continue // a diagnostic with no resolvable file (rare); skip
		}
		byURI[uri] = append(byURI[uri], toLSPDiagnostic(text[uri], d))
	}

	// publish for every project file, plus any uri we previously published, so
	// fixed files get cleared.
	for uri := range s.published {
		uris[uri] = true
	}
	s.published = map[string]bool{}
	for uri := range uris {
		diags := byURI[uri]
		if diags == nil {
			diags = []lspDiagnostic{}
		}
		s.published[uri] = true
		s.notify("textDocument/publishDiagnostics", map[string]any{
			"uri":         uri,
			"diagnostics": diags,
		})
	}
}

// -------------------------------------------------------------- LSP mapping ---

type lspDiagnostic struct {
	Range    lspRange `json:"range"`
	Severity int      `json:"severity"`
	Source   string   `json:"source"`
	Message  string   `json:"message"`
}

type lspRange struct {
	Start lspPosition `json:"start"`
	End   lspPosition `json:"end"`
}

type lspPosition struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

func toLSPDiagnostic(content string, d compiler.Diagnostic) lspDiagnostic {
	sev := 1 // Error
	if d.Severity == compiler.SevWarning {
		sev = 2
	}
	return lspDiagnostic{
		Range:    toLSPRange(content, d.Span),
		Severity: sev,
		Source:   "tmpls",
		Message:  d.Message,
	}
}

func toLSPRange(content string, span compiler.Span) lspRange {
	if span.Start.Line == 0 { // no precise location: point at the start of the file
		return lspRange{End: lspPosition{Character: 1}}
	}
	lines := strings.Split(content, "\n")
	conv := func(p compiler.Pos) lspPosition {
		line := p.Line - 1
		if line < 0 || line >= len(lines) {
			return lspPosition{Line: max(line, 0)}
		}
		b := min(p.Col-1, len(lines[line]))
		return lspPosition{Line: line, Character: utf16Len(lines[line][:b])}
	}
	return lspRange{Start: conv(span.Start), End: conv(span.End)}
}

// utf16Len is the number of UTF-16 code units in s (LSP's column unit).
func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		if r > 0xFFFF {
			n += 2 // surrogate pair
		} else {
			n++
		}
	}
	return n
}

// lspToByte converts a 0-based LSP (line, UTF-16 character) position in content
// to the compiler's 1-based (line, byte column).
func lspToByte(content string, line, character int) (int, int) {
	lines := strings.Split(content, "\n")
	if line < 0 || line >= len(lines) {
		return line + 1, character + 1
	}
	return line + 1, utf16ToByte(lines[line], character) + 1
}

// resolveExternal loads pkgPath's source and returns the LSP Location of the
// package-scope symbol name (e.g. "fmt", "Println" -> fmt's print.go). This is
// the lazy half of go-to-definition for imports/stdlib: only the clicked package
// is loaded, only on the request.
func (s *server) resolveExternal(pkgPath, name string) (string, lspRange, bool) {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedSyntax,
		Dir:  s.moduleDir,
	}, pkgPath)
	if err != nil || len(pkgs) == 0 || pkgs[0].Types == nil {
		return "", lspRange{}, false
	}
	obj := pkgs[0].Types.Scope().Lookup(name)
	if obj == nil || obj.Pos() == 0 {
		return "", lspRange{}, false
	}
	pos := pkgs[0].Fset.Position(obj.Pos())
	if pos.Filename == "" {
		return "", lspRange{}, false
	}
	start := lspPosition{Line: pos.Line - 1, Character: pos.Column - 1}
	end := lspPosition{Line: pos.Line - 1, Character: pos.Column - 1 + utf16Len(name)}
	return pathToURI(pos.Filename), lspRange{Start: start, End: end}, true
}

// utf16ToByte converts a UTF-16 code-unit offset in s to a byte offset.
func utf16ToByte(s string, u16 int) int {
	n := 0
	for i, r := range s {
		if n >= u16 {
			return i
		}
		if r > 0xFFFF {
			n += 2
		} else {
			n++
		}
	}
	return len(s)
}

func uriToPath(uri string) string {
	p := strings.TrimPrefix(uri, "file://")
	if dec, err := url.PathUnescape(p); err == nil {
		return dec
	}
	return p
}

func pathToURI(path string) string {
	return "file://" + (&url.URL{Path: path}).EscapedPath()
}
