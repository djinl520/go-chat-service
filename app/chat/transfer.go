package chat

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
	"ws/app/databases"
	"ws/app/models"
	"ws/app/repositories"
)


const (
	// 转接待接入的用户 sets
	transferUserKey = "user:transfer"
)

func CancelTransfer(transfer *models.ChatTransfer) error {
	chatSessionRepo.Delete([]*repositories.Where{
		{
			Filed: "admin_id = ?",
			Value: 0,
		},
		{
			Filed: "type = ? ",
			Value: models.ChatSessionTypeTransfer,
		},
		{
			Filed: "user_id = ?",
			Value: transfer.UserId,
		},
	})
	transfer.IsCanceled = true
	t := time.Now()
	transfer.CanceledAt = &t
	_ = transferRepo.Save(transfer)
	_ = RemoveTransfer(transfer.UserId)
	return nil
}


func Transfer(fromId int64, toId int64, uid int64, remark  string) error {
	session := GetSession(uid, fromId)
	if session == nil {
		return errors.New("invalid user")
	}
	now := time.Now()
	transfer := &models.ChatTransfer{
		UserId:      uid,
		SessionId:   session.Id,
		FromAdminId: fromId,
		ToAdminId:   toId,
		Remark:      remark,
		CreatedAt:   &now,
	}
	transferRepo.Save(transfer)
	CloseSession(session, true, true)
	_ = AddToTransfer(uid, toId)
	CreateSession(uid, models.ChatSessionTypeTransfer)
	return nil
}

// 从转接hash表移除用户
func RemoveTransfer(uid int64) error {
	ctx := context.Background()
	cmd := databases.Redis.HDel(ctx, transferUserKey, strconv.FormatInt(uid, 10))
	return cmd.Err()
}
// 获取user转接adminId
func GetUserTransferId(uid int64) int64 {
	ctx := context.Background()
	cmd := databases.Redis.HGet(ctx, transferUserKey, strconv.FormatInt(uid, 10))
	if cmd.Err() == redis.Nil {
		return 0
	}
	adminId, _ := strconv.ParseInt(cmd.Val(), 10, 64)
	return adminId
}
//  添加用户到转接哈希表中
func AddToTransfer(uid int64, adminId int64) error {
	ctx := context.Background()
	cmd := databases.Redis.HSet(ctx, transferUserKey, uid, adminId)
	return cmd.Err()
}
