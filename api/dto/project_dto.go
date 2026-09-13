package dto

type ProjectCreateReq struct {
	TutorialID  uint64 `json:"tutorial_id" binding:"required"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Images      string `json:"images"`
	CustomNotes string `json:"custom_notes"`
	Rating      int    `json:"rating"`
}

type ProjectListReq struct {
	Page       int    `form:"page,default=1"`
	Size       int    `form:"size,default=10"`
	TutorialID uint64 `form:"tutorial_id"`
	UserID     uint64 `form:"user_id"`
	Sort       string `form:"sort"`
}

type SearchReq struct {
	Keyword    string `form:"keyword"`
	CategoryID uint64 `form:"category_id"`
	Difficulty string `form:"difficulty"`
	Page       int    `form:"page,default=1"`
	Size       int    `form:"size,default=10"`
}
