package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	results, err := scrape()
	if err != nil {
		errorString := fmt.Sprintf("An error occured handling web request: %s", err.Error())
		log.Println(errorString)
		io.WriteString(w, errorString)
	}
	io.WriteString(w, results)
}
