package domain

import "time"

type Material struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TutorialID uint64    `gorm:"index;not null" json:"tutorial_id"`
	Name       string    `gorm:"size:100;not null" json:"name"`
	Quantity   string    `gorm:"size:50" json:"quantity"`
	Unit       string    `gorm:"size:20" json:"unit"`
	IsTool     bool      `gorm:"default:false" json:"is_tool"`
	Notes      string    `gorm:"size:255" json:"notes"`
	SortOrder  int       `gorm:"default:0" json:"sort_order"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (m *Material) TableName() string {
	return "materials"
}
