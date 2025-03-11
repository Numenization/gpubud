package main

import (
	"fmt"
	"log"
	"net/http"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Env struct {
	DB *gorm.DB
}

func InitEnvironment() (*Env, error) {
	DB, err := gorm.Open(sqlite.Open("gpubud.db"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error in environment initialization: %s", err.Error())
	}

	DB.AutoMigrate(&GPU{})

	env := &Env{
		DB: DB,
	}

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
