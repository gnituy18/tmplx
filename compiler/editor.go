package compiler

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/net/html"
)

// Def is a go-to-definition target. Either it names a tmplx file in the project
// (Path + Span — a state/prop/func decl, or a component's file), or it is an
// external Go symbol (Pkg + Name) the caller resolves in that package's source.
type Def struct {
	Path string // file identity (components/<rel> or pages/<rel>); "" if external
	Span Span
	Pkg  string // import path of an external package; "" if in-project
	Name string // the external symbol's name
}

// Definition resolves what is at the .html line/col in the component or page at
// path and returns its declaration. It uses go/types (so it handles state, props,
// derived, locals, funcs, imports, and stdlib like fmt.Println alike) plus tmplx
// component references (<tx-foo> -> its file, a prop attr -> the child's prop).
// Best-effort: analyzes once, resolves even when the project has diagnostics.
func (c *Compiler) Definition(path string, line, col int) (Def, bool) {
	c.analyze()
	comp := c.byPath(path)
	if comp == nil {
		return Def{}, false
	}
	off := comp.lineMap.Offset(line, col)
	if d, ok := comp.componentRefAt(off); ok {
		return d, true
	}
	sem := comp.semantic()
	if sem == nil {
		return Def{}, false
	}
	obj := sem.objOf(sem.identAt(off))
	if obj == nil || obj.Pkg() == nil {
		return Def{}, false // builtin (len, string, ...) or unresolved
	}
	if obj.Pkg().Path() != "tx_probe" {
		return Def{Pkg: obj.Pkg().Path(), Name: obj.Name()}, true // external symbol
	}
	if obj.Pos() == token.NoPos {
		return Def{}, false
	}
	h, ok := sem.goToHTML(sem.fset.Position(obj.Pos()).Offset)
	if !ok {
		return Def{}, false
	}
	return Def{Path: comp.Path, Span: Span{
		Start: comp.lineMap.Pos(h),
		End:   comp.lineMap.Pos(h + len(obj.Name())),
	}}, true
}

// Hover returns a one-line description (the go/types object string) of the symbol
// at the .html line/col, plus the .html span of the hovered identifier.
func (c *Compiler) Hover(path string, line, col int) (string, Span, bool) {
	c.analyze()
	comp := c.byPath(path)
	if comp == nil {
		return "", Span{}, false
	}
	sem := comp.semantic()
	if sem == nil {
		return "", Span{}, false
	}
	id := sem.identAt(comp.lineMap.Offset(line, col))
	obj := sem.objOf(id)
	if obj == nil {
		return "", Span{}, false
	}
	span := Span{}
	if h, ok := sem.goToHTML(sem.fset.Position(id.Pos()).Offset); ok {
		span = Span{Start: comp.lineMap.Pos(h), End: comp.lineMap.Pos(h + len(id.Name))}
	}
	qual := func(p *types.Package) string {
		if p.Path() == "tx_probe" {
			return "" // the synthetic package: show local symbols unqualified
		}
		return p.Name()
	}
	return types.ObjectString(obj, qual), span, true
}

// Symbol is one entry in a file's outline (document symbols).
type Symbol struct {
	Name string
	Kind int // an LSP SymbolKind: symField, symFunction, or symVariable
	Span Span
}

// LSP SymbolKind values, as the protocol numbers them.
const (
	symField    = 8  // value props (a component's inputs)
	symFunction = 12 // funcs and func-typed props
	symVariable = 13 // state and derived
)

// Symbols returns the outline of the component or page at path: its state, props,
// derived/path vars, and funcs, each with its declaration span.
func (c *Compiler) Symbols(path string) []Symbol {
	c.analyze()
	comp := c.byPath(path)
	if comp == nil {
		return nil
	}
	var out []Symbol
	for _, v := range comp.Vars {
		kind := symVariable
		switch {
		case v.IsFunc():
			kind = symFunction
		case v.Kind == VarKindProp:
			kind = symField
		}
		if v.Span != (Span{}) {
			out = append(out, Symbol{Name: v.Name, Kind: kind, Span: v.Span})
		}
	}
	for _, f := range comp.Funcs {
		if f.Span != (Span{}) {
			out = append(out, Symbol{Name: f.Decl.Name.Name, Kind: symFunction, Span: f.Span})
		}
	}
	return out
}

