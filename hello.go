package main

import (
	"bytes"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	gadio_rss "github.com/luoyayu/go_telegram_bot/gadio-rss"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	homeDevicesInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("light", "control light"),
		),
	)

	homeLightControlInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("open", "opening light"),
			tgbotapi.NewInlineKeyboardButtonData("close", "closing light"),
			tgbotapi.NewInlineKeyboardButtonData("status", "status of light"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("<<back", "back to home devices"),
		),
	)

	gRadiosListInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup()
	helpCmds                  = "/start\n/status\n/sayhi\n/home\n/gradios [5]\t默认最新5条"
	gRadios                   *gadio_rss.Radios
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		log.Fatalln(err)
	}
	bot.Debug = false
	if os.Getenv("BOT_DEBUG") == "true" {
		bot.Debug = true
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)

	grantedIDs := strings.Split(os.Getenv("GRANTEDIDS"), ",")

	/*go func() {
		for {
			handleProactiveNotice(bot)
			time.Sleep(60 * time.Second)
		}
	}()*/

	log.Println(grantedIDs)
	AliToken := os.Getenv("ALI_ACCESS_TOKEN")

	for update := range updates {

		checkChat := "blocked"
		var sendList []tgbotapi.Chattable

		log.Printf("[GET UPDATE]: %+v\n", update)

		if update.CallbackQuery != nil {
			handleChatCallback(bot, &update)
		}

		if update.Message != nil {
			replyMsg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			for _, grantedID := range grantedIDs {
				if strconv.Itoa(update.Message.From.ID) == strings.TrimSpace(grantedID) {
					checkChat = "granted"
					break
				}
			}

			if checkChat == "blocked" {
				replyMsg.Text = "You aren't my master!"
				sticker := tgbotapi.NewStickerShare(update.Message.Chat.ID, "CAADBQADWAAD1zRtDvbsJxzJHDjDFgQ")
				sendList = append(sendList, sticker)
				goto SendChattableMessage
			}

			if update.Message.Sticker != nil {
				log.Printf("%+v\n", update.Message.Sticker)
				sticker := tgbotapi.NewStickerShare(update.Message.Chat.ID, update.Message.Sticker.FileID)
				sendList = append(sendList, sticker)
			}

			if update.Message.IsCommand() {
				handleChatCmd(update.Message, &replyMsg)
			}

			if update.Message.Voice != nil {
				if os.Getenv("ALI_ASR_APPKEY") == "" {
					replyMsg.Text = "no asr application binding to this bot!"
					goto SendChattableMessage
				}

				if err = handleVoiceMsg(bot, update.Message); err != nil {
					replyMsg.Text = err.Error()
					goto SendChattableMessage
				}

				if AliToken == "" {
					var errString string
					AliToken, errString = GetTokenFromSDK()
					if AliToken == "" {
						replyMsg.Text = errString
						goto SendChattableMessage
					}
				}

				asrRet := handleVoiceMsg2Text(
					os.Getenv("ALI_ASR_APPKEY"), AliToken,
					"voice.wav", "pcm", os.Getenv("AUDIO_SAMPLING_RATE_ASR"),
					true, true, false)

				// net error
				if asrRet.Error != nil {
					replyMsg.Text = asrRet.Error.Error()
					goto SendChattableMessage
				}

				if asrRet.Status == 40000001 {
					log.Println("Token has been Expired, try to update it and retry ASR!")
					var errString string
					AliToken, errString = GetTokenFromSDK()
					log.Println("new token: ", AliToken)
					if AliToken == "" {
						replyMsg.Text = errString
						goto SendChattableMessage
					}
					asrRet = handleVoiceMsg2Text(
						os.Getenv("ALI_ASR_APPKEY"), AliToken,
						"voice.wav", "pcm", os.Getenv("AUDIO_SAMPLING_RATE_ASR"),
						true, true, false)
				}

				if asrRet.Status != 20000000 || asrRet.Result == "" {
					replyMsg.Text = asrRet.Message
					goto SendChattableMessage
				} else {
					replyMsg.Text = "you have said: " + asrRet.Result
				}
			}

		SendChattableMessage:
			log.Printf("[%d]|[%s]|[%s]", update.Message.From.ID, update.Message.From.UserName, checkChat)
			sendList = append(sendList, replyMsg)

			for _, sendEntity := range sendList {
				if _, err := bot.Send(sendEntity); err != nil {
					log.Println("send msg error! ", err)
				} else {
					log.Println("sending to ", replyMsg.ReplyToMessageID, replyMsg.Text)
				}
			}
		}
	}
}

