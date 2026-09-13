package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"upcycle-hub/config"
	"upcycle-hub/internal/handler"
	"upcycle-hub/internal/middleware"
	"upcycle-hub/internal/service"
	"upcycle-hub/pkg/logger"
	"upcycle-hub/pkg/utils"

	"github.com/gin-gonic/gin"
)

type Deps struct {
	Cfg            *config.Config
	AuthSvc        *service.AuthService
	TutorialSvc    *service.TutorialService
	ProjectSvc     *service.ProjectService
	CategorySvc    *service.CategoryService
	TagSvc         *service.TagService
	SearchSvc      *service.SearchService
	RecommendSvc   *service.RecommendService
	StatsSvc       *service.StatsService
	InteractSvc    *service.InteractionService
	NotifSvc       *service.NotificationService
	AuditSvc       *service.AuditService
	HistorySvc     *service.TutorialHistoryService
	FrontendDir    string
}

func SetupRouter(d *Deps) *gin.Engine {
	gin.SetMode(d.Cfg.Server.Mode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit(&d.Cfg.Rate))
	api := r.Group("/api/v1")
	authH := handler.NewAuthHandler(d.AuthSvc, d.StatsSvc)
	tutH := handler.NewTutorialHandler(d.TutorialSvc, d.InteractSvc, d.CategorySvc, d.TagSvc)
	projH := handler.NewProjectHandler(d.ProjectSvc, d.InteractSvc)
	searchH := handler.NewSearchHandler(d.SearchSvc, d.RecommendSvc)
	statsH := handler.NewStatsHandler(d.StatsSvc, d.InteractSvc)
	notifH := handler.NewNotificationHandler(d.NotifSvc)
	auditH := handler.NewAuditHandler(d.AuditSvc)
	histH := handler.NewTutorialHistoryHandler(d.HistorySvc)
	api.GET("/home", searchH.Home)
	api.GET("/random", searchH.Random)
	api.GET("/top", searchH.Top)
	api.GET("/search", searchH.Search)
	api.GET("/categories", tutH.Categories)
	api.GET("/tags", tutH.Tags)
	api.GET("/stats", statsH.Dashboard)
	auth := api.Group("/auth")
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
		auth.GET("/me", middleware.Auth(d.AuthSvc), authH.Me)
		auth.PUT("/me", middleware.Auth(d.AuthSvc), authH.UpdateProfile)
		auth.POST("/reset-password", middleware.Auth(d.AuthSvc), authH.ResetPassword)
		auth.GET("/center", middleware.Auth(d.AuthSvc), authH.Center)
	}
	tuts := api.Group("/tutorials")
	{
		tuts.GET("", tutH.List)
		tuts.GET("/:id", middleware.OptionalAuth(d.AuthSvc), tutH.Get)
		tuts.POST("", middleware.Auth(d.AuthSvc), tutH.Create)
		tuts.PUT("/:id", middleware.Auth(d.AuthSvc), tutH.Update)
		tuts.DELETE("/:id", middleware.Auth(d.AuthSvc), tutH.Delete)
		tuts.POST("/:id/reorder", middleware.Auth(d.AuthSvc), tutH.ReorderSteps)
		tuts.GET("/:id/comments", tutH.Comments)
		tuts.POST("/:id/comments", middleware.Auth(d.AuthSvc), tutH.AddComment)
		tuts.POST("/:id/attempt", middleware.Auth(d.AuthSvc), tutH.Attempt)
		tuts.POST("/:id/projects", middleware.Auth(d.AuthSvc), projH.AddProjectUnderTutorial)
		tuts.GET("/:id/history", histH.List)
		tuts.GET("/:id/history/:version", histH.Get)
		tuts.POST("/:id/history/snapshot", middleware.Auth(d.AuthSvc), histH.Snapshot)
		tuts.POST("/:id/history/rollback", middleware.Auth(d.AuthSvc), histH.Rollback)
	}
	projs := api.Group("/projects")
	{
		projs.GET("", projH.List)
		projs.GET("/:id", projH.Get)
		projs.POST("", middleware.Auth(d.AuthSvc), projH.Create)
		projs.PUT("/:id", middleware.Auth(d.AuthSvc), projH.Update)
		projs.DELETE("/:id", middleware.Auth(d.AuthSvc), projH.Delete)
		projs.POST("/:id/like", projH.Like)
		projs.GET("/:id/comments", projH.Comments)
		projs.POST("/:id/comments", middleware.Auth(d.AuthSvc), projH.AddComment)
	}
	me := api.Group("/me", middleware.Auth(d.AuthSvc))
	{
		me.GET("/favorites", statsH.Favorites)
		me.POST("/favorites", statsH.ToggleFavorite)
		me.GET("/attempts", statsH.Attempts)
		me.POST("/follow", statsH.Follow)
		me.DELETE("/follow/:id", statsH.Unfollow)
		me.POST("/messages", statsH.SendMessage)
		me.GET("/messages/:id", statsH.Messages)
		me.GET("/unread", statsH.Unread)
		me.GET("/notifications", notifH.List)
		me.GET("/notifications/unread-count", notifH.CountUnread)
		me.POST("/notifications/:id/read", notifH.MarkRead)
		me.POST("/notifications/read-all", notifH.MarkAllRead)
		me.DELETE("/notifications", notifH.Clear)
	}
	admin := api.Group("/admin", middleware.Auth(d.AuthSvc))
	{
		admin.GET("/audit", auditH.List)
		admin.GET("/audit/stats", auditH.Stats)
	}
	api.GET("/users/:id/follow", middleware.OptionalAuth(d.AuthSvc), statsH.FollowInfo)
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "version": "1.0"})
	})
	if d.FrontendDir != "" {
		r.StaticFS("/uploads", http.Dir(d.Cfg.Upload.Dir))
		r.StaticFS("/static", http.Dir(d.FrontendDir))
		r.NoRoute(serveFrontendDir(d.FrontendDir))
	}
	return r
}

