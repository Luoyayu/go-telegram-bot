package dbRedis

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"log"
	"strings"
)

type User struct {
	name           string
	id             string
	permissionsStr string
	permissionsMap map[string]bool
}

func (r *User) PermissionsMap() map[string]bool {
	return r.permissionsMap
}

func (r *User) SetPermissionsMap(permissionsMap map[string]bool) {
	r.permissionsMap = permissionsMap
}

func (r *User) PermissionsStr() string {
	return r.permissionsStr
}

func (r *User) SetPermissionsStr(permissionsStr string) {
	r.permissionsStr = permissionsStr
}

func (r *User) Id() string {
	return r.id
}

func (r *User) SetId(id string) {
	r.id = id
}

func (r *User) Name() string {
	return r.name
}

func (r *User) SetName(name string) {
	r.name = name
}

type IUser interface {
	PermissionsStr() string
	SetPermissionsStr(string)
	PermissionsMap() map[string]bool
	SetPermissionsMap(map[string]bool)
	Id() string
	SetId(string)
	Name() string
	SetName(string)

	Add(*redis.Client) error
	Del(*redis.Client) int64
}

func NewUser(name, id string, permissionsStr string, permissionsMap map[string]bool) IUser {
	user := User{
		name:           name,
		id:             id,
		permissionsStr: permissionsStr,
		permissionsMap: permissionsMap,
	}
	return &user

}

func (r *User) Add(rc *redis.Client) error {
	if r.id != "" {
		rc.HSet(RedisKeys.UserIdName, r.id, r.name)
	} else {
		return fmt.Errorf("user id is void")
	}

	if r.permissionsStr != "" {
		rc.HSet(RedisKeys.UserIdPermissions, r.id, r.permissionsStr)
	} else {
		return fmt.Errorf("user permissionsStr is void")
	}
	return nil
}

func (r *User) Del(rc *redis.Client) int64 {
	return rc.HDel(RedisKeys.UserIdName, r.id).Val()
}

// FIXME User go 1.13 error

func (r *User) Get(rc *redis.Client) (err error) {
	if rc == nil {
		return fmt.Errorf("redis error")
	}

	if r.id == "" {
		return fmt.Errorf("user id is void")
	}

	if r.name, err = rc.HGet(RedisKeys.UserIdName, r.id).Result(); err != nil {
		return err
	}

	if r.name == "" {
		return fmt.Errorf("user not exists")
	}

	if r.permissionsStr = rc.HGet(RedisKeys.UserIdPermissions, r.id).Val(); r.permissionsStr != "" {
		permissionList := strings.Split(r.permissionsStr, ",")
		log.Println(permissionList)
		r.permissionsMap = make(map[string]bool)
		for _, v := range permissionList {
			r.permissionsMap[strings.TrimSpace(v)] = true
		}
	} else {
		return fmt.Errorf("user permissions string is void")
	}
	return nil
}

/*
func SetAllPermissions(rc *redis.Client, permissionsStr string) error {
	return rc.Set(RedisKeys.AllPermissions, permissionsStr, 0).Err()
}*/

func CheckUserExist(rc *redis.Client, userId string) bool {
	return rc.HExists(RedisKeys.UserIdName, userId).Val()
}
