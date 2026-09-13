package service

import (
	"encoding/json"
	"time"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
)

type NotificationService struct {
	repo *repository.NotificationRepo
}

func NewNotificationService(nr *repository.NotificationRepo) *NotificationService {
	return &NotificationService{repo: nr}
}

func (s *NotificationService) List(uid uint64, page, size int, onlyUnread bool) ([]*domain.Notification, int64, error) {
	return s.repo.ListByUser(uid, page, size, onlyUnread)
}

func (s *NotificationService) CountUnread(uid uint64) (int64, error) {
	return s.repo.CountUnread(uid)
}

func (s *NotificationService) MarkRead(uid, id uint64) error {
	return s.repo.MarkRead(uid, id)
}

func (s *NotificationService) MarkAllRead(uid uint64) (int64, error) {
	return s.repo.MarkAllRead(uid)
}

func (s *NotificationService) ClearOld(uid uint64, days int) (int64, error) {
	var before *time.Time
	if days > 0 {
		t := time.Now().AddDate(0, 0, -days)
		before = &t
	}
	return s.repo.DeleteByUser(uid, before)
}

func (s *NotificationService) NotifyComment(userID, actorID, tutorialID, commentID uint64, actorName, snippet string) error {
	n := &domain.Notification{
		UserID:     userID,
		ActorID:    actorID,
		Type:       domain.NotifComment,
		Title:      "收到新评论",
		Content:    actorName + " 评论了你的教程：" + truncate(snippet, 80),
		TutorialID: tutorialID,
		CommentID:  commentID,
	}
	return s.repo.Create(n)
}

func (s *NotificationService) NotifyReply(userID, actorID, tutorialID, commentID uint64, actorName, snippet string) error {
	n := &domain.Notification{
		UserID:     userID,
		ActorID:    actorID,
		Type:       domain.NotifReply,
		Title:      "回复了你",
		Content:    actorName + " 在评论中回复了你：" + truncate(snippet, 80),
		TutorialID: tutorialID,
		CommentID:  commentID,
	}
	return s.repo.Create(n)
}

func (s *NotificationService) NotifyFavorite(userID, actorID, tutorialID uint64, actorName string, targetType string) error {
	content := actorName + " 收藏了你的"
	title := "收到收藏"
	tid := uint64(0)
	pid := uint64(0)
	if targetType == "tutorial" {
		content += "教程"
		tid = tutorialID
	} else {
		content += "作品"
		pid = tutorialID
	}
	n := &domain.Notification{
		UserID:     userID,
		ActorID:    actorID,
		Type:       domain.NotifFavorite,
		Title:      title,
		Content:    content,
		TutorialID: tid,
		ProjectID:  pid,
	}
	return s.repo.Create(n)
}

func (s *NotificationService) NotifyFollow(userID, actorID uint64, actorName string) error {
	n := &domain.Notification{
		UserID:  userID,
		ActorID: actorID,
		Type:    domain.NotifFollow,
		Title:   "新的关注者",
		Content: actorName + " 关注了你",
	}
	return s.repo.Create(n)
}

func (s *NotificationService) NotifyAttempt(userID, actorID, tutorialID uint64, actorName, tutorialTitle string) error {
	n := &domain.Notification{
		UserID:     userID,
		ActorID:    actorID,
		Type:       domain.NotifAttempt,
		Title:      "有人开始尝试你的教程",
		Content:    actorName + " 标记开始尝试《" + truncate(tutorialTitle, 40) + "》",
		TutorialID: tutorialID,
	}
	return s.repo.Create(n)
}

func (s *NotificationService) NotifyProject(userID, actorID, tutorialID, projectID uint64, actorName, tutorialTitle string) error {
	n := &domain.Notification{
		UserID:     userID,
		ActorID:    actorID,
		Type:       domain.NotifProject,
		Title:      "有新的复刻作品",
		Content:    actorName + " 根据《" + truncate(tutorialTitle, 40) + "》发布了改造作品",
		TutorialID: tutorialID,
		ProjectID:  projectID,
	}
	return s.repo.Create(n)
}

