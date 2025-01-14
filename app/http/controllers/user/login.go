package user

import (
	"ws/app/http/responses"
	"ws/app/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type loginForm struct {
	Username string
	Password string
}

func Login(c *gin.Context) {
	form := &loginForm{}
	err := c.Bind(form)
	if err == nil {
		user := &models.User{}
		user.FindByName(form.Username)
		if user.ID != 0 {
			if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.Password)) == nil {
				responses.RespSuccess(c, gin.H{
					"token": user.Login(),
				})
				return
			}
		}
	}
	responses.RespFail(c, "账号密码错误", 500)
}
