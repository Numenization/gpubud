package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
)

func HandleRoot(env *Env) func(w http.ResponseWriter, r *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		scrape_data, err := Scrape()
		if err != nil {
			errorString := fmt.Sprintf("An error occured handling web request: %s", err.Error())
			log.Println(errorString)
			io.WriteString(w, errorString)
		}

		tmpl := template.Must(template.ParseFiles("./templates/index.html"))
		tmpl_data := map[string]ScrapeData{
			"ScrapeData": scrape_data,
		}
		tmpl.Execute(w, tmpl_data)
	}

	return handler
}
