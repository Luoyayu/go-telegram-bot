package dbRedis

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"os"
)

func NewRedisClient() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("RedisAddress"),
		Password: os.Getenv("RedisPassword"),
		DB:       0,
	})

	if client.Ping().Val() != "PONG" {
		return nil, fmt.Errorf("connect to db error")
	} else {
		return client, nil
	}
}

func FlushDB(rc *redis.Client) {
	rc.FlushDB()
}

type HashKeysStruct struct {
	UserIdName        string
	UserNameId        string
	UserIdPermissions string
	AllPermissions    string
}

var HashKeys = HashKeysStruct{
	UserIdName:        "user:id:name",
	UserNameId:        "user:name:id",
	UserIdPermissions: "user:id:permissions",
	AllPermissions:    "allPermissions",
}
