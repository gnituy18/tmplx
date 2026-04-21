package main

import (
	"bytes"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/scanner"
	"go/token"
	"io/fs"
	"log"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/imports"
)

var (
	pagesDir                 string
	componentsDir            string
	outputFilePath           string
	outputPackageName        string
	outputEventHandlerPrefix string

	componentsByName = map[string]*Component{}
)

func main() {
	// 0. configure logging, find module root, parse CLI flags
	log.SetFlags(0)

	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			log.Fatalln("error: no go.mod found in current or parent directories")
		}
		dir = parent
	}

	flag.StringVar(&componentsDir, "components-dir", filepath.Join(dir, "components"), "directory containing reusable components")
	flag.StringVar(&pagesDir, "pages-dir", filepath.Join(dir, "pages"), "directory containing pages")
	flag.StringVar(&outputFilePath, "output-file", filepath.Join(dir, "routes.go"), "path to the generated Go file")
	flag.StringVar(&outputPackageName, "package-name", "main", "package name for the generated Go code")
	flag.StringVar(&outputEventHandlerPrefix, "handler-prefix", "/tx/", "path prefix for event handler URLs")
	flag.Parse()
	componentsDir = filepath.Clean(componentsDir)
	pagesDir = filepath.Clean(pagesDir)
	if !token.IsIdentifier(outputPackageName) || token.IsKeyword(outputPackageName) {
		log.Fatalf("\"%s\" is not a valid Go package name\n", outputPackageName)
	}
	outputFilePath = filepath.Clean(outputFilePath)

	// 1. register component and page HTML files
	merr := newMultiError()
	if exist, err := dirExist(componentsDir); err != nil {
		log.Fatalf("error: %v\n", err)

	} else if !exist {
		log.Printf("no components directory at %s, skipping\n", componentsDir)

	} else if err := filepath.WalkDir(componentsDir, func(filePath string, entry fs.DirEntry, err error) error {
		if err != nil {
			merr.append(fmt.Errorf("%s: cannot access: %w", filePath, err))
			return nil
		}

		if entry.IsDir() {
			return nil
		}

		if filepath.Ext(filePath) != ".html" {
			return nil
		}

		relPath, _ := filepath.Rel(componentsDir, filePath)
		relPath = filepath.ToSlash(relPath)
		stemPath, _ := strings.CutSuffix(relPath, ".html")
		if stemPath == "" {
			merr.append(fmt.Errorf("%s: invalid filename: .html (missing name before extension)", filePath))
			return nil
		}
		name := "tx-" + strings.ReplaceAll(stemPath, "/", "-")
		for _, r := range name {
			if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
				merr.append(fmt.Errorf("%s: invalid character \"%s\" in <%s>: use only a-z, 0-9, -, _", filePath, string(r), name))
				return nil
			}
		}

		if comp, ok := componentsByName[name]; ok {
			merr.append(fmt.Errorf("%s: duplicate component <%s>, first defined in %s", filePath, name, comp.FilePath))
			return nil
		}

		componentsByName[name] = &Component{
			Type:     CompTypeComp,
			FilePath: filePath,
			RelPath:  relPath,
			Name:     name,
			GoName:   goIdent(name),
		}

		return nil

	}); err != nil {
		log.Fatalf("error: %s: walk failed: %v\n", componentsDir, err)
	}

	pages := []*Component{}
	pageFiles := map[string]string{}
	if exist, err := dirExist(pagesDir); err != nil {
		log.Fatalf("error: %s: cannot access pages directory: %v\n", pagesDir, err)
	} else if !exist {
		log.Fatalf("pages directory not found: %s\n", pagesDir)
	} else if err := filepath.WalkDir(pagesDir, func(filePath string, entry fs.DirEntry, err error) error {
		if err != nil {
			merr.append(fmt.Errorf("%s: cannot access: %w", filePath, err))
			return nil
		}

		if entry.IsDir() {
			return nil
		}

		if filepath.Ext(filePath) != ".html" {
			return nil
		}

		relPath, _ := filepath.Rel(pagesDir, filePath)
		relPath = filepath.ToSlash(relPath)

		urlDir, _ := strings.CutSuffix(relPath, entry.Name())
		baseName, _ := strings.CutSuffix(entry.Name(), ".html")
		if baseName == "" {
			merr.append(fmt.Errorf("%s: invalid filename: .html (missing name before extension)", filePath))
			return nil
		}

		urlPath := "/" + urlDir
		if baseName != "index" {
			urlPath += baseName
		}

		if strings.HasSuffix(urlPath, "/") {
			urlPath += "{$}"
		}

		if existingFile, ok := pageFiles[urlPath]; ok {
			merr.append(fmt.Errorf("%s: duplicate page route %s, first defined in %s", filePath, urlPath, existingFile))
			return nil
		}

		pageFiles[urlPath] = filePath
		pages = append(pages, &Component{
			Type:     CompTypePage,
			FilePath: filePath,
			RelPath:  relPath,
			Name:     urlPath,
			GoName:   goIdent(urlPath),
		})

		return nil

	}); err != nil {
		log.Fatalf("error: %s: walk failed: %v\n", pagesDir, err)
	}
	merr.exitOnErrors()

	// 2. parse component and page script and slot
	var wg sync.WaitGroup
	components := slices.SortedFunc(maps.Values(componentsByName), func(a, b *Component) int {
		return strings.Compare(a.Name, b.Name)
	})
	for _, comp := range components {
		wg.Add(1)
		go func() {
			defer wg.Done()

			file, err := os.Open(comp.FilePath)
			if err != nil {
				merr.append(comp.errf("cannot open file: %w", err))
				return
			}
			defer file.Close()

			nodes, err := html.ParseFragment(file, &html.Node{
				Data:     "body",
				DataAtom: atom.Body,
				Type:     html.ElementNode,
			})
			if err != nil {
				merr.append(comp.errf("invalid HTML: %w", err))
				return
			}

			comp.TemplateNode = newTemplateNode()
			for _, node := range nodes {
				val, found := hasAttr(node, "type")
				if node.DataAtom == atom.Script && found && val == "text/tmplx" {
					if comp.TmplxScriptNode != nil {
						merr.append(comp.errf("multiple <script type=\"text/tmplx\"> elements (only one allowed)"))
						return
					}
					comp.TmplxScriptNode = node
				} else if node.DataAtom == atom.Style {
					if comp.StyleNode != nil {
						merr.append(comp.errf("multiple <style> elements (only one allowed)"))
						return
					}
					comp.StyleNode = node
				} else {
					comp.TemplateNode.AppendChild(node)
				}
			}

			merr.concat(comp.parseTmplxScript())
			merr.concat(comp.parseSlots(comp.TemplateNode, false))
		}()
	}

	for _, page := range pages {
		wg.Add(1)
		go func() {
			defer wg.Done()

			file, err := os.Open(page.FilePath)
			if err != nil {
				merr.append(page.errf("cannot open file: %w", err))
				return
			}
			defer file.Close()

			page.TemplateNode, err = html.Parse(file)
			if err != nil {
				merr.append(page.errf("invalid HTML: %w", err))
				return
			}

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
				merr.append(page.errf("page must have a <head> element (required for state and runtime script injection)"))
				return
			}

			cleanUpTmplxScript(page.TemplateNode)

			merr.concat(page.parseTmplxScript())
		}()
	}
	wg.Wait()
	merr.exitOnErrors()

	// 3. parse used vars has child comps
	for _, comp := range slices.Concat(components, pages) {
		wg.Add(1)
		go func() {
			defer wg.Done()

			comp.ChildCompsIdGen = map[string]*IdGen{}
			for _, c := range components {
				comp.ChildCompsIdGen[c.Name] = newIdGen()
			}
			comp.UsedVars = map[string]struct{}{}
			comp.FillByGoName = map[string]*Fill{}
			merr.concat(comp.parseUsedVars(comp.TemplateNode))

			derivedDeps := map[string]struct{}{}
			for _, v := range comp.Vars {
				if v.Type == VarTypeDerived {
					comp.scanVarRefs(v.InitExprAst, derivedDeps)
				}
			}
			for _, v := range comp.Vars {
				_, used := comp.UsedVars[v.GoName]
				_, dep := derivedDeps[v.GoName]
				if !used && !dep {
					merr.append(comp.errf("%s declared but not used", v.GoName))
				}
			}
		}()
	}
	wg.Wait()
	merr.exitOnErrors()
	for _, comp := range components {
		slices.SortFunc(comp.CompFills, func(a, b *Fill) int {
			return strings.Compare(a.GoName, b.GoName)
		})
		for _, fill := range comp.CompFills {
			if fill.HasChildComps {
				comp.CompFillsHasChildComps = true
				break
			}
		}
	}

	// 4. parse pages and components template
	for _, comp := range components {
		wg.Add(1)
		go func() {
			defer wg.Done()

			comp.ChildCompsIdGen = map[string]*IdGen{}
			for _, c := range components {
				comp.ChildCompsIdGen[c.Name] = newIdGen()
			}
			comp.AnonFuncNameGen = newIdGen()
			comp.RenderFunc = newCode("tx_w")

			comp.RenderFunc = newCode("tx_w")
			comp.RenderFunc.emitStrLit("<!--tx:")
			comp.RenderFunc.emitExpr("tx_id")
			comp.RenderFunc.emitStrLit("-->")
			merr.concat(comp.parseTmpl(comp.TemplateNode, []string{}, false))
			comp.RenderFunc.emitStrLit("<!--tx:")
			comp.RenderFunc.emitExpr("tx_id + \"_e\"")
			comp.RenderFunc.emitStrLit("-->")
		}()
	}

	for _, page := range pages {
		wg.Add(1)
		go func() {
			defer wg.Done()

			page.ChildCompsIdGen = map[string]*IdGen{}
			for _, c := range components {
				page.ChildCompsIdGen[c.Name] = newIdGen()
			}
			page.AnonFuncNameGen = newIdGen()
			page.RenderFunc = newCode("tx_w1")
			merr.concat(page.parseTmpl(page.TemplateNode, []string{}, false))
		}()
	}
	wg.Wait()
	merr.exitOnErrors()

	// 5. generate and write the output Go file
	var code CodeBuilder

	code.write("package %s\n", outputPackageName)

	code.write("import(\n")
	for _, page := range pages {
		for _, im := range page.Imports {
			code.write("%s\n", astToSource(im))
		}
	}
	for _, comp := range components {
		for _, im := range comp.Imports {
			code.write("%s\n", astToSource(im))
		}
	}
	code.write(")\n")

	code.write("var runtimeScript = `%s`\n", strings.Replace(runtimeScript, "TX_HANDLER_PREFIX", outputEventHandlerPrefix, 1))

	for _, comp := range components {
		code.write("type %s struct {\n", comp.GoName)
		for _, v := range comp.Vars {
			if v.Type == VarTypeState || v.Type == VarTypeProp {
				code.write("%s %s `json:\"%s\"`\n", v.SavedField, v.TypeExpr, v.GoName)
			}
		}
		code.write("}\n")

		code.write("func render_%s(tx_w *bytes.Buffer, tx_id string", comp.GoName)
		if len(comp.Slots) > 0 {
			code.write(", tx_pid, tx_loc string")
		}
		if comp.HasChildComps {
			code.write(", tx_curr_saved map[string]string, tx_next_saved map[string]any")
		}
		for _, v := range comp.Vars {
			if _, ok := comp.UsedVars[v.GoName]; ok {
				code.write(", %s %s", v.GoName, v.TypeExpr)
			}
		}
		for _, f := range comp.Funcs {
			code.write(", %s, %s_swap string", f.Name, f.Name)
		}
		for _, slotName := range comp.Slots {
			code.write(", tx_render_fill_%s func()", slotName)
		}
		code.write(") {\n")
		comp.RenderFunc.writeTo(&code)
		code.write("}\n")
		for _, fill := range comp.Fills {
			code.write("func render_fill_%s(tx_w *bytes.Buffer", fill.GoName)
			if fill.HasChildComps {
				code.write(", tx_id string, tx_curr_saved map[string]string, tx_next_saved map[string]any")
			}
			for _, v := range comp.Vars {
				if _, ok := fill.UsedVars[v.GoName]; ok {
					code.write(", %s %s", v.GoName, v.TypeExpr)
				}
			}
			code.write(") {\n")
			fill.RenderFunc.writeTo(&code)
			code.write("}\n")
		}

		if len(comp.CompFills) > 0 {
			code.write("func render_comp_fill_%s(tx_w *bytes.Buffer, tx_loc string, tx_id string, tx_curr_saved map[string]string", comp.GoName)
			if comp.CompFillsHasChildComps {
				code.write(", tx_next_saved map[string]any")
			}
			code.write(") {\n")
			code.write("switch tx_loc {\n")
			for _, fill := range comp.CompFills {
				code.write("case \"%s\":\n", fill.Location)
				code.write("tx_saved := &%s{}\n", fill.ParentComp.GoName)
				code.write("json.Unmarshal([]byte(tx_curr_saved[tx_id]), tx_saved)\n")
				for _, v := range fill.ParentComp.Vars {
					if v.Type == VarTypeDerived {
						if _, ok := fill.UsedVars[v.GoName]; ok {
							code.write("tx_derived_%s := %s\n", v.GoName, v.InitExpr)
						}
					}
				}
				code.write("render_fill_%s(tx_w", fill.GoName)
				if fill.HasChildComps {
					code.write(", tx_id, tx_curr_saved, tx_next_saved")
				}
				for _, v := range fill.ParentComp.Vars {
					if _, ok := fill.UsedVars[v.GoName]; ok {
						switch v.Type {
						case VarTypeState, VarTypeProp:
							code.write(", tx_saved.%s", v.SavedField)
						case VarTypeDerived:
							code.write(", tx_derived_%s", v.GoName)
						}
					}
				}
				code.write(")\n")
			}
			code.write("}\n}\n")
		}
	}

	for _, page := range pages {
		code.write("type %s struct {\n", page.GoName)
		for _, v := range page.Vars {
			if v.Type == VarTypeState {
				code.write("%s %s `json:\"%s\"`\n", v.SavedField, v.TypeExpr, v.GoName)
			}
		}
		code.write("}\n")

		code.write("func render_%s(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer", page.GoName)
		if page.HasChildComps {
			code.write(", tx_curr_saved map[string]string, tx_next_saved map[string]any")
		}
		for _, v := range page.Vars {
			if _, ok := page.UsedVars[v.GoName]; ok {
				code.write(", %s %s", v.GoName, v.TypeExpr)
			}
		}
		for _, f := range page.Funcs {
			code.write(", %s string", f.Name)
		}
		code.write(") {\n")
		page.RenderFunc.writeTo(&code)
		code.write("}\n")
		for _, fill := range page.Fills {
			code.write("func render_fill_%s(tx_w *bytes.Buffer", fill.GoName)
			if fill.HasChildComps {
				code.write(", tx_id string, tx_curr_saved map[string]string, tx_next_saved map[string]any")
			}
			for _, v := range page.Vars {
				if _, ok := fill.UsedVars[v.GoName]; ok {
					code.write(", %s %s", v.GoName, v.TypeExpr)
				}
			}
			code.write(") {\n")
			fill.RenderFunc.writeTo(&code)
			code.write("}\n")
		}
	}

	code.write("type TxRoute struct {\n")
	code.write("Pattern	string\n")
	code.write("Handler	http.HandlerFunc\n")
	code.write("}\n")

	code.write("var txRoutes []TxRoute = []TxRoute{\n")
	for _, page := range pages {
		code.write("{\n")
		code.write("Pattern: \"GET %s\",\n", page.Name)
		code.write("Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {\n")
		code.write("tx_saved := &%s{}\n", page.GoName)
		for _, v := range page.Vars {
			if v.Type == VarTypeState && v.InitExpr != "" {
				code.write("tx_saved.%s = %s\n", v.SavedField, v.InitExpr)
			}
		}
		for _, v := range page.Vars {
			if v.Type == VarTypeDerived {
				code.write("tx_derived_%s := %s\n", v.GoName, v.InitExpr)
			}
		}
		if page.InitFunc != nil {
			code.write("%s", page.InitFunc.Stmts)
		}

		code.write("tx_next_saved := map[string]any{\"page\": tx_saved}\n")
		code.write("var tx_buf1, tx_buf2 bytes.Buffer\n")
		callParams := []string{"&tx_buf1", "&tx_buf2"}
		if page.HasChildComps {
			callParams = append(callParams, "map[string]string{}", "tx_next_saved")
		}
		for _, v := range page.Vars {
			if _, ok := page.UsedVars[v.GoName]; ok {
				switch v.Type {
				case VarTypeState:
					callParams = append(callParams, "tx_saved."+v.SavedField)
				case VarTypeDerived:
					callParams = append(callParams, "tx_derived_"+v.GoName)
				}
			}
		}
		for _, f := range page.Funcs {
			callParams = append(callParams, fmt.Sprintf("\"%s\"", url.PathEscape(page.Name)+":"+f.Name))
		}
		code.write("render_%s(%s)\n", page.GoName, strings.Join(callParams, ", "))
		code.write("tx_savedBytes, _ := json.Marshal(tx_next_saved)\n")
		code.write("tx_w.Write(tx_buf1.Bytes())\n")
		code.write("tx_w.Write(tx_savedBytes)\n")
		code.write("tx_w.Write(tx_buf2.Bytes())\n")
		code.write("},\n")
		code.write("},\n")

		pageFuncs := append(page.Funcs, page.AnonFuncs...)
		for _, f := range pageFuncs {
			code.write("{\n")
			code.write("Pattern: \"POST %s%s:%s\",\n", outputEventHandlerPrefix, url.PathEscape(page.Name), f.Name)
			code.write("Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {\n")
			code.write("tx_r.ParseForm()\n")
			code.write("tx_curr_saved := map[string]string{}\n")
			code.write("for k, v := range tx_r.PostForm {\n")
			code.write("tx_curr_saved[k] = v[0]\n")
			code.write("}\n")
			code.write("tx_saved := &%s{}\n", page.GoName)
			code.write("json.Unmarshal([]byte(tx_curr_saved[\"page\"]), &tx_saved)\n")
			for _, v := range page.Vars {
				if v.Type == VarTypeDerived {
					code.write("tx_derived_%s := %s\n", v.GoName, v.InitExpr)
				}
			}
			for _, list := range f.Decl.Type.Params.List {
				for _, ident := range list.Names {
					code.write("var %s %s\n", ident.Name, astToSource(list.Type))
					code.write("json.Unmarshal([]byte(tx_r.PostFormValue(\"%s\")), &%s)\n", ident.Name, ident.Name)
				}
			}
			code.write("%s", f.Stmts)
			code.write("tx_next_saved := map[string]any{\"page\": tx_saved}\n")
			code.write("var tx_buf1, tx_buf2 bytes.Buffer\n")
			callParams := []string{"&tx_buf1", "&tx_buf2"}
			if page.HasChildComps {
				callParams = append(callParams, "tx_curr_saved", "tx_next_saved")
			}
			for _, v := range page.Vars {
				if _, ok := page.UsedVars[v.GoName]; ok {
					switch v.Type {
					case VarTypeState:
						callParams = append(callParams, "tx_saved."+v.SavedField)
					case VarTypeDerived:
						callParams = append(callParams, "tx_derived_"+v.GoName)
					}
				}
			}
			for _, f := range page.Funcs {
				callParams = append(callParams, fmt.Sprintf("\"%s\"", url.PathEscape(page.Name)+":"+f.Name))
			}
			code.write("render_%s(%s)\n", page.GoName, strings.Join(callParams, ", "))
			code.write("tx_savedBytes, _ := json.Marshal(tx_next_saved)\n")
			code.write("tx_w.Write(tx_buf1.Bytes())\n")
			code.write("tx_w.Write(tx_savedBytes)\n")
			code.write("tx_w.Write(tx_buf2.Bytes())\n")
			code.write("},\n")
			code.write("},\n")
		}
	}
	for _, comp := range components {
		compFuncs := append(comp.Funcs, comp.AnonFuncs...)
		for _, f := range compFuncs {
			if f.Decl.Body == nil {
				return
			}
			code.write("{\n")
			code.write("Pattern: \"POST %s%s:%s\",\n", outputEventHandlerPrefix, comp.Name, f.Name)
			code.write("Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {\n")
			code.write("tx_r.ParseForm()\n")
			code.write("tx_id := tx_r.PostFormValue(\"tx-swap\")\n")
			if len(comp.Slots) > 0 {
				code.write("tx_pid := tx_r.PostFormValue(\"tx-pid\")\n")
				code.write("tx_loc := tx_r.PostFormValue(\"tx-loc\")\n")
			}
			code.write("tx_curr_saved := map[string]string{}\n")
			code.write("for k, v := range tx_r.PostForm {\n")
			if len(comp.Slots) > 0 {
				code.write("if k != \"tx-swap\" && k != \"tx-loc\" && k != \"tx-pid\" {\n")
			} else {
				code.write("if k != \"tx-swap\" {\n")
			}
			code.write("tx_curr_saved[k] = v[0]\n")
			code.write("}\n")
			code.write("}\n")
			code.write("tx_next_saved := map[string]any{}\n")
			code.write("tx_saved := &%s{}\n", comp.GoName)
			code.write("json.Unmarshal([]byte(tx_curr_saved[tx_id]), &tx_saved)\n")
			for _, v := range comp.Vars {
				if v.Type == VarTypeDerived {
					code.write("tx_derived_%s := %s\n", v.GoName, v.InitExpr)
				}
			}
			for _, list := range f.Decl.Type.Params.List {
				for _, ident := range list.Names {
					code.write("var %s %s\n", ident.Name, astToSource(list.Type))
					code.write("json.Unmarshal([]byte(tx_r.PostFormValue(\"%s\")), &%s)\n", ident.Name, ident.Name)
				}
			}
			code.write("%s", f.Stmts)
			code.write("tx_next_saved[tx_id] = tx_saved\n")
			code.write("var tx_buf bytes.Buffer\n")
			callParams := []string{"&tx_buf", "tx_id"}
			if len(comp.Slots) > 0 {
				callParams = append(callParams, "tx_pid", "tx_loc")
			}
			if comp.HasChildComps {
				callParams = append(callParams, "tx_curr_saved", "tx_next_saved")
			}
			for _, v := range comp.Vars {
				if _, ok := comp.UsedVars[v.GoName]; ok {
					switch v.Type {
					case VarTypeState, VarTypeProp:
						callParams = append(callParams, "tx_saved."+v.SavedField)
					case VarTypeDerived:
						callParams = append(callParams, "tx_derived_"+v.GoName)
					}
				}
			}
			for _, f := range comp.Funcs {
				callParams = append(callParams, fmt.Sprintf("\"%s\"", comp.Name+":"+f.Name), "tx_id")
			}
			code.write("render_%s(%s", comp.GoName, strings.Join(callParams, ", "))
			for _, slotName := range comp.Slots {
				if len(comp.CompFills) == 0 {
					code.write(", nil")
				} else {
					code.write(", func() {\n")
					code.write("render_comp_fill_%s(&tx_buf, tx_loc+\"_%s\", tx_pid, tx_curr_saved", comp.GoName, slotName)
					if comp.CompFillsHasChildComps {
						code.write(", tx_next_saved")
					}
					code.write(")\n")
					code.write("}")
				}
			}
			code.write(")\n")
			code.write("tx_w.Write(tx_buf.Bytes())\n")
			code.write("tx_w.Write([]byte(\"<script id=\\\"tx-saved\\\" type=\\\"application/json\\\">\"))\n")
			code.write("tx_savedBytes, _ := json.Marshal(tx_next_saved)\n")
			code.write("tx_w.Write(tx_savedBytes)\n")
			code.write("tx_w.Write([]byte(\"</script>\"))\n")
			code.write("},\n")
			code.write("},\n")
		}
	}
	code.write("}\n")

	code.write("func Routes() []TxRoute { return txRoutes }")

	data := []byte(code.String())
	formatted, err := imports.Process(outputFilePath, data, nil)
	if err != nil {
		lines := strings.Split(string(data), "\n")
		start, end := 0, len(lines)
		var errs scanner.ErrorList
		if errors.As(err, &errs) && len(errs) > 0 {
			errLine := errs[0].Pos.Line
			start = max(errLine-6, 0)
			end = min(errLine+5, len(lines))
		}
		for i := start; i < end; i++ {
			log.Printf("%d: %s\n", i+1, lines[i])
		}
		log.Fatalln(fmt.Errorf("format generated code: %w", err))
	}

	dir = filepath.Dir(outputFilePath)
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
	log.Printf("%s generated successfully (%d pages, %d components)\n", outputFilePath, len(pages), len(componentsByName))

}

