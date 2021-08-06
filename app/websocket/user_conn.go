package websocket

import (
	"github.com/gorilla/websocket"
	"time"
	"ws/app/auth"
	"ws/app/chat"
	"ws/app/databases"
	"ws/app/models"
)

type UserConn struct {
	BaseConn
	User      auth.User
	CreatedAt int64
}

func (c *UserConn) GetUserId() int64 {
	return c.User.GetPrimaryKey()
}
func (c *UserConn) triggerMessageEvent(scene string, message *models.Message)  {
	rules := make([]*models.AutoRule, 0)
	databases.Db.
		Where("is_system", 0).
		Where("is_open", 1).
		Order("sort").
		Preload("Message").
		Preload("Scenes").
		Find(&rules)
LOOP:
	for _, rule := range rules {
		if rule.IsMatch(message.Content) && rule.SceneInclude(scene) {
			switch rule.ReplyType {
			// 转接人工客服
			case models.ReplyTypeTransfer:
				onlineServerCount := len(AdminHub.Clients)
				// 没有客服在线时
				if onlineServerCount == 0 {
					otherRule := models.AutoRule{}
					query := databases.Db.Where("is_system", 1).
						Where("match", models.MatchServiceAllOffLine).Preload("Message").First(&rule)
					if query.RowsAffected > 0 {
						msg := otherRule.GetReplyMessage(c.User.GetPrimaryKey())
						switch otherRule.ReplyType {
						case models.ReplyTypeTransfer:
							UserHub.addToManual(c.GetUserId())
						case models.ReplyTypeMessage:
							if msg != nil {
								databases.Db.Save(msg)
								otherRule.Count++
								databases.Db.Save(&otherRule)
								c.Deliver(NewReceiveAction(msg))
							}
						}
					} else {
						UserHub.addToManual(c.GetUserId())
					}
				} else {
					UserHub.addToManual(c.GetUserId())
				}
			// 回复消息
			case models.ReplyTypeMessage:
				msg := rule.GetReplyMessage(c.User.GetPrimaryKey())
				if msg != nil {
					databases.Db.Save(msg)
					c.Deliver(NewReceiveAction(msg))
				}
				//触发事件
			case models.ReplyTypeEvent:
				switch rule.Key {
				case "break":
					adminId := chat.GetUserLastAdminId(c.GetUserId())
					if adminId > 0 {
						_ = chat.RemoveUserAdminId(c.GetUserId(), adminId)
					}
					msg := rule.GetReplyMessage(c.User.GetPrimaryKey())
					if msg != nil {
						databases.Db.Save(msg)
						c.Deliver(NewReceiveAction(msg))
					}
				}
			}
			rule.Count++
			databases.Db.Save(rule)
			break LOOP
		}
	}
}

func (c *UserConn) onReceiveMessage(act *Action) {
	switch act.Action {
	case SendMessageAction:
		msg, err := act.GetMessage()
		if err == nil {
			if len(msg.Content) != 0 {
				msg.Source = models.SourceUser
				msg.UserId = c.GetUserId()
				msg.ReceivedAT = time.Now().Unix()
				msg.Avatar = c.User.GetAvatarUrl()
				msg.AdminId = chat.GetUserLastAdminId(c.GetUserId())
				c.Deliver(NewReceiptAction(msg))
				// 有对应的客服对象
				if msg.AdminId > 0 {
					// 更新会话有效期
					session := chat.GetSession(c.GetUserId(), msg.AdminId)
					if session == nil {
						return
					}
					addTime := chat.GetServiceSessionSecond()
					_ = chat.UpdateUserAdminId(msg.UserId, msg.AdminId, addTime)
					msg.SessionId = session.Id
					session.BrokeAt = time.Now().Unix() + addTime
					databases.Db.Save(session)
					databases.Db.Save(msg)
					adminConn, exist := AdminHub.GetConn(msg.AdminId)
					if exist {
						c.triggerMessageEvent(models.SceneAdminOnline, msg)
						adminConn.Deliver(NewReceiveAction(msg))
					} else {
						c.triggerMessageEvent(models.SceneAdminOffline, msg)
						adminSetting := &models.AdminChatSetting{}
						databases.Db.Where("admin_id = ?" , msg.AdminId).Find(adminSetting)
						if adminSetting.OfflineContent != "" {
							offlineMsg := adminSetting.GetOfflineMsg(c.GetUserId(), session.Id)
							c.Deliver(NewReceiveAction(offlineMsg))
						}
						// 客服不在线，判断是否超过了不在线自动断开的时间设置，超过了则自动断开会话
						if adminSetting.Id > 0 {
							lastOnline := adminSetting.LastOnline
							duration := chat.GetOfflineDuration()
							if (lastOnline.Unix() + duration) < time.Now().Unix() {
								_ = chat.RemoveUserAdminId(msg.UserId, msg.AdminId )
							}
						}
					}
				} else {
					databases.Db.Save(msg)
					if chat.IsInManual(c.GetUserId()) {
						AdminHub.BroadcastWaitingUser()
					} else {
						isAutoTransfer, exist := chat.Settings[chat.IsAutoTransfer]
						if exist  && isAutoTransfer.GetValue() == "1"{ // 自动转人工
							if !chat.IsInManual(c.GetUserId()) {
								UserHub.addToManual(c.GetUserId())
							}
						} else {
							if !chat.IsInManual(c.GetUserId()) {
								c.triggerMessageEvent(models.SceneNotAccepted, msg)
							}
						}
					}

				}
			}
		}
		break
	}
}
func (c *UserConn) Setup() {
	c.Register(onEnter, func(i ...interface{}) {
		if chat.GetUserLastAdminId(c.GetUserId()) == 0 {
			rule := models.AutoRule{}
			query := databases.Db.
				Where("is_system", 1).
				Where("match", models.MatchEnter).
				Preload("Message").
				First(&rule)
			if query.RowsAffected > 0 {
				if rule.Message != nil {
					msg := rule.GetReplyMessage(c.User.GetPrimaryKey())
					if msg != nil {
						databases.Db.Save(msg)
						rule.Count++
						databases.Db.Save(&rule)
						c.Deliver(NewReceiveAction(msg))
					}
				}
			}
		}
	})
	c.Register(onClose, func(i ...interface{}) {
		UserHub.Logout(c)
	})
	c.Register(onReceiveMessage, func(i ...interface{}) {
		length := len(i)
		if length >= 1 {
			ai := i[0]
			act, ok := ai.(*Action)
			if ok {
				c.onReceiveMessage(act)
			}
		}
	})
	c.Register(onSendSuccess, func(i ...interface{}) {
	})
}
func NewUserConn(user auth.User, conn *websocket.Conn) *UserConn {
	return &UserConn{
		User: user,
		BaseConn: BaseConn{
			conn:        conn,
			closeSignal: make(chan interface{}),
			send:        make(chan *Action, 100),
		},
	}
}
