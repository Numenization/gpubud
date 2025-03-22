package main

import (
	"fmt"
	"log"
	"strings"

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

//type optionMap = map[string]*discordgo.ApplicationCommandInteractionDataOption

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "subscribe",
		Description: "Subscribes current channel to GPU notifier",
	},
	{
		Name:        "unsubscribe",
		Description: "Unsubscribes current channel from GPU notifier",
	},
	{
		Name:        "rules",
		Description: "Returns a list of all the notification rules for the current channel",
	},
	{
		Name:        "add-rule",
		Description: "Create a new rule for GPU Bud to send notifications",
	},
}

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot){
	"help": func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot) {
		// TODO: Write help function. This function will give a more detailed explanation of the commands the bot offers
	},

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
			log.Panicf("could not respond to interaction %s: %s", i.ApplicationCommandData().Name, resErr.Error())
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
			log.Panicf("could not respond to interaction %s: %s", i.ApplicationCommandData().Name, resErr.Error())
		}
	},

	"rules": func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot) {
		channel, err := s.Channel(i.ChannelID)
		if err != nil {
			log.Panicf("Could not respond to interaction %s: %s", i.ApplicationCommandData().Name, err.Error())
		}

		response := ""

		if c, ok := b.config.NotifierChannels[channel.ID]; ok {
			response = "Rules for current channel: "
			var sb strings.Builder
			for _, r := range c.Rules {
				sb.WriteString(fmt.Sprintf("`%s` ", r.Query))
			}

			response = response + sb.String()
		} else {
			response = "Could not get rule data for channel"
		}

		resErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})

		if resErr != nil {
			log.Panicf("could not respond to interaction %s: %s", i.ApplicationCommandData().Name, resErr.Error())
		}
	},

	"add-rule": func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot) {
		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: "ar_submit",
				Title:    "New Rule",
				Flags:    discordgo.MessageFlagsEphemeral,
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								Label:       "GPU Model to get updates for",
								Style:       discordgo.TextInputShort,
								Placeholder: "Model...",
								MinLength:   1,
								MaxLength:   16,
								Required:    true,
							},
						},
					},
				},
			},
		}

		err := s.InteractionRespond(i.Interaction, response)
		if err != nil {
			log.Panicf("could not respond to interaction %s: %s", i.ApplicationCommandData().Name, err.Error())
		}
	},

	"remove-rule": func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot) {
		// TODO: Write remove-rule function. This will let users remove a rule from the channel config
	},
}

var modalHandlers = map[string]func(data *discordgo.ModalSubmitInteractionData, s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot){
	"ar_submit": func(data *discordgo.ModalSubmitInteractionData, s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot) {
		model := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

		if c, ok := b.config.NotifierChannels[i.ChannelID]; ok {
			err := c.AddRule(model, b.config.Env)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("Unable to create rule: %s", err.Error()),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})

				if err != nil {
					log.Panicf("could not process modal response: %s", err.Error())
				}

				return
			}

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("New rule submitted for `%s`\nYou will now recieve notifications in this channel when a GPU matching this model is updated", model),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				log.Panicf("could not process modal response: %s", err.Error())
			}
		} else {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Error in form submission: channel not in configuration map",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				log.Panicf("could not process modal response: %s", err.Error())
			}
		}
	},
}

/*
// Maps the options for a discord command interaction
func parseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) optionMap {
	om := make(optionMap)
	for _, opt := range options {
		om[opt.Name] = opt
	}
	return om
}
*/

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
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if handler, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				handler(s, i, bot)
			}
		case discordgo.InteractionModalSubmit:
			data := i.ModalSubmitData()

			if handler, ok := modalHandlers[data.CustomID]; ok {
				handler(&data, s, i, bot)
			}
		}
	})

	bot.session = s

	return bot, nil
}
