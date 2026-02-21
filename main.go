package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"log"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"golang.org/x/tools/imports"
)

const (
	mimeType = "text/tmplx"

	txCommentPath   = "tx:path"
	txCommentProp   = "tx:prop"
	txCommentMethod = "tx:method"

	txIgnoreKey  = "tx-ignore"
	txForKey     = "tx-for"
	txKeyKey     = "tx-key"
	txIfKey      = "tx-if"
	txElseIfKey  = "tx-else-if"
	txElseKey    = "tx-else"
	txRuntimeVal = "tx-runtime"
)

var (
	errModuleNotFound = errors.New("no go.mod found in current or parent directories")

	pagesDir          string
	componentsDir     string
	outputFilePath    string
	outputPackageName string
	handlerPrefix     string

	components = map[string]*Component{}
)

func main() {
	log.SetFlags(0)

	root, err := findModuleRoot()
	if errors.Is(err, errModuleNotFound) {
		log.Fatalf("error: %v\n", err)
	} else if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	flag.StringVar(&componentsDir, "components-dir", filepath.Join(root, "components"), "directory containing reusable components")
	flag.StringVar(&pagesDir, "pages-dir", filepath.Join(root, "pages"), "directory containing pages")
	flag.StringVar(&outputFilePath, "output-file", filepath.Join(root, "routes.go"), "path to the generated Go file")
	flag.StringVar(&outputPackageName, "package-name", "main", "package name for the generated Go code")
	flag.StringVar(&handlerPrefix, "handler-prefix", "/tx/", "path prefix for event handler URLs")
	flag.Parse()
	componentsDir = filepath.Clean(componentsDir)
	pagesDir = filepath.Clean(pagesDir)
	outputFilePath = filepath.Clean(outputFilePath)
	if !(token.IsIdentifier(outputPackageName) && !token.IsKeyword(outputPackageName)) {
		log.Fatalf("%q is not a valid Go package name\n", outputPackageName)
	}

	merr := newMultiError()
	if exist, err := dirExist(componentsDir); err != nil {
		log.Fatalf("error: %v\n", err)
	} else if !exist {
		log.Printf("no components directory at %s, skipping\n", componentsDir)
	} else {
		if err := filepath.WalkDir(componentsDir, func(filePath string, entry fs.DirEntry, err error) error {
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

			ident := strings.ReplaceAll(stemPath, "/", "-")
			for _, r := range ident {
				if !isValidComponentNameRune(r) {
					merr.append(fmt.Errorf("%s: invalid character %q in <tx-%s>: use only a-z, 0-9, -, _", filePath, r, ident))
					return nil
				}
			}

			name := "tx-" + ident

			if comp, ok := components[name]; ok {
				merr.append(fmt.Errorf("%s: duplicate component <%s>, first defined in %s", filePath, name, comp.FilePath))
				return nil
			}

			components[name] = &Component{
				Type:     CompTypeComp,
				FilePath: filePath,
				RelPath:  relPath,
				Name:     name,
				GoIdent:  goIdent(name),
			}

			return nil
		}); err != nil {
			log.Fatalf("error: %s: walk failed: %v\n", componentsDir, err)
		}
	}

	pages := []*Component{}
	pageNames := map[string]string{}
	exist, err := dirExist(pagesDir)
	if err != nil {
		log.Fatalf("error: %s: cannot access pages directory: %v\n", pagesDir, err)
	}
	if !exist {
		log.Fatalf("pages directory not found: %s\n", pagesDir)
	}

	if err := filepath.WalkDir(pagesDir, func(filePath string, entry fs.DirEntry, err error) error {
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

		if existingFile, ok := pageNames[urlPath]; ok {
			merr.append(fmt.Errorf("%s: duplicate page route %s, first defined in %s", filePath, urlPath, existingFile))
			return nil
		}

		pageNames[urlPath] = filePath
		pages = append(pages, &Component{
			Type:     CompTypePage,
			FilePath: filePath,
			RelPath:  relPath,
			Name:     urlPath,
			GoIdent:  goIdent(urlPath),
		})

		return nil
	}); err != nil {
		log.Fatalf("error: %s: walk failed: %v\n", pagesDir, err)
	}
	merr.exitOnErrors()

	var wg sync.WaitGroup
	componentNames := maps.Keys(components)
	for name := range componentNames {
		wg.Add(1)
		go func() {
			defer wg.Done()

			comp := components[name]
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

			comp.TemplateNode = &html.Node{
				Type:     html.ElementNode,
				DataAtom: atom.Template,
				Data:     "template",
			}
			for _, node := range nodes {
				val, found := hasAttr(node, "type")
				if node.DataAtom == atom.Script && found && val == mimeType {
					if comp.TmplxScriptNode != nil {
						merr.append(comp.errf("multiple <script type=\"%s\"> elements (only one allowed)", mimeType))
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

			comp.SlotNames = []string{}
			comp.Slots = map[string]struct{}{}

			merr.concat(comp.parseSlots(comp.TemplateNode, false))
		}()
	}
	wg.Wait()
	merr.exitOnErrors()

	for name := range componentNames {
		wg.Add(1)
		comp := components[name]
		go func() {
			defer wg.Done()
			comp.ChildCompsIdGen = map[string]*IdGen{}
			for childName := range componentNames {
				comp.ChildCompsIdGen[childName] = newIdGen(comp.Name + "_" + childName)
			}

			comp.AnonFuncNameGen = newIdGen("anon_func")
			comp.writeStrLit("<template id=\"")
			comp.writeExpr("tx_key")
			comp.writeStrLit("\"></template>")
			merr.concat(comp.parseTmpl(comp.TemplateNode, []string{}))
			comp.writeStrLit("<template id=\"")
			comp.writeExpr("tx_key + \"_e\"")
			comp.writeStrLit("\"></template>")

			if len(comp.CurrRenderFuncContent) > 0 {
				comp.RenderFuncCodes = append(comp.RenderFuncCodes, RenderFunc{
					Type:    comp.CurrRenderFuncType,
					Content: comp.CurrRenderFuncContent,
				})
			}
		}()
	}

	for _, page := range pages {
		wg.Add(1)
		go func() {
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
					node.AppendChild(&html.Node{
						Type:     html.ElementNode,
						DataAtom: atom.Script,
						Data:     "script",
						Attr: []html.Attribute{
							{Key: "id", Val: txRuntimeVal},
						},
					})
					node.AppendChild(txStateNode)
					break
				}
			}

			merr.concat(page.parseTmplxScript())

			page.ChildCompsIdGen = map[string]*IdGen{}
			for name := range componentNames {
				page.ChildCompsIdGen[name] = newIdGen(page.Name + "_" + name)
			}

			page.AnonFuncNameGen = newIdGen(page.GoIdent)
			merr.concat(page.parseTmpl(page.TemplateNode, []string{}))

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
	merr.exitOnErrors()

	var out strings.Builder
	out.WriteString("package " + outputPackageName + "\n")
	out.WriteString("import(\n")

	for _, page := range pages {
		for _, im := range page.Imports {
			if _, err := out.WriteString(astToSource(im) + "\n"); err != nil {
				log.Fatalln(fmt.Errorf("write imports: %w", err))
			}
		}
	}

	out.WriteString(")\n")
	out.WriteString("var runtimeScript = `" + strings.Replace(runtimeScript, "TX_HANDLER_PREFIX", handlerPrefix, 1) + "`\n")
	out.WriteString(`
type TxRoute struct {
        Pattern	string
	Handler	http.HandlerFunc
}
`)
	for name := range componentNames {
		comp := components[name]
		out.WriteString(fmt.Sprintf("type state_%s struct {\n", comp.GoIdent))
		for _, varName := range comp.VarNames {
			v := comp.Vars[varName]
			if v.Type == VarTypeState || v.Type == VarTypeProp {
				out.WriteString(fmt.Sprintf("%s %s `json:\"%s\"`\n", v.StructField, astToSource(v.TypeExpr), v.Name))
			}
		}
		out.WriteString("}\n")

		paramsStr := ""
		for _, varName := range comp.VarNames {
			v := comp.Vars[varName]
			paramsStr += fmt.Sprintf(", %s %s", v.Name, astToSource(v.TypeExpr))
		}
		for _, funcName := range comp.FuncNames {
			f := comp.Funcs[funcName]
			paramsStr += fmt.Sprintf(", %s, %s string", f.Name, f.Name+"_swap")
		}
		for _, slotName := range comp.SlotNames {
			if slotName != "" {
				paramsStr += fmt.Sprintf(",tx_render_slot_%s func()", slotName)
			} else {
				paramsStr += ",tx_render_default_slot func()"
			}
		}

		out.WriteString(fmt.Sprintf("func render_%s(tx_w *bytes.Buffer, tx_key string, tx_states map[string]string, tx_newStates map[string]any  %s) {\n", comp.GoIdent, paramsStr))
		comp.implRenderFunc(&out)
		out.WriteString("}\n")
	}

	for _, page := range pages {
		out.WriteString(fmt.Sprintf("type state_%s struct {\n", page.GoIdent))
		for _, varName := range page.VarNames {
			v := page.Vars[varName]
			if v.Type == VarTypeState {
				out.WriteString(fmt.Sprintf("%s %s `json:\"%s\"`\n", v.StructField, astToSource(v.TypeExpr), v.Name))
			}
		}
		out.WriteString("}\n")

		params := []string{}
		for _, varName := range page.VarNames {
			v := page.Vars[varName]
			params = append(params, fmt.Sprintf("%s %s", v.Name, astToSource(v.TypeExpr)))
		}
		for _, funcName := range page.FuncNames {
			f := page.Funcs[funcName]
			params = append(params, fmt.Sprintf("%s, %s string", f.Name, f.Name+"_swap"))
		}
		out.WriteString(fmt.Sprintf("func render_%s(tx_w *bytes.Buffer, tx_key string, tx_states map[string]string, tx_newStates map[string]any, %s) {\n", page.GoIdent, strings.Join(params, ", ")))
		page.implRenderFunc(&out)
		out.WriteString("}\n")
	}

	out.WriteString("var txRoutes []TxRoute = []TxRoute{\n")
	for _, page := range pages {
		out.WriteString("{\n")
		out.WriteString("Pattern: \"GET " + page.Name + "\",\n")
		out.WriteString("Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {\n")
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
				if v.Type == VarTypeDerived {
					out.WriteString(fmt.Sprintf("%s = %s\n", name, astToSource(v.InitExpr)))
				}
			}
		}
		out.WriteString(fmt.Sprintf("tx_state := &state_%s{\n", page.GoIdent))
		for _, name := range page.VarNames {
			v := page.Vars[name]
			if v.Type == VarTypeState {
				out.WriteString(fmt.Sprintf("%s: %s,\n", v.StructField, v.Name))
			}
		}
		out.WriteString("}\n")
		out.WriteString("tx_newStates := map[string]any{}\n")
		out.WriteString("tx_newStates[\"tx_\"] = tx_state\n")
		out.WriteString("var tx_buf bytes.Buffer\n")
		out.WriteString(fmt.Sprintf("render_%s(&tx_buf, \"tx_\", map[string]string{}, tx_newStates", page.GoIdent))
		for _, name := range page.VarNames {
			out.WriteString(fmt.Sprintf(", %s", name))
		}
		for _, name := range page.FuncNames {
			out.WriteString(fmt.Sprintf(", \"%s\", \"tx_\"", page.funcId(name)))
		}
		out.WriteString(")\n")
		out.WriteString("tx_stateBytes, _ := json.Marshal(tx_newStates)\n")
		out.WriteString("tx_w.Write(bytes.Replace(tx_buf.Bytes(), []byte(\"TX_STATE_JSON\"), tx_stateBytes, 1))\n")
		out.WriteString("},\n")
		out.WriteString("},\n")
		for _, funcName := range page.FuncNames {
			if funcName == "init" {
				continue
			}

			f := page.Funcs[funcName]
			out.WriteString("{\n")
			out.WriteString(fmt.Sprintf("Pattern: \"%s %s%s\",\n", f.Method, handlerPrefix, page.funcId(funcName)))
			out.WriteString("Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {\n")
			out.WriteString("tx_query := tx_r.URL.Query()\n")
			out.WriteString("tx_states := map[string]string{}\n")
			out.WriteString("for k, v := range tx_query {\n")
			out.WriteString("if strings.HasPrefix(k, \"tx_\") {\n")
			out.WriteString("tx_states[k] = v[0]\n")
			out.WriteString("}\n")
			out.WriteString("}\n")
			out.WriteString("tx_newStates := map[string]any{}\n")
			out.WriteString(fmt.Sprintf("tx_state := &state_%s{}\n", page.GoIdent))
			out.WriteString("json.Unmarshal([]byte(tx_states[\"tx_\"]), &tx_state)\n")
			for _, name := range page.VarNames {
				v := page.Vars[name]
				if v.Type == VarTypeState {
					out.WriteString(fmt.Sprintf("%s := tx_state.%s\n", v.Name, v.StructField))
				} else if v.Type == VarTypeDerived {
					out.WriteString(fmt.Sprintf("%s := %s\n", v.Name, astToSource(v.InitExpr)))
				}
			}
			for _, list := range f.Decl.Type.Params.List {
				for _, ident := range list.Names {
					out.WriteString(fmt.Sprintf("var %s %s\n", ident.Name, astToSource(list.Type)))
					out.WriteString(fmt.Sprintf("json.Unmarshal([]byte(tx_query.Get(\"%s\")), &%s)\n", ident.Name, ident.Name))
				}
			}
			for _, stmt := range f.Decl.Body.List {
				out.WriteString(astToSource(stmt) + "\n")
			}
			for _, name := range page.VarNames {
				v := page.Vars[name]
				if v.Type == VarTypeDerived {
					out.WriteString(fmt.Sprintf("%s = %s\n", v.Name, astToSource(v.InitExpr)))
				}
			}
			out.WriteString("var tx_buf bytes.Buffer\n")
			out.WriteString(fmt.Sprintf("render_%s(&tx_buf, \"tx_\", tx_states, tx_newStates", page.GoIdent))
			for _, name := range page.VarNames {
				out.WriteString(fmt.Sprintf(", %s", name))
			}
			for _, name := range page.FuncNames {
				out.WriteString(fmt.Sprintf(", \"%s\", \"tx_\"", page.funcId(name)))
			}
			out.WriteString(")\n")
			out.WriteString(fmt.Sprintf("tx_newStates[\"tx_\"] = &state_%s{\n", page.GoIdent))
			for _, name := range page.VarNames {
				v := page.Vars[name]
				if v.Type == VarTypeState {
					out.WriteString(fmt.Sprintf("%s: %s,\n", v.StructField, v.Name))
				}
			}
			out.WriteString("}\n")
			out.WriteString("tx_stateBytes, _ := json.Marshal(tx_newStates)\n")
			out.WriteString("tx_w.Write(bytes.Replace(tx_buf.Bytes(), []byte(\"TX_STATE_JSON\"), tx_stateBytes, 1))\n")
			out.WriteString("},\n")
			out.WriteString("},\n")
		}
	}
	for name := range componentNames {
		comp := components[name]
		for _, funcName := range comp.FuncNames {
			if funcName == "init" {
				continue
			}

			f := comp.Funcs[funcName]
			if f.Decl.Body == nil {
				continue
			}

			out.WriteString("{\n")
			out.WriteString(fmt.Sprintf("Pattern: \"%s %s%s\",\n", f.Method, handlerPrefix, comp.funcId(funcName)))
			out.WriteString("Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {\n")
			out.WriteString("tx_w.Header().Set(\"Content-Type\", \"text/html\")\n")
			out.WriteString("tx_query := tx_r.URL.Query()\n")
			out.WriteString("tx_swap := tx_query.Get(\"tx-swap\")\n")
			out.WriteString("tx_states := map[string]string{}\n")
			out.WriteString("for k, v := range tx_query {\n")
			out.WriteString("if strings.HasPrefix(k, tx_swap) {\n")
			out.WriteString("tx_states[k] = v[0]\n")
			out.WriteString("}\n")
			out.WriteString("}\n")
			out.WriteString("tx_newStates := map[string]any{}\n")
			out.WriteString(fmt.Sprintf("tx_state := &state_%s{}\n", comp.GoIdent))
			out.WriteString("json.Unmarshal([]byte(tx_states[tx_swap]), &tx_state)\n")
			for _, name := range comp.VarNames {
				v := comp.Vars[name]
				if v.Type == VarTypeState || v.Type == VarTypeProp {
					out.WriteString(fmt.Sprintf("%s := tx_state.%s\n", v.Name, v.StructField))
				} else if v.Type == VarTypeDerived {
					out.WriteString(fmt.Sprintf("%s := %s\n", v.Name, astToSource(v.InitExpr)))
				}
			}
			for _, list := range f.Decl.Type.Params.List {
				for _, ident := range list.Names {
					out.WriteString(fmt.Sprintf("var %s %s\n", ident.Name, astToSource(list.Type)))
					out.WriteString(fmt.Sprintf("json.Unmarshal([]byte(tx_query.Get(\"%s\")), &%s)\n", ident.Name, ident.Name))
				}
			}

			for _, stmt := range f.Decl.Body.List {

				out.WriteString(astToSource(stmt) + "\n")
			}
			for _, name := range comp.VarNames {
				v := comp.Vars[name]
				if v.Type == VarTypeDerived {
					out.WriteString(fmt.Sprintf("%s = %s\n", v.Name, astToSource(v.InitExpr)))
				}
			}

			out.WriteString("var tx_buf bytes.Buffer\n")
			out.WriteString(fmt.Sprintf("render_%s(&tx_buf, tx_swap, tx_states, tx_newStates", comp.GoIdent))
			for _, name := range comp.VarNames {
				out.WriteString(fmt.Sprintf(", %s", name))
			}
			for _, name := range comp.FuncNames {
				out.WriteString(fmt.Sprintf(", \"%s\", tx_swap", comp.funcId(name)))
			}
			out.WriteString(")\n")
			out.WriteString("tx_w.Write(tx_buf.Bytes())\n")
			out.WriteString("tx_w.Write([]byte(\"<script id=\\\"tx-state\\\" type=\\\"application/json\\\">\"))\n")
			out.WriteString(fmt.Sprintf("tx_newStates[tx_swap] = &state_%s{\n", comp.GoIdent))
			for _, name := range comp.VarNames {
				v := comp.Vars[name]
				if v.Type == VarTypeState || v.Type == VarTypeProp {
					out.WriteString(fmt.Sprintf("%s: %s,\n", v.StructField, v.Name))
				}
			}
			out.WriteString("}\n")
			out.WriteString("tx_stateBytes, _ := json.Marshal(tx_newStates)\n")
			out.WriteString("tx_w.Write(tx_stateBytes)\n")
			out.WriteString("tx_w.Write([]byte(\"</script>\"))\n")
			out.WriteString("},\n")
			out.WriteString("},\n")
		}
	}
	out.WriteString("}\n")

	out.WriteString("func Routes() []TxRoute { return txRoutes }\n\n")

	data := []byte(out.String())
	formatted, err := imports.Process(outputFilePath, data, nil)
	if err != nil {
		scanner := bufio.NewScanner(bytes.NewReader(data))
		lineNum := 1
		for scanner.Scan() {
			log.Printf("%d: %s\n", lineNum, scanner.Text())
			lineNum++
		}
		log.Fatalln(fmt.Errorf("format generated code: %w", err))
	}

	if outputFilePath == "" {
		if _, err := os.Stdout.Write(formatted); err != nil {
			log.Fatalln(err)
		}
	} else {
		outputFilePath = filepath.Clean(outputFilePath)
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
		log.Printf("%s generated successfully (%d pages, %d components)\n", outputFilePath, len(pages), len(components))
	}

}

type CompType int

const (
	CompTypePage = iota + 1
	CompTypeComp
)

type Component struct {
	Type     CompType
	FilePath string
	RelPath  string

	Name    string
	GoIdent string

	TmplxScriptNode  *html.Node
	Imports          []*ast.ImportSpec
	VarNames         []string
	Vars             map[string]*Var
	SavedVarsLen     int
	FuncNames        []string
	Funcs            map[string]*Func
	AnonFuncNameGen  *IdGen
	InputFuncHandler bool

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
	return fmt.Errorf(comp.RelPath+": "+msg, a...)
}

func (comp *Component) parseTmplxScript() *MultiError {
	merr := newMultiError()

	comp.Imports = []*ast.ImportSpec{}
	comp.VarNames = []string{}
	comp.Vars = map[string]*Var{}
	comp.FuncNames = []string{}
	comp.Funcs = map[string]*Func{}

	if comp.TmplxScriptNode != nil {
		// TODO: save position into errors
		scriptAst, err := parser.ParseFile(token.NewFileSet(), "", "package p\n"+comp.TmplxScriptNode.FirstChild.Data, parser.ParseComments)
		if err != nil {
			merr.append(comp.errf("syntax error in <script type=\"text/tmplx\">: %w", err))
			return merr
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
					comp.VarNames = append(comp.VarNames, ident.Name)
					comp.Vars[ident.Name] = &Var{
						Name:        ident.Name,
						StructField: "S_" + ident.Name,
						TypeExpr:    s.Type,
					}

					isProp := false
					isPath := false

					if d.Doc != nil {
						comments := []Comment{}
						for _, comment := range d.Doc.List {
							comments = append(comments, parseComments(comment.Text)...)
						}

						for _, comment := range comments {
							if comment.Name == CommentProp {
								isProp = true
							} else if comment.Name == CommentPath {
								isPath = true
								comp.Vars[ident.Name].InitExpr = &ast.CallExpr{
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
							comp.Vars[ident.Name].InitExpr = s.Values[0]
						}
						comp.Vars[ident.Name].Type = VarTypeProp
					} else if isPath {
						if len(s.Values) > 0 {
							merr.append(comp.errf("//tx:path variable cannot have an initial value: %s", astToSource(spec)))
						}

						if astToSource(s.Type) != "string" {
							merr.append(comp.errf("//tx:path variable must be type string: %s", astToSource(spec)))
						}

						comp.Vars[ident.Name].Type = VarTypeState
					} else if len(s.Values) == 1 || len(s.Values) == 0 {

						found := false
						if len(s.Values) == 1 {
							v := s.Values[0]
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
							comp.Vars[ident.Name].InitExpr = v
						}

						if found {
							comp.Vars[ident.Name].Type = VarTypeDerived
						} else {
							comp.Vars[ident.Name].Type = VarTypeState
						}

					} else if len(s.Values) > 1 {
						merr.append(comp.errf("declare one variable per var statement: %s", astToSource(spec)))
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
				merr.append(comp.errf("%s: methods (func with receiver) not allowed, use plain functions", d.Name))
			}

			if d.Type.Results != nil {
				merr.append(comp.errf("%s: return values not allowed", d.Name))
			}

			for _, field := range d.Type.Params.List {
				for _, name := range field.Names {
					if comp.Vars[name.Name] != nil {
						merr.append(comp.errf("%s: parameter %s shadows a state variable", d.Name, name.Name))
					}
				}
			}

			if d.Body != nil {
				deriveds := []string{}
				comp.modifiedDerived(d.Body, &deriveds)
				for _, derived := range deriveds {
					merr.append(comp.errf("%s: cannot assign to %s (derived/prop is read-only)", d.Name, derived))
				}
			}

			method := http.MethodGet
			if d.Doc != nil {
				for _, list := range d.Doc.List {
					comments := parseComments(list.Text)
					for _, comment := range comments {
						if comment.Name == txCommentMethod {
							method = comment.Value
						}
					}
				}
			}

			comp.FuncNames = append(comp.FuncNames, d.Name.Name)
			comp.Funcs[d.Name.Name] = &Func{
				Method: method,
				Name:   d.Name.Name,
				Decl:   d,
			}
		}
	}

	for _, v := range comp.Vars {
		if v.Type == VarTypeState || v.Type == VarTypeProp {
			comp.SavedVarsLen++
		}
	}

	return merr
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

				if v.Type != VarTypeDerived && v.Type != VarTypeProp {
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

func (comp *Component) parseTmpl(node *html.Node, forKeys []string) *MultiError {
	merr := newMultiError()
	switch node.Type {
	case html.CommentNode:
		return nil
	case html.DocumentNode:
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			merr.concat(comp.parseTmpl(c, forKeys))
		}
		return merr
	case html.DoctypeNode:
		comp.writeStrLit("<!DOCTYPE ")
		comp.writeStrLit(node.Data)
		comp.writeStrLit(">")
		return nil
	case html.TextNode:
		_, hasTxIgnore := hasAttr(node.Parent, txIgnoreKey)
		isScriptOrStyle := node.Parent.DataAtom == atom.Script || node.Parent.DataAtom == atom.Style
		isRawText := isChildNodeRawText(node.Parent.Data)

		// Skip interpolation for tx-ignore or script/style tags
		if hasTxIgnore || isScriptOrStyle {
			if isRawText {
				comp.writeStrLit(node.Data)
			} else {
				comp.writeStrLit(html.EscapeString(node.Data))
			}
			return nil
		}

		// Parse with escaping based on whether parent is raw text
		return newMultiError(comp.parseTmplStr(node.Data, !isRawText))
	case html.ElementNode:
		// handle component
		if components[node.Data] != nil {
			childComp := components[node.Data]

			// Validate that attributes on the component tag are actual props, not state variables
			for _, attr := range node.Attr {
				if attr.Key == txIfKey || attr.Key == txElseIfKey || attr.Key == txElseKey || attr.Key == txForKey || attr.Key == txKeyKey || attr.Key == "slot" {
					continue
				}
				if _, ok := childComp.Funcs[attr.Key]; ok {
					continue
				}

				v, ok := childComp.Vars[attr.Key]
				if !ok {
					merr.append(comp.errf("<%s %s=\"...\">: %s is not a prop or function in component %s", node.Data, attr.Key, attr.Key, childComp.Name))
					continue
				}

				if v.Type != VarTypeProp {
					merr.append(comp.errf("<%s %s=\"...\">: %s is a state variable, not a prop; use //tx:prop to declare it as a prop", node.Data, attr.Key, attr.Key))
					continue
				}

				parentVar, ok := comp.Vars[attr.Val]
				if !ok {
					continue
				}

				// Type check: if attr.Val is a variable in parent, compare types
				propType := astToSource(v.TypeExpr)
				parentType := astToSource(parentVar.TypeExpr)
				if propType != parentType {
					merr.append(comp.errf("<%s %s=\"%s\">: type mismatch; prop %s expects %s but got %s", node.Data, attr.Key, attr.Val, attr.Key, propType, parentType))
					continue
				}
			}

			id := comp.ChildCompsIdGen[childComp.Name].next()

			comp.writeGo("{\n")
			// create key for component
			if len(forKeys) > 0 {
				comp.writeGo("tx_key := tx_key")
				for _, key := range forKeys {
					comp.writeGo(` + "-" + fmt.Sprint(` + key + ")")
				}
				comp.writeGo("\n")
			}
			comp.writeGo(fmt.Sprintf("tx_ckey := tx_key + \"_%s\"\n", id))
			comp.writeGo(fmt.Sprintf("tx_state := &state_%s{}\n", childComp.GoIdent))
			comp.writeGo("tx_old_state, tx_old_state_exist := tx_states[tx_ckey]\n")
			comp.writeGo("if tx_old_state_exist {\n")
			comp.writeGo("json.Unmarshal([]byte(tx_old_state), tx_state)\n")
			comp.writeGo("}\n")
			for _, varName := range childComp.VarNames {
				v := childComp.Vars[varName]
				if v.Type == VarTypeProp {
					if val, found := hasAttr(node, varName); found {
						comp.writeGo(fmt.Sprintf("%s := %s\n", v.Name, val))
					} else if v.InitExpr != nil {
						comp.writeGo(fmt.Sprintf("%s := %s\n", v.Name, astToSource(v.InitExpr)))
					} else {
						comp.writeGo(fmt.Sprintf("var %s %s\n", v.Name, astToSource(v.TypeExpr)))
					}
					comp.writeGo(fmt.Sprintf("tx_state.%s = %s\n", v.StructField, v.Name))
				}
			}

			if childComp.SavedVarsLen > 0 {
				initStrs := ""
				for _, varName := range childComp.VarNames {
					v := childComp.Vars[varName]
					if v.Type == VarTypeState {
						if v.InitExpr != nil {
							initStrs += fmt.Sprintf("tx_state.%s = %s\n", v.StructField, astToSource(v.InitExpr))
						}
					}
				}

				if initStrs != "" {
					comp.writeGo("if !tx_old_state_exist {\n")
					comp.writeGo(initStrs)
					comp.writeGo("}\n")
				}

				for _, varName := range childComp.VarNames {
					v := childComp.Vars[varName]
					if v.Type == VarTypeState {
						comp.writeGo(fmt.Sprintf("%s := tx_state.%s\n", v.Name, v.StructField))
					}
				}
			}

			for _, varName := range childComp.VarNames {
				v := childComp.Vars[varName]
				if v.Type == VarTypeDerived {
					comp.writeGo(fmt.Sprintf("%s := %s\n", v.Name, astToSource(v.InitExpr)))
				}
			}

			if f, ok := childComp.Funcs["init"]; ok {
				comp.writeGo("if !tx_old_state_exist {\n")
				for _, stmt := range f.Decl.Body.List {
					comp.writeGo(astToSource(stmt) + "\n")
				}
				for _, name := range childComp.VarNames {
					v := childComp.Vars[name]
					if v.Type == VarTypeDerived {
						comp.writeGo(fmt.Sprintf("%s = %s\n", name, astToSource(v.InitExpr)))
					}
				}
				for _, name := range childComp.VarNames {
					v := childComp.Vars[name]
					if v.Type == VarTypeState {
						comp.writeGo(fmt.Sprintf("tx_state.%s = %s\n", v.StructField, v.Name))
					}
				}
				comp.writeGo("}\n")
			}
			comp.writeGo("tx_newStates[tx_ckey] = tx_state\n")
			params := []string{}
			for _, varName := range childComp.VarNames {
				params = append(params, varName)
			}

			for _, funcName := range childComp.FuncNames {
				if val, found := hasAttr(node, funcName); found {
					if f, ok := comp.Funcs[val]; ok {
						params = append(params, f.Name, f.Name+"_swap")
					} else {
						merr.append(comp.errf("undefined function: %s", val))
					}
				} else {
					f := childComp.Funcs[funcName]
					if f.Decl.Body == nil {
						merr.append(comp.errf("function %s has no body in %s and must be passed as a prop", funcName, childComp.Name))
					} else {
						params = append(params, fmt.Sprintf("\"%s\"", childComp.funcId(f.Name)), "tx_ckey")
					}
				}
			}

			comp.writeGo(fmt.Sprintf("render_%s(tx_w, tx_ckey, tx_states, tx_newStates", childComp.GoIdent))
			for _, param := range params {
				comp.writeGo(", " + param)
			}

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
					merr.concat(comp.parseTmpl(n, forKeys))
					comp.writeGo("\n},\n")
				} else {
					comp.writeGo("nil,\n")
				}
			}
			comp.writeGo(")\n")
			comp.writeGo("}\n")
			return merr
		}

		// handle slot
		if node.DataAtom == atom.Slot {
			renderSlotFuncName := "tx_render_default_slot"
			if name, found := hasAttr(node, "name"); found {
				renderSlotFuncName = "tx_render_slot_" + name
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
				merr.concat(comp.parseTmpl(children, forKeys))
			} else {
				comp.writeStrLit(" ")
			}

			comp.writeGo("\n}\n")
			return merr
		}

		// handle non-template tags
		if node.DataAtom != atom.Template {
			comp.writeStrLit("<")
			comp.writeStrLit(node.Data)

			_, isIgnore := hasAttr(node, txIgnoreKey)

			for _, attr := range node.Attr {
				if attr.Key == txIfKey || attr.Key == txElseIfKey || attr.Key == txElseKey || attr.Key == txForKey {
					continue
				}

				comp.writeStrLit(" ")
				if strings.HasPrefix(attr.Key, "tx-on") {
					comp.writeStrLit(attr.Key)
					comp.writeStrLit(`="`)

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
										merr.append(comp.errf("wrong number of arguments: %s", astToSource(callExpr)))
										continue
									}

									comp.writeExpr(fun.Name)
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
											merr.append(comp.errf("cannot pass state/derived variable as event handler argument: %s", callExpr.Args[i]))
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
									comp.writeStrLit(" tx-swap=\"")
									comp.writeExpr(fun.Name + "_swap")
									comp.writeStrLit(`"`)

									continue
								}
							}
						}
					}

					funcName := comp.AnonFuncNameGen.next()
					fileAst, err := parser.ParseFile(token.NewFileSet(), comp.FilePath, fmt.Sprintf("package p\nfunc %s() {\n%s\n}", funcName, attr.Val), 0)
					if err != nil {
						merr.append(comp.errf("invalid inline handler: %s", attr.Val))
						continue
					}

					decl, ok := fileAst.Decls[0].(*ast.FuncDecl)
					if !ok {
						merr.append(comp.errf("invalid inline handler: %s", attr.Val))
						continue
					}

					modifiedDerived := []string{}
					comp.modifiedDerived(decl, &modifiedDerived)
					if len(modifiedDerived) > 0 {
						merr.append(comp.errf("cannot assign to derived/prop variable in handler: %v", modifiedDerived))
						continue
					}

					comp.FuncNames = append(comp.FuncNames, decl.Name.Name)
					comp.Funcs[decl.Name.Name] = &Func{
						Method: http.MethodGet,
						Name:   decl.Name.Name,
						Decl:   decl,
					}

					comp.writeStrLit(comp.funcId(decl.Name.Name))
					comp.writeStrLit("\"")

					comp.writeStrLit(" tx-swap=\"")
					comp.writeExpr(decl.Name.Name + "_swap")
					comp.writeStrLit("\"")

				} else if attr.Key == "tx-value" {
					if comp.Vars[attr.Val] == nil {
						merr.append(comp.errf("undefined variable: %s", attr.Val))
						continue
					}

					comp.writeStrLit(fmt.Sprintf("tx-value=\"%s\"", attr.Val))

					comp.writeStrLit(" tx-swap=\"")
					comp.writeExpr("tx_key")
					comp.writeStrLit(`"`)

					comp.writeStrLit("value=\"")
					comp.writeExpr(attr.Val)
					comp.writeStrLit("\"")
				} else {
					if attr.Namespace != "" {
						comp.writeStrLit(node.Namespace)
						comp.writeStrLit(":")
					}
					comp.writeStrLit(attr.Key)
					comp.writeStrLit(`="`)
					if isIgnore {
						comp.writeStrLit(attr.Val)
					} else if err := comp.parseTmplStr(attr.Val, false); err != nil {
						merr.append(comp.errf("invalid expression in attribute: %s", attr.Val))
					}
					comp.writeStrLit(`"`)
				}
			}

			// https://html.spec.whatwg.org/#void-elements
			if isVoidElement(node.Data) {
				if node.FirstChild != nil {
					merr.append(comp.errf("void element <%s> cannot have children", node.Data))
				}

				comp.writeStrLit("/>")
				return merr
			}

			comp.writeStrLit(">")
		}

		// 0: no control flow
		// 1: if
		// 2: else-if
		// 3: else
		if val, found := hasAttr(node, "id"); node.DataAtom == atom.Script && found && val == txRuntimeVal {
			comp.writeGo("tx_w.WriteString(runtimeScript)\n")
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
							comp.writeGo("\n}\n")
						}
					case CondStateElseIf:
						if currCondState < prevCondState {
							comp.writeGo("\n}\n")
						}
					case CondStateElse:
						if currCondState == CondStateElseIf || currCondState == CondStateElse {
							merr.append(comp.errf("tx-else-if/tx-else on <%s> after tx-else (nothing can follow tx-else)", c.Data))
						}
						comp.writeGo("\n}\n")
					}

					switch currCondState {
					case CondStateDefault:
					case CondStateIf:
						comp.writeGo("if " + field + " {\n")
					case CondStateElseIf:
						comp.writeGo("} else if " + field + " {\n")
					case CondStateElse:
						comp.writeGo("} else {\n")
					}

					prevCondState = currCondState

					if stmt, ok := hasAttr(c, txForKey); ok {
						val, found := hasAttr(c, txKeyKey)
						if !found {
							merr.append(comp.errf("tx-for requires a tx-key attribute"))
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

				merr.concat(comp.parseTmpl(c, childForKeys))

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
		return merr

	}

	return nil
}

