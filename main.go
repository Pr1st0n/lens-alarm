package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DATE_LAYOUT = "2006-01-02"
	MENU_NEW    = "/menu"
	MENU_OP_ADD = "menu-add"
	MENU_OP_SUB = "menu-sub"
	MENU_NEXT   = "menu-next"
	MENU_SAVE   = "menu-save"
	DURATION    = "duration"
	STOCK       = "stock"
	LAST_OPEN   = "last-open"
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

var menuKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Add new lens pack", ":"+MENU_NEW+":"),
	),
)

func getNumericKeyboard(prefix string, value int) tgbotapi.InlineKeyboardMarkup {
	strVal := strconv.Itoa(value)
	prev := value

	if prev-1 < 1 {
		prev = 1
	} else {
		prev -= 1
	}

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("-", prefix+":"+MENU_OP_SUB+":"+strconv.Itoa(prev)),
			tgbotapi.NewInlineKeyboardButtonData(strVal, prefix+":"+MENU_NEXT+":"+strVal),
			tgbotapi.NewInlineKeyboardButtonData("+", prefix+":"+MENU_OP_ADD+":"+strconv.Itoa(value+1)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("next", prefix+":"+MENU_NEXT+":"+strVal),
		),
	)
}

func getNextMenu(cur string) string {
	switch cur {
	case DURATION:
		return LAST_OPEN
	case LAST_OPEN:
		return STOCK
	case STOCK:
		return MENU_SAVE
	default:
		return DURATION
	}
}

func main() {
	token := os.Getenv("LENS_API_TOKEN")
	bot, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	go ScheduleJob(func() {
		for _, msg := range getMessages() {
			bot.Send(msg)
		}
	})

	for update := range updates {
		if update.Message != nil {
			if MENU_NEW == update.Message.Text {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "How can I help you?")
				msg.ReplyMarkup = menuKeyboard
				bot.Send(msg)
			}
		} else if update.CallbackQuery != nil {
			rawStr := strings.Split(update.CallbackQuery.Data, ":")
			field, operation, value := rawStr[0], rawStr[1], rawStr[2]

			switch operation {
			case MENU_OP_ADD, MENU_OP_SUB:
				msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "What is the wearing duration?")
				intVal, _ := strconv.Atoi(value)
				markup := getNumericKeyboard(field, intVal)
				msg.ReplyMarkup = &markup
				chat := getChat(update.CallbackQuery.Message.Chat.ID)
				ack := make(chan bool)
				chat.write <- writeScope{key: field, val: value, ack: ack}
				<-ack
				bot.Send(msg)
			case MENU_NEXT, MENU_NEW:
				nextMenu := getNextMenu(field)
				if nextMenu == MENU_SAVE {
					chat := getChat(update.CallbackQuery.Message.Chat.ID)
					res := make(chan User)
					chat.read <- readScope{res: res}
					user := <-res
					msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, user.String())
					bot.Send(msg)
				} else {
					msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "What is the "+nextMenu+"?")
					markup := getNumericKeyboard(nextMenu, 1)
					msg.ReplyMarkup = &markup
					bot.Send(msg)
				}
			}
		}
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
