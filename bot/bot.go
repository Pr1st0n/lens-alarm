package bot

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
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

type ChatBot struct {
	Bot *tgbotapi.BotAPI
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

func (chatBot ChatBot) HandleUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		chatBot.handleMessageUpdate(update)
	} else if update.CallbackQuery != nil {
		chatBot.handleQueryCallbackUpdate(update)
	}
}

func (chatBot ChatBot) handleMessageUpdate(update tgbotapi.Update) {
	if MENU_NEW == update.Message.Text {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "How can I help you?")
		msg.ReplyMarkup = menuKeyboard
		chatBot.Bot.Send(msg)
	}
}

func (chatBot ChatBot) handleQueryCallbackUpdate(update tgbotapi.Update) {
	rawStr := strings.Split(update.CallbackQuery.Data, ":")
	field, operation, value := rawStr[0], rawStr[1], rawStr[2]

	switch operation {
	case MENU_OP_ADD, MENU_OP_SUB:
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "What is the "+field+"?")
		intVal, _ := strconv.Atoi(value)
		markup := getNumericKeyboard(field, intVal)
		msg.ReplyMarkup = &markup
		chatBot.Bot.Send(msg)
	case MENU_NEXT, MENU_NEW:
		if len(field) > 0 {
			updateUser(update.CallbackQuery.Message.Chat.ID, field, value)
		}

		nextMenu := getNextMenu(field)

		if nextMenu == MENU_SAVE {
			chat := getChatScope(update.CallbackQuery.Message.Chat.ID)
			res := make(chan User)
			chat.read <- readScope{res: res}
			user := <-res
			lastOpen, _ := time.Parse(DATE_LAYOUT, user.Packs.LastOpenDate)
			user.Packs.ChangeDate = lastOpen.AddDate(0, 0, user.Packs.Duration).Format(DATE_LAYOUT)
			if user.Packs.Stock > 0 {
				user.Packs.LastChangeDate = lastOpen.AddDate(0, 0, user.Packs.Stock*user.Packs.Duration).Format(DATE_LAYOUT)
			} else {
				user.Packs.LastChangeDate = user.Packs.ChangeDate
			}
			saveUser(user)
			msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, user.String())
			chatBot.Bot.Send(msg)
		} else {
			msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "What is the "+nextMenu+"?")
			markup := getNumericKeyboard(nextMenu, 1)
			msg.ReplyMarkup = &markup
			chatBot.Bot.Send(msg)
		}
	}
}

func updateUser(chatId int64, key, val string) {
	chat := getChatScope(chatId)
	ack := make(chan bool)
	chat.write <- writeScope{key: key, val: val, ack: ack}
	<-ack
}

func saveUser(user User) {
	f, err := os.OpenFile("data.json", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Errorf("failed to read data.json")
	}
	defer f.Close()

	bytes, _ := json.Marshal(user)
	_, err = f.WriteString(string(bytes) + "\n")
	if err != nil {
		fmt.Errorf("failed to write user data: %s", err)
	}
}
