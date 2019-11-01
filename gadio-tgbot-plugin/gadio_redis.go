package gadioRss

import (
	"github.com/go-redis/redis/v7"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	logger_tgbot "github.com/luoyayu/go_telegram_bot/logger-tgbot-plugin"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

// update data and included to redis

func GadioQueryAndStoreAndSend(bot *tgbotapi.BotAPI, gadio *Radios, rc *redis.Client, send bool, sendToUser string, Logger logger_tgbot.ILogger,
	proactiveNotice func(*tgbotapi.BotAPI, string, string, *tgbotapi.InlineKeyboardMarkup)) {
	// redis Hash store
	Logger.InfoService("gadio-redis", "begin store Included to redis")
	for _, includedEntity := range *gadio.Included {
		// two type `categories` and `users`
		// 1. gadio:categories:62
		// 2. gadio:users:124832
		key := strings.Join([]string{"gadio", includedEntity.Type, includedEntity.ID}, ":")
		logrus.Info("key:", key)

		if rc.Exists(key).Val() == 0 {
			if err := rc.HMSet(key, map[string]interface{}{
				"attributes:title":        includedEntity.Attributes.Title,
				"attributes:desc":         includedEntity.Attributes.Desc,
				"attributes:cover":        includedEntity.Attributes.Cover,
				"attributes:published-at": includedEntity.Attributes.PublishedAt,
				"attributes:nickname":     includedEntity.Attributes.Nickname,
				"attributes:thumb":        includedEntity.Attributes.Thumb,
				"attributes:name":         includedEntity.Attributes.Name,
				"attributes:logo":         includedEntity.Attributes.Logo,
				"attributes:background":   includedEntity.Attributes.Background,
			}).Err(); err != nil {
				Logger.ErrorService("gadio-redis-included", err)
			}
		}
	}

	Logger.InfoService("gadio-redis-data", "begin store data to redis")
	for _, dataEntity := range *gadio.Data {
		// gadio:radios:116484
		key := strings.Join([]string{"gadio", dataEntity.Type, dataEntity.ID}, ":")
		Logger.Info("key:", key)

		dur, _ := time.ParseDuration(strconv.Itoa(dataEntity.Attributes.Duration) + "s")

		if rc.Exists(key).Val() == 0 {
			categoryName := rc.HGet(
				strings.Join([]string{"gadio", "categories", dataEntity.Relationships.Category.Data.ID}, ":"),
				"attributes:name").Val()

			Logger.Info("categoryName: ", strings.Join([]string{"gadio", "categories", dataEntity.Relationships.Category.Data.ID}, ":"))

			if err := rc.HMSet(key, map[string]interface{}{
				"attributes:title":        dataEntity.Attributes.Title,
				"attributes:desc":         dataEntity.Attributes.Desc,
				"attributes:cover":        dataEntity.Attributes.Cover,
				"attributes:published-at": dataEntity.Attributes.PublishedAt,
				"attributes:duration":     dur.String(),

				"relationships:category:id":   dataEntity.Relationships.Category.Data.ID,
				"relationships:category:name": categoryName,
			}).Err(); err != nil {
				Logger.Error(err)
			}

			for _, djEntity := range *dataEntity.Relationships.Djs.Data {
				djName := rc.HGet(
					strings.Join([]string{"gadio", "users", djEntity.ID}, ":"),
					"attributes:nickname",
				).Val()

				// gadio:radios:116484:djs:nickname

				if err := rc.LPush(key+":djs:nickname", djName).Err(); err != nil {
					Logger.Error(err)
				}

				//gadio:radios:116484:djs:id
				if err := rc.LPush(key+":djs:id", djEntity.ID).Err(); err != nil {
					Logger.Error(err)
				}

			}

			if send == true {
				proactiveNotice(bot, sendToUser, dataEntity.Attributes.PublishedAt+"\n"+
					"https://www.gcores.com/radios/"+dataEntity.ID, nil)
				Logger.InfoService("gadio", "send msg to"+sendToUser)
			}
		}
	}
}
