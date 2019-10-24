package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"os"
	"strconv"
)

func handleProactiveNotice(bot *tgbotapi.BotAPI, ) {
	chatID, err := strconv.ParseInt(os.Getenv("CHAT_ID"), 10, 64)
	if err != nil {
		Logger.Error("no chat id")
	} else {
		_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Hello"))
	}
}
