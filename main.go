package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"golang.org/x/sync/errgroup"
	goimports "golang.org/x/tools/imports"
)

const mimeType = "text/tmplx"

var dirPages string
var dirComponents string
var output string

func main() {
	flag.StringVar(&dirPages, "pages", path.Clean("pages"), "pages directory")
	flag.StringVar(&dirComponents, "components", path.Clean("components"), "components directory")
	flag.StringVar(&output, "output", "", "output file")
	flag.Parse()
	dirPages = path.Clean(dirPages)
	dirComponents = path.Clean(dirComponents)
	if output != "" {
		output = path.Clean(output)
	}

	pages := []*Page{}
	if err := filepath.WalkDir(dirPages, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("pages directory not found (file: %s): %w", path, err)
		}

		if d.IsDir() {
			return nil
		}

		_, filename := filepath.Split(path)
		ext := filepath.Ext(filename)
		if ext != ".tmplx" && ext != ".html" {
			log.Println("skip file without .tmplx or .html (file: %s)" + path)
			return nil
		}

		pages = append(pages, &Page{FilePath: path})
		return nil
	}); err != nil {
		log.Fatalln(err)
	}

	g := new(errgroup.Group)
	for _, page := range pages {
		g.Go(page.Parse)
	}
	if err := g.Wait(); err != nil {
		log.Fatalln(err)
	}

	var imports strings.Builder
	var tmplDefs strings.Builder
	var tmplHandlers strings.Builder
	tmplxHandlerTmpl := template.Must(template.New("tmplx_handler").Parse(`
{
	Url: "{{.Url}}",
	HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
		{{ .Code }}
		stateBytes, _ := json.Marshal({{ .State }})
		state := string(stateBytes)
		tmpl.ExecuteTemplate(w, "{{ .TmplName }}", {{ .Fields }})
	},
},`))
	for _, page := range pages {
		for _, im := range page.Imports {
			if _, err := imports.WriteString(astToCode(im) + "\n"); err != nil {
				log.Fatalln(fmt.Errorf("imports WriteString failed: %w", err))
			}
		}
		if _, err := tmplDefs.WriteString(fmt.Sprintf("{{define \"%s\"}}\n", page.urlPath())); err != nil {
			log.Fatalln(fmt.Errorf("tmpl defs WriteString failed: %w", err))
		}
		if err := page.render(&tmplDefs, page.HtmlNode); err != nil {
			log.Fatalln(fmt.Errorf("tmpl defs WriteString failed: %w", err))
		}
		if _, err := tmplDefs.WriteString("\n{{end}}\n"); err != nil {
			log.Fatalln(fmt.Errorf("tmpl defs WriteString failed: %w", err))
		}

		tmplxHandlerTmpl.Execute(&tmplHandlers, page.pageHandlerFields())
		for _, fields := range page.funcHandlerFields() {
			tmplxHandlerTmpl.Execute(&tmplHandlers, fields)
		}
	}

	var out strings.Builder
	out.WriteString("package tmplx\n")
	out.WriteString(fmt.Sprintf(`
import (
%s
)`, imports.String()))

	out.WriteString(`
type TmplxHandler struct {
        Url		string
	HandlerFunc 	http.HandlerFunc
}
`)
	out.WriteString(fmt.Sprintf("var tmpl = template.Must(template.New(\"tmplx_handlers\").Parse(`%s`))\n", tmplDefs.String()))
	out.WriteString("func Handlers() []TmplxHandler { return tmplxHandlers }\n\n")
	out.WriteString(fmt.Sprintf("var tmplxHandlers []TmplxHandler = []TmplxHandler{\n%s}\n", tmplHandlers.String()))

	outStr := out.String()
	formatted, err := goimports.Process("temp.go", []byte(outStr), nil)
	if err != nil {
		log.Fatalln(fmt.Errorf("format output file failed: %w", err))
	}

	if output != "" {
		f, err := createFile(output)
		if err != nil {
			log.Fatal(err)
		}

		f.Write(formatted)
		defer f.Close()
	}

	os.Stdout.Write(formatted)
}

