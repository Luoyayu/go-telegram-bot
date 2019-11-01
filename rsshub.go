package main

import (
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
	AllRssSupportSubscribeInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup()

	// RssHubAllSubMap rssAllMap["youtube"]["user"]["path"]
	RssHubAllSubMap = map[string]map[string]map[string]string{
		"youtube": {
			"user": {
				"path": "youtube/user/",
				"help": ":username",
			},
			"channel": {
				"path": "youtube/channel/",
				"help": ":id",
			},
		},
		"bilibili": {
			"user": {
				"path": "bilibili/user/video/",
				"help": ":uid",
			},
			"live": {
				"path": "bilibili/live/room/",
				"help": ":roomID",
			},
		},
		"github": {
			"repos": {
				"path": "github/repos/",
				"help": ":user",
			},
		},
	}
)

func getAllRssSupportedSubscribe() {
	// http://127.0.0.1:1200/bilibili/user/video/388155334
	// http://127.0.0.1:1200/youtube/user/ryoya1983
	// http://127.0.0.1:1200/youtube/channel/UCkDbLiXbx6CIRZuyW9sZK1g

	AllRssSupportSubscribeInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup()
	for k, _ := range RssHubAllSubMap {
		Logger.InfofService("rssHub", k+" is up.")
		tg_tgbot.OneRowOneBtn(k, "rssSub_"+k, &AllRssSupportSubscribeInlineKeyboard, Logger)
	}
	tg_tgbot.OneRowOneBtn(">> close", BtnIdClose, &AllRssSupportSubscribeInlineKeyboard, Logger)

	Logger.InfoService("rssHub", RssHubAllSubMap["youtube"]["user"]["path"] == "youtube/user/")
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
