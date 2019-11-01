package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/luoyayu/go_telegram_bot/gadio-tgbot-plugin"
	dbRedis "github.com/luoyayu/go_telegram_bot/redis-tgbot-plugin"
	tg_tgbot "github.com/luoyayu/go_telegram_bot/tg-tgbot-plugin"

	"strconv"
	"strings"
	"time"
)

// Callback Suffix with `light`
const (
	CallbackSuffixLight = "light"
	BtnIdControlLight   = "control light"
	BtnIDOpeningLight   = "opening light"
	BtnIdClosingLight   = "closing light"
	BtnIdStatusOfLight  = "status of light"
)

// Callback Prefix with `back to`
const (
	CallbackPrefixBackTo                  = "back to"
	BtnIdBackToRadiosInfo                 = "back to radios info"
	BtnIdBackToHomeDevices                = "back to home devices"
	BtnIdBackToAllRssSupportingSubscribes = "back to all rss supporting subscribes"
)

// Callback Prefix with `close`
const (
	CallbackPrefixClose             = "close"
	BtnIdClose                      = "close"
	BtnIdCloseGRadiosInlineKeyboard = "close GRadios inline keyboard"
)

// Callback Prefix with `update`
const (
	CallbackPrefixUpdate   = "update"
	BtnIdUpdateGRadiosList = "update GRadios list"
)

const (
	CallbackSuffixRegister = "Register"
	BtnIdRegister          = "Register"
	BtnIdNotRegister       = "NotRegister"
)

// Callback Prefix with `radio`
const (
	CallbackPrefixRadio = "radio"
)

const (
	CallbackPrefixRssSub = "rssSub_"
)

// change replay markup for specified chat id
func editMessageReplyMarkup(
	chatCallbackQuery *tgbotapi.CallbackQuery,
	newReplyMarkup *tgbotapi.InlineKeyboardMarkup) tgbotapi.EditMessageReplyMarkupConfig {

	editText := tgbotapi.NewEditMessageReplyMarkup(
		chatCallbackQuery.Message.Chat.ID,
		chatCallbackQuery.Message.MessageID,
		tgbotapi.InlineKeyboardMarkup{})
	editText.ReplyMarkup = newReplyMarkup
	return editText
}

