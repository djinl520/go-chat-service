package repositories

import (
	"github.com/gin-gonic/gin"
	"ws/internal/databases"
	"ws/internal/models"
)


func UpdateMessages(wheres []*Where, values map[string]interface{}) int64 {
	query := databases.Db.Table("messages").Scopes(AddWhere(wheres))
	query.Updates(values)
	return query.RowsAffected
}

func GetMessages(wheres []*Where, limit int, load []string) []*models.Message  {
	messages := make([]*models.Message, 0)
	query := databases.Db.Order("received_at desc").
		Order("id desc").
		Limit(limit).
		Scopes(AddWhere(wheres))
	for _, relate := range load {
		query = query.Preload(relate)
	}
	query.Find(&messages)
	return messages
}

func GetUnSendMessage(wheres ...*Where) []*models.Message {
	wheres = append(wheres, &Where{
		Filed: "admin_id = ?",
		Value: 0,
	})
	return GetMessages(wheres, -1, []string{})
}

func GetAutoMessagePagination(c *gin.Context, wheres ...*Where) *Pagination {
	messages := make([]*models.AutoMessage, 0)
	databases.Db.Order("id desc").
		Scopes(Filter(c, []string{"type"})).
		Scopes(Paginate(c)).
		Scopes(AddWhere(wheres)).
		Find(&messages)
	var total int64
	databases.Db.Model(&models.AutoMessage{}).
		Scopes(Filter(c, []string{"type"})).
		Scopes(AddWhere(wheres)).
		Count(&total)
	return NewPagination(messages, total)
}


func GetAutoRulePagination(c *gin.Context, wheres ...*Where) *Pagination {
	rules := make([]*models.AutoRule, 0)
	wheres = append(wheres, &Where{
		Value: 0,
		Filed: "is_system = ?",
	})
	databases.Db.Order("id desc").
		Scopes(Filter(c, []string{"reply_type"})).
		Scopes(Paginate(c)).
		Scopes(AddWhere(wheres)).
		Preload("Message").
		Find(&rules)
	var total int64
	databases.Db.Model(&models.AutoRule{}).
		Scopes(Filter(c, []string{"reply_type"})).
		Scopes(AddWhere(wheres)).
		Count(&total)
	return NewPagination(rules, total)
}