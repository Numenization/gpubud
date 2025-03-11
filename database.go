package main

import "gorm.io/gorm"

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

type ScrapeData struct {
	GPUs      []*GPU
	Source    string
	Timestamp string
}
