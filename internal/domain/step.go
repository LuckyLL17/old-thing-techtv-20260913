package domain

import "time"

type Step struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TutorialID  uint64    `gorm:"index;not null" json:"tutorial_id"`
	StepOrder   int       `gorm:"default:0;index" json:"step_order"`
	Title       string    `gorm:"size:200" json:"title"`
	Content     string    `gorm:"type:text;not null" json:"content"`
	Image       string    `gorm:"size:255" json:"image"`
	Reminder    string    `gorm:"size:500" json:"reminder"`
	EstimatedMinutes int  `gorm:"default:0" json:"estimated_minutes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *Step) TableName() string {
	return "steps"
}
