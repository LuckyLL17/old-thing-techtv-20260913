package repository

import (
	"upcycle-hub/internal/domain"
	apperr "upcycle-hub/pkg/errors"

	"gorm.io/gorm"
)

type MaterialRepo struct {
	db *gorm.DB
}

func NewMaterialRepo(db *gorm.DB) *MaterialRepo {
	return &MaterialRepo{db: db}
}

func (r *MaterialRepo) Create(m *domain.Material) error {
	err := r.db.Create(m).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "创建材料失败", err)
	}
	return nil
}

func (r *MaterialRepo) BatchCreate(list []*domain.Material) error {
	if len(list) == 0 {
		return nil
	}
	err := r.db.Create(&list).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "批量创建材料失败", err)
	}
	return nil
}

func (r *MaterialRepo) Update(m *domain.Material) error {
	err := r.db.Save(m).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "更新材料失败", err)
	}
	return nil
}

func (r *MaterialRepo) Delete(id uint64) error {
	err := r.db.Delete(&domain.Material{}, id).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "删除材料失败", err)
	}
	return nil
}

func (r *MaterialRepo) DeleteByTutorial(tid uint64) error {
	err := r.db.Where("tutorial_id = ?", tid).Delete(&domain.Material{}).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "删除材料失败", err)
	}
	return nil
}

func (r *MaterialRepo) ListByTutorial(tid uint64) ([]*domain.Material, []*domain.Material, error) {
	var all []*domain.Material
	err := r.db.Where("tutorial_id = ?", tid).Order("is_tool ASC, sort_order ASC").Find(&all).Error
	if err != nil {
		return nil, nil, apperr.Wrap(apperr.CodeDB, "查询材料列表失败", err)
	}
	var materials, tools []*domain.Material
	for _, m := range all {
		if m.IsTool {
			tools = append(tools, m)
		} else {
			materials = append(materials, m)
		}
	}
	return materials, tools, nil
}

func (r *MaterialRepo) GetByID(id uint64) (*domain.Material, error) {
	m := &domain.Material{}
	err := r.db.First(m, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeDB, "查询材料失败", err)
	}
	return m, nil
}
