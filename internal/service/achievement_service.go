package service

import (
	"encoding/json"
	"strconv"
	"time"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
	"upcycle-hub/pkg/logger"
)

// GrantedBadge 一次事件中新颁发的徽章，供接口返回/前端提示
type GrantedBadge struct {
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Icon        string    `json:"icon"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	AwardedAt   time.Time `json:"awarded_at"`
}

// BadgeView 个人中心徽章视图：定义 + 获得状态 + 当前进度
type BadgeView struct {
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Icon        string     `json:"icon"`
	Category    string     `json:"category"`
	Condition   string     `json:"condition"`
	Threshold   int        `json:"threshold"`
	Sort        int        `json:"sort"`
	Earned      bool       `json:"earned"`
	Progress    int        `json:"progress"`
	AwardedAt   *time.Time `json:"awarded_at,omitempty"`
}

type AchievementService struct {
	badgeRepo *repository.BadgeRepo
	notifSvc  *NotificationService
}

func NewAchievementService(br *repository.BadgeRepo, ns *NotificationService) *AchievementService {
	return &AchievementService{badgeRepo: br, notifSvc: ns}
}

// ---- 事件入口：业务侧调用。任何错误只记日志，不影响主流程 ----

// OnTutorialPublished 教程首次发布或由草稿转为发布时触发
func (s *AchievementService) OnTutorialPublished(userID uint64) []GrantedBadge {
	if s == nil || userID == 0 {
		return nil
	}
	granted, err := s.evaluate(userID, domain.BadgeEventTutorialPublished, 0)
	if err != nil {
		logger.Errorf("achievement OnTutorialPublished user=%d: %v", userID, err)
	}
	return granted
}

// OnProjectLiked 作品被点赞后触发（判定对象是作品作者）
func (s *AchievementService) OnProjectLiked(ownerID, projectID uint64) {
	if s == nil || ownerID == 0 {
		return
	}
	if _, err := s.evaluate(ownerID, domain.BadgeEventProjectLiked, projectID); err != nil {
		logger.Errorf("achievement OnProjectLiked user=%d project=%d: %v", ownerID, projectID, err)
	}
}

// OnTutorialFavorited 收藏教程后触发
func (s *AchievementService) OnTutorialFavorited(userID uint64) []GrantedBadge {
	if s == nil || userID == 0 {
		return nil
	}
	granted, err := s.evaluate(userID, domain.BadgeEventTutorialFavorited, 0)
	if err != nil {
		logger.Errorf("achievement OnTutorialFavorited user=%d: %v", userID, err)
	}
	return granted
}

// Sweep 全量对账：遍历所有用户，补发因任何原因漏掉的徽章。
// 对账只补发、不收回；为避免流水膨胀，未达成不写流水。
func (s *AchievementService) Sweep() (int, error) {
	if s == nil {
		return 0, nil
	}
	ids, err := s.badgeRepo.ListAllUserIDs()
	if err != nil {
		return 0, err
	}
	total := 0
	for _, uid := range ids {
		granted, err := s.evaluate(uid, domain.BadgeEventSweep, 0)
		if err != nil {
			logger.Errorf("achievement sweep user=%d: %v", uid, err)
			continue
		}
		total += len(granted)
	}
	return total, nil
}

// evaluate 采集指标 → 逐个判定事件相关徽章 → 幂等发放 + 写流水
func (s *AchievementService) evaluate(userID uint64, event string, projectID uint64) ([]GrantedBadge, error) {
	conds := conditionsForEvent(event)
	if len(conds) == 0 {
		return nil, nil
	}
	metrics, err := s.collectMetrics(userID, conds, projectID)
	if err != nil {
		return nil, err
	}
	earned, err := s.earnedSet(userID)
	if err != nil {
		return nil, err
	}
	var granted []GrantedBadge
	for _, def := range domain.BadgeRegistry {
		if !containsCond(conds, def.Condition) {
			continue
		}
		met, value := s.checkCondition(def, metrics)
		if !met {
			// 对账场景跳过未达成记录；实时事件完整记录判定结果
			if event != domain.BadgeEventSweep {
				s.writeLog(userID, def.Code, event, domain.BadgeAwardNotMet, metrics,
					"条件未达成，当前值 "+strconv.Itoa(value)+"/"+strconv.Itoa(def.Threshold))
			}
			continue
		}
		if _, ok := earned[def.Code]; ok {
			if event != domain.BadgeEventSweep {
				s.writeLog(userID, def.Code, event, domain.BadgeAwardDuplicate, metrics, "已拥有该徽章，不重复颁发")
			}
			continue
		}
		ub := &domain.UserBadge{
			UserID:         userID,
			BadgeCode:      def.Code,
			AwardEvent:     event,
			MetricSnapshot: snapshotJSON(metrics),
		}
		inserted, gerr := s.badgeRepo.Grant(ub)
		if gerr != nil {
			logger.Errorf("grant badge %s user=%d: %v", def.Code, userID, gerr)
			continue
		}
		if !inserted {
			// 并发下唯一索引兜底：视为重复
			s.writeLog(userID, def.Code, event, domain.BadgeAwardDuplicate, metrics, "唯一索引兜底：已存在")
			continue
		}
		s.writeLog(userID, def.Code, event, domain.BadgeAwardGranted, metrics,
			"达成并颁发，当前值 "+strconv.Itoa(value)+"/"+strconv.Itoa(def.Threshold))
		gb := GrantedBadge{
			Code: def.Code, Name: def.Name, Icon: def.Icon,
			Description: def.Description, Category: def.Category,
			AwardedAt: ub.CreatedAt,
		}
		if ub.CreatedAt.IsZero() {
			gb.AwardedAt = time.Now()
		}
		granted = append(granted, gb)
		earned[def.Code] = struct{}{}
		s.notify(userID, def)
		logger.Infof("badge granted: user=%d badge=%s event=%s value=%d", userID, def.Code, event, value)
	}
	return granted, nil
}

// checkCondition 返回是否达成及对应指标当前值
func (s *AchievementService) checkCondition(def domain.BadgeDef, m *domain.AchievementMetrics) (bool, int) {
	switch def.Condition {
	case domain.BadgeCondFirstPublish:
		return m.PublishedCount >= def.Threshold, m.PublishedCount
	case domain.BadgeCondProjectLikes:
		return m.MaxProjectLikes >= def.Threshold, m.MaxProjectLikes
	case domain.BadgeCondPublishStreak:
		return m.PublishStreak >= def.Threshold, m.PublishStreak
	case domain.BadgeCondFavoriteOthers:
		return m.FavoriteOthers >= def.Threshold, m.FavoriteOthers
	}
	return false, 0
}

// collectMetrics 只采集本次条件组需要的指标，减少无谓查询
func (s *AchievementService) collectMetrics(userID uint64, conds []string, projectID uint64) (*domain.AchievementMetrics, error) {
	m := &domain.AchievementMetrics{}
	need := func(c string) bool { return containsCond(conds, c) }
	if need(domain.BadgeCondFirstPublish) {
		n, err := s.badgeRepo.CountPublishedTutorials(userID)
		if err != nil {
			return nil, err
		}
		m.PublishedCount = n
	}
	if need(domain.BadgeCondProjectLikes) {
		maxLikes, pid, err := s.badgeRepo.MaxProjectLikes(userID)
		if err != nil {
			return nil, err
		}
		m.MaxProjectLikes = maxLikes
		m.MaxLikesProjectID = pid
		_ = projectID // 触发作品信息保留在快照中由事件日志体现
	}
	if need(domain.BadgeCondPublishStreak) {
		months, err := s.badgeRepo.PublishedMonths(userID)
		if err != nil {
			return nil, err
		}
		m.PublishMonths = months
		m.PublishStreak = longestMonthRun(months)
	}
	if need(domain.BadgeCondFavoriteOthers) {
		n, err := s.badgeRepo.CountFavoriteOthers(userID)
		if err != nil {
			return nil, err
		}
		m.FavoriteOthers = n
	}
	return m, nil
}

func (s *AchievementService) earnedSet(userID uint64) (map[string]struct{}, error) {
	list, err := s.badgeRepo.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	set := make(map[string]struct{}, len(list))
	for _, ub := range list {
		set[ub.BadgeCode] = struct{}{}
	}
	return set, nil
}

// MyBadges 个人中心：全部徽章定义 + 获得状态 + 实时进度
func (s *AchievementService) MyBadges(userID uint64) ([]*BadgeView, int, error) {
	owned, err := s.badgeRepo.ListByUser(userID)
	if err != nil {
		return nil, 0, err
	}
	ownedMap := map[string]*domain.UserBadge{}
	for _, ub := range owned {
		ownedMap[ub.BadgeCode] = ub
	}
	allConds := []string{
		domain.BadgeCondFirstPublish, domain.BadgeCondProjectLikes,
		domain.BadgeCondPublishStreak, domain.BadgeCondFavoriteOthers,
	}
	metrics, err := s.collectMetrics(userID, allConds, 0)
	if err != nil {
		return nil, 0, err
	}
	views := make([]*BadgeView, 0, len(domain.BadgeRegistry))
	earnedCount := 0
	for _, def := range domain.BadgeRegistry {
		_, value := s.checkCondition(def, metrics)
		v := &BadgeView{
			Code: def.Code, Name: def.Name, Description: def.Description,
			Icon: def.Icon, Category: def.Category, Condition: def.Condition,
			Threshold: def.Threshold, Sort: def.Sort, Progress: value,
		}
		if value > def.Threshold {
			v.Progress = def.Threshold
		}
		if ub, ok := ownedMap[def.Code]; ok {
			v.Earned = true
			v.Progress = def.Threshold
			t := ub.CreatedAt
			v.AwardedAt = &t
			earnedCount++
		}
		views = append(views, v)
	}
	return views, earnedCount, nil
}

// PublicBadges 他人主页只展示已获得的徽章
func (s *AchievementService) PublicBadges(userID uint64) ([]*BadgeView, error) {
	owned, err := s.badgeRepo.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	views := make([]*BadgeView, 0, len(owned))
	for _, ub := range owned {
		def, ok := domain.GetBadgeDef(ub.BadgeCode)
		if !ok {
			continue
		}
		t := ub.CreatedAt
		views = append(views, &BadgeView{
			Code: def.Code, Name: def.Name, Description: def.Description,
			Icon: def.Icon, Category: def.Category, Condition: def.Condition,
			Threshold: def.Threshold, Sort: def.Sort, Earned: true,
			Progress: def.Threshold, AwardedAt: &t,
		})
	}
	return views, nil
}

// ListLogs 管理端追踪判定条件与发放结果
func (s *AchievementService) ListLogs(page, size int, userID uint64, badgeCode, result, event string) ([]*domain.BadgeAwardLog, int64, error) {
	return s.badgeRepo.ListLogs(page, size, userID, badgeCode, result, event)
}

func (s *AchievementService) writeLog(userID uint64, code, event, result string, m *domain.AchievementMetrics, remark string) {
	l := &domain.BadgeAwardLog{
		UserID:         userID,
		BadgeCode:      code,
		Event:          event,
		Result:         result,
		MetricSnapshot: snapshotJSON(m),
		Remark:         remark,
	}
	if err := s.badgeRepo.CreateLog(l); err != nil {
		logger.Errorf("write badge log: %v", err)
	}
}

func (s *AchievementService) notify(userID uint64, def domain.BadgeDef) {
	if s.notifSvc == nil {
		return
	}
	title := "🏅 获得新徽章：" + def.Name
	content := "恭喜达成「" + def.Description + "」，徽章已放入个人中心"
	if err := s.notifSvc.NotifySystem(userID, title, content); err != nil {
		logger.Errorf("notify badge user=%d: %v", userID, err)
	}
}

// ---- 辅助函数 ----

func conditionsForEvent(event string) []string {
	switch event {
	case domain.BadgeEventTutorialPublished:
		return []string{domain.BadgeCondFirstPublish, domain.BadgeCondPublishStreak}
	case domain.BadgeEventProjectLiked:
		return []string{domain.BadgeCondProjectLikes}
	case domain.BadgeEventTutorialFavorited:
		return []string{domain.BadgeCondFavoriteOthers}
	case domain.BadgeEventSweep:
		return []string{
			domain.BadgeCondFirstPublish, domain.BadgeCondProjectLikes,
			domain.BadgeCondPublishStreak, domain.BadgeCondFavoriteOthers,
		}
	}
	return nil
}

func containsCond(list []string, c string) bool {
	for _, v := range list {
		if v == c {
			return true
		}
	}
	return false
}

// longestMonthRun 计算月份序列中最长的连续段长度。
// 以"历史最长连续"为准：曾经连续达成过即保留，后续中断不收回，
// 重新连续发布时也不会再次满足同一门槛（徽章已在）。
func longestMonthRun(months []string) int {
	if len(months) == 0 {
		return 0
	}
	best, run := 1, 1
	for i := 1; i < len(months); i++ {
		if prevMonth(months[i]) == months[i-1] {
			run++
			if run > best {
				best = run
			}
		} else if months[i] != months[i-1] {
			run = 1
		}
	}
	return best
}

// prevMonth 返回 t 的上一个月，t 格式 YYYY-mm
func prevMonth(t string) string {
	tm, err := time.Parse("2006-01", t)
	if err != nil {
		return ""
	}
	return tm.AddDate(0, -1, 0).Format("2006-01")
}

func snapshotJSON(m *domain.AchievementMetrics) string {
	if m == nil {
		return ""
	}
	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(b)
}
