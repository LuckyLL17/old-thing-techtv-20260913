package service

import (
	"path/filepath"
	"testing"
	"time"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := filepath.Join(t.TempDir(), "test.db")
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(
		&domain.User{}, &domain.Category{}, &domain.Tutorial{}, &domain.TutorialTag{}, &domain.Tag{},
		&domain.Project{}, &domain.Favorite{}, &domain.UserBadge{}, &domain.BadgeAwardLog{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func seedUser(t *testing.T, db *gorm.DB, id uint64, name string) {
	t.Helper()
	u := &domain.User{ID: id, Username: name, Email: name + "@test.com", PasswordHash: "x", Status: 1}
	if err := db.Create(u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
}

func seedTutorial(t *testing.T, db *gorm.DB, userID uint64, status, month string) {
	t.Helper()
	ts, err := time.Parse("2006-01", month)
	if err != nil {
		t.Fatal(err)
	}
	tut := &domain.Tutorial{
		UserID: userID, CategoryID: 1, Title: "t-" + month + "-" + time.Now().Format("150405.000000"),
		CoverBefore: "b.jpg", CoverAfter: "a.jpg", Status: status, CreatedAt: ts,
	}
	if err := db.Create(tut).Error; err != nil {
		t.Fatalf("create tutorial: %v", err)
	}
}

func countBadges(t *testing.T, db *gorm.DB, userID uint64) map[string]int {
	t.Helper()
	var list []domain.UserBadge
	db.Where("user_id = ?", userID).Find(&list)
	m := map[string]int{}
	for _, b := range list {
		m[b.BadgeCode]++
	}
	return m
}

func hasBadge(db *gorm.DB, userID uint64, code string) bool {
	var n int64
	db.Model(&domain.UserBadge{}).Where("user_id = ? AND badge_code = ?", userID, code).Count(&n)
	return n > 0
}

// 首次发布教程：发布前不发，发布后颁发，再次触发不重发
func TestFirstPublishBadge(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, 1, "alice")
	svc := NewAchievementService(repository.NewBadgeRepo(db), nil)

	if g := svc.OnTutorialPublished(1); len(g) != 0 {
		t.Fatalf("无已发布教程时不应获得徽章, got %d", len(g))
	}
	seedTutorial(t, db, 1, domain.TutorialStatusPublished, "2026-01")
	g := svc.OnTutorialPublished(1)
	if len(g) != 1 || g[0].Code != "first_publish_1" {
		t.Fatalf("首次发布应颁发 first_publish_1, got %+v", g)
	}
	// 草稿不计入
	seedTutorial(t, db, 1, domain.TutorialStatusDraft, "2026-02")
	if g := svc.OnTutorialPublished(1); len(g) != 0 {
		t.Fatalf("重复达成不应发第二枚, got %d", len(g))
	}
	m := countBadges(t, db, 1)
	if m["first_publish_1"] != 1 {
		t.Fatalf("first_publish_1 应且仅应存在 1 枚, got %d", m["first_publish_1"])
	}
}

// 收藏他人教程达到 10/50：自己的教程不计入；取消收藏不收回
func TestFavoriteOthersBadge(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, 1, "bob")
	seedUser(t, db, 2, "carol")
	br := repository.NewBadgeRepo(db)
	svc := NewAchievementService(br, nil)

	// carol 发布 60 篇教程
	for i := 0; i < 60; i++ {
		seedTutorial(t, db, 2, domain.TutorialStatusPublished, "2026-01")
	}
	// bob 自己也有 1 篇，自己收藏自己不应计数
	seedTutorial(t, db, 1, domain.TutorialStatusPublished, "2026-01")
	var ownTuts []domain.Tutorial
	db.Where("user_id = ?", 1).Find(&ownTuts)
	db.Create(&domain.Favorite{UserID: 1, TargetType: domain.FavTypeTutorial, TargetID: ownTuts[0].ID})

	var others []domain.Tutorial
	db.Where("user_id = ?", 2).Limit(9).Find(&others)
	for _, tut := range others {
		db.Create(&domain.Favorite{UserID: 1, TargetType: domain.FavTypeTutorial, TargetID: tut.ID})
	}
	if g := svc.OnTutorialFavorited(1); len(g) != 0 {
		t.Fatalf("收藏 9 篇他人教程不应达成 10 门槛, got %+v", g)
	}
	var more []domain.Tutorial
	db.Where("user_id = ?", 2).Offset(9).Limit(1).Find(&more)
	db.Create(&domain.Favorite{UserID: 1, TargetType: domain.FavTypeTutorial, TargetID: more[0].ID})
	g := svc.OnTutorialFavorited(1)
	if len(g) != 1 || g[0].Code != "favorite_others_10" {
		t.Fatalf("收藏 10 篇应颁发 favorite_others_10, got %+v", g)
	}
	// 数据回落：取消一个收藏，徽章不收回
	db.Where("user_id = ? AND target_id = ?", 1, more[0].ID).Delete(&domain.Favorite{})
	n, _ := br.CountFavoriteOthers(1)
	if n != 9 {
		t.Fatalf("取消收藏后计数应为 9, got %d", n)
	}
	if !hasBadge(db, 1, "favorite_others_10") {
		t.Fatal("数据回落后已颁发的徽章被收回")
	}
	// 重复触发不发第二枚
	if g := svc.OnTutorialFavorited(1); len(g) != 0 {
		t.Fatalf("回落区间内重复触发不应再发, got %d", len(g))
	}
	// 重新收藏到 50 枚（当前 9，再补 41 篇），颁发 50 档；10 档不重发
	var rest []domain.Tutorial
	db.Where("user_id = ?", 2).Offset(9).Limit(41).Find(&rest)
	if len(rest) != 41 {
		t.Fatalf("测试数据不足，需要 41 篇可收藏教程, got %d", len(rest))
	}
	for _, tut := range rest {
		db.Create(&domain.Favorite{UserID: 1, TargetType: domain.FavTypeTutorial, TargetID: tut.ID})
	}
	g = svc.OnTutorialFavorited(1)
	if len(g) != 1 || g[0].Code != "favorite_others_50" {
		t.Fatalf("收藏 50 篇应只新颁发 favorite_others_50, got %+v", g)
	}
	m := countBadges(t, db, 1)
	if m["favorite_others_10"] != 1 || m["favorite_others_50"] != 1 {
		t.Fatalf("两档收藏徽章应各一枚, got %+v", m)
	}
}

// 单个作品点赞达到门槛：10/50/100 三档按最高跨度一次颁发且各只一枚
func TestProjectLikesBadge(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, 1, "dave")
	svc := NewAchievementService(repository.NewBadgeRepo(db), nil)

	p := &domain.Project{UserID: 1, TutorialID: 1, Status: 1, LikeCount: 55}
	if err := db.Create(p).Error; err != nil {
		t.Fatal(err)
	}
	svc.OnProjectLiked(1, p.ID)
	m := countBadges(t, db, 1)
	if m["project_likes_10"] != 1 || m["project_likes_50"] != 1 {
		t.Fatalf("点赞 55 应同时持有 10/50 两档, got %+v", m)
	}
	if m["project_likes_100"] != 0 {
		t.Fatal("点赞 55 不应获得 100 档")
	}
	// 再次点赞事件，不重复颁发
	svc.OnProjectLiked(1, p.ID)
	m = countBadges(t, db, 1)
	if m["project_likes_10"] != 1 || m["project_likes_50"] != 1 {
		t.Fatalf("重复点赞事件导致重发: %+v", m)
	}
	// 点赞数回落，已获得的不收回
	db.Model(&domain.Project{}).Where("id = ?", p.ID).Update("like_count", 0)
	svc.OnProjectLiked(1, p.ID)
	if !hasBadge(db, 1, "project_likes_50") {
		t.Fatal("点赞回落后徽章被收回")
	}
}

// 连续 N 个月发布：历史最长连段达成后颁发，中断不收回
func TestPublishStreakBadge(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, 1, "erin")
	svc := NewAchievementService(repository.NewBadgeRepo(db), nil)

	for _, mth := range []string{"2026-01", "2026-02", "2026-03"} {
		seedTutorial(t, db, 1, domain.TutorialStatusPublished, mth)
	}
	g := svc.OnTutorialPublished(1)
	if len(g) != 2 {
		t.Fatalf("首发+连续3月应颁发 2 枚, got %+v", g)
	}
	if !hasBadge(db, 1, "publish_streak_3") || !hasBadge(db, 1, "first_publish_1") {
		t.Fatal("缺少 publish_streak_3 或 first_publish_1")
	}
	// 4 月断更，5-10 月连续 6 个月（另一段），应补到 6 月档
	for _, mth := range []string{"2026-05", "2026-06", "2026-07", "2026-08", "2026-09", "2026-10"} {
		seedTutorial(t, db, 1, domain.TutorialStatusPublished, mth)
	}
	g = svc.OnTutorialPublished(1)
	var codes []string
	for _, b := range g {
		codes = append(codes, b.Code)
	}
	if len(codes) != 1 || codes[0] != "publish_streak_6" {
		t.Fatalf("最长连续 6 个月应只新颁发 streak_6, got %v", codes)
	}
	m := countBadges(t, db, 1)
	if m["publish_streak_3"] != 1 || m["publish_streak_6"] != 1 {
		t.Fatalf("连续月徽章应各一枚, got %+v", m)
	}
}

