package models

import "ws/app/databases"

const ChatSessionTypeNormal = 0
const ChatSessionTypeTransfer = 1

type ChatSession struct {
	Id         uint64 `gorm:"primaryKey" json:"id"`
	UserId     int64  `gorm:"index"`
	QueriedAt  int64
	AcceptedAt int64
	BrokeAt    int64
	AdminId    int64  `gorm:"index"`
	Admin      *Admin `gorm:"foreignKey:admin_id"`
	Type       int    `gorm:"default:0"`
	User       *User  `gorm:"foreignKey:user_id"`
}

func (chatSession *ChatSession) getTypeLabel() string {
	switch chatSession.Type {
	case ChatSessionTypeTransfer:
		return "转接"
	case ChatSessionTypeNormal:
		return "普通"
	default:
		return ""
	}
}
func (chatSession *ChatSession) ToJson() *ChatSessionJson {
	var userName, adminName string
	if chatSession.Admin == nil {
		admin := &Admin{}
		databases.Db.Model(chatSession).Association("Admin").Find(admin)
		chatSession.Admin = admin
	}
	adminName = chatSession.Admin.Username
	if chatSession.User == nil {
		user := &User{}
		databases.Db.Model(chatSession).Association("User").Find(user)
		chatSession.User = user
	}
	userName = chatSession.User.Username
	return &ChatSessionJson{
		Id:         chatSession.Id,
		UserId:     chatSession.UserId,
		QueriedAt:  chatSession.QueriedAt * 1000,
		AcceptedAt: chatSession.AcceptedAt * 1000,
		BrokeAt:    chatSession.BrokeAt * 1000,
		AdminId:    chatSession.AdminId,
		TypeLabel:  chatSession.getTypeLabel(),
		UserName:   userName,
		AdminName:  adminName,
	}
}

type ChatSessionJson struct {
	Id         uint64 `json:"id"`
	UserId     int64  `json:"-"`
	QueriedAt  int64  `json:"queried_at"`
	AcceptedAt int64  `json:"accepted_at"`
	BrokeAt    int64  `json:"broke_at"`
	AdminId    int64  `json:"Admin_id"`
	UserName   string `json:"user_name"`
	AdminName  string `json:"admin_name"`
	TypeLabel  string `json:"type_label"`
}
