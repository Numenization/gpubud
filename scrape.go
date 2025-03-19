package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

// Scrapes the Microcenter website for GPU data
func Scrape(env *Env) error {
	log.Println("Attempting to update GPU list from scraper")

	var data ScrapeData

	// check to make sure we have the URL in our env
	url, set := os.LookupEnv("MICROCENTER_URL")
	if !set {
		return errors.New("missing MICROCENTER_URL env")
	}

	// execute the python scraper and get the data back
	cmd := exec.Command("python3", "./scrapers/scrape_microcenter.py", "-s", url)
	out, command_err := cmd.CombinedOutput()
	if command_err != nil {
		return fmt.Errorf("error in scraping microcenter: %s", command_err.Error())
	}

	// unpack the data from json format into a ScrapeData struct
	convert_err := json.Unmarshal(out, &data)
	if convert_err != nil {
		return fmt.Errorf("error in scraping microcenter: %s", convert_err.Error())
	}

	for _, gpu := range data.GPUs {
		UpsertGPU(env, gpu)
	}
	UpdateMissingGPUs(env, data.GPUs)

	env.LastScrapeTime = time.Now()

	return nil
}

func ReportGPUData(env *Env) error {
	gpus, err := GetAllGPUs(env)
	if err != nil {
		return fmt.Errorf("error in updating GPU data: %s", err.Error())
	}
	log.Printf("New GPU data: %d GPUs in database\n", len(gpus))

	return nil
}
