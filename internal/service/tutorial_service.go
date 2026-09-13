package service

import (
	"encoding/json"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
	"upcycle-hub/pkg/utils"
)

type TutorialService struct {
	tutorialRepo *repository.TutorialRepo
	stepRepo     *repository.StepRepo
	materialRepo *repository.MaterialRepo
	tagRepo      *repository.TagRepo
	categoryRepo *repository.CategoryRepo
	userRepo     *repository.UserRepo
}

func NewTutorialService(tur *repository.TutorialRepo, sr *repository.StepRepo, mr *repository.MaterialRepo,
	tr *repository.TagRepo, cr *repository.CategoryRepo, ur *repository.UserRepo) *TutorialService {
	return &TutorialService{tutorialRepo: tur, stepRepo: sr, materialRepo: mr, tagRepo: tr, categoryRepo: cr, userRepo: ur}
}

type TutorialCreateReq struct {
	UserID         uint64
	CategoryID     uint64
	Title          string
	Summary        string
	CoverBefore    string
	CoverAfter     string
	Difficulty     string
	EstimatedHours float64
	Status         string
	TagNames       []string
	Steps          []*domain.Step
	Materials      []*domain.Material
	Tools          []*domain.Material
}

func (s *TutorialService) Create(r *TutorialCreateReq) (*domain.Tutorial, error) {
	if r.Title == "" || r.CoverBefore == "" || r.CoverAfter == "" {
		return nil, ErrValidation("标题和改造前后图为必填项")
	}
	if _, err := s.categoryRepo.GetByID(r.CategoryID); err != nil {
		return nil, ErrValidation("分类不存在")
	}
	t := &domain.Tutorial{
		UserID:         r.UserID,
		CategoryID:     r.CategoryID,
		Title:          r.Title,
		Slug:           utils.Slugify(r.Title),
		Summary:        r.Summary,
		CoverBefore:    r.CoverBefore,
		CoverAfter:     r.CoverAfter,
		Difficulty:     normalizeDifficulty(r.Difficulty),
		EstimatedHours: r.EstimatedHours,
		Status:         normalizeStatus(r.Status),
		Version:        1,
	}
	if err := s.tutorialRepo.Create(t); err != nil {
		return nil, err
	}
	for i, st := range r.Steps {
		st.TutorialID = t.ID
		st.StepOrder = i + 1
	}
	if err := s.stepRepo.BatchCreate(r.Steps); err != nil {
		return nil, err
	}
	for i, m := range r.Materials {
		m.TutorialID = t.ID
		m.SortOrder = i + 1
		m.IsTool = false
	}
	for i, m := range r.Tools {
		m.TutorialID = t.ID
		m.SortOrder = i + 1
		m.IsTool = true
		r.Materials = append(r.Materials, m)
	}
	if err := s.materialRepo.BatchCreate(r.Materials); err != nil {
		return nil, err
	}
	tags, err := s.tagRepo.UpsertByName(r.TagNames)
	if err != nil {
		return nil, err
	}
	if err := s.tagRepo.LinkTutorial(t.ID, tags); err != nil {
		return nil, err
	}
	if t.Status == domain.TutorialStatusPublished {
		s.categoryRepo.IncCount(t.CategoryID, 1)
		s.userRepo.IncStats(t.UserID, 1, 0, 50)
	}
	return s.tutorialRepo.GetByID(t.ID, true)
}

