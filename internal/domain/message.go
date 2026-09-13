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
