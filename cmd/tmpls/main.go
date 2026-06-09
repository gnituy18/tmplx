// Command tmpls is the tmplx language server (like gopls). It speaks LSP over
// stdio; the implementation lives in the tmpls package.
package main

import (
	"io"
	"log"
	"os"

	"github.com/gnituy18/tmplx/tmpls"
)

func main() {
	log.SetFlags(0) // logs go to stderr; stdout is the LSP channel
	if err := tmpls.Run(os.Stdin, os.Stdout); err != nil && err != io.EOF {
		log.Fatalf("tmpls: %v", err)
	}
}
