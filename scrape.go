package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"time"

	"gorm.io/gorm"
)

type GPUDifference struct {
	ID       int32
	GPUID    int32
	GPU      *GPU
	PriceNew float64
	PriceOld float64
	StockNew int32
	StockOld int32
	IsDiff   bool
}

func (diff *GPUDifference) String() string {
	return fmt.Sprintf("%v (%v/%v)", diff.GPUID, diff.PriceNew-diff.PriceOld, diff.StockNew-diff.StockOld)
}

// Scrapes the Microcenter website for GPU data
func Scrape(env *Env) error {
	log.Println("Attempting to update GPU list from scraper")

	var data ScrapeData

	// execute the python scraper and get the data back
	cmd := exec.Command("python3", "./scrapers/scrape_microcenter.py", "-s", env.MicrocenterUrl)
	out, command_err := cmd.CombinedOutput()
	if command_err != nil {
		return fmt.Errorf("error in scraping microcenter: %s", command_err.Error())
	}

	// unpack the data from json format into a ScrapeData struct
	convert_err := json.Unmarshal(out, &data)
	if convert_err != nil {
		return fmt.Errorf("error in scraping microcenter: %s", convert_err.Error())
	}

	var diffs []*GPUDifference
	for _, gpu := range data.GPUs {
		diff, err := Difference(gpu, env)
		if err != nil {
			return fmt.Errorf("error in scraping microcenter: %s", err.Error())
		}

		diffs = append(diffs, diff)
		InsertGPU(env, gpu)
	}

	go env.DiscordBot.NotifyChannels(diffs)
	UpdateMissingGPUs(env, data.GPUs)

	env.LastScrapeTime = time.Now()

	return nil
}

func Difference(gpu *GPU, env *Env) (*GPUDifference, error) {
	old, err := FindGPU(env, gpu.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			diff := &GPUDifference{
				GPUID:    gpu.ID,
				GPU:      gpu,
				PriceOld: 0,
				StockOld: 0,
				PriceNew: gpu.Price,
				StockNew: gpu.Stock,
				IsDiff:   true,
			}

			return diff, nil
		}
		return nil, fmt.Errorf("error in calculating GPU diff: %s", err.Error())
	}

	isDiff := false
	if gpu.Price != old.Price || gpu.Stock != old.Stock {
		isDiff = true
	}

	diff := &GPUDifference{
		GPUID:    gpu.ID,
		GPU:      gpu,
		PriceOld: old.Price,
		StockOld: old.Stock,
		PriceNew: gpu.Price,
		StockNew: gpu.Stock,
		IsDiff:   isDiff,
	}

	return diff, nil
}

// Log the number of GPUs we are currently tracking in the database
func ReportGPUData(env *Env) error {
	gpus, err := GetAllGPUs(env)
	if err != nil {
		return fmt.Errorf("error in updating GPU data: %s", err.Error())
	}
	log.Printf("New GPU data: %d GPUs in database\n", len(gpus))

	return nil
}
