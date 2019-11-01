package dbRedis

import (
	"testing"
)

func Test_connectToRedis(t *testing.T) {
	dbClient, _ := NewRedisClient()

	t.Logf("%q\n", dbClient.Get("ALI_ACCESS_TOKEN").Val())

}
