package domain

import "time"

type Follow struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	FollowerID  uint64    `gorm:"index;not null" json:"follower_id"`
	FollowingID uint64    `gorm:"index;not null" json:"following_id"`
	CreatedAt   time.Time `json:"created_at"`
}

func (f *Follow) TableName() string {
	return "follows"
}