type Page struct {
	FilePath string
	RelPath  string

	HtmlNode   *html.Node
	ScriptNode *html.Node

	Imports   []*ast.ImportSpec
	VarNames  []string
	Vars      map[string]*Var
	FuncNames []string
	Funcs     map[string]*Func

	FieldExprs map[string]Expr
}

func (page *Page) Parse() error {
	relPath, err := filepath.Rel(dirPages, page.FilePath)
	if err != nil {
		return fmt.Errorf("relative path not found: %w", err)
	}
	page.RelPath = relPath

	f, err := os.Open(page.FilePath)
	if err != nil {
		return fmt.Errorf("open page file failed: %w", err)
	}

	documentNode, err := html.Parse(f)
	if err != nil {
		return fmt.Errorf("parse html failed (file: %s): %w", page.FilePath, err)
	}
	page.HtmlNode = documentNode.FirstChild

	for n := range page.HtmlNode.Descendants() {
		if isTmplxScriptNode(n) {
			page.ScriptNode = n
			break
		}
	}
	cleanUpTmplxScript(page.HtmlNode)

	page.Imports = []*ast.ImportSpec{}
	page.VarNames = []string{}
	page.Vars = map[string]*Var{}
	page.FuncNames = []string{}
	page.Funcs = map[string]*Func{}
	funcNameGen := newIdGen("func")
	if page.ScriptNode != nil {
		f, err := parser.ParseFile(token.NewFileSet(), page.FilePath, "package p\n"+page.ScriptNode.FirstChild.Data, 0)
		if err != nil {
			return fmt.Errorf("parse script failed (file: %s): %w", page.FilePath, err)
		}

		// parse imports
		for _, decl := range f.Decls {
			d, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			if d.Tok != token.IMPORT {
				continue
			}

			for _, spec := range d.Specs {
				s, ok := spec.(*ast.ImportSpec)
				if !ok {
					return fmt.Errorf("not a import spec(file: %s): %s", page.FilePath, astToCode(spec))
				}

				page.Imports = append(page.Imports, s)
			}
		}

		// parse vars
		for _, decl := range f.Decls {
			d, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			if d.Tok != token.VAR {
				continue
			}

			for _, spec := range d.Specs {
				s, ok := spec.(*ast.ValueSpec)
				if !ok {
					return fmt.Errorf("not a value spec (file: %s): %s", page.FilePath, astToCode(spec))
				}

				if s.Type == nil {
					return fmt.Errorf("must specify a type in declaration (file: %s): %s", page.FilePath, astToCode(spec))
				}

				if len(s.Values) == 0 {
					for _, ident := range s.Names {
						page.VarNames = append(page.VarNames, ident.Name)
						page.Vars[ident.Name] = &Var{
							Name:     ident.Name,
							Type:     VarTypeState,
							TypeExpr: s.Type,
						}
					}
					continue
				}

				if len(s.Values) > len(s.Names) {
					return fmt.Errorf("extra init exprs (file: %s): %s", page.FilePath, astToCode(spec))
				}

				if len(s.Values) < len(s.Names) {
					return fmt.Errorf("missin init exprs (file: %s): %s", page.FilePath, astToCode(spec))
				}

				for i, v := range s.Values {
					found := false
					t := VarTypeState
					ast.Inspect(v, func(n ast.Node) bool {
						if found {
							return false
						}

						ident, ok := n.(*ast.Ident)
						if !ok {
							return true
						}

						if _, ok := page.Vars[ident.Name]; ok {
							found = true
							t = VarTypeDerived
							return false
						}

						return true
					})
					name := s.Names[i].Name
					page.VarNames = append(page.VarNames, name)
					page.Vars[name] = &Var{
						Name:     name,
						Type:     VarType(t),
						TypeExpr: s.Type,
						InitExpr: v,
					}
				}
			}
		}

		// parse funcs
		for _, decl := range f.Decls {
			d, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if d.Recv != nil {
				return fmt.Errorf("no method declaration (file: %s)", page.FilePath)
			}
			if d.Type.Results != nil {
				return fmt.Errorf("functions must not have return values (file: %s)", page.FilePath)
			}

			modifiedDerived := []string{}
			page.modifiedVars(d, &modifiedDerived)

			if len(modifiedDerived) > 0 {
				return fmt.Errorf("derived can not be modified (file: %s): %v", page.FilePath, modifiedDerived)
			}

			page.FuncNames = append(page.FuncNames, d.Name.Name)
			page.Funcs[d.Name.Name] = &Func{
				Name: d.Name.Name,
				Decl: d,
			}
		}
	}

	runtimeScriptNode := &html.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Script,
		Data:     "script",
		Attr: []html.Attribute{
			{Key: "tx-ignore"},
		},
	}

	runtimeScriptNode.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: `
document.addEventListener('DOMContentLoaded', function() {
  const state = JSON.parse(this.getElementById("tx-state").innerHTML)
  const addHandler = (node) => {
    const walker = document.createTreeWalker(
      node,
      NodeFilter.SHOW_ELEMENT,
      (n) => {
        for (let attr of n.attributes) {
          if (attr.name.startsWith('tx-on')) {
            return NodeFilter.FILTER_ACCEPT;
          }
        }
        return NodeFilter.FILTER_SKIP
      }
    );
    while (walker.nextNode()) {
      const cn = walker.currentNode;
      for (let attr of cn.attributes) {
        if (attr.name.startsWith('tx-on')) {
          const eName = attr.name.slice(5);
          cn.addEventListener(eName, async () => {
            const states = {}

            for (let key in state) {
              states[key] = JSON.stringify(state[key])
            }
            const res = await fetch("/tx/" + attr.value + "?" + new URLSearchParams(states).toString())
            res.text().then(html => {
              document.open()
              document.write(html)
              document.close()
            })
          })
        }
      }
    }
  }

  new MutationObserver((records) => {
    records.forEach((record) => {
      if (record.type !== 'childList') return
      records.addedNodes()
    })
  }).observe(document.documentElement, { childList: true, subList: true })
  addHandler(document.documentElement)
});
`,
	})

	page.HtmlNode.FirstChild.AppendChild(runtimeScriptNode)

	stateNode := &html.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Script,
		Data:     "script",
		Attr: []html.Attribute{
			{Key: "type", Val: "application/json"},
			{Key: "id", Val: "tx-state"},
			{Key: "tx-ignore"},
		},
	}
	stateNode.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: "{{.state}}",
	})
	page.HtmlNode.FirstChild.AppendChild(stateNode)

	page.FieldExprs = map[string]Expr{}
	fieldIdGen := newIdGen("field")
	for node := range page.HtmlNode.Descendants() {
		switch node.Type {
		case html.TextNode:
			if node.Parent.DataAtom == atom.Script || node.Parent.Data == "script" {
				continue
			}

			res, err := page.parseTmpl(node.Data, fieldIdGen)
			if err != nil {
				return err
			}

			node.Data = string(res)
		case html.ElementNode:
			ignore := false
			for _, attr := range node.Attr {
				if attr.Key == "tx-ignore" {
					ignore = true
				}
			}

			if ignore {
				continue
			}

			for i, attr := range node.Attr {
				if strings.HasPrefix(attr.Key, "tx-on") {
					expr, err := parser.ParseExpr(attr.Val)
					if err == nil {
						switch e := expr.(type) {
						case *ast.CallExpr:

							ident, isIdent := e.Fun.(*ast.Ident)
							if !isIdent {
								continue
							}

							f, ok := page.Funcs[ident.Name]
							if !ok {
								continue
							}

							params := []string{}
							for _, list := range f.Decl.Type.Params.List {
								for _, ident := range list.Names {
									params = append(params, ident.Name)
								}
							}

							if len(params) != len(e.Args) {
								return fmt.Errorf("params length not match (file: %s): %s", page.FilePath, astToCode(e))
							}

							paramsStr := ""
							if len(params) > 0 {
								for i, p := range params {
									foundVar := false
									ast.Inspect(e.Args[i], func(n ast.Node) bool {
										if foundVar {
											return false
										}

										ident, ok := n.(*ast.Ident)
										if !ok {
											return true
										}

										if _, ok := page.Vars[ident.Name]; ok {
											foundVar = true
											return false
										}

										return true
									})

									if foundVar {
										return fmt.Errorf("state and derived variables cannot be used as function parameters (file: %s): %s", page.FilePath, e.Args[i])
									}

									arg := astToCode(e.Args[i])
									if i == 0 {
										paramsStr += fmt.Sprintf("?%s={{$%s}}", p, arg)
										continue
									}

									paramsStr += fmt.Sprintf("&%s={{$%s}}", p, arg)
								}
							}

							node.Attr[i] = html.Attribute{
								Key: attr.Key,
								Val: fmt.Sprintf("%s%s", page.funcId(f.Name), paramsStr),
							}
						}
						continue
					}

					fName := funcNameGen.next()
					f, err := parser.ParseFile(token.NewFileSet(), page.FilePath, fmt.Sprintf("package p\nfunc %s() {"+attr.Val+"}", fName), 0)
					if err != nil {
						return fmt.Errorf("parse inline statement failed (file: %s): %s", page.FilePath, attr.Val)
					}
					decl, ok := f.Decls[0].(*ast.FuncDecl)
					if !ok {
						return fmt.Errorf("parse inline statement failed (file: %s): %s", page.FilePath, attr.Val)
					}

					modifiedDerived := []string{}
					page.modifiedVars(decl, &modifiedDerived)

					if len(modifiedDerived) > 0 {
						return fmt.Errorf("derived can not be modified (file: %s): %v", page.FilePath, modifiedDerived)
					}

					page.FuncNames = append(page.FuncNames, decl.Name.Name)
					page.Funcs[decl.Name.Name] = &Func{
						Name: decl.Name.Name,
						Decl: decl,
					}
					node.Attr[i] = html.Attribute{
						Key: attr.Key,
						Val: page.funcId(decl.Name.Name),
					}
					continue
				}

				res, err := page.parseTmpl(attr.Val, fieldIdGen)
				if err != nil {
					return err
				}

				node.Attr[i] = html.Attribute{
					Key: attr.Key,
					Val: string(res),
				}
			}
		}
	}

	return nil
}

