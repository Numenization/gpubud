package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Env struct {
	DB              *gorm.DB
	DiscordBot      *DiscordBot
	LastScrapeTime  time.Time
	RunUpdateLoop   bool
	UpdateManager   *UpdateManager
	MicrocenterUrl  string
	DiscordBotToken string
}

func GetEnvironmentVariable(v string) (string, error) {
	result, set := os.LookupEnv(v)
	if !set {
		return "", fmt.Errorf("cannot find environment variable: %s", v)
	}

	return result, nil
}

func SendDiscordMessage(s *discordgo.Session, message string) {
	s.UserGuilds(200, "", "", false)
}

func InitEnvironment() (*Env, error) {
	// Get environment variables from OS
	var err error
	microcenterUrl, err := GetEnvironmentVariable("MICROCENTER_URL")
	if err != nil {
		return nil, fmt.Errorf("error in initialization: %s", err.Error())
	}

	discordBotToken, err := GetEnvironmentVariable("DISCORD_BOT_TOKEN")
	if err != nil {
		return nil, fmt.Errorf("error in initialization: %s", err.Error())
	}

	// Open and run migrations for database
	DB, err := gorm.Open(sqlite.Open("gpubud.db"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error in initialization: %s", err.Error())
	}

	DB.AutoMigrate(&GPU{}, &Price{})

	// Setup Env struct
	env := &Env{
		DB:              DB,
		LastScrapeTime:  time.Now(),
		RunUpdateLoop:   true,
		MicrocenterUrl:  microcenterUrl,
		DiscordBotToken: discordBotToken,
	}

	env.UpdateManager = NewUpdateManager(env, 5*time.Minute)

	env.UpdateManager.Add(Scrape)
	env.UpdateManager.Add(ReportGPUData)

	bot, err := NewDiscordBot(&DiscordBotConfig{
		Token: discordBotToken,
	})
	if err != nil {
		return nil, fmt.Errorf("error in creating discord bot: %s", err.Error())
	}

	env.DiscordBot = bot

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
