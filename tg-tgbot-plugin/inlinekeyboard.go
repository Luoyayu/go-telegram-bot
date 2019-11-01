package tg_tgbot_plugin

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	logger_tgbot "github.com/luoyayu/go_telegram_bot/logger-tgbot-plugin"
	"strings"
)


func OneRowOneBtn(btnText string, callbackTextID string, inlineKeyboardMarkup *tgbotapi.InlineKeyboardMarkup, logger logger_tgbot.ILogger) {
	var keyboardRow = make([]tgbotapi.InlineKeyboardButton, 1)
	if strings.HasSuffix(callbackTextID, "url") == true {
		logger.Info("find url in ", btnText)
		g := make([]tgbotapi.InlineKeyboardButton, 1)
		g[0] = tgbotapi.NewInlineKeyboardButtonURL(btnText, btnText)
		inlineKeyboardMarkup.InlineKeyboard = append(inlineKeyboardMarkup.InlineKeyboard, g)
	} else {
		keyboardRow[0].Text = btnText
		keyboardRow[0].CallbackData = &callbackTextID
		inlineKeyboardMarkup.InlineKeyboard = append(inlineKeyboardMarkup.InlineKeyboard, keyboardRow)
	}
}
