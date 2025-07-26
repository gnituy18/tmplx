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
						Expr:      v,
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

			funcs[d.Name.Name] = &Func{
				Name: d.Name.Name,
			}
		}
	}

	for _, v := range vars {
		s := ""
		for _, d := range v.Dependent {
			s += d.Name + " "
		}
	}

	fmt.Println(vars)
	fmt.Println(varNames)
	fmt.Println(funcs)

	exprs := map[string]Expr{}
	fieldId := newId("field")
	// TODO use Walk to pass idents
	for node := range htmlNode.Descendants() {
		switch node.Type {
		case html.TextNode:
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
			for _, attr := range node.Attr {
				if strings.HasPrefix(attr.Key, "tx-on") {
					if expr, err := parser.ParseExpr(attr.Val); err == nil {
						callExpr, isCall := expr.(*ast.CallExpr)
						if isCall {
							if ident, isIdent := callExpr.Fun.(*ast.Ident); isIdent {
								if funcs[ident.Name] != nil {
									// TODO named handler
									fmt.Println("tx-on " + ident.Name)
								}
							}
						}
					}
					// TODO anonymous handler

					continue
				}
			}
		}
	}

	return &Page{
		Path:       path,
		ScriptNode: scriptNode,
		HtmlNode:   htmlNode,

		FieldExprs: exprs,
	}, nil
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

				if err := page.render(w, c); err != nil {
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

func (page *Page) handlerFields() HandlerFields {
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

	var fields strings.Builder
	printer.Fprint(&fields, token.NewFileSet(), fieldsAst)
	return HandlerFields{
		Path:   page.urlPath(),
		Code:   page.ScriptNode.FirstChild.Data,
		Fields: fields.String(),
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
	Expr      ast.Expr
	Dependent []*Var
}

type Func struct {
	Name           string
	AffectedStates []string
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
	Path   string
	Code   string
	Fields string
}

const defaultPageHandlerTmpl = `
http.HandleFunc("GET {{ .Path }}", func(w http.ResponseWriter, r *http.Request) {
	{{ .Code }}
	txTmpl.ExecuteTemplate(w, "{{ .Path }}", {{ .Fields }})
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
)

func main() {
	// tmplx //
	log.Fatal(http.ListenAndServe(":8080", nil))
}
`
