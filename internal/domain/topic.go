package domain

import "time"

const (
	TopicStatusOffline = 0 // 下架
	TopicStatusOnline  = 1 // 上线

	TopicItemTypeTutorial = "tutorial"
	TopicItemTypeProject  = "project"
)

type Topic struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Title     string    `gorm:"size:100;not null" json:"title"`
	Summary   string    `gorm:"size:500" json:"summary"`
	Cover     string    `gorm:"size:255" json:"cover"`
	Status    int       `gorm:"default:1;index" json:"status"`
	ItemCount int       `gorm:"default:0" json:"item_count"`
	CreatedBy uint64    `gorm:"index" json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (t *Topic) TableName() string {
	return "topics"
}

type TopicItem struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TopicID   uint64    `gorm:"index;uniqueIndex:idx_topic_item_unique;not null" json:"topic_id"`
	ItemType  string    `gorm:"size:20;uniqueIndex:idx_topic_item_unique;not null" json:"item_type"`
	ItemID    uint64    `gorm:"uniqueIndex:idx_topic_item_unique;not null" json:"item_id"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

func (ti *TopicItem) TableName() string {
	return "topic_items"
}
