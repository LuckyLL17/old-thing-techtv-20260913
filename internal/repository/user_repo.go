package repository

import (
	"upcycle-hub/internal/domain"
	apperr "upcycle-hub/pkg/errors"

	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) DB() *gorm.DB {
	return r.db
}

func (r *UserRepo) Create(u *domain.User) error {
	err := r.db.Create(u).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "创建用户失败", err)
	}
	return nil
}

func (r *UserRepo) GetByID(id uint64) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.First(u, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.ErrUserNotFound
		}
		return nil, apperr.Wrap(apperr.CodeDB, "查询用户失败", err)
	}
	return u, nil
}

func (r *UserRepo) GetByEmail(email string) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.Where("email = ?", email).First(u).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.ErrUserNotFound
		}
		return nil, apperr.Wrap(apperr.CodeDB, "查询用户失败", err)
	}
	return u, nil
}

func (r *UserRepo) GetByUsername(username string) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.Where("username = ?", username).First(u).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.ErrUserNotFound
		}
		return nil, apperr.Wrap(apperr.CodeDB, "查询用户失败", err)
	}
	return u, nil
}

func (r *UserRepo) Update(u *domain.User) error {
	err := r.db.Save(u).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "更新用户失败", err)
	}
	return nil
}

func (r *UserRepo) UpdatePassword(id uint64, hash string) error {
	err := r.db.Model(&domain.User{}).Where("id = ?", id).Update("password_hash", hash).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "更新密码失败", err)
	}
	return nil
}

func (r *UserRepo) List(page, size int, keyword string) ([]*domain.User, int64, error) {
	var total int64
	var list []*domain.User
	q := r.db.Model(&domain.User{})
	if keyword != "" {
		q = q.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	q.Count(&total)
	if size > 0 {
		q = q.Offset((page - 1) * size).Limit(size)
	}
	err := q.Order("id DESC").Find(&list).Error
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeDB, "查询用户列表失败", err)
	}
	return list, total, nil
}

func (r *UserRepo) IncStats(id uint64, tutorialDelta, projectDelta, scoreDelta int) error {
	err := r.db.Model(&domain.User{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"tutorial_count": gorm.Expr("tutorial_count + ?", tutorialDelta),
			"project_count":  gorm.Expr("project_count + ?", projectDelta),
			"score":          gorm.Expr("score + ?", scoreDelta),
		}).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "更新用户统计失败", err)
	}
	return nil
}

func (r *UserRepo) TopUsers(n int) ([]*domain.User, error) {
	var list []*domain.User
	err := r.db.Order("score DESC").Limit(n).Find(&list).Error
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDB, "查询活跃用户失败", err)
	}
	return list, nil
}

func (r *UserRepo) Count() (int64, error) {
	var n int64
	err := r.db.Model(&domain.User{}).Count(&n).Error
	return n, err
}
