package handler

import (
	"net/http"
	"upcycle-hub/api/dto"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/middleware"
	"upcycle-hub/internal/service"
	apperr "upcycle-hub/pkg/errors"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authSvc    *service.AuthService
	statsSvc   *service.StatsService
}

func NewAuthHandler(a *service.AuthService, s *service.StatsService) *AuthHandler {
	return &AuthHandler{authSvc: a, statsSvc: s}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, apperr.Wrap(apperr.CodeValidation, "参数错误", err))
		return
	}
	u, err := h.authSvc.Register(req.Username, req.Email, req.Password)
	if err != nil {
		Fail(c, err)
		return
	}
	token, err := h.authSvc.GenerateToken(u)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, dto.AuthResp{Token: token, User: u})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, apperr.Wrap(apperr.CodeValidation, "参数错误", err))
		return
	}
	u, token, err := h.authSvc.Login(req.Account, req.Password)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, dto.AuthResp{Token: token, User: u})
}

func (h *AuthHandler) Me(c *gin.Context) {
	uid := middleware.MustLogin(c)
	if uid == 0 {
		Fail(c, apperr.ErrUnauthorized)
		return
	}
	u, err := h.authSvc.GetUserByID(uid)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, u)
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	uid := middleware.MustLogin(c)
	if uid == 0 {
		Fail(c, apperr.ErrUnauthorized)
		return
	}
	var req dto.UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, apperr.Wrap(apperr.CodeValidation, "参数错误", err))
		return
	}
	u := &domain.User{
		ID:        uid,
		Nickname:  req.Nickname,
		Avatar:    req.Avatar,
		Specialty: req.Specialty,
		Bio:       req.Bio,
	}
	if err := h.authSvc.UpdateProfile(u); err != nil {
		Fail(c, err)
		return
	}
	OK(c, gin.H{"updated": true})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	uid := middleware.MustLogin(c)
	if uid == 0 {
		Fail(c, apperr.ErrUnauthorized)
		return
	}
	var req dto.ResetPwdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, apperr.Wrap(apperr.CodeValidation, "参数错误", err))
		return
	}
	if err := h.authSvc.ResetPassword(uid, req.OldPassword, req.NewPassword); err != nil {
		Fail(c, err)
		return
	}
	OK(c, gin.H{"reset": true})
}

func (h *AuthHandler) Center(c *gin.Context) {
	uid := middleware.MustLogin(c)
	if uid == 0 {
		Fail(c, apperr.ErrUnauthorized)
		return
	}
	st, err := h.statsSvc.UserCenter(uid)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, st)
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "success": true, "data": data})
}

func Fail(c *gin.Context, err error) {
	if ae, ok := err.(*apperr.AppError); ok {
		c.JSON(httpCode(ae.Code), gin.H{"code": ae.Code, "message": ae.Message, "success": false})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"code": 50000, "message": err.Error(), "success": false})
}

func PageOK(c *gin.Context, list interface{}, total int64, page, size int) {
	OK(c, dto.PageResp{List: list, Total: total, Page: page, Size: size})
}

func httpCode(code int) int {
	switch {
	case code >= 50000:
		return http.StatusInternalServerError
	case code >= 42900:
		return http.StatusTooManyRequests
	case code >= 42200:
		return http.StatusUnprocessableEntity
	case code >= 40900:
		return http.StatusConflict
	case code >= 40400:
		return http.StatusNotFound
	case code >= 40300:
		return http.StatusForbidden
	case code >= 40100:
		return http.StatusUnauthorized
	case code >= 40000:
		return http.StatusBadRequest
	default:
		return http.StatusOK
	}
}
