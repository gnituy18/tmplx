package main

import (
	"io"
	"log"
	"os"

	"github.com/gnituy18/tmplx/tmpls"
)

func main() {
	log.SetFlags(0)
	if err := tmpls.Run(os.Stdin, os.Stdout); err != nil && err != io.EOF {
		log.Fatalf("tmpls: %v", err)
	}
}
