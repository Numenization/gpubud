package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Starting server")

	h1 := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello World\n")
	}

	http.HandleFunc("/", h1)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
