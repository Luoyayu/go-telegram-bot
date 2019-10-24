package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"os"
	"strconv"
	"strings"
)

var (
	helpCmds = "/start\n/status\n/sayhi\n/home\n/gradios [5]\t默认最新5条"
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
