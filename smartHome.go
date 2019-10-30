package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

var (
	HomeDevicesInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("light", BtnIdControlLight),
		),
	)

	HomeLightControlInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("open", BtnIDOpeningLight),
			tgbotapi.NewInlineKeyboardButtonData("close", BtnIdClosingLight),
			tgbotapi.NewInlineKeyboardButtonData("status", BtnIdStatusOfLight),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("<<back", BtnIdBackToHomeDevices),
		),
	)
)

func handleSmartHomeDevices(ops string, dev string) (text string, err error) {
	apiUrl := os.Getenv("SMART_HOME_API_URL") + "/api/" + dev + "/" + ops
	resp, err := http.PostForm(apiUrl, url.Values{
		"apikey": {os.Getenv("SMART_HOME_APITOKEN")},
	})

	if err != nil {
		return text, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return text, err
	}

	switch string(body) {
	case "\"0\"\n":
		text = "已关闭"
	case "\"1\"\n":
		text = "开着"
	default:
		Logger.Infof("%q\n", string(body))
		text = "未知状态"
	}

	switch dev {
	case "light":
		text = "台灯: " + text
	case "ps":
		text = "插线板: " + text
	default:
		text = "设备: " + text
	}

	return
}
