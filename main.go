package main

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	gadioRss "github.com/luoyayu/go_telegram_bot/gadio-tgbot-plugin"
	logger_tgbot "github.com/luoyayu/go_telegram_bot/logger-tgbot-plugin"
	dbRedis "github.com/luoyayu/go_telegram_bot/redis-tgbot-plugin"
	tg_tgbot "github.com/luoyayu/go_telegram_bot/tg-tgbot-plugin"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	helpCmds = "/start\n/redis\n/sayhi\n/home\n/rss\n/gradios [5]\n/free_jp_v2ray"

	dbClient      *redis.Client
	dbClientMutex sync.Mutex

	RssHubAvailable      bool
	RssHubAvailableMutex sync.Mutex
)

var (
	Logger          = logger_tgbot.NewLogger()
	AliToken        string
	SuperUserID     string
	SuperUserName   string
	SmartHomeApiUrl string
	RSSHubUrl       string
)

func init() {
	SuperUserID = os.Getenv("SUPER_USER_ID")
	SuperUserName = os.Getenv("SUPER_USER_NAME")
	SmartHomeApiUrl = os.Getenv("SMART_HOME_API_URL")
	RSSHubUrl = os.Getenv("RSSHub_Url")

}

func initAndTestDB(bot *tgbotapi.BotAPI, err error) {
	if err != nil {
		Logger.Error("<<< redis is down >>>")
	} else {
		Logger.Info("<<< redis is up >>>")

		/*if err := dbRedis.SetAllPermissions(dbClient, "voice,gadio,superuser,homedevice"); err != nil {
			Logger.ErrorService("redis", "access all permissions is down")
		}*/

		// new super user
		superUser := dbRedis.User{}
		superUser.SetId(SuperUserID)
		if superUser.Id() == "" {
			Logger.FatalfService("init service", "SUPER_USER_ID is void")
		}

		// super user exists
		if dbRedis.CheckUserExist(dbClient, superUser.Id()) {
			Logger.Info("superuser already exists")
		} else {
			if SuperUserName == "" {
				SuperUserName = "superuser"
			}
			superUser.SetName(SuperUserName)
			superUser.SetPermissionsMap(map[string]bool{"superuser": true})
			superUser.SetPermissionsStr("superuser")
			// super user not exists

			if err = superUser.Add(dbClient); err != nil {
				Logger.ErrorService("redis", err)
			} else {
				Logger.Info("add super user ok")
			}
		}

		if AliToken = dbClient.Get("ALI_ACCESS_TOKEN").Val(); AliToken == "" {
			if err := updateTokenAndStore(); err != nil {
				Logger.Error(err)
			}
		}
		Logger.InfofService("redis", "ALI_ACCESS_TOKEN ttl: %.0fs", dbClient.TTL("ALI_ACCESS_TOKEN").Val().Seconds())
	}
}

