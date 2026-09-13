package domain

import "time"

type Tag struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string    `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Slug         string    `gorm:"size:50;uniqueIndex" json:"slug"`
	TutorialCount int      `gorm:"default:0;index" json:"tutorial_count"`
	Color        string    `gorm:"size:20" json:"color"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (t *Tag) TableName() string {
	return "tags"
}

type TutorialTag struct {
	TutorialID uint64    `gorm:"primaryKey" json:"tutorial_id"`
	TagID      uint64    `gorm:"primaryKey" json:"tag_id"`
	CreatedAt  time.Time `json:"created_at"`
}

func (TutorialTag) TableName() string {
	return "tutorial_tags"
}