func (s *TutorialService) Update(id, userID uint64, r *TutorialCreateReq) (*domain.Tutorial, error) {
	t, err := s.tutorialRepo.GetByID(id, false)
	if err != nil {
		return nil, err
	}
	if t.UserID != userID {
		return nil, ErrForbidden("无权修改此教程")
	}
	prev := &domain.TutorialVersion{
		TutorialID: t.ID,
		Version:    t.Version,
		Title:      t.Title,
		Summary:    t.Summary,
	}
	dump, _ := json.Marshal(t)
	prev.ContentDump = string(dump)
	s.tutorialRepo.SaveVersion(prev)
	oldCat := t.CategoryID
	oldStatus := t.Status
	if r.Title != "" {
		t.Title = r.Title
		t.Slug = utils.Slugify(r.Title)
	}
	if r.Summary != "" {
		t.Summary = r.Summary
	}
	if r.CoverBefore != "" {
		t.CoverBefore = r.CoverBefore
	}
	if r.CoverAfter != "" {
		t.CoverAfter = r.CoverAfter
	}
	if r.Difficulty != "" {
		t.Difficulty = normalizeDifficulty(r.Difficulty)
	}
	if r.EstimatedHours > 0 {
		t.EstimatedHours = r.EstimatedHours
	}
	if r.Status != "" {
		t.Status = normalizeStatus(r.Status)
	}
	if r.CategoryID > 0 {
		t.CategoryID = r.CategoryID
	}
	t.Version++
	if err := s.tutorialRepo.Update(t); err != nil {
		return nil, err
	}
	if len(r.Steps) > 0 {
		s.stepRepo.DeleteByTutorial(id)
		for i, st := range r.Steps {
			st.ID = 0
			st.TutorialID = id
			st.StepOrder = i + 1
		}
		s.stepRepo.BatchCreate(r.Steps)
	}
	if len(r.Materials) > 0 || len(r.Tools) > 0 {
		s.materialRepo.DeleteByTutorial(id)
		all := make([]*domain.Material, 0, len(r.Materials)+len(r.Tools))
		for i, m := range r.Materials {
			m.ID = 0
			m.TutorialID = id
			m.SortOrder = i + 1
			m.IsTool = false
			all = append(all, m)
		}
		for i, m := range r.Tools {
			m.ID = 0
			m.TutorialID = id
			m.SortOrder = i + 1
			m.IsTool = true
			all = append(all, m)
		}
		s.materialRepo.BatchCreate(all)
	}
	if len(r.TagNames) > 0 {
		tags, _ := s.tagRepo.UpsertByName(r.TagNames)
		s.tagRepo.LinkTutorial(id, tags)
	}
	if oldStatus != domain.TutorialStatusPublished && t.Status == domain.TutorialStatusPublished {
		s.categoryRepo.IncCount(t.CategoryID, 1)
		s.userRepo.IncStats(t.UserID, 1, 0, 50)
	} else if oldStatus == domain.TutorialStatusPublished && t.Status != domain.TutorialStatusPublished {
		s.categoryRepo.IncCount(oldCat, -1)
		s.userRepo.IncStats(t.UserID, -1, 0, -50)
	}
	return s.tutorialRepo.GetByID(id, true)
}

func (s *TutorialService) Delete(id, userID uint64) error {
	t, err := s.tutorialRepo.GetByID(id, false)
	if err != nil {
		return err
	}
	if t.UserID != userID {
		return ErrForbidden("无权删除此教程")
	}
	if t.Status == domain.TutorialStatusPublished {
		s.categoryRepo.IncCount(t.CategoryID, -1)
		s.userRepo.IncStats(t.UserID, -1, 0, -50)
	}
	return s.tutorialRepo.Delete(id)
}

func (s *TutorialService) Get(id uint64, incView bool) (*domain.Tutorial, error) {
	if incView {
		s.tutorialRepo.IncView(id)
	}
	return s.tutorialRepo.GetByID(id, true)
}

func (s *TutorialService) List(page, size int, catID uint64, diff, status, sort, keyword string, userID uint64) ([]*domain.Tutorial, int64, error) {
	return s.tutorialRepo.List(page, size, catID, diff, status, sort, keyword, userID)
}

func (s *TutorialService) ReorderSteps(id, userID uint64, order []uint64) error {
	t, err := s.tutorialRepo.GetByID(id, false)
	if err != nil {
		return err
	}
	if t.UserID != userID {
		return ErrForbidden("无权修改此教程")
	}
	return s.stepRepo.UpdateOrder(id, order)
}

func (s *TutorialService) Attempt(userID, tutorialID uint64) error {
	a, err := (&repository.AttemptRepo{}).GetByUserAndTutorial(userID, tutorialID)
	if err != nil {
		return err
	}
	if a != nil {
		return nil
	}
	a = &domain.Attempt{UserID: userID, TutorialID: tutorialID}
	(&repository.AttemptRepo{}).Create(a)
	s.tutorialRepo.IncCounts(tutorialID, 0, 1, 0, 0)
	return nil
}

func normalizeDifficulty(d string) string {
	switch d {
	case domain.DifficultyEasy, domain.DifficultyMedium, domain.DifficultyHard:
		return d
	}
	return domain.DifficultyMedium
}

func normalizeStatus(s string) string {
	switch s {
	case domain.TutorialStatusDraft, domain.TutorialStatusPublished, domain.TutorialStatusArchived:
		return s
	}
	return domain.TutorialStatusDraft
}
