package service

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
	apperr "upcycle-hub/pkg/errors"
	"upcycle-hub/pkg/logger"
)

const (
	AuditExportErrorPermission   = "permission"
	AuditExportErrorRange        = "range"
	AuditExportErrorGeneration   = "generation"
	auditExportBatchSize         = 1000
	auditExportTTL               = 7 * 24 * time.Hour
	auditExportCleanupInterval   = time.Hour
)

type AuditExportService struct {
	repo    *repository.AuditExportRepo
	logRepo *repository.AuditLogRepo
	dir     string
	workers chan struct{}
	queue   chan uint64
	mu      sync.Mutex
	pending map[uint64]struct{}
}

func NewAuditExportService(repo *repository.AuditExportRepo, logRepo *repository.AuditLogRepo, dir string) *AuditExportService {
	if dir == "" {
		dir = filepath.Join("data", "audit-exports")
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}
	if err := os.MkdirAll(absDir, 0o750); err == nil {
		if entries, readErr := os.ReadDir(absDir); readErr == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tmp") {
					_ = os.Remove(filepath.Join(absDir, entry.Name()))
				}
			}
		}
	} else {
		logger.Warnf("创建审计导出目录失败: %v", err)
	}
	s := &AuditExportService{
		repo:    repo,
		logRepo: logRepo,
		dir:     absDir,
		workers: make(chan struct{}, 2),
		queue:   make(chan uint64, 100),
		pending: map[uint64]struct{}{},
	}
	s.resetInterrupted()
	go s.dispatchLoop()
	go s.cleanupLoop()
	return s
}

func (s *AuditExportService) Create(requesterID uint64, f repository.AuditFilter) (*domain.AuditExport, error) {
	if f.From != nil && f.To != nil && f.From.After(*f.To) {
		return nil, apperr.New(apperr.CodeValidation, "时间范围无效：开始时间不能晚于结束时间")
	}
	if f.From != nil && f.To != nil && f.To.Sub(*f.From) > 365*24*time.Hour {
		return nil, apperr.New(apperr.CodeValidation, "时间范围过大：单次导出不能超过 365 天")
	}

	now := time.Now()
	expiresAt := now.Add(auditExportTTL)
	job := &domain.AuditExport{
		RequesterID: requesterID,
		Status:      domain.AuditExportQueued,
		UserID:      f.UserID,
		Operator:    strings.TrimSpace(f.Operator),
		Action:      strings.TrimSpace(f.Action),
		TargetType:  strings.TrimSpace(f.TargetType),
		From:        f.From,
		To:          f.To,
		ExpiresAt:   &expiresAt,
	}
	if err := s.repo.Create(job); err != nil {
		return nil, err
	}
	s.enqueue(job.ID)
	return job, nil
}

func (s *AuditExportService) Get(requesterID uint64, id uint64, isAdmin bool) (*domain.AuditExport, error) {
	job, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if job.RequesterID != requesterID && !isAdmin {
		return nil, apperr.New(apperr.CodeForbidden, "无权访问该导出任务")
	}
	return job, nil
}

func (s *AuditExportService) List(requesterID uint64) ([]*domain.AuditExport, error) {
	return s.repo.ListByRequester(requesterID, 20)
}

func (s *AuditExportService) FilePath(job *domain.AuditExport) (string, string, error) {
	if job.Status != domain.AuditExportSucceeded {
		return "", "", apperr.New(apperr.CodeBadRequest, "导出文件尚未生成")
	}
	if job.FilePath == "" || job.FileName == "" {
		return "", "", apperr.New(apperr.CodeInternal, "导出文件信息缺失")
	}
	cleanPath := filepath.Clean(job.FilePath)
	if filepath.Dir(cleanPath) != filepath.Clean(s.dir) {
		return "", "", apperr.New(apperr.CodeForbidden, "非法的导出文件路径")
	}
	info, err := os.Stat(cleanPath)
	if err != nil || info.IsDir() {
		return "", "", apperr.New(apperr.CodeNotFound, "导出文件已过期或不存在")
	}
	return cleanPath, job.FileName, nil
}