func (s *NotificationService) NotifyAuditResult(userID, tutorialID uint64, pass bool, reason string) error {
	typ := domain.NotifAuditPass
	title := "审核通过"
	content := "你的教程审核已通过"
	if !pass {
		typ = domain.NotifAuditReject
		title = "审核未通过"
		content = "你的教程未通过审核，原因：" + truncate(reason, 120)
	}
	n := &domain.Notification{
		UserID:     userID,
		Type:       typ,
		Title:      title,
		Content:    content,
		TutorialID: tutorialID,
	}
	return s.repo.Create(n)
}

func (s *NotificationService) NotifySystem(userID uint64, title, content string) error {
	n := &domain.Notification{
		UserID:  userID,
		Type:    domain.NotifSystem,
		Title:   title,
		Content: content,
	}
	return s.repo.Create(n)
}

func (s *NotificationService) BroadcastSystem(title, content string, userIDs []uint64) (int, error) {
	if len(userIDs) == 0 {
		return 0, nil
	}
	list := make([]*domain.Notification, 0, len(userIDs))
	for _, uid := range userIDs {
		list = append(list, &domain.Notification{
			UserID:  uid,
			Type:    domain.NotifSystem,
			Title:   title,
			Content: content,
		})
	}
	if err := s.repo.BatchCreate(list); err != nil {
		return 0, err
	}
	return len(list), nil
}

func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

type AuditService struct {
	repo *repository.AuditLogRepo
}

func NewAuditService(ar *repository.AuditLogRepo) *AuditService {
	return &AuditService{repo: ar}
}

func (s *AuditService) Log(userID uint64, action, targetType string, targetID uint64, ip, ua string, before, after any, remark string) error {
	beforeStr := ""
	afterStr := ""
	if before != nil {
		if b, err := json.Marshal(before); err == nil {
			beforeStr = string(b)
		}
	}
	if after != nil {
		if b, err := json.Marshal(after); err == nil {
			afterStr = string(b)
		}
	}
	l := &domain.AuditLog{
		UserID:     userID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		IP:         ip,
		UserAgent:  ua,
		Before:     beforeStr,
		After:      afterStr,
		Remark:     remark,
	}
	return s.repo.Create(l)
}

func (s *AuditService) List(page, size int, userID uint64, action, targetType string, from, to *time.Time) ([]*domain.AuditLog, int64, error) {
	return s.repo.List(page, size, userID, action, targetType, from, to)
}

func (s *AuditService) Stats(days int) (map[string]int64, error) {
	return s.repo.StatsByAction(days)
}

type TutorialHistoryService struct {
	versionRepo *repository.TutorialVersionRepo
	tutorialRepo *repository.TutorialRepo
	stepRepo     *repository.StepRepo
	materialRepo *repository.MaterialRepo
	toolRepo     *repository.ToolRepo
}

func NewTutorialHistoryService(vr *repository.TutorialVersionRepo, tr *repository.TutorialRepo, sr *repository.StepRepo, mr *repository.MaterialRepo, tor *repository.ToolRepo) *TutorialHistoryService {
	return &TutorialHistoryService{versionRepo: vr, tutorialRepo: tr, stepRepo: sr, materialRepo: mr, toolRepo: tor}
}

func (s *TutorialHistoryService) Snapshot(tid uint64) error {
	t, err := s.tutorialRepo.GetByID(tid, false)
	if err != nil {
		return err
	}
	steps, _ := s.stepRepo.ListByTutorial(tid)
	materials, tools, _ := s.materialRepo.ListByTutorial(tid)
	dump := map[string]any{
		"title":       t.Title,
		"summary":     t.Summary,
		"cover_before": t.CoverBefore,
		"cover_after": t.CoverAfter,
		"difficulty":  t.Difficulty,
		"hours":       t.EstimatedHours,
		"category_id": t.CategoryID,
		"status":      t.Status,
		"slug":        t.Slug,
		"steps":       steps,
		"materials":   materials,
		"tools":       tools,
	}
	b, _ := json.Marshal(dump)
	v := &domain.TutorialVersion{
		TutorialID:  tid,
		Version:     t.Version,
		Title:       t.Title,
		Summary:     t.Summary,
		ContentDump: string(b),
	}
	return s.versionRepo.Create(v)
}