// The first identifier appearing on the LHS of an assignment statement or in an inc/dec statement is a modified variable.
func (page *Page) modifiedVars(node ast.Node, md *[]string) {
	ast.Inspect(node, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			for _, lhs := range stmt.Lhs {
				found := false
				ast.Inspect(lhs, func(n ast.Node) bool {
					if found {
						return false
					}

					ident, ok := n.(*ast.Ident)
					if !ok {
						return true
					}
					found = true

					v, ok := page.Vars[ident.Name]
					if !ok {
						return false
					}

					if v.Type != VarTypeDerived {
						return false
					}

					(*md) = append((*md), v.Name)
					return false
				})
			}

			for _, rhs := range stmt.Rhs {
				page.modifiedVars(rhs, md)
			}

			return false
		case *ast.IncDecStmt:
			found := false
			ast.Inspect(stmt.X, func(n ast.Node) bool {
				if found {
					return false
				}

				ident, ok := n.(*ast.Ident)
				if !ok {
					return true
				}
				found = true

				v, ok := page.Vars[ident.Name]
				if !ok {
					return false
				}

				if v.Type != VarTypeDerived {
					return false
				}

				(*md) = append((*md), v.Name)
				return false
			})
			return false
		}

		return true
	})
}

func (page *Page) parseTmpl(str string, idGen *IdGen) ([]byte, error) {
	braceStack := 0
	isInDoubleQuote := false
	isInSingleQuote := false
	isInBackQuote := false
	skipNext := false

	expr := []byte{}
	res := []byte{}
	for _, r := range str {
		if skipNext {
			expr = append(expr, byte(r))
			skipNext = false
			continue
		}

		switch r {
		case '{':
			if braceStack == 0 {
				braceStack++
			} else if isInDoubleQuote || isInSingleQuote || isInBackQuote {
				expr = append(expr, byte(r))
			} else {
				braceStack++
				expr = append(expr, byte(r))
			}
		case '}':
			if braceStack == 0 {
				res = append(res, byte(r))
			} else if isInDoubleQuote || isInSingleQuote || isInBackQuote {
				expr = append(expr, byte(r))
			} else if braceStack == 1 {
				braceStack--
				trimmedCurrExpr := bytes.TrimSpace(expr)
				if len(trimmedCurrExpr) == 0 {
					continue
				}

				if expr, found := page.FieldExprs[string(trimmedCurrExpr)]; found {
					res = append(res, []byte(fmt.Sprintf("{{.%s}}", expr.FieldId))...)
					continue
				}

				fieldId := idGen.next()
				exprAst, err := parser.ParseExpr(string(trimmedCurrExpr))
				if err != nil {
					return nil, fmt.Errorf("parse expression error (file: %s): %s: %w", page.FilePath, string(trimmedCurrExpr), err)
				}
				page.FieldExprs[string(trimmedCurrExpr)] = Expr{
					Ast:     exprAst,
					FieldId: fieldId,
				}
				res = append(res, []byte(fmt.Sprintf("{{.%s}}", fieldId))...)
				expr = []byte{}
			} else {
				braceStack--
				expr = append(expr, byte(r))
			}
		case '"':
			if braceStack == 0 {
				res = append(res, byte(r))
			} else if isInSingleQuote || isInBackQuote {
				expr = append(expr, byte(r))
			} else {
				isInDoubleQuote = !isInDoubleQuote
				expr = append(expr, byte(r))
			}
		case '\'':
			if braceStack == 0 {
				res = append(res, byte(r))
			} else if isInDoubleQuote || isInBackQuote {
				expr = append(expr, byte(r))
			} else {
				isInSingleQuote = !isInSingleQuote
				expr = append(expr, byte(r))
			}
		case '`':
			if braceStack == 0 {
				res = append(res, byte(r))
			} else if isInDoubleQuote || isInSingleQuote {
				expr = append(expr, byte(r))
			} else {
				isInBackQuote = !isInBackQuote
				expr = append(expr, byte(r))
			}
		case '\\':
			if braceStack == 0 {
				res = append(res, byte(r))
			} else if isInDoubleQuote || isInSingleQuote {
				skipNext = true
				expr = append(expr, byte(r))
			} else {
				expr = append(expr, byte(r))
			}
		default:
			if braceStack == 0 {
				res = append(res, byte(r))
			} else {
				expr = append(expr, byte(r))
			}
		}

	}

	if isInDoubleQuote || isInBackQuote || isInSingleQuote {
		return nil, fmt.Errorf("unclosed quote in expression (file: %s): %s", page.FilePath, str)
	}
	if braceStack != 0 {
		return nil, fmt.Errorf("unclosed brace in expression (file: %s): %s", page.FilePath, str)
	}

	return res, nil
}

