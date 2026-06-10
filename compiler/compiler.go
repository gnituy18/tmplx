package compiler

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/importer"
	"go/parser"
	"go/printer"
	"go/scanner"
	"go/token"
	"go/types"
	"maps"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"golang.org/x/text/unicode/norm"
	"golang.org/x/tools/go/ast/astutil"
)

// Compiler holds one compilation: add sources with NewComponent / NewPage,
// then call Compile. Importer resolves the user's packages (PackageImporter);
// nil resolves the standard library only. ComponentsDir and PagesDir are only
// how errors and editor queries name files (default components / pages); set
// them before adding sources. A Compiler keeps no package-level state, so
// distinct Compilers run concurrently.
type Compiler struct {
	Importer      types.Importer
	PackageName   string
	HandlerPrefix string

	ComponentsDir string
	components    map[string]*Component
	PagesDir      string
	pages         []*Component

	errs []error
}

// NewComponent registers a reusable component read from path (slash-separated,
// relative to the components dir, e.g. "foo/bar.html") with the given file
// bytes. The component is named tx-<stem> (slashes become dashes) and the file
// is identified as <ComponentsDir>/<path> in diagnostics and editor queries.
// An invalid name or a duplicate is recorded and surfaced by Compile.
func (c *Compiler) NewComponent(path string, content []byte) {
	if c.components == nil {
		c.components = map[string]*Component{}
	}
	if c.ComponentsDir == "" {
		c.ComponentsDir = "components"
	}

	stem, ok := strings.CutSuffix(path, ".html")
	// from here on path is namespaced: a page and a component may share a file
	// name, so errors and editor queries identify files by their dir
	path = strings.TrimSuffix(c.ComponentsDir, "/") + "/" + path
	if !ok || stem == "" || strings.HasSuffix(stem, "/") {
		c.errs = append(c.errs, fmt.Errorf("%s: invalid filename: want name.html", path))
		return
	}

	name := "tx-" + strings.ReplaceAll(stem, "/", "-")
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			c.errs = append(c.errs, fmt.Errorf("%s: invalid character %q in <%s>: use only a-z, 0-9, -, _", path, string(r), name))
			return
		}
	}
	if existing, ok := c.components[name]; ok {
		c.errs = append(c.errs, fmt.Errorf("%s: duplicate component <%s>, first defined in %s", path, name, existing.Path))
		return
	}

	c.components[name] = &Component{
		compiler: c,
		content:  content,
		lineMap:  newLineMap(content),
		Type:     CompTypeComp,
		Path:     path,
		Name:     name,
		GoName:   goIdent(name),

		ChildCounters:       map[string]*Counter{},
		Children:            map[*html.Node]*Child{},
		EventHandlers:       map[*html.Node][]*EventHandler{},
		EventHandlerCounter: &Counter{},
	}
}

// NewPage registers a page read from path (slash-separated, relative to the
// pages dir) with the given file bytes. The route is derived from the path
// (index.html -> "/", foo/bar.html -> "/foo/bar") and the file is identified
// as <PagesDir>/<path> in diagnostics and editor queries. A duplicate route is
// recorded and surfaced by Compile.
func (c *Compiler) NewPage(path string, content []byte) {
	if c.PagesDir == "" {
		c.PagesDir = "pages"
	}

	base := path[strings.LastIndex(path, "/")+1:]
	dir, _ := strings.CutSuffix(path, base)
	stem, ok := strings.CutSuffix(base, ".html")
	path = strings.TrimSuffix(c.PagesDir, "/") + "/" + path // namespaced; see NewComponent
	if !ok || stem == "" {
		c.errs = append(c.errs, fmt.Errorf("%s: invalid filename: want name.html", path))
		return
	}

	route := "/" + dir
	if stem != "index" {
		route += stem
	}
	// macOS filenames arrive NFD-decomposed (e + combining accent) but
	// browsers request NFC (precomposed); ServeMux compares bytes, so
	// normalize the route to the form browsers send
	route = norm.NFC.String(route)
	// a ServeMux pattern ending in "/" matches the whole subtree, so a bare
	// index route would swallow every unmatched URL under it; {$} pins it to
	// exactly the directory URL (a {rest...}.html page is the explicit
	// catch-all)
	if strings.HasSuffix(route, "/") {
		route += "{$}"
	}
	for _, p := range c.pages {
		if p.Name == route {
			c.errs = append(c.errs, fmt.Errorf("%s: duplicate page route %s, first defined in %s", path, route, p.Path))
			return
		}
	}

	c.pages = append(c.pages, &Component{
		compiler: c,
		content:  content,
		lineMap:  newLineMap(content),
		Type:     CompTypePage,
		Path:     path,
		Name:     route,
		// route starts with "/" (-> _SL_), so this is always tx_SL_... -
		// inside the reserved tx_ namespace, the encoding owning the underscore
		GoName: "tx" + goIdent(route),

		ChildCounters:       map[string]*Counter{},
		Children:            map[*html.Node]*Child{},
		EventHandlers:       map[*html.Node][]*EventHandler{},
		EventHandlerCounter: &Counter{},
	})
}

// analyze runs the front of the pipeline - naming, parse, type-check (probe) -
// and returns the sorted components plus any diagnostics, without codegen, so
// Diagnose and the editor queries reuse it. It mutates the registered
// Components, so call it once per Compiler.
func (c *Compiler) analyze() ([]*Component, error) {
	if c.Importer == nil {
		c.Importer = &lockedImporter{imp: importer.Default()}
	}
	if c.PackageName == "" {
		c.PackageName = "main"
	}
	if c.HandlerPrefix == "" {
		c.HandlerPrefix = "/tx/"
	}
	if err := sortedJoin(errors.Join(c.errs...)); err != nil {
		return nil, err
	}

	// parse the components in parallel; each goroutine owns one error slot, so
	// there is no shared mutable state to lock
	var wg sync.WaitGroup
	components := slices.SortedFunc(maps.Values(c.components), func(a, b *Component) int {
		return strings.Compare(a.Name, b.Name)
	})
	compErrs := make([]error, len(components))
	for i, comp := range components {
		wg.Go(func() {
			var errs []error
			defer func() { compErrs[i] = errors.Join(errs...) }()

			nodes, err := html.ParseFragment(bytes.NewReader(comp.content), &html.Node{
				Data:     "body",
				DataAtom: atom.Body,
				Type:     html.ElementNode,
			})
			if err != nil {
				errs = append(errs, comp.errf("invalid HTML: %w", err))
				return
			}

			comp.TemplateNode = newTemplateNode()
			for _, node := range nodes {
				if isTmplxScriptNode(node) {
					if comp.TmplxScriptNode != nil {
						errs = append(errs, comp.errf("multiple <script type=\"text/tmplx\"> elements (only one allowed)"))
						return
					}
					comp.TmplxScriptNode = node
				} else if node.DataAtom == atom.Style {
					if comp.StyleNode != nil {
						errs = append(errs, comp.errf("multiple <style> elements (only one allowed)"))
						return
					}
					comp.StyleNode = node
				} else {
					comp.TemplateNode.AppendChild(node)
				}
			}

			comp.annotatePositions()
			errs = append(errs, comp.parseScript())
			if err := comp.checkImports(); err != nil {
				errs = append(errs, err)
				return
			}
			errs = append(errs, comp.inferVarTypes())
			errs = append(errs, comp.parseSlots(comp.TemplateNode, false))
			comp.inferEffects()
		})
	}

	pageErrs := make([]error, len(c.pages))
	for i, page := range c.pages {
		wg.Go(func() {
			var errs []error
			defer func() { pageErrs[i] = errors.Join(errs...) }()

			parsed, err := html.Parse(bytes.NewReader(page.content))
			if err != nil {
				errs = append(errs, page.errf("invalid HTML: %w", err))
				return
			}
			page.TemplateNode = parsed
			page.annotatePositions()

			txSavedNode := &html.Node{
				Type:     html.ElementNode,
				DataAtom: atom.Script,
				Data:     "script",
				Attr: []html.Attribute{
					{Key: "type", Val: "application/json"},
					{Key: "id", Val: "tx-saved"},
				},
			}

			var foundScript, foundHead bool
			for node := range page.TemplateNode.Descendants() {
				if !foundScript && isTmplxScriptNode(node) {
					page.TmplxScriptNode = node
					foundScript = true
				}
				if !foundHead && node.DataAtom == atom.Head {
					node.AppendChild(txSavedNode)
					node.AppendChild(&html.Node{
						Type:     html.ElementNode,
						DataAtom: atom.Script,
						Data:     "script",
						Attr: []html.Attribute{
							{Key: "id", Val: "tx-runtime"},
						},
					})
					foundHead = true
				}
				if foundScript && foundHead {
					break
				}
			}
			if !foundHead {
				errs = append(errs, page.errf("page must have a <head> element (required for state and runtime script injection)"))
				return
			}

			cleanUpTmplxScript(page.TemplateNode)

			errs = append(errs, page.parseScript())
			if err := page.checkImports(); err != nil {
				errs = append(errs, err)
				return
			}
			errs = append(errs, page.inferVarTypes())
			page.inferEffects()
		})
	}
	wg.Wait()
	if err := sortedJoin(errors.Join(append(compErrs, pageErrs...)...)); err != nil {
		return nil, err
	}

	// type-check (go/types) and collect child/fill metadata in one walk.
	// Sequential on purpose: a parent's probeTmpl reads its child components'
	// Vars (prop kinds/types) while probe also rewrites Vars, so parallel
	// probes race. (Parallelizing again needs a build/apply barrier inside
	// probe.)
	var probeErrs []error
	for _, comp := range slices.Concat(components, c.pages) {
		probeErrs = append(probeErrs, comp.probe())
	}
	if err := sortedJoin(errors.Join(probeErrs...)); err != nil {
		return nil, err
	}
	return components, nil
}

// Diagnose runs analyze and returns the diagnostics (parse + type errors)
// without generating code.
func (c *Compiler) Diagnose() []Diagnostic {
	_, err := c.analyze()
	return diagnostics(err)
}

