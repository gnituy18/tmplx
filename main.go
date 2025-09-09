package main

import (
	"bufio"
	"bytes"
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
	"sync"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"golang.org/x/tools/imports"
)

const (
	mimeType     = "text/tmplx"
	txIgnoreKey  = "tx-ignore"
	txForKey     = "tx-for"
	txKeyKey     = "tx-key"
	txIfKey      = "tx-if"
	txElseIfKey  = "tx-else-if"
	txElseKey    = "tx-else"
	txRuntimeVal = "tx-runtime"
)

var (
	pagesDir          string
	componentsDir     string
	outputFilePath    string
	outputPackageName string

	componentNames   = []string{}
	componentsByName = map[string]*Component{}
)

func main() {
	flag.StringVar(&pagesDir, "pages", "pages", "pages directory")
	flag.StringVar(&componentsDir, "components", "components", "components directory")
	flag.StringVar(&outputFilePath, "out-file", "tmplx/handler.go", "output file path")
	flag.StringVar(&outputPackageName, "out-pkg-name", "tmplx", "output package name")
	flag.Parse()
	pagesDir = filepath.Clean(pagesDir)
	componentsDir = filepath.Clean(componentsDir)
	outputFilePath = filepath.Clean(outputFilePath)

	if outputPackageName == "" {
		log.Fatalln("output package name cannot be empty string")
	}

	if err := filepath.WalkDir(componentsDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing %s: %w", path, err)
		}

		if entry.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".html" {
			log.Printf("skipping non-HTML file: %s\n", path)
			return nil
		}

		relPath, err := filepath.Rel(componentsDir, path)
		if err != nil {
			return fmt.Errorf("relative path not found: %w", err)
		}

		basePath, _ := strings.CutSuffix(relPath, ext)

		// https://html.spec.whatwg.org/multipage/custom-elements.html#valid-custom-element-name
		name := "tx-" + strings.ToLower(strings.ReplaceAll(basePath, "/", "-"))

		componentNames = append(componentNames, name)
		componentsByName[name] = &Component{
			FilePath: path,
			RelPath:  relPath,
			Name:     name,
		}

		return nil
	}); err != nil {
		log.Fatalln(err)
	}

	pages := []*Component{}
	if err := filepath.WalkDir(pagesDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing %s: %w", path, err)
		}

		if entry.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".html" {
			log.Printf("skipping non-HTML file: %s\n", path)
			return nil
		}

		relPath, err := filepath.Rel(pagesDir, path)
		if err != nil {
			return fmt.Errorf("relative path not found: %w", err)
		}

		basePath, _ := strings.CutSuffix(relPath, ext)

		pages = append(pages, &Component{
			FilePath: path,
			RelPath:  relPath,
			Name:     basePath,
		})
		return nil
	}); err != nil {
		log.Fatalln(err)
	}

	errs := newErrors()
	var wg sync.WaitGroup
	for _, name := range componentNames {
		wg.Add(1)
		go func() {
			comp := componentsByName[name]
			file, err := os.Open(comp.FilePath)
			if err != nil {
				errs.append(comp.errf("open file failed: %w", err))
				return
			}
			defer file.Close()

			nodes, err := html.ParseFragment(file, &html.Node{
				Data:     "body",
				DataAtom: atom.Body,
				Type:     html.ElementNode,
			})
			if err != nil {
				errs.append(comp.errf("parse html failed: %w", err))
				return
			}

			comp.TemplateNode = &html.Node{
				Type:     html.ElementNode,
				DataAtom: atom.Template,
				Data:     "template",
			}
			for _, node := range nodes {
				if node.Type != html.ElementNode {
					continue
				}

				val, found := hasAttr(node, "type")
				if node.DataAtom == atom.Script && found && val == mimeType {
					if comp.TmplxScriptNode != nil {
						errs.append(comp.errf("multiple <script type=\"%s\"> node found", mimeType))
						continue
					}
					comp.TmplxScriptNode = node
				} else if node.DataAtom == atom.Style {
					if comp.StyleNode != nil {
						errs.append(comp.errf("multiple <style> node found"))
						continue
					}
					comp.StyleNode = node
				} else {
					comp.TemplateNode.AppendChild(node)
				}
			}

			tsErrs := comp.parseTmplxScript()
			errs.append(tsErrs.Errs...)

			comp.SlotNames = []string{}
			comp.Slots = map[string]struct{}{}
			psErrs := comp.parseSlots(comp.TemplateNode, false)
			errs.append(psErrs.Errs...)

			wg.Done()
		}()
	}
	wg.Wait()
	if len(errs.Errs) > 0 {
		for _, err := range errs.Errs {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	for _, name := range componentNames {
		wg.Add(1)
		comp := componentsByName[name]
		go func() {
			comp.ChildCompsIdGen = map[string]*IdGen{}
			for _, name = range componentNames {
				comp.ChildCompsIdGen[name] = newIdGen(comp.Name + "_" + name)
			}

			comp.FuncNameGen = newIdGen(comp.Id())
			ptErrs := comp.parseTmpl(comp.TemplateNode, []string{})
			errs.append(ptErrs.Errs...)

			if len(comp.CurrRenderFuncContent) > 0 {
				comp.RenderFuncCodes = append(comp.RenderFuncCodes, RenderFunc{
					Type:    comp.CurrRenderFuncType,
					Content: comp.CurrRenderFuncContent,
				})
			}

			wg.Done()
		}()
	}
	for _, page := range pages {
		wg.Add(1)
		go func() {
			file, err := os.Open(page.FilePath)
			if err != nil {
				errs.append(page.errf("open page file failed: %w", err))
				return
			}
			defer file.Close()

			page.TemplateNode, err = html.Parse(file)
			if err != nil {
				errs.append(page.errf("parse html failed: %w", err))
				return
			}

			for node := range page.TemplateNode.Descendants() {
				if isTmplxScriptNode(node) {
					page.TmplxScriptNode = node
					break
				}
			}
			cleanUpTmplxScript(page.TemplateNode)

			txStateNode := &html.Node{
				Type:     html.ElementNode,
				DataAtom: atom.Script,
				Data:     "script",
				Attr: []html.Attribute{
					{Key: "type", Val: "application/json"},
					{Key: "id", Val: "tx-state"},
				},
			}

			txStateNode.AppendChild(&html.Node{
				Type: html.TextNode,
				Data: "TX_STATE_JSON",
			})

			for node := range page.TemplateNode.Descendants() {
				if node.DataAtom == atom.Head {
					if page.TmplxScriptNode != nil {
						node.AppendChild(&html.Node{
							Type:     html.ElementNode,
							DataAtom: atom.Script,
							Data:     "script",
							Attr: []html.Attribute{
								{Key: "id", Val: txRuntimeVal},
							},
						})
					}
					node.AppendChild(txStateNode)
					break
				}
			}

			if tsErrs := page.parseTmplxScript(); tsErrs != nil {
				errs.append(tsErrs.Errs...)
			}

			page.ChildCompsIdGen = map[string]*IdGen{}
			for _, name := range componentNames {
				page.ChildCompsIdGen[name] = newIdGen(page.Name + "_" + name)
			}

			page.FuncNameGen = newIdGen(page.Id())
			if ptErrs := page.parseTmpl(page.TemplateNode, []string{}); ptErrs != nil {
				errs.append(ptErrs.Errs...)
			}

			if len(page.CurrRenderFuncContent) > 0 {
				page.RenderFuncCodes = append(page.RenderFuncCodes, RenderFunc{
					Type:    page.CurrRenderFuncType,
					Content: page.CurrRenderFuncContent,
				})
			}

			wg.Done()
		}()
	}

	wg.Wait()
	if len(errs.Errs) > 0 {
		for _, err := range errs.Errs {
			fmt.Println(err)
		}
		os.Exit(1)
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
	out.WriteString("var runtimeScript = `" + runtimeScript + "`\n")
	out.WriteString(`
type TmplxHandler struct {
        Url		string
	HandlerFunc 	http.HandlerFunc
}
`)
	for _, name := range componentNames {
		comp := componentsByName[name]
		out.WriteString(fmt.Sprintf("type state_%s struct {\n", comp.Id()))
		for _, varName := range comp.VarNames {
			v := comp.Vars[varName]
			out.WriteString(fmt.Sprintf("%s %s `json:\"%s\"`\n", v.StructField, astToSource(v.TypeExpr), v.Name))
		}
		out.WriteString("}\n")

		paramsStr := ""
		for _, varName := range comp.VarNames {
			v := comp.Vars[varName]
			if v.Type == VarTypeState {
				paramsStr += fmt.Sprintf(", %s %s", v.Name, astToSource(v.TypeExpr))
			}
		}
		for _, slotName := range comp.SlotNames {
			if slotName != "" {
				paramsStr += fmt.Sprintf(",render_slot_%s func()", slotName)
			} else {
				paramsStr += fmt.Sprintf(",render_default_slot func()")
			}
		}

		out.WriteString(fmt.Sprintf("func render_%s(w io.Writer, key string, states, newStates map[string]string %s) {\n", comp.Id(), paramsStr))
		comp.implRenderFunc(&out)
		out.WriteString("}\n")
	}

	for _, page := range pages {
		out.WriteString(fmt.Sprintf("type state_%s struct {\n", page.Id()))
		for _, varName := range page.VarNames {
			v := page.Vars[varName]
			out.WriteString(fmt.Sprintf("%s %s\n", v.Name, astToSource(v.TypeExpr)))
		}
		out.WriteString("}\n")

		params := []string{}
		for _, varName := range page.VarNames {
			v := page.Vars[varName]
			params = append(params, fmt.Sprintf("%s %s", v.Name, astToSource(v.TypeExpr)))
		}
		out.WriteString(fmt.Sprintf("func render_%s(w io.Writer, key string, states, newStates map[string]string, %s) {\n", page.Id(), strings.Join(params, ", ")))
		page.implRenderFunc(&out)
		out.WriteString("}\n")
	}

	out.WriteString("var tmplxHandlers []TmplxHandler = []TmplxHandler{\n")
	for _, page := range pages {
		out.WriteString("{\n")
		out.WriteString("Url: \"" + page.urlPath() + "\",\n")
		out.WriteString("HandlerFunc: func(w http.ResponseWriter, r *http.Request) {\n")
		for _, name := range page.VarNames {
			v := page.Vars[name]
			out.WriteString(fmt.Sprintf("var %s %s", v.Name, astToSource(v.TypeExpr)))
			if v.InitExpr != nil {
				out.WriteString(fmt.Sprintf(" = %s\n", astToSource(v.InitExpr)))
			} else {
				out.WriteString("\n")
			}
		}
		if f, ok := page.Funcs["init"]; ok {
			for _, stmt := range f.Decl.Body.List {
				out.WriteString(astToSource(stmt) + "\n")
			}
			for _, name := range page.VarNames {
				v := page.Vars[name]
				if v.Type != VarTypeDerived {
					continue
				}

				out.WriteString(fmt.Sprintf("%s = %s\n", name, astToSource(v.InitExpr)))
			}
		}
		out.WriteString(fmt.Sprintf("state := &state_%s{\n", page.Id()))
		for _, name := range page.VarNames {
			out.WriteString(fmt.Sprintf("%s: %s,\n", name, name))
		}
		out.WriteString("}\n")
		out.WriteString("newStates := map[string]string{}\n")
		out.WriteString("txStateBytes, _ := json.Marshal(state)\n")
		out.WriteString("newStates[\"\"] = string(txStateBytes)\n")
		out.WriteString("var buf bytes.Buffer\n")
		out.WriteString(fmt.Sprintf("render_%s(&buf, \"\", map[string]string{}, newStates", page.Id()))
		for _, name := range page.VarNames {
			out.WriteString(fmt.Sprintf(", %s", name))
		}
		out.WriteString(")\n")
		out.WriteString("stateStr, _ := json.Marshal(newStates)\n")
		out.WriteString("w.Write(bytes.Replace(buf.Bytes(), []byte(\"TX_STATE_JSON\"), stateStr, 1))\n")
		out.WriteString("},\n")
		out.WriteString("},\n")
	}
	out.WriteString("}\n")

	out.WriteString("func Handlers() []TmplxHandler { return tmplxHandlers }\n\n")

	data := []byte(out.String())
	formatted, err := imports.Process(outputFilePath, data, nil)
	if err != nil {
		scanner := bufio.NewScanner(bytes.NewReader(data))
		lineNum := 1
		for scanner.Scan() {
			log.Printf("%d: %s\n", lineNum, scanner.Text())
			lineNum++
		}
		log.Fatalln(fmt.Errorf("format output file failed: %w", err))
	}

	dir := filepath.Dir(outputFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalln(err)
	}
	file, err := os.OpenFile(outputFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	if _, err := file.Write(formatted); err != nil {
		log.Fatal(err)
	}
}

type Component struct {
	FilePath string
	RelPath  string

	Name string

	TmplxScriptNode *html.Node
	Imports         []*ast.ImportSpec
	VarNames        []string
	Vars            map[string]*Var
	FuncNames       []string
	Funcs           map[string]*Func
	FuncNameGen     *IdGen

	TemplateNode    *html.Node
	SlotNames       []string
	Slots           map[string]struct{}
	ChildCompsIdGen map[string]*IdGen

	CurrRenderFuncType    RenderFuncType
	CurrRenderFuncContent []byte
	RenderFuncCodes       []RenderFunc

	StyleNode *html.Node
}

func (comp *Component) errf(msg string, a ...any) error {
	return fmt.Errorf("comp:"+comp.RelPath+": "+msg, a...)
}

func (comp *Component) parseTmplxScript() *Errors {
	errs := newErrors()

	comp.Imports = []*ast.ImportSpec{}
	comp.VarNames = []string{}
	comp.Vars = map[string]*Var{}
	comp.FuncNames = []string{}
	comp.Funcs = map[string]*Func{}

	if comp.TmplxScriptNode != nil {
		// TODO: save position into errors
		scriptAst, err := parser.ParseFile(token.NewFileSet(), "", "package p\n"+comp.TmplxScriptNode.FirstChild.Data, 0)
		if err != nil {
			errs.append(comp.errf("parse tmplx script failed: %w", err))
			return errs
		}

		for _, decl := range scriptAst.Decls {
			switch d := decl.(type) {
			case *ast.BadDecl:
				errs.append(comp.errf("bad declaration: %s", astToSource(decl)))
			case *ast.GenDecl:
				switch d.Tok {
				case token.IMPORT:
					for _, spec := range d.Specs {
						s, ok := spec.(*ast.ImportSpec)
						if !ok {
							errs.append(comp.errf("not a import spec: %s", astToSource(spec)))
							continue
						}

						comp.Imports = append(comp.Imports, s)
					}
				case token.VAR:
					for _, spec := range d.Specs {
						s, ok := spec.(*ast.ValueSpec)
						if !ok {
							errs.append(comp.errf("not a value spec: %s", astToSource(spec)))
							continue
						}

						if s.Type == nil {
							errs.append(comp.errf("must specify a type in declaration: %s", astToSource(spec)))
						}

						for _, ident := range s.Names {
							comp.VarNames = append(comp.VarNames, ident.Name)
							comp.Vars[ident.Name] = &Var{
								Name:        ident.Name,
								StructField: "S_" + ident.Name,
								TypeExpr:    s.Type,
							}
						}

						if len(s.Values) == 0 {
							continue
						}

						if len(s.Values) > len(s.Names) {
							errs.append(comp.errf("extra init exprs: %s", astToSource(spec)))
							continue
						}

						if len(s.Values) < len(s.Names) {
							errs.append(comp.errf("missin init exprs: %s", astToSource(spec)))
							continue
						}

						for i, v := range s.Values {
							found := false
							ast.Inspect(v, func(n ast.Node) bool {
								if found {
									return false
								}

								ident, ok := n.(*ast.Ident)
								if !ok {
									return true
								}

								if _, ok := comp.Vars[ident.Name]; !ok {
									return true
								}

								found = true
								return false
							})

							if found {
								comp.Vars[s.Names[i].Name].Type = VarTypeDerived
							} else {
								comp.Vars[s.Names[i].Name].Type = VarTypeState
							}

							comp.Vars[s.Names[i].Name].InitExpr = v
						}
					}
				}
			}
		}

		for _, decl := range scriptAst.Decls {
			d, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if d.Recv != nil {
				errs.append(comp.errf("%s: no method declaration", d.Name))
			}

			if d.Type.Results != nil {
				errs.append(comp.errf("%s: functions must not have return values", d.Name))
			}

			for _, field := range d.Type.Params.List {
				for _, name := range field.Names {
					if comp.Vars[name.Name] != nil {
						errs.append(comp.errf("%s: cannot use state names as handler parameter names: %s", d.Name, name.Name))
					}
				}
			}

			if d.Body != nil {
				deriveds := []string{}
				comp.modifiedDerived(d.Body, &deriveds)
				for _, derived := range deriveds {
					errs.append(comp.errf("%s: derived state cannot be modified: %s", d.Name, derived))
				}
			}

			comp.FuncNames = append(comp.FuncNames, d.Name.Name)
			comp.Funcs[d.Name.Name] = &Func{
				Name: d.Name.Name,
				Decl: d,
			}
		}
	}

	return errs
}

func (comp *Component) modifiedDerived(node ast.Node, md *[]string) {
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

					v, ok := comp.Vars[ident.Name]
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
				comp.modifiedDerived(rhs, md)
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

				v, ok := comp.Vars[ident.Name]
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

func (comp *Component) parseTmpl(node *html.Node, forKeys []string) *Errors {
	switch node.Type {
	case html.CommentNode:
		return nil
	case html.DocumentNode:
		errs := newErrors()
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			errs.append(comp.parseTmpl(c, forKeys).Errs...)
		}
		return errs
	case html.DoctypeNode:
		comp.writeStrLit("<!DOCTYPE ")
		comp.writeStrLit(node.Data)
		comp.writeStrLit(">")
		return nil
	case html.TextNode:
		if isChildNodeRawText(node.Parent.Data) {
			if _, found := hasAttr(node.Parent, txIgnoreKey); found || node.Parent.DataAtom == atom.Script || node.Parent.DataAtom == atom.Style {
				comp.writeStrLit(node.Data)
				return nil
			} else {
				return newErrors(comp.parseTmplStr(node.Data, false))
			}

		} else {
			if _, found := hasAttr(node.Parent, txIgnoreKey); found {
				comp.writeStrLit(html.EscapeString(node.Data))
				return nil
			} else {

				return newErrors(comp.parseTmplStr(node.Data, true))
			}
		}
	case html.ElementNode:
		errs := newErrors()
		if componentsByName[node.Data] != nil {
			childComp := componentsByName[node.Data]

			id := comp.ChildCompsIdGen[childComp.Name].next()

			comp.writeGo("{\n")
			if len(forKeys) > 0 {
				comp.writeGo("key := key")
				for _, key := range forKeys {
					comp.writeGo(` + "-" + fmt.Sprint(` + key + ")")
				}
				comp.writeGo("\n")
			}
			comp.writeGo(fmt.Sprintf("ckey := key + \"%s\"\n", id))
			comp.writeGo(fmt.Sprintf("state := &state_%s{}\n", childComp.Id()))
			comp.writeGo("if _, ok := states[ckey]; ok {\n")
			comp.writeGo("json.Unmarshal([]byte(states[ckey]), state)\n")
			comp.writeGo("newStates[ckey] = states[ckey]")
			comp.writeGo("} else {\n")
			for _, varName := range childComp.VarNames {
				v := childComp.Vars[varName]
				comp.writeGo(fmt.Sprintf("state.%s", v.StructField))
				if val, found := hasAttr(node, varName); found {
					comp.writeGo(fmt.Sprintf(" = %s\n", val))
				} else if v.InitExpr != nil {
					comp.writeGo(fmt.Sprintf(" = %s\n", astToSource(v.InitExpr)))
				} else {
					comp.writeGo("\n")
				}
			}
			comp.writeGo("txStateBytes, _ := json.Marshal(state)\n")
			comp.writeGo("newStates[ckey] = string(txStateBytes)\n")
			comp.writeGo("}\n")
			states := []string{}
			for _, varName := range childComp.VarNames {
				states = append(states, "state."+childComp.Vars[varName].StructField)
			}
			comp.writeGo(fmt.Sprintf("render_%s(w, ckey, states, newStates, %s", childComp.Id(), strings.Join(states, ",")))

			slotNodes := map[string]*html.Node{}

			for c := node.FirstChild; c != nil; c = c.NextSibling {
				if slotName, found := hasAttr(c, "slot"); found {
					slotNodes[slotName] = c
					continue
				} else {
					if slotNodes[""] == nil {

						slotNodes[""] = &html.Node{
							Type:     html.ElementNode,
							DataAtom: atom.Template,
							Data:     "template",
						}
					}

					slotNodes[""].AppendChild(&html.Node{
						FirstChild: c.FirstChild,
						LastChild:  c.LastChild,

						Type:      c.Type,
						DataAtom:  c.DataAtom,
						Data:      c.Data,
						Namespace: c.Namespace,
						Attr:      c.Attr,
					})
				}
			}

			if len(childComp.SlotNames) > 0 {
				comp.writeGo(",\n")
			}
			for _, slotName := range childComp.SlotNames {
				n, ok := slotNodes[slotName]
				if ok {
					comp.writeGo("func() {\n")
					ptErrs := comp.parseTmpl(n, forKeys)
					errs.append(ptErrs.Errs...)
					comp.writeGo("\n},\n")
				} else {
					comp.writeGo("nil,\n")
				}
			}
			comp.writeGo(")\n")
			comp.writeGo("}\n")
			return errs
		}

		if node.DataAtom == atom.Slot {
			renderSlotFuncName := "render_default_slot"
			if name, found := hasAttr(node, "name"); found {
				renderSlotFuncName = "render_slot_" + name
			}

			comp.writeGo(fmt.Sprintf("if %s != nil {\n", renderSlotFuncName))
			comp.writeGo(fmt.Sprintf("%s()\n", renderSlotFuncName))
			comp.writeGo("} else {\n")

			if node.FirstChild != nil {
				children := &html.Node{
					Type:     html.ElementNode,
					DataAtom: atom.Template,
					Data:     "template",
				}
				for c := node.FirstChild; c != nil; c = c.NextSibling {
					children.AppendChild(&html.Node{
						FirstChild: c.FirstChild,
						LastChild:  c.LastChild,

						Type:      c.Type,
						DataAtom:  c.DataAtom,
						Namespace: c.Namespace,
						Data:      c.Data,
						Attr:      c.Attr,
					})
				}
				ptErrs := comp.parseTmpl(children, forKeys)
				errs.append(ptErrs.Errs...)
			} else {
				comp.writeStrLit(" ")
			}

			comp.writeGo("\n}\n")
			return errs
		}

		if node.DataAtom != atom.Template {
			comp.writeStrLit("<")
			comp.writeStrLit(node.Data)

			_, isIgnore := hasAttr(node, txIgnoreKey)

			for _, attr := range node.Attr {
				if attr.Key == txIfKey || attr.Key == txElseIfKey || attr.Key == txElseKey || attr.Key == txForKey {
					continue
				}

				comp.writeStrLit(" ")
				if attr.Namespace != "" {
					comp.writeStrLit(node.Namespace)
					comp.writeStrLit(":")
				}
				comp.writeStrLit(attr.Key)
				comp.writeStrLit(`="`)

				if strings.HasPrefix(attr.Key, "tx-on") {
					if expr, err := parser.ParseExpr(attr.Val); err == nil {
						if callExpr, ok := expr.(*ast.CallExpr); ok {
							if ident, ok := callExpr.Fun.(*ast.Ident); ok {
								if fun, ok := comp.Funcs[ident.Name]; ok {
									params := []string{}
									for _, list := range fun.Decl.Type.Params.List {
										for _, ident := range list.Names {
											params = append(params, ident.Name)
										}
									}

									if len(params) != len(callExpr.Args) {
										errs.append(comp.errf("params length not match: %s", astToSource(callExpr)))
										continue
									}

									comp.writeStrLit(comp.funcId(fun.Name))
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

											if _, ok := comp.Vars[ident.Name]; ok {
												foundVar = true
												return false
											}

											return true
										})

										if foundVar {
											errs.append(comp.errf("state and derived variables cannot be used as function parameters: %s", callExpr.Args[i]))
											continue
										}

										if i == 0 {
											comp.writeStrLit("?" + param + "=")
										} else {
											comp.writeStrLit("&" + param + "=")
										}

										arg := astToSource(callExpr.Args[i])
										comp.writeUrlEscapeExpr(arg)
									}
									comp.writeStrLit(`"`)
									continue
								}
							}
						}
					}

					funcName := comp.FuncNameGen.next()
					fileAst, err := parser.ParseFile(token.NewFileSet(), comp.FilePath, fmt.Sprintf("package p\nfunc %s() {\n%s\n}", funcName, attr.Val), 0)
					if err != nil {
						errs.append(comp.errf("parse inline statement failed: %s", attr.Val))
						continue
					}

					decl, ok := fileAst.Decls[0].(*ast.FuncDecl)
					if !ok {
						errs.append(comp.errf("parse inline statement failed: %s", attr.Val))
						continue
					}

					modifiedDerived := []string{}
					comp.modifiedDerived(decl, &modifiedDerived)
					if len(modifiedDerived) > 0 {
						errs.append(comp.errf("derived cannot be modified: %v", modifiedDerived))
						continue
					}

					comp.FuncNames = append(comp.FuncNames, decl.Name.Name)
					comp.Funcs[decl.Name.Name] = &Func{
						Name: decl.Name.Name,
						Decl: decl,
					}

					comp.writeStrLit(comp.funcId(decl.Name.Name))

				} else if isIgnore {
					comp.writeStrLit(attr.Val)
				} else if err := comp.parseTmplStr(attr.Val, false); err != nil {
					errs.append(comp.errf("parse attr value failed: %s", attr.Val))
				}
				comp.writeStrLit(`"`)
			}

			// https://html.spec.whatwg.org/#void-elements
			if isVoidElement(node.Data) {
				if node.FirstChild != nil {
					errs.append(comp.errf("invalid void elements: %s" + node.Data))
				}

				comp.writeStrLit("/>")
				return errs
			}

			comp.writeStrLit(">")
		}

		// 0: no control flow
		// 1: if
		// 2: else-if
		// 3: else
		if val, found := hasAttr(node, "id"); node.DataAtom == atom.Script && found && val == txRuntimeVal {
			comp.writeGo("w.Write([]byte(runtimeScript))\n")
		} else {

			var prevCondState CondState
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				hasFor := false
				forKey := ""
				if c.Type == html.ElementNode {
					currCondState, field := condState(c)

					switch prevCondState {
					case CondStateDefault:
						if currCondState == CondStateElseIf || currCondState == CondStateElse {
							errs.append(comp.errf("detect tx-else-if or tx-else right after non-cond node: %s", c.Data))
						}
					case CondStateIf:
						if currCondState <= prevCondState {
							comp.writeGo("\n}\n")
						}
					case CondStateElseIf:
						if currCondState < prevCondState {
							comp.writeGo("\n}\n")
						}
					case CondStateElse:
						if currCondState == CondStateElseIf || currCondState == CondStateElse {
							errs.append(comp.errf("detect tx-else-if or tx-else right after tx-else: %s", c.Data))
						}
						comp.writeGo("\n}\n")
					}

					switch currCondState {
					case CondStateDefault:
					case CondStateIf:
						comp.writeGo("\nif " + field + " {\n")
					case CondStateElseIf:
						comp.writeGo("\n} else if " + field + " {\n")
					case CondStateElse:
						comp.writeGo("\n} else {\n")
					}

					prevCondState = currCondState

					if stmt, ok := hasAttr(c, txForKey); ok {
						val, found := hasAttr(c, txKeyKey)
						if !found {
							errs.append(comp.errf("tx-for loop must have tx-key attr"))
						} else {
							hasFor = true
							forKey = val
							comp.writeGo("\nfor " + stmt + " {\n")
						}
					}
				}

				childForKeys := forKeys
				if hasFor {
					childForKeys = append(forKeys, forKey)
				}

				if childErrs := comp.parseTmpl(c, childForKeys); childErrs != nil {
					errs.append(childErrs.Errs...)
				}

				if hasFor {
					comp.writeGo("\n}\n")
				}

				if c.NextSibling == nil && (prevCondState == CondStateIf || prevCondState == CondStateElseIf || prevCondState == CondStateElse) {
					comp.writeGo("\n}\n")
				}
			}
		}

		if node.DataAtom != atom.Template {
			comp.writeStrLit("</")
			comp.writeStrLit(node.Data)
			comp.writeStrLit(">")
		}
		return errs

	}

	return nil
}

