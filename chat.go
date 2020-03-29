package main

import (
	"strconv"
	"time"
)

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

func getChat(chatId int64) chatScope {
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
			case "lastOpenDate":
				user.Packs.LastOpenDate = time.Now().String()
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
