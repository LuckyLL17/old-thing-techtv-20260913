package handler

import (
	"upcycle-hub/api/dto"
	"upcycle-hub/internal/middleware"
	"upcycle-hub/internal/service"
	apperr "upcycle-hub/pkg/errors"

	"github.com/gin-gonic/gin"
)

type FeedHandler struct {
	feedSvc *service.FeedService
}

func NewFeedHandler(s *service.FeedService) *FeedHandler {
	return &FeedHandler{feedSvc: s}
}

// Feed 关注动态：所关注作者最近发布的教程与改造作品按时间混排分页
func (h *FeedHandler) Feed(c *gin.Context) {
	uid := middleware.MustLogin(c)
	if uid == 0 {
		Fail(c, apperr.ErrUnauthorized)
		return
	}
	p := getPage(c)
	res, err := h.feedSvc.Feed(uid, p.Page, p.Size)
	if err != nil {
		Fail(c, err)
		return
	}
	list := make([]*dto.FeedItemResp, 0, len(res.Items))
	for _, it := range res.Items {
		list = append(list, &dto.FeedItemResp{
			Type:      it.Type,
			Tutorial:  it.Tutorial,
			Project:   it.Project,
			CreatedAt: it.CreatedAt,
		})
	}
	OK(c, dto.FeedResp{
		List:           list,
		Total:          res.Total,
		Page:           res.Page,
		Size:           res.Size,
		HasMore:        res.HasMore,
		FollowingCount: res.FollowingCount,
		Suggestions:    res.Suggestions,
	})
}
