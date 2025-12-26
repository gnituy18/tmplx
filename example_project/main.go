package main

import (
	"log"
	"net/http"

	"example_project/tmplx"
)

func main() {
	for _, r := range tmplx.Routes() {
		http.HandleFunc(r.Pattern, r.Handler)
	}

	log.Fatal(http.ListenAndServe(":8081", nil))
}