func (page *Page) render(w io.StringWriter, node *html.Node) error {
	switch node.Type {
	case html.TextNode:
		if _, err := w.WriteString(html.EscapeString(node.Data)); err != nil {
			return err
		}

	case html.ElementNode:
		if _, err := w.WriteString("<"); err != nil {
			return err
		}
		if _, err := w.WriteString(node.Data); err != nil {
			return err
		}

		for _, attr := range node.Attr {
			if _, err := w.WriteString(" "); err != nil {
				return err
			}
			if attr.Namespace != "" {
				if _, err := w.WriteString(node.Namespace); err != nil {
					return err
				}
				if _, err := w.WriteString(":"); err != nil {
					return err
				}
			}
			if _, err := w.WriteString(attr.Key); err != nil {
				return err
			}
			if _, err := w.WriteString(`="`); err != nil {
				return err
			}
			if _, err := w.WriteString(html.EscapeString(attr.Val)); err != nil {
				return err
			}
			if _, err := w.WriteString(`"`); err != nil {
				return err
			}
		}

		// https://html.spec.whatwg.org/#void-elements
		if isVoidElement(node.Data) {
			if node.FirstChild != nil {
				return errors.New("invalid void elements: " + node.Data)
			}
			if _, err := w.WriteString("/>"); err != nil {
				return err
			}
			return nil
		}

		if _, err := w.WriteString(">"); err != nil {
			return err
		}

		// https://html.spec.whatwg.org/multipage/parsing.html
		if c := node.FirstChild; c != nil && c.Type == html.TextNode && strings.HasPrefix(c.Data, "\n") {
			switch node.Data {
			case "pre", "listing", "textarea":
				if _, err := w.WriteString("\n"); err != nil {
					return err
				}
			}
		}

		// https://html.spec.whatwg.org/#parsing-html-fragments
		if isChildNodeRawText(node.Data) {
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				if c.Type != html.TextNode {
					continue
				}

				if _, err := w.WriteString(c.Data); err != nil {
					return err
				}
			}
		} else {
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				if err := page.render(w, c); err != nil {
					return err
				}
			}
		}

		if _, err := w.WriteString("</"); err != nil {
			return err
		}
		if _, err := w.WriteString(node.Data); err != nil {
			return err
		}
		if _, err := w.WriteString(">"); err != nil {
			return err
		}
		return nil
	}

	return nil
}

