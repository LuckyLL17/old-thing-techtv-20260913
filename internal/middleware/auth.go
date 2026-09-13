package middleware

import (
	"strings"
	"upcycle-hub/internal/service"
	apperr "upcycle-hub/pkg/errors"

	"github.com/gin-gonic/gin"
)

type UserCtxKey struct{}

func Auth(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			q := c.Query("token")
			if q != "" {
				header = "Bearer " + q
			}
		}
		if header == "" {
			c.JSON(401, respErr(apperr.ErrUnauthorized))
			c.Abort()
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(401, respErr(apperr.ErrUnauthorized))
			c.Abort()
			return
		}
		claims, err := authSvc.ParseToken(parts[1])
		if err != nil {
			c.JSON(401, respErr(err))
			c.Abort()
			return
		}
		c.Set("uid", claims.UserID)
		c.Set("uname", claims.Username)
		c.Next()
	}
}

func OptionalAuth(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			q := c.Query("token")
			if q != "" {
				header = "Bearer " + q
			}
		}
		if header == "" {
			c.Next()
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.Next()
			return
		}
		claims, err := authSvc.ParseToken(parts[1])
		if err == nil {
			c.Set("uid", claims.UserID)
			c.Set("uname", claims.Username)
		}
		c.Next()
	}
}

func MustLogin(c *gin.Context) uint64 {
	v, ok := c.Get("uid")
	if !ok {
		return 0
	}
	id, ok := v.(uint64)
	if !ok {
		return 0
	}
	return id
}

func respErr(err error) gin.H {
	if ae, ok := err.(*apperr.AppError); ok {
		return gin.H{"code": ae.Code, "message": ae.Message, "success": false}
	}
	return gin.H{"code": 50000, "message": err.Error(), "success": false}
}