// Compile type-checks the registered components and pages and returns the
// formatted routes.go bytes. It does no file I/O: the caller adds sources with
// NewComponent / NewPage and writes the result. On a gofmt failure it returns
// the unformatted bytes plus the error; on tmplx errors, nil plus a joined error.
func (c *Compiler) Compile() ([]byte, error) {
	components, err := c.analyze()
	if err != nil {
		return nil, err
	}

	// the body is emitted first, the package/import header after: the import
	// block depends on what the body used (code.needsFmt / needsHTML)
	var code CodeBuilder

	for _, comp := range components {
		code.write("type %s struct {\n", comp.GoName)
		code.write("tx_target string `json:\"-\"`\n")
		code.write("tx_prev url.Values `json:\"-\"`\n")
		code.write("tx_next map[string]any `json:\"-\"`\n")
		code.write("tx_trigger string `json:\"-\"`\n")
		code.write("tx_trigger_handler string `json:\"-\"`\n\n")
		for _, v := range comp.Vars {
			switch v.Kind {
			case VarKindState:
				code.write("V_%s %s `json:\"%s\"`\n", v.Name, v.Type, v.Name)
			case VarKindDerived:
				code.write("V_%s %s `json:\"-\"`\n", v.Name, v.Type)
			case VarKindProp:
				if v.IsFunc() {
					code.write("V_%s %s `json:\"-\"`\n", v.Name, v.Type)
				} else {
					code.write("V_%s *%s `json:\"-\"`\n", v.Name, v.Type)
				}
			}
		}
		code.write("}\n\n")

		code.write("func tx_new_%s(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string", comp.GoName)
		for _, v := range comp.Vars {
			if v.Kind == VarKindProp {
				if v.IsFunc() {
					code.write(", %s %s", v.Name, v.Type)
				} else {
					code.write(", %s *%s", v.Name, v.Type)
				}
			}
		}
		code.write(") *%s {\n", comp.GoName)
		code.write("tx_comp := &%s{}\n", comp.GoName)
		code.write("tx_comp.tx_target = tx_target\n")
		code.write("tx_comp.tx_prev = tx_prev\n")
		code.write("tx_comp.tx_next = tx_next\n")
		code.write("tx_comp.tx_trigger = tx_trigger\n")
		code.write("tx_comp.tx_trigger_handler = tx_trigger_handler\n")
		for _, v := range comp.Vars {
			if v.Kind == VarKindProp {
				if v.InitExpr != nil {
					code.write("if %s != nil {\n", v.Name)
					code.write("tx_comp.V_%s = %s\n", v.Name, v.Name)
					code.write("} else {\n")
					if v.IsFunc() {
						code.write("tx_comp.V_%s = %s\n", v.Name, v.InitCode)
					} else {
						code.write("val_%s := %s\n", v.Name, v.InitCode)
						code.write("tx_comp.V_%s = &val_%s\n", v.Name, v.Name)
					}
					code.write("}\n")
				} else {
					code.write("tx_comp.V_%s = %s\n", v.Name, v.Name)
				}
			}
		}
		code.write("tx_prev_str := tx_prev.Get(tx_id)\n")
		code.write("if tx_prev_str != \"\" {\n")
		code.write("json.Unmarshal([]byte(tx_prev_str), tx_comp)\n")
		for _, v := range comp.Vars {
			if v.Kind == VarKindDerived {
				code.write("tx_comp.V_%s = %s\n", v.Name, v.InitCode)
			}
		}
		if slices.ContainsFunc(comp.Vars, func(v *Var) bool {
			return v.Kind == VarKindDerived || (v.Kind == VarKindState && v.InitCode != "")
		}) || (comp.InitFunc != nil && comp.InitFunc.Code != "") {
			code.write("} else {\n")
			for _, v := range comp.Vars {
				switch v.Kind {
				case VarKindState:
					if v.InitCode != "" {
						code.write("tx_comp.V_%s = %s\n", v.Name, v.InitCode)
					}
				case VarKindDerived:
					code.write("tx_comp.V_%s = %s\n", v.Name, v.InitCode)
				}
			}
			if comp.InitFunc != nil {
				code.write("%s", comp.InitFunc.Code)
			}
		}
		code.write("}\n")
		code.write("return tx_comp\n")
		code.write("}\n\n")

		for _, f := range comp.Funcs {
			code.write("func (tx_comp *%s) %s%s {\n%s}\n\n", comp.GoName, f.Decl.Name.Name, strings.TrimPrefix(astToSource(f.Decl.Type), "func"), f.Code)
		}

		for _, eh := range comp.sortedEventHandlers() {
			code.write("func (tx_comp *%s) tx_eh%d(", comp.GoName, eh.ID)
			for i, p := range eh.Args {
				if i > 0 {
					code.write(", ")
				}
				code.write("%s %s", p.Name, p.Type)
			}
			code.write(") {\n%s\n}\n\n", eh.Code)
		}

		if comp.hasCompute() {
			code.write("func (tx_comp *%s) tx_compute(tx_id string) {\n", comp.GoName)
			if ehs := comp.sortedEventHandlers(); len(ehs) > 0 {
				code.write("if tx_id == tx_comp.tx_trigger {\n")
				code.write("switch tx_comp.tx_trigger_handler {\n")
				for _, eh := range ehs {
					code.write("case \"eh%d\":\n", eh.ID)
					for _, p := range eh.Args {
						code.write("var %s %s\n", p.Name, p.Type)
						code.write("json.Unmarshal([]byte(tx_comp.tx_prev.Get(\"%s\")), &%s)\n", p.Name, p.Name)
					}
					code.write("tx_comp.tx_eh%d(", eh.ID)
					for i, p := range eh.Args {
						if i > 0 {
							code.write(", ")
						}
						code.write("%s", p.Name)
					}
					code.write(")\n")
				}
				code.write("}\n")
				code.write("}\n")
			}
			cc := newCode("")
			comp.emitCompute(&cc, comp.TemplateNode, []LocalVar{}, []string{}, false)
			cc.writeTo(&code)
			code.write("}\n\n")
		}

		code.write("func (tx_comp *%s) tx_render(tx_w *bytes.Buffer, tx_id string", comp.GoName)
		for _, slotName := range comp.Slots {
			code.write(", tx_render_fill_%s func()", goIdent(slotName))
		}
		code.write(") {\n")
		w := newCode("tx_w")
		w.emitGo("if tx_comp.tx_target == tx_id {\n")
		w.emitStrLit("<!--tx:")
		w.emitExpr("tx_id")
		w.emitStrLit("-->")
		w.emitGo("}\n")
		comp.emitRender(&w, comp.TemplateNode, []LocalVar{}, []string{}, false)
		w.emitGo("if tx_comp.tx_target == tx_id {\n")
		w.emitStrLit("<!--tx:")
		w.emitExpr("tx_id + \"_e\"")
		w.emitStrLit("-->")
		w.emitGo("}\n")
		w.writeTo(&code)
		code.write("}\n\n")

		children := slices.SortedFunc(maps.Values(comp.Children), func(a, b *Child) int {
			return strings.Compare(a.pos(), b.pos())
		})

		for _, child := range children {
			for _, slotName := range child.Comp.Slots {
				if fill := child.Fills[slotName]; fill != nil {
					code.write("func (tx_comp *%s) tx_render_fill_%s_%s_%s(tx_w *bytes.Buffer", comp.GoName, child.Comp.GoName, child.Pos, goIdent(slotName))
					if len(fill.Children) > 0 {
						code.write(", tx_id string")
					}
					code.write(") {\n")
					fill.Code.writeTo(&code)
					code.write("}\n\n")
				}
			}
		}
	}

	for _, page := range c.pages {
		code.write("type %s struct {\n", page.GoName)
		code.write("tx_prev url.Values `json:\"-\"`\n")
		code.write("tx_next map[string]any `json:\"-\"`\n")
		code.write("tx_trigger string `json:\"-\"`\n")
		code.write("tx_trigger_handler string `json:\"-\"`\n\n")
		for _, v := range page.Vars {
			switch v.Kind {
			case VarKindState:
				code.write("V_%s %s `json:\"%s\"`\n", v.Name, v.Type, v.Name)
			case VarKindDerived, VarKindPath:
				code.write("V_%s %s `json:\"-\"`\n", v.Name, v.Type)
			}
		}
		code.write("}\n\n")

		code.write("func tx_new_%s(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string", page.GoName)
		for _, v := range page.Vars {
			if v.Kind == VarKindPath {
				code.write(", %s %s", v.Name, v.Type)
			}
		}
		code.write(") *%s {\n", page.GoName)
		code.write("tx_comp := &%s{}\n", page.GoName)
		code.write("tx_comp.tx_prev = tx_prev\n")
		code.write("tx_comp.tx_next = tx_next\n")
		code.write("tx_comp.tx_trigger = tx_trigger\n")
		code.write("tx_comp.tx_trigger_handler = tx_trigger_handler\n")
		for _, v := range page.Vars {
			if v.Kind == VarKindPath {
				code.write("tx_comp.V_%s = %s\n", v.Name, v.Name)
			}
		}
		code.write("tx_prev_str := tx_prev.Get(\"page\")\n")
		code.write("if tx_prev_str != \"\" {\n")
		code.write("json.Unmarshal([]byte(tx_prev_str), tx_comp)\n")
		for _, v := range page.Vars {
			if v.Kind == VarKindDerived {
				code.write("tx_comp.V_%s = %s\n", v.Name, v.InitCode)
			}
		}
		if slices.ContainsFunc(page.Vars, func(v *Var) bool {
			return v.Kind == VarKindDerived || (v.Kind == VarKindState && v.InitCode != "")
		}) || (page.InitFunc != nil && page.InitFunc.Code != "") {
			code.write("} else {\n")
			for _, v := range page.Vars {
				switch v.Kind {
				case VarKindState:
					if v.InitCode != "" {
						code.write("tx_comp.V_%s = %s\n", v.Name, v.InitCode)
					}
				case VarKindDerived:
					code.write("tx_comp.V_%s = %s\n", v.Name, v.InitCode)
				}
			}
			if page.InitFunc != nil {
				code.write("%s", page.InitFunc.Code)
			}
		}
		code.write("}\n")
		code.write("return tx_comp\n")
		code.write("}\n\n")

		for _, f := range page.Funcs {
			code.write("func (tx_comp *%s) %s%s {\n%s}\n\n", page.GoName, f.Decl.Name.Name, strings.TrimPrefix(astToSource(f.Decl.Type), "func"), f.Code)
		}
		for _, eh := range page.sortedEventHandlers() {
			code.write("func (tx_comp *%s) tx_eh%d(", page.GoName, eh.ID)
			for i, p := range eh.Args {
				if i > 0 {
					code.write(", ")
				}
				code.write("%s %s", p.Name, p.Type)
			}
			code.write(") {\n%s\n}\n\n", eh.Code)
		}

		if page.hasCompute() {
			code.write("func (tx_comp *%s) tx_compute() {\n", page.GoName)
			if ehs := page.sortedEventHandlers(); len(ehs) > 0 {
				code.write("if tx_comp.tx_trigger == \"page\" {\n")
				code.write("switch tx_comp.tx_trigger_handler {\n")
				for _, eh := range ehs {
					code.write("case \"eh%d\":\n", eh.ID)
					for _, p := range eh.Args {
						code.write("var %s %s\n", p.Name, p.Type)
						code.write("json.Unmarshal([]byte(tx_comp.tx_prev.Get(\"%s\")), &%s)\n", p.Name, p.Name)
					}
					code.write("tx_comp.tx_eh%d(", eh.ID)
					for i, p := range eh.Args {
						if i > 0 {
							code.write(", ")
						}
						code.write("%s", p.Name)
					}
					code.write(")\n")
				}
				code.write("}\n")
				code.write("}\n")
			}
			cc := newCode("")
			page.emitCompute(&cc, page.TemplateNode, []LocalVar{}, []string{}, false)
			cc.writeTo(&code)
			code.write("}\n\n")
		}

		code.write("func (tx_comp *%s) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {\n", page.GoName)
		w := newCode("tx_w1")
		page.emitRender(&w, page.TemplateNode, []LocalVar{}, []string{}, false)
		w.writeTo(&code)
		code.write("}\n\n")

		children := slices.SortedFunc(maps.Values(page.Children), func(a, b *Child) int {
			return strings.Compare(a.pos(), b.pos())
		})

		for _, child := range children {
			for _, slotName := range child.Comp.Slots {
				if fill := child.Fills[slotName]; fill != nil {
					code.write("func (tx_comp *%s) tx_render_fill_%s_%s_%s(tx_w *bytes.Buffer", page.GoName, child.Comp.GoName, child.Pos, goIdent(slotName))
					if len(fill.Children) > 0 {
						code.write(", tx_id string")
					}
					code.write(") {\n")
					fill.Code.writeTo(&code)
					code.write("}\n\n")
				}
			}
		}
	}

	code.write("func tx_dispatch(tx_w http.ResponseWriter, tx_r *http.Request) {\n")
	code.write("tx_r.ParseForm()\n")
	code.write("tx_prev := tx_r.PostForm\n")
	code.write("tx_target := tx_r.PostFormValue(\"target\")\n")
	code.write("tx_trigger := tx_r.PostFormValue(\"trigger\")\n")
	code.write("tx_trigger_handler := tx_r.URL.Path\n")
	code.write("if i := strings.LastIndexByte(tx_trigger_handler, '/'); i >= 0 {\ntx_trigger_handler = tx_trigger_handler[i+1:]\n}\n")
	code.write("tx_next := map[string]any{}\n")
	code.write("switch tx_target {\n")
	for _, page := range c.pages {
		code.write("case \"%s\":\n", page.Name)
		code.write("var tx_buf1, tx_buf2 bytes.Buffer\n")
		code.write("tx_comp := tx_new_%s(tx_prev, tx_next, tx_trigger, tx_trigger_handler", page.GoName)
		for _, v := range page.Vars {
			if v.Kind == VarKindPath {
				code.write(", \"\"")
			}
		}
		code.write(")\n")
		code.write("tx_next[\"page\"] = tx_comp\n")
		if page.hasCompute() {
			code.write("tx_comp.tx_compute()\n")
		}
		code.write("tx_comp.tx_render(&tx_buf1, &tx_buf2)\n")
		code.write("tx_json, _ := json.Marshal(tx_next)\n")
		code.write("tx_w.Write(tx_buf1.Bytes())\n")
		code.write("tx_w.Write(tx_json)\n")
		code.write("tx_w.Write(tx_buf2.Bytes())\n")
		code.write("return\n")
	}
	code.write("}\n")
	code.write("seg := tx_target\n")
	code.write("if i := max(strings.LastIndexByte(seg, ':'), strings.LastIndexByte(seg, '@')); i >= 0 {\nseg = seg[i+1:]\n}\n")
	code.write("name := seg\n")
	code.write("if i := strings.LastIndexByte(name, '-'); i >= 0 {\nname = name[:i]\n}\n")
	code.write("var buf bytes.Buffer\n")
	code.write("switch name {\n")
	for _, comp := range components {
		if comp.sealable() {
			code.write("case \"%s\":\n", comp.Name)
			code.write("tx_comp := tx_new_%s(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target", comp.GoName)
			for _, v := range comp.Vars {
				if v.Kind == VarKindProp {
					code.write(", nil")
				}
			}
			code.write(")\n")
			code.write("tx_next[tx_target] = tx_comp\n")
			if comp.hasCompute() {
				code.write("tx_comp.tx_compute(tx_target)\n")
			}
			code.write("tx_comp.tx_render(&buf, tx_target")
			for range comp.Slots {
				code.write(", nil")
			}
			code.write(")\n")
		}
	}
	code.write("default:\nreturn\n")
	code.write("}\n")
	code.write("tx_json, _ := json.Marshal(tx_next)\n")
	code.write("buf.WriteString(\"<script type=\\\"application/json\\\" id=\\\"tx-saved\\\">\")\n")
	code.write("buf.Write(tx_json)\n")
	code.write("buf.WriteString(\"</script>\")\n")
	code.write("tx_w.Write(buf.Bytes())\n")
	code.write("}\n\n")

	code.write("type TxRoute struct {\n")
	code.write("Pattern\tstring\n")
	code.write("Handler\thttp.HandlerFunc\n")
	code.write("}\n")

	code.write("var tx_routes []TxRoute = []TxRoute{\n")
	for _, page := range c.pages {
		code.write("{\n")
		code.write("Pattern: \"GET %s\",\n", page.Name)
		code.write("Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {\n")
		code.write("tx_next := map[string]any{}\n")
		code.write("tx_comp := tx_new_%s(nil, tx_next, \"\", \"\"", page.GoName)
		for _, v := range page.Vars {
			if v.Kind == VarKindPath {
				code.write(", %s", v.InitCode)
			}
		}
		code.write(")\n")
		code.write("tx_next[\"page\"] = tx_comp\n")
		// initial render: no event yet, so compute only runs to mount children
		if len(page.Children) > 0 {
			code.write("tx_comp.tx_compute()\n")
		}
		code.write("var tx_buf1, tx_buf2 bytes.Buffer\n")
		code.write("tx_comp.tx_render(&tx_buf1, &tx_buf2)\n")
		code.write("tx_json, _ := json.Marshal(tx_next)\n")
		code.write("tx_w.Write(tx_buf1.Bytes())\n")
		code.write("tx_w.Write(tx_json)\n")
		code.write("tx_w.Write(tx_buf2.Bytes())\n")
		code.write("},\n")
		code.write("},\n")
		for _, eh := range page.sortedEventHandlers() {
			code.write("{\n")
			code.write("Pattern: \"POST %s%s/eh%d\",\n", c.HandlerPrefix, url.PathEscape(page.Name), eh.ID)
			code.write("Handler: tx_dispatch,\n")
			code.write("},\n")
		}
	}
	for _, comp := range components {
		compUrl := url.PathEscape(comp.Name)
		for _, eh := range comp.sortedEventHandlers() {
			code.write("{\n")
			code.write("Pattern: \"POST %s%s/eh%d\",\n", c.HandlerPrefix, compUrl, eh.ID)
			code.write("Handler: tx_dispatch,\n")
			code.write("},\n")
		}
	}

	code.write("}\n")

	code.write("func Routes() []TxRoute { return tx_routes }\n\n")

	code.write("var tx_runtime_script = `%s`\n", strings.Replace(runtimeScript, "TX_HANDLER_PREFIX", c.HandlerPrefix, 1))

	var file CodeBuilder
	file.write("package %s\n", c.PackageName)
	file.write("import (\n")
	seen := map[string]bool{}  // import specs already written (dedup across files)
	bound := map[string]bool{} // paths whose package name the user's imports already bind
	for _, comp := range slices.Concat(c.pages, components) {
		for _, im := range comp.Imports {
			src := astToSource(im)
			if seen[src] {
				continue
			}
			seen[src] = true
			path, _ := strconv.Unquote(im.Path.Value)
			if im.Name == nil || im.Name.Name == path[strings.LastIndexByte(path, '/')+1:] {
				bound[path] = true
			}
			file.write("%s\n", src)
		}
	}
	generated := []string{"bytes", "encoding/json", "net/http", "strings"}
	if len(c.pages)+len(components) > 0 {
		generated = append(generated, "net/url")
	}
	if code.needsFmt {
		generated = append(generated, "fmt")
	}
	if code.needsHTML {
		generated = append(generated, "html")
	}
	for _, path := range generated {
		if !bound[path] {
			file.write("\"%s\"\n", path)
		}
	}
	file.write(")\n")
	file.WriteString(code.String())

	data := []byte(file.String())
	formatted, err := format.Source(data)
	if err != nil {
		return data, fmt.Errorf("format generated code: %w", err)
	}
	return formatted, nil
}

