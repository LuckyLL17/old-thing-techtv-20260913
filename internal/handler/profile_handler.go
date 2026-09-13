package handler

import (
	"strconv"
	"upcycle-hub/internal/middleware"
	"upcycle-hub/internal/service"
	apperr "upcycle-hub/pkg/errors"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	profileSvc *service.ProfileService
}

func NewProfileHandler(p *service.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileSvc: p}
}

// Profile 用户主页：游客可访问，登录后额外返回关注状态
func (h *ProfileHandler) Profile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		Fail(c, apperr.ErrBadRequest)
		return
	}
	viewerID := middleware.MustLogin(c)
	data, err := h.profileSvc.GetProfile(viewerID, id)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, data)
}
