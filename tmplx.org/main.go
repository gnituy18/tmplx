package main

import (
	"log"
	"net/http"
	"os"

	"tmplx.org/tmplx"
)

func main() {
	env := os.Getenv("ENV")
	cert := os.Getenv("CERT")
	pk := os.Getenv("PK")

	for _, th := range tmplx.Routes() {
		http.HandleFunc(th.Pattern, th.Handler)
	}

	fs := http.FileServer(http.Dir("./assets"))
	http.Handle("/", http.StripPrefix("/", fs))
	if env == "prod" {
		go http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
		}))
		log.Fatal(http.ListenAndServeTLS(":443", cert, pk, nil))
	} else {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
}
