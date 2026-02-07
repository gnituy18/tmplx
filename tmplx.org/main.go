package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	env := os.Getenv("ENV")

	for _, route := range Routes() {
		http.HandleFunc(route.Pattern, route.Handler)
	}
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./assets"))))

	if env == "prod" {
		http80 := http.NewServeMux()
		http80.Handle("/.well-known/acme-challenge/", http.StripPrefix("/.well-known/acme-challenge/", http.FileServer(http.Dir("/var/www/letsencrypt"))))
		http80.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
		})
		go http.ListenAndServe(":80", http80)

		cert := os.Getenv("CERT")
		pk := os.Getenv("PK")
		log.Fatal(http.ListenAndServeTLS(":443", cert, pk, nil))
	} else {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
}