func (s *AuditExportService) resetInterrupted() {
	if err := s.repo.MarkInterrupted(); err != nil {
		logger.Warnf("重置未完成审计导出任务失败: %v", err)
	}
}

func (s *AuditExportService) enqueue(id uint64) {
	s.mu.Lock()
	if _, ok := s.pending[id]; ok {
		s.mu.Unlock()
		return
	}
	s.pending[id] = struct{}{}
	s.mu.Unlock()

	select {
	case s.queue <- id:
	default:
		s.markFailed(id, AuditExportErrorGeneration, "导出队列已满，请稍后重试")
		s.forget(id)
	}
}

func (s *AuditExportService) forget(id uint64) {
	s.mu.Lock()
	delete(s.pending, id)
	s.mu.Unlock()
}

func (s *AuditExportService) dispatchLoop() {
	for id := range s.queue {
		id := id
		s.workers <- struct{}{}
		go func() {
			defer func() { <-s.workers; s.forget(id) }()
			if err := s.run(id); err != nil {
				logger.Errorf("生成审计导出 %d 失败: %v", id, err)
			}
		}()
	}
}

func (s *AuditExportService) run(id uint64) error {
	job, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if job.Status != domain.AuditExportQueued {
		return nil
	}
	now := time.Now()
	job.Status = domain.AuditExportRunning
	job.StartedAt = &now
	if err := s.repo.Update(job); err != nil {
		return err
	}

	filter := repository.AuditFilter{
		UserID:     job.UserID,
		Operator:   job.Operator,
		Action:     job.Action,
		TargetType: job.TargetType,
		From:       job.From,
		To:         job.To,
	}
	fileName := fmt.Sprintf("audit_logs_%s_%d.csv", now.Format("20060102_150405"), id)
	tempPath := filepath.Join(s.dir, fileName+".tmp")
	finalPath := filepath.Join(s.dir, fileName)
	out, err := os.Create(tempPath)
	if err != nil {
		s.markFailed(id, AuditExportErrorGeneration, "创建导出文件失败")
		return err
	}

	if _, err := out.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		out.Close()
		os.Remove(tempPath)
		s.markFailed(id, AuditExportErrorGeneration, "写入导出文件失败")
		return err
	}
	w := csv.NewWriter(out)
	if err := w.Write([]string{"时间", "操作人", "动作", "目标", "来源地址", "备注"}); err != nil {
		out.Close()
		os.Remove(tempPath)
		s.markFailed(id, AuditExportErrorGeneration, "写入 CSV 表头失败")
		return err
	}

	var written int64
	err = s.logRepo.ForEachFiltered(filter, auditExportBatchSize, func(logs []*domain.AuditLog) error {
		for _, l := range logs {
			record := []string{
				l.CreatedAt.Format("2006-01-02 15:04:05"),
				operatorName(l),
				actionLabel(l.Action),
				targetName(l),
				l.IP,
				l.Remark,
			}
			if err := w.Write(record); err != nil {
				return err
			}
			written++
		}
		w.Flush()
		return w.Error()
	})
	if err != nil {
		out.Close()
		os.Remove(tempPath)
		s.markFailed(id, AuditExportErrorGeneration, "写入审计日志失败")
		return err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		out.Close()
		os.Remove(tempPath)
		s.markFailed(id, AuditExportErrorGeneration, "写入 CSV 数据失败")
		return err
	}
	if err := out.Sync(); err != nil {
		out.Close()
		os.Remove(tempPath)
		s.markFailed(id, AuditExportErrorGeneration, "保存导出文件失败")
		return err
	}
	if err := out.Close(); err != nil {
		os.Remove(tempPath)
		s.markFailed(id, AuditExportErrorGeneration, "关闭导出文件失败")
		return err
	}
	if err := os.Rename(tempPath, finalPath); err != nil {
		os.Remove(tempPath)
		s.markFailed(id, AuditExportErrorGeneration, "保存导出文件失败")
		return err
	}

	job, err = s.repo.GetByID(id)
	if err != nil {
		return err
	}
	finished := time.Now()
	job.Status = domain.AuditExportSucceeded
	job.FilePath = finalPath
	job.FileName = fileName
	job.RowCount = written
	job.StartedAt = firstTime(job.StartedAt, &now)
	job.FinishedAt = &finished
	job.ErrorCategory = ""
	job.ErrorMessage = ""
	if err := s.repo.Update(job); err != nil {
		os.Remove(finalPath)
		return err
	}
	return nil
}

