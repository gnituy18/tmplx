package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

const mimeType = "text/tmplx"

var dirPages string
var dirComponents string
var dirGen string

const defaultTargetPath = "main.go"

var targetPath string

var a = change()

func change() string {
	return "a"
}

func main() {
	flag.StringVar(&dirPages, "pages", path.Clean("pages"), "pages directory")
	flag.StringVar(&dirComponents, "components", path.Clean("components"), "components directory")
	flag.StringVar(&dirGen, "gen", path.Clean("gen"), "generation directory")
	flag.StringVar(&targetPath, "target", "", "file for injecting handler codes")
	flag.Parse()
	dirPages = path.Clean(dirPages)
	dirComponents = path.Clean(dirComponents)
	dirGen = path.Clean(dirGen)

	pageHandlerTmpl := template.Must(template.New("handler").Parse(defaultPageHandlerTmpl))

	// create pages
	pages := []*Page{}
	if err := filepath.WalkDir(dirPages, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(dirPages, path)
		if err != nil {
			return err
		}

		_, file := filepath.Split(relPath)
		ext := filepath.Ext(file)
		if ext != ".tmplx" && ext != ".html" {
			log.Println("skip parsing file: " + path)
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		page, err := ParsePage(relPath, f)
		if err != nil {
			return err
		}

		pages = append(pages, page)
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	var handlers strings.Builder
	handlers.WriteString(`
		pages := []string{}
		filepath.WalkDir("pages", func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() { return nil }
			pages = append(pages, path)
			return nil
		})
		txTmpl := template.Must(template.ParseFiles(pages...))
	`)
	for _, page := range pages {
		if err := page.compileTemplate(); err != nil {
			log.Fatal(err)
		}
		pageHandlerTmpl.Execute(&handlers, page.handlerFields())
		for _, fields := range page.funcHandlerFields() {
			pageHandlerTmpl.Execute(&handlers, fields)
		}
	}

	if targetPath == "" {
		target, err := os.Create(path.Join(dirGen, "main.go"))
		if err != nil {
			log.Fatal(err)
		}

		target.WriteString(strings.Replace(defaultTargetTmpl, "// tmplx //", handlers.String(), 1))
	}
}

type Page struct {
	Path string

	ScriptNode *html.Node
	HtmlNode   *html.Node

	Vars       map[string]*Var
	VarNames   []string
	Funcs      map[string]*Func
	FuncNames  []string
	FieldExprs map[string]Expr
}

func ParsePage(path string, f *os.File) (*Page, error) {
	documentNode, err := html.Parse(f)
	if err != nil {
		return nil, err
	}

	var scriptNode *html.Node
	htmlNode := documentNode.FirstChild
	for n := range htmlNode.Descendants() {
		if isTmplxScriptNode(n) {
			scriptNode = n
			break
		}
	}
	cleanUpTmplxScript(htmlNode)

	vars := map[string]*Var{}
	varNames := []string{}
	funcs := map[string]*Func{}
	funcNames := []string{}
	if scriptNode != nil {
		f, err := parser.ParseFile(token.NewFileSet(), path, "package p\n"+scriptNode.FirstChild.Data, 0)
		if err != nil {
			return nil, err
		}

		for _, decl := range f.Decls {
			d, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			// TODO handle imports
			if d.Tok == token.CONST || d.Tok == token.TYPE || d.Tok == token.IMPORT {
				return nil, errors.New(`no const, type and import declaration`)
			}

			for _, spec := range d.Specs {
				s, ok := spec.(*ast.ValueSpec)
				if !ok {
					return nil, errors.New("not a value spec")
				}

				if s.Type == nil {
					return nil, errors.New("must specify a type ")
				}

				if len(s.Values) == 0 {
					for _, ident := range s.Names {
						vars[ident.Name] = &Var{
							Name:      ident.Name,
							Type:      VarTypeState,
							TypeExpr:  s.Type,
							Dependent: []*Var{},
						}
						varNames = append(varNames, ident.Name)
					}
					continue
				}

				if len(s.Values) > len(s.Names) {
					return nil, errors.New("extra init exprs")
				}

				if len(s.Values) < len(s.Names) {
					return nil, errors.New("missing init exprs")
				}

				for i, v := range s.Values {
					name := s.Names[i].Name
					newVar := &Var{
						Name:      name,
						TypeExpr:  s.Type,
						InitExpr:  v,
						Dependent: []*Var{},
					}
					ast.Inspect(v, func(n ast.Node) bool {
						ident, ok := n.(*ast.Ident)
						if !ok {
							return true
						}

						p, ok := vars[ident.Name]
						if ok {
							newVar.Type = VarTypeDerived
							p.Dependent = append(p.Dependent, newVar)
						}

						return true
					})
					vars[name] = newVar
					varNames = append(varNames, name)
				}
			}
		}
		for _, decl := range f.Decls {
			d, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if d.Recv != nil {
				return nil, errors.New("no method declaration")
			}
			if d.Type.Results != nil {
				return nil, errors.New("func must not have returns")
			}

			modifiedStates := []string{}
			modifiedDerived := []string{}
			modifiedVars(vars, &modifiedStates, &modifiedDerived, d)

			if len(modifiedDerived) > 0 {
				return nil, errors.New("can not modify derived")
			}

			funcs[d.Name.Name] = &Func{
				Name:           d.Name.Name,
				ModifiedStates: modifiedStates,
				Decl:           d,
			}
			funcNames = append(funcNames, d.Name.Name)
		}
	}

	tmplxJsNode := &html.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Script,
		Data:     "script",
	}

	tmplxJsNode.AppendChild(&html.Node{
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

	htmlNode.FirstChild.AppendChild(tmplxJsNode)

	stateNode := &html.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Script,
		Data:     "script",
		Attr: []html.Attribute{
			{Key: "type", Val: "application/json"},
			{Key: "id", Val: "tx-state"},
		},
	}
	stateNode.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: "{{.state}}",
	})
	htmlNode.FirstChild.AppendChild(stateNode)

	exprs := map[string]Expr{}
	fieldId := newId("field")
	// funcId := newId("func")
	for node := range htmlNode.Descendants() {
		switch node.Type {
		case html.TextNode:
			if node.Parent.DataAtom == atom.Script && node.Parent.Data == "script" {
				continue
			}

			braceStack := 0
			isInDoubleQuote := false
			isInSingleQuote := false
			isInBackQuote := false
			skipNext := false

			expr := []byte{}
			res := []byte{}
			for _, r := range node.Data {
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

						if expr, found := exprs[string(trimmedCurrExpr)]; found {
							res = append(res, []byte(fmt.Sprintf("{{.%s}}", expr.FieldId))...)
							continue
						}

						fieldId := fieldId.next()
						exprAst, err := parser.ParseExpr(string(trimmedCurrExpr))
						if err != nil {
							return nil, err
						}
						exprs[string(trimmedCurrExpr)] = Expr{
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
				return nil, errors.New(fmt.Sprintf("unclosed quote in expression: \"%s\"", node.Data))
			}
			if braceStack != 0 {
				return nil, errors.New(fmt.Sprintf("unclosed brace in expression: \"%s\"", node.Data))
			}

			node.Data = string(res)
		case html.ElementNode:
			for i, attr := range node.Attr {
				if !strings.HasPrefix(attr.Key, "tx-on") {
					continue
				}
				expr, err := parser.ParseExpr(attr.Val)
				if err == nil {
					callExpr, isCall := expr.(*ast.CallExpr)
					if !isCall {
						continue
					}

					ident, isIdent := callExpr.Fun.(*ast.Ident)
					if !isIdent {
						continue
					}

					f, ok := funcs[ident.Name]
					if !ok {
						continue
					}

					p, _ := strings.CutSuffix(path, filepath.Ext(path))
					p = strings.ReplaceAll(p, "/", "-")
					p = strings.ReplaceAll(p, "{", "")
					p = strings.ReplaceAll(p, "}", "")
					p += "-" + f.Name

					params := []string{}
					for _, list := range f.Decl.Type.Params.List {
						for _, ident := range list.Names {
							params = append(params, ident.Name)
						}
					}

					if len(params) != len(callExpr.Args) {
						return nil, errors.New("params length not match: " + f.Name)
					}

					paramsStr := ""
					if len(params) > 0 {
						for i, p := range params {
							found := false
							ast.Inspect(callExpr.Args[i], func(n ast.Node) bool {
								if found {
									return false
								}

								ident, ok := n.(*ast.Ident)
								if !ok {
									return true
								}

								if _, ok := vars[ident.Name]; ok {
									found = true
								}

								return true
							})

							if found {
								return nil, errors.New("state and derived can not be in params")
							}

							var sb strings.Builder
							printer.Fprint(&sb, token.NewFileSet(), callExpr.Args[i])
							exprStr := sb.String()
							if i == 0 {
								paramsStr += fmt.Sprintf("?%s={{$%s}}", p, exprStr)
								continue
							}
							paramsStr += fmt.Sprintf("&%s={{$%s}}", p, exprStr)
						}
					}

					node.Attr[i] = html.Attribute{
						Key: attr.Key,
						Val: fmt.Sprintf("%s%s", p, paramsStr),
					}
				}

				// TODO anonymous handler

				continue
			}
		}
	}

	return &Page{
		Path: path,

		ScriptNode: scriptNode,
		Vars:       vars,
		VarNames:   varNames,
		Funcs:      funcs,
		FuncNames:  funcNames,
		FieldExprs: exprs,

		HtmlNode: htmlNode,
	}, nil
}

func modifiedVars(vars map[string]*Var, ms, md *[]string, node ast.Node) {
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

					v, ok := vars[ident.Name]
					if !ok {
						return false
					}

					if v.Type == VarTypeState {
						(*ms) = append((*ms), v.Name)
					} else {
						(*md) = append((*md), v.Name)
					}
					found = true
					return false
				})

			}

			for _, rhs := range stmt.Rhs {
				modifiedVars(vars, ms, md, rhs)
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

				v, ok := vars[ident.Name]
				if !ok {
					return false
				}

				if v.Type == VarTypeState {
					(*ms) = append((*ms), v.Name)
				} else {
					(*md) = append((*md), v.Name)
				}
				found = true
				return false
			})
			return false
		}

		return true
	})
}