func handleChatCallback(bot *tgbotapi.BotAPI, user *dbRedis.User, update *tgbotapi.Update) {
	callbackQuery := update.CallbackQuery
	callbackID := callbackQuery.Data
	Logger.Info("callbackID: ", callbackID)

	_, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(callbackQuery.ID, callbackID))

	if err != nil {
		Logger.Errorf("get Answer from Callback Query Error: %+v\n", err)
		return
	}

	if strings.HasPrefix(callbackID, "_") == true { // no need to reply if callback data with "_"
		return
	}

	replyMsg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, callbackID)

	// user not Authorized access
	if user.PermissionsStr() == "" {
		goto HandleRegister
	}

	// callback with Smart Home Devices
	if strings.HasSuffix(callbackID, CallbackSuffixLight) {
		var lightError error
		switch callbackID {
		case BtnIdControlLight:
			_, lightError = bot.Send(editMessageReplyMarkup(callbackQuery, &HomeLightControlInlineKeyboard))
		case BtnIDOpeningLight:
			replyMsg.Text, lightError = handleSmartHomeDevices("on", "light")
		case BtnIdClosingLight:
			replyMsg.Text, lightError = handleSmartHomeDevices("off", "light")
		case BtnIdStatusOfLight:
			replyMsg.Text, lightError = handleSmartHomeDevices("status", "light")
		}

		if lightError != nil {
			Logger.ErrorAndSend(&replyMsg, lightError)
		} else if replyMsg.Text == "" {
			Logger.ErrorAndSend(&replyMsg, "Unknown error!")
		} else {
			return
		}
		goto ReplyMsg
	} else
	// callback with `back to ...` inline keyboard
	if strings.HasPrefix(callbackID, CallbackPrefixBackTo) {
		var backToError error
		switch callbackID {
		case BtnIdBackToRadiosInfo:
			_, backToError = bot.Send(editMessageReplyMarkup(callbackQuery, &GRadiosListInlineKeyboard))
		case BtnIdBackToHomeDevices:
			_, backToError = bot.Send(editMessageReplyMarkup(callbackQuery, &HomeDevicesInlineKeyboard))
		case BtnIdBackToAllRssSupportingSubscribes:
			_, backToError = bot.Send(editMessageReplyMarkup(callbackQuery, &AllRssSupportSubscribeInlineKeyboard))
		}
		if backToError != nil {
			Logger.ErrorAndSend(&replyMsg, backToError)
			goto ReplyMsg
		} else {
			return
		}
	} else
	// callback with `close ...`
	if strings.HasPrefix(callbackID, CallbackPrefixClose) {
		var closeError error
		switch callbackID {
		case BtnIdCloseGRadiosInlineKeyboard:
			_, closeError = bot.Send(editMessageReplyMarkup(callbackQuery, &UpdateGRadiosList))
		default:
			_, closeError = bot.Send(editMessageReplyMarkup(callbackQuery, nil))
		}
		if closeError != nil {
			Logger.ErrorAndSend(&replyMsg, closeError)
		} else {
			return
		}
		goto ReplyMsg
	} else
	// callback with `update ...`
	if strings.HasPrefix(callbackID, CallbackPrefixUpdate) {
		var updateError error
		switch callbackID {
		case BtnIdUpdateGRadiosList:
			if updateError = newGRadioListInlineKeyboard(5); updateError != nil {
				Logger.ErrorAndSend(&replyMsg, err)
				goto ReplyMsg
			} else {
				_, updateError = bot.Send(editMessageReplyMarkup(callbackQuery, &GRadiosListInlineKeyboard))
			}
		}
		if updateError != nil {
			Logger.ErrorAndSend(&replyMsg, updateError)
		} else {
			return
		}
		goto ReplyMsg
	} else
	// callback with `radio ...`
	if strings.HasPrefix(callbackID, CallbackPrefixRadio) {
		radioId := callbackID[5:]
		radioUrl := "https://www.gcores.com/radios/" + radioId
		var radioSelected gadioRss.RadioDataEntity

		Logger.Info("radioUrl: ", radioUrl)
		if GRadios == nil {
			Logger.ErrorAndSend(&replyMsg, "Please retry /gradios to get latest info")
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
			Logger.ErrorAndSend(&replyMsg, "not found this radio [%d] info!", radioId)
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
			"_" + callbackID + "_date":     "发布日期: " + strings.Split(radioSelected.Attributes.PublishedAt, "T")[0],
			"_" + callbackID + "_title":    "标题: " + radioSelected.Attributes.Title,
			"_" + callbackID + "_desc":     "描述: " + radioSelected.Attributes.Desc,
			"_" + callbackID + "_url":      radioUrl,
			"_" + callbackID + "_duration": "时长: " + radioDuration,
		}
		gRadioInfoKeyboard := tgbotapi.NewInlineKeyboardMarkup()

		for callbackText, btnText := range btnEntities {
			tg_tgbot.OneRowOneBtn(btnText, callbackText, &gRadioInfoKeyboard, Logger)
		}
		tg_tgbot.OneRowOneBtn(">>back", BtnIdBackToRadiosInfo, &gRadioInfoKeyboard, Logger)
		_, err = bot.Send(editMessageReplyMarkup(callbackQuery, &gRadioInfoKeyboard))
		if err != nil {
			Logger.ErrorAndSend(&replyMsg, err)
		} else {
			return
		}
		goto ReplyMsg
	} else
	//
	if strings.HasPrefix(callbackID, CallbackPrefixRssSub) {
		// rssSub_bilibili
		// rssSub_bilibili_user_2333
		// rssSub_bilibili_live_1
		requestRssList := strings.Split(callbackID, "_")

		requestRss := requestRssList[1:]
		requestRssName := requestRss[0]
		show := false
		if len(requestRss) == 1 {
			show = true
		}

		if show {
			if v, ok := RssHubAllSubMap[requestRssName]; ok == false {
				Logger.ErrorfService("rssHub", "not found %s in Rss Hub All Sub Map", requestRssName)
			} else {
				rssAvailableSubsInlineKeyboard := tgbotapi.NewInlineKeyboardMarkup()
				for subSubName, subSubItem := range v {
					Logger.Infof("%+v %+v\n", subSubName, subSubItem)
					Logger.Infof("%+v\n", subSubItem)
					tg_tgbot.OneRowOneBtn(
						subSubItem["path"]+subSubItem["help"], CallbackPrefixRssSub+requestRssName+"_"+subSubName,
						&rssAvailableSubsInlineKeyboard, Logger)
				}
				tg_tgbot.OneRowOneBtn("<<back", "back to all rss supporting subscribes",
					&rssAvailableSubsInlineKeyboard, Logger)
				_, err := bot.Send(editMessageReplyMarkup(callbackQuery, &rssAvailableSubsInlineKeyboard))
				if err != nil {
					Logger.Error(err)
				}
				return
			}
		} else {
			params := requestRss
			Logger.Infof("%+v\n", params)
			// store message state by Chat ID
			Logger.Info(update.UpdateID)

			// push params to list user:user.Id():task_msgID
			dbClient.LPush("user:"+user.Id()+":tasks", callbackID+"_"+fmt.Sprint(callbackQuery.Message.MessageID))
			replyMsg.Text = "please input by the format: `:` mean the param is necessary" + RssHubAllSubMap[requestRssName][params[0]]["help"]

		}

	}

HandleRegister:
	if strings.HasSuffix(callbackID, CallbackSuffixRegister) {
		var registerErr error

		switch callbackID {
		// TODO
		case BtnIdRegister:
			// TODO
		case BtnIdNotRegister:
			_, registerErr = bot.Send(editMessageReplyMarkup(callbackQuery, nil))
			dbClient.SRem(dbRedis.RedisKeys.UserNotAuthorizedAccess, user.Id())
		}

		if registerErr != nil {
			// TODO
		} else {
			return
		}

	}

ReplyMsg:
	if replyMsg.Text == "" {
		Logger.Error("replyMsg is void, do nothing!")
		return
	}
	_, err = bot.Send(replyMsg)
	if err != nil {
		Logger.Info("send msg error! ", err)
	} else {
		Logger.Info("sending to ", replyMsg.ChatID, replyMsg.Text)
	}
}
