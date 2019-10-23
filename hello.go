package main

import (
	"bytes"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	gadioRss "github.com/luoyayu/go_telegram_bot/gadio-rss"
	"io"
	"io/ioutil"
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
	updateGRadiosList         = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("update", "update GRadios list"),
		),
	)

	helpCmds = "/start\n/status\n/sayhi\n/home\n/gradios [5]\t默认最新5条"
	gRadios  *gadioRss.Radios
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		Logger.Fatal(err)
	}
	bot.Debug = false
	if os.Getenv("BOT_DEBUG") == "true" {
		bot.Debug = true
	}

	Logger.Infof("Authorized on account %s", bot.Self.UserName)
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

	Logger.Debug("grantedIDs: ", grantedIDs)
	AliToken := os.Getenv("ALI_ACCESS_TOKEN")

	for update := range updates {

		chatChecker := "blocked"
		chatMsg := update.Message

		var sendList []tgbotapi.Chattable

		Logger.Infof("[GET UPDATE]: %+v\n", update)

		if update.CallbackQuery != nil {
			handleChatCallback(bot, &update)
		}

		if chatMsg != nil {
			chatID := update.Message.Chat.ID
			fromID := update.Message.From.ID
			replyMsg := tgbotapi.NewMessage(chatID, chatMsg.Text)
			for _, grantedID := range grantedIDs {
				if strconv.Itoa(fromID) == strings.TrimSpace(grantedID) {
					chatChecker = "granted"
					break
				}
			}

			if chatChecker == "blocked" && chatMsg.Command() != "gradios" {
				replyMsg.Text = "You aren't my master!"
				sticker := tgbotapi.NewStickerShare(chatID, "CAADBQADWAAD1zRtDvbsJxzJHDjDFgQ")
				sendList = append(sendList, sticker)
				goto SendChattableMessages
			}

			if chatMsg.Sticker != nil {
				Logger.Infof("%+v\n", chatMsg.Sticker)
				sticker := tgbotapi.NewStickerShare(chatID, chatMsg.Sticker.FileID)
				sendList = append(sendList, sticker)
			}

			if chatMsg.IsCommand() {
				handleChatCommand(chatMsg, &replyMsg)
			}

			if chatMsg.Voice != nil {
				var errString string
				if os.Getenv("ALI_ASR_APPKEY") == "" {
					replyMsg.Text = "no asr application binding to this bot!"
					Logger.Warn(replyMsg.Text)
					goto SendChattableMessages
				}

				if err = handleVoiceMsg(bot, chatMsg); err != nil {
					replyMsg.Text = err.Error()
					Logger.Error(replyMsg.Text)
					goto SendChattableMessages
				}

				if AliToken == "" {
					AliToken, errString = GetTokenFromSDK()
					if AliToken == "" {
						replyMsg.Text = errString
						goto SendChattableMessages
					}
				}

				asrResponse := handleVoiceMsg2Text(
					os.Getenv("ALI_ASR_APPKEY"), AliToken,
					"voice.wav", "pcm", os.Getenv("AUDIO_SAMPLING_RATE_ASR"),
					true, true, false)

				// net error
				if asrResponse.Error != nil {
					replyMsg.Text = asrResponse.Error.Error()
					Logger.Error(replyMsg.Text)
					goto SendChattableMessages
				}

				if asrResponse.Status == 40000001 {
					Logger.Warn("Token has been Expired, try to update it and retry query Ali ASR service!")
					AliToken, errString = GetTokenFromSDK()
					Logger.Info("get new token: ", AliToken)
					if AliToken == "" {
						replyMsg.Text = errString
						goto SendChattableMessages
					}
					asrResponse = handleVoiceMsg2Text(
						os.Getenv("ALI_ASR_APPKEY"), AliToken,
						"voice.wav", "pcm", os.Getenv("AUDIO_SAMPLING_RATE_ASR"),
						true, true, false)
				}

				if asrResponse.Status != 20000000 || asrResponse.Result == "" {
					replyMsg.Text = asrResponse.Message
					goto SendChattableMessages
				} else {
					replyMsg.Text = "you have said: " + asrResponse.Result
					Logger.Info(replyMsg.Text)
				}
			}

		SendChattableMessages:
			Logger.Infof("[%d]|[%s]|[%s]", fromID, chatMsg.From.UserName, chatChecker)
			sendList = append(sendList, replyMsg)

			for _, sendEntity := range sendList {
				if _, err := bot.Send(sendEntity); err != nil {
					Logger.Error("send msg error! ", err)
				} else {
					Logger.Info("sending to ", replyMsg.ReplyToMessageID, " ", replyMsg.Text)
				}
			}
		}
	}
}

func convertOga2Wav48K(fileNameWOExt string) (error, string) {
	cmd := exec.Command("ffmpeg", "-y", "-i",
		fileNameWOExt+".oga", "-ar",
		os.Getenv("AUDIO_SAMPLING_RATE_ASR"),
		fileNameWOExt+".wav")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		Logger.Error(err, stderr.String())
	} else {
		return err, stderr.String()
	}
	return nil, ""
}