type CompType int

const (
	CompTypeComp CompType = iota
	CompTypePage
)

type Component struct {
	// set at construction
	compiler *Compiler
	content  []byte
	lineMap  *LineMap // byte offset -> line/col in content (for diagnostics)
	Type     CompType
	Path     string
	Name     string
	GoName   string

	// the parsed HTML, set in analyze
	TmplxScriptNode *html.Node
	TemplateNode    *html.Node
	StyleNode       *html.Node
	nodePos         map[*html.Node]int // node -> byte offset in content (annotatePositions)

	// the script block's declarations, set by parseScript and parseSlots
	scriptStart int // byte offset of the <script> text in content
	Imports     []*ast.ImportSpec
	Vars        []*Var
	InitFunc    *Func
	Funcs       []*Func
	Slots       []string

	// template metadata, filled by probe
	ChildCounters       map[string]*Counter
	Children            map[*html.Node]*Child
	EventHandlers       map[*html.Node][]*EventHandler
	EventHandlerCounter *Counter
}

// checkImports verifies every import resolves through comp.compiler.Importer.
// An import that does not resolve is reported here as a plain "cannot import X"
// so it fails clearly instead of cascading into a cryptic probe error; the
// caller skips type-checking the component when this returns non-nil.
func (comp *Component) checkImports() error {
	var errs []error
	for _, im := range comp.Imports {
		path, _ := strconv.Unquote(im.Path.Value)
		if _, err := comp.compiler.Importer.Import(path); err != nil {
			errs = append(errs, comp.errf("cannot import %s: %w", path, err))
		}
	}
	return errors.Join(errs...)
}

const scriptWrapPrefix = "package p\n"

func (comp *Component) parseScript() error {
	var errs []error

	if comp.TmplxScriptNode == nil || comp.TmplxScriptNode.FirstChild == nil {
		return errors.Join(errs...)
	}

	scriptData := comp.TmplxScriptNode.FirstChild.Data
	comp.scriptStart = bytes.Index(comp.content, []byte(scriptData))
	fset := token.NewFileSet()
	scriptAst, err := parser.ParseFile(fset, "", scriptWrapPrefix+scriptData, parser.ParseComments)
	if err != nil {
		span := Span{}
		msg := err.Error()
		if list, ok := err.(scanner.ErrorList); ok && len(list) > 0 {
			msg = list[0].Msg
			if comp.lineMap != nil {
				off := comp.scriptStart - len(scriptWrapPrefix) + list[0].Pos.Offset
				span = Span{Start: comp.lineMap.Pos(off), End: comp.lineMap.Pos(off)}
			}
		}
		errs = append(errs, comp.errAt(span, "syntax error in <script type=\"text/tmplx\">: %s", msg))
		return errors.Join(errs...)
	}

	for _, decl := range scriptAst.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok == token.IMPORT {
				for _, spec := range d.Specs {
					comp.Imports = append(comp.Imports, spec.(*ast.ImportSpec))
				}
				continue
			}
			if d.Tok == token.TYPE {
				errs = append(errs, comp.errAt(comp.spanOf(fset, d.Pos(), d.End()), "type declarations are not supported in the script block (declare types in a separate Go package and import them): %s", astToSource(d)))
				continue
			}
			if d.Tok != token.VAR || len(d.Specs) == 0 {
				continue
			}
			if len(d.Specs) > 1 {
				errs = append(errs, comp.errAt(comp.spanOf(fset, d.Pos(), d.End()), "more than one variable in var block: %s", astToSource(d)))
				continue
			}
			spec := d.Specs[0].(*ast.ValueSpec)
			if len(spec.Names) > 1 {
				errs = append(errs, comp.errAt(comp.spanOf(fset, d.Pos(), d.End()), "more than one name in var statement: %s", astToSource(spec)))
				continue
			}
			ident := spec.Names[0]
			if strings.HasPrefix(ident.Name, "tx_") {
				errs = append(errs, comp.errAt(comp.spanOf(fset, d.Pos(), d.End()), "%s: variable name cannot start with tx_ (reserved prefix)", ident.Name))
				continue
			}
			if len(spec.Values) > 1 {
				errs = append(errs, comp.errAt(comp.spanOf(fset, d.Pos(), d.End()), "more than one value in var statement: %s", astToSource(spec)))
				continue
			}

			typ := ""
			if spec.Type != nil {
				if _, ok := spec.Type.(*ast.StarExpr); ok {
					errs = append(errs, comp.errAt(comp.spanOf(fset, d.Pos(), d.End()), "%s: pointer types are not allowed at top level (use a value type)", ident.Name))
					continue
				}
				typ = astToSource(spec.Type)
			}

			foundProp := false
			foundPath := false
			pathValue := ""
			if d.Doc != nil {
				for _, comment := range d.Doc.List {
					for _, c := range parseComments(comment.Text) {
						switch c.Name {
						case CommentProp:
							foundProp = true
						case CommentPath:
							foundPath = true
							pathValue = c.Value
						}
					}
				}
			}
			if foundProp && foundPath {
				errs = append(errs, comp.errAt(comp.spanOf(fset, d.Pos(), d.End()), "cannot combine //tx:prop and //tx:path on %s", ident.Name))
				continue
			}

			kind := VarKindState
			switch {
			case foundProp:
				kind = VarKindProp
			case foundPath:
				kind = VarKindPath
			}

			switch kind {
			case VarKindProp:
				if comp.Type == CompTypePage {
					errs = append(errs, comp.errAt(comp.spanOf(fset, d.Pos(), d.End()), "//tx:prop on %s: pages cannot have props", ident.Name))
					continue
				}
				if ident.Name == "slot" {
					errs = append(errs, comp.errAt(comp.spanOf(fset, d.Pos(), d.End()), "//tx:prop on slot: \"slot\" is reserved (used for slot placement on the comp element); rename the prop"))
					continue
				}
			case VarKindPath:
				if comp.Type != CompTypePage {
					errs = append(errs, comp.errAt(comp.spanOf(fset, d.Pos(), d.End()), "//tx:path on %s: components cannot have path variables (only pages bind URL path values)", ident.Name))
					continue
				}
				if len(spec.Values) > 0 {
					errs = append(errs, comp.errAt(comp.spanOf(fset, d.Pos(), d.End()), "//tx:path variable cannot have an initial value: %s", astToSource(spec)))
					continue
				}
				if typ != "string" {
					errs = append(errs, comp.errAt(comp.spanOf(fset, d.Pos(), d.End()), "//tx:path variable must be type string: %s", astToSource(spec)))
					continue
				}
			case VarKindState:
				if _, ok := spec.Type.(*ast.FuncType); ok {
					errs = append(errs, comp.errAt(comp.spanOf(fset, d.Pos(), d.End()), "%s: state vars cannot be func type (use //tx:prop for callbacks)", ident.Name))
					continue
				}
			}

			v := &Var{
				Kind:     kind,
				Name:     ident.Name,
				Type:     typ,
				TypeExpr: spec.Type,
				Span:     comp.spanOf(fset, spec.Pos(), spec.End()),
			}
			if len(spec.Values) == 1 {
				v.InitExpr = spec.Values[0]
			}
			if kind == VarKindPath {
				v.InitCode = fmt.Sprintf("tx_r.PathValue(\"%s\")", pathValue)
			}
			comp.Vars = append(comp.Vars, v)

		case *ast.FuncDecl:
			if d.Recv != nil {
				errs = append(errs, comp.errAt(comp.spanOf(fset, d.Pos(), d.End()), "%s: methods (func with receiver) not allowed, use plain functions", d.Name))
				continue
			}
			for _, field := range d.Type.Params.List {
				for _, name := range field.Names {
					if strings.HasPrefix(name.Name, "tx_") {
						errs = append(errs, comp.errAt(comp.spanOf(fset, d.Pos(), d.End()), "%s: parameter %s cannot start with tx_ (reserved prefix)", d.Name, name.Name))
					}
				}
			}
			if strings.HasPrefix(d.Name.Name, "tx_") {
				errs = append(errs, comp.errAt(comp.spanOf(fset, d.Pos(), d.End()), "%s: function name cannot start with tx_ (reserved prefix)", d.Name.Name))
				continue
			}
			if d.Doc != nil {
				for _, comment := range d.Doc.List {
					for _, c := range parseComments(comment.Text) {
						errs = append(errs, comp.errAt(comp.spanOf(fset, d.Pos(), d.End()), "%s: //tx:%s is only valid on var declarations", d.Name.Name, c.Name))
					}
				}
			}
			if d.Body == nil {
				errs = append(errs, comp.errAt(comp.spanOf(fset, d.Pos(), d.End()), "%s: function declarations must have a body; use `var %s func(...) ...` for a function-typed prop", d.Name.Name, d.Name.Name))
				continue
			}

			f := &Func{Decl: d, Span: comp.spanOf(fset, d.Pos(), d.End())}
			if d.Name.Name == "init" {
				comp.InitFunc = f
			} else {
				comp.Funcs = append(comp.Funcs, f)
			}
		}
	}

	return errors.Join(errs...)
}

func (comp *Component) inferVarTypes() error {
	var errs []error

	var b strings.Builder
	b.WriteString("package tx_infer\n")
	for _, im := range comp.Imports {
		fmt.Fprintf(&b, "import %s\n", astToSource(im))
	}
	for _, v := range comp.Vars {
		switch {
		case v.Type != "" && v.InitExpr != nil:
			fmt.Fprintf(&b, "var %s %s = %s\n", v.Name, v.Type, astToSource(v.InitExpr))
		case v.Type != "":
			fmt.Fprintf(&b, "var %s %s\n", v.Name, v.Type)
		case v.InitExpr != nil:
			fmt.Fprintf(&b, "var %s = %s\n", v.Name, astToSource(v.InitExpr))
		}
	}
	for _, f := range comp.Funcs {
		fmt.Fprintf(&b, "%s\n", astToSource(f.Decl))
	}

	fset := token.NewFileSet()
	file, perr := parser.ParseFile(fset, "infer.go", b.String(), 0)
	if perr != nil {
		return errors.Join(errs...)
	}

	conf := types.Config{Importer: comp.compiler.Importer, Error: func(error) {}}
	info := &types.Info{Defs: map[*ast.Ident]types.Object{}}
	conf.Check("tx_infer", fset, []*ast.File{file}, info)

	typeByName := map[string]string{}
	ast.Inspect(file, func(n ast.Node) bool {
		if vs, ok := n.(*ast.ValueSpec); ok {
			for _, ident := range vs.Names {
				if obj := info.Defs[ident]; obj != nil && obj.Type() != nil {
					typeByName[ident.Name] = types.TypeString(obj.Type(), func(p *types.Package) string { return p.Name() })
				}
			}
		}
		return true
	})
	for _, v := range comp.Vars {
		if v.Type == "" {
			if t, ok := typeByName[v.Name]; ok {
				v.Type = t
			} else {
				errs = append(errs, comp.errf("%s: cannot resolve type", v.Name))
			}
		}
	}
	return errors.Join(errs...)
}

