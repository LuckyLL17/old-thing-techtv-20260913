package handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"upcycle-hub/internal/middleware"
	"upcycle-hub/internal/repository"
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
	Operator   string `form:"operator"`
	Action     string `form:"action"`
	TargetType string `form:"target_type"`
	From       string `form:"from"`
	To         string `form:"to"`
}

type CreateAuditExportReq struct {
	UserID     uint64 `json:"user_id"`
	Operator   string `json:"operator"`
	Action     string `json:"action"`
	TargetType string `json:"target_type"`
	From       string `json:"from"`
	To         string `json:"to"`
}

type AuditHandler struct {
	svc       *service.AuditService
	exportSvc *service.AuditExportService
}

func NewAuditHandler(s *service.AuditService, es *service.AuditExportService) *AuditHandler {
	return &AuditHandler{svc: s, exportSvc: es}
}

func parseAuditFilter(userID uint64, operator, action, targetType, from, to string) (repository.AuditFilter, error) {
	filter := repository.AuditFilter{
		UserID:     userID,
		Operator:   strings.TrimSpace(operator),
		Action:     strings.TrimSpace(action),
		TargetType: strings.TrimSpace(targetType),
	}
	if filter.UserID == 0 {
		if id, err := strconv.ParseUint(filter.Operator, 10, 64); err == nil && id > 0 {
			filter.UserID = id
			filter.Operator = ""
		}
	}
	var err error
	if from != "" {
		if filter.From, err = parseAuditTime(from, false); err != nil {
			return filter, apperr.New(apperr.CodeValidation, "开始时间格式无效，支持 YYYY-MM-DD 或 YYYY-MM-DD HH:mm:ss")
		}
	}
	if to != "" {
		if filter.To, err = parseAuditTime(to, true); err != nil {
			return filter, apperr.New(apperr.CodeValidation, "结束时间格式无效，支持 YYYY-MM-DD 或 YYYY-MM-DD HH:mm:ss")
		}
	}
	if filter.From != nil && filter.To != nil && filter.From.After(*filter.To) {
		return filter, apperr.New(apperr.CodeValidation, "时间范围无效：开始时间不能晚于结束时间")
	}
	if filter.From != nil && filter.To != nil && filter.To.Sub(*filter.From) > 365*24*time.Hour {		return filter, apperr.New(apperr.CodeValidation, "时间范围过大：单次导出不能超过 365 天")
	}
	return filter, nil
}

func parseAuditTime(value string, endOfDay bool) (*time.Time, error) {
	value = strings.TrimSpace(value)
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02T15:04", "2006-01-02"}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			if endOfDay && layout == "2006-01-02" {
				t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			}
			return &t, nil
		}
	}
	return nil, fmt.Errorf("invalid audit time: %s", value)
}

func auditFail(c *gin.Context, err error, exportCategory string) {
	if ae, ok := err.(*apperr.AppError); ok {
		category := exportCategory
		if ae.Code == apperr.CodeForbidden || ae.Code == apperr.CodeUnauthorized {
			category = service.AuditExportErrorPermission
		}
		c.JSON(httpCode(ae.Code), gin.H{"code": ae.Code, "message": ae.Message, "error_category": category, "success": false})
		return
	}
	c.JSON(500, gin.H{"code": 50000, "message": err.Error(), "error_category": service.AuditExportErrorGeneration, "success": false})
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
	filter, err := parseAuditFilter(q.UserID, q.Operator, q.Action, q.TargetType, q.From, q.To)
	if err != nil {
		Fail(c, err)
		return
	}
	list, total, err := h.svc.List(q.Page, q.Size, filter)
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

func (h *AuditHandler) CreateExport(c *gin.Context) {
	uid := middleware.MustLogin(c)
	req := &CreateAuditExportReq{}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(req); err != nil {
			c.JSON(422, gin.H{"code": 42200, "message": "请求参数错误", "error_category": service.AuditExportErrorRange, "success": false})
			return
		}
	}
	filter, err := parseAuditFilter(req.UserID, req.Operator, req.Action, req.TargetType, req.From, req.To)
	if err != nil {
		c.JSON(422, gin.H{"code": 42200, "message": err.Error(), "error_category": service.AuditExportErrorRange, "success": false})
		return
	}
	job, err := h.exportSvc.Create(uid, filter)
	if err != nil {
		category := service.AuditExportErrorGeneration
		if ae, ok := err.(*apperr.AppError); ok && ae.Code == apperr.CodeValidation {
			category = service.AuditExportErrorRange
		}
		auditFail(c, err, category)
		return
	}
	c.JSON(202, gin.H{"code": 0, "message": "导出任务已提交，正在后台生成", "success": true, "data": job})
}

func (h *AuditHandler) ListExports(c *gin.Context) {
	uid := middleware.MustLogin(c)
	list, err := h.exportSvc.List(uid)
	if err != nil {
		auditFail(c, err, service.AuditExportErrorGeneration)
		return
	}
	OK(c, list)
}

func (h *AuditHandler) GetExport(c *gin.Context) {
	uid := middleware.MustLogin(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		Fail(c, apperr.ErrBadRequest)
		return
	}
	job, err := h.exportSvc.Get(uid, id, c.GetBool("is_admin"))
	if err != nil {
		auditFail(c, err, service.AuditExportErrorGeneration)
		return
	}
	OK(c, job)
}

func (h *AuditHandler) DownloadExport(c *gin.Context) {
	uid := middleware.MustLogin(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		Fail(c, apperr.ErrBadRequest)
		return
	}
	job, err := h.exportSvc.Get(uid, id, c.GetBool("is_admin"))
	if err != nil {
		auditFail(c, err, service.AuditExportErrorGeneration)
		return
	}
	path, fileName, err := h.exportSvc.FilePath(job)
	if err != nil {
		auditFail(c, err, service.AuditExportErrorGeneration)
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename*=UTF-8''"+fileName)
	c.FileAttachment(path, fileName)
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