type CompType int

const (
	CompTypeComp CompType = iota
	CompTypePage
)

type Component struct {
	Type     CompType
	FilePath string
	RelPath  string
	Name     string
	GoName   string

	TmplxScriptNode *html.Node
	TemplateNode    *html.Node
	StyleNode       *html.Node
	Slots           []string

	Imports    []*ast.ImportSpec
	Vars       []*Var
	VarByName  map[string]*Var
	InitFunc   *Func
	Funcs      []*Func
	FuncByName map[string]*Func

	ChildCompsIdGen        map[string]*IdGen
	HasChildComps          bool
	UsedVars               map[string]struct{}
	Fills                  []*Fill
	FillByGoName           map[string]*Fill
	CompFills              []*Fill
	CompFillsMu            sync.Mutex
	CompFillsHasChildComps bool

	AnonFuncNameGen *IdGen
	AnonFuncs       []*Func
	RenderFunc      Code
}

func (comp *Component) errf(msg string, a ...any) error {
	return fmt.Errorf(comp.RelPath+": "+msg, a...)
}

func (comp *Component) parseSlots(node *html.Node, inSlot bool) *MultiError {
	merr := newMultiError()
	if node.Type != html.ElementNode {
		return nil
	}

	isSlot := node.DataAtom == atom.Slot
	if isSlot {
		if inSlot {
			merr.append(comp.errf("<slot> cannot be nested inside another <slot>"))
		}

		slotName := ""
		if name, found := hasAttr(node, "name"); found {
			slotName = name
		}

		if !slices.Contains(comp.Slots, slotName) {
			comp.Slots = append(comp.Slots, slotName)
		} else {
			if slotName == "" {
				merr.append(comp.errf("duplicate default <slot> (only one allowed)"))
			} else {
				merr.append(comp.errf("duplicate <slot name=\"%s\"> (only one allowed)", slotName))
			}
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		merr.concat(comp.parseSlots(c, isSlot))
	}

	return merr
}

func (comp *Component) parseTmplxScript() *MultiError {
	merr := newMultiError()

	comp.Imports = []*ast.ImportSpec{}
	comp.Vars = []*Var{}
	comp.VarByName = map[string]*Var{}
	comp.Funcs = []*Func{}
	comp.FuncByName = map[string]*Func{}

	if comp.TmplxScriptNode != nil {
		scriptAst, err := parser.ParseFile(token.NewFileSet(), "", "package p\n"+comp.TmplxScriptNode.FirstChild.Data, parser.ParseComments)
		if err != nil {
			merr.append(comp.errf("syntax error in <script type=\"text/tmplx\">: %w", err))
			return merr
		}

		allVarNames := map[string]struct{}{}
		for _, decl := range scriptAst.Decls {
			if d, ok := decl.(*ast.GenDecl); ok && d.Tok == token.VAR && len(d.Specs) > 0 {
				if s, ok := d.Specs[0].(*ast.ValueSpec); ok && len(s.Names) > 0 {
					allVarNames[s.Names[0].Name] = struct{}{}
				}
			}
		}

		for _, decl := range scriptAst.Decls {
			switch d := decl.(type) {
			case *ast.BadDecl:
				merr.append(comp.errf("invalid declaration: %s", astToSource(decl)))
			case *ast.GenDecl:
				switch d.Tok {
				case token.IMPORT:
					for _, spec := range d.Specs {
						s, ok := spec.(*ast.ImportSpec)
						if !ok {
							merr.append(comp.errf("invalid import: %s", astToSource(spec)))
							continue
						}

						comp.Imports = append(comp.Imports, s)
					}
				case token.VAR:
					if len(d.Specs) == 0 {
						continue
					}
					if len(d.Specs) > 1 {
						merr.append(comp.errf("declare one variable per var statement: %s", astToSource(d)))
						continue
					}

					spec := d.Specs[0]
					s, ok := spec.(*ast.ValueSpec)
					if !ok {
						merr.append(comp.errf("invalid variable declaration: %s", astToSource(spec)))
						continue
					}

					if s.Type == nil {
						merr.append(comp.errf("missing type annotation: %s", astToSource(spec)))
					}

					if len(s.Names) > 1 {
						merr.append(comp.errf("declare one variable per var statement: %s", astToSource(spec)))
						continue
					}

					ident := s.Names[0]
					if strings.HasPrefix(ident.Name, "tx_") {
						merr.append(comp.errf("%s: variable name cannot start with tx_ (reserved prefix)", ident.Name))
						continue
					}

					newVar := &Var{
						GoName:     ident.Name,
						SavedField: "S_" + ident.Name,
						TypeExpr:   astToSource(s.Type),
					}

					isProp := false
					isPath := false
					if d.Doc != nil {
						comments := []Comment{}
						for _, comment := range d.Doc.List {
							comments = append(comments, parseComments(comment.Text)...)
						}

						for _, comment := range comments {
							switch comment.Name {
							case CommentProp:
								isProp = true
							case CommentPath:
								isPath = true
								pathAst := &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "tx_r"},
										Sel: &ast.Ident{Name: "PathValue"},
									},
									Args: []ast.Expr{
										&ast.BasicLit{
											Value: fmt.Sprintf("\"%s\"", comment.Value),
											Kind:  token.STRING,
										},
									},
								}
								newVar.InitExprAst = pathAst
								newVar.InitExpr = astToSource(pathAst)
							}
						}
					}

					if isProp && isPath {
						merr.append(comp.errf("cannot combine //tx:prop and //tx:path on %s", ident.Name))
					} else if isProp {
						if comp.Type == CompTypePage {
							merr.append(comp.errf("//tx:prop on %s: pages cannot have props", ident.Name))
						}
						if len(s.Values) == 1 {
							newVar.InitExprAst = s.Values[0]
							newVar.InitExpr = astToSource(s.Values[0])
						}
						newVar.Type = VarTypeProp

					} else if isPath {
						if len(s.Values) > 0 {
							merr.append(comp.errf("//tx:path variable cannot have an initial value: %s", astToSource(spec)))
						}
						if astToSource(s.Type) != "string" {
							merr.append(comp.errf("//tx:path variable must be type string: %s", astToSource(spec)))
						}
						newVar.Type = VarTypeState

					} else if len(s.Values) == 1 || len(s.Values) == 0 {
						found := false
						if len(s.Values) == 1 {
							v := s.Values[0]
							astutil.Apply(v, func(c *astutil.Cursor) bool {
								id, ok := c.Node().(*ast.Ident)
								if !ok {
									return true
								}
								if !atVarRefPos(c) {
									return false
								}
								if _, inAll := allVarNames[id.Name]; inAll {
									if _, inDeclared := comp.VarByName[id.Name]; !inDeclared {
										merr.append(comp.errf("%s: variable %s used before declaration", s.Names[0].Name, id.Name))
									}
									found = true
								}
								return false
							}, nil)
							newVar.InitExprAst = v
							cloned, _ := parser.ParseExpr(astToSource(v))
							newVar.InitExpr = astToSource(comp.rewriteVarRefs(cloned))
						}

						if found {
							newVar.Type = VarTypeDerived
						} else {
							newVar.Type = VarTypeState
						}

					} else if len(s.Values) > 1 {
						merr.append(comp.errf("declare one variable per var statement: %s", astToSource(spec)))
					}

					comp.Vars = append(comp.Vars, newVar)
					comp.VarByName[ident.Name] = newVar
				}
			}
		}

		for _, decl := range scriptAst.Decls {
			d, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if d.Recv != nil {
				merr.append(comp.errf("%s: methods (func with receiver) not allowed, use plain functions", d.Name))
			}

			if d.Type.Results != nil {
				merr.append(comp.errf("%s: return values not allowed", d.Name))
			}

			for _, field := range d.Type.Params.List {
				for _, name := range field.Names {
					if strings.HasPrefix(name.Name, "tx_") {
						merr.append(comp.errf("%s: parameter %s cannot start with tx_ (reserved prefix)", d.Name, name.Name))
					}
					if comp.VarByName[name.Name] != nil {
						merr.append(comp.errf("%s: parameter %s shadows a state variable", d.Name, name.Name))
					}
				}
			}

			if d.Body != nil {
				for _, name := range comp.readOnlyMutations(d.Body) {
					merr.append(comp.errf("%s: cannot assign to %s (derived/prop is read-only)", d.Name, name))
				}
				for _, name := range comp.shadowingLocals(d.Body) {
					merr.append(comp.errf("%s: local variable %s shadows a state/prop/derived variable (rename the local)", d.Name, name))
				}
			}

			if strings.HasPrefix(d.Name.Name, "tx_") {
				merr.append(comp.errf("%s: function name cannot start with tx_ (reserved prefix)", d.Name.Name))
				continue
			}

			newFunc := &Func{
				Name: d.Name.Name,
				Decl: d,
			}
			dirtyDerived := comp.dirtyDerivedNames(d.Body)
			var b strings.Builder
			for _, stmt := range d.Body.List {
				b.WriteString(astToSource(comp.rewriteVarRefs(stmt)))
				b.WriteByte('\n')
			}
			for _, name := range dirtyDerived {
				v := comp.VarByName[name]
				fmt.Fprintf(&b, "tx_derived_%s = %s\n", v.GoName, v.InitExpr)
			}
			newFunc.Stmts = b.String()

			if d.Name.Name == "init" {
				comp.InitFunc = newFunc
			} else {
				comp.FuncByName[d.Name.Name] = newFunc
				comp.Funcs = append(comp.Funcs, newFunc)
			}
		}
	}

	return merr
}

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

