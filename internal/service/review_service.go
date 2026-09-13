package service

import (
	"time"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
)

// ReviewService 负责教程审核流：作者提交 -> 管理员通过/驳回 -> 通知作者。
type ReviewService struct {
	tutorialRepo *repository.TutorialRepo
	categoryRepo *repository.CategoryRepo
	userRepo     *repository.UserRepo
	notifSvc     *NotificationService
	auditSvc     *AuditService
}

func NewReviewService(tr *repository.TutorialRepo, cr *repository.CategoryRepo, ur *repository.UserRepo,
	ns *NotificationService, as *AuditService) *ReviewService {
	return &ReviewService{tutorialRepo: tr, categoryRepo: cr, userRepo: ur, notifSvc: ns, auditSvc: as}
}

// Submit 作者将草稿/被驳回的教程提交审核，状态流转为 pending。
func (s *ReviewService) Submit(id, userID uint64) (*domain.Tutorial, error) {
	t, err := s.tutorialRepo.GetByID(id, false)
	if err != nil {
		return nil, err
	}
	if t.UserID != userID {
		return nil, ErrForbidden("无权提交此教程")
	}
	switch t.Status {
	case domain.TutorialStatusDraft, domain.TutorialStatusRejected, domain.TutorialStatusArchived:
		// 允许提交
	case domain.TutorialStatusPending:
		return nil, ErrValidation("教程已在审核队列中，请耐心等待")
	default:
		return nil, ErrValidation("当前状态无需提交审核")
	}
	t.Status = domain.TutorialStatusPending
	t.ReviewNote = ""
	t.ReviewedAt = nil
	if err := s.tutorialRepo.Update(t); err != nil {
		return nil, err
	}
	return t, nil
}

// ListByStatus 管理端按状态查看教程（默认待审）。
func (s *ReviewService) ListByStatus(status string, page, size int) ([]*domain.Tutorial, int64, error) {
	switch status {
	case domain.TutorialStatusPending, domain.TutorialStatusRejected,
		domain.TutorialStatusPublished, domain.TutorialStatusDraft, domain.TutorialStatusArchived:
	default:
		status = domain.TutorialStatusPending
	}
	return s.tutorialRepo.List(page, size, 0, "", status, "new", "", 0)
}

// Approve 审核通过：pending -> published，计数入账并通知作者。
func (s *ReviewService) Approve(id, adminID uint64, ip, ua string) (*domain.Tutorial, error) {
	t, err := s.tutorialRepo.GetByID(id, false)
	if err != nil {
		return nil, err
	}
	if t.Status != domain.TutorialStatusPending {
		return nil, ErrValidation("仅待审核的教程可以通过")
	}
	now := time.Now()
	t.Status = domain.TutorialStatusPublished
	t.ReviewNote = ""
	t.ReviewedAt = &now
	if err := s.tutorialRepo.Update(t); err != nil {
		return nil, err
	}
	s.categoryRepo.IncCount(t.CategoryID, 1)
	s.userRepo.IncStats(t.UserID, 1, 0, 50)
	s.notifSvc.NotifyAuditResult(t.UserID, t.ID, true, "")
	s.auditSvc.Log(adminID, domain.AuditAdminAuditPass, "tutorial", t.ID, ip, ua, nil,
		map[string]any{"title": t.Title, "status": t.Status}, "审核通过")
	return t, nil
}

// Reject 审核驳回：pending -> rejected，记录原因并通知作者。
func (s *ReviewService) Reject(id, adminID uint64, reason, ip, ua string) (*domain.Tutorial, error) {
	if reason == "" {
		return nil, ErrValidation("驳回时必须填写原因")
	}
	t, err := s.tutorialRepo.GetByID(id, false)
	if err != nil {
		return nil, err
	}
	if t.Status != domain.TutorialStatusPending {
		return nil, ErrValidation("仅待审核的教程可以驳回")
	}
	now := time.Now()
	t.Status = domain.TutorialStatusRejected
	t.ReviewNote = reason
	t.ReviewedAt = &now
	if err := s.tutorialRepo.Update(t); err != nil {
		return nil, err
	}
	s.notifSvc.NotifyAuditResult(t.UserID, t.ID, false, reason)
	s.auditSvc.Log(adminID, domain.AuditAdminAuditReject, "tutorial", t.ID, ip, ua, nil,
		map[string]any{"title": t.Title, "status": t.Status, "reason": reason}, "审核驳回")
	return t, nil
}
