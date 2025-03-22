package main

import (
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GPU is a struct that holds the data for a GPU
type GPU struct {
	gorm.Model
	ID           int32   `json:"id" gorm:"primaryKey;autoIncrement:false"`
	Brand        string  `json:"brand"`
	Line         string  `json:"line"`
	Link         string  `json:"link"`
	Manufacturer string  `json:"manufacturer"`
	ProductModel string  `json:"model"`
	Name         string  `json:"name"`
	Stock        int32   `json:"stock"`
	Price        float64 `json:"price"`
}

// Price is a snapshot of the price of a GPU at a given time
type Price struct {
	gorm.Model
	Price float64
	Stock int32
	GPUID int32
	GPU   *GPU
	Time  time.Time
}

type ChannelConfig struct {
	gorm.Model
	ID         int32                `gorm:"primaryKey"`
	ChannelID  string               `gorm:"unique;not null"`
	Rules      []*ChannelConfigRule `gorm:"foreignKey:ChannelConfigRefer"`
	Subscribed bool                 `gorm:"default:false"`
}

type ChannelConfigRule struct {
	gorm.Model
	ID                 int32 `gorm:"primaryKey"`
	ChannelConfigRefer uint
	Query              string
}

// ScrapeData is a struct that holds the data scraped from the Microcenter website
type ScrapeData struct {
	GPUs      []*GPU
	Source    string
	Timestamp string
}

// Commit the config to the database, updating the database with any changes
func (c *ChannelConfig) commit(env *Env) error {
	result := env.DB.Save(c)
	if result.Error != nil {
		return fmt.Errorf("error in commiting config to db: %s", result.Error)
	}

	return nil
}

func (c *ChannelConfig) AddRule(q string, env *Env) error {
	// TODO: This is raw user input that gets put into database queries. Sanitization would be preferable
	cleansedInput := strings.TrimSpace(q)
	for _, v := range c.Rules {
		if v.Query == cleansedInput {
			return fmt.Errorf("rule already exists in config")
		}
	}
	c.Rules = append(c.Rules, &ChannelConfigRule{Query: cleansedInput})
	return c.commit(env)
}

func (c *ChannelConfig) RemoveRule(q string, env *Env) error {
	// TODO: Same as Addrule(). This user input should be sanitized more.
	cleansedInput := strings.TrimSpace(q)
	for i, v := range c.Rules {
		if v.Query == cleansedInput {
			c.Rules = slices.Delete(c.Rules, i, i+1)
			return c.commit(env)
		}
	}

	return fmt.Errorf("could not find rule in config")
}

func (c *ChannelConfig) Subscribe(env *Env) error {
	if c.Subscribed {
		return fmt.Errorf("already subscribed for notifications")
	}

	c.Subscribed = true
	return c.commit(env)
}

func (c *ChannelConfig) Unsubscribe(env *Env) error {
	if !c.Subscribed {
		return fmt.Errorf("not subscribed for notifications")
	}

	c.Subscribed = false
	return c.commit(env)
}

func CreateChannelConfig(env *Env, channel *discordgo.Channel) (*ChannelConfig, error) {
	newChannel := ChannelConfig{
		ChannelID: channel.ID,
	}

	result := env.DB.Create(&newChannel)
	if result.Error != nil {
		return nil, fmt.Errorf("could not create channel config: %s", result.Error)
	}

	return &newChannel, nil
}

func LoadChannelConfigs(env *Env) ([]*ChannelConfig, error) {
	var configs []*ChannelConfig
	result := env.DB.Preload(clause.Associations).Find(&configs)
	if result.Error != nil {
		return nil, fmt.Errorf("could not load channel configurations: %s", result.Error)
	}

	//env.ChannelConfigs = configs
	return configs, nil
}

// Inserts or updates a GPU in the database and create a new price for the current time
func InsertGPU(env *Env, gpu *GPU) {
	env.DB.Clauses(clause.OnConflict{UpdateAll: true}).Create(gpu)
	CreatePrice(env, gpu)
}

func CreatePrice(env *Env, gpu *GPU) {
	price := Price{
		Price: gpu.Price,
		Stock: gpu.Stock,
		GPUID: gpu.ID,
		GPU:   gpu,
		Time:  time.Now(),
	}
	env.DB.Create(&price)
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
			dbGPU.Stock = 0
			env.DB.Save(&dbGPU)
		}
	}
}
