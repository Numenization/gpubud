package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
)

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	_, err := scrape()
	if err != nil {
		errorString := fmt.Sprintf("An error occured handling web request: %s", err.Error())
		log.Println(errorString)
		io.WriteString(w, errorString)
	}

	tmpl := template.Must(template.ParseFiles("./templates/index.html"))
	tmpl.Execute(w, nil)
	// io.WriteString(w, results)
}
