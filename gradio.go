package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/luoyayu/go_telegram_bot/gadio-rss"

	"strconv"
	"strings"
)

var (
	GRadios *gadioRss.Radios

	UpdateGRadiosList = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("update", "update GRadios list"),
		),
	)

	GRadiosListInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup()
)

func newGRadioListInlineKeyboard(radiosNum int) error {
	GRadiosListInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup()
	var err error
	GRadios, err = gadioRss.GetGRadios(radiosNum)

	if err != nil {
		Logger.Error(err)
		return err
	} else {
		if GRadios == nil {
			return fmt.Errorf("net is down!")
		}

		for _, radio := range *GRadios.Data {
			handleOneRowOneBtn(radio.Attributes.Title, fmt.Sprint("radio", radio.ID), &GRadiosListInlineKeyboard)
			//radiosTitle += strings.Split(radio.Attributes.PublishedAt, "T")[0] + "\n\t"
			//radiosTitle += radio.Attributes.Title + "\n"
		}
		handleOneRowOneBtn("close", "close GRadios inline keyboard", &GRadiosListInlineKeyboard)
	}
	return nil
}

func handleChatGRadios(message *tgbotapi.Message, replyMsg *tgbotapi.MessageConfig) string {
	radiosNum := 5
	Logger.Info("CommandArguments: ", message.CommandArguments())
	if commandArgument, err := strconv.Atoi(strings.TrimSpace(message.CommandArguments())); err == nil {
		radiosNum = commandArgument
	} else if strings.TrimSpace(message.CommandArguments()) != "" {
		Logger.Warn(err)
		return "wrong Command Arguments:" + message.CommandArguments()
	}
	if err := newGRadioListInlineKeyboard(radiosNum); err != nil {
		Logger.Error(err)
		return err.Error()
	}
	replyMsg.ReplyMarkup = GRadiosListInlineKeyboard
	return "机核近期电台\n"
}
