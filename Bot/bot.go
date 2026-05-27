// Package bot provides the bot functionalities.
package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	crafty "github.com/chipotlegroove/crafty-discord-bot/Crafty"
)

var token string

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "list-servers",
			Description: "List all available servers",
		},
		{
			Name:        "start-server",
			Description: "Starts the selected server",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "server-name",
					Description: "Name of the server to start",
					Required:    true,
				},
			},
		},
		{
			Name:        "stop-server",
			Description: "Stops the selected server",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "server-name",
					Description: "Name of the server to stop",
					Required:    true,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"list-servers": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})
			servers, err := crafty.GetServersWithStatus()
			var embed *discordgo.MessageEmbed
			if err != nil {
				embed = &discordgo.MessageEmbed{
					Title:       "Server unreachable",
					Description: "Crafty Controller could not be reached 😔",
					Color:       0xf54927,
					Timestamp:   time.Now().Format(time.RFC3339),
				}
			} else {
				var fields []*discordgo.MessageEmbedField
				for _, server := range servers {
					field := &discordgo.MessageEmbedField{
						Name:  server.ServerName,
						Value: server.Status.Label(),
					}
					fields = append(fields, field)
				}
				embed = &discordgo.MessageEmbed{
					Title:     "Servers",
					Color:     0x00ff00,
					Fields:    fields,
					Timestamp: time.Now().Format(time.RFC3339),
				}
			}
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds: &[]*discordgo.MessageEmbed{embed},
			})
		},
		"start-server": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})

			options := i.ApplicationCommandData().Options

			serverName := options[0].Value.(string)

			err := crafty.ExecuteAction(serverName, "start_server")

			var embed *discordgo.MessageEmbed

			if err != nil {
				embed = &discordgo.MessageEmbed{
					Title: "Server could not start 🤬",
					Color: 0xf54927,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "Error details",
							Value: fmt.Sprintf("Error trying to start server %s: %s. Pendejo.", serverName, err),
						},
					},
				}
			} else {
				embed = &discordgo.MessageEmbed{
					Title:       "Server start signal sent 🥵",
					Color:       0x00ff00,
					Description: "Starting " + serverName + "...",
				}
			}

			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds: &[]*discordgo.MessageEmbed{embed},
			})
		},
		"stop-server": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})

			options := i.ApplicationCommandData().Options

			serverName := options[0].Value.(string)

			err := crafty.ExecuteAction(serverName, "stop_server")

			var embed *discordgo.MessageEmbed

			if err != nil {
				embed = &discordgo.MessageEmbed{
					Title: "Server could not stop 🤬",
					Color: 0xf54927,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "Error details",
							Value: fmt.Sprintf("Error trying to stop server %s: %s. Pendejo.", serverName, err),
						},
					},
				}
			} else {
				embed = &discordgo.MessageEmbed{
					Title: "Server stop signal sent 🥵",
					Color: 0x00ff00,
				}
			}

			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds: &[]*discordgo.MessageEmbed{embed},
			})
		},
	}
)

func Init(t string) {
	token = t
}

func Run() {
	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error initialising bot: %v", err)
	}

	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	discord.Identify.Intents = discordgo.IntentsGuildMessages

	err = discord.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
		return
	}

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := discord.ApplicationCommandCreate(discord.State.User.ID, "", v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}

		registeredCommands[i] = cmd
	}

	fmt.Println("Bot running. Press CTRL-C to exit")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, os.Interrupt)
	<-sc

	discord.Close()
}
