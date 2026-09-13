package repository

import (
	"upcycle-hub/internal/domain"
	"upcycle-hub/pkg/utils"
	apperr "upcycle-hub/pkg/errors"

	"gorm.io/gorm"
)

type TagRepo struct {
	db *gorm.DB
}

func NewTagRepo(db *gorm.DB) *TagRepo {
	return &TagRepo{db: db}
}

func (r *TagRepo) Create(t *domain.Tag) error {
	if t.Slug == "" {
		t.Slug = utils.Slugify(t.Name)
	}
	err := r.db.Create(t).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "创建标签失败", err)
	}
	return nil
}

func (r *TagRepo) GetByID(id uint64) (*domain.Tag, error) {
	t := &domain.Tag{}
	err := r.db.First(t, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeDB, "查询标签失败", err)
	}
	return t, nil
}

func (r *TagRepo) GetByName(name string) (*domain.Tag, error) {
	t := &domain.Tag{}
	err := r.db.Where("name = ?", name).First(t).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, apperr.Wrap(apperr.CodeDB, "查询标签失败", err)
	}
	return t, nil
}

func (r *TagRepo) ListAll(limit int) ([]*domain.Tag, error) {
	var list []*domain.Tag
	q := r.db.Where("tutorial_count > 0").Order("tutorial_count DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&list).Error
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDB, "查询标签列表失败", err)
	}
	return list, nil
}

func (r *TagRepo) UpsertByName(names []string) ([]*domain.Tag, error) {
	tags := make([]*domain.Tag, 0, len(names))
	for _, name := range names {
		name = utils.Truncate(name, 40)
		if name == "" {
			continue
		}
		t, err := r.GetByName(name)
		if err != nil {
			return nil, err
		}
		if t == nil {
			t = &domain.Tag{Name: name, Slug: utils.Slugify(name)}
			if err := r.Create(t); err != nil {
				return nil, err
			}
		}
		tags = append(tags, t)
	}
	return tags, nil
}

func (r *TagRepo) LinkTutorial(tid uint64, tags []*domain.Tag) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tutorial_id = ?", tid).Delete(&domain.TutorialTag{}).Error; err != nil {
			return err
		}
		for _, t := range tags {
			tt := &domain.TutorialTag{TutorialID: tid, TagID: t.ID}
			if err := tx.Create(tt).Error; err != nil {
				return err
			}
			if err := tx.Model(&domain.Tag{}).Where("id = ?", t.ID).
				UpdateColumn("tutorial_count", gorm.Expr("tutorial_count + 1")).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *TagRepo) IncCount(id uint64, delta int) error {
	err := r.db.Model(&domain.Tag{}).Where("id = ?", id).
		Update("tutorial_count", gorm.Expr("tutorial_count + ?", delta)).Error
	return err
}

func (r *TagRepo) Search(keyword string, limit int) ([]*domain.Tag, error) {
	var list []*domain.Tag
	err := r.db.Where("name LIKE ?", "%"+keyword+"%").Order("tutorial_count DESC").Limit(limit).Find(&list).Error
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDB, "搜索标签失败", err)
	}
	return list, nil
}