func (comp *Component) parseSlots(node *html.Node, inSlot bool) error {
	var errs []error

	isSlot := node.Type == html.ElementNode && node.DataAtom == atom.Slot
	if isSlot && inSlot {
		errs = append(errs, comp.errf("<slot> cannot be nested inside another <slot>"))
		return errors.Join(errs...)
	}

	if isSlot {
		slotName := ""
		if name, found := hasAttr(node, "name"); found {
			slotName = name
		}

		if slices.Contains(comp.Slots, slotName) {
			if slotName == "" {
				errs = append(errs, comp.errf("duplicate default <slot> (only one allowed)"))
			} else {
				errs = append(errs, comp.errf("duplicate <slot name=\"%s\"> (only one allowed)", slotName))
			}
		} else {
			comp.Slots = append(comp.Slots, slotName)
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		errs = append(errs, comp.parseSlots(c, inSlot || isSlot))
	}
	return errors.Join(errs...)
}

func (comp *Component) inferEffects() {
	selfOf := map[*Func]bool{}
	propsOf := map[*Func][]*Var{}
	callsOf := map[*Func][]*Func{}
	for _, f := range comp.Funcs {
		selfOf[f], propsOf[f], callsOf[f] = comp.directFuncFacts(f.Decl.Body)
	}
	for _, f := range comp.Funcs {
		eff := Effect{}
		seen := map[*Var]struct{}{}
		visited := map[*Func]struct{}{}
		var walk func(*Func)
		walk = func(g *Func) {
			if _, ok := visited[g]; ok {
				return
			}
			visited[g] = struct{}{}
			if selfOf[g] {
				eff.Self = true
			}
			for _, p := range propsOf[g] {
				if _, ok := seen[p]; !ok {
					seen[p] = struct{}{}
					eff.FuncProps = append(eff.FuncProps, p)
				}
			}
			for _, callee := range callsOf[g] {
				walk(callee)
			}
		}
		walk(f)
		f.Effect = eff
	}
}

func (comp *Component) probe() error {
	var errs []error

	p := newProbeState()
	p.write("package tx_probe\n")
	for _, im := range comp.Imports {
		p.writef("import %s\n", astToSource(im))
	}
	for _, v := range comp.Vars {
		p.anchor(v.Span, "var "+v.Name)
		if v.InitExpr != nil {
			p.writef("var %s %s = %s\n", v.Name, v.Type, astToSource(v.InitExpr))
		} else {
			p.writef("var %s %s\n", v.Name, v.Type)
		}
	}

	allFuncs := comp.Funcs
	if comp.InitFunc != nil {
		allFuncs = append(allFuncs, comp.InitFunc)
	}

	for _, f := range allFuncs {
		p.anchor(f.Span, "func "+f.Decl.Name.Name)
		p.writef("%s\n", astToSource(f.Decl))
	}

	p.write("func tx_tmpl() {\n")
	comp.probeTmpl(comp.TemplateNode, p)
	p.write("}\n")

	fset := token.NewFileSet()
	probeFile, perr := parser.ParseFile(fset, "probe.go", p.buf.String(), 0)
	if perr != nil {
		errs = append(errs, comp.errf("internal: probe parse failed: %v", perr))
		return errors.Join(errs...)
	}

	var typeErrs []types.Error
	conf := types.Config{
		Importer: comp.compiler.Importer,
		Error: func(e error) {
			if terr, ok := e.(types.Error); ok {
				typeErrs = append(typeErrs, terr)
			}
		},
	}
	info := &types.Info{
		Defs: map[*ast.Ident]types.Object{},
		Uses: map[*ast.Ident]types.Object{},
	}
	conf.Check("tx_probe", fset, []*ast.File{probeFile}, info)

	localTypes := map[string]string{}
	ast.Inspect(probeFile, func(n ast.Node) bool {
		var lhs []ast.Expr
		switch s := n.(type) {
		case *ast.RangeStmt:
			if s.Tok == token.DEFINE {
				lhs = []ast.Expr{s.Key, s.Value}
			}
		case *ast.ForStmt:
			if a, ok := s.Init.(*ast.AssignStmt); ok && a.Tok == token.DEFINE {
				lhs = a.Lhs
			}
		case *ast.IfStmt:
			if a, ok := s.Init.(*ast.AssignStmt); ok && a.Tok == token.DEFINE {
				lhs = a.Lhs
			}
		}
		for _, e := range lhs {
			if id, ok := e.(*ast.Ident); ok && id.Name != "_" {
				if obj := info.Defs[id]; obj != nil {
					localTypes[id.Name] = obj.Type().String()
				}
			}
		}
		return true
	})

	usedVars := map[string]struct{}{}
	for ident := range info.Uses {
		usedVars[ident.Name] = struct{}{}
	}
	for _, v := range comp.Vars {
		if _, ok := usedVars[v.Name]; !ok {
			errs = append(errs, comp.errf("%s declared but not used", v.Name))
		}
	}

	for _, v := range comp.Vars {
		switch v.Kind {
		case VarKindPath:
			// InitCode already set in parseScript (tx_r.PathValue); nothing to do here
			continue

		case VarKindProp:
			if v.InitExpr != nil {
				if !v.IsFunc() {
					astutil.Apply(v.InitExpr, func(c *astutil.Cursor) bool {
						id, ok := c.Node().(*ast.Ident)
						if !ok || !atVarRefPos(c) {
							return true
						}
						if vv := comp.varByName(id.Name); vv != nil && (vv.Kind == VarKindState || vv.Kind == VarKindDerived) {
							errs = append(errs, comp.errf("prop %s default cannot reference state variable %s: props are resolved before state exists", v.Name, id.Name))
						}
						return true
					}, nil)
				}
				v.InitCode = comp.rewriteExpr(astToSource(v.InitExpr), nil)
			}

		case VarKindState:
			if v.InitExpr != nil {
				// init referencing a comp var makes this a derived var (recomputed each
				// request, not serialized)
				isDerived := false
				astutil.Apply(v.InitExpr, func(c *astutil.Cursor) bool {
					id, ok := c.Node().(*ast.Ident)
					if !ok {
						return true
					}
					if !atVarRefPos(c) {
						return false
					}
					if comp.varByName(id.Name) != nil {
						isDerived = true
					}
					return false
				}, nil)
				if isDerived {
					v.Kind = VarKindDerived
				}

				v.InitCode = comp.rewriteExpr(astToSource(v.InitExpr), nil)
			}
		}
	}

	reordered := make([]*Var, 0, len(comp.Vars))
	for _, v := range comp.Vars {
		if v.InitExpr == nil {
			reordered = append(reordered, v)
		}
	}
	for _, init := range info.InitOrder {
		for _, lhs := range init.Lhs {
			if v := comp.varByName(lhs.Name()); v != nil {
				reordered = append(reordered, v)
			}
		}
	}
	comp.Vars = reordered

	for _, f := range allFuncs {
		d := f.Decl
		for _, field := range d.Type.Params.List {
			for _, name := range field.Names {
				if d.Body != nil && comp.varByName(name.Name) != nil {
					errs = append(errs, comp.errf("%s: parameter %s shadows a state variable", d.Name, name.Name))
				}
			}
		}
		if d.Body != nil {
			for _, name := range comp.readOnlyMutations(d.Body) {
				errs = append(errs, comp.errf("%s: cannot assign to %s — only state vars are writable", d.Name, name))
			}
			for _, name := range comp.shadowingLocals(d.Body) {
				errs = append(errs, comp.errf("%s: local variable %s shadows a state/prop/derived/path variable (rename the local)", d.Name, name))
			}
			dirty := comp.dirtyDerivedNames(d.Body)
			var b strings.Builder
			for _, stmt := range d.Body.List {
				b.WriteString(astToSource(comp.rewriteVarRefs(stmt, nil)))
				b.WriteByte('\n')
			}
			for _, name := range dirty {
				v := comp.varByName(name)
				fmt.Fprintf(&b, "%s.V_%s = %s\n", "tx_comp", v.Name, v.InitCode)
			}
			f.Code = strings.TrimSpace(b.String())
		}
	}

	for node, ehs := range comp.EventHandlers {
		for _, eh := range ehs {
			fileAst, ferr := parser.ParseFile(token.NewFileSet(), "", "package p\nfunc f() {\n"+eh.Val+"\n}", 0)
			if ferr != nil {
				continue
			}
			decl, ok := fileAst.Decls[0].(*ast.FuncDecl)
			if !ok || decl.Body == nil {
				continue
			}
			for _, lv := range comp.capturedLocals(decl.Body, node) {
				arg := EventHandlerArg{Name: lv.Name, Type: localTypes[lv.Name]}
				eh.Args = append(eh.Args, arg)
				eh.Captures = append(eh.Captures, arg)
			}
			if eventName := strings.TrimPrefix(eh.Key, "tx-on"); eventCarriesValue(eventName) {
				astutil.Apply(decl.Body, func(c *astutil.Cursor) bool {
					sel, ok := c.Node().(*ast.SelectorExpr)
					if !ok || sel.Sel.Name != "value" {
						return true
					}
					inner, ok := sel.X.(*ast.SelectorExpr)
					if !ok || inner.Sel.Name != "target" {
						return true
					}
					if base, ok := inner.X.(*ast.Ident); ok && base.Name == "event" {
						c.Replace(&ast.Ident{Name: "tx_ev_target_value"})
						return false
					}
					return true
				}, nil)
				eh.Args = append(eh.Args, EventHandlerArg{Name: "tx_ev_target_value", Type: "string"})
			}
			dirty := comp.dirtyDerivedNames(decl.Body)
			var b strings.Builder
			for _, stmt := range decl.Body.List {
				b.WriteString(astToSource(comp.rewriteVarRefs(stmt, nil)))
				b.WriteByte('\n')
			}
			for _, name := range dirty {
				v := comp.varByName(name)
				fmt.Fprintf(&b, "%s.V_%s = %s\n", "tx_comp", v.Name, v.InitCode)
			}
			eh.Code = strings.TrimSpace(b.String())
		}
	}

	for _, d := range p.parseErrs {
		d.Path = comp.Path
		errs = append(errs, d)
	}
	for _, terr := range typeErrs {
		pos := fset.Position(terr.Pos)
		span, desc := p.descAt(pos.Line)
		if desc != "" {
			errs = append(errs, comp.errAt(span, "%s: %s", desc, terr.Msg))
		} else {
			errs = append(errs, comp.errAt(span, "%s", terr.Msg))
		}
	}
	return errors.Join(errs...)
}

func (comp *Component) probeTmpl(node *html.Node, p *probeState) {
	switch node.Type {
	case html.TextNode:
		_, hasTxIgnore := hasAttr(node.Parent, "tx-ignore")
		if hasTxIgnore || isVerbatimSerialize(node.Parent.DataAtom) {
			return
		}

		elemName := node.Parent.Data
		nodeBase, hasBase := comp.nodePos[node] // byte offset of this text node in comp.content
		comp.scanTmplStr(node.Data, false, func(rune) {}, func(expr string, off int) error {
			desc := fmt.Sprintf("text in <%s>: { %s }", elemName, expr)
			p.span = Span{}
			if hasBase && comp.lineMap != nil {
				p.span = Span{Start: comp.lineMap.Pos(nodeBase + off), End: comp.lineMap.Pos(nodeBase + off + len(expr))}
			}
			if _, err := parser.ParseExpr(expr); err != nil {
				p.parseErr(desc, err)
				return nil
			}
			p.anchor(Span{}, desc)
			p.writef("_ = %s\n", expr)
			return nil
		})
		return

	case html.ElementNode:
		p.span = comp.nodeSpan(node)
		if usedComp, ok := comp.compiler.components[node.Data]; ok {
			if _, ok := comp.ChildCounters[usedComp.Name]; !ok {
				comp.ChildCounters[usedComp.Name] = &Counter{}
			}

			args := []Arg{}
			for _, attr := range node.Attr {
				if attr.Key == "tx-if" || attr.Key == "tx-else-if" || attr.Key == "tx-else" || attr.Key == "tx-for" || attr.Key == "tx-key" || attr.Key == "slot" {
					continue
				}
				v := usedComp.varByName(attr.Key)
				if v == nil || v.Kind != VarKindProp {
					p.perr("<%s %s=...>: %s is not a prop on <%s>", node.Data, attr.Key, attr.Key, usedComp.Name)
					continue
				}

				desc := fmt.Sprintf("<%s %s=\"%s\">", node.Data, attr.Key, attr.Val)
				expr, err := parser.ParseExpr(attr.Val)
				if err != nil {
					p.parseErr(desc, err)
					continue
				}

				p.anchor(Span{}, desc)
				p.writef("{ var _ %s = %s }\n", v.Type, attr.Val)

				arg := Arg{PropName: v.Name, Kind: ArgKindValue, Val: attr.Val}
				if ident, ok := expr.(*ast.Ident); ok {
					if f := comp.funcByName(ident.Name); v.IsFunc() && f != nil {
						arg = Arg{PropName: v.Name, Kind: ArgKindFunc, Func: f}
					} else if vr := comp.varByName(ident.Name); vr != nil {
						arg = Arg{PropName: v.Name, Kind: ArgKindVar, Var: vr}
					}
				}

				args = append(args, arg)
			}

			child := &Child{
				Pos:   fmt.Sprint(comp.ChildCounters[usedComp.Name].next()),
				Comp:  usedComp,
				Args:  args,
				Fills: map[string]*Fill{},
			}

			comp.Children[node] = child

			fillNodes := map[string]*html.Node{}
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				slotName, _ := hasAttr(c, "slot")
				if fillNodes[slotName] == nil {
					fillNodes[slotName] = newTemplateNode()
				}
				fillNodes[slotName].AppendChild(&html.Node{
					FirstChild: c.FirstChild,
					LastChild:  c.LastChild,
					Type:       c.Type,
					DataAtom:   c.DataAtom,
					Data:       c.Data,
					Namespace:  c.Namespace,
					Attr:       c.Attr,
				})
			}
			savedChildren := comp.Children
			for _, slotName := range usedComp.Slots {
				if n := fillNodes[slotName]; n != nil {
					comp.Children = map[*html.Node]*Child{}
					comp.probeTmpl(n, p)
					child.Fills[slotName] = &Fill{
						Node:     n,
						Children: comp.Children,
					}
					maps.Copy(savedChildren, comp.Children)
				}
			}
			comp.Children = savedChildren

			return

		} else if node.DataAtom != atom.Slot && node.DataAtom != atom.Template {
			if _, isIgnored := hasAttr(node, "tx-ignore"); !isIgnored {
				for _, attr := range node.Attr {
					if attr.Key == "tx-if" || attr.Key == "tx-else-if" || attr.Key == "tx-else" || attr.Key == "tx-for" || attr.Key == "tx-key" {
						continue
					}
					if strings.HasPrefix(attr.Key, "tx-on") {
						desc := fmt.Sprintf("%s on <%s>", attr.Key, node.Data)
						fileAst, err := parser.ParseFile(token.NewFileSet(), "", fmt.Sprintf("package p\nfunc f() {\n%s\n}", attr.Val), 0)
						if err != nil {
							p.parseErr(desc, err)
							continue
						}
						// reject a handler value that escapes the func wrapper (e.g. "}func g(){"):
						// it parses as two decls, so require exactly one func decl with a body
						if len(fileAst.Decls) != 1 {
							continue
						}
						decl, ok := fileAst.Decls[0].(*ast.FuncDecl)
						if !ok || decl.Body == nil {
							continue
						}
						for _, name := range comp.readOnlyMutations(decl.Body) {
							p.perr("%s: cannot assign to %s — only state vars are writable", desc, name)
						}
						eventName := strings.TrimPrefix(attr.Key, "tx-on")
						p.anchor(Span{}, desc)
						p.write("{\n")
						if eventCarriesValue(eventName) {
							p.write("var event struct{ target struct{ value string } }\n_ = event\n")
						}
						for _, stmt := range decl.Body.List {
							p.writef("%s\n", astToSource(stmt))
						}
						p.write("}\n")
						comp.EventHandlers[node] = append(comp.EventHandlers[node], &EventHandler{
							ID:  comp.EventHandlerCounter.next(),
							Key: attr.Key,
							Val: attr.Val,
						})
					} else if attr.Key == "tx-action" {
						desc := fmt.Sprintf("tx-action on <%s>", node.Data)
						f := comp.funcByName(attr.Val)
						if f == nil {
							p.perr("%s: %s is not a known function", desc, attr.Val)
							continue
						}
						p.anchor(Span{}, desc)
						p.writef("_ = %s\n", attr.Val)
						eh := &EventHandler{ID: comp.EventHandlerCounter.next(), Key: "tx-action"}
						var names []string
						for _, field := range f.Decl.Type.Params.List {
							typeStr := astToSource(field.Type)
							for _, n := range field.Names {
								eh.Args = append(eh.Args, EventHandlerArg{Name: n.Name, Type: typeStr})
								names = append(names, n.Name)
							}
						}
						eh.Val = fmt.Sprintf("%s(%s)", attr.Val, strings.Join(names, ", "))
						comp.EventHandlers[node] = append(comp.EventHandlers[node], eh)
					} else {
						elemName := node.Data
						attrKey := attr.Key
						comp.scanTmplStr(attr.Val, false, func(rune) {}, func(expr string, _ int) error {
							desc := fmt.Sprintf("<%s %s>: { %s }", elemName, attrKey, expr)
							if _, err := parser.ParseExpr(expr); err != nil {
								p.parseErr(desc, err)
								return nil
							}
							p.anchor(Span{}, desc)
							p.writef("_ = %s\n", expr)
							return nil
						})
					}
				}
			}
		}
	}

	var prevCondState CondState
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		hasFor := false
		branchFailed := false
		if c.Type == html.ElementNode {
			p.span = comp.nodeSpan(c)
			currCondState, field := condState(c)
			switch prevCondState {
			case CondStateIf, CondStateElseIf:
				if currCondState != CondStateElseIf && currCondState != CondStateElse {
					p.write("}\n")
				}
			case CondStateElse:
				p.write("}\n")
			}
			validChain := true
			switch currCondState {
			case CondStateIf:
				_, err := parser.ParseFile(token.NewFileSet(), "", "package p\nfunc _(){if "+field+"{}}", 0)
				if err == nil {
					p.anchor(Span{}, fmt.Sprintf("tx-if on <%s>", c.Data))
					p.writef("if %s {\n", field)
				} else {
					p.parseErr(fmt.Sprintf("tx-if on <%s>", c.Data), err)
					validChain = false
					branchFailed = true
				}
			case CondStateElseIf:
				chainOk := prevCondState == CondStateIf || prevCondState == CondStateElseIf
				_, err := parser.ParseFile(token.NewFileSet(), "", "package p\nfunc _(){if "+field+"{}}", 0)
				if chainOk && err == nil {
					p.anchor(Span{}, fmt.Sprintf("tx-else-if on <%s>", c.Data))
					p.writef("} else if %s {\n", field)
				} else {
					if err != nil {
						p.parseErr(fmt.Sprintf("tx-else-if on <%s>", c.Data), err)
						branchFailed = true
					}
					if !chainOk {
						p.perr("tx-else-if on <%s>: must follow tx-if or tx-else-if", c.Data)
					}
					validChain = false
				}
			case CondStateElse:
				if prevCondState == CondStateIf || prevCondState == CondStateElseIf {
					p.anchor(Span{}, fmt.Sprintf("tx-else on <%s>", c.Data))
					p.write("} else {\n")
				} else {
					p.perr("tx-else on <%s>: must follow tx-if or tx-else-if", c.Data)
					validChain = false
				}
			}
			if validChain {
				prevCondState = currCondState
			} else {
				prevCondState = CondStateDefault
			}

			for _, lv := range comp.ifLocals(field) {
				if comp.varByName(lv.Name) != nil {
					p.perr("tx-if on <%s>: %s shadows a state/prop/derived/path variable (rename it)", c.Data, lv.Name)
				}
			}
			if eff := comp.inlineEffect("func(){\nif " + field + " {}\n}"); eff.Self || len(eff.FuncProps) > 0 {
				p.perr("tx-if on <%s>: condition must not mutate state (move effects into a handler): %s", c.Data, field)
			}

			if stmt, ok := hasAttr(c, "tx-for"); ok {
				if _, ok := hasAttr(c, "tx-key"); !ok {
					p.perr("tx-for on <%s>: requires a tx-key attribute", c.Data)
				}
				for _, lv := range comp.forLocals(stmt) {
					if comp.varByName(lv.Name) != nil {
						p.perr("tx-for on <%s>: %s shadows a state/prop/derived/path variable (rename it)", c.Data, lv.Name)
					}
				}
				if eff := comp.inlineEffect("func(){\nfor " + stmt + " {}\n}"); eff.Self || len(eff.FuncProps) > 0 {
					p.perr("tx-for on <%s>: range must not mutate state (move effects into a handler): %s", c.Data, stmt)
				}
				_, ferr := parser.ParseFile(token.NewFileSet(), "", "package p\nfunc _(){for "+stmt+"{}}", 0)
				if ferr == nil {
					hasFor = true
					p.anchor(Span{}, fmt.Sprintf("tx-for on <%s>", c.Data))
					p.writef("for %s {\n", stmt)
				} else {
					p.parseErr(fmt.Sprintf("tx-for on <%s>", c.Data), ferr)
					branchFailed = true
				}
			}
			if key, ok := hasAttr(c, "tx-key"); ok {
				if _, err := parser.ParseExpr(key); err == nil {
					p.anchor(Span{}, fmt.Sprintf("tx-key on <%s>", c.Data))
					p.writef("_ = %s\n", key)
				} else {
					p.parseErr(fmt.Sprintf("tx-key on <%s>", c.Data), err)
				}
			}
		}

		if !branchFailed {
			comp.probeTmpl(c, p)
		}

		if hasFor {
			p.write("}\n")
		}

		if c.NextSibling == nil && (prevCondState == CondStateIf || prevCondState == CondStateElseIf || prevCondState == CondStateElse) {
			p.write("}\n")
		}
	}
}

