package handler

import (
	"strconv"
	"upcycle-hub/internal/middleware"
	"upcycle-hub/internal/service"
	apperr "upcycle-hub/pkg/errors"

	"github.com/gin-gonic/gin"
)

type ExportHandler struct {
	svc *service.ExportService
}

func NewExportHandler(s *service.ExportService) *ExportHandler {
	return &ExportHandler{svc: s}
}

// Create 生成教程的 A4 打印快照。游客可用；登录用户会记录在 CreatedBy 中。
func (h *ExportHandler) Create(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		Fail(c, apperr.ErrBadRequest)
		return
	}
	uid := middleware.MustLogin(c) // OptionalAuth 下未登录为 0
	rec, err := h.svc.Export(id, uid)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, rec)
}

// List 返回教程的历史导出快照列表（新→旧）。
func (h *ExportHandler) List(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		Fail(c, apperr.ErrBadRequest)
		return
	}
	list, err := h.svc.List(id)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, list)
}