func (comp *Component) parseTmplStr(str string, escape bool) error {
	str = strings.TrimSpace(str)
	if len(str) == 0 {
		str = " "
	}
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
				comp.writeStrLit(html.EscapeString(string(r)))
			} else {
				comp.writeStrLit(string(r))
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
					return comp.errf("parse expression error: %s: %w", string(trimmedCurrExpr), err)
				}

				if escape {
					comp.writeHtmlEscapeExpr(string(trimmedCurrExpr))
				} else {
					comp.writeExpr(string(trimmedCurrExpr))
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
		return comp.errf("unclosed quote in expression: %s", str)
	}
	if braceStack != 0 {
		return comp.errf("unclosed brace in expression: %s", str)
	}

	return nil
}

func (comp *Component) parseSlots(node *html.Node, inSlot bool) *Errors {
	errs := newErrors()
	if node.Type != html.ElementNode {
		return errs
	}

	isSlot := node.DataAtom == atom.Slot
	if isSlot {
		if inSlot {
			errs.append(comp.errf("no nested slots"))
		}

		slotName := ""
		if name, found := hasAttr(node, "name"); found {
			slotName = name
		}

		if _, ok := comp.Slots[slotName]; !ok {
			comp.SlotNames = append(comp.SlotNames, slotName)
			comp.Slots[slotName] = struct{}{}
		} else {
			if slotName == "" {
				errs.append(comp.errf("multiple <slot>"))
			} else {
				errs.append(comp.errf("multiple <slot name=\"%s\" >", slotName))
			}
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		psErrs := comp.parseSlots(c, isSlot)
		errs.append(psErrs.Errs...)
	}

	return errs
}

func (comp *Component) implRenderFunc(out *strings.Builder) {
	for _, tmpl := range comp.RenderFuncCodes {
		switch tmpl.Type {
		case RenderFuncTypeGo:
			if _, err := out.WriteString(string(tmpl.Content)); err != nil {
				log.Fatalln(err)
			}
		case RenderFuncTypeStrLit:
			if _, err := out.WriteString(fmt.Sprintf("w.Write([]byte(`%s`))\n", string(tmpl.Content))); err != nil {
				log.Fatalln(err)
			}
		case RenderFuncTypeExpr:
			if _, err := out.WriteString(fmt.Sprintf("w.Write([]byte(fmt.Sprint(%s)))\n", string(tmpl.Content))); err != nil {
				log.Fatalln(err)
			}
		case RenderFuncTypeHtmlEscapeExpr:
			if _, err := out.WriteString(fmt.Sprintf("w.Write([]byte(html.EscapeString(fmt.Sprint(%s))))\n", string(tmpl.Content))); err != nil {
				log.Fatalln(err)
			}
		case RenderFuncTypeUrlEscapeExpr:
			if _, err := out.WriteString(fmt.Sprintf("if param, err := json.Marshal(%s); err != nil {\nlog.Panic(err)\n} else {\nw.Write([]byte(url.QueryEscape(string(param))))}\n", string(tmpl.Content))); err != nil {
				log.Fatalln(err)
			}
		}
	}
}

type VarType int

const (
	VarTypeState = iota + 1
	VarTypeDerived
)

type Var struct {
	Name        string
	StructField string
	Type        VarType
	TypeExpr    ast.Expr
	InitExpr    ast.Expr
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

type RenderFuncType int

const (
	RenderFuncTypeGo RenderFuncType = iota + 1
	RenderFuncTypeStrLit
	RenderFuncTypeExpr
	RenderFuncTypeHtmlEscapeExpr
	RenderFuncTypeUrlEscapeExpr
	RenderFuncTypeComp
)

type RenderFunc struct {
	Type    RenderFuncType
	Content []byte
}

func (comp *Component) writeTmpl(t RenderFuncType, content string) {
	if comp.CurrRenderFuncType != t {
		if len(comp.CurrRenderFuncContent) != 0 {
			comp.RenderFuncCodes = append(comp.RenderFuncCodes, RenderFunc{
				Type:    comp.CurrRenderFuncType,
				Content: comp.CurrRenderFuncContent,
			})
		}

		comp.CurrRenderFuncType = t
		comp.CurrRenderFuncContent = []byte{}
	}

	comp.CurrRenderFuncContent = append(comp.CurrRenderFuncContent, content...)
}

func (comp *Component) writeGo(content string) {
	comp.writeTmpl(RenderFuncTypeGo, content)
}

func (comp *Component) writeStrLit(content string) {
	comp.writeTmpl(RenderFuncTypeStrLit, content)
}

func (comp *Component) writeExpr(content string) {
	comp.writeTmpl(RenderFuncTypeExpr, content)
}

func (comp *Component) writeHtmlEscapeExpr(content string) {
	comp.writeTmpl(RenderFuncTypeHtmlEscapeExpr, content)
}

func (comp *Component) writeUrlEscapeExpr(content string) {
	comp.writeTmpl(RenderFuncTypeUrlEscapeExpr, content)
}

func (comp *Component) Id() string {
	p, _ := strings.CutSuffix(comp.Name, filepath.Ext(comp.Name))
	p = strings.ReplaceAll(p, "/", "_")
	p = strings.ReplaceAll(p, "-", "_d_")
	p = strings.ReplaceAll(p, "{", "_lb_")
	p = strings.ReplaceAll(p, "}", "_rb_")
	return p
}

func (comp *Component) funcId(funcName string) string {
	return comp.Id() + "_" + funcName
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
		if attr.Key == txIfKey {
			return CondStateIf, attr.Val
		}

		if attr.Key == txElseIfKey {
			return CondStateElseIf, attr.Val
		}

		if attr.Key == txElseKey {
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

func (comp *Component) urlPath() string {
	dir, file := filepath.Split(comp.RelPath)
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

func (comp *Component) funcHandlerFields() []HandlerFields {
	var codeVar strings.Builder
	codeVar.WriteString("query := r.URL.Query()\n")

	for _, name := range comp.VarNames {
		v := comp.Vars[name]
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
	comp.writeDerivedAst(&codeDerived)

	handlers := []HandlerFields{}
	for _, funcName := range comp.FuncNames {
		var code strings.Builder
		code.WriteString(codeVar.String())

		fun := comp.Funcs[funcName]

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
		comp.writeStateFieldsAst(&state)

		params := []string{}
		for _, varName := range comp.VarNames {
			params = append(params, varName)
		}
		render := fmt.Sprintf("render_%s(w, state, %s)\n", comp.Id(), strings.Join(params, ", "))

		handlers = append(handlers, HandlerFields{
			Url:    "/tx/" + comp.funcId(funcName),
			Code:   code.String(),
			State:  state.String(),
			Render: render,
		})
	}

	return handlers
}

func (comp *Component) writeStateFieldsAst(sb *strings.Builder) {
	stateAst := &ast.CompositeLit{
		Type: &ast.MapType{
			Key:   &ast.Ident{Name: "string"},
			Value: &ast.Ident{Name: "any"},
		},
	}

	for _, varName := range comp.VarNames {
		name := comp.Vars[varName].Name
		stateAst.Elts = append(stateAst.Elts, &ast.KeyValueExpr{
			Key:   &ast.BasicLit{Kind: token.STRING, Value: `"` + name + `"`},
			Value: &ast.Ident{Name: name},
		})
	}

	sb.WriteString(astToSource(stateAst))
}

func (comp *Component) writeDerivedAst(sb *strings.Builder) {
	for _, name := range comp.VarNames {
		v := comp.Vars[name]
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

func newErrors(errs ...error) *Errors {
	Errs := &Errors{
		Errs: []error{},
	}
	for _, err := range errs {
		if err != nil {
			Errs.Errs = append(Errs.Errs, err)
		}
	}
	return Errs
}

type Errors struct {
	Errs []error
	mux  sync.Mutex
}

func (es *Errors) append(errs ...error) {
	es.mux.Lock()
	for _, err := range errs {
		if err != nil {
			es.Errs = append(es.Errs, err)
		}
	}
	es.mux.Unlock()
}

type IdGen struct {
	Curr   int
	Prefix string
}

func (id *IdGen) curr() string {
	return fmt.Sprintf("%s_%d", id.Prefix, id.Curr)
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
