package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type DiscordBot struct {
	session *discordgo.Session
	config  *DiscordBotConfig
}

type DiscordBotConfig struct {
	Token            string
	NotifierChannels []*discordgo.Channel
	Env              *Env
}

func (bot *DiscordBot) Open() error {
	return bot.session.Open()
}

func (bot *DiscordBot) SendMessage(msg string) {
	for _, channel := range bot.config.NotifierChannels {
		bot.session.ChannelMessageSend(channel.ID, msg)

	}
}

func NewDiscordBot(config *DiscordBotConfig) (*DiscordBot, error) {
	bot := &DiscordBot{
		config: config,
	}

	s, err := discordgo.New(fmt.Sprintf("Bot %s", config.Token))
	if err != nil {
		return nil, fmt.Errorf("error in initializing bot: %s", err.Error())
	}

	bot.session = s

	return bot, nil
}
