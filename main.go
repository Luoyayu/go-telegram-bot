package main

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	dbRedis "github.com/luoyayu/go_telegram_bot/redis-tgbot"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	helpCmds      = "/start\n/redis\n/sayhi\n/home\n/gradios [5]\t默认最新5条"
	dbClient      *redis.Client
	dbClientmutex sync.Mutex
	AliToken      string
)

func initAndTestDB(bot *tgbotapi.BotAPI, err error) {
	if err != nil {
		Logger.Error("<<< redis is still down <<<")
		handleProactiveNotice(bot, "", "redis is down", nil)
	} else {
		Logger.Info("<<< redis is ok <<<")

		if err := dbRedis.SetAllPermissions(dbClient, "voice,gadio,superuser,homedevice"); err != nil {
			Logger.ErrorService("redis", "access all permissions is down")
			handleProactiveNotice(bot, "", "get AllPermissions error: "+err.Error(), nil)
		}

		// new super user
		superUser := dbRedis.User{}
		superUser.SetId(os.Getenv("SUPER_USER_ID"))
		// super user exists
		if dbRedis.CheckUserExist(dbClient, superUser.Id()) {
			logrus.Info("user already exists")
		} else {

			// FIXME get superuser name from env
			superUser.SetName("luoyayu")
			superUser.SetPermissionsMap(map[string]bool{"superuser": true})
			superUser.SetPermissionsStr("superuser")
			// super user not exists

			if err = superUser.Add(dbClient); err != nil {
				Logger.ErrorService("redis", err)
			} else {
				Logger.Info("add super user ok")
			}
		}

		// get ALI_ACCESS_TOKEN from redis
		AliToken = dbClient.Get("ALI_ACCESS_TOKEN").Val()
		Logger.Infof("AliToken in redis: %q", AliToken)

		if AliToken == "" {
			if err := updateTokenAndStore(); err != nil {
				Logger.Error(err)
			}
		} else {
			Logger.InfofService("redis", "ALI_ACCESS_TOKEN ttl: %.0fs", dbClient.TTL("ALI_ACCESS_TOKEN").Val().Seconds())
		}
	}
}

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

	dbClientmutex.Lock()
	dbClient, err = dbRedis.NewRedisClient()
	dbClientmutex.Unlock()

	initAndTestDB(bot, err)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)

	grantedIDs := strings.Split(os.Getenv("GRANTEDIDS"), ",")

	// goroutine for checking redis connection every 1min
	go func() {
		for {
			Logger.Info("test redis connection>>>")

			dbClientmutex.Lock()
			dbClient, err = dbRedis.NewRedisClient()
			dbClientmutex.Unlock()

			initAndTestDB(bot, err)
			time.Sleep(60 * time.Second) // FIXME just for test
		}
	}()

	Logger.Debug("grantedIDs: ", grantedIDs)
	for update := range updates {

		// check if new user
		user := dbRedis.User{}
		var userIDint int
		if update.Message != nil {
			userIDint = update.Message.From.ID
		} else if update.CallbackQuery != nil {
			userIDint = update.CallbackQuery.From.ID
		} else if update.InlineQuery != nil {
			userIDint = update.InlineQuery.From.ID
		}

		user.SetId(fmt.Sprint(userIDint))
		if dbClient == nil {
			Logger.ErrorService("redis", "redis is down when Get user Info")
			handleProactiveNotice(bot, fmt.Sprint(update.Message.From.ID), "redis is down", nil)
		} else {
			if err := user.Get(dbClient); err != nil {
				switch err.Error() {
				case "user id is not exists":
				case "user id is void":
				case "user permissions Str is void":
				default:
					// net error
					Logger.ErrorService("redis", err)
					handleProactiveNotice(bot, fmt.Sprint(update.Message.From.ID),
						"redis is unavailable now", nil)
					continue
				}

				Logger.Info("the user is not in database, err: ", err)
				handleProactiveNotice(bot, fmt.Sprint(update.Message.From.ID),
					"you are unauthorized user, do you want to pull a request?", nil)

				// show new user inline keyboard
				continue // TODO inline keyboard yes or no

			} else {
				Logger.Infof("welcome user [%s][%s]", user.Name(), user.Id())
				Logger.Info("you have Permissions:", user.PermissionsStr())
			}
		}

		// FIXME use redis user:id:permissions to replace the hard code
		chatChecker := "blocked"
		chatMsg := update.Message

		var sendList []tgbotapi.Chattable

		Logger.Infof("[GET UPDATE]: %+v\n", update)

		// if update is callback
		if update.CallbackQuery != nil {
			handleChatCallback(bot, &update)
		}

		// if update is message
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

			// if user is blocked # FIXME
			if chatChecker == "blocked" && chatMsg.Command() != "gradios" {
				replyMsg.Text = "You aren't my master!"
				sticker := tgbotapi.NewStickerShare(chatID, "CAADBQADWAAD1zRtDvbsJxzJHDjDFgQ")
				sendList = append(sendList, sticker)
				goto SendChattableMessages
			}

			// TODO reply random Sticker
			// if message contain Sticker
			if chatMsg.Sticker != nil {
				Logger.Infof("%+v\n", chatMsg.Sticker)
				sticker := tgbotapi.NewStickerShare(chatID, chatMsg.Sticker.FileID)
				sendList = append(sendList, sticker)
				goto SendChattableMessages
			}

			// if message is command
			if chatMsg.IsCommand() {
				err := handleChatCommand(chatMsg, &replyMsg)
				if err != nil {
					Logger.ErrorAndSend(&replyMsg, err)
					goto SendChattableMessages
				}
			}

			// TODO understand sentence
			// if message contains voice
			if chatMsg.Voice != nil {

				var errString string
				if os.Getenv("ALI_ASR_APPKEY") == "" {
					replyMsg.Text = "no asr application binding to this bot!"
					Logger.Warn(replyMsg.Text)
					goto SendChattableMessages
				}

				// convery voice from oga to wav(16K sampling rate)
				if err = handleVoiceMsg(bot, chatMsg); err != nil {
					Logger.Error(&replyMsg, err)
					goto SendChattableMessages
				}

				// first try asr may broken token
				asrResponse := handleVoiceMsg2Text(
					os.Getenv("ALI_ASR_APPKEY"), AliToken,
					"voice.wav", "pcm", os.Getenv("AUDIO_SAMPLING_RATE_ASR"),
					true, true, false)

				// net error
				if asrResponse.Error != nil {
					Logger.ErrorAndSend(&replyMsg, asrResponse.Error)
					goto SendChattableMessages
				}

				if asrResponse.Status == 40000001 {
					Logger.Warnf("Token %q has been Expired, try to update it and retry query Ali ASR service!", AliToken)
					if err := updateTokenAndStore(); err != nil {
						AliToken = ""
						Logger.Error(err)
					} else {
						Logger.Info("get new token: ", AliToken)
						if AliToken == "" {
							Logger.ErrorAndSend(&replyMsg, errString)
							goto SendChattableMessages
						}
						// sencond try asr from broken token
						asrResponse = handleVoiceMsg2Text(
							os.Getenv("ALI_ASR_APPKEY"), AliToken,
							"voice.wav", "pcm", os.Getenv("AUDIO_SAMPLING_RATE_ASR"),
							true, true, false)
					}

				}

				if asrResponse.Status != 20000000 {
					replyMsg.Text = asrResponse.Message
					goto SendChattableMessages
				} else if asrResponse.Result == "" {
					replyMsg.Text = " recognize nothing, please retry"
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
				//SendAndLog(bot, &sendEntity)
				if _, err := bot.Send(sendEntity); err != nil {
					Logger.Warn(err)
				} else {
					Logger.Info("sending to ", replyMsg.ReplyToMessageID, " ", replyMsg.Text)
				}
			}
		}
	}
}
