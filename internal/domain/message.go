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

// Conversation 会话列表项，由 messages 表聚合查询得到（非独立数据表）
type Conversation struct {
	OtherID      uint64    `gorm:"column:other_id" json:"other_id"`
	LastSenderID uint64    `gorm:"column:last_sender_id" json:"last_sender_id"`
	LastContent  string    `gorm:"column:last_content" json:"last_content"`
	LastAt       time.Time `gorm:"column:last_at" json:"last_at"`
	Unread       int64     `gorm:"column:unread" json:"unread"`
	User         *User     `gorm:"-" json:"user"`
}
