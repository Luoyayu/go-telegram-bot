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

func handleSmartHomeDevices(ops string, dev string) (replyMsg string) {
	apiUrl := os.Getenv("SMART_HOME_API_URL") + "/api/" + dev + "/" + ops
	resp, err := http.PostForm(apiUrl, url.Values{
		"apikey": {os.Getenv("SMART_HOME_APITOKEN")},
	})

	if err != nil {
		Logger.Error(err)
		replyMsg = err.Error()
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Logger.Error(err)
		replyMsg = err.Error()
		return
	}

	switch string(body) {
	case "\"0\"\n":
		replyMsg = "已关闭"
	case "\"1\"\n":
		replyMsg = "开着"
	default:
		Logger.Infof("%q\n", string(body))
		replyMsg = "未知状态"
	}

	switch dev {
	case "light":
		replyMsg = "台灯: " + replyMsg
	case "ps":
		replyMsg = "插线板: " + replyMsg
	default:
		replyMsg = "设备: " + replyMsg
	}
	return
}