func main() {
	var bot *tgbotapi.BotAPI
	var err error

	if bot, err = tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN")); err != nil {
		Logger.Fatal(err)
	}

	bot.Debug = false
	if os.Getenv("BOT_DEBUG") == "true" {
		bot.Debug = true
	}

	Logger.Infof("Authorized on account Name: %s", bot.Self.UserName)
	Logger.Infof("Authorized on account ID: %s", bot.Self.ID)

	dbClientMutex.Lock()
	dbClient, err = dbRedis.NewRedisClient()
	dbClientMutex.Unlock()

	// the error is from dbRedis.NewRedisClient()
	initAndTestDB(bot, err)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)

	if err != nil {
		Logger.ErrorService("bot", "get update filed!")
	}

	go func() {
		for {
			if radios, err := gadioRss.GetGRadios(5); err == nil && dbClient != nil && radios != nil {
				gadioRss.GadioQueryAndStoreAndSend(bot, radios, dbClient, true, "", Logger, tg_tgbot.ProactiveNotice)
			} else {
				if err != nil {
					Logger.ErrorService("gadio", err.Error())
				}
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
			time.Sleep(5 * time.Minute) // FIXME just for test
		}
	}()

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
			tg_tgbot.ProactiveNotice(bot, fmt.Sprint(update.Message.From.ID), "service is not available now.", nil, Logger)
		} else {
			if err := user.Get(dbClient); err != nil {
				if strings.HasPrefix(err.Error(), "redis") {
					Logger.ErrorService("redis", err)
					tg_tgbot.ProactiveNotice(bot, fmt.Sprint(update.Message.From.ID),
						"user service is not available now.", nil, Logger)
					continue
				} else if err.Error() != "" {
					Logger.Info("the user info is not in database, err: ", err)
					// show Register Inline Keyboard
					showRegisterInlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("YES", BtnIdRegister),
							tgbotapi.NewInlineKeyboardButtonData("No", BtnIdNotRegister),
						),
					)
					opVal := dbClient.SAdd(dbRedis.RedisKeys.UserNotAuthorizedAccess, user.Id()).Val()
					Logger.Infof("add the user[%s] to %s: %d\n", user.Id(), dbRedis.RedisKeys.UserNotAuthorizedAccess, opVal)

					if opVal == 0 { // user still in UserNotAuthorizedAccess
						logrus.Infof("the user[%s] is already in %s\n", user.Id(), dbRedis.RedisKeys.UserNotAuthorizedAccess)
						// ! allow UserNotAuthorizedAccess goto handleChatCallback
					} else if opVal == 1 {
						tg_tgbot.ProactiveNotice(bot, user.Id(),
							"you are unauthorized user, do you want to pull a request?", &showRegisterInlineKeyboard, Logger)
						continue
					}
				}
			} else {
				Logger.Infof("welcome user [%s][%s]", user.Name(), user.Id())
				Logger.Infof("the user[%s] have Permissions: %s", user.Name(), user.PermissionsStr())
			}

			// TODO check user has unfinished task
			/*if userTasksStackName := "user:" + user.Id() + ":tasks"; dbClient.LLen(userTasksStackName).Val() != 0 {
				if update.Message != nil {
					err = HandleUserUnFinishedTask(userTasksStackName, dbClient.LPop(userTasksStackName).Val(), update.Message)
					if err != nil {
						Logger.Warnf("user task ", userTasksStackName, " still unfinished")
					} else {
						Logger.Info("finish user all task")
					}
				} else {
					tg_tgbot.ProactiveNotice(bot, user.Id(), "the input is illegal", nil, Logger)
				}
				continue
			} else {
				Logger.Info("user don't have unfinished task")
			}*/
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

			// user replay to something, must handle it now!
			if chatMsg.ReplyToMessage != nil {
				Logger.Warnf("user reply: %q, ans is : %q\n", chatMsg.ReplyToMessage.Text, chatMsg.Text)
				handleChatReply(bot, &user, &update)
				continue
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
				err := HandleChatCommand(bot, chatMsg, &replyMsg)
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
				if err = HandleVoiceMsg(bot, chatMsg); err != nil {
					Logger.Error(&replyMsg, err)
					goto SendChattableMessages
				}

				// first try, asr may broken token
				asrResponse := handleVoiceMsg2Text(
					os.Getenv("ALI_ASR_APPKEY"), AliToken,
					"./tmp/voice.wav", "pcm", os.Getenv("AUDIO_SAMPLING_RATE_ASR"),
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
							"./tmp/voice.wav", "pcm", os.Getenv("AUDIO_SAMPLING_RATE_ASR"),
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
				if resp, err := bot.Send(sendEntity); err != nil {
					Logger.Warn(err)
				} else {
					go func() {
						time.Sleep(time.Second * 30)
						if chatMsg.Command() == "free_jp_v2ray" {
							if _, err := bot.DeleteMessage(tgbotapi.DeleteMessageConfig{
								ChatID:    resp.Chat.ID,
								MessageID: resp.MessageID,
							}); err != nil {
								Logger.ErrorService("delete vmess code msg i sent", err)
							} else {
								Logger.Info("delete vmess code msg i sent ok!")
							}
						}
					}()

					Logger.Info("sending to ", replyMsg.ChatID, " ", replyMsg.Text)
				}
			}
		}
	}
}
