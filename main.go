package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
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
)

const mimeType = "text/tmplx"

var dirPages string
var dirComponents string
var dirGen string

func main() {
	flag.StringVar(&dirPages, "pages", path.Clean("pages"), "pages directory")
	flag.StringVar(&dirComponents, "components", path.Clean("components"), "components directory")
	flag.StringVar(&dirGen, "gen", path.Clean("gen"), "generation directory")
	flag.Parse()
	dirPages = path.Clean(dirPages)
	dirComponents = path.Clean(dirComponents)
	dirGen = path.Clean(dirGen)

	// parse components
	comps := map[string]*Comp{}
	if err := filepath.WalkDir(dirComponents, func(path string, d fs.DirEntry, err error) error {
		relPath, err := filepath.Rel(dirComponents, path)
		if err != nil {
			return err
		}

		dir, file := filepath.Split(relPath)
		ext := filepath.Ext(file)
		if ext != ".tmplx" {
			return nil
		}

		name := strings.ReplaceAll(filepath.Join(dir, strings.TrimSuffix(file, ext)), "/", "-")
		bs, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		comp, err := newComp(name, string(bs))
		if err != nil {
			return err
		}

		comps[name] = comp

		return nil
	}); err != nil {
		log.Fatal(err)
	}

	// parse pages
	bs, err := os.ReadFile(filepath.Join(dirPages, "index.tmplx"))
	if err != nil {
		log.Fatal(err)
	}
	page, err := newPage("index", string(bs))
	if err != nil {
		log.Fatal(err)
	}

	if err := page.generate(); err != nil {
		log.Fatal(err)
	}

	t := template.Must(template.New("index").Parse(defaultTargetTmpl))
	template.Must(t.Parse(defaultHandlerTmpl))
}

type IdentType int

const (
	IdentTypeNonFunc = iota
	IdentTypeFunc
)

type Page struct {
	Name         string
	ScriptNode   *html.Node
	ScriptIdents map[string]IdentType
	Exprs        map[string]string
	ExprId       *Id

	HtmlNode *html.Node
}

func newPage(name, content string) (*Page, error) {
	var scriptNode *html.Node
	nodes, err := html.ParseFragment(strings.NewReader(content), &html.Node{Type: html.ElementNode, DataAtom: atom.Head, Data: "head"})
	if err != nil {
		return nil, err
	}

	for _, node := range nodes {
		if isTmplxScriptNode(node) && scriptNode == nil {
			scriptNode = node
		}
	}

	documentNode, err := html.Parse(strings.NewReader(content))
	if err != nil {
		log.Fatal(err)
	}

	htmlNode := documentNode.FirstChild
	cleanUpTmplxScript(htmlNode)

	scriptIdents := map[string]IdentType{}
	if scriptNode != nil {
		f, err := parser.ParseFile(token.NewFileSet(), "index", "package p\n func _() { "+scriptNode.FirstChild.Data+"}", 0)
		if err != nil {
			log.Fatal(err)
		}
		fAst, _ := f.Decls[0].(*ast.FuncDecl)
		for _, stmt := range fAst.Body.List {
			switch s := stmt.(type) {
			case *ast.DeclStmt:
				decl, ok := s.Decl.(*ast.GenDecl)
				if !ok {
					log.Println("s.Decl.(*ast.GenDecl) type assert failed")
				}

				for _, spec := range decl.Specs {
					s, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}

					var t IdentType = IdentTypeNonFunc
					if _, ok := s.Type.(*ast.FuncType); ok {
						t = IdentTypeFunc
					}

					for _, name := range s.Names {
						scriptIdents[name.Name] = t
					}
				}
			case *ast.AssignStmt:
				if s.Tok != token.DEFINE {
					continue
				}

				for i, expr := range s.Lhs {
					ident, ok := expr.(*ast.Ident)
					if !ok {
						continue
					}
					var t IdentType = IdentTypeNonFunc
					if _, ok := s.Rhs[i].(*ast.FuncLit); ok {
						t = IdentTypeFunc
					}
					scriptIdents[ident.Name] = t
				}
			}
		}
	}

	return &Page{
		Name:         name,
		ScriptNode:   scriptNode,
		ScriptIdents: scriptIdents,
		Exprs:        map[string]string{},
		ExprId:       newId("expr"),

		HtmlNode: htmlNode,
	}, nil
}

func cleanUpTmplxScript(node *html.Node) {
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if isTmplxScriptNode(c) {
			node.RemoveChild(c)
			continue
		}
		cleanUpTmplxScript(c)
	}
}

func (page *Page) generate() error {
	dir := path.Join(dirGen, "pages")
	fmt.Println(dir)
	os.MkdirAll(dir, 0755)

	file, err := os.Create(path.Join(dir, page.Name+".tmpl"))
	if err != nil {
		return err
	}
	page.render(file, page.HtmlNode)

	return nil
}

func (page *Page) render(w io.StringWriter, node *html.Node) error {
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

					id := page.ExprId.Next()
					res = append(res, []byte(fmt.Sprintf("{{.%s}}", id))...)
					page.Exprs[id] = string(trimmedCurrExpr)
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
			return errors.New(fmt.Sprintf("unclosed quote in expression: \"%s\"", node.Data))
		}
		if braceStack != 0 {
			return errors.New(fmt.Sprintf("unclosed brace in expression: \"%s\"", node.Data))
		}

		if _, err := w.WriteString(html.EscapeString(string(res))); err != nil {
			return err
		}
	case html.ElementNode:
		if _, err := w.WriteString("<"); err != nil {
			return err
		}
		if _, err := w.WriteString(node.Data); err != nil {
			return err
		}

		// TODO handle tx-
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

type Comp struct {
	Name         string
	ScriptNode   *html.Node
	TemplateNode *html.Node
}

func newComp(name, content string) (*Comp, error) {
	nodes, err := html.ParseFragment(strings.NewReader(content), &html.Node{Type: html.ElementNode, DataAtom: atom.Body, Data: "body"})
	if err != nil {
		return nil, err
	}

	var scriptNode, templateNode *html.Node

	for _, node := range nodes {
		if isTmplxScriptNode(node) && scriptNode == nil {
			scriptNode = node
		} else if node.DataAtom == atom.Template && templateNode == nil {
			templateNode = node
		}
	}

	if templateNode == nil {
		return nil, errors.New("<template> not found in " + name)
	}

	return &Comp{
		Name:         name,
		ScriptNode:   scriptNode,
		TemplateNode: templateNode,
	}, nil
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

type Id struct {
	Curr   int
	Prefix string
}

func (id *Id) Next() string {
	id.Curr++
	return fmt.Sprintf("%s_%d", id.Prefix, id.Curr)
}

func newId(prefix string) *Id {
	return &Id{
		Prefix: prefix,
	}
}

const defaultTargetTmpl = `
package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"text/template"
)
func main() {
	// {{ .tmplx }}
	log.Fatal(http.ListenAndServe(":8080", nil))
}
`

const defaultHandlerTmpl = `
{{ define "handler" }}
http.HandleFunc("{{ .method }} { .url }", func(w http.ResponseWriter, r *http.Request) {
	{{ .code }}
	
})
{{ end }}
`
