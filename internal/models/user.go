package models

import (
	"github.com/gin-gonic/gin"
	"time"
	"ws/internal/databases"
	"ws/util"
)

type User struct {
	ID        int64      `json:"id"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	Username  string     `gorm:"string;size:255" json:"username"`
	Password  string     `gorm:"string;size:255" json:"-"`
	ApiToken  string     `gogm:"string;size:255"  json:"-"`
}

func (user *User) GetUsername() string  {
	return user.Username
}
func (user *User) GetAvatarUrl() string {
	return ""
}
func (user *User) GetPrimaryKey() int64 {
	return user.ID
}

func (user *User) Auth(c *gin.Context) bool {
	databases.Db.Where("api_token= ?", util.GetToken(c)).Limit(1).First(user)
	return user.ID > 0
}

func (user *User) Login() (token string) {
	token = util.RandomStr(32)
	databases.Db.Model(user).Update("api_token", token)
	return
}
func (user *User) FindByName(username string) {
	databases.Db.Where("username= ?", username).Limit(1).First(user)
}

