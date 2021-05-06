package hub

import (
	"github.com/gorilla/websocket"
	"time"
	"ws/action"
	"ws/db"
	"ws/models"
)

type ServiceConn struct {
	User *models.ServiceUser
	BaseConn
}

func (c *ServiceConn) Setup() {
	c.Register(onReceiveMessage, func(i ...interface{}) {
		length := len(i)
		if length >= 1 {
			ai := i[0]
			act, ok := ai.(*action.Action)
			if ok {
				switch act.Action {
				case action.SendMessageAction:
					msg, err := act.GetMessage()
					if err == nil {
						if msg.UserId > 0 && len(msg.Content) != 0 && c.User.CheckChatUserLegal(msg.UserId) {
							msg.ServiceId = c.User.ID
							msg.IsServer = true
							msg.ReceivedAT = time.Now().Unix()
							db.Db.Save(msg)
							_ = c.User.UpdateChatUser(msg.UserId)
							c.Deliver(action.NewReceiptAction(msg))
							userConn, ok := UserHub.GetConn(msg.UserId)
							if ok { // 在线
								userConn.Deliver(action.NewReceiveAction(msg))
							}
						}
					}
					break
				}
			}
		}
	})
}
func (c *ServiceConn) GetUserId() int64 {
	return c.User.ID
}

func NewServiceConn(user *models.ServiceUser, conn *websocket.Conn) *ServiceConn {
	return &ServiceConn{
		User: user,
		BaseConn: BaseConn{
			conn:        conn,
			closeSignal: make(chan interface{}),
			send:        make(chan *action.Action, 100),
		},
	}
}