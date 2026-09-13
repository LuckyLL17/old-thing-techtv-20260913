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

type TopicHandler struct {
	topicSvc *service.TopicService
}

func NewTopicHandler(t *service.TopicService) *TopicHandler {
	return &TopicHandler{topicSvc: t}
}

func parseTopicID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		Fail(c, apperr.ErrBadRequest)
		return 0, false
	}
	return id, true
}

// List 前台专题列表（仅上线）
func (h *TopicHandler) List(c *gin.Context) {
	p := getPage(c)
	list, total, err := h.topicSvc.List(p.Page, p.Size)
	if err != nil {
		Fail(c, err)
		return
	}
	PageOK(c, list, total, p.Page, p.Size)
}

// Detail 前台专题详情 + 按添加顺序分页的条目
func (h *TopicHandler) Detail(c *gin.Context) {
	id, ok := parseTopicID(c)
	if !ok {
		return
	}
	t, err := h.topicSvc.Get(id)
	if err != nil {
		Fail(c, err)
		return
	}
	if t.Status != domain.TopicStatusOnline {
		Fail(c, apperr.ErrNotFound)
		return
	}
	p := getPage(c)
	items, total, err := h.topicSvc.Items(id, p.Page, p.Size)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, gin.H{
		"topic": t,
		"items": dto.PageResp{List: items, Total: total, Page: p.Page, Size: p.Size},
	})
}

// AdminList 管理端专题列表（含已下架）
func (h *TopicHandler) AdminList(c *gin.Context) {
	p := getPage(c)
	list, total, err := h.topicSvc.AdminList(p.Page, p.Size)
	if err != nil {
		Fail(c, err)
		return
	}
	PageOK(c, list, total, p.Page, p.Size)
}

func (h *TopicHandler) Create(c *gin.Context) {
	uid := middleware.MustLogin(c)
	if uid == 0 {
		Fail(c, apperr.ErrUnauthorized)
		return
	}
	var req dto.TopicSaveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, apperr.Wrap(apperr.CodeValidation, "参数错误", err))
		return
	}
	t, err := h.topicSvc.Create(uid, &service.TopicSaveReq{
		Title: req.Title, Summary: req.Summary, Cover: req.Cover, Status: req.Status,
	})
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, t)
}

func (h *TopicHandler) Update(c *gin.Context) {
	id, ok := parseTopicID(c)
	if !ok {
		return
	}
	var req dto.TopicUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, apperr.Wrap(apperr.CodeValidation, "参数错误", err))
		return
	}
	t, err := h.topicSvc.Update(id, &service.TopicUpdateReq{
		Title: req.Title, Summary: req.Summary, Cover: req.Cover, Status: req.Status,
	})
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, t)
}

func (h *TopicHandler) Delete(c *gin.Context) {
	id, ok := parseTopicID(c)
	if !ok {
		return
	}
	if err := h.topicSvc.Delete(id); err != nil {
		Fail(c, err)
		return
	}
	OK(c, gin.H{"deleted": true})
}

// AdminItems 管理端查看专题条目（不限专题状态）
func (h *TopicHandler) AdminItems(c *gin.Context) {
	id, ok := parseTopicID(c)
	if !ok {
		return
	}
	if _, err := h.topicSvc.Get(id); err != nil {
		Fail(c, err)
		return
	}
	p := getPage(c)
	items, total, err := h.topicSvc.Items(id, p.Page, p.Size)
	if err != nil {
		Fail(c, err)
		return
	}
	PageOK(c, items, total, p.Page, p.Size)
}

func (h *TopicHandler) AddItem(c *gin.Context) {
	id, ok := parseTopicID(c)
	if !ok {
		return
	}
	var req dto.TopicAddItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, apperr.Wrap(apperr.CodeValidation, "参数错误", err))
		return
	}
	item, added, err := h.topicSvc.AddItem(id, req.ItemType, req.ItemID)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, gin.H{"item": item, "added": added})
}

func (h *TopicHandler) RemoveItem(c *gin.Context) {
	id, ok := parseTopicID(c)
	if !ok {
		return
	}
	itemID, err := strconv.ParseUint(c.Param("itemId"), 10, 64)
	if err != nil || itemID == 0 {
		Fail(c, apperr.ErrBadRequest)
		return
	}
	if err := h.topicSvc.RemoveItem(id, itemID); err != nil {
		Fail(c, err)
		return
	}
	OK(c, gin.H{"removed": true})
}