// 定时对账：历史用户未触发过事件也能补发；再次对账零新增
func TestSweepBackfill(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, 1, "frank")
	seedUser(t, db, 2, "grace")
	br := repository.NewBadgeRepo(db)
	svc := NewAchievementService(br, nil)

	seedTutorial(t, db, 1, domain.TutorialStatusPublished, "2026-01")
	// grace 有 3 个月连续发布
	for _, mth := range []string{"2026-02", "2026-03", "2026-04"} {
		seedTutorial(t, db, 2, domain.TutorialStatusPublished, mth)
	}
	n, err := svc.Sweep()
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("首次对账应补发 3 枚（frank 1 + grace 2）, got %d", n)
	}
	if !hasBadge(db, 1, "first_publish_1") {
		t.Fatal("对账未补发 frank 的首发徽章")
	}
	if !hasBadge(db, 2, "first_publish_1") || !hasBadge(db, 2, "publish_streak_3") {
		t.Fatal("对账未补发 grace 的徽章")
	}
	// 再次对账不产生新徽章
	n, _ = svc.Sweep()
	if n != 0 {
		t.Fatalf("重复对账不应再补发, got %d", n)
	}
}

// Grant 的唯一索引兜底：并发/重复插入只算一次
func TestGrantIdempotent(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, 1, "henry")
	br := repository.NewBadgeRepo(db)
	ok1, err := br.Grant(&domain.UserBadge{UserID: 1, BadgeCode: "first_publish_1", AwardEvent: domain.BadgeEventTutorialPublished})
	if err != nil || !ok1 {
		t.Fatalf("首次颁发应成功, ok=%v err=%v", ok1, err)
	}
	ok2, err := br.Grant(&domain.UserBadge{UserID: 1, BadgeCode: "first_publish_1", AwardEvent: domain.BadgeEventSweep})
	if err != nil || ok2 {
		t.Fatalf("重复颁发应被唯一索引拦截且无错误, ok=%v err=%v", ok2, err)
	}
	m := countBadges(t, db, 1)
	if m["first_publish_1"] != 1 {
		t.Fatalf("唯一索引兜底失败: %d 枚", m["first_publish_1"])
	}
}

