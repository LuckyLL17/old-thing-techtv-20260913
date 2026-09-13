package service

import (
	"strings"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
)

type SearchService struct {
	tutorialRepo *repository.TutorialRepo
	tagRepo      *repository.TagRepo
	categoryRepo *repository.CategoryRepo
}

func NewSearchService(tr *repository.TutorialRepo, tgr *repository.TagRepo, cr *repository.CategoryRepo) *SearchService {
	return &SearchService{tutorialRepo: tr, tagRepo: tgr, categoryRepo: cr}
}

type SearchResult struct {
	Tutorials     []*domain.Tutorial `json:"tutorials"`
	Categories    []*domain.Category `json:"categories"`
	Tags          []*domain.Tag      `json:"tags"`
	TutorialTotal int64              `json:"tutorial_total"`
	Keyword       string             `json:"keyword"`
}

func (s *SearchService) All(keyword string, page, size int) (*SearchResult, error) {
	keyword = strings.TrimSpace(keyword)
	res := &SearchResult{Keyword: keyword}
	if keyword == "" {
		return res, nil
	}
	list, total, err := s.tutorialRepo.List(page, size, 0, "", domain.TutorialStatusPublished, "new", keyword, 0)
	if err != nil {
		return nil, err
	}
	res.Tutorials = list
	res.TutorialTotal = total
	tags, err := s.tagRepo.Search(keyword, 10)
	if err == nil {
		res.Tags = tags
	}
	cats, err := s.categoryRepo.ListAll()
	if err == nil {
		match := make([]*domain.Category, 0)
		kw := strings.ToLower(keyword)
		for _, c := range cats {
			if strings.Contains(strings.ToLower(c.Name), kw) || strings.Contains(strings.ToLower(c.Description), kw) {
				match = append(match, c)
			}
		}
		res.Categories = match
	}
	return res, nil
}

func (s *SearchService) Tutorials(keyword string, catID uint64, diff string, page, size int) ([]*domain.Tutorial, int64, error) {
	return s.tutorialRepo.List(page, size, catID, diff, domain.TutorialStatusPublished, "new", keyword, 0)
}