func (comp *Component) rewriteVarRefs(node ast.Node) ast.Node {
	return astutil.Apply(node, func(c *astutil.Cursor) bool {
		ident, ok := c.Node().(*ast.Ident)
		if !ok {
			return true
		}
		if !atVarRefPos(c) {
			return false
		}
		v, ok := comp.VarByName[ident.Name]
		if !ok {
			return false
		}
		switch v.Type {
		case VarTypeState, VarTypeProp:
			c.Replace(&ast.SelectorExpr{
				X:   &ast.Ident{Name: "tx_saved"},
				Sel: &ast.Ident{Name: v.SavedField},
			})
		case VarTypeDerived:
			c.Replace(&ast.Ident{Name: "tx_derived_" + v.GoName})
		}
		return false
	}, nil)
}

func (comp *Component) readOnlyMutations(node ast.Node) []string {
	seen := map[string]struct{}{}
	var result []string
	check := func(expr ast.Expr) {
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
			return
		}
		v, ok := comp.VarByName[ident.Name]
		if !ok {
			return
		}
		if _, dup := seen[v.GoName]; dup {
			return
		}
		if v.Type == VarTypeDerived || v.Type == VarTypeProp {
			seen[v.GoName] = struct{}{}
			result = append(result, v.GoName)
		}
	}
	ast.Inspect(node, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			for _, lhs := range stmt.Lhs {
				check(lhs)
			}
		case *ast.IncDecStmt:
			check(stmt.X)
		}
		return true
	})
	return result
}

