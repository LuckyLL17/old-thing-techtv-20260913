package repository

import (
	"upcycle-hub/internal/domain"
	apperr "upcycle-hub/pkg/errors"

	"gorm.io/gorm"
)

type ToolRepo struct {
	db *gorm.DB
}

func NewToolRepo(db *gorm.DB) *ToolRepo {
	return &ToolRepo{db: db}
}

func (r *ToolRepo) Create(t *domain.Tool) error {
	err := r.db.Create(t).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "创建工具失败", err)
	}
	return nil
}

func (r *ToolRepo) BatchCreate(list []*domain.Tool) error {
	if len(list) == 0 {
		return nil
	}
	err := r.db.Create(&list).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "批量创建工具失败", err)
	}
	return nil
}

func (r *ToolRepo) Update(t *domain.Tool) error {
	err := r.db.Save(t).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "更新工具失败", err)
	}
	return nil
}

func (r *ToolRepo) Delete(id uint64) error {
	err := r.db.Delete(&domain.Tool{}, id).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "删除工具失败", err)
	}
	return nil
}

func (r *ToolRepo) DeleteByTutorial(tid uint64) error {
	err := r.db.Where("tutorial_id = ?", tid).Delete(&domain.Tool{}).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "删除工具失败", err)
	}
	return nil
}

func (r *ToolRepo) ListByTutorial(tid uint64) ([]*domain.Tool, error) {
	var list []*domain.Tool
	err := r.db.Where("tutorial_id = ?", tid).Order("sort_order ASC, id ASC").Find(&list).Error
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDB, "查询工具列表失败", err)
	}
	return list, nil
}

func (r *ToolRepo) GetByID(id uint64) (*domain.Tool, error) {
	t := &domain.Tool{}
	err := r.db.First(t, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeDB, "查询工具失败", err)
	}
	return t, nil
}

func (r *ToolRepo) ReplaceByTutorial(tid uint64, list []*domain.Tool) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return apperr.Wrap(apperr.CodeDB, "启动事务失败", tx.Error)
	}
	if err := tx.Where("tutorial_id = ?", tid).Delete(&domain.Tool{}).Error; err != nil {
		tx.Rollback()
		return apperr.Wrap(apperr.CodeDB, "重置工具失败", err)
	}
	if len(list) > 0 {
		if err := tx.Create(&list).Error; err != nil {
			tx.Rollback()
			return apperr.Wrap(apperr.CodeDB, "批量创建工具失败", err)
		}
	}
	if err := tx.Commit().Error; err != nil {
		return apperr.Wrap(apperr.CodeDB, "提交事务失败", err)
	}
	return nil
}
