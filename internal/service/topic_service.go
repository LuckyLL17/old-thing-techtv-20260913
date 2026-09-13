package service

import (
	"strings"
	"time"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
)

type TopicService struct {
	topicRepo *repository.TopicRepo
}

func NewTopicService(tr *repository.TopicRepo) *TopicService {
	return &TopicService{topicRepo: tr}
}

type TopicSaveReq struct {
	Title   string
	Summary string
	Cover   string
	Status  *int // nil 表示不修改
}

func (s *TopicService) Create(operatorID uint64, r *TopicSaveReq) (*domain.Topic, error) {
	title := strings.TrimSpace(r.Title)
	if title == "" {
		return nil, ErrValidation("专题标题必填")
	}
	status := domain.TopicStatusOnline
	if r.Status != nil && *r.Status == domain.TopicStatusOffline {
		status = domain.TopicStatusOffline
	}
	t := &domain.Topic{
		Title:     title,
		Summary:   strings.TrimSpace(r.Summary),
		Cover:     strings.TrimSpace(r.Cover),
		Status:    status,
		CreatedBy: operatorID,
	}
	if err := s.topicRepo.Create(t); err != nil {
		return nil, err
	}
	return t, nil
}

// TopicUpdateReq 部分更新：nil 字段保持不变，避免仅切换状态时清空其他字段
type TopicUpdateReq struct {
	Title   *string
	Summary *string
	Cover   *string
	Status  *int
}

func (s *TopicService) Update(id uint64, r *TopicUpdateReq) (*domain.Topic, error) {
	t, err := s.topicRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if r.Title != nil {
		title := strings.TrimSpace(*r.Title)
		if title == "" {
			return nil, ErrValidation("专题标题不能为空")
		}
		t.Title = title
	}
	if r.Summary != nil {
		t.Summary = strings.TrimSpace(*r.Summary)
	}
	if r.Cover != nil {
		t.Cover = strings.TrimSpace(*r.Cover)
	}
	if r.Status != nil {
		if *r.Status != domain.TopicStatusOnline && *r.Status != domain.TopicStatusOffline {
			return nil, ErrValidation("状态值非法")
		}
		t.Status = *r.Status
	}
	if err := s.topicRepo.Update(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TopicService) Delete(id uint64) error {
	if _, err := s.topicRepo.GetByID(id); err != nil {
		return err
	}
	return s.topicRepo.Delete(id)
}

func (s *TopicService) Get(id uint64) (*domain.Topic, error) {
	return s.topicRepo.GetByID(id)
}

// List 前台列表：仅上线专题
func (s *TopicService) List(page, size int) ([]*domain.Topic, int64, error) {
	return s.topicRepo.List(page, size, true)
}

// AdminList 管理端列表：含已下架
func (s *TopicService) AdminList(page, size int) ([]*domain.Topic, int64, error) {
	return s.topicRepo.List(page, size, false)
}

// AddItem 添加条目；同一条目重复加入时幂等返回已有记录，只保留一次
func (s *TopicService) AddItem(topicID uint64, itemType string, itemID uint64) (*domain.TopicItem, bool, error) {
	t, err := s.topicRepo.GetByID(topicID)
	if err != nil {
		return nil, false, err
	}
	if itemType != domain.TopicItemTypeTutorial && itemType != domain.TopicItemTypeProject {
		return nil, false, ErrValidation("条目类型仅支持 tutorial / project")
	}
	if itemID == 0 {
		return nil, false, ErrValidation("条目ID必填")
	}
	// 校验目标存在
	if itemType == domain.TopicItemTypeTutorial {
		m, err := s.topicRepo.TutorialsByIDs([]uint64{itemID})
		if err != nil {
			return nil, false, err
		}
		if _, ok := m[itemID]; !ok {
			return nil, false, ErrNotFound("教程不存在")
		}
	} else {
		m, err := s.topicRepo.ProjectsByIDs([]uint64{itemID})
		if err != nil {
			return nil, false, err
		}
		if _, ok := m[itemID]; !ok {
			return nil, false, ErrNotFound("作品不存在")
		}
	}
	// 去重：已存在则直接返回
	exist, err := s.topicRepo.FindItem(topicID, itemType, itemID)
	if err != nil {
		return nil, false, err
	}
	if exist != nil {
		return exist, false, nil
	}
	max, err := s.topicRepo.MaxSortOrder(topicID)
	if err != nil {
		return nil, false, err
	}
	ti := &domain.TopicItem{
		TopicID:   t.ID,
		ItemType:  itemType,
		ItemID:    itemID,
		SortOrder: max + 1,
	}
	if err := s.topicRepo.AddItem(ti); err != nil {
		return nil, false, err
	}
	s.topicRepo.IncItemCount(topicID, 1)
	return ti, true, nil
}

func (s *TopicService) RemoveItem(topicID, itemID uint64) error {
	if _, err := s.topicRepo.GetByID(topicID); err != nil {
		return err
	}
	if err := s.topicRepo.RemoveItem(topicID, itemID); err != nil {
		return err
	}
	s.topicRepo.IncItemCount(topicID, -1)
	return nil
}

// TopicItemView 条目视图：目标被删除或下架时 Available=false，专题页仍可正常打开
type TopicItemView struct {
	ID        uint64           `json:"id"`
	ItemType  string           `json:"item_type"`
	ItemID    uint64           `json:"item_id"`
	SortOrder int              `json:"sort_order"`
	Available bool             `json:"available"`
	Tutorial  *domain.Tutorial `json:"tutorial,omitempty"`
	Project   *domain.Project  `json:"project,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
}

// Items 按添加顺序分页返回条目视图，批量装配教程/作品数据
func (s *TopicService) Items(topicID uint64, page, size int) ([]*TopicItemView, int64, error) {
	items, total, err := s.topicRepo.ListItems(topicID, page, size)
	if err != nil {
		return nil, 0, err
	}
	tutIDs := make([]uint64, 0, len(items))
	projIDs := make([]uint64, 0, len(items))
	for _, it := range items {
		if it.ItemType == domain.TopicItemTypeTutorial {
			tutIDs = append(tutIDs, it.ItemID)
		} else {
			projIDs = append(projIDs, it.ItemID)
		}
	}
	tuts, err := s.topicRepo.TutorialsByIDs(tutIDs)
	if err != nil {
		return nil, 0, err
	}
	projs, err := s.topicRepo.ProjectsByIDs(projIDs)
	if err != nil {
		return nil, 0, err
	}
	views := make([]*TopicItemView, 0, len(items))
	for _, it := range items {
		v := &TopicItemView{
			ID:        it.ID,
			ItemType:  it.ItemType,
			ItemID:    it.ItemID,
			SortOrder: it.SortOrder,
			CreatedAt: it.CreatedAt,
		}
		if it.ItemType == domain.TopicItemTypeTutorial {
			if t, ok := tuts[it.ItemID]; ok && t.Status == domain.TutorialStatusPublished {
				v.Available = true
				v.Tutorial = t
			}
		} else {
			if p, ok := projs[it.ItemID]; ok && p.Status == 1 {
				v.Available = true
				v.Project = p
			}
		}
		views = append(views, v)
	}
	return views, total, nil
}
