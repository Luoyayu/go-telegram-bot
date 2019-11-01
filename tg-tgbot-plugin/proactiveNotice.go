package tg_tgbot_plugin

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	logger_tgbot "github.com/luoyayu/go_telegram_bot/logger-tgbot-plugin"
	"os"
	"strconv"
)

func ProactiveNotice(bot *tgbotapi.BotAPI, userid string, messageText string, inlineKeyboard *tgbotapi.InlineKeyboardMarkup, Logger logger_tgbot.ILogger) {
	var userID64 int64
	var err error
	if userid == "" {
		userID64, err = strconv.ParseInt(os.Getenv("SUPER_USER_ID"), 10, 64)
	} else {
		userID64, err = strconv.ParseInt(userid, 10, 64)
	}
	if err != nil {
		Logger.Error("no chat id")
	} else {
		msg := tgbotapi.NewMessage(userID64, messageText)
		if inlineKeyboard != nil {
			msg.ReplyMarkup = inlineKeyboard
		}
		_, _ = bot.Send(msg)
	}
}
