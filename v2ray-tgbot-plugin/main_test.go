package v2ray_tgbot

import (
	logger_tgbot "github.com/luoyayu/go_telegram_bot/logger-tgbot-plugin"
	"testing"
)

func TestGetVmessCode(t *testing.T) {
	GetVmessCode(logger_tgbot.NewLogger())
}