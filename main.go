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


	// create pages
	pages := []*Page{}
	filepath.WalkDir(dirPages, func(path string, d fs.DirEntry, err error) error {
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
	})

	if err := pages[0].compileTemplate(); err != nil {
		log.Println(err)
	}

	t := template.Must(template.New("index").Parse(defaultTargetTmpl))
	template.Must(t.Parse(defaultHandlerTmpl))
}

type Page struct {
	Path string

	ScriptNode *html.Node
	HtmlNode   *html.Node

	Exprs map[string]Expr
}

func ParsePage(path string, f *os.File) (*Page, error) {
	documentNode, err := html.Parse(f)
	if err != nil {
		log.Fatal(err)
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

	exprs := map[string]Expr{}
	fieldId := newId("field")
	for node := range htmlNode.Descendants() {
		if node.Type != html.TextNode {
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
	}


	return &Page{
		Path:       path,
		ScriptNode: scriptNode,
		HtmlNode:   htmlNode,

		Exprs: exprs,
	}, nil
}

func (page *Page) compileTemplate() error {
	relDir, file := filepath.Split(page.Path)
	if ext := filepath.Ext(file); ext != "" {
		file, _ = strings.CutSuffix(file, ext)
	}

	dir := path.Join(dirGen, "pages", relDir)
	os.MkdirAll(dir, 0755)

	f, err := os.Create(path.Join(dir, file+".html"))
	if err != nil {
		return err
	}

	buf := bufio.NewWriter(f)
	if err := page.render(buf, page.HtmlNode); err != nil {
		return err
	}

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

func (page *Page) generate() error {
	var code strings.Builder
	var data strings.Builder

	mapAst := &ast.CompositeLit{
		Type: &ast.MapType{
			Key:   &ast.Ident{Name: "string"},
			Value: &ast.Ident{Name: "any"},
		},
	}
	// for key, expr := range page.Exprs {
	// 	mapAst.Elts = append(mapAst.Elts, &ast.KeyValueExpr{
	// 		Key:   &ast.BasicLit{Kind: token.STRING, Value: `"` + key + `"`},
	// 		Value: expr,
	// 	})
	// }

	handlerTmpl := template.Must(template.New("handler").Parse(defaultHandlerTmpl))
	code.WriteString(page.ScriptNode.FirstChild.Data)
	printer.Fprint(&data, token.NewFileSet(), mapAst)

	var handlerStr strings.Builder
	handlerTmpl.ExecuteTemplate(&handlerStr, "handler", map[string]any{
		"method": "GET",
		"path":   "/",
		"code":   code.String(),
		"data":   data.String(),
	})

	target, err := os.Create(path.Join(dirGen, "main.go"))
	if err != nil {
		return err
	}
	targetHandler := template.Must(template.New("index").Parse(defaultTargetTmpl))
	targetHandler.Execute(target, map[string]any{
		"tmplx": handlerStr.String(),
	})

	return nil
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
			node.RemoveChild(c)
			continue
		}
		cleanUpTmplxScript(c)
	}
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

const defaultTargetTmpl = `
package main

import (
	"log"
	"net/http"
	"html/template"
)

func main() {
	tmpl := template.Must(template.ParseFiles("pages/index.tmpl"))
	// {{ .tmplx }}
	log.Fatal(http.ListenAndServe(":8080", nil))
}
`

const defaultHandlerTmpl = `
http.HandleFunc("{{ .method }} {{ .path }}", func(w http.ResponseWriter, r *http.Request) {
	{{ .code }}
	tmpl.Execute(w, {{ .data }})
})
`
