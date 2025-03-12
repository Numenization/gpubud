package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Env struct {
	DB             *gorm.DB
	LastScrapeTime time.Time
}

func InitEnvironment() (*Env, error) {
	DB, err := gorm.Open(sqlite.Open("gpubud.db"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error in environment initialization: %s", err.Error())
	}

	DB.AutoMigrate(&GPU{}, &Price{})

	env := &Env{
		DB:             DB,
		LastScrapeTime: time.Now(),
	}

	Scrape(env)

	return env, nil
}

func main() {
	log.Println("Starting server")

	env, err := InitEnvironment()
	if err != nil {
		log.Fatal(err.Error())
	}

	http.HandleFunc("/", HandleRoot(env))

	log.Fatal(http.ListenAndServe(":8000", nil))
}