func (comp *Component) emitCompute(code *Code, node *html.Node, localVars []LocalVar, forKeys []string, inSlot bool) {
	switch node.Type {
	case html.DocumentNode:
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			comp.emitCompute(code, child, localVars, forKeys, inSlot)
		}

	case html.ElementNode:
		if child, ok := comp.Children[node]; ok {
			code.emitGo("{\n")

			comp.emitChildKey(code, child, forKeys, inSlot)

			argStr := ""
			for _, v := range child.Comp.Vars {
				if v.Kind == VarKindProp {
					var arg *Arg
					for _, a := range child.Args {
						if a.PropName == v.Name {
							arg = &a
							break
						}
					}
					if arg == nil {
						argStr += ", nil"
						continue
					}
					if v.IsFunc() {
						switch arg.Kind {
						case ArgKindValue:
							code.emitGo(fmt.Sprintf("tx_val_%s := %s\n", arg.PropName, comp.rewriteExpr(arg.Val, localVars)))
							argStr += fmt.Sprintf(", tx_val_%s", arg.PropName)
						case ArgKindVar:
							argStr += fmt.Sprintf(", tx_comp.V_%s", arg.Var.Name)
						case ArgKindFunc:
							argStr += fmt.Sprintf(", tx_comp.%s", arg.Func.Decl.Name.Name)
						}
					} else {
						switch arg.Kind {
						case ArgKindValue:
							code.emitGo(fmt.Sprintf("tx_val_%s := %s\n", arg.PropName, comp.rewriteExpr(arg.Val, localVars)))
							argStr += fmt.Sprintf(", &tx_val_%s", arg.PropName)
						case ArgKindVar:
							argStr += fmt.Sprintf(", &tx_comp.V_%s", arg.Var.Name)
						}
					}
				}

			}
			txTarget := comp.rootExpr()
			if len(child.Args) == 0 && child.Comp.sealable() {
				txTarget = "tx_id"
			}
			code.emitGo(fmt.Sprintf("tx_child := tx_new_%s(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, %s%s)\n", child.Comp.GoName, txTarget, argStr))
			if child.Comp.hasCompute() {
				code.emitGo("tx_child.tx_compute(tx_id)\n")
			}
			code.emitGo("tx_comp.tx_next[tx_id] = tx_child\n")
			for _, slotName := range child.Comp.Slots {
				if fill := child.Fills[slotName]; fill != nil {
					comp.emitCompute(code, fill.Node, localVars, forKeys, true)
				}
			}
			code.emitGo("}\n")
			return
		}

		if node.DataAtom == atom.Slot {
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				comp.emitCompute(code, child, localVars, forKeys, false)
			}
			return
		}

		comp.emitChildren(code, node, localVars, forKeys, inSlot, comp.emitCompute)
	}
}

func (comp *Component) emitRender(code *Code, node *html.Node, localVars []LocalVar, forKeys []string, inSlot bool) {
	switch node.Type {
	case html.DocumentNode:
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			comp.emitRender(code, child, localVars, forKeys, inSlot)
		}

	case html.DoctypeNode:
		code.emitStrLit("<!DOCTYPE " + node.Data + ">")

	case html.CommentNode:
		code.emitStrLit("<!--" + node.Data + "-->")

	case html.TextNode:
		if isVerbatimSerialize(node.Parent.DataAtom) {
			code.emitStrLit(node.Data)
		} else if _, txIgnore := hasAttr(node.Parent, "tx-ignore"); txIgnore {
			code.emitStrLit(html.EscapeString(node.Data))
		} else {
			collapseWs := !isWhitespacePreserving(node.Parent.DataAtom)
			comp.scanTmplStr(node.Data, collapseWs, func(r rune) {
				code.emitStrLit(html.EscapeString(string(r)))
			}, func(expr string, _ int) error {
				code.emitHtmlEscapeExpr(comp.rewriteExpr(expr, localVars))
				return nil
			})
		}

	case html.ElementNode:
		if child, ok := comp.Children[node]; ok {
			code.emitGo("{\n")

			comp.emitChildKey(code, child, forKeys, inSlot)

			code.emitGo(fmt.Sprintf("tx_child := tx_comp.tx_next[tx_id].(*%s)\n", child.Comp.GoName))
			code.emitGo(fmt.Sprintf("tx_child.tx_render(%s, tx_id", code.PendingSegment.BufName))

			for _, slotName := range child.Comp.Slots {
				if fill := child.Fills[slotName]; fill != nil {
					code.emitGo(", func() {\n")

					code.emitGo(fmt.Sprintf("tx_comp.tx_render_fill_%s_%s_%s(%s", child.Comp.GoName, child.Pos, goIdent(slotName), code.PendingSegment.BufName))
					if len(fill.Children) > 0 {
						code.emitGo(", tx_id")
					}
					code.emitGo(")\n")

					c := newCode("tx_w")
					comp.emitRender(&c, fill.Node, localVars, forKeys, true)
					fill.Code = &c

					code.emitGo("}")
				} else {
					code.emitGo(", nil")
				}
			}

			code.emitGo(")\n")
			code.emitGo("}\n")
			return
		}

		if node.DataAtom == atom.Slot {
			name, _ := hasAttr(node, "name")
			goName := goIdent(name)
			code.emitGo(fmt.Sprintf("if tx_render_fill_%s != nil {\n", goName))
			code.emitGo(fmt.Sprintf("tx_render_fill_%s()\n", goName))
			if node.FirstChild != nil {
				code.emitGo("} else {\n")
				for child := node.FirstChild; child != nil; child = child.NextSibling {
					comp.emitRender(code, child, localVars, forKeys, false)
				}
			}
			code.emitGo("}\n")
			return
		}

		if node.DataAtom != atom.Template {
			code.emitStrLit("<" + node.Data)
			for _, attr := range node.Attr {
				if attr.Key == "tx-if" || attr.Key == "tx-else-if" || attr.Key == "tx-else" || attr.Key == "tx-for" || attr.Key == "tx-key" || attr.Key == "slot" || attr.Key == "tx-action" || attr.Key == "tx-ignore" || strings.HasPrefix(attr.Key, "tx-on") {
					continue
				}
				code.emitStrLit(fmt.Sprintf(" %s=\"", attr.Key))
				comp.scanTmplStr(attr.Val, false, func(r rune) {
					code.emitStrLit(string(r))
				}, func(expr string, _ int) error {
					code.emitHtmlEscapeExpr(comp.rewriteExpr(expr, localVars))
					return nil
				})
				code.emitStrLit("\"")
			}
			if comp.Type == CompTypePage && node.DataAtom == atom.Script {
				if id, _ := hasAttr(node, "id"); id == "tx-saved" {
					code.emitStrLit(fmt.Sprintf(" data-tx-page=\"%s\"", comp.Name))
				}
			}
			if ehs := comp.EventHandlers[node]; len(ehs) > 0 {
				if comp.Type == CompTypePage {
					code.emitStrLit(" data-tx-trigger=\"page\"")
				} else {
					code.emitStrLit(" data-tx-trigger=\"")
					code.emitExpr("tx_id")
					code.emitStrLit("\"")
				}
				code.emitStrLit(" data-tx-target=\"")
				code.emitExpr(comp.rootExpr())
				code.emitStrLit("\"")
				for _, eh := range ehs {
					if eh.Key == "tx-action" {
						code.emitStrLit(fmt.Sprintf(" data-tx-action=\"%s/eh%d\"", url.PathEscape(comp.Name), eh.ID))
					} else {
						code.emitStrLit(fmt.Sprintf(" data-tx-eh%d-on=\"%s\"", eh.ID, strings.TrimPrefix(eh.Key, "tx-on")))
					}
					for _, p := range eh.Captures {
						code.emitStrLit(fmt.Sprintf(" data-tx-eh%d-arg-%s=\"", eh.ID, p.Name))
						code.emitHtmlEscapeExpr(fmt.Sprintf("func() string { tx_b, _ := json.Marshal(%s); return string(tx_b) }()", p.Name))
						code.emitStrLit("\"")
					}
				}
			}
			if isVoidElement(node.Data) {
				code.emitStrLit("/>")
				return
			}
			code.emitStrLit(">")
		}

		nodeId, _ := hasAttr(node, "id")
		if node.DataAtom == atom.Script && nodeId == "tx-runtime" {
			code.emitExpr("tx_runtime_script")
		} else if node.DataAtom == atom.Script && nodeId == "tx-saved" {
			code.emitSplit()
		} else {
			comp.emitChildren(code, node, localVars, forKeys, inSlot, comp.emitRender)
		}

		if node.DataAtom != atom.Template {
			code.emitStrLit("</" + node.Data + ">")
		}
	}
}