func (page *Page) compileTemplate() error {
	relDir, file := filepath.Split(page.Path)
	if ext := filepath.Ext(file); ext != "" {
		file, _ = strings.CutSuffix(file, ext)
	}

	dir := path.Join(dirGen, "pages", relDir)
	os.MkdirAll(dir, 0755)

	f, err := os.Create(path.Join(dir, file+".tmpl"))
	if err != nil {
		return err
	}

	buf := bufio.NewWriter(f)
	buf.WriteString(fmt.Sprintf("{{define \"%s\"}}\n", page.urlPath()))
	if err := page.render(buf, page.HtmlNode); err != nil {
		return err
	}
	buf.WriteString("{{end}}")

	return buf.Flush()
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
	dir, file := filepath.Split(page.Path)
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

func (page *Page) funcHandlerFields() []HandlerFields {
	var varCode strings.Builder
	varCode.WriteString("query := r.URL.Query()\n")

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

		printer.Fprint(&varCode, token.NewFileSet(), decl)
		varCode.WriteString("\n")

		if v.Type == VarTypeDerived {
			continue
		}

		varCode.WriteString(fmt.Sprintf("json.Unmarshal([]byte(query.Get(\"%s\")), &%s)\n", name, name))
	}

	var varCodeDerived strings.Builder
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

		printer.Fprint(&varCodeDerived, token.NewFileSet(), assign)
		varCodeDerived.WriteString("\n")
	}

	handlers := []HandlerFields{}
	for _, funcName := range page.FuncNames {
		var code strings.Builder
		code.WriteString(varCode.String())

		f := page.Funcs[funcName]
		for _, stmt := range f.Decl.Body.List {
			printer.Fprint(&code, token.NewFileSet(), stmt)
			code.WriteString("\n")
		}
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

		fieldsAst.Elts = append(fieldsAst.Elts, &ast.KeyValueExpr{
			Key:   &ast.BasicLit{Kind: token.STRING, Value: `"state"`},
			Value: stateAst,
		})

		var fields strings.Builder
		printer.Fprint(&fields, token.NewFileSet(), fieldsAst)
		code.WriteString(varCodeDerived.String())

		p, _ := strings.CutSuffix(page.Path, filepath.Ext(page.Path))
		p = strings.ReplaceAll(p, "/", "-")
		p = strings.ReplaceAll(p, "{", "")
		p = strings.ReplaceAll(p, "}", "")
		p += "-" + f.Name
		handlers = append(handlers, HandlerFields{
			Path:     "/tx/" + p,
			Code:     code.String(),
			Fields:   fields.String(),
			TmplName: page.urlPath(),
		})
	}

	return handlers
}

