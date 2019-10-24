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
			tgbotapi.NewInlineKeyboardButtonData("light", "control light"),
		),
	)

	HomeLightControlInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("open", "opening light"),
			tgbotapi.NewInlineKeyboardButtonData("close", "closing light"),
			tgbotapi.NewInlineKeyboardButtonData("status", "status of light"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("<<back", "back to home devices"),
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
