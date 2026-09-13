package service

import (
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
)

type CategoryService struct {
	categoryRepo *repository.CategoryRepo
}

func NewCategoryService(cr *repository.CategoryRepo) *CategoryService {
	return &CategoryService{categoryRepo: cr}
}

func (s *CategoryService) ListAll() ([]*domain.Category, error) {
	return s.categoryRepo.ListAll()
}

func (s *CategoryService) Get(id uint64) (*domain.Category, error) {
	return s.categoryRepo.GetByID(id)
}

func (s *CategoryService) GetByCode(code string) (*domain.Category, error) {
	return s.categoryRepo.GetByCode(code)
}

func (s *CategoryService) Create(c *domain.Category) error {
	if c.Name == "" || c.Code == "" {
		return ErrValidation("名称和编码必填")
	}
	return s.categoryRepo.Create(c)
}

func (s *CategoryService) Update(c *domain.Category) error {
	return s.categoryRepo.Update(c)
}

func (s *CategoryService) InitDefaults() error {
	return s.categoryRepo.InitDefaults()
}

type TagService struct {
	tagRepo *repository.TagRepo
}

func NewTagService(tr *repository.TagRepo) *TagService {
	return &TagService{tagRepo: tr}
}

func (s *TagService) ListPopular(limit int) ([]*domain.Tag, error) {
	return s.tagRepo.ListAll(limit)
}

func (s *TagService) Search(keyword string, limit int) ([]*domain.Tag, error) {
	return s.tagRepo.Search(keyword, limit)
}

func (s *TagService) Upsert(names []string) ([]*domain.Tag, error) {
	return s.tagRepo.UpsertByName(names)
}
