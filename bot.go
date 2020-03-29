package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
	"strings"
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
	bot *tgbotapi.BotAPI
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
		chatBot.bot.Send(msg)
	}
}

func (chatBot ChatBot) handleQueryCallbackUpdate(update tgbotapi.Update) {
	rawStr := strings.Split(update.CallbackQuery.Data, ":")
	field, operation, value := rawStr[0], rawStr[1], rawStr[2]

	switch operation {
	case MENU_OP_ADD, MENU_OP_SUB:
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "What is the wearing duration?")
		intVal, _ := strconv.Atoi(value)
		markup := getNumericKeyboard(field, intVal)
		msg.ReplyMarkup = &markup
		chat := GetChat(update.CallbackQuery.Message.Chat.ID)
		ack := make(chan bool)
		chat.write <- WriteScope{key: field, val: value, ack: ack}
		<-ack
		chatBot.bot.Send(msg)
	case MENU_NEXT, MENU_NEW:
		nextMenu := getNextMenu(field)
		if nextMenu == MENU_SAVE {
			chat := GetChat(update.CallbackQuery.Message.Chat.ID)
			res := make(chan User)
			chat.read <- ReadScope{res: res}
			user := <-res
			msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, user.String())
			chatBot.bot.Send(msg)
		} else {
			msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "What is the "+nextMenu+"?")
			markup := getNumericKeyboard(nextMenu, 1)
			msg.ReplyMarkup = &markup
			chatBot.bot.Send(msg)
		}
	}
}
