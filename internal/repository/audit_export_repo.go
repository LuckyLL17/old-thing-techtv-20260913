package repository

import (
	"time"
	"upcycle-hub/internal/domain"
	apperr "upcycle-hub/pkg/errors"

	"gorm.io/gorm"
)

type AuditExportRepo struct {
	db *gorm.DB
}

func NewAuditExportRepo(db *gorm.DB) *AuditExportRepo {
	return &AuditExportRepo{db: db}
}

func (r *AuditExportRepo) Create(e *domain.AuditExport) error {
	if err := r.db.Create(e).Error; err != nil {
		return apperr.Wrap(apperr.CodeDB, "创建审计导出任务失败", err)
	}
	return nil
}

func (r *AuditExportRepo) GetByID(id uint64) (*domain.AuditExport, error) {
	e := &domain.AuditExport{}
	if err := r.db.First(e, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.New(apperr.CodeNotFound, "导出任务不存在")
		}
		return nil, apperr.Wrap(apperr.CodeDB, "查询审计导出任务失败", err)
	}
	return e, nil
}

func (r *AuditExportRepo) Update(e *domain.AuditExport) error {
	if err := r.db.Save(e).Error; err != nil {
		return apperr.Wrap(apperr.CodeDB, "更新审计导出任务失败", err)
	}
	return nil
}

func (r *AuditExportRepo) ListByRequester(requesterID uint64, limit int) ([]*domain.AuditExport, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var list []*domain.AuditExport
	err := r.db.Where("requester_id = ?", requesterID).Order("id DESC").Limit(limit).Find(&list).Error
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDB, "查询审计导出任务失败", err)
	}
	return list, nil
}

func (r *AuditExportRepo) MarkInterrupted() error {
	now := time.Now()
	return r.db.Model(&domain.AuditExport{}).
		Where("status IN ?", []string{domain.AuditExportQueued, domain.AuditExportRunning}).
		Updates(map[string]any{
			"status":         domain.AuditExportFailed,
			"error_category": "generation",
			"error_message":  "服务重启，导出任务被中断",
			"finished_at":    now,
			"updated_at":     now,
		}).Error
}

func (r *AuditExportRepo) DeleteExpired(before time.Time) ([]*domain.AuditExport, error) {
	var list []*domain.AuditExport
	if err := r.db.Where("expires_at IS NOT NULL AND expires_at < ?", before).Find(&list).Error; err != nil {
		return nil, apperr.Wrap(apperr.CodeDB, "查询过期审计导出任务失败", err)
	}
	if err := r.db.Where("expires_at IS NOT NULL AND expires_at < ?", before).Delete(&domain.AuditExport{}).Error; err != nil {
		return nil, apperr.Wrap(apperr.CodeDB, "删除过期审计导出任务失败", err)
	}
	return list, nil
}