// References returns every .html span in the component/page at path that refers
// to the symbol at line/col (its declaration plus all uses, across the script and
// template). Resolution is within the one file via go/types.
func (c *Compiler) References(path string, line, col int) []Span {
	c.analyze()
	comp := c.byPath(path)
	if comp == nil {
		return nil
	}
	sem := comp.semantic()
	if sem == nil {
		return nil
	}
	target := sem.objOf(sem.identAt(comp.lineMap.Offset(line, col)))
	if target == nil {
		return nil
	}
	var spans []Span
	mark := func(id *ast.Ident, obj types.Object) {
		if obj != target {
			return
		}
		if h, ok := sem.goToHTML(sem.fset.Position(id.Pos()).Offset); ok {
			spans = append(spans, Span{Start: comp.lineMap.Pos(h), End: comp.lineMap.Pos(h + len(id.Name))})
		}
	}
	for id, obj := range sem.info.Uses {
		mark(id, obj)
	}
	for id, obj := range sem.info.Defs {
		mark(id, obj)
	}
	return spans
}

func (c *Compiler) byPath(path string) *Component {
	for _, comp := range c.components {
		if comp.Path == path {
			return comp
		}
	}
	for _, page := range c.pages {
		if page.Path == path {
			return page
		}
	}
	return nil
}

// semantic is a go/types view of a component built for editor queries: the
// script block verbatim plus every template Go fragment (see collectExprs),
// type-checked, with a source map and captured Uses/Defs. It is independent of
// the diagnostic probe and best-effort (type errors are ignored).
type semantic struct {
	fset *token.FileSet
	file *ast.File
	info *types.Info
	maps []semMap
}

// semMap ties length bytes at goStart in sem.go to htmlStart in the .html
// source. The text is copied verbatim, so offsets convert by addition.
type semMap struct{ goStart, length, htmlStart int }

func (comp *Component) semantic() *semantic {
	if comp.TmplxScriptNode == nil || comp.TmplxScriptNode.FirstChild == nil {
		return nil
	}
	var b strings.Builder
	var maps []semMap
	b.WriteString("package tx_probe\n")
	script := comp.TmplxScriptNode.FirstChild.Data
	maps = append(maps, semMap{goStart: b.Len(), length: len(script), htmlStart: comp.scriptStart})
	b.WriteString(script)
	b.WriteString("\nfunc tx_tmpl() {\n")
	comp.collectExprs(comp.TemplateNode, &b, &maps)
	b.WriteString("\n}\n")

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "sem.go", b.String(), parser.SkipObjectResolution)
	if err != nil {
		return nil
	}
	info := &types.Info{
		Defs: map[*ast.Ident]types.Object{},
		Uses: map[*ast.Ident]types.Object{},
	}
	(&types.Config{Importer: comp.compiler.Importer, Error: func(error) {}}).Check("tx_probe", fset, []*ast.File{file}, info)
	return &semantic{fset: fset, file: file, info: info, maps: maps}
}