func (comp *Component) emitChildKey(code *Code, child *Child, forKeys []string, inSlot bool) {
	if len(forKeys) > 0 {
		code.NeedsFmt = true // the key segments below call fmt.Sprint
	}
	code.emitGo("tx_id := ")
	if inSlot {
		code.emitGo("tx_id")
		// loop keys get a ';' prefix: fmt.Sprint(key) is arbitrary user data
		// that could look like a real id segment (e.g. "tx-counter-1"), so the
		// ';' marks it as a key and keeps id parsing from mistaking it for one
		for _, key := range forKeys {
			code.emitGo(fmt.Sprintf(" + \";\" + fmt.Sprint(%s)", key))
		}
		code.emitGo(fmt.Sprintf(" + \"@%s\"", child.pos()))
	} else {
		switch comp.Type {
		case CompTypePage:
			if len(forKeys) == 0 {
				code.emitGo(fmt.Sprintf("\"%s\"", child.pos()))
			} else {
				code.emitGo(fmt.Sprintf("\";\" + fmt.Sprint(%s)", forKeys[0]))
				for _, key := range forKeys[1:] {
					code.emitGo(fmt.Sprintf(" + \";\" + fmt.Sprint(%s)", key))
				}
				code.emitGo(fmt.Sprintf(" + \":%s\"", child.pos()))
			}
		case CompTypeComp:
			code.emitGo("tx_id")
			for _, key := range forKeys {
				code.emitGo(fmt.Sprintf(" + \";\" + fmt.Sprint(%s)", key))
			}
			code.emitGo(fmt.Sprintf(" + \":%s\"", child.pos()))
		}
	}
	code.emitGo("\n")
}

func (comp *Component) emitChildren(code *Code, node *html.Node, localVars []LocalVar, forKeys []string, inSlot bool, emit func(*Code, *html.Node, []LocalVar, []string, bool)) {
	var prevCondState CondState
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		hasFor := false
		forKey := ""
		childLocalVars := localVars
		if c.Type == html.ElementNode {
			currCondState, field := condState(c)

			switch prevCondState {
			case CondStateIf:
				if currCondState <= prevCondState {
					code.emitGo("\n}\n")
				}
			case CondStateElseIf:
				if currCondState < prevCondState {
					code.emitGo("\n}\n")
				}
			case CondStateElse:
				code.emitGo("\n}\n")
			}

			switch currCondState {
			case CondStateIf:
				code.emitGo("if " + comp.rewriteIfCond(field, localVars) + " {\n")
			case CondStateElseIf:
				code.emitGo("} else if " + comp.rewriteIfCond(field, localVars) + " {\n")
			case CondStateElse:
				code.emitGo("} else {\n")
			}
			if currCondState == CondStateIf || currCondState == CondStateElseIf {
				for _, lv := range comp.ifLocals(field) {
					code.emitGo(fmt.Sprintf("_ = %s\n", lv.Name))
				}
			}

			prevCondState = currCondState

			if stmt, ok := hasAttr(c, "tx-for"); ok {
				forKey, _ = hasAttr(c, "tx-key")
				hasFor = true
				forLoc := comp.forLocals(stmt)
				code.emitGo("\nfor " + comp.rewriteForStmt(stmt, localVars) + " {\n")
				for _, lv := range forLoc {
					code.emitGo(fmt.Sprintf("_ = %s\n", lv.Name))
				}
				childLocalVars = append(slices.Clone(localVars), forLoc...)
			}
		}

		childForKeys := forKeys
		if hasFor {
			childForKeys = append(forKeys, forKey)
		}

		emit(code, c, childLocalVars, childForKeys, inSlot)

		if hasFor {
			code.emitGo("\n}\n")
		}

		if c.NextSibling == nil && (prevCondState == CondStateIf || prevCondState == CondStateElseIf || prevCondState == CondStateElse) {
			code.emitGo("\n}\n")
		}
	}
}

func (comp *Component) errf(msg string, a ...any) *Diagnostic {
	return comp.errAt(Span{}, msg, a...)
}

// errAt is errf with a source span in comp.content.
func (comp *Component) errAt(span Span, msg string, a ...any) *Diagnostic {
	return &Diagnostic{Path: comp.Path, Span: span, Message: fmt.Errorf(msg, a...).Error()}
}

// spanOf maps a [start,end) token range in the parsed script (scriptWrapPrefix +
// the <script> text) back to a Span in comp.content. Zero Span if unavailable.
func (comp *Component) spanOf(fset *token.FileSet, start, end token.Pos) Span {
	if comp.lineMap == nil {
		return Span{}
	}
	base := comp.scriptStart - len(scriptWrapPrefix)
	return Span{
		Start: comp.lineMap.Pos(base + fset.Position(start).Offset),
		End:   comp.lineMap.Pos(base + fset.Position(end).Offset),
	}
}

// nodeSpan returns the .html span of an element's open tag (<tag ...>), using
// the byte offsets recorded by annotatePositions. Returns the zero Span when the
// node has no recorded offset (so a diagnostic falls back to file level rather
// than pointing at the wrong place).
func (comp *Component) nodeSpan(node *html.Node) Span {
	start, ok := comp.nodePos[node]
	if !ok || comp.lineMap == nil {
		return Span{}
	}
	end := start
	if gt := bytes.IndexByte(comp.content[start:], '>'); gt >= 0 {
		end = start + gt + 1
	}
	return Span{Start: comp.lineMap.Pos(start), End: comp.lineMap.Pos(end)}
}

// annotatePositions records the byte offset of every element and text node in
// the parsed template by sweeping comp.content with a forward cursor in document
// order (golang.org/x/net/html drops source positions). It must run on the
// pristine parse tree, before any synthetic nodes are injected. A node whose
// text can't be located keeps no entry, so callers fall back to file level.
func (comp *Component) annotatePositions() {
	comp.nodePos = map[*html.Node]int{}
	if comp.TemplateNode == nil {
		return
	}
	cursor := 0
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		switch n.Type {
		case html.ElementNode:
			if idx := bytes.Index(comp.content[cursor:], []byte("<"+n.Data)); idx >= 0 {
				comp.nodePos[n] = cursor + idx
				cursor += idx + 1 + len(n.Data)
			}
		case html.TextNode:
			if n.Data != "" {
				if idx := bytes.Index(comp.content[cursor:], []byte(n.Data)); idx >= 0 {
					comp.nodePos[n] = cursor + idx
					cursor += idx + len(n.Data)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(comp.TemplateNode)
}

func (comp *Component) sealable() bool {
	for _, v := range comp.Vars {
		if v.Kind == VarKindProp && v.InitExpr == nil {
			return false
		}
	}
	return true
}

func (comp *Component) rootExpr() string {
	if comp.Type == CompTypePage {
		return "\"page\""
	}
	return "tx_comp.tx_target"
}

func (comp *Component) hasCompute() bool {
	for _, ehs := range comp.EventHandlers {
		if len(ehs) > 0 {
			return true
		}
	}
	return len(comp.Children) > 0
}

// rewriteVarRefs rewrites bare references to component vars and funcs into
// tx_comp selectors (count -> tx_comp.V_count), dereferencing value props. It
// mutates node in place, so raw-AST analyses (dirtyDerivedNames) must run
// before it - afterward they see the selectors and miss the bare idents.
func (comp *Component) rewriteVarRefs(node ast.Node, localVars []LocalVar) ast.Node {
	return astutil.Apply(node, func(c *astutil.Cursor) bool {
		ident, ok := c.Node().(*ast.Ident)
		if !ok {
			return true
		}
		if !atVarRefPos(c) {
			return false
		}
		for _, lv := range localVars {
			if lv.Name == ident.Name {
				return false
			}
		}
		if v := comp.varByName(ident.Name); v != nil {
			sel := &ast.SelectorExpr{
				X:   &ast.Ident{Name: "tx_comp"},
				Sel: &ast.Ident{Name: "V_" + v.Name},
			}
			switch {
			case v.Kind == VarKindProp && !v.IsFunc():
				c.Replace(&ast.StarExpr{X: sel})
			default:
				c.Replace(sel)
			}
			return false
		}
		if f := comp.funcByName(ident.Name); f != nil {
			c.Replace(&ast.SelectorExpr{
				X:   &ast.Ident{Name: "tx_comp"},
				Sel: &ast.Ident{Name: ident.Name},
			})
		}
		return false
	}, nil)
}

func (comp *Component) forLocals(s string) []LocalVar {
	f, err := parser.ParseFile(token.NewFileSet(), "", "package p\nfunc _() { for "+s+" {} }", 0)
	if err != nil {
		return nil
	}
	binder := "for " + s
	var locals []LocalVar
	switch stmt := f.Decls[0].(*ast.FuncDecl).Body.List[0].(type) {
	case *ast.RangeStmt:
		if stmt.Tok == token.DEFINE {
			for _, e := range []ast.Expr{stmt.Key, stmt.Value} {
				if id, ok := e.(*ast.Ident); ok && id.Name != "_" {
					locals = append(locals, LocalVar{Name: id.Name, Stmt: binder})
				}
			}
		}
	case *ast.ForStmt:
		if a, ok := stmt.Init.(*ast.AssignStmt); ok && a.Tok == token.DEFINE {
			for _, lhs := range a.Lhs {
				if id, ok := lhs.(*ast.Ident); ok && id.Name != "_" {
					locals = append(locals, LocalVar{Name: id.Name, Stmt: binder})
				}
			}
		}
	}
	return locals
}

func (comp *Component) ifLocals(s string) []LocalVar {
	f, err := parser.ParseFile(token.NewFileSet(), "", "package p\nfunc _() { if "+s+" {} }", 0)
	if err != nil {
		return nil
	}
	ifStmt, ok := f.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.IfStmt)
	if !ok {
		return nil
	}
	a, ok := ifStmt.Init.(*ast.AssignStmt)
	if !ok || a.Tok != token.DEFINE {
		return nil
	}
	binder := "if " + s
	var locals []LocalVar
	for _, lhs := range a.Lhs {
		if id, ok := lhs.(*ast.Ident); ok && id.Name != "_" {
			locals = append(locals, LocalVar{Name: id.Name, Stmt: binder})
		}
	}
	return locals
}

func (comp *Component) capturedLocals(body *ast.BlockStmt, node *html.Node) []LocalVar {
	var locals []LocalVar
	for n := node; n != nil; n = n.Parent {
		if stmt, ok := hasAttr(n, "tx-for"); ok {
			locals = append(locals, comp.forLocals(stmt)...)
		}
		if cond, ok := hasAttr(n, "tx-if"); ok {
			locals = append(locals, comp.ifLocals(cond)...)
		}
		if cond, ok := hasAttr(n, "tx-else-if"); ok {
			locals = append(locals, comp.ifLocals(cond)...)
		}
	}
	referenced := map[string]bool{}
	ast.Inspect(body, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			referenced[id.Name] = true
		}
		return true
	})
	var out []LocalVar
	seen := map[string]bool{}
	for _, lv := range locals {
		if referenced[lv.Name] && !seen[lv.Name] {
			out = append(out, lv)
			seen[lv.Name] = true
		}
	}
	return out
}

func (comp *Component) rewriteForStmt(s string, localVars []LocalVar) string {
	src := "package p\nfunc _() { for " + s + " {} }"
	f, err := parser.ParseFile(token.NewFileSet(), "", src, 0)
	if err != nil {
		return s
	}
	locals := append(slices.Clone(localVars), comp.forLocals(s)...)
	body := f.Decls[0].(*ast.FuncDecl).Body
	switch stmt := body.List[0].(type) {
	case *ast.ForStmt:
		r := comp.rewriteVarRefs(stmt, locals).(*ast.ForStmt)
		var b strings.Builder
		if r.Init != nil {
			b.WriteString(astToSource(r.Init))
		}
		b.WriteString("; ")
		if r.Cond != nil {
			b.WriteString(astToSource(r.Cond))
		}
		b.WriteString("; ")
		if r.Post != nil {
			b.WriteString(astToSource(r.Post))
		}
		return b.String()
	case *ast.RangeStmt:
		r := comp.rewriteVarRefs(stmt, locals).(*ast.RangeStmt)
		var b strings.Builder
		if r.Key != nil {
			b.WriteString(astToSource(r.Key))
		}
		if r.Value != nil {
			b.WriteString(", ")
			b.WriteString(astToSource(r.Value))
		}
		b.WriteString(" ")
		switch r.Tok {
		case token.DEFINE:
			b.WriteString(":=")
		case token.ASSIGN:
			b.WriteString("=")
		}
		b.WriteString(" range ")
		b.WriteString(astToSource(r.X))
		return b.String()
	}
	return s
}

