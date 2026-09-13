package domain

import "time"

const (
	AuditExportQueued    = "queued"
	AuditExportRunning   = "running"
	AuditExportSucceeded = "succeeded"
	AuditExportFailed    = "failed"
)

type AuditExport struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	RequesterID    uint64     `gorm:"index;not null" json:"requester_id"`
	Status         string     `gorm:"size:20;index;not null;default:queued" json:"status"`
	UserID         uint64     `gorm:"index" json:"user_id"`
	Operator       string     `gorm:"size:100;index" json:"operator"`
	Action         string     `gorm:"size:50;index" json:"action"`
	TargetType     string     `gorm:"size:30;index" json:"target_type"`
	From           *time.Time `gorm:"column:from_time" json:"from,omitempty"`
	To             *time.Time `gorm:"column:to_time" json:"to,omitempty"`
	FilePath       string     `gorm:"size:500" json:"-"`
	FileName       string     `gorm:"size:200" json:"file_name"`
	RowCount       int64      `json:"row_count"`
	ErrorCategory  string     `gorm:"size:20" json:"error_category,omitempty"`
	ErrorMessage   string     `gorm:"size:500" json:"error_message,omitempty"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
	ExpiresAt      *time.Time `gorm:"index" json:"expires_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (AuditExport) TableName() string {
	return "audit_exports"
}
