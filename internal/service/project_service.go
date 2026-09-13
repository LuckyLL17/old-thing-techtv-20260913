package service

import (
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
)

type ProjectService struct {
	projectRepo  *repository.ProjectRepo
	tutorialRepo *repository.TutorialRepo
	userRepo     *repository.UserRepo
}

func NewProjectService(pr *repository.ProjectRepo, tr *repository.TutorialRepo, ur *repository.UserRepo) *ProjectService {
	return &ProjectService{projectRepo: pr, tutorialRepo: tr, userRepo: ur}
}

type ProjectCreateReq struct {
	UserID      uint64
	TutorialID  uint64
	Title       string
	Description string
	Images      string
	CustomNotes string
	Rating      int
}

func (s *ProjectService) Create(r *ProjectCreateReq) (*domain.Project, error) {
	if r.TutorialID == 0 {
		return nil, ErrValidation("必须关联教程")
	}
	t, err := s.tutorialRepo.GetByID(r.TutorialID, false)
	if err != nil {
		return nil, err
	}
	if t.Status != domain.TutorialStatusPublished {
		return nil, ErrValidation("关联的教程尚未发布")
	}
	if r.Rating < 0 || r.Rating > 5 {
		r.Rating = 0
	}
	p := &domain.Project{
		UserID:      r.UserID,
		TutorialID:  r.TutorialID,
		Title:       r.Title,
		Description: r.Description,
		Images:      r.Images,
		CustomNotes: r.CustomNotes,
		Rating:      r.Rating,
		Status:      1,
	}
	if err := s.projectRepo.Create(p); err != nil {
		return nil, err
	}
	s.tutorialRepo.IncCounts(t.ID, 0, 0, 0, 1)
	s.userRepo.IncStats(r.UserID, 0, 1, 20+r.Rating*10)
	return s.projectRepo.GetByID(p.ID)
}

func (s *ProjectService) Update(id, userID uint64, r *ProjectCreateReq) (*domain.Project, error) {
	p, err := s.projectRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if p.UserID != userID {
		return nil, ErrForbidden("无权修改此作品")
	}
	if r.Title != "" {
		p.Title = r.Title
	}
	if r.Description != "" {
		p.Description = r.Description
	}
	if r.Images != "" {
		p.Images = r.Images
	}
	if r.CustomNotes != "" {
		p.CustomNotes = r.CustomNotes
	}
	if r.Rating >= 0 && r.Rating <= 5 {
		p.Rating = r.Rating
	}
	if err := s.projectRepo.Update(p); err != nil {
		return nil, err
	}
	return s.projectRepo.GetByID(id)
}

func (s *ProjectService) Delete(id, userID uint64) error {
	p, err := s.projectRepo.GetByID(id)
	if err != nil {
		return err
	}
	if p.UserID != userID {
		return ErrForbidden("无权删除此作品")
	}
	s.tutorialRepo.IncCounts(p.TutorialID, 0, 0, 0, -1)
	s.userRepo.IncStats(userID, 0, -1, -(20 + p.Rating*10))
	return s.projectRepo.Delete(id)
}

func (s *ProjectService) Get(id uint64) (*domain.Project, error) {
	return s.projectRepo.GetByID(id)
}

func (s *ProjectService) List(page, size int, tutorialID, userID uint64, sort string) ([]*domain.Project, int64, error) {
	return s.projectRepo.List(page, size, tutorialID, userID, sort)
}

func (s *ProjectService) Like(id uint64) error {
	return s.projectRepo.IncLike(id, 1)
}