// collectExprs emits every Go fragment in the template — text interpolations,
// attribute interpolations, tx-if/tx-for/tx-key/tx-action, tx-on handlers, and
// component prop values — into the probe function, recording each fragment's
// .html source span so a click anywhere in template Go resolves through go/types.
// Emission is flat (no scope nesting), so a tx-for loop variable used in a child
// element won't resolve; everything else does.
func (comp *Component) collectExprs(node *html.Node, b *strings.Builder, maps *[]semMap) {
	emit := func(prefix, raw, suffix string, htmlOff int) {
		b.WriteString(prefix)
		*maps = append(*maps, semMap{goStart: b.Len(), length: len(raw), htmlStart: htmlOff})
		b.WriteString(raw)
		b.WriteString(suffix)
	}
	interp := func(s string, base int) {
		comp.scanTmplStr(s, false, func(rune) {}, func(expr string, off int) error {
			if _, err := parser.ParseExpr(expr); err == nil {
				emit("_ = (", expr, ")\n", base+off)
			}
			return nil
		})
	}
	switch node.Type {
	case html.TextNode:
		if node.Parent != nil {
			if _, ign := hasAttr(node.Parent, "tx-ignore"); !ign && !isVerbatimSerialize(node.Parent.DataAtom) {
				if base, ok := comp.nodePos[node]; ok {
					interp(node.Data, base)
				}
			}
		}
	case html.ElementNode:
		isComp := comp.compiler.components[node.Data] != nil
		for _, attr := range node.Attr {
			off, ok := comp.attrValueOffset(node, attr.Key)
			if ok && attr.Val != "" {
				parses := func() bool { _, err := parser.ParseExpr(attr.Val); return err == nil }
				switch {
				case attr.Key == "tx-if" || attr.Key == "tx-else-if":
					if parses() {
						emit("if (", attr.Val, ") {\n}\n", off)
					}
				case attr.Key == "tx-for":
					emit("for ", attr.Val, " {\n}\n", off)
				case attr.Key == "tx-key" || attr.Key == "tx-action":
					if parses() {
						emit("_ = (", attr.Val, ")\n", off)
					}
				case strings.HasPrefix(attr.Key, "tx-on"):
					emit("{\n", attr.Val, "\n}\n", off)
				case isComp && !strings.HasPrefix(attr.Key, "tx-") && attr.Key != "slot":
					if parses() {
						emit("_ = (", attr.Val, ")\n", off)
					}
				default:
					if strings.Contains(attr.Val, "{") {
						interp(attr.Val, off)
					}
				}
			}
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		comp.collectExprs(c, b, maps)
	}
}

func (s *semantic) htmlToGo(off int) (int, bool) {
	for _, m := range s.maps {
		if off >= m.htmlStart && off < m.htmlStart+m.length {
			return m.goStart + (off - m.htmlStart), true
		}
	}
	return 0, false
}

func (s *semantic) goToHTML(off int) (int, bool) {
	for _, m := range s.maps {
		if off >= m.goStart && off < m.goStart+m.length {
			return m.htmlStart + (off - m.goStart), true
		}
	}
	return 0, false
}

// identAt finds the innermost identifier covering the .html byte offset.
func (s *semantic) identAt(htmlOff int) *ast.Ident {
	goOff, ok := s.htmlToGo(htmlOff)
	if !ok {
		return nil
	}
	var found *ast.Ident
	ast.Inspect(s.file, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			if p := s.fset.Position(id.Pos()).Offset; goOff >= p && goOff < p+len(id.Name) {
				found = id
			}
		}
		return true
	})
	return found
}

func (s *semantic) objOf(id *ast.Ident) types.Object {
	if id == nil {
		return nil
	}
	if obj := s.info.Uses[id]; obj != nil {
		return obj
	}
	return s.info.Defs[id]
}

// componentRefAt resolves a click on a <tx-foo> usage: on the tag name -> the
// component's file; on a prop attribute's key -> the child component's prop decl.
func (comp *Component) componentRefAt(off int) (Def, bool) {
	for node, start := range comp.nodePos {
		if node.Type != html.ElementNode {
			continue
		}
		used, isComp := comp.compiler.components[node.Data]
		if !isComp {
			continue
		}
		gt := bytes.IndexByte(comp.content[start:], '>')
		if gt < 0 || off < start || off > start+gt {
			continue
		}
		if off <= start+1+len(node.Data) { // on "<tx-foo"
			return Def{Path: used.Path}, true
		}
		for _, attr := range node.Attr {
			ks, ke, found := comp.attrKeyRange(node, attr.Key)
			if found && off >= ks && off < ke {
				if v := used.varByName(attr.Key); v != nil && v.Kind == VarKindProp {
					return Def{Path: used.Path, Span: v.Span}, true
				}
			}
		}
	}
	return Def{}, false
}

// attrKeyRange returns the byte range of an attribute's key within node's open tag.
func (comp *Component) attrKeyRange(node *html.Node, key string) (int, int, bool) {
	start, ok := comp.nodePos[node]
	if !ok {
		return 0, 0, false
	}
	gt := bytes.IndexByte(comp.content[start:], '>')
	if gt < 0 {
		return 0, 0, false
	}
	tag := comp.content[start : start+gt]
	for i := 1; i+len(key) <= len(tag); i++ {
		if (tag[i-1] == ' ' || tag[i-1] == '\t' || tag[i-1] == '\n') && string(tag[i:i+len(key)]) == key {
			after := i + len(key)
			if after == len(tag) || tag[after] == '=' || tag[after] == ' ' || tag[after] == '\t' || tag[after] == '\n' || tag[after] == '/' {
				return start + i, start + i + len(key), true
			}
		}
	}
	return 0, 0, false
}

