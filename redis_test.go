package main

import (
	dbRedis "github.com/luoyayu/go_telegram_bot/redis-tgbot"
	"testing"
)

func Test_connectToRedis(t *testing.T) {
	dbClient, _ = dbRedis.NewRedisClient()

	t.Logf("%q\n", dbClient.Get("ALI_ACCESS_TOKEN").Val())

}
