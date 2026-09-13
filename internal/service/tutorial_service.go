package service

import (
	"encoding/json"
	"time"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
	apperr "upcycle-hub/pkg/errors"
	"upcycle-hub/pkg/logger"
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
	ScheduledAt    *time.Time
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
	status := normalizeStatus(r.Status)
	var scheduledAt *time.Time
	if r.ScheduledAt != nil {
		if !r.ScheduledAt.After(time.Now()) {
			return nil, ErrValidation("定时发布时间必须晚于当前时间")
		}
		status = domain.TutorialStatusScheduled
		scheduledAt = utcTime(r.ScheduledAt)
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
		Status:         status,
		ScheduledAt:    scheduledAt,
		Version:        1,
	}
	if status == domain.TutorialStatusPublished {
		now := time.Now()
		t.PublishedAt = &now
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
	if r.ScheduledAt != nil {
		if !r.ScheduledAt.After(time.Now()) {
			return nil, ErrValidation("定时发布时间必须晚于当前时间")
		}
		if oldStatus == domain.TutorialStatusPublished {
			return nil, ErrValidation("已发布的教程不能改为定时发布")
		}
		t.Status = domain.TutorialStatusScheduled
		t.ScheduledAt = utcTime(r.ScheduledAt)
	} else if r.Status != "" {
		t.Status = normalizeStatus(r.Status)
		// 显式切换为其他状态时，撤销未生效的定时
		t.ScheduledAt = nil
	}
	if r.CategoryID > 0 {
		t.CategoryID = r.CategoryID
	}
	becamePublished := oldStatus != domain.TutorialStatusPublished && t.Status == domain.TutorialStatusPublished
	if becamePublished {
		now := time.Now()
		t.PublishedAt = &now
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
	if becamePublished {
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

func (s *TutorialService) Get(id, viewerID uint64, incView bool) (*domain.Tutorial, error) {
	t, err := s.tutorialRepo.GetByID(id, true)
	if err != nil {
		return nil, err
	}
	// 未发布内容（草稿/定时中/已归档）仅作者本人可见
	if t.Status != domain.TutorialStatusPublished && t.UserID != viewerID {
		return nil, apperr.ErrTutorialNotFound
	}
	if incView {
		s.tutorialRepo.IncView(id)
	}
	return t, nil
}

// Schedule 设置或修改定时发布时间
func (s *TutorialService) Schedule(id, userID uint64, at time.Time) (*domain.Tutorial, error) {
	if !at.After(time.Now()) {
		return nil, ErrValidation("定时发布时间必须晚于当前时间")
	}
	t, err := s.tutorialRepo.GetByID(id, false)
	if err != nil {
		return nil, err
	}
	if t.UserID != userID {
		return nil, ErrForbidden("无权修改此教程")
	}
	if t.Status == domain.TutorialStatusPublished {
		return nil, ErrValidation("已发布的教程不能设置定时")
	}
	t.Status = domain.TutorialStatusScheduled
	t.ScheduledAt = utcTime(&at)
	if err := s.tutorialRepo.Update(t); err != nil {
		return nil, err
	}
	return s.tutorialRepo.GetByID(id, true)
}

// CancelSchedule 取消定时发布，教程回到草稿状态
func (s *TutorialService) CancelSchedule(id, userID uint64) (*domain.Tutorial, error) {
	t, err := s.tutorialRepo.GetByID(id, false)
	if err != nil {
		return nil, err
	}
	if t.UserID != userID {
		return nil, ErrForbidden("无权修改此教程")
	}
	if t.Status != domain.TutorialStatusScheduled {
		return nil, ErrValidation("该教程不在定时发布状态")
	}
	t.Status = domain.TutorialStatusDraft
	t.ScheduledAt = nil
	if err := s.tutorialRepo.Update(t); err != nil {
		return nil, err
	}
	return s.tutorialRepo.GetByID(id, true)
}

// PublishNow 立即发布（草稿或定时中的教程）
func (s *TutorialService) PublishNow(id, userID uint64) (*domain.Tutorial, error) {
	t, err := s.tutorialRepo.GetByID(id, false)
	if err != nil {
		return nil, err
	}
	if t.UserID != userID {
		return nil, ErrForbidden("无权修改此教程")
	}
	if t.Status == domain.TutorialStatusPublished {
		return nil, ErrValidation("教程已发布")
	}
	t.Status = domain.TutorialStatusPublished
	t.ScheduledAt = nil
	now := time.Now()
	t.PublishedAt = &now
	if err := s.tutorialRepo.Update(t); err != nil {
		return nil, err
	}
	s.categoryRepo.IncCount(t.CategoryID, 1)
	s.userRepo.IncStats(t.UserID, 1, 0, 50)
	return s.tutorialRepo.GetByID(id, true)
}

// PublishDue 发布所有到期的定时教程，由调度器周期调用。
// 每条教程的发布结果都会写日志，便于排查未发出的情况。
func (s *TutorialService) PublishDue() {
	now := time.Now().UTC()
	list, err := s.tutorialRepo.ListDueScheduled(now)
	if err != nil {
		logger.Errorf("定时发布：查询到期任务失败: %v", err)
		return
	}
	for _, t := range list {
		ok, err := s.tutorialRepo.PublishIfScheduled(t.ID, now)
		if err != nil {
			logger.Errorf("定时发布失败 tutorial_id=%d title=%q: %v", t.ID, t.Title, err)
			continue
		}
		if !ok {
			// 状态已被并发修改（如作者取消定时或手动发布），跳过
			continue
		}
		s.categoryRepo.IncCount(t.CategoryID, 1)
		s.userRepo.IncStats(t.UserID, 1, 0, 50)
		scheduledAt := ""
		if t.ScheduledAt != nil {
			scheduledAt = t.ScheduledAt.Format(time.RFC3339)
		}
		logger.Infof("定时发布成功 tutorial_id=%d title=%q scheduled_at=%s", t.ID, t.Title, scheduledAt)
	}
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

// utcTime 统一转为 UTC 存储，保证 SQLite 字符串形式的时间比较不受时区影响
func utcTime(t *time.Time) *time.Time {
	u := t.UTC()
	return &u
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