func convertOga2Wav48K(fileNameWOExt string) (error, string) {
	cmd := exec.Command("ffmpeg", "-y", "-i", fileNameWOExt+".oga", "-ar", os.Getenv("AUDIO_SAMPLING_RATE_ASR"), fileNameWOExt+".wav")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Println(err, stderr.String())
	} else {
		return err, stderr.String()
	}
	return nil, ""
}

func handleVoiceMsg(bot *tgbotapi.BotAPI, update *tgbotapi.Message) error {
	log.Println("get voice message, fileID: ", update.Voice.FileID)
	fileUrl, err := bot.GetFileDirectURL(update.Voice.FileID)
	ogaFile, _ := bot.GetFile(tgbotapi.FileConfig{
		FileID: update.Voice.FileID,
	})
	log.Println("file url:", fileUrl)
	resp, err := http.Get(fileUrl)
	if err != nil {
		log.Println("download from telegram error: ", err)
	} else {
		fileNameWOExtension := "voice"
		file, _ := os.Create(fileNameWOExtension + ".oga")
		n, _ := io.Copy(file, resp.Body)
		if n != int64(ogaFile.FileSize) {
			log.Println("download size: ", n, "\ttelegram file size:", ogaFile.FileSize)
		} else {
			err, _ := convertOga2Wav48K(fileNameWOExtension)
			if err != nil {
				return err
			}
			log.Println("Write telegram voice to wav file", fileNameWOExtension)
		}
	}
	return nil
}

func handleChatCmd(message *tgbotapi.Message, replyMsg *tgbotapi.MessageConfig) {
	if message.IsCommand() {
		switch message.Command() {
		case "start":
			replyMsg.Text = "ヽ(ﾟ∀ﾟ)ﾒ(ﾟ∀ﾟ)ﾉ /help to start"
		case "help":
			replyMsg.Text = "you can ask me by these commands:\n" + helpCmds
		case "status":
			replyMsg.Text = "(￣.￣)), I'm fine."
		case "sayhi":
			replyMsg.Text = "Hi :)"
		case "home":
			if os.Getenv("SMART_HOME_API_URL") != "" {
				replyMsg.ReplyMarkup = homeDevicesInlineKeyboard
			} else {
				replyMsg.Text = "Not support control Smart Home!"
			}
		case "gradios":
			replyMsg.Text = handleGetGRadios(message, replyMsg)

		default:
			replyMsg.Text = "!?(･_･;?"
			replyMsg.ReplyToMessageID = message.MessageID
		}
	}
}
func handleEditMessageReplyMarkup(callbackQ *tgbotapi.CallbackQuery, newReplyMarkup *tgbotapi.InlineKeyboardMarkup) tgbotapi.EditMessageReplyMarkupConfig {
	editText := tgbotapi.NewEditMessageReplyMarkup(
		callbackQ.Message.Chat.ID,
		callbackQ.Message.MessageID,
		tgbotapi.InlineKeyboardMarkup{})
	editText.ReplyMarkup = newReplyMarkup
	return editText

}

func handleOneRowOneBtn(text string, callbackData string, keyboard *tgbotapi.InlineKeyboardMarkup) {
	var keyboardRow = make([]tgbotapi.InlineKeyboardButton, 1)
	if strings.HasSuffix(callbackData, "url") == true {
		log.Println("find url in ", text)
		g := make([]tgbotapi.InlineKeyboardButton, 1)
		g[0] = tgbotapi.NewInlineKeyboardButtonURL(text, text)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, g)
	} else {
		keyboardRow[0].Text = text
		keyboardRow[0].CallbackData = &callbackData
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, keyboardRow)
	}

}

