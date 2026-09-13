package domain

import "time"

const (
	CommentTypeTutorial = "tutorial"
	CommentTypeProject  = "project"
)

type Comment struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64    `gorm:"index;not null" json:"user_id"`
	User       *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	TargetType string    `gorm:"size:20;index;not null" json:"target_type"`
	TargetID   uint64    `gorm:"index;not null" json:"target_id"`
	ParentID   uint64    `gorm:"index;default:0" json:"parent_id"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	LikeCount  int       `gorm:"default:0" json:"like_count"`
	Status     int       `gorm:"default:1" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (c *Comment) TableName() string {
	return "comments"
}
