package repository

import (
	"upcycle-hub/internal/domain"
	apperr "upcycle-hub/pkg/errors"

	"gorm.io/gorm"
)

type ProjectRepo struct {
	db *gorm.DB
}

func NewProjectRepo(db *gorm.DB) *ProjectRepo {
	return &ProjectRepo{db: db}
}

func (r *ProjectRepo) Create(p *domain.Project) error {
	err := r.db.Create(p).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "创建作品失败", err)
	}
	return nil
}

func (r *ProjectRepo) GetByID(id uint64) (*domain.Project, error) {
	p := &domain.Project{}
	err := r.db.Preload("User").Preload("Tutorial").First(p, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.ErrProjectNotFound
		}
		return nil, apperr.Wrap(apperr.CodeDB, "查询作品失败", err)
	}
	return p, nil
}

func (r *ProjectRepo) Update(p *domain.Project) error {
	err := r.db.Save(p).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "更新作品失败", err)
	}
	return nil
}

func (r *ProjectRepo) Delete(id uint64) error {
	err := r.db.Delete(&domain.Project{}, id).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "删除作品失败", err)
	}
	return nil
}

func (r *ProjectRepo) List(page, size int, tutorialID, userID uint64, sort string) ([]*domain.Project, int64, error) {
	var total int64
	var list []*domain.Project
	q := r.db.Model(&domain.Project{}).Preload("User").Preload("Tutorial").Where("status = ?", 1)
	if tutorialID > 0 {
		q = q.Where("tutorial_id = ?", tutorialID)
	}
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	q.Count(&total)
	order := "id DESC"
	switch sort {
	case "rating":
		order = "rating DESC, like_count DESC"
	case "likes":
		order = "like_count DESC"
	}
	if size > 0 {
		q = q.Offset((page - 1) * size).Limit(size)
	}
	err := q.Order(order).Find(&list).Error
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeDB, "查询作品列表失败", err)
	}
	return list, total, nil
}

func (r *ProjectRepo) IncLike(id uint64, delta int) error {
	err := r.db.Model(&domain.Project{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + ?", delta)).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "更新点赞数失败", err)
	}
	return nil
}

func (r *ProjectRepo) IncComment(id uint64, delta int) error {
	err := r.db.Model(&domain.Project{}).Where("id = ?", id).
		UpdateColumn("comment_count", gorm.Expr("comment_count + ?", delta)).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "更新评论数失败", err)
	}
	return nil
}

func (r *ProjectRepo) Count() (int64, error) {
	var n int64
	err := r.db.Model(&domain.Project{}).Where("status = ?", 1).Count(&n).Error
	return n, err
}
