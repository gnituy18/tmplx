package main

import (
	"errors"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)


// parse components
// comps := map[string]*Comp{}
// if err := filepath.WalkDir(dirComponents, func(path string, d fs.DirEntry, err error) error {
// 	if d.IsDir() {
// 		return nil
// 	}

// 	relPath, err := filepath.Rel(dirComponents, path)
// 	if err != nil {
// 		return err
// 	}

// 	dir, file := filepath.Split(relPath)
// 	ext := filepath.Ext(file)
// 	if ext != ".tmplx" {
// 		return nil
// 	}

// 	name := strings.ReplaceAll(filepath.Join(dir, strings.TrimSuffix(file, ext)), "/", "-")
// 	bs, err := os.ReadFile(path)
// 	if err != nil {
// 		return err
// 	}

// 	comp, err := NewComp(name, string(bs))
// 	if err != nil {
// 		return err
// 	}

// 	comps[name] = comp

// 	return nil
// }); err != nil {
// 	log.Fatal(err)
// }

type Comp struct {
	Name         string
	ScriptNode   *html.Node
	TemplateNode *html.Node
}

func NewComp(name, content string) (*Comp, error) {
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
