package main

import (
	"log"
	"os"

	bot "github.com/chipotlegroove/crafty-discord-bot/Bot"
	crafty "github.com/chipotlegroove/crafty-discord-bot/Crafty"
	requests "github.com/chipotlegroove/crafty-discord-bot/Requests"
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
	requests.Init()

	bot.Run()
}
