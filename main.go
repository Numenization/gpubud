package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("Starting server")

	http.HandleFunc("/", HandleRoot)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