func (comp *Component) parseTmplStr(str string, escape bool) error {
	str = "'" + str + "'"
	str = strings.Join(strings.Fields(str), " ")
	str = str[1 : len(str)-1]

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
					return comp.errf("invalid expression {%s}: %w", string(trimmedCurrExpr), err)
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
		return comp.errf("unclosed quote in: %s", str)
	}
	if braceStack != 0 {
		return comp.errf("unclosed { in: %s", str)
	}

	return nil
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

		if _, ok := comp.Slots[slotName]; !ok {
			comp.SlotNames = append(comp.SlotNames, slotName)
			comp.Slots[slotName] = struct{}{}
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

func (comp *Component) implRenderFunc(out *strings.Builder) {
	codes := comp.RenderFuncCodes
	for i := 0; i < len(codes); i++ {
		tmpl := codes[i]
		switch tmpl.Type {
		case RenderFuncTypeGo:
			if _, err := out.WriteString(string(tmpl.Content)); err != nil {
				log.Fatalln(err)
			}
		case RenderFuncTypeStrLit:
			combined := string(tmpl.Content)
			for i+1 < len(codes) && codes[i+1].Type == RenderFuncTypeStrLit {
				i++
				combined += string(codes[i].Content)
			}
			if _, err := fmt.Fprintf(out, "tx_w.WriteString(%s)\n", strconv.Quote(combined)); err != nil {
				log.Fatalln(err)
			}
		case RenderFuncTypeExpr:
			if _, err := fmt.Fprintf(out, "fmt.Fprint(tx_w, %s)\n", string(tmpl.Content)); err != nil {
				log.Fatalln(err)
			}
		case RenderFuncTypeHtmlEscapeExpr:
			if _, err := fmt.Fprintf(out, "tx_w.WriteString(html.EscapeString(fmt.Sprint(%s)))\n", string(tmpl.Content)); err != nil {
				log.Fatalln(err)
			}
		case RenderFuncTypeUrlEscapeExpr:
			if _, err := fmt.Fprintf(out, "if param, err := json.Marshal(%s); err != nil {\nlog.Panic(err)\n} else {\ntx_w.WriteString(url.QueryEscape(string(param)))}\n", string(tmpl.Content)); err != nil {
				log.Fatalln(err)
			}
		}
	}
}

type VarType int

const (
	VarTypeState = iota + 1
	VarTypeDerived
	VarTypeProp
)

type Var struct {
	Name        string
	StructField string
	Type        VarType
	TypeExpr    ast.Expr
	InitExpr    ast.Expr
}

type Func struct {
	Method string
	Name   string
	Decl   *ast.FuncDecl
}

//go:embed runtime.js
var runtimeScript string

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

func (comp *Component) funcId(funcName string) string {
	return comp.GoIdent + "_" + funcName
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

func hasAttr(n *html.Node, str string) (string, bool) {
	for _, attr := range n.Attr {
		if attr.Key == str {
			return attr.Val, true
		}
	}

	return "", false
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

func newIdGen(prefix string) *IdGen {
	return &IdGen{
		Prefix: prefix,
	}
}

type IdGen struct {
	Curr   int
	Prefix string
}

func (id *IdGen) next() string {
	id.Curr++
	return fmt.Sprintf("%s_%d", id.Prefix, id.Curr)
}

type CommentName string

const (
	CommentMethod CommentName = "method"
	CommentPath   CommentName = "path"
	CommentProp   CommentName = "prop"
)

type Comment struct {
	Name  CommentName
	Value string
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
		if str == txCommentProp {
			comments = append(comments, Comment{
				Name: CommentProp,
			})
		} else if strings.HasPrefix(str, txCommentPath) {
			val := strings.TrimSpace(str[len(txCommentPath):])
			comments = append(comments, Comment{
				Name:  CommentPath,
				Value: val,
			})
		} else if strings.HasPrefix(str, txCommentMethod) {
			val := strings.TrimSpace(str[len(txCommentMethod):])
			comments = append(comments, Comment{
				Name:  txCommentMethod,
				Value: val,
			})
		}
	}

	return comments
}

func isValidComponentNameRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_'
}

func findModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errModuleNotFound
		}
		dir = parent
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
