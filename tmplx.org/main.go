package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gnituy18/tmplx/compiler"
)

// playgroundCompile compiles a tmplx snippet posted by the playground and
// returns the generated Go (or diagnostics) as JSON. It only type-checks and
// generates code — it never runs the user's Go — so it is safe to expose.
func playgroundCompile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	src, err := io.ReadAll(io.LimitReader(r.Body, 64<<10))
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}
	c := &compiler.Compiler{}
	c.NewPage("index.html", src)
	code, err := c.Compile()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"code": string(code), "diagnostics": compiler.Messages(err)})
}

func main() {
	env := os.Getenv("ENV")

	for _, route := range Routes() {
		http.HandleFunc(route.Pattern, route.Handler)
	}
	http.HandleFunc("/playground/compile", playgroundCompile)
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./assets"))))

	if env == "prod" {
		http80 := http.NewServeMux()
		http80.Handle("/.well-known/", http.FileServer(http.Dir("/var/www/letsencrypt")))
		http80.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
		})
		go http.ListenAndServe(":80", http80)

		cert := os.Getenv("CERT")
		pk := os.Getenv("PK")
		log.Fatal(http.ListenAndServeTLS(":443", cert, pk, nil))
	} else {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		log.Fatal(http.ListenAndServe(":"+port, nil))
	}
}
