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

type ChatBot struct {
	Bot *tgbotapi.BotAPI
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
		msg.ReplyMarkup = mainMenuKeyboard
		chatBot.Bot.Send(msg)
	}
}

func (chatBot ChatBot) handleQueryCallbackUpdate(update tgbotapi.Update) {
	var markup tgbotapi.InlineKeyboardMarkup
	var msg tgbotapi.EditMessageTextConfig
	rawStr := strings.Split(update.CallbackQuery.Data, ":")
	field, operation, value := rawStr[0], rawStr[1], rawStr[2]
	intVal := 0

	if len(field) > 0 {
		updateUser(update.CallbackQuery.Message.Chat.ID, field, value)
	}

	switch operation {
	case MENU_OP_ADD, MENU_OP_SUB:
		msg = tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, update.CallbackQuery.Message.Text)
		intVal, _ = strconv.Atoi(value)
		switch field {
		case LAST_OPEN:
			markup = getLastOpenKeyboard(intVal)
		case STOCK:
			markup = getStockKeyboard(intVal)
		}
	case DURATION:
		markup = getDurationKeyboard()
		msg = tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "What is the wear duration?")
	case LAST_OPEN:
		markup = getLastOpenKeyboard(intVal)
		msg = tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "How many days have passed since you opened your current pair of lenses?")
	case STOCK:
		markup = getStockKeyboard(intVal)
		msg = tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "How many additional pair do you have?")
	case MENU_SAVE:
		chat := getChatScope(update.CallbackQuery.Message.Chat.ID)
		res := make(chan User)
		chat.read <- readScope{res: res}
		user := <-res
		lastOpen, _ := time.Parse(DATE_LAYOUT, user.Packs.LastOpenDate)
		user.Packs.ChangeDate = lastOpen.AddDate(0, 0, user.Packs.Duration).Format(DATE_LAYOUT)
		if user.Packs.Stock > 0 {
			user.Packs.LastChangeDate = lastOpen.AddDate(0, 0, (user.Packs.Stock+1)*user.Packs.Duration).Format(DATE_LAYOUT)
		} else {
			user.Packs.LastChangeDate = user.Packs.ChangeDate
		}
		saveUser(user)
		msg = tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, user.String())
	}
	if operation != MENU_SAVE {
		msg.ReplyMarkup = &markup
	}
	chatBot.Bot.Send(msg)
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