func handleChatCallback(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	callbackQuery := update.CallbackQuery
	_, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(callbackQuery.ID, callbackQuery.Data))
	if err != nil {
		log.Printf("get Answer from Callback Query Error: %+v\n", err)
	}
	log.Printf("Callback Data: %+v\n", callbackQuery.Data)

	if strings.HasPrefix(callbackQuery.Data, "_") == true { // no need to reply
		return
	}

	replyMsg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, callbackQuery.Data)
	switch callbackQuery.Data {
	case "control light":
		_, err = bot.Send(handleEditMessageReplyMarkup(callbackQuery, &homeLightControlInlineKeyboard))
		return
	case "back to home devices":
		_, err = bot.Send(handleEditMessageReplyMarkup(callbackQuery, &homeDevicesInlineKeyboard))
		return
	case "opening light":
		replyMsg.Text = handHomeDevices("on", "light")
	case "closing light":
		replyMsg.Text = handHomeDevices("off", "light")
	case "status of light":
		replyMsg.Text = handHomeDevices("status", "light")
	case "close inline keyboard":
		_, err = bot.Send(handleEditMessageReplyMarkup(callbackQuery, nil))
		return
	case "back to radios info":
		_, err = bot.Send(handleEditMessageReplyMarkup(callbackQuery, &gRadiosListInlineKeyboard))
		return
	}

	if strings.HasPrefix(callbackQuery.Data, "radio") == true {
		radioId := callbackQuery.Data[5:]
		radioUrl := "https://www.gcores.com/radios/" + radioId
		var radioInCallback gadio_rss.RadioDataEntity

		log.Println(radioUrl)
		for _, gRadio := range *gRadios.Data {
			if gRadio.ID == radioId {
				radioInCallback = gRadio
				log.Println("found in database!")
				break
			}
		}
		var btnEntitis map[string]string
		var radioDuration string

		dur, err := time.ParseDuration(strconv.Itoa(radioInCallback.Attributes.Duration) + "s")
		if err != nil {
			radioDuration = strconv.Itoa(radioInCallback.Attributes.Duration)
		} else {
			radioDuration = dur.String()
		}

		btnEntitis = map[string]string{
			"_" + callbackQuery.Data + "_date":     "发布日期: " + strings.Split(radioInCallback.Attributes.PublishedAt, "T")[0],
			"_" + callbackQuery.Data + "_title":    "标题: " + radioInCallback.Attributes.Title,
			"_" + callbackQuery.Data + "_desc":     "描述: " + radioInCallback.Attributes.Desc,
			"_" + callbackQuery.Data + "_url":      radioUrl,
			"_" + callbackQuery.Data + "_duration": "时长: " + radioDuration,
		}
		gRadioInfoKeyboard := tgbotapi.NewInlineKeyboardMarkup()

		for callbackText, btnText := range btnEntitis {
			handleOneRowOneBtn(btnText, callbackText, &gRadioInfoKeyboard)
		}
		handleOneRowOneBtn("返回", "back to radios info", &gRadioInfoKeyboard)
		_, err = bot.Send(handleEditMessageReplyMarkup(callbackQuery, &gRadioInfoKeyboard))
		return
	}

	_, err = bot.Send(replyMsg)
	if err != nil {
		log.Println("send msg error! ", err)
	} else {
		log.Println("sending to ", replyMsg.ReplyToMessageID, replyMsg.Text)
	}
}

func handHomeDevices(ops string, dev string) (replyMsg string) {
	apiUrl := os.Getenv("SMART_HOME_API_URL") + "/api/" + dev + "/" + ops
	resp, err := http.PostForm(apiUrl, url.Values{
		"apikey": {os.Getenv("SMART_HOME_APITOKEN")},
	})

	if err != nil {
		replyMsg = err.Error()
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		replyMsg = err.Error()
		return
	}

	switch string(body) {
	case "\"0\"\n":
		replyMsg = "已关闭"
	case "\"1\"\n":
		replyMsg = "开着"
	default:
		log.Printf("%q\n", string(body))
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

func handleGetGRadios(message *tgbotapi.Message, replyMsg *tgbotapi.MessageConfig) string {
	radiosNum := 5
	gRadiosListInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup()
	log.Println("CommandArguments: ", message.CommandArguments())
	if commandArgument, err := strconv.Atoi(strings.TrimSpace(message.CommandArguments())); err == nil {
		radiosNum = commandArgument
	} else if strings.TrimSpace(message.CommandArguments()) != "" {
		return "wrong Command Arguments:" + message.CommandArguments()
	}
	var err error
	gRadios, err = gadio_rss.GetGRadios(radiosNum)
	//radiosTitle := ""

	if err != nil {
		return err.Error()
	} else {
		for _, radio := range *gRadios.Data {
			handleOneRowOneBtn(radio.Attributes.Title, fmt.Sprint("radio", radio.ID), &gRadiosListInlineKeyboard)

			//radiosTitle += strings.Split(radio.Attributes.PublishedAt, "T")[0] + "\n\t"
			//radiosTitle += radio.Attributes.Title + "\n"
		}
		handleOneRowOneBtn("关闭", "close inline keyboard", &gRadiosListInlineKeyboard)
	}
	replyMsg.ReplyMarkup = gRadiosListInlineKeyboard
	return "机核近期电台情报\n"
}

func handleProactiveNotice(bot *tgbotapi.BotAPI, ) {
	chatID, err := strconv.ParseInt(os.Getenv("CHAT_ID"), 10, 64)
	if err != nil {
		log.Println("no chat id")
	} else {
		_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Hello"))
	}

}