// attrValueOffset returns the byte offset of an attribute's value (the first
// character inside the quotes) within node's open tag.
func (comp *Component) attrValueOffset(node *html.Node, key string) (int, bool) {
	start, ok := comp.nodePos[node]
	if !ok {
		return 0, false
	}
	_, ke, found := comp.attrKeyRange(node, key)
	if !found {
		return 0, false
	}
	tag := comp.content[start : start+bytes.IndexByte(comp.content[start:], '>')]
	j := ke - start
	for j < len(tag) && (tag[j] == ' ' || tag[j] == '\t' || tag[j] == '\n') {
		j++
	}
	if j >= len(tag) || tag[j] != '=' {
		return 0, false // boolean attribute, no value
	}
	j++
	for j < len(tag) && (tag[j] == ' ' || tag[j] == '\t' || tag[j] == '\n') {
		j++
	}
	if j < len(tag) && (tag[j] == '"' || tag[j] == '\'') {
		j++
	}
	return start + j, true
}

// Format formats the <script type="text/tmplx"> block of a tmplx source file
// the way gopls would (gofmt: tabs for nested code) while preserving the HTML
// indentation of the <script> tag: every Go line is prefixed with the tag's own
// indent, so a <script> nested 8 spaces deep keeps its body at 8 spaces + tabs.
// The surrounding HTML is left untouched. With no tmplx script src is returned
// unchanged; if the script isn't valid Go, src is returned plus the error.
func Format(src []byte) ([]byte, error) {
	tagStart, contentStart, contentEnd := -1, -1, -1
	z := html.NewTokenizer(bytes.NewReader(src))
	for offset := 0; ; {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}
		raw := z.Raw()
		if tt == html.StartTagToken {
			if name, hasAttr := z.TagName(); string(name) == "script" {
				tmplx := false
				for hasAttr {
					var k, v []byte
					k, v, hasAttr = z.TagAttr()
					if string(k) == "type" && string(v) == "text/tmplx" {
						tmplx = true
					}
				}
				if tmplx {
					tagStart, contentStart = offset, offset+len(raw)
					if z.Next() == html.TextToken {
						contentEnd = contentStart + len(z.Raw())
					}
					break
				}
			}
		}
		offset += len(raw)
	}
	if contentEnd < 0 {
		return src, nil // no tmplx script, or it's empty
	}

	// Format as a full file (synthetic package clause) so every top-level decl
	// lands at column 0. format.Source's fragment mode instead indents the whole
	// result by the first code line's indentation, which over-indents lines 2+ of
	// an already-indented script. Then drop the package clause.
	formatted, err := format.Source(append([]byte("package p\n"), src[contentStart:contentEnd]...))
	if err != nil {
		return src, fmt.Errorf("format script: %w", err)
	}
	if nl := bytes.IndexByte(formatted, '\n'); nl >= 0 {
		formatted = formatted[nl+1:]
	}

	// the <script> tag's own indentation: the whitespace before '<' on its line
	// (only when the tag is alone on its line, else don't add a base indent)
	indent := src[bytes.LastIndexByte(src[:tagStart], '\n')+1 : tagStart]
	if len(bytes.TrimLeft(indent, " \t")) != 0 {
		indent = nil
	}

	var body []byte
	for i, line := range bytes.Split(bytes.TrimSpace(formatted), []byte("\n")) {
		if i > 0 {
			body = append(body, '\n')
		}
		if len(line) > 0 {
			body = append(body, indent...)
			body = append(body, line...)
		}
	}

	out := make([]byte, 0, len(src)+len(body))
	out = append(out, src[:contentStart]...)
	out = append(out, '\n')
	out = append(out, body...)
	out = append(out, '\n')
	out = append(out, indent...)
	out = append(out, src[contentEnd:]...)
	return out, nil
}
