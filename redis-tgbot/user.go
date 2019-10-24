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
		rc.HSet(HashKeys.UserIdName, r.id, r.name)
	} else {
		return fmt.Errorf("user id is void")
	}

	if r.permissionsStr != "" {
		rc.HSet(HashKeys.UserIdPermissions, r.id, r.permissionsStr)
	} else {
		return fmt.Errorf("user permissionsStr is void")
	}
	return nil
}

func (r *User) Del(rc *redis.Client) int64 {
	return rc.HDel(HashKeys.UserIdName, r.id).Val()
}

func (r *User) Get(rc *redis.Client) error {
	if rc == nil {
		return fmt.Errorf("redis connect down")
	}

	if r.id == "" {
		return fmt.Errorf("user id is void")
	} else {
		var err error
		r.name, err = rc.HGet(HashKeys.UserIdName, r.id).Result()
		if err != nil {
			return err
		}
		if r.name == "" {
			return fmt.Errorf("user id is not exists")
		} else {
			r.permissionsStr = rc.HGet(HashKeys.UserIdPermissions, r.id).Val()
			if r.permissionsStr != "" {
				permissionList := strings.Split(r.permissionsStr, ",")
				log.Println(permissionList)
				var tmpMap = make(map[string]bool)
				for _, v := range permissionList {
					tmpMap[strings.TrimSpace(v)] = true
				}
				r.permissionsMap = tmpMap
			} else {
				return fmt.Errorf("user permissions Str is void")
			}
		}
	}
	return nil
}

/*func GetAllPermissions(rc *redis.Client) (map[string]bool, error) {
	permissionsMap := map[string]bool{}
	allPermissions, err := rc.HKeys(HashKeys.AllPermissions).Result()
	if err != nil {
		return nil, err
	}

	for _, p := range allPermissions {
		permissionsMap[p] = true
	}
	return permissionsMap, nil
}*/

func SetAllPermissions(rc *redis.Client, permissions string) error {
	return rc.Set(HashKeys.AllPermissions, permissions, 0).Err()
}

func CheckUserExist(rc *redis.Client, userId string) bool {
	return rc.HExists(HashKeys.UserIdName, userId).Val()
}
