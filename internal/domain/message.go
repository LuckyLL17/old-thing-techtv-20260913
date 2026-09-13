package domain

import "time"

type Message struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	SenderID    uint64    `gorm:"index;not null" json:"sender_id"`
	ReceiverID  uint64    `gorm:"index;not null" json:"receiver_id"`
	Content     string    `gorm:"type:text;not null" json:"content"`
	IsRead      bool      `gorm:"default:false;index" json:"is_read"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (m *Message) TableName() string {
	return "messages"
}

// ConversationUser 会话列表中对方用户的展示字段，不含邮箱、账号状态等敏感信息
type ConversationUser struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// Conversation 会话列表项，由 messages 表聚合查询得到（非独立数据表）
type Conversation struct {
	OtherID      uint64            `gorm:"column:other_id" json:"other_id"`
	LastSenderID uint64            `gorm:"column:last_sender_id" json:"last_sender_id"`
	LastContent  string            `gorm:"column:last_content" json:"last_content"`
	LastAt       time.Time         `gorm:"column:last_at" json:"last_at"`
	Unread       int64             `gorm:"column:unread" json:"unread"`
	User         *ConversationUser `gorm:"-" json:"user"`
}
