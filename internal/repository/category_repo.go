package repository

import (
	"upcycle-hub/internal/domain"
	apperr "upcycle-hub/pkg/errors"

	"gorm.io/gorm"
)

type CategoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepo(db *gorm.DB) *CategoryRepo {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) Create(c *domain.Category) error {
	err := r.db.Create(c).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "创建分类失败", err)
	}
	return nil
}

func (r *CategoryRepo) GetByID(id uint64) (*domain.Category, error) {
	c := &domain.Category{}
	err := r.db.First(c, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeDB, "查询分类失败", err)
	}
	return c, nil
}

func (r *CategoryRepo) GetByCode(code string) (*domain.Category, error) {
	c := &domain.Category{}
	err := r.db.Where("code = ?", code).First(c).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeDB, "查询分类失败", err)
	}
	return c, nil
}

func (r *CategoryRepo) ListAll() ([]*domain.Category, error) {
	var list []*domain.Category
	err := r.db.Where("status = ?", 1).Order("sort_order ASC, id ASC").Find(&list).Error
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDB, "查询分类列表失败", err)
	}
	return list, nil
}

func (r *CategoryRepo) Update(c *domain.Category) error {
	err := r.db.Save(c).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "更新分类失败", err)
	}
	return nil
}

func (r *CategoryRepo) IncCount(id uint64, delta int) error {
	err := r.db.Model(&domain.Category{}).Where("id = ?", id).
		Update("tutorial_count", gorm.Expr("tutorial_count + ?", delta)).Error
	return err
}

func (r *CategoryRepo) InitDefaults() error {
	for _, c := range domain.DefaultCategories {
		var cnt int64
		r.db.Model(&domain.Category{}).Where("code = ?", c.Code).Count(&cnt)
		if cnt == 0 {
			if err := r.db.Create(&c).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
