package main

import (
	"errors"
	"flag"
	"go/scanner"
	"go/token"
	"go/types"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnituy18/tmplx/compiler"
	"golang.org/x/tools/go/packages"
)

func main() {
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

	componentsDir := flag.String("components-dir", filepath.Join(dir, "components"), "directory containing reusable components")
	pagesDir := flag.String("pages-dir", filepath.Join(dir, "pages"), "directory containing pages")
	outputFile := flag.String("output-file", filepath.Join(dir, "routes.go"), "path to the generated Go file")
	packageName := flag.String("package-name", "main", "package name for the generated Go code")
	handlerPrefix := flag.String("handler-prefix", "/tx/", "path prefix for event handler URLs")
	flag.Parse()
	if !token.IsIdentifier(*packageName) || token.IsKeyword(*packageName) {
		log.Fatalf("\"%s\" is not a valid Go package name\n", *packageName)
	}

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedImports | packages.NeedDeps,
		Dir:  dir,
	}, "./...")
	if err != nil {
		log.Fatalf("error: load packages: %v\n", err)
	}
	loaded := map[string]*types.Package{}
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		if pkg.Types != nil {
			loaded[pkg.PkgPath] = pkg.Types
		}
	})

	c := &compiler.Compiler{
		PackageName:   *packageName,
		Importer:      compiler.PackageImporter(loaded),
		HandlerPrefix: *handlerPrefix,
		OutputFile:    *outputFile,
	}

	cDir := filepath.Clean(*componentsDir)
	if info, err := os.Stat(cDir); err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("error: %s: %v\n", cDir, err)
		}
		log.Printf("no components directory at %s, skipping\n", cDir)
	} else if !info.IsDir() {
		log.Fatalf("error: %s is not a directory\n", cDir)
	} else if err := filepath.WalkDir(cDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".html" {
			return nil
		}
		rel, err := filepath.Rel(cDir, path)
		if err != nil {
			return err
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		c.NewComponent(filepath.ToSlash(rel), body)
		return nil
	}); err != nil {
		log.Fatalf("error: %s: walk failed: %v\n", cDir, err)
	}

	pageCount := 0
	pDir := filepath.Clean(*pagesDir)
	if info, err := os.Stat(pDir); err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("pages directory not found: %s\n", pDir)
		}
		log.Fatalf("error: %s: %v\n", pDir, err)
	} else if !info.IsDir() {
		log.Fatalf("error: %s is not a directory\n", pDir)
	} else if err := filepath.WalkDir(pDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".html" {
			return nil
		}
		rel, err := filepath.Rel(pDir, path)
		if err != nil {
			return err
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		c.NewPage(filepath.ToSlash(rel), body)
		pageCount++
		return nil
	}); err != nil {
		log.Fatalf("error: %s: walk failed: %v\n", pDir, err)
	}
	if pageCount == 0 {
		log.Printf("warning: no pages found in %s\n", pDir)
	}

	code, err := c.Compile()
	if err != nil {
		var errs scanner.ErrorList
		if errors.As(err, &errs) && len(errs) > 0 {
			lines := strings.Split(string(code), "\n")
			start := max(errs[0].Pos.Line-6, 0)
			end := min(errs[0].Pos.Line+5, len(lines))
			for i := start; i < end; i++ {
				log.Printf("%d: %s\n", i+1, lines[i])
			}
		}
		log.Fatalln(err)
	}

	out := filepath.Clean(*outputFile)
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		log.Fatalln(err)
	}
	if err := os.WriteFile(out, code, 0644); err != nil {
		log.Fatalln(err)
	}
	log.Printf("%s generated successfully\n", out)
}
