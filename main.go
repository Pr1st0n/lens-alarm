package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
	"time"
)

type User struct {
	ChatId int64 `json:"chatId"`
	Packs  Packs `json:"packs"`
}

type Packs struct {
	Stock          int    `json:"stock"`
	Duration       int    `json:"duration"`
	LastOpenDate   string `json:"lastOpenDate"`
	ChangeDate     string `json:"changeDate"`
	LastChangeDate string `json:"lastChangeDate"`
}

func (user User) String() string {
	return fmt.Sprintf("Duration: %s, Days in use: %s, Stock: %s, Next change: %s, Last change: %s",
		user.Packs.Duration,
		user.Packs.LastOpenDate,
		user.Packs.Stock,
		user.Packs.ChangeDate,
		user.Packs.LastChangeDate)
}

func main() {
	token := os.Getenv("LENS_API_TOKEN")
	tgbot, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", tgbot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := tgbot.GetUpdatesChan(u)
	chatBot := ChatBot{bot: tgbot}

	go ScheduleJob(func() {
		for _, msg := range getMessages() {
			tgbot.Send(msg)
		}
	})

	for update := range updates {
		chatBot.HandleUpdate(update)
	}
}

func getMessages() []tgbotapi.MessageConfig {
	var messages []tgbotapi.MessageConfig
	now := time.Now()
	fileReader, err := os.Open("data.json")
	if err != nil {
		fmt.Printf("error: %v", err)
		return messages
	}

	defer func() {
		err := fileReader.Close()
		if err != nil {
			fmt.Printf("error: %v", err)
		}
	}()

	scanner := bufio.NewScanner(fileReader)

	for scanner.Scan() {
		bytes := scanner.Bytes()
		user := User{}

		err := json.Unmarshal(bytes, &user)
		if err != nil {
			fmt.Printf("error: %v", err)
			return messages
		}

		if changeDate, _ := time.Parse(DATE_LAYOUT, user.Packs.ChangeDate); now.Sub(changeDate) > 0 {
			message := tgbotapi.NewMessage(user.ChatId, "Don't forget to change your contact lenses today")
			messages = append(messages, message)
		}
	}

	return messages
}
