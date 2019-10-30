package main

import (
	"bytes"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	dbRedis "github.com/luoyayu/go_telegram_bot/redis-tgbot"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func handleChatCommand(chatMsg *tgbotapi.Message, replyMsg *tgbotapi.MessageConfig) (err error) {
	switch chatMsg.Command() {
	case "start":
		replyMsg.Text = "ヽ(ﾟ∀ﾟ)ﾒ(ﾟ∀ﾟ)ﾉ /help to start"
	case "help":
		replyMsg.Text = "you can ask me by these commands:\n" + helpCmds
	case "redis":
		if dbClient != nil {
			replyMsg.Text = "(￣.￣)) redis is ok."
		} else {
			replyMsg.Text = "(￣^￣)), redis is down."
		}
	case "sayhi":
		replyMsg.Text = "Hi :)"
	case "home":
		if os.Getenv("SMART_HOME_API_URL") != "" {
			replyMsg.ReplyMarkup = HomeDevicesInlineKeyboard
		} else {
			err = fmt.Errorf("not support control Smart Home")
		}
	case "gradios":
		var text string
		text, err = handleChatGRadios(chatMsg)
		if err != nil {
			return
		} else {
			replyMsg.Text = text
			replyMsg.ReplyMarkup = GRadiosListInlineKeyboard
		}
	case "rss":
		replyMsg.Text = "rss services supported currently:"
		replyMsg.ReplyMarkup = AllRssSupportSubscribeInlineKeyboard

	default:
		replyMsg.Text = "!?(･_･;?"
		replyMsg.ReplyToMessageID = chatMsg.MessageID
	}
	return
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

func handleUserUnFinishedTask(taskName, task string, message *tgbotapi.Message) (err error) {
	if strings.HasPrefix(task, CallbackPrefixRssSub) {
		param := strings.Split(task, "_")
		Logger.Info("user unfinished task: ", param)
		//taskMsgId := param[len(param)-1]

		// TODO
		err = nil

	} else {

	}
	if err != nil {
		dbClient.LPush(taskName, task)
	} else {
		return nil
	}
	return err
}

/*func SendAndLog(bot *tgbotapi.BotAPI, replyMsg *tgbotapi.Chattable) {
	if replyMsg.Text == "" {
		Logger.Error("reply message is void, do nothing")
		return
	}

	_, err := bot.Send(*replyMsg)
	if err != nil {
		Logger.Info("send msg error: ", err)
	} else {
		Logger.Info("sending to ", replyMsg.ReplyToMessageID, replyMsg.Text)
	}
}*/

func checkUserPermissions(user *dbRedis.User, expectPermission string) (withThePermission bool) {
	_, withThePermission = user.PermissionsMap()[expectPermission]
	return
}