func (page *Page) urlPath() string {
	dir, file := filepath.Split(page.RelPath)
	name, _ := strings.CutSuffix(file, filepath.Ext(file))
	if name == "index" {
		name = ""
	}

	p := "/" + path.Join(dir, name)
	if found := strings.HasSuffix(p, "/"); found {
		p += "{$}"
	}

	return p
}

type HandlerFields struct {
	Url      string
	Code     string
	State    string
	Fields   string
	TmplName string
}

func (page *Page) pageHandlerFields() HandlerFields {
	var code strings.Builder
	for _, name := range page.VarNames {
		v := page.Vars[name]
		spec := &ast.ValueSpec{
			Names: []*ast.Ident{{Name: name}},
			Type:  v.TypeExpr,
		}
		if v.InitExpr != nil {
			spec.Values = []ast.Expr{v.InitExpr}
		}

		decl := &ast.GenDecl{
			Tok:   token.VAR,
			Specs: []ast.Spec{spec},
		}

		code.WriteString(astToCode(decl) + "\n")
	}
	if f, ok := page.Funcs["init"]; ok {
		for _, stmt := range f.Decl.Body.List {
			code.WriteString(astToCode(stmt) + "\n")
		}
	}
	page.writeDerivedAst(&code)

	var fields strings.Builder
	page.writeFieldsAst(&fields)

	var state strings.Builder
	page.writeStateFieldsAst(&state)

	return HandlerFields{
		Url:      page.urlPath(),
		Code:     code.String(),
		State:    state.String(),
		TmplName: page.urlPath(),
		Fields:   fields.String(),
	}
}

