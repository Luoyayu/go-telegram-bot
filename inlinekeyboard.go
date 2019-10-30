package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
)


func oneRowOneBtn(btnText string, callbackTextID string, inlineKeyboardMarkup *tgbotapi.InlineKeyboardMarkup) {
	var keyboardRow = make([]tgbotapi.InlineKeyboardButton, 1)
	if strings.HasSuffix(callbackTextID, "url") == true {
		Logger.Info("find url in ", btnText)
		g := make([]tgbotapi.InlineKeyboardButton, 1)
		g[0] = tgbotapi.NewInlineKeyboardButtonURL(btnText, btnText)
		inlineKeyboardMarkup.InlineKeyboard = append(inlineKeyboardMarkup.InlineKeyboard, g)
	} else {
		keyboardRow[0].Text = btnText
		keyboardRow[0].CallbackData = &callbackTextID
		inlineKeyboardMarkup.InlineKeyboard = append(inlineKeyboardMarkup.InlineKeyboard, keyboardRow)
	}
}
