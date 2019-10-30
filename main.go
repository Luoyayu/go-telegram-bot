package main

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	gadioRss "github.com/luoyayu/go_telegram_bot/gadio-rss"
	logger_tgbot "github.com/luoyayu/go_telegram_bot/logger-tgbot"
	dbRedis "github.com/luoyayu/go_telegram_bot/redis-tgbot"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	helpCmds             = "/start\n/redis\n/sayhi\n/home\n/gradios [5]\t默认最新5条"
	dbClient             *redis.Client
	dbClientMutex        sync.Mutex
	AliToken             string
	RssHubAvailable      bool
	RssHubAvailableMutex sync.Mutex
)

var (
	Logger = logger_tgbot.NewLogger()
)

func initAndTestDB(bot *tgbotapi.BotAPI, err error) {
	if err != nil {
		Logger.Error("<<< redis is still down <<<")
		proactiveNotice(bot, "", "redis is down", nil)
	} else {
		Logger.Info("<<< redis is ok <<<")

		if err := dbRedis.SetAllPermissions(dbClient, "voice,gadio,superuser,homedevice"); err != nil {
			Logger.ErrorService("redis", "access all permissions is down")
			proactiveNotice(bot, "", "get AllPermissions error: "+err.Error(), nil)
		}

		// new super user
		superUser := dbRedis.User{}
		superUser.SetId(os.Getenv("SUPER_USER_ID"))
		// super user exists
		if dbRedis.CheckUserExist(dbClient, superUser.Id()) {
			Logger.Info("superuser already exists")
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

	dbClientMutex.Lock()
	dbClient, err = dbRedis.NewRedisClient()
	dbClientMutex.Unlock()

	initAndTestDB(bot, err)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)

	grantedIDs := strings.Split(os.Getenv("GRANTEDIDS"), ",")

	go func() {
		for {

			if radios, err := gadioRss.GetGRadios(5); err == nil && dbClient != nil {
				GadioStore(bot, radios, dbClient)
			} else {
				Logger.ErrorService("gradio", err.Error())
			}
			time.Sleep(time.Minute * 20)
		}
	}()

	// goroutine for checking redis connection every 1min
	go func() {
		for {
			Logger.Info("test redis connection>>>")

			dbClientMutex.Lock()
			dbClient, err = dbRedis.NewRedisClient()
			dbClientMutex.Unlock()

			initAndTestDB(bot, err)
			if err := checkRssHubAvailable(); err != nil {
				Logger.ErrorService("rssHub", err)
			} else {
				getAllRssSupportedSubscribe()
			}
			time.Sleep(1 * time.Minute) // FIXME just for test
		}
	}()

	Logger.Debug("grantedIDs: ", grantedIDs)
	for update := range updates {

		// check if update is rom new user
		user := dbRedis.User{}
		var userIDint int
		if update.Message != nil {
			userIDint = update.Message.From.ID
		} else if update.CallbackQuery != nil {
			userIDint = update.CallbackQuery.From.ID
		} else if update.InlineQuery != nil {
			userIDint = update.InlineQuery.From.ID
		}
		Logger.Info("update.UpdateID: ", update.UpdateID)

		user.SetId(fmt.Sprint(userIDint))
		if dbClient == nil {
			Logger.ErrorService("redis", "redis is down when Get user Info")
			proactiveNotice(bot, fmt.Sprint(update.Message.From.ID), "redis is down", nil)
		} else {
			if err := user.Get(dbClient); err != nil {
				if err.Error() == "redis error" {
					Logger.ErrorService("redis", err)
					proactiveNotice(bot, fmt.Sprint(update.Message.From.ID),
						"redis is unavailable now", nil)
					continue
				} else if err.Error() != "" {
					Logger.Info("the user info is not in database, err: ", err)
					showRegisterInlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("YES", BtnIdRegister),
							tgbotapi.NewInlineKeyboardButtonData("No", BtnIdNotRegister),
						),
					)
					opVal := dbClient.SAdd(dbRedis.RedisKeys.UserNotAuthorizedAccess, user.Id()).Val()
					Logger.Infof("add the user[%s] to %s: %d\n", user.Id(), dbRedis.RedisKeys.UserNotAuthorizedAccess, opVal)

					if opVal == 0 {
						logrus.Infof("the user[%s] is already in %s\n", user.Id(), dbRedis.RedisKeys.UserNotAuthorizedAccess)
						// allow goto handleChatCallback
					} else if opVal == 1 {
						proactiveNotice(bot, user.Id(),
							"you are unauthorized user, do you want to pull a request?", &showRegisterInlineKeyboard)
						continue
					}
				}
			} else {
				Logger.Infof("welcome user [%s][%s]", user.Name(), user.Id())
				Logger.Info("you have Permissions:", user.PermissionsStr())
			}
		}

		// TODO check user has unfinished task
		if userTasksStackName := "user:" + user.Id() + ":tasks"; dbClient.LLen(userTasksStackName).Val() != 0 {
			if update.Message != nil {
				err = handleUserUnFinishedTask(userTasksStackName, dbClient.LPop(userTasksStackName).Val(), update.Message)
				if err != nil {
					Logger.Warnf("user task ", userTasksStackName, " still unfinished")
				} else {
					Logger.Info("finish user all task")
				}
			} else {
				proactiveNotice(bot, user.Id(), "the input is illegal", nil)
			}
			continue
		} else {
			Logger.Info("user don't have unfinished task")
		}

		// FIXME use redis user:id:permissions to replace the hard code
		//chatAuthorized := "blocked"
		chatMsg := update.Message

		var sendList []tgbotapi.Chattable

		Logger.Infof("[GET UPDATE]: %+v\n", update)

		// if update is callback
		if update.CallbackQuery != nil {
			handleChatCallback(bot, &user, &update)
		}

		// if update is message
		if chatMsg != nil {
			chatID := update.Message.Chat.ID
			fromID := update.Message.From.ID

			Logger.Info("update.Message.Chat.ID: ", chatID)
			Logger.Info("update.Message.From.ID: ", fromID)
			replyMsg := tgbotapi.NewMessage(chatID, chatMsg.Text)
			//for _, grantedID := range grantedIDs {
			//	if strconv.Itoa(fromID) == strings.TrimSpace(grantedID) {
			//		chatAuthorized = "granted"
			//		break
			//	}
			//}

			// if user is blocked # FIXME
			//if chatAuthorized == "blocked" && chatMsg.Command() != "gradios" {
			//	replyMsg.Text = "You aren't my master!"
			//	sticker := tgbotapi.NewStickerShare(chatID, "CAADBQADWAAD1zRtDvbsJxzJHDjDFgQ")
			//	sendList = append(sendList, sticker)
			//	goto SendChattableMessages
			//}

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

				// first try, asr may broken token
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
						// second try, asr from broken token
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
			Logger.Infof("[%d]|[%s]", fromID, chatMsg.From.UserName)
			sendList = append(sendList, replyMsg)

			for _, sendEntity := range sendList {
				//SendAndLog(bot, &sendEntity)
				if _, err := bot.Send(sendEntity); err != nil {
					Logger.Warn(err)
				} else {
					Logger.Info("sending to ", replyMsg.ChatID, " ", replyMsg.Text)
				}
			}
		}
	}
}
