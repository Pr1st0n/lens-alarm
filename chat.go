package main

import (
	"strconv"
	"time"
)

// Validate instead of read. Check fields and save, or ask for missing data
type WriteScope struct {
	key string
	val string
	ack chan bool
}

type ReadScope struct {
	res chan User
}

type ChatScope struct {
	id    int64
	read  chan ReadScope
	write chan WriteScope
}

var chats = make(map[int64]ChatScope)

func GetChat(chatId int64) ChatScope {
	chat, ok := chats[chatId]

	if !ok {
		chat = ChatScope{
			id:    chatId,
			read:  make(chan ReadScope),
			write: make(chan WriteScope),
		}
		go chat.exec()
		chats[chatId] = chat
	}

	return chat
}

func (chat ChatScope) exec() {
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
