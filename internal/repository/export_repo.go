package repository

import (
	"upcycle-hub/internal/domain"
	apperr "upcycle-hub/pkg/errors"

	"gorm.io/gorm"
)

type ExportRepo struct {
	db *gorm.DB
}

func NewExportRepo(db *gorm.DB) *ExportRepo {
	return &ExportRepo{db: db}
}

func (r *ExportRepo) Create(e *domain.TutorialExport) error {
	if err := r.db.Create(e).Error; err != nil {
		return apperr.Wrap(apperr.CodeDB, "保存导出记录失败", err)
	}
	return nil
}

func (r *ExportRepo) ListByTutorial(tid uint64, limit int) ([]*domain.TutorialExport, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var list []*domain.TutorialExport
	err := r.db.Where("tutorial_id = ?", tid).
		Order("id DESC").Limit(limit).Find(&list).Error
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDB, "查询导出记录失败", err)
	}
	return list, nil
}
