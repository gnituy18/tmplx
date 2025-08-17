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

var pagesDir string
var componentsDir string
var output string
var outputPackageName string

func main() {
	flag.StringVar(&pagesDir, "pages", path.Clean("pages"), "pages directory")
	flag.StringVar(&componentsDir, "components", path.Clean("components"), "components directory")
	flag.StringVar(&output, "output", path.Clean("./tmplx/handler.go"), "output file")
	flag.StringVar(&outputPackageName, "package", "tmplx", "output package name")
	flag.Parse()
	pagesDir = path.Clean(pagesDir)
	componentsDir = path.Clean(componentsDir)
	output = path.Clean(output)

	componentNames := []string{}
	components := map[string]*Component{}
	if err := filepath.WalkDir(componentsDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing %s: %w", path, err)
		}

		if entry.IsDir() {
			return nil
		}

		_, filename := filepath.Split(path)
		if ext := filepath.Ext(filename); ext != ".html" {
			log.Printf("skipping non-HTML file: %s\n", path)
			return nil
		}

		relPath, err := filepath.Rel(componentsDir, path)
		if err != nil {
			return fmt.Errorf("relative path not found: %w", err)
		}

		name := "tx-" + strings.ToLower(strings.ReplaceAll(relPath, "/", "-"))

		componentNames = append(componentNames, name)
		components[name] = &Component{
			FilePath: path,
			Name:     name,
		}

		return nil
	}); err != nil {
		log.Fatalln(err)
	}

	eg := new(errgroup.Group)
	for _, name := range componentNames {
		eg.Go(components[name].parse)
	}
	if err := eg.Wait(); err != nil {
		log.Fatalln(err)
	}

	pages := []*Page{}
	if err := filepath.WalkDir(pagesDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing %s: %w", path, err)
		}

		if entry.IsDir() {
			return nil
		}

		_, filename := filepath.Split(path)
		if ext := filepath.Ext(filename); ext != ".html" {
			log.Printf("skipping non-HTML file: %s\n", path)
			return nil
		}

		pages = append(pages, &Page{FilePath: path})
		return nil
	}); err != nil {
		log.Fatalln(err)
	}

	eg = new(errgroup.Group)
	for _, page := range pages {
		eg.Go(page.parse)
	}
	if err := eg.Wait(); err != nil {
		log.Fatalln(err)
	}

	var out strings.Builder
	out.WriteString("package " + outputPackageName + "\n")
	out.WriteString("import(\n")

	for _, page := range pages {
		for _, im := range page.Imports {
			if _, err := out.WriteString(astToSource(im) + "\n"); err != nil {
				log.Fatalln(fmt.Errorf("imports WriteString failed: %w", err))
			}
		}
	}
	out.WriteString(")\n")
	out.WriteString(`
type TmplxHandler struct {
        Url		string
	HandlerFunc 	http.HandlerFunc
}
`)
	out.WriteString("var runtimeScript = `" + runtimeScript + "`\n")
	for _, page := range pages {
		params := []string{}
		for _, varName := range page.VarNames {
			v := page.Vars[varName]
			params = append(params, fmt.Sprintf("%s %s", v.Name, astToSource(v.TypeExpr)))
		}
		out.WriteString(fmt.Sprintf("func render_%s(w io.Writer, state string, %s) {\n", page.pageId(), strings.Join(params, ", ")))
		for _, tmpl := range page.Tmpls {
			switch tmpl.Type {
			case TmplTypeGo:
				if _, err := out.WriteString(string(tmpl.Content)); err != nil {
					log.Fatalln(err)
				}
			case TmplTypeStrLit:
				if _, err := out.WriteString(fmt.Sprintf("w.Write([]byte(`%s`))\n", string(tmpl.Content))); err != nil {
					log.Fatalln(err)
				}
			case TmplTypeExpr:
				if _, err := out.WriteString(fmt.Sprintf("w.Write([]byte(fmt.Sprint(%s)))\n", string(tmpl.Content))); err != nil {
					log.Fatalln(err)
				}
			case TmplTypeHtmlEscapeExpr:
				if _, err := out.WriteString(fmt.Sprintf("w.Write([]byte(html.EscapeString(fmt.Sprint(%s))))\n", string(tmpl.Content))); err != nil {
					log.Fatalln(err)
				}
			case TmplTypeUrlEscapeExpr:
				if _, err := out.WriteString(fmt.Sprintf("if param, err := json.Marshal(%s); err != nil {\nlog.Panic(err)\n} else {\nw.Write([]byte(url.QueryEscape(string(param))))}\n", string(tmpl.Content))); err != nil {
					log.Fatalln(err)
				}
			}
		}

		out.WriteString("}\n")
	}

	var tmplHandlers strings.Builder
	tmplxHandlerTmpl := template.Must(template.New("tmplx_handler").Parse(`
{
	Url: "{{.Url}}",
	HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
		{{ .Code }}
		stateBytes, _ := json.Marshal({{ .State }})
		state := string(stateBytes)
		{{ .Render }}
	},
},`))
	for _, page := range pages {
		tmplxHandlerTmpl.Execute(&tmplHandlers, page.pageHandlerFields())
		for _, fields := range page.funcHandlerFields() {
			tmplxHandlerTmpl.Execute(&tmplHandlers, fields)
		}
	}

	out.WriteString("func Handlers() []TmplxHandler { return tmplxHandlers }\n\n")
	out.WriteString(fmt.Sprintf("var tmplxHandlers []TmplxHandler = []TmplxHandler{\n%s\n}\n", tmplHandlers.String()))

	data := []byte(out.String())
	formatted, err := goimports.Process(output, data, nil)
	if err != nil {
		printSourceWithLineNum(data)
		log.Fatalln(fmt.Errorf("format output file failed: %w", err))
	}

	dir := filepath.Dir(output)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalln(err)
	}
	file, err := os.OpenFile(output, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	if _, err := file.Write(formatted); err != nil {
		log.Fatal(err)
	}
}

