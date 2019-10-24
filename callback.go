package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/luoyayu/go_telegram_bot/gadio-rss"

	"strconv"
	"strings"
	"time"
)

func handleEditMessageReplyMarkup(chatCallbackQuery *tgbotapi.CallbackQuery, newReplyMarkup *tgbotapi.InlineKeyboardMarkup) tgbotapi.EditMessageReplyMarkupConfig {
	editText := tgbotapi.NewEditMessageReplyMarkup(
		chatCallbackQuery.Message.Chat.ID,
		chatCallbackQuery.Message.MessageID,
		tgbotapi.InlineKeyboardMarkup{})
	editText.ReplyMarkup = newReplyMarkup
	return editText
}

func handleChatCallback(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	callbackQuery := update.CallbackQuery
	callbackData := callbackQuery.Data

	_, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(callbackQuery.ID, callbackData))

	if err != nil {
		Logger.Errorf("get Answer from Callback Query Error: %+v\n", err)
		return
	}

	if strings.HasPrefix(callbackData, "_") == true { // no need to reply if callback data with "_"
		return
	}

	replyMsg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, callbackData)

	// callback with Smart Home Devices
	if strings.HasSuffix(callbackData, "light") == true {
		switch callbackData {
		case "control light":
			_, err = bot.Send(handleEditMessageReplyMarkup(callbackQuery, &HomeLightControlInlineKeyboard))
			return
		case "opening light":
			replyMsg.Text = handleSmartHomeDevices("on", "light")
		case "closing light":
			replyMsg.Text = handleSmartHomeDevices("off", "light")
		case "status of light":
			replyMsg.Text = handleSmartHomeDevices("status", "light")
		}
		goto ReplyMsg
	} else
	// callback with `back to ...` inline keyboard
	if strings.HasPrefix(callbackData, "back to") == true {
		switch callbackData {
		case "back to radios info":
			_, err = bot.Send(handleEditMessageReplyMarkup(callbackQuery, &GRadiosListInlineKeyboard))
			return
		case "back to home devices":
			_, err = bot.Send(handleEditMessageReplyMarkup(callbackQuery, &HomeDevicesInlineKeyboard))
			return
		}

	} else
	// callback with `close ...`
	if strings.HasPrefix(callbackData, "close") == true {
		switch callbackData {
		case "close GRadios inline keyboard":
			_, err = bot.Send(handleEditMessageReplyMarkup(callbackQuery, &UpdateGRadiosList))
			return
		}
	} else
	// callback with `update ...`
	if strings.HasPrefix(callbackData, "update") == true {
		switch callbackData {
		case "update GRadios list":
			if err := newGRadioListInlineKeyboard(5); err != nil {
				Logger.Error(err)
				replyMsg.Text = "update error"
				goto ReplyMsg
			} else {
				_, err = bot.Send(handleEditMessageReplyMarkup(callbackQuery, &GRadiosListInlineKeyboard))
				return
			}
		}
	}

	if strings.HasPrefix(callbackData, "radio") == true {
		radioId := callbackData[5:]
		radioUrl := "https://www.gcores.com/radios/" + radioId
		var radioSelected gadioRss.RadioDataEntity

		Logger.Info("radioUrl: ", radioUrl)
		if GRadios == nil {
			replyMsg.Text = "Please retry this /gradios to get latest info"
			goto ReplyMsg
		}

		for _, gRadio := range *GRadios.Data {
			if gRadio.ID == radioId {
				radioSelected = gRadio
				Logger.Info("radio found in response body")
				break
			}
		}

		if radioSelected.ID == "" {
			replyMsg.Text = "not found this radio info!"
			goto ReplyMsg
		}

		var btnEntities map[string]string
		var radioDuration string

		dur, err := time.ParseDuration(strconv.Itoa(radioSelected.Attributes.Duration) + "s")
		if err != nil {
			Logger.Warn(err)
			radioDuration = strconv.Itoa(radioSelected.Attributes.Duration)
		} else {
			radioDuration = dur.String()
		}

		btnEntities = map[string]string{
			"_" + callbackData + "_date":     "发布日期: " + strings.Split(radioSelected.Attributes.PublishedAt, "T")[0],
			"_" + callbackData + "_title":    "标题: " + radioSelected.Attributes.Title,
			"_" + callbackData + "_desc":     "描述: " + radioSelected.Attributes.Desc,
			"_" + callbackData + "_url":      radioUrl,
			"_" + callbackData + "_duration": "时长: " + radioDuration,
		}
		gRadioInfoKeyboard := tgbotapi.NewInlineKeyboardMarkup()

		for callbackText, btnText := range btnEntities {
			handleOneRowOneBtn(btnText, callbackText, &gRadioInfoKeyboard)
		}
		handleOneRowOneBtn(">>back", "back to radios info", &gRadioInfoKeyboard)
		_, err = bot.Send(handleEditMessageReplyMarkup(callbackQuery, &gRadioInfoKeyboard))
		return
	}

ReplyMsg:
	_, err = bot.Send(replyMsg)
	if err != nil {
		Logger.Info("send msg error! ", err)
	} else {
		Logger.Info("sending to ", replyMsg.ReplyToMessageID, replyMsg.Text)
	}
}
