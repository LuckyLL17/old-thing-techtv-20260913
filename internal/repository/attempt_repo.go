package repository

import (
	"upcycle-hub/internal/domain"
	apperr "upcycle-hub/pkg/errors"

	"gorm.io/gorm"
)

type AttemptRepo struct {
	db *gorm.DB
}

func NewAttemptRepo(db *gorm.DB) *AttemptRepo {
	return &AttemptRepo{db: db}
}

func (r *AttemptRepo) Create(a *domain.Attempt) error {
	err := r.db.Create(a).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "创建尝试记录失败", err)
	}
	return nil
}

func (r *AttemptRepo) Update(a *domain.Attempt) error {
	err := r.db.Save(a).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "更新尝试记录失败", err)
	}
	return nil
}

func (r *AttemptRepo) GetByUserAndTutorial(userID, tutorialID uint64) (*domain.Attempt, error) {
	a := &domain.Attempt{}
	err := r.db.Where("user_id = ? AND tutorial_id = ?", userID, tutorialID).First(a).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, apperr.Wrap(apperr.CodeDB, "查询尝试记录失败", err)
	}
	return a, nil
}

func (r *AttemptRepo) ListByUser(userID uint64, page, size int) ([]*domain.Attempt, int64, error) {
	var total int64
	var list []*domain.Attempt
	q := r.db.Model(&domain.Attempt{}).Where("user_id = ?", userID)
	q.Count(&total)
	if size > 0 {
		q = q.Offset((page - 1) * size).Limit(size)
	}
	err := q.Order("id DESC").Find(&list).Error
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeDB, "查询尝试记录失败", err)
	}
	return list, total, nil
}

func (r *AttemptRepo) CountByUser(userID uint64) (int64, int64, error) {
	var total, done int64
	err := r.db.Model(&domain.Attempt{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return 0, 0, err
	}
	err = r.db.Model(&domain.Attempt{}).Where("user_id = ? AND completed = ?", userID, true).Count(&done).Error
	if err != nil {
		return 0, 0, err
	}
	return total, done, nil
}
