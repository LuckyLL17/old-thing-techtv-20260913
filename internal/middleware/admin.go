package middleware

import (
	"upcycle-hub/internal/service"

	"github.com/gin-gonic/gin"
)

func RequireAdmin(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := MustLogin(c)
		if uid == 0 {
			c.JSON(401, gin.H{"code": 40100, "message": "未授权访问", "error_category": "permission", "success": false})
			c.Abort()
			return
		}
		u, err := authSvc.GetUserByID(uid)
		if err != nil || !u.IsAdmin || u.Status != 1 {
			c.JSON(403, gin.H{"code": 40300, "message": "需要管理员权限", "error_category": "permission", "success": false})
			c.Abort()
			return
		}
		c.Set("is_admin", true)
		c.Next()
	}
}