func printSourceWithLineNum(data []byte) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	lineNum := 1
	for scanner.Scan() {
		log.Printf("%d: %s\n", lineNum, scanner.Text())
		lineNum++
	}
}

type Component struct {
	FilePath string
	Name     string

	ScriptNode   *html.Node
	TemplateNode *html.Node
	StyleNode    *html.Node
}

func (comp *Component) parse() error {
	file, err := os.Open(comp.FilePath)
	if err != nil {
		return fmt.Errorf("open component file failed: %w", err)
	}
	defer file.Close()

	nodes, err := html.ParseFragment(file, &html.Node{
		Data:     "body",
		DataAtom: atom.Body,
		Type:     html.ElementNode,
	})
	if err != nil {
		return fmt.Errorf("parse component html failed (file: %s): %w", comp.FilePath, err)
	}

	for _, node := range nodes {
		if node.Type != html.ElementNode {
			continue
		}
		switch node.DataAtom {
		case atom.Script:
			if comp.ScriptNode != nil {
				return fmt.Errorf("multiple <script> node found (file: %s)", comp.FilePath)
			}
			if val, found := hasAttr(node, "type"); !found || val != mimeType {
				return fmt.Errorf("you must have type=\"%s\" in <script> (file: %s)", mimeType, comp.FilePath)
			}

			comp.ScriptNode = node
		case atom.Template:
			if comp.TemplateNode != nil {
				return fmt.Errorf("multiple <template> node found (file: %s)", comp.FilePath)
			}
			comp.TemplateNode = node
		case atom.Style:
			if comp.StyleNode != nil {
				return fmt.Errorf("multiple <style> node found (file: %s)", comp.FilePath)
			}
			comp.StyleNode = node
		default:
			return fmt.Errorf("component must start with <script type=\"%s\"> or <template> or <style> node (file: %s)", mimeType, comp.FilePath)
		}
	}
	if comp.TemplateNode == nil {
		return fmt.Errorf("component <template> node not found (file: %s)", comp.FilePath)
	}

	fmt.Println(comp)

	return nil
}

type Page struct {
	FilePath string
	RelPath  string

	DocumentNode *html.Node
	ScriptNode   *html.Node

	Imports []*ast.ImportSpec

	VarNames []string
	Vars     map[string]*Var

	FuncNameGen *IdGen
	FuncNames   []string
	Funcs       map[string]*Func

	CurrTmplType    TmplType
	CurrTmplContent []byte
	Tmpls           []PageTmpl
}