func handleVoiceMsg(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
	Logger.Info("get voice message, fileID: ", msg.Voice.FileID)
	voiceFileDirectUrl, err := bot.GetFileDirectURL(msg.Voice.FileID)
	ogaFile, _ := bot.GetFile(tgbotapi.FileConfig{FileID: msg.Voice.FileID})
	Logger.Info("file url:", voiceFileDirectUrl)

	resp, err := http.Get(voiceFileDirectUrl)
	if err != nil {
		Logger.Error("Download voice from telegram server error: ", err)
	} else {
		fileNameWOExtension := "voice"
		file, _ := os.Create(fileNameWOExtension + ".oga")
		n, _ := io.Copy(file, resp.Body)
		if n != int64(ogaFile.FileSize) {
			errorString := fmt.Sprintf("download size: %d\tvoice file  in telegram server size: %d", n, ogaFile.FileSize)
			Logger.Error(errorString)
			return fmt.Errorf(errorString)
		} else {
			err, _ := convertOga2Wav48K(fileNameWOExtension)
			if err != nil {
				Logger.Error(err)
				return err
			}
			Logger.Info("Write telegram voice to wav file", fileNameWOExtension)
		}
	}
	return nil
}

func handleChatCommand(chatMsg *tgbotapi.Message, replyMsg *tgbotapi.MessageConfig) {
	switch chatMsg.Command() {
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
		replyMsg.Text = handleChatGRadios(chatMsg, replyMsg)
	default:
		replyMsg.Text = "!?(･_･;?"
		replyMsg.ReplyToMessageID = chatMsg.MessageID
	}

}
func handleEditMessageReplyMarkup(chatCallbackQuery *tgbotapi.CallbackQuery, newReplyMarkup *tgbotapi.InlineKeyboardMarkup) tgbotapi.EditMessageReplyMarkupConfig {
	editText := tgbotapi.NewEditMessageReplyMarkup(
		chatCallbackQuery.Message.Chat.ID,
		chatCallbackQuery.Message.MessageID,
		tgbotapi.InlineKeyboardMarkup{})
	editText.ReplyMarkup = newReplyMarkup
	return editText

}

func handleOneRowOneBtn(btnText string, callbackText string, inlineKeyboardMarkup *tgbotapi.InlineKeyboardMarkup) {
	var keyboardRow = make([]tgbotapi.InlineKeyboardButton, 1)
	if strings.HasSuffix(callbackText, "url") == true {
		Logger.Info("find url in ", btnText)
		g := make([]tgbotapi.InlineKeyboardButton, 1)
		g[0] = tgbotapi.NewInlineKeyboardButtonURL(btnText, btnText)
		inlineKeyboardMarkup.InlineKeyboard = append(inlineKeyboardMarkup.InlineKeyboard, g)
	} else {
		keyboardRow[0].Text = btnText
		keyboardRow[0].CallbackData = &callbackText
		inlineKeyboardMarkup.InlineKeyboard = append(inlineKeyboardMarkup.InlineKeyboard, keyboardRow)
	}
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
			_, err = bot.Send(handleEditMessageReplyMarkup(callbackQuery, &homeLightControlInlineKeyboard))
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
			_, err = bot.Send(handleEditMessageReplyMarkup(callbackQuery, &gRadiosListInlineKeyboard))
			return
		case "back to home devices":
			_, err = bot.Send(handleEditMessageReplyMarkup(callbackQuery, &homeDevicesInlineKeyboard))
			return
		}

	} else
	// callback with `close ...`
	if strings.HasPrefix(callbackData, "close") == true {
		switch callbackData {
		case "close GRadios inline keyboard":
			_, err = bot.Send(handleEditMessageReplyMarkup(callbackQuery, &updateGRadiosList))
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
				_, err = bot.Send(handleEditMessageReplyMarkup(callbackQuery, &gRadiosListInlineKeyboard))
				return
			}
		}
	}

	if strings.HasPrefix(callbackData, "radio") == true {
		radioId := callbackData[5:]
		radioUrl := "https://www.gcores.com/radios/" + radioId
		var radioSelected gadioRss.RadioDataEntity

		Logger.Info("radioUrl: ", radioUrl)
		if gRadios == nil {
			replyMsg.Text = "Please retry this /gradios to get latest info"
			goto ReplyMsg
		}

		for _, gRadio := range *gRadios.Data {
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

func newGRadioListInlineKeyboard(radiosNum int) error {
	gRadiosListInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup()
	var err error
	gRadios, err = gadioRss.GetGRadios(radiosNum)

	if err != nil {
		Logger.Error(err)
		return err
	} else {
		for _, radio := range *gRadios.Data {
			handleOneRowOneBtn(radio.Attributes.Title, fmt.Sprint("radio", radio.ID), &gRadiosListInlineKeyboard)
			//radiosTitle += strings.Split(radio.Attributes.PublishedAt, "T")[0] + "\n\t"
			//radiosTitle += radio.Attributes.Title + "\n"
		}
		handleOneRowOneBtn("close", "close GRadios inline keyboard", &gRadiosListInlineKeyboard)
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
	replyMsg.ReplyMarkup = gRadiosListInlineKeyboard
	return "机核近期电台\n"
}

func handleProactiveNotice(bot *tgbotapi.BotAPI, ) {
	chatID, err := strconv.ParseInt(os.Getenv("CHAT_ID"), 10, 64)
	if err != nil {
		Logger.Error("no chat id")
	} else {
		_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Hello"))
	}

}
