package domain

import "time"

type Tool struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TutorialID uint64    `gorm:"index;not null" json:"tutorial_id"`
	Name       string    `gorm:"size:100;not null" json:"name"`
	Quantity   string    `gorm:"size:50" json:"quantity"`
	Optional   bool      `gorm:"default:false" json:"optional"`
	SortOrder  int       `gorm:"default:0" json:"sort_order"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (Tool) TableName() string {
	return "tools"
}

type Notification struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64     `gorm:"index;not null" json:"user_id"`
	ActorID    uint64     `json:"actor_id"`
	Actor      *User      `gorm:"foreignKey:ActorID" json:"actor,omitempty"`
	Type       string     `gorm:"size:30;index;not null" json:"type"`
	Title      string     `gorm:"size:200;not null" json:"title"`
	Content    string     `gorm:"type:text" json:"content"`
	TutorialID uint64     `json:"tutorial_id,omitempty"`
	ProjectID  uint64     `json:"project_id,omitempty"`
	CommentID  uint64     `json:"comment_id,omitempty"`
	Read       bool       `gorm:"default:false;index" json:"read"`
	ReadAt     *time.Time `json:"read_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (Notification) TableName() string {
	return "notifications"
}

const (
	NotifComment     = "comment"
	NotifReply       = "reply"
	NotifFavorite    = "favorite"
	NotifFollow      = "follow"
	NotifAttempt     = "attempt"
	NotifProject     = "project"
	NotifLike        = "like"
	NotifSystem      = "system"
	NotifAuditPass   = "audit_pass"
	NotifAuditReject = "audit_reject"
)

type AuditLog struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64    `gorm:"index;not null" json:"user_id"`
	User       *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Action     string    `gorm:"size:50;index;not null" json:"action"`
	TargetType string    `gorm:"size:30;index" json:"target_type"`
	TargetID   uint64    `json:"target_id"`
	IP         string    `gorm:"size:64" json:"ip"`
	UserAgent  string    `gorm:"size:500" json:"user_agent"`
	Before     string    `gorm:"type:text" json:"before"`
	After      string    `gorm:"type:text" json:"after"`
	Remark     string    `gorm:"size:500" json:"remark"`
	CreatedAt  time.Time `json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

const (
	AuditLogin             = "login"
	AuditLogout            = "logout"
	AuditRegister          = "register"
	AuditPasswordReset     = "password_reset"
	AuditProfileUpdate     = "profile_update"
	AuditTutorialCreate    = "tutorial_create"
	AuditTutorialUpdate    = "tutorial_update"
	AuditTutorialDelete    = "tutorial_delete"
	AuditTutorialPublish   = "tutorial_publish"
	AuditTutorialArchive   = "tutorial_archive"
	AuditTutorialRollback  = "tutorial_rollback"
	AuditProjectCreate     = "project_create"
	AuditProjectDelete     = "project_delete"
	AuditCommentCreate     = "comment_create"
	AuditCommentDelete     = "comment_delete"
	AuditFavoriteToggle    = "favorite_toggle"
	AuditFollowToggle      = "follow_toggle"
	AuditAttemptToggle     = "attempt_toggle"
	AuditAdminAuditPass    = "admin_audit_pass"
	AuditAdminAuditReject  = "admin_audit_reject"
	AuditAdminUserBan      = "admin_user_ban"
	AuditAdminCategoryEdit = "admin_category_edit"
)
