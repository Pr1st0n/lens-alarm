package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
)

const (
	DATE_LAYOUT = "2006-01-02"
	MENU_NEW    = "/menu"
	MENU_OP_ADD = "menu-add"
	MENU_OP_SUB = "menu-sub"
	MENU_SAVE   = "menu-save"
	DURATION    = "duration"
	STOCK       = "stock"
	LAST_OPEN   = "last-open"
)

var mainMenuKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Add new lens pack", ":"+DURATION+":"),
	),
)

func getDurationKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("7", DURATION+":"+LAST_OPEN+":"+"7"),
			tgbotapi.NewInlineKeyboardButtonData("14", DURATION+":"+LAST_OPEN+":"+"14"),
			tgbotapi.NewInlineKeyboardButtonData("30", DURATION+":"+LAST_OPEN+":"+"30"),
		),
	)
}

func getLastOpenKeyboard(value int) tgbotapi.InlineKeyboardMarkup {
	strVal := strconv.Itoa(value)
	prev := value - 1
	prevPrev := value - 5

	if prev < 0 {
		prev = 0
		prevPrev = 0
	} else if prevPrev < 0 {
		prevPrev = 0
	}

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("<<", LAST_OPEN+":"+MENU_OP_SUB+":"+strconv.Itoa(prevPrev)),
			tgbotapi.NewInlineKeyboardButtonData("<", LAST_OPEN+":"+MENU_OP_SUB+":"+strconv.Itoa(prev)),
			tgbotapi.NewInlineKeyboardButtonData(strVal, LAST_OPEN+":"+STOCK+":"+strVal),
			tgbotapi.NewInlineKeyboardButtonData(">", LAST_OPEN+":"+MENU_OP_ADD+":"+strconv.Itoa(value+1)),
			tgbotapi.NewInlineKeyboardButtonData(">>", LAST_OPEN+":"+MENU_OP_ADD+":"+strconv.Itoa(value+5)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("next", LAST_OPEN+":"+STOCK+":"+strVal),
		),
	)
}

func getStockKeyboard(value int) tgbotapi.InlineKeyboardMarkup {
	strVal := strconv.Itoa(value)
	prev := value - 1
	prevPrev := value - 5

	if prev < 0 {
		prev = 0
		prevPrev = 0
	} else if prevPrev < 0 {
		prevPrev = 0
	}

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("<<", STOCK+":"+MENU_OP_SUB+":"+strconv.Itoa(prevPrev)),
			tgbotapi.NewInlineKeyboardButtonData("<", STOCK+":"+MENU_OP_SUB+":"+strconv.Itoa(prev)),
			tgbotapi.NewInlineKeyboardButtonData(strVal, STOCK+":"+MENU_SAVE+":"+strVal),
			tgbotapi.NewInlineKeyboardButtonData(">", STOCK+":"+MENU_OP_ADD+":"+strconv.Itoa(value+1)),
			tgbotapi.NewInlineKeyboardButtonData(">>", STOCK+":"+MENU_OP_ADD+":"+strconv.Itoa(value+5)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("next", STOCK+":"+MENU_SAVE+":"+strVal),
		),
	)
}
