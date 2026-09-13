package repository

import (
	"upcycle-hub/internal/domain"
	apperr "upcycle-hub/pkg/errors"

	"gorm.io/gorm"
)

type TopicRepo struct {
	db *gorm.DB
}

func NewTopicRepo(db *gorm.DB) *TopicRepo {
	return &TopicRepo{db: db}
}

func (r *TopicRepo) Create(t *domain.Topic) error {
	err := r.db.Create(t).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "创建专题失败", err)
	}
	return nil
}

func (r *TopicRepo) GetByID(id uint64) (*domain.Topic, error) {
	t := &domain.Topic{}
	err := r.db.First(t, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.CodeDB, "查询专题失败", err)
	}
	return t, nil
}

func (r *TopicRepo) Update(t *domain.Topic) error {
	err := r.db.Save(t).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "更新专题失败", err)
	}
	return nil
}

// Delete 删除专题及其全部条目（事务）
func (r *TopicRepo) Delete(id uint64) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("topic_id = ?", id).Delete(&domain.TopicItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.Topic{}, id).Error
	})
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "删除专题失败", err)
	}
	return nil
}

// List 专题列表；onlineOnly 为 true 时只返回上线专题
func (r *TopicRepo) List(page, size int, onlineOnly bool) ([]*domain.Topic, int64, error) {
	var total int64
	var list []*domain.Topic
	q := r.db.Model(&domain.Topic{})
	if onlineOnly {
		q = q.Where("status = ?", domain.TopicStatusOnline)
	}
	q.Count(&total)
	if size > 0 {
		q = q.Offset((page - 1) * size).Limit(size)
	}
	err := q.Order("id DESC").Find(&list).Error
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeDB, "查询专题列表失败", err)
	}
	return list, total, nil
}

// FindItem 按专题+类型+目标查找条目，未找到返回 (nil, nil)
func (r *TopicRepo) FindItem(topicID uint64, itemType string, itemID uint64) (*domain.TopicItem, error) {
	ti := &domain.TopicItem{}
	err := r.db.Where("topic_id = ? AND item_type = ? AND item_id = ?", topicID, itemType, itemID).First(ti).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, apperr.Wrap(apperr.CodeDB, "查询专题条目失败", err)
	}
	return ti, nil
}

func (r *TopicRepo) MaxSortOrder(topicID uint64) (int, error) {
	var max *int
	err := r.db.Model(&domain.TopicItem{}).Where("topic_id = ?", topicID).
		Select("MAX(sort_order)").Scan(&max).Error
	if err != nil {
		return 0, apperr.Wrap(apperr.CodeDB, "查询条目排序失败", err)
	}
	if max == nil {
		return 0, nil
	}
	return *max, nil
}

func (r *TopicRepo) AddItem(ti *domain.TopicItem) error {
	err := r.db.Create(ti).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "添加专题条目失败", err)
	}
	return nil
}

func (r *TopicRepo) RemoveItem(topicID, itemID uint64) error {
	err := r.db.Where("id = ? AND topic_id = ?", itemID, topicID).Delete(&domain.TopicItem{}).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "移除专题条目失败", err)
	}
	return nil
}

// ListItems 按添加顺序（sort_order 升序）分页返回条目
func (r *TopicRepo) ListItems(topicID uint64, page, size int) ([]*domain.TopicItem, int64, error) {
	var total int64
	var list []*domain.TopicItem
	q := r.db.Model(&domain.TopicItem{}).Where("topic_id = ?", topicID)
	q.Count(&total)
	if size > 0 {
		q = q.Offset((page - 1) * size).Limit(size)
	}
	err := q.Order("sort_order ASC, id ASC").Find(&list).Error
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.CodeDB, "查询专题条目失败", err)
	}
	return list, total, nil
}

func (r *TopicRepo) IncItemCount(topicID uint64, delta int) error {
	err := r.db.Model(&domain.Topic{}).Where("id = ?", topicID).
		Update("item_count", gorm.Expr("item_count + ?", delta)).Error
	if err != nil {
		return apperr.Wrap(apperr.CodeDB, "更新专题条目数失败", err)
	}
	return nil
}

// TutorialsByIDs 批量取教程（含已下架，便于条目标记可用性）
func (r *TopicRepo) TutorialsByIDs(ids []uint64) (map[uint64]*domain.Tutorial, error) {
	m := make(map[uint64]*domain.Tutorial, len(ids))
	if len(ids) == 0 {
		return m, nil
	}
	var list []*domain.Tutorial
	err := r.db.Preload("User").Preload("Category").Preload("Tags").Where("id IN ?", ids).Find(&list).Error
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDB, "查询教程失败", err)
	}
	for _, t := range list {
		m[t.ID] = t
	}
	return m, nil
}

// ProjectsByIDs 批量取作品（含已下架）
func (r *TopicRepo) ProjectsByIDs(ids []uint64) (map[uint64]*domain.Project, error) {
	m := make(map[uint64]*domain.Project, len(ids))
	if len(ids) == 0 {
		return m, nil
	}
	var list []*domain.Project
	err := r.db.Preload("User").Preload("Tutorial").Where("id IN ?", ids).Find(&list).Error
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDB, "查询作品失败", err)
	}
	for _, p := range list {
		m[p.ID] = p
	}
	return m, nil
}
