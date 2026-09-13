package dto

import "upcycle-hub/internal/domain"

// FeedItemResp 关注动态混排条目：教程与改造作品二选一
type FeedItemResp struct {
	Type      string      `json:"type"` // tutorial | project
	Tutorial  interface{} `json:"tutorial,omitempty"`
	Project   interface{} `json:"project,omitempty"`
	CreatedAt string      `json:"created_at"`
}

// FeedResp 关注动态响应
type FeedResp struct {
	List           []*FeedItemResp `json:"list"`
	Total          int64           `json:"total"`
	Page           int             `json:"page"`
	Size           int             `json:"size"`
	HasMore        bool            `json:"has_more"`
	FollowingCount int             `json:"following_count"`
	// Suggestions 仅在没有关注任何作者时返回，供前端展示发现作者引导
	Suggestions []*domain.User `json:"suggestions,omitempty"`
}