func (page *Page) parse() error {
	// 0. Read the page file
	file, err := os.Open(page.FilePath)
	if err != nil {
		return fmt.Errorf("open page file failed: %w", err)
	}
	defer file.Close()

	page.RelPath, err = filepath.Rel(pagesDir, page.FilePath)
	if err != nil {
		return fmt.Errorf("relative path not found: %w", err)
	}

	// 1. Parse the html syntax into script node and tmpl node
	page.DocumentNode, err = html.Parse(file)
	if err != nil {
		return fmt.Errorf("parse html failed (file: %s): %w", page.FilePath, err)
	}

	for node := range page.DocumentNode.Descendants() {
		if isTmplxScriptNode(node) {
			page.ScriptNode = node
			break
		}
	}
	cleanUpTmplxScript(page.DocumentNode)

	if page.ScriptNode != nil {
		for node := range page.DocumentNode.Descendants() {
			if node.DataAtom == atom.Head {
				node.AppendChild(&html.Node{
					Type:     html.ElementNode,
					DataAtom: atom.Script,
					Data:     "script",
					Attr: []html.Attribute{
						{Key: "id", Val: "tx-runtime"},
					},
				})
				node.AppendChild(&html.Node{
					Type:     html.ElementNode,
					DataAtom: atom.Script,
					Data:     "script",
					Attr: []html.Attribute{
						{Key: "type", Val: "application/json"},
						{Key: "id", Val: "tx-state"},
					},
				})
				break
			}
		}
	}

	// 3. Parse script node
	page.Imports = []*ast.ImportSpec{}
	page.VarNames = []string{}
	page.Vars = map[string]*Var{}
	page.FuncNames = []string{}
	page.Funcs = map[string]*Func{}
	page.FuncNameGen = newIdGen("func")

	if page.ScriptNode != nil {
		scriptAst, err := parser.ParseFile(token.NewFileSet(), page.FilePath, "package p\n"+page.ScriptNode.FirstChild.Data, 0)
		if err != nil {
			return fmt.Errorf("parse script failed (file: %s): %w", page.FilePath, err)
		}

		// 3.1 Parse imports
		for _, decl := range scriptAst.Decls {
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
					return fmt.Errorf("not a import spec (file: %s): %s", page.FilePath, astToSource(spec))
				}

				page.Imports = append(page.Imports, s)
			}
		}

		// 3.2 Parse variables
		for _, decl := range scriptAst.Decls {
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
					return fmt.Errorf("not a value spec (file: %s): %s", page.FilePath, astToSource(spec))
				}

				if s.Type == nil {
					return fmt.Errorf("must specify a type in declaration (file: %s): %s", page.FilePath, astToSource(spec))
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
					return fmt.Errorf("extra init exprs (file: %s): %s", page.FilePath, astToSource(spec))
				}

				if len(s.Values) < len(s.Names) {
					return fmt.Errorf("missin init exprs (file: %s): %s", page.FilePath, astToSource(spec))
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

		// 3.3 Parse functions
		for _, decl := range scriptAst.Decls {
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
			for _, field := range d.Type.Params.List {
				for _, name := range field.Names {
					if page.Vars[name.Name] != nil {
						return fmt.Errorf("You cannot use state names as handler parameter names (file: %s): %v", page.FilePath, name)
					}
				}
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

	// 4. Parse HTML node
	if err := page.parseTmpl(page.DocumentNode); err != nil {
		return err
	}
	page.doneParsingTmpl()

	return nil
}

type VarType int

const (
	VarTypeState = iota + 1
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

var runtimeScript = `document.addEventListener('DOMContentLoaded', function() {
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
          const [fun, params] = attr.value.split("?")
          const searchParams = new URLSearchParams(params)
          const eventName = attr.name.slice(5);
          cn.addEventListener(eventName, async () => {
            for (let key in state) {
              searchParams.append(key, JSON.stringify(state[key]))
            }
            const res = await fetch("/tx/" + fun + "?" + searchParams.toString())
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
`

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

func (page *Page) parseTmpl(node *html.Node) error {
	switch node.Type {
	case html.CommentNode:
		page.writeStrLit("<!--")
		page.writeStrLit(html.EscapeString(node.Data))
		page.writeStrLit("-->")
	case html.DocumentNode:
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			page.parseTmpl(c)
		}
	case html.DoctypeNode:
		page.writeStrLit("<!DOCTYPE ")
		page.writeStrLit(node.Data)
		page.writeStrLit(">")
	case html.TextNode:
		if isChildNodeRawText(node.Parent.Data) {
			// Disable template in script and style
			if hasTxIgnoreAttr(node.Parent) || node.Parent.DataAtom == atom.Script || node.Parent.DataAtom == atom.Style {
				page.writeStrLit(node.Data)
				return nil
			}
			return page.parseTmplStr(node.Data, false)
		}

		if hasTxIgnoreAttr(node.Parent) {
			page.writeStrLit(html.EscapeString(node.Data))
			return nil
		}

		return page.parseTmplStr(node.Data, true)
	case html.ElementNode:
		isTxRuntimeScript := false
		isTxState := false
		if node.DataAtom != atom.Template {
			page.writeStrLit("<")
			page.writeStrLit(node.Data)

			isIgnore := hasTxIgnoreAttr(node)

			for _, attr := range node.Attr {
				if attr.Key == "tx-if" || attr.Key == "tx-else-if" || attr.Key == "tx-else" || attr.Key == "tx-for" {
					continue
				}
				if attr.Key == "id" && attr.Val == "tx-state" {
					isTxState = true
				}

				if attr.Key == "id" && attr.Val == "tx-runtime" {
					isTxRuntimeScript = true
				}

				page.writeStrLit(" ")
				if attr.Namespace != "" {
					page.writeStrLit(node.Namespace)
					page.writeStrLit(":")
				}
				page.writeStrLit(attr.Key)
				page.writeStrLit(`="`)

				if strings.HasPrefix(attr.Key, "tx-on") {
					if expr, err := parser.ParseExpr(attr.Val); err == nil {
						if callExpr, ok := expr.(*ast.CallExpr); ok {
							if ident, ok := callExpr.Fun.(*ast.Ident); ok {
								if fun, ok := page.Funcs[ident.Name]; ok {
									params := []string{}
									for _, list := range fun.Decl.Type.Params.List {
										for _, ident := range list.Names {
											params = append(params, ident.Name)
										}
									}

									if len(params) != len(callExpr.Args) {
										return fmt.Errorf("params length not match (file: %s): %s", page.FilePath, astToSource(callExpr))
									}

									page.writeStrLit(page.funcId(fun.Name))
									for i, param := range params {
										foundVar := false
										ast.Inspect(callExpr.Args[i], func(n ast.Node) bool {
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
											return fmt.Errorf("state and derived variables cannot be used as function parameters (file: %s): %s", page.FilePath, callExpr.Args[i])
										}

										if i == 0 {
											page.writeStrLit("?" + param + "=")
										} else {
											page.writeStrLit("&" + param + "=")
										}

										arg := astToSource(callExpr.Args[i])
										page.writeUrlEscapeExpr(arg)
									}
									page.writeStrLit(`"`)
									continue
								}
							}
						}
					}

					funcName := page.FuncNameGen.next()
					fileAst, err := parser.ParseFile(token.NewFileSet(), page.FilePath, fmt.Sprintf("package p\nfunc %s() {\n%s\n}", funcName, attr.Val), 0)
					if err != nil {
						return fmt.Errorf("parse inline statement failed (file: %s): %s", page.FilePath, attr.Val)
					}

					decl, ok := fileAst.Decls[0].(*ast.FuncDecl)
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

					page.writeStrLit(page.funcId(decl.Name.Name))
				} else if isIgnore {
					page.writeStrLit(attr.Val)
				} else if err := page.parseTmplStr(attr.Val, false); err != nil {
					return err
				}
				page.writeStrLit(`"`)
			}

			// https://html.spec.whatwg.org/#void-elements
			if isVoidElement(node.Data) {
				if node.FirstChild != nil {
					return errors.New("invalid void elements: " + node.Data)
				}

				page.writeStrLit("/>")
				return nil
			}

			page.writeStrLit(">")
		}

		// https://html.spec.whatwg.org/multipage/parsing.html
		if c := node.FirstChild; c != nil && c.Type == html.TextNode && strings.HasPrefix(c.Data, "\n") {
			switch node.Data {
			case "pre", "listing", "textarea":
				page.writeStrLit("\n")
			}
		}

		if isTxRuntimeScript {
			page.writeGo("w.Write([]byte(runtimeScript))\n")
		} else if isTxState {
			page.writeExpr("state")
		} else {
			// 0: no control flow
			// 1: if
			// 2: else-if
			// 3: else
			var prevCondState CondState
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				hasFor := false
				if c.Type == html.ElementNode {
					currCondState, field := condState(c)

					switch prevCondState {
					case CondStateDefault:
						if currCondState == CondStateElseIf || currCondState == CondStateElse {
							return fmt.Errorf("detect tx-else-if or tx-else right after non-cond node (file: %s): %s", page.FilePath, c.Data)
						}
					case CondStateIf:
						if currCondState <= prevCondState {
							page.writeGo("\n}\n")
						}
					case CondStateElseIf:
						if currCondState < prevCondState {
							page.writeGo("\n}\n")
						}
					case CondStateElse:
						if currCondState == CondStateElseIf || currCondState == CondStateElse {
							return fmt.Errorf("detect tx-else-if or tx-else right after tx-else (file: %s): %s", page.FilePath, c.Data)
						}
						page.writeGo("\n}\n")
					}

					switch currCondState {
					case CondStateDefault:
					case CondStateIf:
						page.writeGo("\nif " + field + " {\n")
					case CondStateElseIf:
						page.writeGo("\n} else if " + field + " {\n")
					case CondStateElse:
						page.writeGo("\n} else {\n")
					}

					prevCondState = currCondState

					if stmt, ok := hasForAttr(c); ok {
						hasFor = true
						page.writeGo("\nfor " + stmt + " {\n")
					}
				}

				if err := page.parseTmpl(c); err != nil {
					return err
				}

				if hasFor {
					page.writeGo("\n}\n")
				}

				if c.NextSibling == nil && (prevCondState == CondStateIf || prevCondState == CondStateElseIf || prevCondState == CondStateElse) {
					page.writeGo("\n}\n")
				}
			}
		}

		if node.DataAtom != atom.Template {
			page.writeStrLit("</")
			page.writeStrLit(node.Data)
			page.writeStrLit(">")
		}
	}
	return nil
}

func (page *Page) parseTmplStr(str string, escape bool) error {
	braceStack := 0
	isInDoubleQuote := false
	isInSingleQuote := false
	isInBackQuote := false
	skipNext := false

	expr := []byte{}
	res := []byte{}
	for _, r := range str {
		if skipNext {
			expr = append(expr, []byte(string(r))...)
			skipNext = false
			continue
		}

		if braceStack == 0 && r != '{' {
			if escape {
				page.writeStrLit(html.EscapeString(string(r)))
			} else {
				page.writeStrLit(string(r))
			}
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

				_, err := parser.ParseExpr(string(trimmedCurrExpr))
				if err != nil {
					return fmt.Errorf("parse expression error (file: %s): %s: %w", page.FilePath, string(trimmedCurrExpr), err)
				}

				if escape {
					page.writeHtmlEscapeExpr(string(trimmedCurrExpr))
				} else {
					page.writeExpr(string(trimmedCurrExpr))
				}
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
		return fmt.Errorf("unclosed quote in expression (file: %s): %s", page.FilePath, str)
	}
	if braceStack != 0 {
		return fmt.Errorf("unclosed brace in expression (file: %s): %s", page.FilePath, str)
	}

	return nil
}

type TmplType int

const (
	TmplTypeGo TmplType = iota + 1
	TmplTypeStrLit
	TmplTypeExpr
	TmplTypeHtmlEscapeExpr
	TmplTypeUrlEscapeExpr
)

type PageTmpl struct {
	Type    TmplType
	Content []byte
}

func (page *Page) writeTmpl(t TmplType, content string) {
	if page.CurrTmplType != t {
		if len(page.CurrTmplContent) != 0 {
			page.Tmpls = append(page.Tmpls, PageTmpl{
				Type:    page.CurrTmplType,
				Content: page.CurrTmplContent,
			})
		}
		page.CurrTmplType = t
		page.CurrTmplContent = []byte{}
	}

	page.CurrTmplContent = append(page.CurrTmplContent, content...)
}

func (page *Page) writeGo(content string) {
	page.writeTmpl(TmplTypeGo, content)
}

func (page *Page) writeStrLit(content string) {
	page.writeTmpl(TmplTypeStrLit, content)
}

func (page *Page) writeExpr(content string) {
	page.writeTmpl(TmplTypeExpr, content)
}

func (page *Page) writeHtmlEscapeExpr(content string) {
	page.writeTmpl(TmplTypeHtmlEscapeExpr, content)
}

func (page *Page) writeUrlEscapeExpr(content string) {
	page.writeTmpl(TmplTypeUrlEscapeExpr, content)
}

func (page *Page) doneParsingTmpl() {
	if len(page.CurrTmplContent) == 0 {
		return
	}

	page.Tmpls = append(page.Tmpls, PageTmpl{
		Type:    page.CurrTmplType,
		Content: page.CurrTmplContent,
	})
}

type CondState int

const (
	CondStateDefault CondState = iota
	CondStateIf
	CondStateElseIf
	CondStateElse
)

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

func hasAttr(n *html.Node, str string) (string, bool) {
	for _, attr := range n.Attr {
		if attr.Key == str {
			return attr.Val, true
		}
	}

	return "", false
}

func hasForAttr(n *html.Node) (string, bool) {
	for _, attr := range n.Attr {
		if attr.Key == "tx-for" {
			return attr.Val, true
		}
	}

	return "", false
}

func hasTxIgnoreAttr(n *html.Node) bool {
	for _, attr := range n.Attr {
		if attr.Key == "tx-ignore" {
			return true
		}
	}

	return false
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
	Url    string
	Code   string
	State  string
	Render string
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

		code.WriteString(astToSource(decl) + "\n")
	}
	if f, ok := page.Funcs["init"]; ok {
		for _, stmt := range f.Decl.Body.List {
			code.WriteString(astToSource(stmt) + "\n")
		}
	}
	page.writeDerivedAst(&code)

	var state strings.Builder
	page.writeStateFieldsAst(&state)

	params := []string{}
	for _, v := range page.VarNames {
		params = append(params, v)
	}
	render := fmt.Sprintf("render_%s(w, state, %s)\n", page.pageId(), strings.Join(params, ", "))

	return HandlerFields{
		Url:    page.urlPath(),
		Code:   code.String(),
		State:  state.String(),
		Render: render,
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

		fun := page.Funcs[funcName]

		for _, params := range fun.Decl.Type.Params.List {
			for _, name := range params.Names {
				spec := &ast.ValueSpec{
					Names: []*ast.Ident{{Name: name.Name}},
					Type:  params.Type,
				}
				decl := &ast.GenDecl{
					Tok:   token.VAR,
					Specs: []ast.Spec{spec},
				}
				printer.Fprint(&code, token.NewFileSet(), decl)
				code.WriteString("\n")
				code.WriteString(fmt.Sprintf("json.Unmarshal([]byte(query.Get(\"%s\")), &%s)\n", name, name))
			}
		}

		for _, stmt := range fun.Decl.Body.List {
			printer.Fprint(&code, token.NewFileSet(), stmt)
			code.WriteString("\n")
		}

		code.WriteString(codeDerived.String())

		var state strings.Builder
		page.writeStateFieldsAst(&state)

		params := []string{}
		for _, varName := range page.VarNames {
			params = append(params, varName)
		}
		render := fmt.Sprintf("render_%s(w, state, %s)\n", page.pageId(), strings.Join(params, ", "))

		handlers = append(handlers, HandlerFields{
			Url:    "/tx/" + page.funcId(funcName),
			Code:   code.String(),
			State:  state.String(),
			Render: render,
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

	sb.WriteString(astToSource(stateAst))
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

func (page *Page) pageId() string {
	p, _ := strings.CutSuffix(page.RelPath, filepath.Ext(page.RelPath))
	p = strings.ReplaceAll(p, "/", "_")
	p = strings.ReplaceAll(p, "-", "_d_")
	p = strings.ReplaceAll(p, "{", "_lb_")
	p = strings.ReplaceAll(p, "}", "_rb_")
	return p
}

func (page *Page) funcId(funcName string) string {
	return page.pageId() + "_" + funcName
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

func astToSource(a ast.Node) string {
	var buf strings.Builder
	printer.Fprint(&buf, token.NewFileSet(), a)
	return buf.String()
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
