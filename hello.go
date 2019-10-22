package main

import (
	"bytes"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var (
	myHomeKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("light", "control light"),
		),
	)

	myHomeLightKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("open", "opening light"),
			tgbotapi.NewInlineKeyboardButtonData("close", "closing light"),
			tgbotapi.NewInlineKeyboardButtonData("status", "status of light"),
		),
	)
	helpCmds = "/start\n/status\n/sayhi\n/home"
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
			replyMsg.Text = "ヽ(ﾟ∀ﾟ)ﾒ(ﾟ∀ﾟ)ﾉ"
		case "help":
			replyMsg.Text = "you can ask me by these commands:\n" + helpCmds
		case "status":
			replyMsg.Text = "(￣.￣)), I'm fine."
		case "sayhi":
			replyMsg.Text = "Hi :)"
		case "home":
			if os.Getenv("SMART_HOME_API_URL") != "" {
				replyMsg.ReplyMarkup = myHomeKeyboard
			} else {
				replyMsg.Text = "Not support control Smart Home!"
			}
		default:
			replyMsg.Text = "!?(･_･;?"
			replyMsg.ReplyToMessageID = message.MessageID
		}
	}
}

func handleChatCallback(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	log.Printf("Callback update: %+v\n", update)
	resp, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))
	if err != nil {
		log.Printf("get Answer from Callback Query Error: %+v\n", err)
	}
	log.Printf("Callback response: %+v\n", resp)
	replyMsg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data)
	switch update.CallbackQuery.Data {
	case "control light":
		log.Println("show control light")
		replyMsg.Text = "support operations"
		replyMsg.ReplyMarkup = myHomeLightKeyboard
	case "opening light":
		log.Println("opening light")
		replyMsg.Text = handHomeDevices("on", "light")
	case "closing light":
		log.Println("closing light")
		replyMsg.Text = handHomeDevices("off", "light")
	case "status of light":
		log.Println("status of light")
		replyMsg.Text = handHomeDevices("status", "light")

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
