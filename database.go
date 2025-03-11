package main

import (
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GPU is a struct that holds the data for a GPU
type GPU struct {
	gorm.Model
	ID           string `json:"id" gorm:"primaryKey"`
	Brand        string `json:"brand"`
	Line         string `json:"line"`
	Link         string `json:"link"`
	Manufacturer string `json:"manufacturer"`
	ProductModel string `json:"model"`
	Name         string `json:"name"`
	PriceString  string `json:"price"`
	Stock        string `json:"stock"`
	Price        float64
}

// ScrapeData is a struct that holds the data scraped from the Microcenter website
type ScrapeData struct {
	GPUs      []*GPU
	Source    string
	Timestamp string
}

// Inserts or updates a GPU in the database
func UpsertGPU(env *Env, gpu *GPU) {
	env.DB.Clauses(clause.OnConflict{UpdateAll: true}).Create(gpu)
}

// Finds a GPU in the database by its ID
func FindGPU(env *Env, id string) (*GPU, error) {
	var gpu GPU
	result := env.DB.First(&gpu, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &gpu, nil
}

// Gets all GPUs in the database
func GetAllGPUs(env *Env) ([]*GPU, error) {
	var gpus []*GPU
	result := env.DB.Find(&gpus)
	if result.Error != nil {
		return nil, result.Error
	}
	return gpus, nil
}

// Gets the GPUs in the database and compares it to another list of GPUs. Any GPU found in the database
// but not in the given list will be assumed to be out of stock and will be updated in the database.
func UpdateMissingGPUs(env *Env, gpu []*GPU) {
	var dbGPUs []*GPU
	env.DB.Find(&dbGPUs)

	for _, dbGPU := range dbGPUs {
		found := false
		for _, gpu := range gpu {
			if dbGPU.ID == gpu.ID {
				found = true
				break
			}
		}
		if !found {
			log.Println("GPU out of stock: ", dbGPU.ID)
			dbGPU.Stock = "0"
			env.DB.Save(&dbGPU)
		}
	}
}