func serveFrontendDir(dir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/" || path == "" {
			path = "index.html"
		}
		fp := filepath.Join(dir, filepath.Clean("/"+strings.TrimPrefix(path, "/")))
		if _, err := os.Stat(fp); err != nil {
			fp = filepath.Join(dir, "index.html")
		}
		ext := strings.ToLower(filepath.Ext(fp))
		ct := "text/html; charset=utf-8"
		switch ext {
		case ".css":
			ct = "text/css; charset=utf-8"
		case ".js":
			ct = "application/javascript; charset=utf-8"
		case ".png":
			ct = "image/png"
		case ".jpg", ".jpeg":
			ct = "image/jpeg"
		case ".svg":
			ct = "image/svg+xml"
		case ".ico":
			ct = "image/x-icon"
		case ".json":
			ct = "application/json"
		case ".woff":
			ct = "font/woff"
		case ".woff2":
			ct = "font/woff2"
		}
		data, err := os.ReadFile(fp)
		if err != nil {
			c.String(404, "not found")
			return
		}
		c.Data(200, ct, data)
	}
}

func UploadHandler(cfg *config.UploadConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		utils.EnsureDir(cfg.Dir)
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(400, gin.H{"code": 400, "message": "缺少上传文件", "success": false})
			return
		}
		path, err := utils.SaveUploadedFile(file, cfg.Dir, cfg.MaxSize)
		if err != nil {
			logger.Errorf("save upload: %v", err)
			c.JSON(500, gin.H{"code": 50003, "message": err.Error(), "success": false})
			return
		}
		thumb, _ := utils.GenerateThumbnail(path, 300)
		_ = utils.ResizeImage(path, path, cfg.ImageMax, cfg.ImageMax)
		c.JSON(200, gin.H{
			"code": 0, "message": "ok", "success": true,
			"data": gin.H{"path": "/uploads/" + basename(path), "thumb": "/uploads/" + basename(thumb), "size": file.Size},
		})
	}
}

func basename(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' || p[i] == '\\' {
			return p[i+1:]
		}
	}
	return p
}