func (comp *Component) rewriteIfCond(s string, localVars []LocalVar) string {
	src := "package p\nfunc _() { if " + s + " {} }"
	f, err := parser.ParseFile(token.NewFileSet(), "", src, 0)
	if err != nil {
		return s
	}
	body := f.Decls[0].(*ast.FuncDecl).Body
	ifStmt, ok := body.List[0].(*ast.IfStmt)
	if !ok {
		return s
	}
	rewritten := comp.rewriteVarRefs(ifStmt, localVars).(*ast.IfStmt)
	var b strings.Builder
	if rewritten.Init != nil {
		b.WriteString(astToSource(rewritten.Init))
		b.WriteString("; ")
	}
	b.WriteString(astToSource(rewritten.Cond))
	return b.String()
}

func (comp *Component) rewriteExpr(exprStr string, localVars []LocalVar) string {
	expr, err := parser.ParseExpr(exprStr)
	if err != nil {
		return exprStr
	}
	return astToSource(comp.rewriteVarRefs(expr, localVars))
}

func (comp *Component) varByName(name string) *Var {
	for _, v := range comp.Vars {
		if v.Name == name {
			return v
		}
	}
	return nil
}

func (comp *Component) funcByName(name string) *Func {
	for _, f := range comp.Funcs {
		if f.Decl.Name.Name == name {
			return f
		}
	}
	return nil
}

func (comp *Component) directFuncFacts(body *ast.BlockStmt) (self bool, props []*Var, calls []*Func) {
	if body == nil {
		return
	}
	ast.Inspect(body, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.AssignStmt:
			for _, lhs := range s.Lhs {
				if comp.assignsState(lhs) {
					self = true
				}
			}
		case *ast.IncDecStmt:
			if comp.assignsState(s.X) {
				self = true
			}
		case *ast.CallExpr:
			if id, ok := s.Fun.(*ast.Ident); ok {
				if v := comp.varByName(id.Name); v != nil && v.Kind == VarKindProp && v.IsFunc() {
					props = append(props, v)
				} else if g := comp.funcByName(id.Name); g != nil {
					calls = append(calls, g)
				}
			}
		}
		return true
	})
	return
}

func (comp *Component) assignsState(expr ast.Expr) bool {
	var ident *ast.Ident
	ast.Inspect(expr, func(n ast.Node) bool {
		if ident != nil {
			return false
		}
		if id, ok := n.(*ast.Ident); ok {
			ident = id
			return false
		}
		return true
	})
	if ident == nil {
		return false
	}
	v := comp.varByName(ident.Name)
	return v != nil && v.Kind == VarKindState
}

func (comp *Component) inlineEffect(src string) Effect {
	eff := Effect{}
	expr, err := parser.ParseExpr(src)
	if err != nil {
		return eff
	}
	lit, ok := expr.(*ast.FuncLit)
	if !ok {
		return eff
	}
	self, props, calls := comp.directFuncFacts(lit.Body)
	eff.Self = self
	seen := map[*Var]struct{}{}
	for _, p := range props {
		if _, ok := seen[p]; !ok {
			seen[p] = struct{}{}
			eff.FuncProps = append(eff.FuncProps, p)
		}
	}
	for _, g := range calls {
		if g.Effect.Self {
			eff.Self = true
		}
		for _, p := range g.Effect.FuncProps {
			if _, ok := seen[p]; !ok {
				seen[p] = struct{}{}
				eff.FuncProps = append(eff.FuncProps, p)
			}
		}
	}
	return eff
}

func (comp *Component) readOnlyMutations(node ast.Node) []string {
	seen := map[string]struct{}{}
	var result []string
	lhsExprs := []ast.Expr{}
	ast.Inspect(node, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			lhsExprs = append(lhsExprs, stmt.Lhs...)
		case *ast.IncDecStmt:
			lhsExprs = append(lhsExprs, stmt.X)
		}
		return true
	})
	for _, expr := range lhsExprs {
		var ident *ast.Ident
		ast.Inspect(expr, func(n ast.Node) bool {
			if ident != nil {
				return false
			}
			if id, ok := n.(*ast.Ident); ok {
				ident = id
				return false
			}
			return true
		})
		if ident == nil {
			continue
		}
		v := comp.varByName(ident.Name)
		if v == nil || v.Kind == VarKindState {
			continue
		}
		if _, dup := seen[v.Name]; dup {
			continue
		}
		seen[v.Name] = struct{}{}
		result = append(result, v.Name)
	}
	return result
}

func (comp *Component) dirtyDerivedNames(body *ast.BlockStmt) []string {
	dirty := map[string]struct{}{}
	lhsExprs := []ast.Expr{}
	ast.Inspect(body, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			lhsExprs = append(lhsExprs, stmt.Lhs...)
		case *ast.IncDecStmt:
			lhsExprs = append(lhsExprs, stmt.X)
		}
		return true
	})
	for _, expr := range lhsExprs {
		var ident *ast.Ident
		ast.Inspect(expr, func(n ast.Node) bool {
			if ident != nil {
				return false
			}
			if id, ok := n.(*ast.Ident); ok {
				ident = id
				return false
			}
			return true
		})
		if ident == nil {
			continue
		}
		if v := comp.varByName(ident.Name); v != nil && v.Kind == VarKindState {
			dirty[ident.Name] = struct{}{}
		}
	}
	result := []string{}
	for _, v := range comp.Vars {
		if v.Kind != VarKindDerived {
			continue
		}
		needsRecalc := false
		astutil.Apply(v.InitExpr, func(c *astutil.Cursor) bool {
			if needsRecalc {
				return false
			}
			id, ok := c.Node().(*ast.Ident)
			if !ok {
				return true
			}
			if !atVarRefPos(c) {
				return false
			}
			if _, ok := dirty[id.Name]; ok {
				needsRecalc = true
			}
			return false
		}, nil)
		if needsRecalc {
			result = append(result, v.Name)
			dirty[v.Name] = struct{}{}
		}
	}
	return result
}

func (comp *Component) shadowingLocals(body ast.Node) []string {
	candidates := []*ast.Ident{}
	ast.Inspect(body, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			if stmt.Tok != token.DEFINE {
				return true
			}
			for _, lhs := range stmt.Lhs {
				if id, ok := lhs.(*ast.Ident); ok {
					candidates = append(candidates, id)
				}
			}
		case *ast.RangeStmt:
			if stmt.Tok != token.DEFINE {
				return true
			}
			for _, expr := range []ast.Expr{stmt.Key, stmt.Value} {
				if id, ok := expr.(*ast.Ident); ok {
					candidates = append(candidates, id)
				}
			}
		case *ast.DeclStmt:
			genDecl, ok := stmt.Decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.VAR {
				return true
			}
			for _, spec := range genDecl.Specs {
				if valSpec, ok := spec.(*ast.ValueSpec); ok {
					candidates = append(candidates, valSpec.Names...)
				}
			}
		}
		return true
	})
	var found []string
	for _, id := range candidates {
		if id.Name == "_" {
			continue
		}
		if comp.varByName(id.Name) != nil {
			found = append(found, id.Name)
		}
	}
	return found
}

// scanTmplStr scans template text: onRaw receives each literal rune (runs of
// whitespace collapse to one space when collapseWs), onExpr receives each
// top-level { expr } trimmed, with its byte offset in str (the sourcemaps
// depend on it). Braces inside Go string and rune literals don't open or close
// expressions. Errors on an unclosed quote or brace.
func (comp *Component) scanTmplStr(str string, collapseWs bool, onRaw func(r rune), onExpr func(expr string, off int) error) error {
	if str == "" {
		return nil
	}

	braceStack := 0
	isInDoubleQuote := false
	isInSingleQuote := false
	isInBackQuote := false
	skipNext := false
	lastWasSpace := false

	expr := []byte{}
	exprStart := 0
	for i, r := range str {
		if skipNext {
			expr = append(expr, []byte(string(r))...)
			skipNext = false
			continue
		}

		if braceStack == 0 && r != '{' {
			if collapseWs && (r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\f') {
				if !lastWasSpace {
					onRaw(' ')
					lastWasSpace = true
				}
				continue
			}
			onRaw(r)
			lastWasSpace = false
			continue
		}
		lastWasSpace = false

		switch r {
		case '{':
			if braceStack == 0 {
				braceStack++
				exprStart = i + 1
			} else if isInDoubleQuote || isInSingleQuote || isInBackQuote {
				expr = append(expr, byte(r))
			} else {
				braceStack++
				expr = append(expr, byte(r))
			}
		case '}':
			if isInDoubleQuote || isInSingleQuote || isInBackQuote {
				expr = append(expr, byte(r))
			} else if braceStack == 1 {
				braceStack--
				trimmedCurrExpr := bytes.TrimSpace(expr)
				if len(trimmedCurrExpr) == 0 {
					continue
				}
				leading := len(expr) - len(bytes.TrimLeft(expr, " \t\n\r\f"))
				if err := onExpr(string(trimmedCurrExpr), exprStart+leading); err != nil {
					return err
				}
				expr = []byte{}
			} else {
				braceStack--
				expr = append(expr, byte(r))
			}
		case '"':
			if isInSingleQuote || isInBackQuote {
				expr = append(expr, byte(r))
			} else {
				isInDoubleQuote = !isInDoubleQuote
				expr = append(expr, byte(r))
			}
		case '\'':
			if isInDoubleQuote || isInBackQuote {
				expr = append(expr, byte(r))
			} else {
				isInSingleQuote = !isInSingleQuote
				expr = append(expr, byte(r))
			}
		case '`':
			if isInDoubleQuote || isInSingleQuote {
				expr = append(expr, byte(r))
			} else {
				isInBackQuote = !isInBackQuote
				expr = append(expr, byte(r))
			}
		case '\\':
			if isInDoubleQuote || isInSingleQuote {
				skipNext = true
				expr = append(expr, byte(r))
			} else {
				expr = append(expr, byte(r))
			}
		default:
			expr = append(expr, byte(r))
		}
	}

	if isInDoubleQuote || isInBackQuote || isInSingleQuote {
		return comp.errf("unclosed quote in: %s", str)
	}
	if braceStack != 0 {
		return comp.errf("unclosed { in: %s", str)
	}

	return nil
}

// Everything below is the glossary: vocabulary shared across phases, grouped
// by topic - IR data types, diagnostics, probe infra, the codegen buffer, and
// pure leaf helpers. Placement rule for the whole file: breadth decides the
// layer. Single-caller code lives right after its caller, in pipeline order;
// anything used across phases sinks down here. First appearance only breaks
// ties within a section.

type VarKind int

const (
	VarKindState VarKind = iota
	VarKindDerived
	VarKindProp
	VarKindPath
)

type Var struct {
	Kind VarKind
	Name string

	Type     string
	TypeExpr ast.Expr
	InitExpr ast.Expr
	InitCode string
	Span     Span // .html source span of the declaration (for diagnostics)
}

func (v *Var) IsFunc() bool {
	_, ok := v.TypeExpr.(*ast.FuncType)
	return ok
}

type LocalVar struct {
	Name string
	Stmt string
}

type CommentName string

const (
	CommentPath CommentName = "path"
	CommentProp CommentName = "prop"
)

type Comment struct {
	Name  CommentName
	Value string
}

type Func struct {
	Decl   *ast.FuncDecl
	Code   string
	Effect Effect
	Span   Span // .html source span of the declaration (for diagnostics)
}

type Effect struct {
	Self      bool
	FuncProps []*Var
}

type CondState int

const (
	CondStateDefault CondState = iota
	CondStateIf
	CondStateElseIf
	CondStateElse
)

type Child struct {
	Pos  string
	Comp *Component

	Args  []Arg
	Fills map[string]*Fill
}

type Fill struct {
	Node *html.Node

	Children map[*html.Node]*Child
	Code     *Code
}

func (c *Child) pos() string {
	return fmt.Sprintf("%s-%s", c.Comp.Name, c.Pos)
}

type Arg struct {
	Kind     ArgKind
	PropName string

	Val  string
	Var  *Var
	Func *Func
}

type ArgKind int

const (
	ArgKindValue ArgKind = iota
	ArgKindVar
	ArgKindFunc
)

type EventHandler struct {
	ID       int
	Key      string
	Val      string
	Code     string
	Args     []EventHandlerArg
	Captures []EventHandlerArg
}

func (comp *Component) sortedEventHandlers() []*EventHandler {
	var all []*EventHandler
	for _, ehs := range comp.EventHandlers {
		all = append(all, ehs...)
	}
	slices.SortFunc(all, func(a, b *EventHandler) int {
		return a.ID - b.ID
	})
	return all
}

type EventHandlerArg struct {
	Name string
	Type string
}

// Pos is a 1-based line/column in a source file. Col counts bytes, not UTF-16
// code units.
type Pos struct {
	Line, Col int
}

// Span is a half-open range in one .html file. The zero Span means "no precise
// location" - the diagnostic points at the whole file.
type Span struct {
	Start, End Pos
}

type Severity int

const (
	SevError Severity = iota
	SevWarning
)

// Diagnostic is a positioned compiler message. It implements error so it flows
// through errors.Join with the rest of the pipeline; Diagnose collects the
// leaves back into []Diagnostic.
type Diagnostic struct {
	Path     string
	Span     Span
	Message  string
	Severity Severity
	Related  []RelatedInfo
}

// RelatedInfo points at a secondary location, e.g. a component's prop
// declaration referenced from a usage-site error in another file.
type RelatedInfo struct {
	Path    string
	Span    Span
	Message string
}

