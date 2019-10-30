package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"os"
	"strconv"
)

func proactiveNotice(bot *tgbotapi.BotAPI, userid string, messageText string, inlineKeyboard *tgbotapi.InlineKeyboardMarkup) {
	var userid64 int64
	var err error
	if userid == "" {
		userid64, err = strconv.ParseInt(os.Getenv("SUPER_USER_ID"), 10, 64)
	} else {
		userid64, err = strconv.ParseInt(userid, 10, 64)
	}
	if err != nil {
		Logger.Error("no chat id")
	} else {
		msg := tgbotapi.NewMessage(userid64, messageText)
		if inlineKeyboard != nil {
			msg.ReplyMarkup = inlineKeyboard
		}
		_, _ = bot.Send(msg)
	}
}
