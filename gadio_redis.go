package main

import (
	"github.com/go-redis/redis/v7"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	gadioRss "github.com/luoyayu/go_telegram_bot/gadio-rss"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

// update data and included to redis

func GadioStore(bot *tgbotapi.BotAPI, gadio *gadioRss.Radios, rc *redis.Client) {
	// redis Hash store
	Logger.Info("store Included")
	for _, includedEntity := range *gadio.Included {
		// two type `categories` and `users`
		// gadio:categories:62
		// gadio:users:124832
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
				Logger.Error(err)
			}
		}
	}

	Logger.Info("store data")
	for _, dataEntity := range *gadio.Data {
		// gadio:radios:116484
		key := strings.Join([]string{"gadio", dataEntity.Type, dataEntity.ID}, ":")
		Logger.Info("key:", key)
		Logger.Info("current key:", rc.Exists(key).Val())

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

			proactiveNotice(bot, "",
				dataEntity.Attributes.PublishedAt+"\n"+
					//dataEntity.Attributes.Title+"\n"+
					"https://www.gcores.com/radios/"+dataEntity.ID,
				nil)
			Logger.Info("send msg to superuser!")
		}
	}
}
