package domain

import "time"

// 成就事件类型：用户行为发生后触发对应条件组的判定
const (
	BadgeEventTutorialPublished = "tutorial_published" // 首次发布 / 连续发布
	BadgeEventProjectLiked      = "project_liked"      // 作品点赞
	BadgeEventTutorialFavorited = "tutorial_favorited" // 收藏他人教程
	BadgeEventSweep             = "sweep"              // 定时对账补发
)

// 徽章条件类型
const (
	BadgeCondFirstPublish   = "first_publish"   // 首次发布教程
	BadgeCondProjectLikes   = "project_likes"   // 单个作品点赞达到门槛
	BadgeCondPublishStreak  = "publish_streak"  // 连续 N 个月每月都有发布
	BadgeCondFavoriteOthers = "favorite_others" // 收藏他人教程数量达到门槛
)

// 发放结果（写入 award log，便于追踪）
const (
	BadgeAwardGranted   = "granted"   // 满足条件并成功颁发
	BadgeAwardDuplicate = "duplicate" // 已拥有，重复触发不重复发
	BadgeAwardNotMet    = "not_met"   // 条件尚未达成
)

// BadgeDef 是代码内置的徽章定义。徽章元数据不入库，
// 由 BadgeRegistry 统一管理，运营调整门槛只需改这里。
type BadgeDef struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Category    string `json:"category"`
	Condition   string `json:"condition"`
	// Threshold 判定门槛：
	// first_publish 固定为 1；project_likes / favorite_others 为数量；
	// publish_streak 为连续月数。
	Threshold int `json:"threshold"`
	Sort      int `json:"sort"`
}

// BadgeRegistry 全部徽章定义，顺序即个人中心展示顺序。
var BadgeRegistry = []BadgeDef{
	{
		Code: "first_publish_1", Name: "初次创作", Icon: "🌱",
		Category: "发布", Condition: BadgeCondFirstPublish, Threshold: 1, Sort: 10,
		Description: "发布第一篇教程",
	},
	{
		Code: "project_likes_10", Name: "小有人气", Icon: "👍",
		Category: "点赞", Condition: BadgeCondProjectLikes, Threshold: 10, Sort: 20,
		Description: "单个改造作品获得 10 个赞",
	},
	{
		Code: "project_likes_50", Name: "人气创作者", Icon: "🔥",
		Category: "点赞", Condition: BadgeCondProjectLikes, Threshold: 50, Sort: 21,
		Description: "单个改造作品获得 50 个赞",
	},
	{
		Code: "project_likes_100", Name: "爆款制造机", Icon: "🏅",
		Category: "点赞", Condition: BadgeCondProjectLikes, Threshold: 100, Sort: 22,
		Description: "单个改造作品获得 100 个赞",
	},
	{
		Code: "publish_streak_3", Name: "三月坚持", Icon: "📅",
		Category: "坚持", Condition: BadgeCondPublishStreak, Threshold: 3, Sort: 30,
		Description: "连续 3 个月每月都发布教程",
	},
	{
		Code: "publish_streak_6", Name: "半年笔耕", Icon: "🗓️",
		Category: "坚持", Condition: BadgeCondPublishStreak, Threshold: 6, Sort: 31,
		Description: "连续 6 个月每月都发布教程",
	},
	{
		Code: "favorite_others_10", Name: "收藏家学徒", Icon: "⭐",
		Category: "收藏", Condition: BadgeCondFavoriteOthers, Threshold: 10, Sort: 40,
		Description: "收藏 10 篇他人的教程",
	},
	{
		Code: "favorite_others_50", Name: "灵感收藏家", Icon: "💎",
		Category: "收藏", Condition: BadgeCondFavoriteOthers, Threshold: 50, Sort: 41,
		Description: "收藏 50 篇他人的教程",
	},
}

// BadgeDefByMap 便于按 code 反查定义
var badgeDefByMap = func() map[string]BadgeDef {
	m := make(map[string]BadgeDef, len(BadgeRegistry))
	for _, b := range BadgeRegistry {
		m[b.Code] = b
	}
	return m
}()

func GetBadgeDef(code string) (BadgeDef, bool) {
	b, ok := badgeDefByMap[code]
	return b, ok
}

// UserBadge 用户已获得的徽章。一旦写入永不删除（数据回落不收回）。
type UserBadge struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64 `gorm:"index;not null;uniqueIndex:uk_user_badge,priority:1" json:"user_id"`
	BadgeCode string `gorm:"size:50;not null;uniqueIndex:uk_user_badge,priority:2" json:"badge_code"`
	// AwardEvent 记录由哪类事件触发颁发（实时事件 / sweep 对账补发）
	AwardEvent string `gorm:"size:30;not null;default:sweep" json:"award_event"`
	// MetricSnapshot 颁发瞬间的指标快照，如 {"project_likes":52,"project_id":7}
	MetricSnapshot string    `gorm:"type:text" json:"metric_snapshot"`
	CreatedAt      time.Time `json:"created_at"`
}

func (UserBadge) TableName() string {
	return "user_badges"
}

// BadgeAwardLog 徽章判定与发放流水：每次条件判定都会留痕，
// granted / duplicate / not_met 三种结果均可追踪。
type BadgeAwardLog struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         uint64    `gorm:"index;not null" json:"user_id"`
	BadgeCode      string    `gorm:"size:50;index;not null" json:"badge_code"`
	Event          string    `gorm:"size:30;not null" json:"event"`
	Result         string    `gorm:"size:20;index;not null" json:"result"`
	MetricSnapshot string    `gorm:"type:text" json:"metric_snapshot"`
	Remark         string    `gorm:"size:255" json:"remark"`
	CreatedAt      time.Time `json:"created_at"`
}

func (BadgeAwardLog) TableName() string {
	return "badge_award_logs"
}

// AchievementMetrics 判定时采集的用户成就指标
type AchievementMetrics struct {
	PublishedCount    int      `json:"published_count"`   // 已发布教程数
	MaxProjectLikes   int      `json:"max_project_likes"` // 单个作品最高点赞
	MaxLikesProjectID uint64   `json:"max_likes_project_id"`
	PublishMonths     []string `json:"publish_months"`  // 有发布的月份，格式 YYYY-mm（升序）
	PublishStreak     int      `json:"publish_streak"`  // 截止最近发布月的连续月数（历史最长连段）
	FavoriteOthers    int      `json:"favorite_others"` // 收藏他人教程数
}