func (s *TutorialHistoryService) List(tid uint64, page, size int) ([]*domain.TutorialVersion, int64, error) {
	return s.versionRepo.ListByTutorial(tid, page, size)
}

func (s *TutorialHistoryService) Get(tid uint64, version int) (*domain.TutorialVersion, error) {
	return s.versionRepo.GetByVersion(tid, version)
}

func (s *TutorialHistoryService) Rollback(tid, actorID uint64, version int) error {
	v, err := s.versionRepo.GetByVersion(tid, version)
	if err != nil {
		return err
	}
	dump := map[string]any{}
	if e := json.Unmarshal([]byte(v.ContentDump), &dump); e != nil {
		return ErrValidation("版本内容损坏")
	}
	t, err := s.tutorialRepo.GetByID(tid, false)
	if err != nil {
		return err
	}
	if t.UserID != actorID {
		return ErrForbidden("无权回滚该教程")
	}
	if title, ok := dump["title"].(string); ok {
		t.Title = title
	}
	if summary, ok := dump["summary"].(string); ok {
		t.Summary = summary
	}
	if cb, ok := dump["cover_before"].(string); ok {
		t.CoverBefore = cb
	}
	if ca, ok := dump["cover_after"].(string); ok {
		t.CoverAfter = ca
	}
	if diff, ok := dump["difficulty"].(string); ok {
		t.Difficulty = diff
	}
	if slug, ok := dump["slug"].(string); ok {
		t.Slug = slug
	}
	if h, ok := dump["hours"].(float64); ok {
		t.EstimatedHours = h
	}
	if cid, ok := dump["category_id"].(float64); ok {
		t.CategoryID = uint64(cid)
	}
	t.Version++
	if err := s.tutorialRepo.Update(t); err != nil {
		return err
	}
	steps := []*domain.Step{}
	if raw, ok := dump["steps"].([]any); ok {
		for _, item := range raw {
			b, _ := json.Marshal(item)
			st := &domain.Step{}
			if json.Unmarshal(b, st) == nil {
				st.ID = 0
				st.TutorialID = tid
				steps = append(steps, st)
			}
		}
	}
	if err := s.stepRepo.ReplaceByTutorial(tid, steps); err != nil {
		return err
	}
	mats := []*domain.Material{}
	if raw, ok := dump["materials"].([]any); ok {
		for _, item := range raw {
			b, _ := json.Marshal(item)
			m := &domain.Material{}
			if json.Unmarshal(b, m) == nil {
				m.ID = 0
				m.TutorialID = tid
				mats = append(mats, m)
			}
		}
	}
	toolItems := []*domain.Material{}
	if raw, ok := dump["tools"].([]any); ok {
		for _, item := range raw {
			b, _ := json.Marshal(item)
			m := &domain.Material{}
			if json.Unmarshal(b, m) == nil {
				m.ID = 0
				m.TutorialID = tid
				toolItems = append(toolItems, m)
			}
		}
	}
	_ = s.materialRepo.DeleteByTutorial(tid)
	all := append(mats, toolItems...)
	if len(all) > 0 {
		_ = s.materialRepo.BatchCreate(all)
	}
	newTools := []*domain.Tool{}
	for idx, mt := range toolItems {
		newTools = append(newTools, &domain.Tool{
			TutorialID: tid,
			Name:       mt.Name,
			Quantity:   mt.Quantity,
			Optional:   false,
			SortOrder:  idx,
		})
	}
	_ = s.toolRepo.ReplaceByTutorial(tid, newTools)
	return s.Snapshot(tid)
}
