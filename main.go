package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func scrape() (string, error) {
	url, set := os.LookupEnv("MICROCENTER_URL")
	if !set {
		return url, errors.New("missing MICROCENTER_URL env")
	}

	cmd := exec.Command("python3", "./scrapers/scrape_microcenter.py", "-p", "-s", url)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return url, fmt.Errorf("error in scraping microcenter: %s", err.Error())
	}

	return string(out), nil
}

func main() {
	fmt.Println("Starting server")

	http.HandleFunc("/", HandleRoot)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
