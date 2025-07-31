package main

import (
	"log"
	"net/http"

	"example_project/tmplx"
)

func main() {
	for _, th := range tmplx.Handlers() {
		http.HandleFunc("GET "+th.Url, th.HandlerFunc)
	}

	log.Fatal(http.ListenAndServe(":8080", nil))
}
