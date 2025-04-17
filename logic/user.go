package logic

import (
	"Go-Chat/common/e"
	"time"
)

type User struct {
	Id         int       `gorm:"column:id;primaryKey;comment:自增id"`
	Uuid       string    `gorm:"column:uuid;uniqueIndex;type:char(20);comment:用户唯一id"`
	UserName   string    `gorm:"column:username;type:varchar(20);not null;comment:昵称"`
	Password   string    `gorm:"column:password;type:char(18);not null;comment:密码"`
	CreateTime time.Time `gorm:"column:created_at;index;type:datetime;not null;comment:创建时间"`
	//LastOnlineAt  sql.NullTime `gorm:"column:last_online_at;type:datetime;comment:上次登录时间"`
	//LastOfflineAt sql.NullTime `gorm:"column:last_offline_at;type:datetime;comment:最近离线时间"`
}

func (u *User) TableName() string {
	return "user"
}

func (u *User) Add() (userId int, err error) {
	if u.UserName == "" || u.Password == "" {
		return 0, e.Error_Register_NAMEORPWD
	}
	oUser := u.CheckHaveUserName(u.UserName)
	if oUser.Id > 0 {
		return oUser.Id, nil
	}
	u.CreateTime = time.Now()
	if err = GormDB.Table(u.TableName()).Create(&u).Error; err != nil {
		return 0, err
	}
	return u.Id, nil
}

func (u *User) CheckHaveUserName(userName string) (data User) {
	GormDB.Table(u.TableName()).Where("user_name=?", userName).Take(&data)
	return
}

func (u *User) GetUserNameByUserId(userId int) (userName string) {
	var data User
	GormDB.Table(u.TableName()).Where("id=?", userId).Take(&data)
	return data.UserName
}

func SetUserOnline(userUuid string, ttl time.Duration) error {
	key := "online:user:" + userUuid
	return RedisClient.Set(key, "1", ttl).Err()
}

func IsUserOnline(userUuid string) (bool, error) {
	key := "online:user:" + userUuid
	v, err := RedisClient.Exists(key).Result()
	if err != nil {
		return false, err
	}
	return v != 0, nil
}

func GetOnlineUserCount() (int, error) {
	var count int
	var cursor uint64
	var keys []string
	var err error

	for {
		keys, cursor, err = RedisClient.Scan(cursor, "online:user:*", 100).Result()
		if err != nil {
			return count, err
		}
		count += len(keys)
		if cursor == 0 {
			break
		}
	}
	return count, nil
}
