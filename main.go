package main

import (
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Callback struct {
	ID       string
	Function func(*Env) error
}

type Env struct {
	DB              *gorm.DB
	LastScrapeTime  time.Time
	RunUpdateLoop   bool
	UpdateSleepTime time.Duration
	UpdateCallback  []Callback
}

func (env *Env) AddUpdateCallback(id string, callback func(*Env) error) {
	env.UpdateCallback = append(env.UpdateCallback, Callback{ID: id, Function: callback})
}

func (env *Env) RemoveUpdateCallback(id string) {
	for i, cb := range env.UpdateCallback {
		if cb.ID == id {
			env.UpdateCallback = slices.Delete(env.UpdateCallback, i, i+1)
			return
		}
	}
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
		UpdateCallback:  []Callback{},
	}

	env.AddUpdateCallback("scrape", Scrape)

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
			callbackErr := callback.Function(env)
			if callbackErr != nil {
				log.Fatal(fmt.Errorf("error in update callback id: %s: %s", callback.ID, callbackErr.Error()))
			}
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
