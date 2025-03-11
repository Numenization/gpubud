package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"
)

func HandleRoot(env *Env) func(w http.ResponseWriter, r *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var gpus []*GPU
		if time.Since(env.LastScrapeTime).Minutes() < 5 {
			dbGpus, err := GetAllGPUs(env)
			if err != nil {
				errorString := fmt.Sprintf("An error occured getting GPUs from the database: %s", err.Error())
				log.Println(errorString)
				io.WriteString(w, errorString)
				return
			}
			gpus = dbGpus
		} else {
			scrape_data, err := Scrape(env)
			if err != nil {
				errorString := fmt.Sprintf("An error occured handling web request: %s", err.Error())
				log.Println(errorString)
				io.WriteString(w, errorString)
				return
			}

			gpus = scrape_data.GPUs
		}

		tmpl := template.Must(template.ParseFiles("./templates/index.html"))
		tmpl_data := map[string][]*GPU{
			"GPUs": gpus,
		}
		tmpl.Execute(w, tmpl_data)
	}

	return handler
}
