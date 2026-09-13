package service

import (
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
)

type RecommendService struct {
	tutorialRepo *repository.TutorialRepo
	tagRepo      *repository.TagRepo
	categoryRepo *repository.CategoryRepo
}

func NewRecommendService(tr *repository.TutorialRepo, tgr *repository.TagRepo, cr *repository.CategoryRepo) *RecommendService {
	return &RecommendService{tutorialRepo: tr, tagRepo: tgr, categoryRepo: cr}
}

type HomeData struct {
	HotTutorials   []*domain.Tutorial `json:"hot_tutorials"`
	NewTutorials   []*domain.Tutorial `json:"new_tutorials"`
	PopularTags    []*domain.Tag      `json:"popular_tags"`
	Categories     []*domain.Category `json:"categories"`
	RandomPick     []*domain.Tutorial `json:"random_pick"`
	MostAttempted  []*domain.Tutorial `json:"most_attempted"`
}

func (s *RecommendService) Home() (*HomeData, error) {
	hd := &HomeData{}
	var err error
	var total int64
	hd.HotTutorials, total, err = s.tutorialRepo.List(1, 8, 0, "", domain.TutorialStatusPublished, "popular", "", 0)
	if err != nil {
		return nil, err
	}
	hd.NewTutorials, _, err = s.tutorialRepo.List(1, 8, 0, "", domain.TutorialStatusPublished, "new", "", 0)
	if err != nil {
		return nil, err
	}
	hd.MostAttempted, _, err = s.tutorialRepo.List(1, 6, 0, "", domain.TutorialStatusPublished, "attempted", "", 0)
	if err != nil {
		return nil, err
	}
	hd.PopularTags, err = s.tagRepo.ListAll(30)
	if err != nil {
		return nil, err
	}
	hd.Categories, err = s.categoryRepo.ListAll()
	if err != nil {
		return nil, err
	}
	hd.RandomPick, err = s.tutorialRepo.Random(6)
	if err != nil {
		hd.RandomPick = hd.HotTutorials[:min(6, len(hd.HotTutorials))]
	}
	_ = total
	return hd, nil
}

func (s *RecommendService) TopTutorials(n int) ([]*domain.Tutorial, error) {
	return s.tutorialRepo.TopTutorials(n)
}

func (s *RecommendService) RandomInspiration(n int) ([]*domain.Tutorial, error) {
	list, err := s.tutorialRepo.Random(n)
	if err != nil || len(list) == 0 {
		l, _, e := s.tutorialRepo.List(1, n, 0, "", domain.TutorialStatusPublished, "popular", "", 0)
		return l, e
	}
	return list, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