func firstTime(values ...*time.Time) *time.Time {
	for _, v := range values {
		if v != nil {
			return v
		}
	}
	return nil
}

func (s *AuditExportService) markFailed(id uint64, category, message string) {
	job, err := s.repo.GetByID(id)
	if err != nil {
		return
	}
	now := time.Now()
	job.Status = domain.AuditExportFailed
	job.ErrorCategory = category
	if len(message) > 500 {
		message = message[:500]
	}
	job.ErrorMessage = message
	job.FinishedAt = &now
	if job.StartedAt == nil {
		job.StartedAt = &now
	}
	if err := s.repo.Update(job); err != nil {
		logger.Warnf("标记审计导出任务失败: %v", err)
	}
}

func (s *AuditExportService) cleanupLoop() {
	ticker := time.NewTicker(auditExportCleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		s.cleanupExpired()
	}
}

func (s *AuditExportService) cleanupExpired() {
	list, err := s.repo.DeleteExpired(time.Now())
	if err != nil {
		logger.Warnf("清理过期审计导出任务失败: %v", err)
		return
	}
	for _, job := range list {
		if job.FilePath != "" {
			if cleanPath := filepath.Clean(job.FilePath); filepath.Dir(cleanPath) == filepath.Clean(s.dir) {
				if err := os.Remove(cleanPath); err != nil && !os.IsNotExist(err) {
					logger.Warnf("删除过期审计导出文件失败: %v", err)
				}
			}
		}
	}
}

func operatorName(l *domain.AuditLog) string {
	if l.User != nil {
		name := l.User.DisplayName()
		if l.User.Username != "" && l.User.Username != name {
			return fmt.Sprintf("%s(%s,ID:%d)", name, l.User.Username, l.UserID)
		}
		return fmt.Sprintf("%s(ID:%d)", name, l.UserID)
	}
	return "用户ID:" + strconv.FormatUint(l.UserID, 10)
}

func targetName(l *domain.AuditLog) string {
	targetType := strings.TrimSpace(l.TargetType)
	if targetType == "" {
		targetType = "-"
	}
	if l.TargetID > 0 {
		return targetType + "#" + strconv.FormatUint(l.TargetID, 10)
	}
	return targetType
}

func actionLabel(action string) string {
	labels := map[string]string{
		domain.AuditLogin:             "登录",
		domain.AuditLogout:            "退出登录",
		domain.AuditRegister:          "注册",
		domain.AuditPasswordReset:     "重置密码",
		domain.AuditProfileUpdate:     "更新资料",
		domain.AuditTutorialCreate:    "创建教程",
		domain.AuditTutorialUpdate:    "更新教程",
		domain.AuditTutorialDelete:    "删除教程",
		domain.AuditTutorialPublish:   "发布教程",
		domain.AuditTutorialArchive:   "归档教程",
		domain.AuditTutorialRollback:  "回滚教程版本",
		domain.AuditProjectCreate:     "创建作品",
		domain.AuditProjectDelete:     "删除作品",
		domain.AuditCommentCreate:     "发表评论",
		domain.AuditCommentDelete:     "删除评论",
		domain.AuditFavoriteToggle:    "切换收藏",
		domain.AuditFollowToggle:      "切换关注",
		domain.AuditAttemptToggle:     "切换尝试",
		domain.AuditAdminAuditPass:    "审核通过",
		domain.AuditAdminAuditReject:  "审核驳回",
		domain.AuditAdminUserBan:      "禁用用户",
		domain.AuditAdminCategoryEdit: "编辑分类",
	}
	if label := labels[action]; label != "" {
		return label + "(" + action + ")"
	}
	return action
}