func (comp *Component) shadowingLocals(body ast.Node) []string {
	var found []string
	ast.Inspect(body, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			if stmt.Tok != token.DEFINE {
				return true
			}
			for _, lhs := range stmt.Lhs {
				ident, ok := lhs.(*ast.Ident)
				if !ok {
					continue
				}
				if ident.Name == "_" {
					continue
				}
				if _, ok := comp.VarByName[ident.Name]; ok {
					found = append(found, ident.Name)
				}
			}
		case *ast.RangeStmt:
			if stmt.Tok != token.DEFINE {
				return true
			}
			for _, expr := range []ast.Expr{stmt.Key, stmt.Value} {
				if expr == nil {
					continue
				}
				ident, ok := expr.(*ast.Ident)
				if !ok {
					continue
				}
				if ident.Name == "_" {
					continue
				}
				if _, ok := comp.VarByName[ident.Name]; ok {
					found = append(found, ident.Name)
				}
			}
		case *ast.DeclStmt:
			genDecl, ok := stmt.Decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.VAR {
				return true
			}
			for _, spec := range genDecl.Specs {
				valSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range valSpec.Names {
					if name.Name == "_" {
						continue
					}
					if _, ok := comp.VarByName[name.Name]; ok {
						found = append(found, name.Name)
					}
				}
			}
		}
		return true
	})
	return found
}

