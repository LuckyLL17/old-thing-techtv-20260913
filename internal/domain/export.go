package domain

import "time"

// TutorialExport 是一次 A4 打印导出的快照记录。
// 导出文件在触发那一刻生成并落盘，之后教程再修改不影响已生成的文件。
type TutorialExport struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TutorialID uint64    `gorm:"index;not null" json:"tutorial_id"`
	Version    int       `gorm:"default:1" json:"version"` // 触发导出时教程的版本号
	Title      string    `gorm:"size:200" json:"title"`    // 触发导出时的教程标题
	FileName   string    `gorm:"size:255;not null" json:"file_name"`
	FileURL    string    `gorm:"size:500;not null" json:"file_url"` // 可访问的静态 URL
	FileSize   int64     `gorm:"default:0" json:"file_size"`
	CreatedBy  uint64    `gorm:"index" json:"created_by"` // 0 表示游客
	CreatedAt  time.Time `json:"created_at"`
}

func (TutorialExport) TableName() string {
	return "tutorial_exports"
}
