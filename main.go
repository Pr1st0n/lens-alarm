package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pr1st0n/lens-alarm/bot"
	"github.com/pr1st0n/lens-alarm/util"
	"log"
	"os"
)

func main() {
	token := os.Getenv("LENS_API_TOKEN")
	tgbot, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", tgbot.Self.UserName)

	// Create data.json, if it does not exist
	_, err = os.OpenFile("data.json", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := tgbot.GetUpdatesChan(u)
	chatBot := bot.ChatBot{Bot: tgbot}

	// Notifications handler
	go util.ScheduleJob(chatBot.SendNotification)

	// User updates handler
	for update := range updates {
		chatBot.HandleUpdate(update)
	}
}