func (comp *Component) dirtyDerivedNames(body *ast.BlockStmt) []string {
	dirty := map[string]struct{}{}
	markBase := func(lhs ast.Expr) {
		var ident *ast.Ident
		ast.Inspect(lhs, func(n ast.Node) bool {
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
			return
		}
		if v, ok := comp.VarByName[ident.Name]; ok {
			if v.Type == VarTypeState || v.Type == VarTypeProp {
				dirty[ident.Name] = struct{}{}
			}
		}
	}
	ast.Inspect(body, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			for _, lhs := range stmt.Lhs {
				markBase(lhs)
			}
		case *ast.IncDecStmt:
			markBase(stmt.X)
		}
		return true
	})
	result := []string{}
	for _, v := range comp.Vars {
		if v.Type != VarTypeDerived {
			continue
		}
		needsRecalc := false
		astutil.Apply(v.InitExprAst, func(c *astutil.Cursor) bool {
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
			result = append(result, v.GoName)
			dirty[v.GoName] = struct{}{}
		}
	}
	return result
}

func (comp *Component) parseUsedVars(node *html.Node) *MultiError {
	merr := newMultiError()
	switch node.Type {
	case html.DocumentNode:
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			merr.concat(comp.parseUsedVars(c))
		}
		return merr

	case html.TextNode:
		_, hasTxIgnore := hasAttr(node.Parent, "tx-ignore")
		// script/style bodies are JS/CSS where { is not template syntax, so skip interpolation.
		isScriptOrStyle := node.Parent.DataAtom == atom.Script || node.Parent.DataAtom == atom.Style
		if hasTxIgnore || isScriptOrStyle {
			return nil
		}

		comp.parseUsedVarsStr(node.Data)

	case html.ElementNode:
		if childComp, ok := componentsByName[node.Data]; ok {
			comp.HasChildComps = true
			idNum := comp.ChildCompsIdGen[childComp.Name].next()

			for _, v := range childComp.Vars {
				if v.Type == VarTypeProp {
					if val, found := hasAttr(node, v.GoName); found {
						if propExpr, perr := parser.ParseExpr(val); perr == nil {
							comp.markUsedVars(propExpr)
						}
					}
				}
			}

			fillNodes := parseFillNodes(node)
			for _, slotName := range childComp.Slots {
				if n, ok := fillNodes[slotName]; ok {
					savedHasChildComps := comp.HasChildComps
					savedUsedVars := comp.UsedVars

					comp.HasChildComps = false
					comp.UsedVars = map[string]struct{}{}

					merr.concat(comp.parseUsedVars(n))

					fillHasChildComps := comp.HasChildComps
					fillUsedVars := comp.UsedVars

					comp.HasChildComps = savedHasChildComps || fillHasChildComps
					comp.UsedVars = savedUsedVars
					for v := range fillUsedVars {
						comp.UsedVars[v] = struct{}{}
					}

					fillGoName := fmt.Sprintf("%s_%s_%s_%s", comp.GoName, childComp.GoName, idNum, goIdent(slotName))
					newFill := &Fill{
						GoName:     fillGoName,
						CompName:   childComp.Name,
						ParentComp: comp,
						Location:   fmt.Sprintf("%s_%s_%s", comp.Name, idNum, slotName),

						HasChildComps: fillHasChildComps,
						UsedVars:      fillUsedVars,
						RenderFunc:    newCode("tx_w"),
					}
					comp.Fills = append(comp.Fills, newFill)
					comp.FillByGoName[fillGoName] = newFill
					childComp.CompFillsMu.Lock()
					childComp.CompFills = append(childComp.CompFills, newFill)
					childComp.CompFillsMu.Unlock()
				}
			}
			return merr

		} else if node.DataAtom == atom.Slot {
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				merr.concat(comp.parseUsedVars(c))
			}
			return merr

		} else if node.DataAtom != atom.Template {
			if _, isIgnored := hasAttr(node, "tx-ignore"); !isIgnored {
				for _, attr := range node.Attr {
					if attr.Key == "tx-if" || attr.Key == "tx-else-if" || attr.Key == "tx-else" || attr.Key == "tx-for" {
						continue
					}
					if strings.HasPrefix(attr.Key, "tx-on") || attr.Key == "tx-action" {
						continue
					}
					comp.parseUsedVarsStr(attr.Val)
				}
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				if _, field := condState(c); field != "" {
					if fieldExpr, ferr := parser.ParseExpr(field); ferr == nil {
						comp.markUsedVars(fieldExpr)
					}
				}

				if stmt, ok := hasAttr(c, "tx-for"); ok {
					if forAst, ferr := parser.ParseFile(token.NewFileSet(), "", "package p\nfunc f(){for "+stmt+"{}}", 0); ferr == nil {
						comp.markUsedVars(forAst)
					}
				}

				if key, ok := hasAttr(c, "tx-key"); ok {
					if keyExpr, kerr := parser.ParseExpr(key); kerr == nil {
						comp.markUsedVars(keyExpr)
					}
				}
			}

			merr.concat(comp.parseUsedVars(c))
		}

		return merr
	}

	return nil
}

func (comp *Component) scanVarRefs(node ast.Node, target map[string]struct{}) {
	astutil.Apply(node, func(c *astutil.Cursor) bool {
		id, ok := c.Node().(*ast.Ident)
		if !ok {
			return true
		}
		if !atVarRefPos(c) {
			return false
		}
		if _, ok := comp.VarByName[id.Name]; ok {
			target[id.Name] = struct{}{}
		}
		return false
	}, nil)
}

func (comp *Component) markUsedVars(node ast.Node) {
	comp.scanVarRefs(node, comp.UsedVars)
}

