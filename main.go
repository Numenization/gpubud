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
	RunUpdateLoop  bool
	UpdateManager  *UpdateManager
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
		RunUpdateLoop:  true,
	}

	env.UpdateManager = NewUpdateManager(env, 5*time.Minute)

	env.UpdateManager.Subscribe(Scrape)
	env.UpdateManager.Subscribe(ReportGPUData)

	return env, nil
}

func main() {
	log.Println("Initializing environment")
	env, err := InitEnvironment()
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Starting server")
	http.HandleFunc("/", HandleRoot(env))

	env.UpdateManager.Start()
	env.UpdateManager.UpdateNow()

	log.Fatal(http.ListenAndServe(":8000", nil))
}
