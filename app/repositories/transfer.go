package repositories

import (
	"github.com/gin-gonic/gin"
	"ws/app/databases"
	"ws/app/models"
)

type TransferRepo struct {
}

func (repo *TransferRepo) Get(wheres []*Where, limit int, load []string, orders ...string) []*models.ChatTransfer {
	messages := make([]*models.ChatTransfer, 0)
	databases.Db.
		Limit(limit).
		Scopes(AddOrder(orders)).
		Scopes(AddWhere(wheres)).
		Scopes(AddLoad(load)).
		Find(&messages)
	return messages
}
func (repo *TransferRepo) Save(transfer *models.ChatTransfer) int64 {
	result := databases.Db.Save(transfer)
	return result.RowsAffected
}

func (repo *TransferRepo) First(where []*Where, orders ...string) *models.ChatTransfer  {
	model := &models.ChatTransfer{}
	result := databases.Db.Scopes(AddWhere(where)).Scopes(AddOrder(orders)).First(model)
	if result.Error != nil {
		return nil
	}
	return model
}

func (repo *TransferRepo) Paginate(c *gin.Context, wheres []*Where, load []string, order ...string) *Pagination {
	transfer := make([]*models.ChatTransfer, 0)
	databases.Db.Scopes(AddWhere(wheres)).
		Scopes(AddOrder(load)).
		Scopes(AddLoad(order)).
		Scopes(Paginate(c)).
		Find(&transfer)
	var total int64
	databases.Db.Model(&models.ChatTransfer{}).
		Scopes(AddWhere(wheres)).
		Scopes(AddOrder(order)).
		Count(&total)
	return NewPagination(transfer, total)
}