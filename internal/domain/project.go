package domain

import "time"

type Project struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint64    `gorm:"index;not null" json:"user_id"`
	User        *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	TutorialID  uint64    `gorm:"index;not null" json:"tutorial_id"`
	Tutorial    *Tutorial `gorm:"foreignKey:TutorialID" json:"tutorial,omitempty"`
	Title       string    `gorm:"size:200" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	Images      string    `gorm:"type:text" json:"images"`
	CustomNotes string    `gorm:"type:text" json:"custom_notes"`
	Rating      int       `gorm:"default:0;index" json:"rating"`
	LikeCount   int       `gorm:"default:0" json:"like_count"`
	CommentCount int      `gorm:"default:0" json:"comment_count"`
	Status      int       `gorm:"default:1" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (p *Project) TableName() string {
	return "projects"
}
