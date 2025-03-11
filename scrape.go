package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

func ConvertPriceStrings(data *ScrapeData) error {
	for _, gpu := range data.GPUs {
		price, err := strconv.ParseFloat(gpu.PriceString, 64)
		if err != nil {
			return fmt.Errorf("error in converting GPU price strings: %s", err.Error())
		}
		gpu.Price = price
	}

	return nil
}

func Scrape(env *Env) (ScrapeData, error) {
	var data ScrapeData

	// check to make sure we have the URL in our env
	url, set := os.LookupEnv("MICROCENTER_URL")
	if !set {
		return data, errors.New("missing MICROCENTER_URL env")
	}

	// execute the python scraper and get the data back
	cmd := exec.Command("python3", "./scrapers/scrape_microcenter.py", "-s", url)
	out, command_err := cmd.CombinedOutput()
	if command_err != nil {
		return data, fmt.Errorf("error in scraping microcenter: %s", command_err.Error())
	}

	// unpack the data from json format into a ScrapeData struct
	convert_err := json.Unmarshal(out, &data)
	if convert_err != nil {
		return data, fmt.Errorf("error in scraping microcenter: %s", convert_err.Error())
	}
	convert_err = ConvertPriceStrings(&data)
	if convert_err != nil {
		return data, fmt.Errorf("error in scraping microcenter: %s", convert_err.Error())
	}

	for _, gpu := range data.GPUs {
		UpsertGPU(env, gpu)
	}
	UpdateMissingGPUs(env, data.GPUs)

	return data, nil
}
