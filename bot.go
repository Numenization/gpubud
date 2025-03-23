package main

import (
	"fmt"
	"log"
	"strconv"
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
	{
		Name:        "remove-rule",
		Description: "Remove a notification rule",
	},
	{
		Name:        "list",
		Description: "Lists all the currently in stock GPUs",
	},
}

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot){
	"help": func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot) {
		// TODO: Write help function. This function will give a more detailed explanation of the commands the bot offers
	},

	"subscribe": func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot) {
		channel, err := s.Channel(i.ChannelID)
		if err != nil {
			log.Printf("Could not respond to interaction %s: %s\n", i.ApplicationCommandData().Name, err.Error())
		}

		content := ""
		stop := false

		if c, ok := b.config.NotifierChannels[i.ChannelID]; ok {
			// We have the current channel in the configuration already
			err := c.Subscribe(b.config.Env)
			if err != nil {
				content = fmt.Sprintf("Could not subscribe: %s", err.Error())
				stop = true
			}
			if !stop {
				content = "Subscribed for notifications"
			}
		} else {
			// The current channel is new and not in our configuration map
			newConfig, err := CreateChannelConfig(b.config.Env, channel)
			if err != nil {
				content = fmt.Sprintf("Could not subscribe: %s", err.Error())
				stop = true
			}

			subscribeErr := newConfig.Subscribe(b.config.Env)
			if subscribeErr != nil && !stop {
				content = fmt.Sprintf("Could not subscribe: %s", err.Error())
			} else if subscribeErr == nil && !stop {
				b.config.NotifierChannels[channel.ID] = newConfig
				content = "Subscribed for notifications"
			}
		}

		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}

		Respond(s, i, response)
	},

	"unsubscribe": func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot) {
		content := ""
		stop := false

		if c, ok := b.config.NotifierChannels[i.ChannelID]; ok {
			err := c.Unsubscribe(b.config.Env)
			if err != nil {
				content = fmt.Sprintf("Could not unsubscribe: %s", err.Error())
				stop = true
			}
			if !stop {
				content = "Unsubscribed from notifications"
			}
		} else {
			// Current channel is not in config list
			content = "This channel has not been configured to recieve notifications"
		}

		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}

		Respond(s, i, response)
	},

	"rules": func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot) {
		content := ""

		if c, ok := b.config.NotifierChannels[i.ChannelID]; ok {
			content = "Rules for current channel: "
			var sb strings.Builder
			for _, r := range c.Rules {
				sb.WriteString(fmt.Sprintf("`%s` ", r.Query))
			}

			content = content + sb.String()
		} else {
			content = "Could not get rule data for channel"
		}

		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}

		Respond(s, i, response)
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

		Respond(s, i, response)
	},

	"remove-rule": func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot) {
		var options []discordgo.SelectMenuOption

		if c, ok := b.config.NotifierChannels[i.ChannelID]; ok {
			for _, r := range c.Rules {
				opt := discordgo.SelectMenuOption{
					Label: r.Query,
					Value: r.Query,
					Emoji: &discordgo.ComponentEmoji{
						Name: "🗑️",
					},
					Default: false,
				}

				options = append(options, opt)
			}
		}

		if len(options) <= 0 {
			Respond(s, i, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "There are no rules for the current channel",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Rule Removal",
				Flags:   discordgo.MessageFlagsEphemeral,
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.SelectMenu{
								CustomID:    "remove_rule",
								Placeholder: "Select a rule to remove 👇",
								Options:     options,
							},
						},
					},
				},
			},
		}

		Respond(s, i, response)
	},

	"list": func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot) {
		gpus, err := GetAllGPUs(b.config.Env)
		if err != nil {
			Respond(s, i, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Error in retrieving GPU data: %s", err.Error()),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		if len(gpus) <= 0 {
			Respond(s, i, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "There are no stocked GPUs in database",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		page, err := GetListPage(0, b)
		if err != nil {
			Respond(s, i, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Error generating page: %s", err.Error()),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		Respond(s, i, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: page,
		})
	},
}

var componentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot){
	"remove_rule": func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot) {
		data := i.MessageComponentData()

		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("⚠️ Are you sure you want to remove this rule? ⚠️\n`%s`", data.Values[0]),
				Flags:   discordgo.MessageFlagsEphemeral,
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label:    "Yes",
								Style:    discordgo.PrimaryButton,
								Disabled: false,
								CustomID: fmt.Sprintf("remove_rule_accept_%s", data.Values[0]),
							},
							discordgo.Button{
								Label:    "No",
								Style:    discordgo.DangerButton,
								Disabled: false,
								CustomID: fmt.Sprintf("remove_rule_decline_%s", data.Values[0]),
							},
						},
					},
				},
			},
		}

		Respond(s, i, response)
	},
}

var componentResponseHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot, d string){
	"remove_rule_accept": func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot, d string) {
		if c, ok := b.config.NotifierChannels[i.ChannelID]; ok {
			err := c.RemoveRule(d, b.config.Env)
			if err != nil {
				Respond(s, i, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseUpdateMessage,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("Error in deleting rule: %s", err.Error()),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})

				return
			}

			Respond(s, i, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Rule `%s` removed ✅", d),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
	},

	"remove_rule_decline": func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot, d string) {
		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content: "Rule not deleted",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}

		Respond(s, i, response)
	},

	"list_page": func(s *discordgo.Session, i *discordgo.InteractionCreate, b *DiscordBot, d string) {
		pageIdx, err := strconv.Atoi(d)
		if err != nil {
			Respond(s, i, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Error generating page: %s", err.Error()),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		page, err := GetListPage(pageIdx, b)
		if err != nil {
			Respond(s, i, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Error generating page: %s", err.Error()),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		Respond(s, i, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: page,
		})
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
					log.Printf("Could not process modal response: %s\n", err.Error())
				}

				return
			}

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("New rule created for `%s` ✅\nYou will now recieve notifications in this channel when a GPU matching this model is updated", model),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				log.Printf("Could not process modal response: %s\n", err.Error())
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
				log.Printf("Could not process modal response: %s\n", err.Error())
			}
		}
	},
}

// Helper function for the list command handler to get paginated results
func GetListPage(p int, b *DiscordBot) (*discordgo.InteractionResponseData, error) {
	gpus, err := GetAllGPUs(b.config.Env)
	if err != nil {
		return nil, fmt.Errorf("could not get paginated results: %s", err.Error())
	}

	// Create a new slice of only stocked gpus
	var stockedGpus []*GPU
	for _, gpu := range gpus {
		if gpu.Stock <= 0 {
			continue
		}
		stockedGpus = append(stockedGpus, gpu)
	}

	pageLen := 8
	maxPages := (len(stockedGpus) + (pageLen - 1)) / pageLen
	start := pageLen * p
	end := min(start+pageLen, len(stockedGpus)-1)

	var embeds []*discordgo.MessageEmbed
	for i := start; i < end; i++ {
		gpu := stockedGpus[i]
		imageURL := fmt.Sprintf("https://90a1c75758623581b3f8-5c119c3de181c9857fcb2784776b17ef.ssl.cf2.rackcdn.com/%v_%s_01_front_zoom.jpg", gpu.ID, gpu.SKU)
		embed := &discordgo.MessageEmbed{
			URL:         gpu.Link,
			Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: imageURL},
			Title:       fmt.Sprintf("%s %s %s %s", gpu.Manufacturer, gpu.Brand, gpu.Line, gpu.ProductModel),
			Description: fmt.Sprintf("$%v - %v in stock at Microcenter", gpu.Price, gpu.Stock),
		}
		embeds = append(embeds, embed)
	}

	nextPageDisabled := false
	prevPageDisabled := false
	if p+1 >= maxPages {
		nextPageDisabled = true
	}
	if p-1 < 0 {
		prevPageDisabled = true
	}

	nav := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Style: discordgo.PrimaryButton,
				Emoji: &discordgo.ComponentEmoji{
					Name: "👈",
				},
				Disabled: prevPageDisabled,
				CustomID: fmt.Sprintf("list_page_%v", p-1),
			},
			discordgo.Button{
				Label:    fmt.Sprintf("Page %v/%v", p+1, maxPages),
				Style:    discordgo.SecondaryButton,
				Disabled: true,
				CustomID: "list_page_index",
			},
			discordgo.Button{
				Style: discordgo.PrimaryButton,
				Emoji: &discordgo.ComponentEmoji{
					Name: "👉",
				},
				Disabled: nextPageDisabled,
				CustomID: fmt.Sprintf("list_page_%v", p+1),
			},
		},
	}

	response := &discordgo.InteractionResponseData{
		Content:    "Heres the list of currently in stock GPUs:",
		Flags:      discordgo.MessageFlagsEphemeral,
		Embeds:     embeds,
		Components: []discordgo.MessageComponent{nav},
	}

	return response, nil
}

// Sends a response to an Interaction with error handling. If an error occurs, it will try to send a response notifying the user of the error.
func Respond(s *discordgo.Session, i *discordgo.InteractionCreate, r *discordgo.InteractionResponse) {
	err := s.InteractionRespond(i.Interaction, r)
	if err != nil {
		log.Printf("Could not respond to interaction %s: %s\n", i.ID, err.Error())

		err2 := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error in sending response: %s", err.Error()),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})

		if err2 != nil {
			log.Panicf("Could not send response for error in Respond(): %s", err.Error())
		}
	}
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
		log.Printf("Creating application command '%s'\n", v.Name)
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
	log.Println("Cleaning up discord bot...")

	// Close the bot session and return
	log.Println("Closing discord session...")
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
			data := i.ApplicationCommandData()
			if handler, ok := commandHandlers[data.Name]; ok {
				handler(s, i, bot)
			}
		case discordgo.InteractionMessageComponent:
			data := i.MessageComponentData()
			if handler, ok := componentHandlers[data.CustomID]; ok {
				// Catch any interactions that don't have any custom data with them
				handler(s, i, bot)
			} else {
				// Check to see if this interaction's ID refers to a handler
				for k, handler := range componentResponseHandlers {
					if !strings.HasPrefix(data.CustomID, k) {
						continue
					}

					data, _ := strings.CutPrefix(data.CustomID, k+"_")
					handler(s, i, bot, data)
				}
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
