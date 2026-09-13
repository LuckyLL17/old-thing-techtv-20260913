package handler

import (
	"strconv"
	"strings"
	"upcycle-hub/internal/middleware"
	"upcycle-hub/internal/service"
	apperr "upcycle-hub/pkg/errors"

	"github.com/gin-gonic/gin"
)

type ReviewHandler struct {
	svc *service.ReviewService
}

func NewReviewHandler(s *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{svc: s}
}

// Submit 作者提交教程进入审核队列。
func (h *ReviewHandler) Submit(c *gin.Context) {
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
	t, err := h.svc.Submit(id, uid)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, t)
}

// List 管理端按状态查看教程，默认待审队列。
func (h *ReviewHandler) List(c *gin.Context) {
	p := getPage(c)
	status := c.DefaultQuery("status", "pending")
	list, total, err := h.svc.ListByStatus(status, p.Page, p.Size)
	if err != nil {
		Fail(c, err)
		return
	}
	PageOK(c, list, total, p.Page, p.Size)
}

// Approve 管理端审核通过。
func (h *ReviewHandler) Approve(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		Fail(c, apperr.ErrBadRequest)
		return
	}
	t, err := h.svc.Approve(id, middleware.MustLogin(c), c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, t)
}

type RejectReq struct {
	Reason string `json:"reason" binding:"required"`
}

// Reject 管理端驳回，必须填写原因。
func (h *ReviewHandler) Reject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		Fail(c, apperr.ErrBadRequest)
		return
	}
	var req RejectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, apperr.Wrap(apperr.CodeValidation, "请填写驳回原因", err))
		return
	}
	req.Reason = strings.TrimSpace(req.Reason)
	t, err := h.svc.Reject(id, middleware.MustLogin(c), req.Reason, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, t)
}
