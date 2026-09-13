package handler

import (
	"strconv"
	"upcycle-hub/internal/middleware"
	"upcycle-hub/internal/service"
	apperr "upcycle-hub/pkg/errors"

	"github.com/gin-gonic/gin"
)

type BadgeHandler struct {
	svc *service.AchievementService
}

func NewBadgeHandler(s *service.AchievementService) *BadgeHandler {
	return &BadgeHandler{svc: s}
}

// Mine 个人中心：全部徽章定义 + 已获得/未获得状态 + 当前进度
func (h *BadgeHandler) Mine(c *gin.Context) {
	uid := middleware.MustLogin(c)
	if uid == 0 {
		Fail(c, apperr.ErrUnauthorized)
		return
	}
	badges, earnedCount, err := h.svc.MyBadges(uid)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, gin.H{
		"badges":       badges,
		"earned_count": earnedCount,
		"total_count":  len(badges),
	})
}

// Public 他人主页：只展示已获得的徽章
func (h *BadgeHandler) Public(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		Fail(c, apperr.ErrBadRequest)
		return
	}
	badges, err := h.svc.PublicBadges(id)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, gin.H{"badges": badges, "earned_count": len(badges)})
}

// Logs 管理端：追踪每次条件判定与发放结果
func (h *BadgeHandler) Logs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "30"))
	uid, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)
	list, total, err := h.svc.ListLogs(page, size, uid,
		c.Query("badge_code"), c.Query("result"), c.Query("event"))
	if err != nil {
		Fail(c, err)
		return
	}
	PageOK(c, list, total, page, size)
}