func (page *Page) funcHandlerFields() []HandlerFields {
	var codeVar strings.Builder
	codeVar.WriteString("query := r.URL.Query()\n")

	for _, name := range page.VarNames {
		v := page.Vars[name]
		spec := &ast.ValueSpec{
			Names: []*ast.Ident{{Name: name}},
			Type:  v.TypeExpr,
		}

		if v.Type == VarTypeDerived && v.InitExpr != nil {
			spec.Values = []ast.Expr{v.InitExpr}
		}

		decl := &ast.GenDecl{
			Tok:   token.VAR,
			Specs: []ast.Spec{spec},
		}

		printer.Fprint(&codeVar, token.NewFileSet(), decl)
		codeVar.WriteString("\n")

		if v.Type == VarTypeDerived {
			continue
		}

		codeVar.WriteString(fmt.Sprintf("json.Unmarshal([]byte(query.Get(\"%s\")), &%s)\n", name, name))
	}

	var codeDerived strings.Builder
	page.writeDerivedAst(&codeDerived)

	handlers := []HandlerFields{}
	for _, funcName := range page.FuncNames {
		var code strings.Builder
		code.WriteString(codeVar.String())

		f := page.Funcs[funcName]
		for _, stmt := range f.Decl.Body.List {
			printer.Fprint(&code, token.NewFileSet(), stmt)
			code.WriteString("\n")
		}

		code.WriteString(codeDerived.String())

		var fields strings.Builder
		page.writeFieldsAst(&fields)

		var state strings.Builder
		page.writeStateFieldsAst(&state)

		handlers = append(handlers, HandlerFields{
			Url:      "/tx/" + page.funcId(funcName),
			Code:     code.String(),
			State:    state.String(),
			Fields:   fields.String(),
			TmplName: page.urlPath(),
		})
	}

	return handlers
}

