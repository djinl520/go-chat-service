package admin

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"ws/app/databases"
	"ws/app/file"
	"ws/app/http/requests"
	"ws/app/models"
	"ws/app/repositories"
	"ws/app/util"
)

func StoreAutoMessageImage(c *gin.Context) {
	f, _ := c.FormFile("file")
	ff, err := file.Save(f, "auto_message")
	if err != nil {
		util.RespFail(c, err.Error(), 500)
	} else {
		util.RespSuccess(c, gin.H{
			"url": ff.FullUrl,
		})
	}
}
func GetAutoMessages(c *gin.Context)  {
	pagination := repositories.GetAutoMessagePagination(c)
	util.RespPagination(c , pagination)
}

func ShowAutoMessage(c *gin.Context) {
	id := c.Param("id")
	message := models.AutoMessage{}
	query := databases.Db.Find(&message, id)
	if query.RowsAffected > 0 {
		util.RespSuccess(c, message)
	} else {
		util.RespNotFound(c)
	}
}

func StoreAutoMessage(c *gin.Context)  {
	form := requests.AutoMessageForm{}
	err := c.ShouldBind(&form)
	if err != nil {
		util.RespValidateFail(c, err.Error())
		return
	}
	var  exist int64
	databases.Db.Table("auto_messages").
		Where("name = ?" , form.Name).Count(&exist)
	if exist > 0 {
		util.RespValidateFail(c, "已存在同名的消息")
		return
	}

	message := &models.AutoMessage{
		Name: form.Name,
		Type: form.Type,
	}

	if message.Type == models.TypeText  || message.Type == models.TypeImage {
		message.Content = form.Content
	}
	if message.Type == models.TypeNavigate {
		content := map[string]string{
			"title": form.Title,
			"url": form.Url,
			"content": form.Content,
		}
		jsonBytes, err := json.Marshal(content)
		if err != nil{
			util.RespError(c, err.Error())
			return
		}
		message.Content = string(jsonBytes)
	}
	databases.Db.Save(message)
	util.RespSuccess(c, message)
}
func UpdateAutoMessage(c *gin.Context) {
	message := &models.AutoMessage{}
	query := databases.Db.Find(&message, c.Param("id"))
	if query.Error == gorm.ErrRecordNotFound {
		util.RespNotFound(c)
	} else {
		form := requests.AutoMessageForm{}
		err := c.ShouldBind(&form)
		if err != nil {
			util.RespValidateFail(c, err.Error())
			return
		}
		var  exist int64
		databases.Db.Table("auto_messages").
			Where("name = ?" , form.Name).
			Where("id != ?", c.Param("id")).
			Count(&exist)
		if exist > 0 {
			util.RespValidateFail(c, "已存在同名的其他消息")
			return
		}
		if message.Type == models.TypeText  || message.Type == models.TypeImage {
			message.Content = form.Content
		}
		if message.Type == models.TypeNavigate {
			content := map[string]string{
				"title": form.Title,
				"url": form.Url,
				"content": form.Content,
			}
			jsonBytes, err := json.Marshal(content)
			if err != nil{
				util.RespError(c, err.Error())
				return
			}
			message.Content = string(jsonBytes)
		}
		databases.Db.Save(message)
		util.RespSuccess(c, message)
	}
}
func DeleteAutoMessage(c *gin.Context) {
	message := &models.AutoMessage{}
	databases.Db.Find(&message, c.Param("id"))
	if message.ID <= 0 {
		util.RespNotFound(c)
	} else {
		databases.Db.Delete(message)
		util.RespSuccess(c, message)
	}
}