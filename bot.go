package main

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

type DiscordBot struct {
	session *discordgo.Session
	config  *DiscordBotConfig
}

type DiscordBotConfig struct {
	// The bot's discord API access token
	Token string
	// A map that maps the string channel ID of a channel to its respective configuration
	NotifierChannels map[string]*ChannelConfig
	// The environment structure for the GPU Bud core
	Env *Env
}

type optionMap = map[string]*discordgo.ApplicationCommandInteractionDataOption

var commands = []*discordgo.ApplicationCommand{
	/*
		{
			Name:        "echo",
			Description: "Echoes the user's message",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "message",
					Description: "The message to echo back",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
	*/
	{
		Name:        "subscribe",
		Description: "Subscribes current channel to GPU notifier",
	},
	{
		Name:        "unsubscribe",
		Description: "Unsubscribes current channel from GPU notifier",
	},
}

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot){
	"subscribe": func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot) {
		channel, err := s.Channel(i.ChannelID)
		if err != nil {
			log.Panicf("Could not respond to interaction %s: %s", i.ApplicationCommandData().Name, err.Error())
		}

		response := ""
		stop := false

		if c, ok := b.config.NotifierChannels[channel.ID]; ok {
			// We have the current channel in the configuration already
			err := c.Subscribe(b.config.Env)
			if err != nil {
				response = fmt.Sprintf("Could not subscribe: %s", err.Error())
				stop = true
			}
			if !stop {
				response = "Subscribed for notifications"
			}
		} else {
			// The current channel is new and not in our configuration map
			newConfig, err := CreateChannelConfig(b.config.Env, channel)
			if err != nil {
				response = fmt.Sprintf("Could not subscribe: %s", err.Error())
				stop = true
			}

			subscribeErr := newConfig.Subscribe(b.config.Env)
			if subscribeErr != nil && !stop {
				response = fmt.Sprintf("Could not subscribe: %s", err.Error())
			} else if subscribeErr == nil && !stop {
				b.config.NotifierChannels[channel.ID] = newConfig
				response = "Subscribed for notifications"
			}
		}

		resErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})

		if resErr != nil {
			log.Panicf("could not respond to interaction %s: %s", i.ApplicationCommandData().Name, err.Error())
		}
	},

	"unsubscribe": func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot) {
		channel, err := s.Channel(i.ChannelID)
		if err != nil {
			log.Panicf("Could not respond to interaction %s: %s", i.ApplicationCommandData().Name, err.Error())
		}

		response := ""
		stop := false

		if c, ok := b.config.NotifierChannels[channel.ID]; ok {
			err := c.Unsubscribe(b.config.Env)
			if err != nil {
				response = fmt.Sprintf("Could not unsubscribe: %s", err.Error())
				stop = true
			}
			if !stop {
				response = "Unsubscribed from notifications"
			}
		} else {
			// Current channel is not in config list
			response = "This channel has not been configured to recieve notifications"
		}

		resErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})

		if resErr != nil {
			log.Panicf("could not respond to interaction %s: %s", i.ApplicationCommandData().Name, err.Error())
		}
	},
}

// Maps the options for a discord command interaction
func parseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) optionMap {
	om := make(optionMap)
	for _, opt := range options {
		om[opt.Name] = opt
	}
	return om
}

// Creates the discord API session and registers the bot's commands
func (bot *DiscordBot) Open() error {
	err := bot.session.Open()
	if err != nil {
		return fmt.Errorf("could not open discord bot: %s", err.Error())
	}

	log.Printf("Adding commands to Discord bot...")

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		log.Printf("Trying to add command '%s'\n", v.Name)
		cmd, err := bot.session.ApplicationCommandCreate(bot.session.State.User.ID, "", v)
		if err != nil {
			return fmt.Errorf("cannot create command '%s' command: %s", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	log.Println("Discord bot ready")

	return nil
}

// Closes the discord API session
func (bot *DiscordBot) Close() error {
	log.Println("Closing discord bot...")
	return bot.session.Close()
}

// Send a string message to all the channels in the channel:config map
func (bot *DiscordBot) NotifyChannels(msg string) {
	for k := range bot.config.NotifierChannels {
		bot.session.ChannelMessageSend(k, msg)
	}
}

// Creates a new discord bot with a given configuration
func NewDiscordBot(config *DiscordBotConfig) (*DiscordBot, error) {
	log.Println("Starting Discord bot...")

	bot := &DiscordBot{
		config: config,
	}

	s, err := discordgo.New(fmt.Sprintf("Bot %s", config.Token))
	if err != nil {
		return nil, fmt.Errorf("error in initializing bot: %s", err.Error())
	}

	// Set up event handlers
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Successfully logged in as: %v#%v\n", s.State.User.Username, s.State.User.Discriminator)
	})
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if handler, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			handler(s, i, bot)
		}
	})

	bot.session = s

	return bot, nil
}
