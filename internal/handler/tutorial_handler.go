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

type TutorialHandler struct {
	tutorialSvc  *service.TutorialService
	interactSvc  *service.InteractionService
	categorySvc  *service.CategoryService
	tagSvc       *service.TagService
}

func NewTutorialHandler(t *service.TutorialService, i *service.InteractionService, c *service.CategoryService, tg *service.TagService) *TutorialHandler {
	return &TutorialHandler{tutorialSvc: t, interactSvc: i, categorySvc: c, tagSvc: tg}
}

func (h *TutorialHandler) List(c *gin.Context) {
	var req dto.TutorialListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		Fail(c, apperr.Wrap(apperr.CodeValidation, "参数错误", err))
		return
	}
	list, total, err := h.tutorialSvc.List(req.Page, req.Size, req.Category, req.Difficulty, req.Status, req.Sort, req.Keyword, req.UserID)
	if err != nil {
		Fail(c, err)
		return
	}
	PageOK(c, list, total, req.Page, req.Size)
}

func (h *TutorialHandler) Create(c *gin.Context) {
	uid := middleware.MustLogin(c)
	if uid == 0 {
		Fail(c, apperr.ErrUnauthorized)
		return
	}
	var req dto.TutorialCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, apperr.Wrap(apperr.CodeValidation, "参数错误", err))
		return
	}
	r := &service.TutorialCreateReq{
		UserID:         uid,
		CategoryID:     req.CategoryID,
		Title:          req.Title,
		Summary:        req.Summary,
		CoverBefore:    req.CoverBefore,
		CoverAfter:     req.CoverAfter,
		Difficulty:     req.Difficulty,
		EstimatedHours: req.EstimatedHours,
		Status:         req.Status,
		TagNames:       req.Tags,
	}
	for _, s := range req.Steps {
		r.Steps = append(r.Steps, &domain.Step{
			Title: s.Title, Content: s.Content, Image: s.Image,
			Reminder: s.Reminder, EstimatedMinutes: s.EstimatedMinutes,
		})
	}
	for _, m := range req.Materials {
		r.Materials = append(r.Materials, &domain.Material{
			Name: m.Name, Quantity: m.Quantity, Unit: m.Unit, Notes: m.Notes,
		})
	}
	for _, m := range req.Tools {
		r.Tools = append(r.Tools, &domain.Material{
			Name: m.Name, Quantity: m.Quantity, Unit: m.Unit, Notes: m.Notes,
		})
	}
	t, err := h.tutorialSvc.Create(r)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, t)
}

func (h *TutorialHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		Fail(c, apperr.ErrBadRequest)
		return
	}
	t, err := h.tutorialSvc.Get(id, true)
	if err != nil {
		Fail(c, err)
		return
	}
	uid := middleware.MustLogin(c)
	var faved bool
	if uid > 0 {
		faved, _ = h.interactSvc.IsFavorite(uid, domain.FavTypeTutorial, id)
	}
	OK(c, gin.H{"tutorial": t, "favorited": faved})
}

func (h *TutorialHandler) Update(c *gin.Context) {
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
	var req dto.TutorialCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, apperr.Wrap(apperr.CodeValidation, "参数错误", err))
		return
	}
	r := &service.TutorialCreateReq{
		UserID:         uid,
		CategoryID:     req.CategoryID,
		Title:          req.Title,
		Summary:        req.Summary,
		CoverBefore:    req.CoverBefore,
		CoverAfter:     req.CoverAfter,
		Difficulty:     req.Difficulty,
		EstimatedHours: req.EstimatedHours,
		Status:         req.Status,
		TagNames:       req.Tags,
	}
	for _, s := range req.Steps {
		r.Steps = append(r.Steps, &domain.Step{
			Title: s.Title, Content: s.Content, Image: s.Image,
			Reminder: s.Reminder, EstimatedMinutes: s.EstimatedMinutes,
		})
	}
	for _, m := range req.Materials {
		r.Materials = append(r.Materials, &domain.Material{
			Name: m.Name, Quantity: m.Quantity, Unit: m.Unit, Notes: m.Notes,
		})
	}
	for _, m := range req.Tools {
		r.Tools = append(r.Tools, &domain.Material{
			Name: m.Name, Quantity: m.Quantity, Unit: m.Unit, Notes: m.Notes,
		})
	}
	t, err := h.tutorialSvc.Update(id, uid, r)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, t)
}

func (h *TutorialHandler) Delete(c *gin.Context) {
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
	if err := h.tutorialSvc.Delete(id, uid); err != nil {
		Fail(c, err)
		return
	}
	OK(c, gin.H{"deleted": true})
}

func (h *TutorialHandler) ReorderSteps(c *gin.Context) {
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
	var req dto.ReorderStepsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, apperr.Wrap(apperr.CodeValidation, "参数错误", err))
		return
	}
	if err := h.tutorialSvc.ReorderSteps(id, uid, req.Order); err != nil {
		Fail(c, err)
		return
	}
	OK(c, gin.H{"reordered": true})
}

func (h *TutorialHandler) Comments(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		Fail(c, apperr.ErrBadRequest)
		return
	}
	p := getPage(c)
	list, total, e := h.interactSvc.ListComments(domain.CommentTypeTutorial, id, p.Page, p.Size)
	if e != nil {
		Fail(c, e)
		return
	}
	PageOK(c, list, total, p.Page, p.Size)
}

func (h *TutorialHandler) AddComment(c *gin.Context) {
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
	var req dto.CommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, apperr.Wrap(apperr.CodeValidation, "参数错误", err))
		return
	}
	cmt, err := h.interactSvc.Comment(uid, domain.CommentTypeTutorial, id, req.Content, req.ParentID)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, cmt)
}

func (h *TutorialHandler) Attempt(c *gin.Context) {
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
	var req dto.AttemptReq
	req.TutorialID = id
	if err := c.ShouldBindJSON(&req); err == nil {
		if req.TutorialID == 0 {
			req.TutorialID = id
		}
	}
	err = h.interactSvc.MarkAttempt(uid, id, req.Completed, req.Note)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, gin.H{"attempted": true})
}

func (h *TutorialHandler) Categories(c *gin.Context) {
	list, err := h.categorySvc.ListAll()
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, list)
}

func (h *TutorialHandler) Tags(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	list, err := h.tagSvc.ListPopular(limit)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, list)
}

func getPage(c *gin.Context) dto.PageReq {
	p := dto.PageReq{Page: 1, Size: 10}
	if v := c.Query("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			p.Page = n
		}
	}
	if v := c.Query("size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			p.Size = n
		}
	}
	return p
}
