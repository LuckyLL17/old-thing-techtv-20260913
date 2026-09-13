package dto

type TopicSaveReq struct {
	Title   string `json:"title" binding:"required"`
	Summary string `json:"summary"`
	Cover   string `json:"cover"`
	Status  *int   `json:"status"`
}

// TopicUpdateReq 部分更新：仅显式传入的字段才会被修改
type TopicUpdateReq struct {
	Title   *string `json:"title"`
	Summary *string `json:"summary"`
	Cover   *string `json:"cover"`
	Status  *int    `json:"status"`
}

type TopicAddItemReq struct {
	ItemType string `json:"item_type" binding:"required"`
	ItemID   uint64 `json:"item_id" binding:"required"`
}
