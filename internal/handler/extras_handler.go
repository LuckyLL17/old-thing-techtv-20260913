package handler

import (
	"strconv"
	"time"
	"upcycle-hub/internal/middleware"
	"upcycle-hub/internal/service"
	apperr "upcycle-hub/pkg/errors"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	svc *service.NotificationService
}

func NewNotificationHandler(s *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: s}
}

type NotifQuery struct {
	Page       int  `form:"page"`
	Size       int  `form:"size"`
	OnlyUnread bool `form:"only_unread"`
}

func (h *NotificationHandler) List(c *gin.Context) {
	uid := middleware.MustLogin(c)
	if uid == 0 {
		Fail(c, apperr.ErrUnauthorized)
		return
	}
	q := &NotifQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		q = &NotifQuery{Page: 1, Size: 20}
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Size <= 0 || q.Size > 100 {
		q.Size = 20
	}
	list, total, err := h.svc.List(uid, q.Page, q.Size, q.OnlyUnread)
	if err != nil {
		Fail(c, err)
		return
	}
	PageOK(c, list, total, q.Page, q.Size)
}

func (h *NotificationHandler) CountUnread(c *gin.Context) {
	uid := middleware.MustLogin(c)
	if uid == 0 {
		Fail(c, apperr.ErrUnauthorized)
		return
	}
	n, err := h.svc.CountUnread(uid)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, gin.H{"unread": n})
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
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
	if err := h.svc.MarkRead(uid, id); err != nil {
		Fail(c, err)
		return
	}
	OK(c, nil)
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	uid := middleware.MustLogin(c)
	if uid == 0 {
		Fail(c, apperr.ErrUnauthorized)
		return
	}
	n, err := h.svc.MarkAllRead(uid)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, gin.H{"updated": n})
}

func (h *NotificationHandler) Clear(c *gin.Context) {
	uid := middleware.MustLogin(c)
	if uid == 0 {
		Fail(c, apperr.ErrUnauthorized)
		return
	}
	days := 0
	if s := c.Query("older_than_days"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v >= 0 {
			days = v
		}
	}
	n, err := h.svc.ClearOld(uid, days)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, gin.H{"deleted": n})
}

type AdminAuditQuery struct {
	Page       int    `form:"page"`
	Size       int    `form:"size"`
	UserID     uint64 `form:"user_id"`
	Action     string `form:"action"`
	TargetType string `form:"target_type"`
	From       string `form:"from"`
	To         string `form:"to"`
}

type AuditHandler struct {
	svc *service.AuditService
}

func NewAuditHandler(s *service.AuditService) *AuditHandler {
	return &AuditHandler{svc: s}
}

func (h *AuditHandler) List(c *gin.Context) {
	q := &AdminAuditQuery{}
	c.ShouldBindQuery(q)
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Size <= 0 || q.Size > 200 {
		q.Size = 30
	}
	var from, to *time.Time
	if q.From != "" {
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", q.From, time.Local); err == nil {
			from = &t
		} else if t, err := time.ParseInLocation("2006-01-02", q.From, time.Local); err == nil {
			from = &t
		}
	}
	if q.To != "" {
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", q.To, time.Local); err == nil {
			to = &t
		} else if t, err := time.ParseInLocation("2006-01-02", q.To, time.Local); err == nil {
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			to = &t
		}
	}
	list, total, err := h.svc.List(q.Page, q.Size, q.UserID, q.Action, q.TargetType, from, to)
	if err != nil {
		Fail(c, err)
		return
	}
	PageOK(c, list, total, q.Page, q.Size)
}

func (h *AuditHandler) Stats(c *gin.Context) {
	days := 30
	if s := c.Query("days"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v >= 0 {
			days = v
		}
	}
	data, err := h.svc.Stats(days)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, data)
}

type HistoryQuery struct {
	Page int `form:"page"`
	Size int `form:"size"`
}

type TutorialHistoryHandler struct {
	svc *service.TutorialHistoryService
}

func NewTutorialHistoryHandler(s *service.TutorialHistoryService) *TutorialHistoryHandler {
	return &TutorialHistoryHandler{svc: s}
}

func (h *TutorialHistoryHandler) List(c *gin.Context) {
	tid, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || tid == 0 {
		Fail(c, apperr.ErrBadRequest)
		return
	}
	q := &HistoryQuery{}
	c.ShouldBindQuery(q)
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Size <= 0 || q.Size > 100 {
		q.Size = 20
	}
	list, total, err := h.svc.List(tid, q.Page, q.Size)
	if err != nil {
		Fail(c, err)
		return
	}
	PageOK(c, list, total, q.Page, q.Size)
}

func (h *TutorialHistoryHandler) Get(c *gin.Context) {
	tid, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || tid == 0 {
		Fail(c, apperr.ErrBadRequest)
		return
	}
	version, err := strconv.Atoi(c.Param("version"))
	if err != nil {
		Fail(c, apperr.ErrBadRequest)
		return
	}
	v, err := h.svc.Get(tid, version)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, v)
}

type RollbackReq struct {
	Version int `json:"version" binding:"required,min=1"`
}

func (h *TutorialHistoryHandler) Rollback(c *gin.Context) {
	uid := middleware.MustLogin(c)
	if uid == 0 {
		Fail(c, apperr.ErrUnauthorized)
		return
	}
	tid, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || tid == 0 {
		Fail(c, apperr.ErrBadRequest)
		return
	}
	req := &RollbackReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		Fail(c, apperr.ErrBadRequest)
		return
	}
	if err := h.svc.Rollback(tid, uid, req.Version); err != nil {
		Fail(c, err)
		return
	}
	OK(c, gin.H{"tutorial_id": tid, "rolled_back_to": req.Version})
}

func (h *TutorialHistoryHandler) Snapshot(c *gin.Context) {
	uid := middleware.MustLogin(c)
	if uid == 0 {
		Fail(c, apperr.ErrUnauthorized)
		return
	}
	tid, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || tid == 0 {
		Fail(c, apperr.ErrBadRequest)
		return
	}
	if err := h.svc.Snapshot(tid); err != nil {
		Fail(c, err)
		return
	}
	OK(c, gin.H{"saved": true})
}
