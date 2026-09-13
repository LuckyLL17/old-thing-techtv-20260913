package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
	"upcycle-hub/api"
	"upcycle-hub/config"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/middleware"
	"upcycle-hub/internal/repository"
	"upcycle-hub/internal/service"
	"upcycle-hub/internal/worker"
	"upcycle-hub/pkg/logger"
	"upcycle-hub/pkg/utils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	cfgPath := "config/config.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		wd, _ := os.Getwd()
		cfgPath = filepath.Join(wd, "config/config.yaml")
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}
	if err := logger.Init(&cfg.Log); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()
	if err := utils.EnsureDir(cfg.Upload.Dir); err != nil {
		logger.Errorf("创建上传目录失败: %v", err)
	}
	frontendDir := findFrontendDir(cfgPath)
	db, err := gorm.Open(sqlite.Open(cfg.DB.DSN), &gorm.Config{})
	if err != nil {
		logger.Errorf("连接数据库失败: %v", err)
		os.Exit(1)
	}
	sqlDB, err := db.DB()
	if err != nil {
		logger.Errorf("获取数据库句柄失败: %v", err)
		os.Exit(1)
	}
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)
	if err := autoMigrate(db); err != nil {
		logger.Errorf("数据库迁移失败: %v", err)
		os.Exit(1)
	}
	userRepo := repository.NewUserRepo(db)
	tutorialRepo := repository.NewTutorialRepo(db)
	stepRepo := repository.NewStepRepo(db)
	materialRepo := repository.NewMaterialRepo(db)
	projectRepo := repository.NewProjectRepo(db)
	categoryRepo := repository.NewCategoryRepo(db)
	tagRepo := repository.NewTagRepo(db)
	commentRepo := repository.NewCommentRepo(db)
	favoriteRepo := repository.NewFavoriteRepo(db)
	attemptRepo := repository.NewAttemptRepo(db)
	followRepo := repository.NewFollowRepo(db)
	messageRepo := repository.NewMessageRepo(db)
	toolRepo := repository.NewToolRepo(db)
	versionRepo := repository.NewTutorialVersionRepo(db)
	notifRepo := repository.NewNotificationRepo(db)
	auditRepo := repository.NewAuditLogRepo(db)
	if err := categoryRepo.InitDefaults(); err != nil {
		logger.Warnf("初始化分类失败: %v", err)
	}
	authSvc := service.NewAuthService(userRepo, &cfg.JWT)
	tutorialSvc := service.NewTutorialService(tutorialRepo, stepRepo, materialRepo, tagRepo, categoryRepo, userRepo)
	projectSvc := service.NewProjectService(projectRepo, tutorialRepo, userRepo)
	categorySvc := service.NewCategoryService(categoryRepo)
	tagSvc := service.NewTagService(tagRepo)
	searchSvc := service.NewSearchService(tutorialRepo, tagRepo, categoryRepo)
	recommendSvc := service.NewRecommendService(tutorialRepo, tagRepo, categoryRepo)
	statsSvc := service.NewStatsService(tutorialRepo, projectRepo, userRepo, categoryRepo, favoriteRepo, attemptRepo)
	interactSvc := service.NewInteractionService(commentRepo, favoriteRepo, attemptRepo, followRepo, messageRepo, tutorialRepo, projectRepo)
	notifSvc := service.NewNotificationService(notifRepo)
	auditSvc := service.NewAuditService(auditRepo)
	historySvc := service.NewTutorialHistoryService(versionRepo, tutorialRepo, stepRepo, materialRepo, toolRepo)
	updater := worker.NewStatsUpdater(userRepo, tutorialRepo, commentRepo, categoryRepo, tagRepo)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	updater.Start(ctx)
	r := api.SetupRouter(&api.Deps{
		Cfg: cfg, AuthSvc: authSvc, TutorialSvc: tutorialSvc,
		ProjectSvc: projectSvc, CategorySvc: categorySvc, TagSvc: tagSvc,
		SearchSvc: searchSvc, RecommendSvc: recommendSvc,
		StatsSvc: statsSvc, InteractSvc: interactSvc,
		NotifSvc: notifSvc, AuditSvc: auditSvc, HistorySvc: historySvc,
		FrontendDir: frontendDir,
	})
	r.POST("/api/v1/upload", middleware.Auth(authSvc), api.UploadHandler(&cfg.Upload))
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{Addr: addr, Handler: r}
	go func() {
		logger.Infof("🚀 旧物改造平台启动于 http://%s", addr)
		if frontendDir != "" {
			logger.Infof("📄 前端资源目录: %s", frontendDir)
		}
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("启动服务器失败: %v", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("正在关闭服务...")
	updater.Stop()
	sctx, scancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer scancel()
	if err := srv.Shutdown(sctx); err != nil {
		logger.Errorf("服务关闭异常: %v", err)
	}
	logger.Info("服务已优雅退出")
}

func findFrontendDir(cfgPath string) string {
	candidates := []string{}
	cwd, _ := os.Getwd()
	candidates = append(candidates, filepath.Join(cwd, "frontend"))
	cfgDir := filepath.Dir(cfgPath)
	candidates = append(candidates, filepath.Join(cfgDir, "..", "frontend"))
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "frontend"))
	}
	for _, p := range candidates {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			idx := filepath.Join(p, "index.html")
			if _, err := os.Stat(idx); err == nil {
				abs, _ := filepath.Abs(p)
				return abs
			}
		}
	}
	return ""
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&domain.User{},
		&domain.Category{},
		&domain.Tag{},
		&domain.Tutorial{},
		&domain.TutorialVersion{},
		&domain.TutorialTag{},
		&domain.Step{},
		&domain.Material{},
		&domain.Tool{},
		&domain.Project{},
		&domain.Comment{},
		&domain.Favorite{},
		&domain.Attempt{},
		&domain.Follow{},
		&domain.Message{},
		&domain.Notification{},
		&domain.AuditLog{},
	)
}
