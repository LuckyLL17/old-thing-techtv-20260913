package domain

import "time"

type Attempt struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64    `gorm:"index;not null" json:"user_id"`
	TutorialID uint64    `gorm:"index;not null" json:"tutorial_id"`
	Status     int       `gorm:"default:1" json:"status"`
	Completed  bool      `gorm:"default:false" json:"completed"`
	Note       string    `gorm:"type:text" json:"note"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (a *Attempt) TableName() string {
	return "attempts"
}
