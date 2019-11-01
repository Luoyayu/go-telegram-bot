package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/luoyayu/go_telegram_bot/gadio-tgbot-plugin"
	"strconv"
	"strings"
)

var (
	GRadios *gadioRss.Radios

	UpdateGRadiosList = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("update", BtnIdUpdateGRadiosList),
			tgbotapi.NewInlineKeyboardButtonData(">>close", BtnIdClose),
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
			return fmt.Errorf("net is down")
		}

		for _, radio := range *GRadios.Data {
			oneRowOneBtn(radio.Attributes.Title, fmt.Sprint("radio", radio.ID), &GRadiosListInlineKeyboard)
			//radiosTitle += strings.Split(radio.Attributes.PublishedAt, "T")[0] + "\n\t"
			//radiosTitle += radio.Attributes.Title + "\n"
		}
		oneRowOneBtn(">>close", BtnIdCloseGRadiosInlineKeyboard, &GRadiosListInlineKeyboard)
	}
	return nil
}

func handleChatGRadios(message *tgbotapi.Message) (text string, err error) {
	radiosNum := 5
	Logger.Infof("command with arguments: %q\n", message.CommandArguments())
	commandArgument, err := strconv.Atoi(strings.TrimSpace(message.CommandArguments()))
	if err == nil {
		radiosNum = commandArgument
	} else if strings.TrimSpace(message.CommandArguments()) != "" {
		Logger.Error(err)
	}

	if err = newGRadioListInlineKeyboard(radiosNum); err != nil {
		Logger.Error(err)
		return
	}

	text = "机核近期电台\n"
	return
}
