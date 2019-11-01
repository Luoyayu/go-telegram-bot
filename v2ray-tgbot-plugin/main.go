package v2ray_tgbot

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	logger_tgbot "github.com/luoyayu/go_telegram_bot/logger-tgbot-plugin"
	"net/http"
)

func GetVmessCode(logger logger_tgbot.ILogger) (code string, err error) {
	var resp *http.Response
	var doc *goquery.Document

	defer func() {
		if v := recover(); v != nil {
			err = errors.New(fmt.Sprint(v))
			return
		}
	}()

	if resp, err = http.Get(targetUrl); err == nil {
		if doc, err = goquery.NewDocumentFromReader(resp.Body); err == nil {
			doc.Find("li .copybtn").Each(func(i int, selection *goquery.Selection) {
				code = selection.Nodes[0].Attr[1].Val
			})
		}
	}
	return
}
