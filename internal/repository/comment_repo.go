package repository

import (
	"upcycle-hub/internal/domain"
	apperr "upcycle-hub/pkg/errors"

	"gorm.io/gorm"
)

type CommentRepo struct {
	db *gorm.DB
}

func NewCommentRepo(db *gorm.DB) *CommentRepo {
	return &CommentRepo{db: db}
}

func (r *CommentRepo) Create(c *domain.Comment) error {
	err := r.db.Create(c).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "创建评论失败", err)
	}
	return nil
}

func (r *CommentRepo) GetByID(id uint64) (*domain.Comment, error) {
	c := &domain.Comment{}
	err := r.db.First(c, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeDB, "查询评论失败", err)
	}
	return c, nil
}

func (r *CommentRepo) Delete(id uint64) error {
	err := r.db.Delete(&domain.Comment{}, id).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "删除评论失败", err)
	}
	return nil
}

func (r *CommentRepo) List(targetType string, targetID uint64, page, size int) ([]*domain.Comment, int64, error) {
	var total int64
	var list []*domain.Comment
	q := r.db.Model(&domain.Comment{}).Preload("User").
		Where("target_type = ? AND target_id = ? AND parent_id = 0", targetType, targetID).
		Where("status = ?", 1)
	q.Count(&total)
	if size > 0 {
		q = q.Offset((page - 1) * size).Limit(size)
	}
	err := q.Order("id DESC").Find(&list).Error
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeDB, "查询评论列表失败", err)
	}
	return list, total, nil
}

func (r *CommentRepo) IncLike(id uint64, delta int) error {
	err := r.db.Model(&domain.Comment{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + ?", delta)).Error
	return err
}

func (r *CommentRepo) Count(targetType string, targetID uint64) (int64, error) {
	var n int64
	err := r.db.Model(&domain.Comment{}).
		Where("target_type = ? AND target_id = ? AND status = ?", targetType, targetID, 1).
		Count(&n).Error
	return n, err
}
