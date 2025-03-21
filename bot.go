package main

import (
	"fmt"
	"log"
	"slices"

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

type optionMap = map[string]*discordgo.ApplicationCommandInteractionDataOption

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "hello-world",
		Description: "Says 'Hello World!'",
	},
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
	"hello-world": func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot) {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Hello World!",
			},
		})

		if err != nil {
			log.Panicf("Could not respond to interaction %s: %s", i.ApplicationCommandData().Name, err.Error())
		}
	},

	"echo": func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot) {
		opts := parseOptions(i.ApplicationCommandData().Options)

		author := "Missing Author"

		if i.User != nil {
			// Message sent in DM
			author = i.User.GlobalName
		} else if i.Member != nil {
			// Message sent in Guild
			author = i.Member.User.GlobalName
		}

		response := fmt.Sprintf("%s: %s", author, opts["message"].StringValue())

		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})

		if err != nil {
			log.Panicf("could not respond to interaction %s: %s", i.ApplicationCommandData().Name, err.Error())
		}
	},

	"subscribe": func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot) {
		channel, err := s.Channel(i.ChannelID)
		if err != nil {
			log.Panicf("Could not respond to interaction %s: %s", i.ApplicationCommandData().Name, err.Error())
		}

		// Check to see if channel is already subscribed
		for _, v := range b.config.NotifierChannels {
			if v.ID == channel.ID {
				// This channel is already subscribed to the bot
				response := "Channel is already subscribed to bot."

				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: response,
					},
				})

				if err != nil {
					log.Panicf("could not respond to interaction %s: %s", i.ApplicationCommandData().Name, err.Error())
				}

				return
			}
		}

		// Channel isn't subscribed. Add them to the NotifierChannels list and respond
		b.config.NotifierChannels = append(b.config.NotifierChannels, channel)

		response := "Subscribed channel to bot notifications."

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

		// Check to see if channel is subscribed
		for idx, v := range b.config.NotifierChannels {
			if v.ID == channel.ID {
				// This channel is already subscribed to the bot. Remove it and respond
				b.config.NotifierChannels = slices.Delete(b.config.NotifierChannels, idx, idx+1)

				response := "Unsubscribed from bot notifications."

				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: response,
					},
				})

				if err != nil {
					log.Panicf("could not respond to interaction %s: %s", i.ApplicationCommandData().Name, err.Error())
				}

				return
			}
		}

		// If we're here, then the channel isn't in the notifier list
		response := "Channel is not subscribed for notifications."

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

func parseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (om optionMap) {
	om = make(optionMap)
	for _, opt := range options {
		om[opt.Name] = opt
	}
	return om
}

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

func (bot *DiscordBot) Close() error {
	log.Println("Closing discord bot...")
	return bot.session.Close()
}

func (bot *DiscordBot) NotifyChannels(msg string) {
	for _, channel := range bot.config.NotifierChannels {
		bot.session.ChannelMessageSend(channel.ID, msg)
	}
}

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