// 判定流水可追踪：granted / duplicate / not_met 三种结果均有记录
func TestAwardLogs(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, 1, "ivy")
	svc := NewAchievementService(repository.NewBadgeRepo(db), nil)

	svc.OnTutorialPublished(1) // not_met: first_publish + streak 两个条件组
	seedTutorial(t, db, 1, domain.TutorialStatusPublished, "2026-01")
	svc.OnTutorialPublished(1) // granted: first_publish
	svc.OnTutorialPublished(1) // duplicate: first_publish

	var granted, duplicate, notMet int64
	db.Model(&domain.BadgeAwardLog{}).Where("result = ? AND badge_code = ?", domain.BadgeAwardGranted, "first_publish_1").Count(&granted)
	db.Model(&domain.BadgeAwardLog{}).Where("result = ? AND badge_code = ?", domain.BadgeAwardDuplicate, "first_publish_1").Count(&duplicate)
	db.Model(&domain.BadgeAwardLog{}).Where("result = ? AND badge_code = ?", domain.BadgeAwardNotMet, "first_publish_1").Count(&notMet)
	if granted != 1 || duplicate != 1 || notMet != 1 {
		t.Fatalf("流水应各有 1 条 granted/duplicate/not_met, got %d/%d/%d", granted, duplicate, notMet)
	}
	// granted 流水应携带指标快照
	var log domain.BadgeAwardLog
	db.Where("result = ?", domain.BadgeAwardGranted).First(&log)
	if log.MetricSnapshot == "" || log.Event != domain.BadgeEventTutorialPublished {
		t.Fatalf("发放流水缺少快照或事件信息: %+v", log)
	}
}

// 个人中心视图：未获得显示进度，获得后封顶为门槛值
func TestMyBadgesView(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, 1, "jack")
	br := repository.NewBadgeRepo(db)
	svc := NewAchievementService(br, nil)

	views, earned, err := svc.MyBadges(1)
	if err != nil || earned != 0 || len(views) != len(domain.BadgeRegistry) {
		t.Fatalf("初始视图应为全部未获得, earned=%d len=%d err=%v", earned, len(views), err)
	}
	for _, v := range views {
		if v.Earned {
			t.Fatalf("徽章 %s 初始不应为已获得", v.Code)
		}
	}
	seedTutorial(t, db, 1, domain.TutorialStatusPublished, "2026-01")
	svc.OnTutorialPublished(1)
	views, earned, _ = svc.MyBadges(1)
	if earned != 1 {
		t.Fatalf("应显示 1 枚已获得, got %d", earned)
	}
	var first *BadgeView
	for _, v := range views {
		if v.Code == "first_publish_1" {
			first = v
		}
	}
	if first == nil || !first.Earned || first.AwardedAt == nil || first.Progress != 1 {
		t.Fatalf("首发徽章视图状态错误: %+v", first)
	}
}

func TestLongestMonthRun(t *testing.T) {
	cases := []struct {
		months []string
		want   int
	}{
		{nil, 0},
		{[]string{"2026-01"}, 1},
		{[]string{"2026-01", "2026-02", "2026-03"}, 3},
		{[]string{"2026-01", "2026-02", "2026-04"}, 2}, // 断了一个月
		{[]string{"2025-11", "2025-12", "2026-01"}, 3}, // 跨年连续
		{[]string{"2026-01", "2026-03", "2026-04", "2026-05"}, 3},
	}
	for _, c := range cases {
		if got := longestMonthRun(c.months); got != c.want {
			t.Errorf("longestMonthRun(%v) = %d, want %d", c.months, got, c.want)
		}
	}
}