func (page *Page) handlerFields() HandlerFields {
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

		printer.Fprint(&code, token.NewFileSet(), decl)
		code.WriteString("\n")
	}
	if f, ok := page.Funcs["init"]; ok {
		for _, stmt := range f.Decl.Body.List {
			printer.Fprint(&code, token.NewFileSet(), stmt)
			code.WriteString("\n")
		}
	}

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

	fieldsAst.Elts = append(fieldsAst.Elts, &ast.KeyValueExpr{
		Key:   &ast.BasicLit{Kind: token.STRING, Value: `"state"`},
		Value: stateAst,
	})

	var fields strings.Builder
	printer.Fprint(&fields, token.NewFileSet(), fieldsAst)

	return HandlerFields{
		Path:     page.urlPath(),
		Code:     code.String(),
		Fields:   fields.String(),
		TmplName: page.urlPath(),
	}
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

type VarType int

const (
	VarTypeState = iota
	VarTypeDerived
)

type Var struct {
	Name      string
	Type      VarType
	TypeExpr  ast.Expr
	InitExpr  ast.Expr
	Dependent []*Var
}

type Func struct {
	Name           string
	Decl           *ast.FuncDecl
	ModifiedStates []string
}

type Expr struct {
	Ast     ast.Expr
	FieldId string
}

type Id struct {
	Curr   int
	Prefix string
}

func (id *Id) next() string {
	id.Curr++
	return fmt.Sprintf("%s_%d", id.Prefix, id.Curr)
}

func newId(prefix string) *Id {
	return &Id{
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

type HandlerFields struct {
	Path     string
	Code     string
	Fields   string
	TmplName string
}

const defaultPageHandlerTmpl = `
http.HandleFunc("GET {{ .Path }}", func(w http.ResponseWriter, r *http.Request) {
	{{ .Code }}
	txTmpl.ExecuteTemplate(w, "{{ .TmplName }}", {{ .Fields }})
})
`

const defaultTargetTmpl = `
package main

import (
	"log"
	"net/http"
	"html/template"
	"io/fs"
	"path/filepath"
	"encoding/json"
)

func main() {
	// tmplx //
	log.Fatal(http.ListenAndServe(":8080", nil))
}
`
