package domain

import "time"

const (
	FavTypeTutorial = "tutorial"
	FavTypeProject  = "project"
)

type Favorite struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64    `gorm:"index;not null" json:"user_id"`
	TargetType string    `gorm:"size:20;index;not null" json:"target_type"`
	TargetID   uint64    `gorm:"index;not null" json:"target_id"`
	CreatedAt  time.Time `json:"created_at"`
}

func (f *Favorite) TableName() string {
	return "favorites"
}
