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

type ProjectHandler struct {
	projectSvc  *service.ProjectService
	interactSvc *service.InteractionService
}

func NewProjectHandler(p *service.ProjectService, i *service.InteractionService) *ProjectHandler {
	return &ProjectHandler{projectSvc: p, interactSvc: i}
}

func (h *ProjectHandler) List(c *gin.Context) {
	var req dto.ProjectListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		Fail(c, apperr.Wrap(apperr.CodeValidation, "参数错误", err))
		return
	}
	list, total, err := h.projectSvc.List(req.Page, req.Size, req.TutorialID, req.UserID, req.Sort)
	if err != nil {
		Fail(c, err)
		return
	}
	PageOK(c, list, total, req.Page, req.Size)
}

func (h *ProjectHandler) Create(c *gin.Context) {
	uid := middleware.MustLogin(c)
	if uid == 0 {
		Fail(c, apperr.ErrUnauthorized)
		return
	}
	var req dto.ProjectCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, apperr.Wrap(apperr.CodeValidation, "参数错误", err))
		return
	}
	r := &service.ProjectCreateReq{
		UserID: uid, TutorialID: req.TutorialID, Title: req.Title,
		Description: req.Description, Images: req.Images, CustomNotes: req.CustomNotes, Rating: req.Rating,
	}
	p, err := h.projectSvc.Create(r)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, p)
}

func (h *ProjectHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		Fail(c, apperr.ErrBadRequest)
		return
	}
	p, err := h.projectSvc.Get(id)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, p)
}

func (h *ProjectHandler) Update(c *gin.Context) {
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
	var req dto.ProjectCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, apperr.Wrap(apperr.CodeValidation, "参数错误", err))
		return
	}
	r := &service.ProjectCreateReq{
		UserID: uid, TutorialID: req.TutorialID, Title: req.Title,
		Description: req.Description, Images: req.Images, CustomNotes: req.CustomNotes, Rating: req.Rating,
	}
	p, err := h.projectSvc.Update(id, uid, r)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, p)
}

func (h *ProjectHandler) Delete(c *gin.Context) {
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
	if err := h.projectSvc.Delete(id, uid); err != nil {
		Fail(c, err)
		return
	}
	OK(c, gin.H{"deleted": true})
}

func (h *ProjectHandler) Like(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		Fail(c, apperr.ErrBadRequest)
		return
	}
	if err := h.projectSvc.Like(id); err != nil {
		Fail(c, err)
		return
	}
	OK(c, gin.H{"liked": true})
}

func (h *ProjectHandler) Comments(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		Fail(c, apperr.ErrBadRequest)
		return
	}
	p := getPage(c)
	list, total, e := h.interactSvc.ListComments(domain.CommentTypeProject, id, p.Page, p.Size)
	if e != nil {
		Fail(c, e)
		return
	}
	PageOK(c, list, total, p.Page, p.Size)
}

func (h *ProjectHandler) AddComment(c *gin.Context) {
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
	cmt, err := h.interactSvc.Comment(uid, domain.CommentTypeProject, id, req.Content, req.ParentID)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, cmt)
}

func (h *ProjectHandler) AddProjectUnderTutorial(c *gin.Context) {
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
	var req dto.ProjectCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, apperr.Wrap(apperr.CodeValidation, "参数错误", err))
		return
	}
	req.TutorialID = tid
	r := &service.ProjectCreateReq{
		UserID: uid, TutorialID: req.TutorialID, Title: req.Title,
		Description: req.Description, Images: req.Images, CustomNotes: req.CustomNotes, Rating: req.Rating,
	}
	p, err := h.projectSvc.Create(r)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, p)
}
