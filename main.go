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
	DB              *gorm.DB
	LastScrapeTime  time.Time
	RunUpdateLoop   bool
	UpdateSleepTime time.Duration
	UpdateCallback  []func(*Env) error
}

func InitEnvironment() (*Env, error) {
	DB, err := gorm.Open(sqlite.Open("gpubud.db"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error in environment initialization: %s", err.Error())
	}

	DB.AutoMigrate(&GPU{}, &Price{})

	env := &Env{
		DB:              DB,
		LastScrapeTime:  time.Now(),
		RunUpdateLoop:   true,
		UpdateSleepTime: 5 * time.Minute,
		UpdateCallback: []func(*Env) error{
			Scrape,
		},
	}

	return env, nil
}

// Every 5 minutes, run the scraper to get the latest GPU data
func UpdateLoop(env *Env) {
	log.Println("Starting update loop")

	for env.RunUpdateLoop {
		// update GPU data
		log.Println("Updating GPU data")

		// run update callbacks
		for _, callback := range env.UpdateCallback {
			callback(env)
		}

		gpus, err := GetAllGPUs(env)
		if err != nil {
			log.Fatal("error in updating GPU data: ", err.Error())
		}
		log.Printf("New GPU data: %d GPUs in database\n", len(gpus))

		// sleep 5 minutes
		time.Sleep(env.UpdateSleepTime)
	}

	log.Println("Update loop stopped")
}

func main() {
	log.Println("Initializing environment")
	env, err := InitEnvironment()
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Starting server")
	http.HandleFunc("/", HandleRoot(env))
	go UpdateLoop(env)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
