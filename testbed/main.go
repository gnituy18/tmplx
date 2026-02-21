package main

import (
	"log"
	"net/http"
)

func main() {
	for _, route := range Routes() {
		http.HandleFunc(route.Pattern, route.Handler)
	}

	log.Fatal(http.ListenAndServe(":7327", nil))
}