func (comp *Component) parseTmpl(node *html.Node, forKeys []string, inSlot bool) *MultiError {
	merr := newMultiError()
	switch node.Type {
	case html.DoctypeNode:
		comp.RenderFunc.emitStrLit(fmt.Sprintf("<!DOCTYPE %s>", node.Data))
		return nil
	case html.CommentNode:
		comp.RenderFunc.emitStrLit(fmt.Sprintf("<!--%s-->", node.Data))
		return nil
	case html.DocumentNode:
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			merr.concat(comp.parseTmpl(c, forKeys, inSlot))
		}
		return merr
	case html.TextNode:
		_, hasTxIgnore := hasAttr(node.Parent, "tx-ignore")
		// script/style bodies are JS/CSS where { is not template syntax, so skip interpolation.
		isScriptOrStyle := node.Parent.DataAtom == atom.Script || node.Parent.DataAtom == atom.Style
		isRawText := isChildNodeRawText(node.Parent.Data)

		if hasTxIgnore || isScriptOrStyle {
			if isRawText {
				comp.RenderFunc.emitStrLit(node.Data)
			} else {
				comp.RenderFunc.emitStrLit(html.EscapeString(node.Data))
			}
			return nil
		}

		return newMultiError(comp.parseTmplStr(node.Data, !isRawText))
	case html.ElementNode:
		if childComp, ok := componentsByName[node.Data]; ok {
			comp.HasChildComps = true
			idNum := comp.ChildCompsIdGen[childComp.Name].next()
			id := fmt.Sprintf("%s-%s", childComp.Name, idNum)

			for _, attr := range node.Attr {
				if attr.Key == "tx-if" || attr.Key == "tx-else-if" || attr.Key == "tx-else" || attr.Key == "tx-for" || attr.Key == "tx-key" || attr.Key == "slot" {
					continue
				}
				if _, ok := childComp.FuncByName[attr.Key]; ok {
					continue
				}

				v, ok := childComp.VarByName[attr.Key]
				if !ok {
					merr.append(comp.errf("<%s %s=\"...\">: %s is not a prop or function in component %s", node.Data, attr.Key, attr.Key, childComp.Name))
					continue
				}

				if v.Type != VarTypeProp {
					merr.append(comp.errf("<%s %s=\"...\">: %s is a state variable, not a prop; use //tx:prop to declare it as a prop", node.Data, attr.Key, attr.Key))
					continue
				}

				parentVar, ok := comp.VarByName[attr.Val]
				if !ok {
					continue
				}

				propType := v.TypeExpr
				parentType := parentVar.TypeExpr
				if propType != parentType {
					merr.append(comp.errf("<%s %s=\"%s\">: type mismatch; prop %s expects %s but got %s", node.Data, attr.Key, attr.Val, attr.Key, propType, parentType))
					continue
				}
			}

			comp.RenderFunc.emitGo("{\n")

			comp.RenderFunc.emitGo("tx_cid := ")
			if inSlot {
				comp.RenderFunc.emitGo("tx_id")
				for _, key := range forKeys {
					comp.RenderFunc.emitGo(fmt.Sprintf(" + \":\" + fmt.Sprint(%s)", key))
				}
				comp.RenderFunc.emitGo(fmt.Sprintf(" + \"@%s\"", id))
			} else {
				switch comp.Type {
				case CompTypePage:
					if len(forKeys) == 0 {
						comp.RenderFunc.emitGo(fmt.Sprintf("\"%s\"", id))
					} else {
						comp.RenderFunc.emitGo(fmt.Sprintf("fmt.Sprint(%s)", forKeys[0]))
						for _, key := range forKeys[1:] {
							comp.RenderFunc.emitGo(fmt.Sprintf(" + \":\" + fmt.Sprint(%s)", key))
						}
						comp.RenderFunc.emitGo(fmt.Sprintf(" + \":%s\"", id))
					}
				case CompTypeComp:
					comp.RenderFunc.emitGo("tx_id")
					for _, key := range forKeys {
						comp.RenderFunc.emitGo(fmt.Sprintf(" + \":\" + fmt.Sprint(%s)", key))
					}
					comp.RenderFunc.emitGo(fmt.Sprintf(" + \":%s\"", id))
				}
			}
			comp.RenderFunc.emitGo("\n")

			if len(childComp.Vars) == 0 {
				comp.RenderFunc.emitGo(fmt.Sprintf("tx_next_saved[tx_cid] = &%s{}\n", childComp.GoName))
			} else {
				comp.RenderFunc.emitGo(fmt.Sprintf("tx_saved := &%s{}\n", childComp.GoName))
				for _, v := range childComp.Vars {
					if v.Type == VarTypeDerived {
						comp.RenderFunc.emitGo(fmt.Sprintf("var tx_derived_%s %s\n", v.GoName, v.TypeExpr))
					}
				}

				comp.RenderFunc.emitGo("tx_curr_saved_str, tx_curr_saved_exist := tx_curr_saved[tx_cid]\n")
				comp.RenderFunc.emitGo("if tx_curr_saved_exist {\n")
				comp.RenderFunc.emitGo("json.Unmarshal([]byte(tx_curr_saved_str), tx_saved)\n")
				for _, v := range childComp.Vars {
					if v.Type == VarTypeProp {
						if val, found := hasAttr(node, v.GoName); found {
							comp.RenderFunc.emitGo(fmt.Sprintf("tx_saved.%s = %s\n", v.SavedField, val))
						}
					}
				}
				for _, v := range childComp.Vars {
					if v.Type == VarTypeDerived {
						comp.RenderFunc.emitGo(fmt.Sprintf("tx_derived_%s = %s\n", v.GoName, v.InitExpr))
					}
				}

				hasElseContent := childComp.InitFunc != nil
				for _, v := range childComp.Vars {
					if v.Type == VarTypeDerived {
						hasElseContent = true
					} else if v.Type == VarTypeState && v.InitExpr != "" {
						hasElseContent = true
					} else if v.Type == VarTypeProp {
						if _, found := hasAttr(node, v.GoName); found || v.InitExpr != "" {
							hasElseContent = true
						}
					}
				}

				if hasElseContent {
					comp.RenderFunc.emitGo("} else {\n")
					for _, v := range childComp.Vars {
						switch v.Type {
						case VarTypeState:
							if v.InitExpr != "" {
								comp.RenderFunc.emitGo(fmt.Sprintf("tx_saved.%s = %s\n", v.SavedField, v.InitExpr))
							}
						case VarTypeProp:
							if val, found := hasAttr(node, v.GoName); found {
								comp.RenderFunc.emitGo(fmt.Sprintf("tx_saved.%s = %s\n", v.SavedField, val))
							} else if v.InitExpr != "" {
								comp.RenderFunc.emitGo(fmt.Sprintf("tx_saved.%s = %s\n", v.SavedField, v.InitExpr))
							}
						case VarTypeDerived:
							comp.RenderFunc.emitGo(fmt.Sprintf("tx_derived_%s = %s\n", v.GoName, v.InitExpr))
						}
					}

					if childComp.InitFunc != nil {
						comp.RenderFunc.emitGo(childComp.InitFunc.Stmts)
					}
				}
				comp.RenderFunc.emitGo("}\n")
				comp.RenderFunc.emitGo("tx_next_saved[tx_cid] = tx_saved\n")
			}

			comp.RenderFunc.emitGo(fmt.Sprintf("render_%s(%s, tx_cid", childComp.GoName, comp.RenderFunc.PendingSegment.BufName))

			if len(childComp.Slots) > 0 {
				parent := "\"page\""
				if comp.Type == CompTypeComp {
					parent = "tx_id"
					for _, key := range forKeys {
						parent += ` + ":" + fmt.Sprint(` + key + `)`
					}
				}
				comp.RenderFunc.emitGo(fmt.Sprintf(", %s, \"%s_%s\"", parent, comp.Name, idNum))
			}
			if childComp.HasChildComps {
				comp.RenderFunc.emitGo(", tx_curr_saved, tx_next_saved")
			}

			for _, v := range childComp.Vars {
				if _, ok := childComp.UsedVars[v.GoName]; !ok {
					continue
				}
				switch v.Type {
				case VarTypeProp, VarTypeState:
					comp.RenderFunc.emitGo(fmt.Sprintf(", tx_saved.%s", v.SavedField))
				case VarTypeDerived:
					comp.RenderFunc.emitGo(fmt.Sprintf(", tx_derived_%s", v.GoName))
				}
			}

			for _, f := range childComp.Funcs {
				if val, found := hasAttr(node, f.Name); found {
					if pf, ok := comp.FuncByName[val]; ok {
						comp.RenderFunc.emitGo(fmt.Sprintf(", %s, %s_swap", pf.Name, pf.Name))
					} else {
						merr.append(comp.errf("undefined function: %s", val))
					}
				} else {
					if f.Decl.Body == nil {
						merr.append(comp.errf("function %s has no body in %s and must be passed as a prop", f.Name, childComp.Name))
					} else {
						comp.RenderFunc.emitGo(fmt.Sprintf(",\"%s:%s\", tx_cid", childComp.Name, f.Name))
					}
				}
			}

			fillNodes := parseFillNodes(node)
			if len(childComp.Slots) > 0 {
				comp.RenderFunc.emitGo(",\n")
				for _, slotName := range childComp.Slots {
					if n, ok := fillNodes[slotName]; ok {
						savedRenderFunc := comp.RenderFunc
						comp.RenderFunc = newCode("tx_w")

						merr.concat(comp.parseTmpl(n, forKeys, true))

						currFillRenderFunc := comp.RenderFunc
						comp.RenderFunc = savedRenderFunc

						fill := comp.FillByGoName[fmt.Sprintf("%s_%s_%s_%s", comp.GoName, childComp.GoName, idNum, goIdent(slotName))]
						fill.RenderFunc = currFillRenderFunc
						comp.RenderFunc.emitGo(fmt.Sprintf("func () { render_fill_%s(%s", fill.GoName, comp.RenderFunc.PendingSegment.BufName))
						if fill.HasChildComps {
							comp.RenderFunc.emitGo(", tx_cid, tx_curr_saved, tx_next_saved")
						}
						for _, v := range comp.Vars {
							if _, ok := fill.UsedVars[v.GoName]; ok {
								comp.RenderFunc.emitGo(fmt.Sprintf(", %s", v.GoName))
							}
						}
						comp.RenderFunc.emitGo(") },\n")
					} else {
						comp.RenderFunc.emitGo("nil,\n")
					}
				}
			}
			comp.RenderFunc.emitGo(")\n")
			comp.RenderFunc.emitGo("}\n")
			return merr
		} else if node.DataAtom == atom.Slot {
			val, _ := hasAttr(node, "name")
			comp.RenderFunc.emitGo(fmt.Sprintf("if tx_render_fill_%s != nil {\n", val))
			comp.RenderFunc.emitGo(fmt.Sprintf("tx_render_fill_%s()\n", val))
			if node.FirstChild != nil {
				comp.RenderFunc.emitGo("} else {\n")
				child := newTemplateNode()
				for c := node.FirstChild; c != nil; c = c.NextSibling {
					child.AppendChild(&html.Node{
						FirstChild: c.FirstChild,
						LastChild:  c.LastChild,
						Type:       c.Type,
						DataAtom:   c.DataAtom,
						Namespace:  c.Namespace,
						Data:       c.Data,
						Attr:       c.Attr,
					})
				}
				merr.concat(comp.parseTmpl(child, forKeys, false))
			}
			comp.RenderFunc.emitGo("}\n")
			return merr

		} else if node.DataAtom != atom.Template {
			comp.RenderFunc.emitStrLit("<")
			comp.RenderFunc.emitStrLit(node.Data)

			_, isIgnore := hasAttr(node, "tx-ignore")

			for _, attr := range node.Attr {
				if attr.Key == "tx-if" || attr.Key == "tx-else-if" || attr.Key == "tx-else" || attr.Key == "tx-for" {
					continue
				}

				comp.RenderFunc.emitStrLit(" ")
				if strings.HasPrefix(attr.Key, "tx-on") {
					comp.RenderFunc.emitStrLit(attr.Key)
					comp.RenderFunc.emitStrLit(`="`)

					// Check if it's a function call
					if expr, err := parser.ParseExpr(attr.Val); err == nil {
						if callExpr, ok := expr.(*ast.CallExpr); ok {
							if ident, ok := callExpr.Fun.(*ast.Ident); ok {
								if fun, ok := comp.FuncByName[ident.Name]; ok {
									params := []string{}
									for _, list := range fun.Decl.Type.Params.List {
										for _, ident := range list.Names {
											params = append(params, ident.Name)
										}
									}

									if len(params) != len(callExpr.Args) {
										merr.append(comp.errf("wrong number of arguments: %s", astToSource(callExpr)))
										continue
									}

									comp.RenderFunc.emitExpr(fun.Name)
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

											if _, ok := comp.VarByName[ident.Name]; ok {
												foundVar = true
												return false
											}

											return true
										})

										if foundVar {
											merr.append(comp.errf("cannot pass state/derived variable as event handler argument: %s", callExpr.Args[i]))
											continue
										}

										if i == 0 {
											comp.RenderFunc.emitStrLit("?" + param + "=")
										} else {
											comp.RenderFunc.emitStrLit("&" + param + "=")
										}

										arg := astToSource(callExpr.Args[i])
										comp.RenderFunc.emitUrlEscapeExpr(arg)
									}
									comp.RenderFunc.emitStrLit(`"`)

									if comp.Type == CompTypeComp {
										comp.RenderFunc.emitStrLit(" tx-swap=\"")
										if comp.Type != CompTypePage {
											comp.RenderFunc.emitExpr(fun.Name + "_swap")
										}
										comp.RenderFunc.emitStrLit(`"`)
									}

									if len(comp.Slots) > 0 {
										comp.RenderFunc.emitStrLit(" tx-pid=\"")
										comp.RenderFunc.emitExpr("tx_pid")
										comp.RenderFunc.emitStrLit("\"")
										comp.RenderFunc.emitStrLit(" tx-loc=\"")
										comp.RenderFunc.emitExpr("tx_loc")
										comp.RenderFunc.emitStrLit("\"")
									}

									continue
								}
							}
						}
					}

					idNum := comp.AnonFuncNameGen.next()
					funcName := fmt.Sprintf("af-%s", idNum)
					fileAst, err := parser.ParseFile(token.NewFileSet(), comp.FilePath, fmt.Sprintf("package p\nfunc f() {\n%s\n}", attr.Val), 0)
					if err != nil {
						merr.append(comp.errf("invalid inline handler: %s", attr.Val))
						continue
					}

					decl, ok := fileAst.Decls[0].(*ast.FuncDecl)
					if !ok {
						merr.append(comp.errf("invalid inline handler: %s", attr.Val))
						continue
					}

					if modified := comp.readOnlyMutations(decl); len(modified) > 0 {
						merr.append(comp.errf("cannot assign to derived/prop variable in handler: %v", modified))
						continue
					}
					for _, name := range comp.shadowingLocals(decl) {
						merr.append(comp.errf("inline handler: local variable %s shadows a state/prop/derived variable (rename the local)", name))
					}

					var b strings.Builder
					dirtyDerived := comp.dirtyDerivedNames(decl.Body)
					for _, stmt := range decl.Body.List {
						b.WriteString(astToSource(comp.rewriteVarRefs(stmt)))
						b.WriteByte('\n')
					}
					for _, name := range dirtyDerived {
						v := comp.VarByName[name]
						fmt.Fprintf(&b, "tx_derived_%s = %s\n", v.GoName, v.InitExpr)
					}

					comp.AnonFuncs = append(comp.AnonFuncs, &Func{
						Name:  funcName,
						Decl:  decl,
						Stmts: b.String(),
					})

					comp.RenderFunc.emitStrLit(comp.Name + ":" + funcName)
					comp.RenderFunc.emitStrLit("\"")

					if comp.Type == CompTypeComp {
						comp.RenderFunc.emitStrLit(" tx-swap=\"")
						comp.RenderFunc.emitExpr("tx_id")
						comp.RenderFunc.emitStrLit(`"`)
					}

					if len(comp.Slots) > 0 {
						comp.RenderFunc.emitStrLit(" tx-pid=\"")
						comp.RenderFunc.emitExpr("tx_pid")
						comp.RenderFunc.emitStrLit("\"")
						comp.RenderFunc.emitStrLit(" tx-loc=\"")
						comp.RenderFunc.emitExpr("tx_loc")
						comp.RenderFunc.emitStrLit("\"")
					}

				} else if attr.Key == "tx-action" {
					if node.DataAtom != atom.Form {
						merr.append(comp.errf("tx-action only allowed on <form>, got <%s>", node.Data))
						continue
					}
					if !token.IsIdentifier(attr.Val) {
						merr.append(comp.errf("tx-action value must be a function name, got \"%s\"", attr.Val))
						continue
					}
					fun, ok := comp.FuncByName[attr.Val]
					if !ok {
						merr.append(comp.errf("tx-action: undefined function %s", attr.Val))
						continue
					}
					comp.RenderFunc.emitStrLit("tx-action=\"")
					comp.RenderFunc.emitExpr(fun.Name)
					comp.RenderFunc.emitStrLit("\"")
					if comp.Type == CompTypeComp {
						comp.RenderFunc.emitStrLit(" tx-swap=\"")
						comp.RenderFunc.emitExpr(fun.Name + "_swap")
						comp.RenderFunc.emitStrLit("\"")
					}
					if len(comp.Slots) > 0 {
						comp.RenderFunc.emitStrLit(" tx-pid=\"")
						comp.RenderFunc.emitExpr("tx_pid")
						comp.RenderFunc.emitStrLit("\"")
						comp.RenderFunc.emitStrLit(" tx-loc=\"")
						comp.RenderFunc.emitExpr("tx_loc")
						comp.RenderFunc.emitStrLit("\"")
					}
				} else {
					if attr.Namespace != "" {
						comp.RenderFunc.emitStrLit(node.Namespace)
						comp.RenderFunc.emitStrLit(":")
					}
					comp.RenderFunc.emitStrLit(attr.Key)
					comp.RenderFunc.emitStrLit(`="`)
					if isIgnore {
						comp.RenderFunc.emitStrLit(attr.Val)
					} else if err := comp.parseTmplStr(attr.Val, false); err != nil {
						merr.append(comp.errf("invalid expression in attribute: %s", attr.Val))
					}
					comp.RenderFunc.emitStrLit(`"`)
				}
			}

			// https://html.spec.whatwg.org/#void-elements
			if isVoidElement(node.Data) {
				if node.FirstChild != nil {
					merr.append(comp.errf("void element <%s> cannot have children", node.Data))
				}

				comp.RenderFunc.emitStrLit("/>")
				return merr
			}

			comp.RenderFunc.emitStrLit(">")
		}

		// prevCondState tracks the conditional directive on the previous sibling:
		// 0=none, 1=if, 2=else-if, 3=else. It is used to emit a closing "}" when
		// a conditional chain ends, and to validate that tx-else-if/tx-else only
		// follow tx-if or tx-else-if.
		txNodeId, _ := hasAttr(node, "id")
		if node.DataAtom == atom.Script && txNodeId == "tx-runtime" {
			comp.RenderFunc.emitGo(fmt.Sprintf("%s.WriteString(runtimeScript)\n", comp.RenderFunc.PendingSegment.BufName))
		} else if node.DataAtom == atom.Script && txNodeId == "tx-saved" {
			comp.RenderFunc.emitSplit()
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
							merr.append(comp.errf("tx-else-if/tx-else on <%s> without preceding tx-if", c.Data))
						}
					case CondStateIf:
						if currCondState <= prevCondState {
							comp.RenderFunc.emitGo("\n}\n")
						}
					case CondStateElseIf:
						if currCondState < prevCondState {
							comp.RenderFunc.emitGo("\n}\n")
						}
					case CondStateElse:
						if currCondState == CondStateElseIf || currCondState == CondStateElse {
							merr.append(comp.errf("tx-else-if/tx-else on <%s> after tx-else (nothing can follow tx-else)", c.Data))
						}
						comp.RenderFunc.emitGo("\n}\n")
					}

					switch currCondState {
					case CondStateDefault:
					case CondStateIf:
						comp.RenderFunc.emitGo("if " + field + " {\n")
					case CondStateElseIf:
						comp.RenderFunc.emitGo("} else if " + field + " {\n")
					case CondStateElse:
						comp.RenderFunc.emitGo("} else {\n")
					}

					prevCondState = currCondState

					if stmt, ok := hasAttr(c, "tx-for"); ok {
						val, found := hasAttr(c, "tx-key")
						if !found {
							merr.append(comp.errf("tx-for requires a tx-key attribute"))
						} else {
							hasFor = true
							forKey = val
							comp.RenderFunc.emitGo("\nfor " + stmt + " {\n")
						}
					}
				}

				childForKeys := forKeys
				if hasFor {
					childForKeys = append(forKeys, forKey)
				}

				merr.concat(comp.parseTmpl(c, childForKeys, inSlot))

				if hasFor {
					comp.RenderFunc.emitGo("\n}\n")
				}

				if c.NextSibling == nil && (prevCondState == CondStateIf || prevCondState == CondStateElseIf || prevCondState == CondStateElse) {
					comp.RenderFunc.emitGo("\n}\n")
				}
			}
		}

		if node.DataAtom != atom.Template {
			comp.RenderFunc.emitStrLit("</")
			comp.RenderFunc.emitStrLit(node.Data)
			comp.RenderFunc.emitStrLit(">")
		}
		return merr

	}

	return nil
}

