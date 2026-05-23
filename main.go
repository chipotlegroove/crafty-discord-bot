package main

import (
	"fmt"
	"log"
	"os"

	bot "github.com/chipotlegroove/crafty-discord-bot/Bot"
	crafty "github.com/chipotlegroove/crafty-discord-bot/Crafty"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token := os.Getenv("BOT_TOKEN")
	craftyAPIToken := os.Getenv("CRAFTY_API_TOKEN")
	craftyBaseURL := os.Getenv("CRAFTY_BASE_URL")

	bot.Init(token)
	crafty.Init(craftyAPIToken, craftyBaseURL)

	servers, err := crafty.GetServersWithStatus()
	if err != nil {
		panic(err)
	}

	for _, server := range servers {
		fmt.Printf("%s \t %s\n", server.ServerName, server.Status.Label())
	}

	bot.Run()
}
