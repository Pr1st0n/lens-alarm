package bot

import (
	"bufio"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"os"
	"strconv"
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
	return fmt.Sprintf("Duration: %d, In use since: %s, Stock: %d, Next change: %s, Last change: %s",
		user.Packs.Duration,
		user.Packs.LastOpenDate,
		user.Packs.Stock,
		user.Packs.ChangeDate,
		user.Packs.LastChangeDate)
}

// Validate instead of read. Check fields and save, or ask for missing data
type writeScope struct {
	key string
	val string
	ack chan bool
}

type readScope struct {
	res chan User
}

type chatScope struct {
	id    int64
	read  chan readScope
	write chan writeScope
}

var chats = make(map[int64]chatScope)

func getChatScope(chatId int64) chatScope {
	chat, ok := chats[chatId]

	if !ok {
		chat = chatScope{
			id:    chatId,
			read:  make(chan readScope),
			write: make(chan writeScope),
		}
		go chat.exec()
		chats[chatId] = chat
	}

	return chat
}

func (chat chatScope) exec() {
	user := User{ChatId: chat.id}

	for {
		select {
		case data := <-chat.write:
			switch data.key {
			case "duration":
				user.Packs.Duration, _ = strconv.Atoi(data.val)
				data.ack <- true
			case "last-open":
				now := time.Now()
				days, _ := strconv.Atoi(data.val)
				lastOpen := now.AddDate(0, 0, -days)
				user.Packs.LastOpenDate = lastOpen.Format(DATE_LAYOUT)
				data.ack <- true
			case "stock":
				user.Packs.Stock, _ = strconv.Atoi(data.val)
				data.ack <- true
			default:
				data.ack <- false
			}
		case data := <-chat.read:
			data.res <- user
		}
	}
}

func getNotifications() []tgbotapi.MessageConfig {
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

func (chatBot ChatBot) SendNotification() {
	for _, msg := range getNotifications() {
		chatBot.Bot.Send(msg)
	}
}
