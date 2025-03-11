package main

import (
	"html/template"
	"log"
	"net/http"
)

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
