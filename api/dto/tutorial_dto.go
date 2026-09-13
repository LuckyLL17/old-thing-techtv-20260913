package dto

type TutorialListReq struct {
	Page     int    `form:"page,default=1"`
	Size     int    `form:"size,default=10"`
	Category uint64 `form:"category"`
	Difficulty string `form:"difficulty"`
	Status   string `form:"status"`
	Sort     string `form:"sort"`
	Keyword  string `form:"keyword"`
	UserID   uint64 `form:"user_id"`
}

type StepIn struct {
	Title            string `json:"title"`
	Content          string `json:"content"`
	Image            string `json:"image"`
	Reminder         string `json:"reminder"`
	EstimatedMinutes int    `json:"estimated_minutes"`
}

type MaterialIn struct {
	Name     string `json:"name"`
	Quantity string `json:"quantity"`
	Unit     string `json:"unit"`
	Notes    string `json:"notes"`
}

type TutorialCreateReq struct {
	CategoryID     uint64       `json:"category_id" binding:"required"`
	Title          string       `json:"title" binding:"required"`
	Summary        string       `json:"summary"`
	CoverBefore    string       `json:"cover_before" binding:"required"`
	CoverAfter     string       `json:"cover_after" binding:"required"`
	Difficulty     string       `json:"difficulty"`
	EstimatedHours float64      `json:"estimated_hours"`
	Status         string       `json:"status"`
	Tags           []string     `json:"tags"`
	Steps          []*StepIn    `json:"steps"`
	Materials      []*MaterialIn `json:"materials"`
	Tools          []*MaterialIn `json:"tools"`
}

type ReorderStepsReq struct {
	Order []uint64 `json:"order" binding:"required"`
}

type CommentReq struct {
	Content  string `json:"content" binding:"required"`
	ParentID uint64 `json:"parent_id"`
}

type FavoriteReq struct {
	TargetType string `json:"target_type" binding:"required"`
	TargetID   uint64 `json:"target_id" binding:"required"`
}

type FollowReq struct {
	FollowingID uint64 `json:"following_id" binding:"required"`
}

type MessageReq struct {
	ReceiverID uint64 `json:"receiver_id" binding:"required"`
	Content    string `json:"content" binding:"required"`
}

type AttemptReq struct {
	TutorialID uint64 `json:"tutorial_id" binding:"required"`
	Completed  bool   `json:"completed"`
	Note       string `json:"note"`
}
