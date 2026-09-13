package repository

import (
	"upcycle-hub/internal/domain"
	apperr "upcycle-hub/pkg/errors"

	"gorm.io/gorm"
)

type FavoriteRepo struct {
	db *gorm.DB
}

func NewFavoriteRepo(db *gorm.DB) *FavoriteRepo {
	return &FavoriteRepo{db: db}
}

func (r *FavoriteRepo) Create(f *domain.Favorite) error {
	err := r.db.Create(f).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "创建收藏失败", err)
	}
	return nil
}

func (r *FavoriteRepo) Delete(userID uint64, targetType string, targetID uint64) error {
	err := r.db.Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		Delete(&domain.Favorite{}).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "删除收藏失败", err)
	}
	return nil
}

func (r *FavoriteRepo) Exists(userID uint64, targetType string, targetID uint64) (bool, error) {
	var n int64
	err := r.db.Model(&domain.Favorite{}).
		Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		Count(&n).Error
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *FavoriteRepo) ListByUser(userID uint64, targetType string, page, size int) ([]*domain.Favorite, int64, error) {
	var total int64
	var list []*domain.Favorite
	q := r.db.Model(&domain.Favorite{}).Where("user_id = ?", userID)
	if targetType != "" {
		q = q.Where("target_type = ?", targetType)
	}
	q.Count(&total)
	if size > 0 {
		q = q.Offset((page - 1) * size).Limit(size)
	}
	err := q.Order("id DESC").Find(&list).Error
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeDB, "查询收藏列表失败", err)
	}
	return list, total, nil
}

func (r *FavoriteRepo) CountByUser(userID uint64) (map[string]int64, error) {
	rows, err := r.db.Model(&domain.Favorite{}).
		Select("target_type, count(*)").Where("user_id = ?", userID).
		Group("target_type").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[string]int64)
	for rows.Next() {
		var t string
		var n int64
		rows.Scan(&t, &n)
		m[t] = n
	}
	return m, nil
}
