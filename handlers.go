package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

func GetGPUData(env *Env) ([]*GPU, error) {
	var gpus []*GPU

	if time.Since(env.LastScrapeTime).Minutes() < 5 {
		dbGpus, err := GetAllGPUs(env)
		if err != nil {
			return nil, fmt.Errorf("error in retrieving GPU data from database: %s", err.Error())
		}
		gpus = dbGpus
	} else {
		scrape_data, err := Scrape(env)
		if err != nil {
			return nil, fmt.Errorf("error in scraping GPU data: %s", err.Error())
		}

		gpus = scrape_data.GPUs
	}

	return gpus, nil
}

func HandleRoot(env *Env) func(w http.ResponseWriter, r *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		gpus, err := GetGPUData(env)
		if err != nil {
			log.Println("error in route root: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl := template.Must(template.ParseFiles("./templates/index.html"))
		tmpl_data := map[string][]*GPU{
			"GPUs": gpus,
		}
		tmpl.Execute(w, tmpl_data)
	}

	return handler
}
