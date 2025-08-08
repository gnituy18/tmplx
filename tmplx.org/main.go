package main

import (
	"log"
	"net/http"

	"tmplx.org/tmplx"
)

func main() {
	for _, th := range tmplx.Handlers() {
		http.HandleFunc("GET "+th.Url, th.HandlerFunc)
	}

	log.Fatal(http.ListenAndServe(":8080", nil))
}
