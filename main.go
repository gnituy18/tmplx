package main

import (
	"errors"
	"flag"
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

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const mimeType = "text/tmplx"

var dirPages string
var dirComponents string

func main() {
	flag.StringVar(&dirPages, "pages", path.Clean("pages"), "pages directory")
	flag.StringVar(&dirComponents, "components", path.Clean("components"), "components directory")
	flag.Parse()
	dirPages = path.Clean(dirPages)
	dirComponents = path.Clean(dirComponents)

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

	page.compile(page.HeadNode)
	page.compile(page.BodyNode)
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

	HeadNode *html.Node
	BodyNode *html.Node
}

func newPage(name, content string) (*Page, error) {
	var scriptNode, headNode, bodyNode *html.Node
	nodes, err := html.ParseFragment(strings.NewReader(content), &html.Node{Type: html.ElementNode, DataAtom: atom.Head, Data: "head"})
	if err != nil {
		return nil, err
	}

	for _, node := range nodes {
		if isTmplxScriptNode(node) && scriptNode == nil {
			scriptNode = node
		}
	}

	nodes, err = html.ParseFragment(strings.NewReader(content), &html.Node{Type: html.ElementNode, DataAtom: atom.Html, Data: "html"})
	if err != nil {
		log.Fatal(err)
	}

	for _, node := range nodes {
		if isTmplxScriptNode(node) && scriptNode == nil {
			scriptNode = node
		} else if node.DataAtom == atom.Head && headNode == nil {
			headNode = node
		} else if node.DataAtom == atom.Body && bodyNode == nil {
			bodyNode = node
		}
	}

	cleanUpTmplxScript(headNode)

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

		HeadNode: headNode,
		BodyNode: bodyNode,
	}, nil
}

func cleanUpTmplxScript(node *html.Node) {
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if isTmplxScriptNode(c) {
			node.RemoveChild(c)
		}
		cleanUpTmplxScript(c)
	}
}

func (page *Page) compile(node *html.Node) string {
	var builder = &strings.Builder{}
	compile(builder, node)
	return builder.String()
}

func compile(w io.StringWriter, node *html.Node) error {
	switch node.Type {
	case html.TextNode:
		// TODO parse syntax
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
				if _, err := w.WriteString(c.Data); err != nil {
					return err
				}
			}
		} else {
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				if err := compile(w, c); err != nil {
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