func (d *Diagnostic) Error() string {
	if d.Span == (Span{}) {
		return d.Path + ": " + d.Message
	}
	return fmt.Sprintf("%s:%d:%d: %s", d.Path, d.Span.Start.Line, d.Span.Start.Col, d.Message)
}

// diagnostics flattens a pipeline error tree into sorted Diagnostics: leaves
// that are *Diagnostic keep their span; any others become file-less messages.
func diagnostics(err error) []Diagnostic {
	leaves := flattenErrs(err)
	ds := make([]Diagnostic, 0, len(leaves))
	for _, e := range leaves {
		var d *Diagnostic
		if errors.As(e, &d) {
			ds = append(ds, *d)
		} else {
			ds = append(ds, Diagnostic{Message: e.Error()})
		}
	}
	slices.SortStableFunc(ds, func(a, b Diagnostic) int {
		if a.Path != b.Path {
			return strings.Compare(a.Path, b.Path)
		}
		if a.Span.Start.Line != b.Span.Start.Line {
			return a.Span.Start.Line - b.Span.Start.Line
		}
		return a.Span.Start.Col - b.Span.Start.Col
	})
	return ds
}

// LineMap converts byte offsets in a source file into 1-based line/byte-column.
type LineMap struct {
	lineStarts []int
}

func newLineMap(src []byte) *LineMap {
	starts := []int{0}
	for i, b := range src {
		if b == '\n' {
			starts = append(starts, i+1)
		}
	}
	return &LineMap{lineStarts: starts}
}

// Pos returns the 1-based line and byte-column for a byte offset.
func (m *LineMap) Pos(offset int) Pos {
	lo, hi := 0, len(m.lineStarts)-1
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if m.lineStarts[mid] <= offset {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return Pos{Line: lo + 1, Col: offset - m.lineStarts[lo] + 1}
}

// Offset is the inverse of Pos: a 1-based line/col back to a byte offset.
func (m *LineMap) Offset(line, col int) int {
	if line < 1 || line > len(m.lineStarts) {
		return -1
	}
	return m.lineStarts[line-1] + col - 1
}

// flattenErrs returns the leaf errors of a tree produced by errors.Join, in
// order (errors.Join nests rather than flattens its inputs).
func flattenErrs(err error) []error {
	if err == nil {
		return nil
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		var out []error
		for _, e := range joined.Unwrap() {
			out = append(out, flattenErrs(e)...)
		}
		return out
	}
	return []error{err}
}

// sortedJoin flattens err and rejoins its leaves ordered by the text before the
// first ':', so diagnostics group by file. Returns nil when there are none.
func sortedJoin(err error) error {
	errs := flattenErrs(err)
	slices.SortStableFunc(errs, func(a, b error) int {
		pa, _, _ := strings.Cut(a.Error(), ":")
		pb, _, _ := strings.Cut(b.Error(), ":")
		return strings.Compare(pa, pb)
	})
	return errors.Join(errs...)
}

// Messages returns one diagnostic string per leaf error, sorted, for callers
// that surface diagnostics individually.
func Messages(err error) []string {
	errs := flattenErrs(sortedJoin(err))
	msgs := make([]string, len(errs))
	for i, e := range errs {
		msgs[i] = e.Error()
	}
	return msgs
}

type pkgImporter struct {
	loaded map[string]*types.Package
}

func (p *pkgImporter) Import(path string) (*types.Package, error) {
	if pkg, ok := p.loaded[path]; ok {
		return pkg, nil
	}
	return importer.Default().Import(path)
}

// lockedImporter serializes Import: analyze type-checks files in parallel, and
// importer.Default's internal package map is not safe for concurrent use.
// (pkgImporter needs no lock: its map is read-only and its fallback builds a
// fresh importer per call.)
type lockedImporter struct {
	mu  sync.Mutex
	imp types.Importer
}

func (l *lockedImporter) Import(path string) (*types.Package, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.imp.Import(path)
}

// PackageImporter builds a types.Importer that resolves the given packages
// (keyed by import path) and falls back to the standard library. Load the
// user's module packages with golang.org/x/tools/go/packages; for stdlib-only
// resolution leave Compiler.Importer nil.
func PackageImporter(loaded map[string]*types.Package) types.Importer {
	return &pkgImporter{loaded: loaded}
}

type probeAnchor struct {
	line int
	span Span
	desc string
}

type probeState struct {
	buf       strings.Builder
	line      int
	span      Span // .html span of the construct currently being emitted
	anchors   []probeAnchor
	parseErrs []*Diagnostic
}

func newProbeState() *probeState {
	return &probeState{line: 1}
}

func (p *probeState) write(s string) {
	p.buf.WriteString(s)
	p.line += strings.Count(s, "\n")
}

func (p *probeState) writef(format string, args ...any) {
	p.write(fmt.Sprintf(format, args...))
}

func (p *probeState) anchor(span Span, desc string) {
	if span == (Span{}) {
		span = p.span // fall back to the current construct's span
	}
	p.anchors = append(p.anchors, probeAnchor{line: p.line, span: span, desc: desc})
}

func (p *probeState) descAt(line int) (Span, string) {
	var best probeAnchor
	for _, a := range p.anchors {
		if a.line <= line {
			best = a
		} else {
			break
		}
	}
	return best.span, best.desc
}

func (p *probeState) parseErr(desc string, err error) {
	msg := err.Error()
	if errs, ok := err.(scanner.ErrorList); ok && len(errs) > 0 {
		msg = errs[0].Msg
	}
	p.perr("%s: %s", desc, msg)
}

// perr records a template parse/validation error at the current construct's span.
func (p *probeState) perr(format string, args ...any) {
	p.parseErrs = append(p.parseErrs, &Diagnostic{Span: p.span, Message: fmt.Sprintf(format, args...)})
}

type Counter struct {
	CurrNum int
}

func (id *Counter) next() int {
	id.CurrNum++
	return id.CurrNum
}

type SegmentType int

const (
	SegmentTypeStrLit SegmentType = iota
	SegmentTypeGo
	SegmentTypeExpr
	SegmentTypeHtmlEscapeExpr
)

type Segment struct {
	Type    SegmentType
	BufName string
	Content []byte
}

func (s Segment) empty() bool { return len(s.Content) == 0 }

type Code struct {
	NeedsFmt       bool // a Go segment calls fmt directly (emitChildKey's Sprint)
	PendingSegment Segment
	Segments       []Segment
}

func newCode(buf string) Code {
	return Code{
		PendingSegment: Segment{
			BufName: buf,
		},
	}
}

func (code *Code) emit(t SegmentType, content string) {
	isExpr := t == SegmentTypeExpr || t == SegmentTypeHtmlEscapeExpr
	if code.PendingSegment.Type != t || isExpr {
		bufName := code.flush()
		code.PendingSegment = Segment{Type: t, BufName: bufName}
	}
	code.PendingSegment.Content = append(code.PendingSegment.Content, content...)
}

func (code *Code) emitGo(content string) {
	code.emit(SegmentTypeGo, content)
}

func (code *Code) emitStrLit(content string) {
	code.emit(SegmentTypeStrLit, content)
}

func (code *Code) emitExpr(content string) {
	code.emit(SegmentTypeExpr, content)
}

func (code *Code) emitHtmlEscapeExpr(content string) {
	code.emit(SegmentTypeHtmlEscapeExpr, content)
}

func (code *Code) emitSplit() {
	code.flush()
	code.PendingSegment = Segment{
		BufName: "tx_w2",
	}
}

func (code *Code) writeTo(codeBuilder *CodeBuilder) {
	code.flush()
	codeBuilder.needsFmt = codeBuilder.needsFmt || code.NeedsFmt
	for _, segment := range code.Segments {
		switch segment.Type {
		case SegmentTypeGo:
			codeBuilder.WriteString(string(segment.Content))
		case SegmentTypeStrLit:
			codeBuilder.write("%s.WriteString(%s)\n", segment.BufName, strconv.Quote(string(segment.Content)))
		case SegmentTypeExpr:
			codeBuilder.needsFmt = true
			codeBuilder.write("fmt.Fprint(%s, %s)\n", segment.BufName, string(segment.Content))
		case SegmentTypeHtmlEscapeExpr:
			codeBuilder.needsFmt = true
			codeBuilder.needsHTML = true
			codeBuilder.write("%s.WriteString(html.EscapeString(fmt.Sprint(%s)))\n", segment.BufName, string(segment.Content))
		}
	}
}

func (code *Code) flush() string {
	bufName := code.PendingSegment.BufName
	if !code.PendingSegment.empty() {
		code.Segments = append(code.Segments, code.PendingSegment)
	}
	code.PendingSegment = Segment{}
	return bufName
}

type CodeBuilder struct {
	strings.Builder
	needsFmt  bool
	needsHTML bool
}

func (code *CodeBuilder) write(s string, params ...any) {
	if len(params) == 0 {
		code.WriteString(s)
	} else {
		fmt.Fprintf(&code.Builder, s, params...)
	}
}

//go:embed runtime.js
var runtimeScript string

// goIdent encodes a component name, route, or slot name as a valid Go
// identifier: letters, digits (not leading), and _ pass through; the runes
// that occur in real names get mnemonic codes (_SL_ slash, _HY_ hyphen, _DT_
// dot, _LB_/_RB_ braces, _EX_ for the {$} route suffix); any other rune
// becomes its uppercase hex. Every code contains a letter outside the hex
// alphabet, so a code can never equal a hex escape. Component names cannot
// collide at all - their charset is lowercase-only, the codes uppercase. Page
// routes can in the pathological case of a literal code (a filename
// containing _SL_); the collision emits duplicate decls, so the user's go
// build fails closed.
func goIdent(s string) string {
	var b strings.Builder
	b.Grow(len(s) * 3)
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch {
		case unicode.IsLetter(r) || r == '_' || (i > 0 && unicode.IsDigit(r)):
			b.WriteRune(r)
		case r == '{' && i+2 < len(runes) && runes[i+1] == '$' && runes[i+2] == '}':
			b.WriteString("_EX_")
			i += 2
		case r == '/':
			b.WriteString("_SL_")
		case r == '-':
			b.WriteString("_HY_")
		case r == '.':
			b.WriteString("_DT_")
		case r == '{':
			b.WriteString("_LB_")
		case r == '}':
			b.WriteString("_RB_")
		default:
			fmt.Fprintf(&b, "_%X_", r)
		}
	}
	return b.String()
}

// text inside these elements is serialized verbatim, no entity escaping
// (https://html.spec.whatwg.org/#serializing-html-fragments)
func isVerbatimSerialize(a atom.Atom) bool {
	switch a {
	case atom.Script, atom.Style, atom.Xmp, atom.Iframe, atom.Noembed, atom.Noframes, atom.Plaintext, atom.Noscript:
		return true
	}
	return false
}

func isWhitespacePreserving(a atom.Atom) bool {
	switch a {
	case atom.Pre, atom.Listing, atom.Textarea:
		return true
	}
	return false
}

func newTemplateNode() *html.Node {
	return &html.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Template,
		Data:     "template",
	}
}

func isTmplxScriptNode(node *html.Node) bool {
	if node.DataAtom != atom.Script {
		return false
	}

	for _, attr := range node.Attr {
		if attr.Key == "type" && attr.Val == "text/tmplx" {
			return true
		}
	}

	return false
}

func hasAttr(n *html.Node, str string) (string, bool) {
	for _, attr := range n.Attr {
		if attr.Key == str {
			return attr.Val, true
		}
	}

	return "", false
}

func cleanUpTmplxScript(node *html.Node) {
	for c := node.FirstChild; c != nil; {
		next := c.NextSibling
		if isTmplxScriptNode(c) {
			node.RemoveChild(c)
		} else {
			cleanUpTmplxScript(c)
		}
		c = next
	}
}

func astToSource(a ast.Node) string {
	var buf strings.Builder
	printer.Fprint(&buf, token.NewFileSet(), a)
	return buf.String()
}

// atVarRefPos reports whether the cursor's ident can be a variable reference -
// not a selector field, a declaration name, or a struct-literal key.
func atVarRefPos(c *astutil.Cursor) bool {
	switch c.Name() {
	case "Sel", "Names":
		return false
	case "Key":
		if _, ok := c.Parent().(*ast.KeyValueExpr); ok {
			return false
		}
	}
	return true
}

func parseComments(text string) []Comment {
	comments := []Comment{}

	if text[1] == '*' {
		text = text[2 : len(text)-2]
	} else {
		text = text[2:]
	}

	lines := strings.SplitSeq(text, "\n")
	for line := range lines {
		str := strings.TrimSpace(line)
		if str == "tx:prop" {
			comments = append(comments, Comment{
				Name: CommentProp,
			})
		} else if strings.HasPrefix(str, "tx:path") {
			val := strings.TrimSpace(str[len("tx:path"):])
			comments = append(comments, Comment{
				Name:  CommentPath,
				Value: val,
			})
		}
	}

	return comments
}

func condState(n *html.Node) (CondState, string) {
	for _, attr := range n.Attr {
		if attr.Key == "tx-if" {
			return CondStateIf, attr.Val
		}

		if attr.Key == "tx-else-if" {
			return CondStateElseIf, attr.Val
		}

		if attr.Key == "tx-else" {
			return CondStateElse, ""
		}
	}
	return CondStateDefault, ""
}

// eventCarriesValue reports whether the event's target carries a string
// .value, making event.target.value available in the handler (the runtime
// sends it as tx_ev_target_value). For keydown/keypress this is the value
// BEFORE the keystroke applies; use keyup or input for the post-keystroke
// value.
func eventCarriesValue(name string) bool {
	switch name {
	case "input", "change", "keydown", "keyup", "keypress", "blur", "focus", "focusin", "focusout", "search":
		return true
	}
	return false
}

// https://html.spec.whatwg.org/#void-elements
func isVoidElement(name string) bool {
	switch name {
	case "area", "base", "br", "col", "embed", "hr", "img", "input", "link", "meta", "source", "track", "wbr":
		return true
	}
	return false
}
