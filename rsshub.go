package main

import (
	"github.com/go-redis/redis/v7"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	tg_tgbot "github.com/luoyayu/go_telegram_bot/tg-tgbot-plugin"
	"net/http"
	"os"
)

/*type rssSupportedServiceStruct struct {
	Data []*rssServices `json:"data"`
}

type rssServices struct {
	RssName  string               `json:"rss_name"`
	Services []*rssServicesEntity `json:"services"`
}

type rssServicesEntity struct {
	SubName string `json:"sub_name"`
	Path    string `json:"path"`
	Help    string `json:"help"`
}*/

var (
	AllRssSupportSubscribedDomainInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup()

	// RssHubAllSubMap rssAllMap["youtube"]["user"]["path"]
	RssHubAllSubMap = map[string]map[string]map[string]string{
		"youtube": {
			"user": {
				"path": "youtube/user/",
				"help": "username*",
				"type": "string",
			},
			"channel": {
				"path": "youtube/channel/",
				"help": "id*",
				"type": "string",
			},
		},
		"bilibili": {
			"user": {
				"path": "bilibili/user/video/",
				"help": "uid*",
				"type": "int",
			},
			"live": {
				"path": "bilibili/live/room/",
				"help": "roomID*",
				"type": "int",
			},
		},
		"github": {
			"repos": {
				"path": "github/repos/",
				"help": "userName*",
				"type": "string",
			},
		},
	}
)

func getAllRssSupportedSubscribe() {
	// http://127.0.0.1:1200/bilibili/user/video/388155334
	// http://127.0.0.1:1200/youtube/user/ryoya1983
	// http://127.0.0.1:1200/youtube/channel/UCkDbLiXbx6CIRZuyW9sZK1g

	AllRssSupportSubscribedDomainInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup()
	for k, _ := range RssHubAllSubMap {
		Logger.InfofService("rssHub", k+" is up.")
		tg_tgbot.OneRowOneBtn(k, "rssSub_"+k, &AllRssSupportSubscribedDomainInlineKeyboard, Logger)
	}

	tg_tgbot.OneRowOneBtn("my rss list", BtnIdShowUserAllRss, &AllRssSupportSubscribedDomainInlineKeyboard, Logger)
	tg_tgbot.OneRowOneBtn(">> close", BtnIdClose, &AllRssSupportSubscribedDomainInlineKeyboard, Logger)

	Logger.InfoService("rssHub", "test RssHubAllSubMap string index", RssHubAllSubMap["youtube"]["user"]["path"] == "youtube/user/")
}

func checkRssHubAvailable() (err error) {
	RssHubAvailableMutex.Lock()
	if _, err = http.Get(os.Getenv("RSSHub_Url")); err != nil {
		RssHubAvailable = false
		Logger.ErrorService("rssHub", "rssHub is down")
	} else {
		RssHubAvailable = true
		Logger.InfoService("rssHub", "rssHub is on")
	}
	RssHubAvailableMutex.Unlock()
	return
}

func GetUserAllRss(userId string, dbClient *redis.Client) []string {
	return dbClient.SMembers("user:" + userId + ":rssTasks").Val()
}

func GetUserAllRssInlineKeyboard(userRss []string) *tgbotapi.InlineKeyboardMarkup {
	userAllRssInlineKeyboard := tgbotapi.NewInlineKeyboardMarkup()
	for _, subEntity := range userRss {
		tg_tgbot.OneRowOneBtn(subEntity, "user_rss_"+subEntity, &userAllRssInlineKeyboard, Logger)
	}
	tg_tgbot.OneRowOneBtn("<< back", BtnIdBackToAllRssSupportingSubscribes, &userAllRssInlineKeyboard, Logger)
	return &userAllRssInlineKeyboard
}
