package main

import (
	"log"
	"net/http"
)

func main() {
	for _, route := range Routes() {
		http.Handle(route.Pattern, route.Handler)
	}
	log.Fatal(http.ListenAndServe(":8081", nil))
}
