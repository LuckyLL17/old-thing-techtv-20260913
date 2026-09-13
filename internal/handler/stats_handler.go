package handler

import (
	"strconv"
	"upcycle-hub/api/dto"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/middleware"
	"upcycle-hub/internal/service"
	apperr "upcycle-hub/pkg/errors"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	statsSvc    *service.StatsService
	interactSvc *service.InteractionService
}

func NewStatsHandler(s *service.StatsService, i *service.InteractionService) *StatsHandler {
	return &StatsHandler{statsSvc: s, interactSvc: i}
}

func (h *StatsHandler) Dashboard(c *gin.Context) {
	st, err := h.statsSvc.Dashboard()
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, st)
}

func (h *StatsHandler) Favorites(c *gin.Context) {
	uid := middleware.MustLogin(c)
	if uid == 0 {
		Fail(c, apperr.ErrUnauthorized)
		return
	}
	p := getPage(c)
	t := c.Query("type")
	list, total, err := h.interactSvc.ListFavorites(uid, t, p.Page, p.Size)
	if err != nil {
		Fail(c, err)
		return
	}
	PageOK(c, list, total, p.Page, p.Size)
}

func (h *StatsHandler) ToggleFavorite(c *gin.Context) {
	uid := middleware.MustLogin(c)
	if uid == 0 {
		Fail(c, apperr.ErrUnauthorized)
		return
	}
	var req dto.FavoriteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, apperr.Wrap(apperr.CodeValidation, "参数错误", err))
		return
	}
	added, err := h.interactSvc.ToggleFavorite(uid, req.TargetType, req.TargetID)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, gin.H{"favorited": added})
}

func (h *StatsHandler) Follow(c *gin.Context) {
	uid := middleware.MustLogin(c)
	if uid == 0 {
		Fail(c, apperr.ErrUnauthorized)
		return
	}
	var req dto.FollowReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, apperr.Wrap(apperr.CodeValidation, "参数错误", err))
		return
	}
	if err := h.interactSvc.Follow(uid, req.FollowingID); err != nil {
		Fail(c, err)
		return
	}
	OK(c, gin.H{"followed": true})
}

func (h *StatsHandler) Unfollow(c *gin.Context) {
	uid := middleware.MustLogin(c)
	if uid == 0 {
		Fail(c, apperr.ErrUnauthorized)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		Fail(c, apperr.ErrBadRequest)
		return
	}
	if err := h.interactSvc.Unfollow(uid, id); err != nil {
		Fail(c, err)
		return
	}
	OK(c, gin.H{"unfollowed": true})
}

func (h *StatsHandler) FollowInfo(c *gin.Context) {
	uid := middleware.MustLogin(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		Fail(c, apperr.ErrBadRequest)
		return
	}
	followers, following, err := h.interactSvc.FollowCounts(id)
	if err != nil {
		Fail(c, err)
		return
	}
	var isFollowing bool
	if uid > 0 {
		isFollowing, _ = h.interactSvc.IsFollowing(uid, id)
	}
	OK(c, gin.H{"followers": followers, "following": following, "is_following": isFollowing})
}

func (h *StatsHandler) SendMessage(c *gin.Context) {
	uid := middleware.MustLogin(c)
	if uid == 0 {
		Fail(c, apperr.ErrUnauthorized)
		return
	}
	var req dto.MessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, apperr.Wrap(apperr.CodeValidation, "参数错误", err))
		return
	}
	if err := h.interactSvc.SendMessage(uid, req.ReceiverID, req.Content); err != nil {
		Fail(c, err)
		return
	}
	OK(c, gin.H{"sent": true})
}

func (h *StatsHandler) Messages(c *gin.Context) {
	uid := middleware.MustLogin(c)
	if uid == 0 {
		Fail(c, apperr.ErrUnauthorized)
		return
	}
	other, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || other == 0 {
		Fail(c, apperr.ErrBadRequest)
		return
	}
	p := getPage(c)
	list, err := h.interactSvc.ListMessages(uid, other, p.Page, p.Size)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, list)
}

func (h *StatsHandler) Unread(c *gin.Context) {
	uid := middleware.MustLogin(c)
	if uid == 0 {
		Fail(c, apperr.ErrUnauthorized)
		return
	}
	n, err := h.interactSvc.UnreadCount(uid)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, gin.H{"unread": n})
}

func (h *StatsHandler) Attempts(c *gin.Context) {
	uid := middleware.MustLogin(c)
	if uid == 0 {
		Fail(c, apperr.ErrUnauthorized)
		return
	}
	p := getPage(c)
	list, total, err := h.interactSvc.ListAttempts(uid, p.Page, p.Size)
	if err != nil {
		Fail(c, err)
		return
	}
	PageOK(c, list, total, p.Page, p.Size)
}

func (h *StatsHandler) UserPublishedTutorials(c *gin.Context) {
	_ = domain.UserLevelNovice
}