func (comp *Component) scanTmplStr(str string, onRaw func(r rune), onExpr func(expr string) error) error {
	str = strings.TrimSpace(str)
	if str == "" {
		return nil
	}

	braceStack := 0
	isInDoubleQuote := false
	isInSingleQuote := false
	isInBackQuote := false
	skipNext := false

	expr := []byte{}
	for _, r := range str {
		if skipNext {
			expr = append(expr, []byte(string(r))...)
			skipNext = false
			continue
		}

		if braceStack == 0 && r != '{' {
			onRaw(r)
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
			if isInDoubleQuote || isInSingleQuote || isInBackQuote {
				expr = append(expr, byte(r))
			} else if braceStack == 1 {
				braceStack--
				trimmedCurrExpr := bytes.TrimSpace(expr)
				if len(trimmedCurrExpr) == 0 {
					continue
				}
				if err := onExpr(string(trimmedCurrExpr)); err != nil {
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

func (comp *Component) parseUsedVarsStr(str string) {
	comp.scanTmplStr(str, func(rune) {}, func(expr string) error {
		if parsed, err := parser.ParseExpr(expr); err == nil {
			comp.markUsedVars(parsed)
		}
		return nil
	})
}

func (comp *Component) parseTmplStr(str string, escape bool) error {
	return comp.scanTmplStr(str, func(r rune) {
		s := string(r)
		if escape {
			s = html.EscapeString(s)
		}
		comp.RenderFunc.emitStrLit(s)
	}, func(expr string) error {
		if _, err := parser.ParseExpr(expr); err != nil {
			return comp.errf("invalid expression {%s}: %w", expr, err)
		}
		if escape {
			comp.RenderFunc.emitHtmlEscapeExpr(expr)
		} else {
			comp.RenderFunc.emitExpr(expr)
		}
		return nil
	})
}

type MultiError struct {
	errs []error
	mux  sync.Mutex
}

func newMultiError(errs ...error) *MultiError {
	me := &MultiError{}
	for _, err := range errs {
		if err != nil {
			me.errs = append(me.errs, err)
		}
	}
	return me
}

func (me *MultiError) append(errs ...error) {
	me.mux.Lock()
	defer me.mux.Unlock()
	for _, err := range errs {
		if err != nil {
			me.errs = append(me.errs, err)
		}
	}
}

func (me *MultiError) concat(other *MultiError) {
	if other == nil {
		return
	}
	other.mux.Lock()
	errs := make([]error, len(other.errs))
	copy(errs, other.errs)
	other.mux.Unlock()

	me.append(errs...)
}

func (me *MultiError) exitOnErrors() {
	me.mux.Lock()
	defer me.mux.Unlock()
	if len(me.errs) == 0 {
		return
	}
	log.Printf("%d error(s):\n", len(me.errs))
	for _, err := range me.errs {
		log.Println(err)
	}
	os.Exit(1)
}

type IdGen struct {
	CurrNum int
}

func newIdGen() *IdGen {
	return &IdGen{}
}

func (id *IdGen) next() string {
	id.CurrNum++
	return fmt.Sprint(id.CurrNum)
}

type SegmentType int

const (
	SegmentTypeStrLit SegmentType = iota
	SegmentTypeGo
	SegmentTypeExpr
	SegmentTypeHtmlEscapeExpr
	SegmentTypeUrlEscapeExpr
)

type Segment struct {
	Type    SegmentType
	BufName string
	Content []byte
}

func (s Segment) empty() bool { return len(s.Content) == 0 }

type Code struct {
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
	if code.PendingSegment.Type != t {
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

func (comp *Code) emitUrlEscapeExpr(content string) {
	comp.emit(SegmentTypeUrlEscapeExpr, content)
}

func (code *Code) emitSplit() {
	code.flush()
	code.PendingSegment = Segment{
		BufName: "tx_w2",
	}
}

func (code *Code) writeTo(codeBuilder *CodeBuilder) {
	code.flush()
	for _, segment := range code.Segments {
		switch segment.Type {
		case SegmentTypeGo:
			codeBuilder.WriteString(string(segment.Content))
		case SegmentTypeStrLit:
			codeBuilder.write("%s.WriteString(%s)\n", segment.BufName, strconv.Quote(string(segment.Content)))
		case SegmentTypeExpr:
			codeBuilder.write("fmt.Fprint(%s, %s)\n", segment.BufName, string(segment.Content))
		case SegmentTypeHtmlEscapeExpr:
			codeBuilder.write("%s.WriteString(html.EscapeString(fmt.Sprint(%s)))\n", segment.BufName, string(segment.Content))
		case SegmentTypeUrlEscapeExpr:
			codeBuilder.write("if param, err := json.Marshal(%s); err != nil {\nlog.Panic(err)\n} else {\n%s.WriteString(url.QueryEscape(string(param)))}\n", string(segment.Content), segment.BufName)
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

type VarType int

const (
	VarTypeState VarType = iota
	VarTypeDerived
	VarTypeProp
)

type Var struct {
	Type       VarType
	GoName     string
	SavedField string

	TypeExpr    string
	InitExprAst ast.Expr
	InitExpr    string
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
	Name  string
	Decl  *ast.FuncDecl
	Stmts string
}

type CondState int

const (
	CondStateDefault CondState = iota
	CondStateIf
	CondStateElseIf
	CondStateElse
)

type Fill struct {
	GoName     string
	CompName   string
	ParentComp *Component
	Location   string

	HasChildComps bool
	UsedVars      map[string]struct{}
	RenderFunc    Code
}

//go:embed runtime.js
var runtimeScript string

type CodeBuilder struct {
	strings.Builder
}

func (code *CodeBuilder) write(s string, params ...any) {
	if len(params) == 0 {
		code.WriteString(s)
	} else {
		fmt.Fprintf(&code.Builder, s, params...)
	}
}

func newTemplateNode() *html.Node {
	return &html.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Template,
		Data:     "template",
	}
}

func dirExist(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

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
			b.WriteString("_S_")
		case r == '-':
			b.WriteString("_H_")
		case r == '.':
			b.WriteString("_O_")
		case r == ':':
			b.WriteString("_C_")
		case r == '@':
			b.WriteString("_A_")
		case r == '!':
			b.WriteString("_B_")
		case r == '~':
			b.WriteString("_T_")
		case r == '*':
			b.WriteString("_K_")
		case r == '+':
			b.WriteString("_P_")
		case r == '=':
			b.WriteString("_E_")
		case r == '&':
			b.WriteString("_N_")
		case r == '?':
			b.WriteString("_Q_")
		case r == '#':
			b.WriteString("_F_")
		case r == '$':
			b.WriteString("_D_")
		case r == '{':
			b.WriteString("_L_")
		case r == '}':
			b.WriteString("_R_")
		default:
			fmt.Fprintf(&b, "_%X_", r)
		}
	}
	return b.String()
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

func parseFillNodes(n *html.Node) map[string]*html.Node {
	fillNodes := map[string]*html.Node{}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if slotName, found := hasAttr(c, "slot"); found {
			fillNodes[slotName] = c
			continue
		} else {
			if fillNodes[""] == nil {
				fillNodes[""] = newTemplateNode()
			}

			fillNodes[""].AppendChild(&html.Node{
				FirstChild: c.FirstChild,
				LastChild:  c.LastChild,
				Type:       c.Type,
				DataAtom:   c.DataAtom,
				Data:       c.Data,
				Namespace:  c.Namespace,
				Attr:       c.Attr,
			})
		}
	}
	return fillNodes
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