func (page *Page) writeStateFieldsAst(sb *strings.Builder) {
	stateAst := &ast.CompositeLit{
		Type: &ast.MapType{
			Key:   &ast.Ident{Name: "string"},
			Value: &ast.Ident{Name: "any"},
		},
	}

	for _, varName := range page.VarNames {
		name := page.Vars[varName].Name
		stateAst.Elts = append(stateAst.Elts, &ast.KeyValueExpr{
			Key:   &ast.BasicLit{Kind: token.STRING, Value: `"` + name + `"`},
			Value: &ast.Ident{Name: name},
		})
	}

	sb.WriteString(astToCode(stateAst))
}

func (page *Page) writeFieldsAst(sb *strings.Builder) {
	fieldsAst := &ast.CompositeLit{
		Type: &ast.MapType{
			Key:   &ast.Ident{Name: "string"},
			Value: &ast.Ident{Name: "any"},
		},
	}

	for _, expr := range page.FieldExprs {
		fieldsAst.Elts = append(fieldsAst.Elts, &ast.KeyValueExpr{
			Key:   &ast.BasicLit{Kind: token.STRING, Value: `"` + expr.FieldId + `"`},
			Value: expr.Ast,
		})
	}

	fieldsAst.Elts = append(fieldsAst.Elts, &ast.KeyValueExpr{
		Key:   &ast.BasicLit{Kind: token.STRING, Value: `"state"`},
		Value: &ast.Ident{Name: "state"},
	})

	sb.WriteString(astToCode(fieldsAst))
}

func (page *Page) writeDerivedAst(sb *strings.Builder) {
	for _, name := range page.VarNames {
		v := page.Vars[name]
		if v.Type != VarTypeDerived {
			continue
		}

		assign := &ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: name}},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{v.InitExpr},
		}

		printer.Fprint(sb, token.NewFileSet(), assign)
		sb.WriteString("\n")
	}
}

func (page *Page) funcId(funcName string) string {
	p, _ := strings.CutSuffix(page.RelPath, filepath.Ext(page.RelPath))
	p = strings.ReplaceAll(p, "/", "-")
	p = strings.ReplaceAll(p, "{", "")
	p = strings.ReplaceAll(p, "}", "")
	p += "-" + funcName
	return p
}

type VarType int

const (
	VarTypeState = iota
	VarTypeDerived
)

type Var struct {
	Name     string
	Type     VarType
	TypeExpr ast.Expr
	InitExpr ast.Expr
}

type Func struct {
	Name string
	Decl *ast.FuncDecl
}

type Expr struct {
	Ast     ast.Expr
	FieldId string
}

func isTmplxScriptNode(node *html.Node) bool {
	if node.DataAtom != atom.Script {
		return false
	}

	for _, attr := range node.Attr {
		if attr.Key == "type" && attr.Val == mimeType {
			return true
		}
	}

	return false
}

func cleanUpTmplxScript(node *html.Node) {
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if isTmplxScriptNode(c) {
			n := c.NextSibling
			node.RemoveChild(c)
			c.NextSibling = n
			continue
		}
		cleanUpTmplxScript(c)
	}
}

func astToCode(a ast.Node) string {
	var buf strings.Builder
	printer.Fprint(&buf, token.NewFileSet(), a)
	return buf.String()
}

type IdGen struct {
	Curr   int
	Prefix string
}

func (id *IdGen) next() string {
	id.Curr++
	return fmt.Sprintf("%s_%d", id.Prefix, id.Curr)
}

func newIdGen(prefix string) *IdGen {
	return &IdGen{
		Prefix: prefix,
	}
}

// https://html.spec.whatwg.org/#void-elements
func isVoidElement(name string) bool {
	switch name {
	case "area":
		return true
	case "base":
		return true
	case "br":
		return true
	case "col":
		return true
	case "embed":
		return true
	case "hr":
		return true
	case "img":
		return true
	case "input":
		return true
	case "link":
		return true
	case "meta":
		return true
	case "source":
		return true
	case "track":
		return true
	case "wbr":
		return true
	}
	return false
}

// https://html.spec.whatwg.org/#parsing-html-fragments
func isChildNodeRawText(name string) bool {
	switch name {
	case "title":
		return true
	case "textarea":
		return true
	case "style":
		return true
	case "xmp":
		return true
	case "iframe":
		return true
	case "noembed":
		return true
	case "noframes":
		return true
	case "script":
		return true
	case "noscript":
		return true
	}

	return false
}

func createFile(path string) (*os.File, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}
