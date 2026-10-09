package main

import (
	"flag"
	"fmt"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/gnituy18/tmplx/compiler"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("tmplx: ")

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	root := cwd
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			log.Fatalln("no go.mod found in current or parent directories (run inside a Go module, or create one with 'go mod init')")
		}
		root = parent
	}

	componentsDir := flag.String("components-dir", filepath.Join(root, "components"), "directory containing reusable components")
	pagesDir := flag.String("pages-dir", filepath.Join(root, "pages"), "directory containing pages")
	outputFile := flag.String("output-file", filepath.Join(root, "routes.go"), "path to the generated Go file")
	packageName := flag.String("package-name", "main", "package name for the generated Go code")
	handlerPrefix := flag.String("handler-prefix", "/tx/", "path prefix for event handler URLs")
	flag.Parse()
	if !token.IsIdentifier(*packageName) || *packageName == "_" {
		log.Fatalf("%q is not a valid Go package name", *packageName)
	}

	compDir := filepath.Clean(*componentsDir)
	pageDir := filepath.Clean(*pagesDir)
	compDisplay := filepath.ToSlash(compDir)
	if rel, err := filepath.Rel(cwd, compDir); err == nil {
		compDisplay = filepath.ToSlash(rel)
	}
	pageDisplay := filepath.ToSlash(pageDir)
	if rel, err := filepath.Rel(cwd, pageDir); err == nil {
		pageDisplay = filepath.ToSlash(rel)
	}
	c := &compiler.Compiler{
		PackageName:   *packageName,
		Importer:      compiler.PackageImporter(root),
		HandlerPrefix: *handlerPrefix,
		ComponentsDir: compDisplay,
		PagesDir:      pageDisplay,
	}

	if info, err := os.Stat(compDir); err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("%s: %v", compDir, err)
		}
	} else if !info.IsDir() {
		log.Fatalf("%s is not a directory", compDir)
	} else if err := walkHTML(compDir, c.NewComponent); err != nil {
		log.Fatalln(err)
	}

	pageCount := 0
	if info, err := os.Stat(pageDir); err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("pages directory not found: %s", pageDir)
		}
		log.Fatalf("%s: %v", pageDir, err)
	} else if !info.IsDir() {
		log.Fatalf("%s is not a directory", pageDir)
	} else if err := walkHTML(pageDir, func(rel string, content []byte) {
		c.NewPage(rel, content)
		pageCount++
	}); err != nil {
		log.Fatalln(err)
	}
	if pageCount == 0 {
		log.Printf("warning: no pages found in %s", pageDir)
	}

	code, err := c.Compile()
	if err != nil {
		// bare, not log.Fatalln: parsers expect diagnostics to lead with file:line:col
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	out := filepath.Clean(*outputFile)
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		log.Fatalln(err)
	}
	if err := os.WriteFile(out, code, 0644); err != nil {
		log.Fatalln(err)
	}
}

func walkHTML(dir string, add func(rel string, body []byte)) error {
	return filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".html" {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		add(filepath.ToSlash(rel), content)
		return nil
	})
}
