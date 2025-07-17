package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const MIME_TYPE = "text/tmplx"

func main() {
	dirComponents := path.Join("example_project/components")
	if str, ok := os.LookupEnv("TMPLX_COMPONENTS_PATH"); ok {
		dirComponents = path.Join(str)
	}

	dirPages := path.Join("example_project/pages")
	if str, ok := os.LookupEnv("TMPLX_PAGES_PATH"); ok {
		dirPages = path.Join(str)
	}

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

		bs, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		name := strings.ReplaceAll(filepath.Join(dir, strings.TrimSuffix(file, ext)), "/", "-")
		comp, err := NewComp(name, string(bs))
		if err != nil {
			return err
		}

		comps[name] = comp

		return nil
	}); err != nil {
		log.Fatal(err)
	}

	// parse route
	bs, err := os.ReadFile(filepath.Join(dirPages, "index.html"))
	if err != nil {
		log.Fatal(err)
	}
	page, err := NewPage("index", string(bs))
	if err != nil {
		log.Fatal(err)
	}

	f, err := parser.ParseFile(token.NewFileSet(), "index", "package p\n func _() { "+page.ScriptNode.FirstChild.Data+"}", 0)
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

				for _, name := range s.Names {
					page.ScriptIdents[name.Name] = struct{}{}
				}

				// TODO memorize value type
			}

		case *ast.AssignStmt:
			if s.Tok != token.DEFINE {
				continue
			}

			for _, expr := range s.Lhs {
				ident, ok := expr.(*ast.Ident)
				if !ok {
					continue
				}
				page.ScriptIdents[ident.Name] = struct{}{}
			}

		}
	}
	fmt.Println(page.ScriptIdents)
}

type Page struct {
	Name         string
	ScriptNode   *html.Node
	HeadNode     *html.Node
	BodyNode     *html.Node
	ScriptIdents map[string]struct{}
}

func NewPage(name, snippet string) (*Page, error) {
	var scriptNode *html.Node
	nodes, err := html.ParseFragment(strings.NewReader(snippet), &html.Node{Type: html.ElementNode, DataAtom: atom.Body, Data: "body"})
	if err != nil {
		log.Fatal(err)
	}

	for _, node := range nodes {
		if node.DataAtom == atom.Script && scriptNode == nil {
			for _, attr := range node.Attr {
				if !(attr.Key == "type" && attr.Val == MIME_TYPE) {
					continue
				}

				scriptNode = node
				break
			}
		}
	}

	nodes, err = html.ParseFragment(strings.NewReader(snippet), &html.Node{Type: html.ElementNode, DataAtom: atom.Html, Data: "html"})
	if err != nil {
		log.Fatal(err)
	}

	var headNode, bodyNode *html.Node
	for _, node := range nodes {
		if node.DataAtom == atom.Body && bodyNode == nil {
			bodyNode = node
		} else if node.DataAtom == atom.Head && headNode == nil {
			headNode = node
		}
	}

	return &Page{
		Name:         name,
		ScriptNode:   scriptNode,
		ScriptIdents: map[string]struct{}{},

		HeadNode: headNode,
		BodyNode: bodyNode,
	}, nil
}

type Comp struct {
	Name         string
	ScriptNode   *html.Node
	TemplateNode *html.Node
}

func NewComp(name, snippet string) (*Comp, error) {
	nodes, err := html.ParseFragment(strings.NewReader(snippet), &html.Node{Type: html.ElementNode, DataAtom: atom.Body, Data: "body"})
	if err != nil {
		return nil, err
	}

	var scriptNode, templateNode *html.Node

	for _, node := range nodes {
		if node.DataAtom == atom.Script && scriptNode == nil {
			for _, attr := range node.Attr {
				if !(attr.Key == "type" && attr.Val == MIME_TYPE) {
					continue
				}

				scriptNode = node
				break
			}
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
